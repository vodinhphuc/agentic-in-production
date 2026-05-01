Phase 4# agentic-in-production — Platform Design

**Status:** Draft for review
**Date:** 2026-05-01
**Type:** Foundational design spec (no predecessor)

---

## 1. Summary

`agentic-in-production` is a **learning-oriented but production-shaped** platform for experimenting with AI agent concepts on **cybersecurity data** following the OCSF (Open Cybersecurity Schema Framework).

It is a monorepo of three services:

- **Frontend** — React + Vite + TypeScript
- **Backend** — Go gateway with an **Adapter** layer that abstracts over the underlying AI agent platform
- **Agent Platform** — an *external* system (Mock for MVP; **GoClaw** for Phase 1; later candidates: Dify, NemoClaw, OpenClaw, or custom)

The three services communicate over a **frozen, versioned, machine-checkable protocol** (`protocols/`), so:
- The frontend is decoupled from AI platform choice.
- Switching platforms is a new Adapter implementation in the backend, not a rewrite.
- Each AI concept (tool use, RAG, multi-agent, memory, AG-UI, evals) becomes a self-contained "experiment" plugged into the same skeleton.

The first end-to-end concept is a **Trino tool-use agent investigating synthetic OCSF security events** — a real-shaped SOC analyst workflow.

---

## 2. Goals & non-goals

### 2.1 Goals
1. Production-shaped patterns: auth, audit log, structured streaming, contract-driven service boundaries, conformance tests, one-command local startup.
2. Strong codebase support so concepts stick: layered docs, ADRs, a project dictionary aligning vocabulary between user and AI coding agents.
3. AI coding agents (Claude Code, etc.) can orient themselves in **under 5 minutes**.
4. Frontend development decoupled from AI platform choice via a frozen contract.
5. Each agent concept is a self-contained experiment; ships independently without breaking earlier ones.
6. Locally deployable (`docker compose up` + `make up`); single-tenant.

### 2.2 Non-goals (for now)
1. Cloud deployment, multi-tenant, OAuth/SSO, billing.
2. Building a custom agent runtime from scratch — the **agent platform IS the agent**.
3. Toy ecommerce/blog datasets — domain is cybersecurity throughout.
4. Auto-generated tutorials or hosted documentation sites.
5. OpenTelemetry / Prometheus telemetry pipeline for MVP (Phase 2+).

---

## 3. Architecture

### 3.1 Three services, three responsibilities

```
┌──────────────────┐    HTTP/SSE    ┌────────────────────┐    HTTP/SSE    ┌──────────────────┐
│  Frontend (web)  │ ─────────────► │  Backend (gateway) │ ─────────────► │  Agent Platform  │
│  React + Vite    │ ◄───────────── │  Go (chi/echo)     │ ◄───────────── │  (external)      │
└──────────────────┘                └────────────────────┘                └──────────────────┘
                                              │
                                              │ owns: end-user auth, sessions,
                                              │       rate limit, audit log,
                                              │       agent registry, ADAPTER LAYER
                                              ▼
                                       ┌────────────┐
                                       │ Postgres   │
                                       │ (sessions, │
                                       │  audit,    │
                                       │  registry) │
                                       └────────────┘
```

The Backend is **not a thin proxy**. It is the production access layer:
- End-user authentication (single admin for MVP).
- Mapping our `session_id` → a platform `conversation_id` (the platform is stateful).
- Tool gating, rate limiting, full audit log of every tool call.
- The Adapter layer that translates between the canonical event envelope and each platform's native API.

The Agent Platform is **external** (or in-process via the Mock adapter for MVP). It owns:
- The LLM agent loop and tool execution.
- Conversation history and platform-side authentication (API keys).
- Its own native streaming format (different per platform).

The frontend never sees an LLM API key, never sees a Trino credential, never picks an agent version — the browser only ever talks to the Backend.

### 3.2 Request flow for a single user message

1. Browser → `POST /api/sessions/{id}/messages` (Backend) — body has user text.
2. Backend authenticates, persists user message, opens SSE stream back to browser.
3. Backend → adapter → platform's native `run` API — opens its own SSE stream upstream.
4. Agent runs: LLM call → may emit tool-call events → executes tool → feeds result back to LLM → emits tokens.
5. Adapter translates each platform-native event into a **canonical event** and yields it.
6. Backend forwards each canonical event to the browser AND persists tool-call events to the audit log.
7. Stream closes when the agent emits `run_finished`.

### 3.3 Why three services and not two

If the agent platform were called from the browser directly:
- LLM and tool-platform credentials would have to be in the client.
- There would be no server-side audit trail of what tools were called.
- Switching agent platforms would force a frontend redeploy.

The Go backend exists specifically to make those problems go away.

---

## 4. Protocol-first discipline (the load-bearing principle)

> **The frontend is decoupled from the AI agent platform by a frozen, versioned, machine-checkable contract. The Adapter is the only layer that ever knows what platform is in use.**

If this principle holds, the frontend team never reads platform docs. AI engineers can swap GoClaw → Dify → NemoClaw — the frontend doesn't move, the contract doesn't move, only one Adapter folder changes.

### 4.1 Artifacts in `protocols/`, committed *before any service code*

```
protocols/
├── README.md                  ← what this folder is, versioning policy, hard rules
├── VERSION                    ← single line: "v1" (the lock)
├── openapi.yaml               ← REST: auth, sessions, messages, audit
├── agent-events.schema.json   ← JSON Schema for the canonical event envelope
├── agent-protocol.md          ← human-readable lifecycle, ordering invariants, error semantics
└── examples/                  ← canonical JSON examples for every event type
    ├── text-only-stream.json
    ├── tool-call-success.json
    ├── tool-call-failure.json
    └── multi-step-investigation.json
```

These five things are the contract. Backend, frontend, adapters, mocks, and conformance tests are all downstream of them.

### 4.2 v1 of the canonical event envelope

Every event on the wire (Backend ↔ Frontend, Adapter ↔ Backend) is one of:

```jsonc
{ "type": "run_started",      "run_id": "..." }
{ "type": "text_delta",       "text": "Looking at the orders table..." }
{ "type": "tool_call_start",  "call_id": "c1", "tool": "execute_query",
                              "args": { "sql": "SELECT ..." } }
{ "type": "tool_call_end",    "call_id": "c1", "ok": true,
                              "result_preview": "12 rows, columns: ..." }
{ "type": "state_update",     "key": "current_table", "value": "orders" }
{ "type": "error",            "code": "...", "message": "..." }
{ "type": "run_finished",     "reason": "done" | "stopped" | "error" }
```

This shape is deliberately **AG-UI-aligned** to keep the migration cost to near-zero if/when AG-UI is adopted (see §4.5).

### 4.3 Versioning policy (`protocols/README.md` codifies this)

| Change | Treatment |
|---|---|
| Add an optional field to an event | additive — no version label change |
| Add a new event type | additive |
| Make a field required, rename, change type, remove a field, remove an event | **breaking — requires v2** |
| Breaking rollout | v1 and v2 coexist on `/api/v1/...` and `/api/v2/...` for at least one phase |

Every protocol change requires an ADR.

### 4.4 Generated types and runtime validation

```
protocols/openapi.yaml             ──► frontend/src/api/types.ts        (TS, generated)
protocols/agent-events.schema.json ──► frontend/src/events/types.ts     (TS, generated)
protocols/agent-events.schema.json ──► backend/internal/protocol/       (Go structs + validator, generated)
```

`make gen` produces these artifacts. CI fails if `git diff` after `make gen` is non-empty.
Runtime validation (in dev/test): the Backend validates every adapter-emitted event against the schema before forwarding. Schema-violating adapters fail loudly.

### 4.5 Open question — long-term wire protocol

v1 is a deliberate **AG-UI-aligned minimum**, not a permanent commitment. **Phase 0.0.a** (below) is a 1–2 day research spike evaluating:

- **AG-UI** (CopilotKit's Agent-User Interaction Protocol) — strongest current candidate; v1 is structurally close.
- **A2UI** — flagged by user as interesting; needs verification of current state and adoption.
- MCP-shaped events, OpenAI Realtime events, vendor-native (GoClaw, Dify, etc.) for completeness.

The output is **ADR-0003** recommending v1's shape and naming the migration target. Until then v1 stays minimal and platform-agnostic, with no platform-specific semantics in event payloads (anything platform-specific goes in opaque `metadata.platform`).

### 4.6 Hard rules (added to root `CLAUDE.md`)

```
- Do not modify protocols/ without writing an accompanying ADR.
- Do not edit generated types by hand. Run `make gen`.
- Do not bypass schema validation in adapters or backend.
- Breaking protocol changes require a new version (v2). Never silently change v1.
- Frontend team owns nothing in protocols/. Backend proposes; ADR consensus changes.
```

---

## 5. Adapter architecture (Backend ↔ Platform)

### 5.1 The interface

```go
// backend/internal/adapters/adapter.go
type AgentPlatformAdapter interface {
    Name() string
    Capabilities(ctx context.Context) (Capabilities, error)

    StartConversation(ctx context.Context, sess Session) (PlatformConvID, error)
    EndConversation  (ctx context.Context, id PlatformConvID) error

    Run(ctx context.Context, req RunRequest) (<-chan AgentEvent, error)
}
```

```
backend/internal/adapters/
├── adapter.go        # interface + canonical types
├── registry.go       # name → factory, picked from config / agent registry table
├── mock/             # in-process, scripted scenarios — non-negotiable
├── goclaw/           # (Phase 1)
├── dify/             # (later)
├── nemoclaw/         # (later)
└── openclaw/         # (later)
```

### 5.2 Stateful platform contract

The platform owns conversation history. The Backend stores only:
- Its own audit trail of tool calls and emitted events.
- The mapping `session_id ↔ platform_conv_id` in Postgres.
- Per-platform connection settings (`base_url`, `api_key_ref`, `agent_id_in_platform`).

When the Backend asks the adapter to run, it sends only the new user message — not history. History is the platform's job.

### 5.3 Two-layer auth

| Boundary | Auth mechanism |
|---|---|
| Frontend ↔ Backend | end-user JWT in cookie (single admin for MVP) |
| Backend ↔ Platform | per-platform API key in `.env` / secrets (never visible to frontend) |

### 5.4 Conformance tests — what proves the abstraction holds

Per-adapter test suites:
1. **Unit** — adapter logic with the platform's HTTP API mocked from recorded fixtures.
2. **Conformance** — same scripted scenarios run against every adapter (Mock, Dify, …); each must emit a canonical event sequence matching golden files.

Scenarios live in `backend/internal/adapters/conformance/scenarios/*.yaml`. Adding a new adapter is "implement the interface + pass conformance" — no architectural changes.

---

## 6. Repo layout

```
agentic-in-production/
├── README.md                    # project overview + quickstart
├── CLAUDE.md                    # top-level guidance for AI coding agents
├── AGENTS.md                    # thin pointer to CLAUDE.md (for Cursor / Aider / etc.)
├── docker-compose.yml           # one-command local stack
├── Makefile                     # common dev commands (POSIX-shell-friendly)
├── .env.example
│
├── frontend/                    # React + Vite + TypeScript
│   ├── README.md  CLAUDE.md
│   ├── src/  package.json
│
├── backend/                     # Go gateway service
│   ├── README.md  CLAUDE.md
│   ├── cmd/server/
│   ├── internal/
│   │   ├── auth/  sessions/  audit/
│   │   ├── agentregistry/       # which adapters exist, which are enabled
│   │   ├── adapters/            # the Adapter implementations
│   │   ├── protocol/            # generated types + runtime validator
│   │   └── httpapi/             # HTTP handlers + SSE streaming
│   └── go.mod
│
├── agents/                      # PLATFORM CONFIGURATION (not code)
│   ├── README.md                # how to add a new agent
│   ├── _template/               # scaffold for a new agent's config
│   └── trino-agent/             # Phase 1: prompt, tool defs, platform export
│       ├── README.md            # what THIS agent teaches
│       ├── prompts/
│       └── tools.yaml
│
├── protocols/                   # SHARED CONTRACTS — see §4
│   ├── README.md  VERSION
│   ├── openapi.yaml
│   ├── agent-events.schema.json
│   ├── agent-protocol.md
│   └── examples/
│
├── infra/                       # local dev infrastructure
│   ├── trino/                   # Trino + Hive metastore + catalog config (Phase 1)
│   ├── postgres/                # init SQL for sessions / audit / registry
│   └── seed-data/               # OCSF synthetic dataset generator (Phase 1)
│
├── docs/                        # learning-first documentation — see §7
│   ├── README.md                # learning index — read first
│   ├── dictionary.md            # shared vocabulary (user ↔ AI agent)
│   ├── concepts/                # numbered conceptual essays
│   ├── adr/                     # architecture decision records
│   ├── diagrams/                # mermaid sources
│   └── superpowers/specs/       # design specs (this file lives here)
│
├── scripts/                     # dev helpers
└── .github/workflows/           # CI
```

**Key invariants of the layout:**
- `agents/<name>/` is a hard boundary holding **platform-side configuration** (prompts, tool definitions, exported workflow files). Not Python code, not service deps. Each agent's configuration is self-contained so swapping or removing one never touches the others.
- The corresponding **adapter code** for each agent is the parallel folder under `backend/internal/adapters/<platform>/`. Heavy adapter dependencies (a Dify SDK client, a vector-store client added for a future RAG agent) live in the backend's Go module and are isolated by Go's package boundaries — they never leak across adapters.
- `protocols/` is top-level and authoritative. Both backend and frontend are *consumers*, not owners.
- `docs/concepts/` is for the user (learner). Service READMEs are for the operator. CLAUDE.md files are for AI coding agents.

---

## 7. Documentation strategy

### 7.1 Three audiences, three doc types

| Audience | Reads to… | Doc type |
|---|---|---|
| User (learner) | internalize concepts | `docs/concepts/`, `docs/adr/`, design specs |
| AI coding agents | understand code well enough to edit | `CLAUDE.md` files, `docs/dictionary.md`, `protocols/` |
| User (operator) | run / debug / extend | `README.md` (root + per-service), runbooks |

### 7.2 Five document types

| Doc | Length | Audience | Written when |
|---|---|---|---|
| Concept essay | 300–800 words | learner | **after** an experiment ships |
| ADR | <300 words | learner + future-self | when a non-obvious tradeoff is made |
| Design spec | as long as needed | learner + AI agent | before each implementation plan |
| Service README | <200 words | operator + AI agent | when service is created |
| CLAUDE.md | <150 words per file | AI agent | per service |
| **dictionary.md** | grows over time | **user + AI agent (shared)** | term added the moment a misunderstanding surfaces |

### 7.3 The "concept essay follows code" rule

Each shipped experiment must produce **at least one concept essay before it is considered done**. Concept essays are not written before the code — written before, they're Wikipedia articles. Written after, with `// see backend/internal/audit/audit.go` links to the user's own code, they're personal understanding artifacts the user will re-read in two years and still gain from.

The user writes the prose; AI assistants may write outlines and pose questions to answer.

### 7.4 `docs/dictionary.md` — the shared glossary

The only doc in the repo with **both** user and AI coding agent as primary audience. Aligned vocabulary prevents the most expensive class of misunderstanding (silent term-mismatch where both sides think they're agreeing).

What goes in:
- Project-coined terms ("canonical event envelope", "platform conversation").
- Ambiguous terms ("Agent" — coding agent? platform? LLM loop?).
- Alias-prone terms (we pick one canonical name).
- Domain terms the agent needs context for (OCSF classes, security investigation vocabulary).

Entry format:

```markdown
### Adapter
**Aliases:** `AgentPlatformAdapter`, "platform adapter"
**Definition:** A Go implementation that translates between our **canonical event envelope**
and a specific external Agent Platform. Lives under `backend/internal/adapters/<name>/`.
**Don't confuse with:** Trino's "connector" (data-source plugin, Trino's vocabulary).
**Code:** `backend/internal/adapters/adapter.go`
```

**Starter entries** (seeded in Phase 0):
Agent · Agent Platform · Adapter · Canonical event envelope · Session · Conversation · Run · Tool call · Experiment · Phase · Detection Finding · OCSF class.

**Maintenance rule:** add an entry the moment a misunderstanding surfaces in conversation — that's the highest-value moment because it's evidence the term needed disambiguating.

### 7.5 `docs/README.md` — the learning index

Not a TOC, a map. Concepts in the order they were learned, with links to the experiments that drove them and the code that implements them. The file the user opens after losing the thread.

### 7.6 Diagrams: mermaid in markdown

GitHub renders mermaid; VS Code does too. Lives in the same PR as the code it describes. No Excalidraw PNG exports — they decay.

---

## 8. AI coding agent support

Calibrated to: **a fresh AI coding agent (or future-self) can orient in <5 minutes.**

### 8.1 CLAUDE.md hierarchy

```
/CLAUDE.md                      ← project-wide rules, map of where to look
/AGENTS.md                      ← thin "see CLAUDE.md" pointer
/backend/CLAUDE.md              ← Go conventions, adapter rules
/frontend/CLAUDE.md             ← React/TS conventions, codegen rules
/agents/CLAUDE.md               ← "this folder is platform configuration, not code"
```

Scoped, not one mega-file. When working in `backend/`, only Go-specific guidance loads — not React rules that would confuse the agent.

### 8.2 Root `/CLAUDE.md` — sketch

```markdown
# agentic-in-production

Three-tier learning platform: React → Go gateway (with Adapter pattern) → external Agent Platform.

## Read these before using project vocabulary
- docs/dictionary.md — terms with project-specific meaning. ALWAYS check.
- protocols/ — the wire contracts. Treat as source of truth.

## Where docs live
- docs/README.md — learning index (start here)
- docs/concepts/ — conceptual essays (the "why")
- docs/adr/ — architecture decisions (read before changing them)
- docs/superpowers/specs/ — design specs for in-flight work

## Sub-agent guidance
- backend/ ? Read backend/CLAUDE.md too.
- frontend/ ? Read frontend/CLAUDE.md too.
- agents/ ? Read agents/CLAUDE.md (it's config, not code).

## Hard rules — do NOT do these without asking
- Add a top-level dep to any service.
- Modify protocols/ (the wire contract) without an ADR.
- Add a platform adapter that doesn't pass the conformance suite.
- Bypass the backend's audit log.
- Skip writing the concept essay when finishing an experiment.

## Workflow skills
- Starting any feature → brainstorming
- Implementing → writing-plans then executing-plans
- Bug → systematic-debugging first
- Before "done" → verification-before-completion
- Major step done → requesting-code-review
```

### 8.3 `.claude/settings.json` — minimal

Pre-allow safe dev-loop commands so AI agents don't ping for permission on every test run:

```jsonc
{
  "permissions": {
    "allow": [
      "Bash(make *)",
      "Bash(go test:*)", "Bash(go build:*)", "Bash(go vet:*)", "Bash(go mod:*)",
      "Bash(gofmt:*)",
      "Bash(pnpm install)", "Bash(pnpm run *)", "Bash(pnpm test:*)",
      "Bash(uv run:*)", "Bash(uv sync)",
      "Bash(docker compose ps)", "Bash(docker compose logs *)"
    ]
  }
}
```

Notably **not** allowed: `docker compose down`, `git push`, anything that mutates external state.

### 8.4 Hooks — start with **none**

Claude-specific hooks are easy to over-engineer. The standard `pre-commit` git hook running `make verify` (lint + types + tests) is enough. Add Claude hooks only when a specific recurring problem justifies one.

### 8.5 MCP servers — start with **none specific to this project**

Project-specific MCP (e.g. a Postgres MCP, Trino MCP) would let the AI coding agent skip the actual product flow (frontend → backend → adapter → agent → tools → Trino) and short-circuit the system being built. Default off; revisit only with a concrete justification.

### 8.6 Subagents and skills — use what exists

The available superpowers skills (brainstorming, writing-plans, executing-plans, TDD, debugging, verification-before-completion) cover the workflow. No custom subagents or skills until a real repeated workflow demands one.

### 8.7 The 5-minute orientation test

1. Open `/CLAUDE.md` → see structure, rules, where docs live (1 min).
2. Open `docs/README.md` → see learning index, find relevant concept (1 min).
3. Open `docs/dictionary.md` → understand project-specific vocabulary (1 min).
4. Open the relevant service's `CLAUDE.md` → load language-specific rules (1 min).
5. Read the most recent design spec for in-flight context (1 min).

---

## 9. Tooling, tests, dev loop

### 9.1 Makefile — single source for dev commands

POSIX-shell-friendly so it works in Git Bash today and natively on Ubuntu after the user migrates.

```makefile
.PHONY: up down seed test test-backend test-frontend lint lint-backend lint-frontend \
        fmt gen verify logs

up:
	docker compose up -d postgres
	cd backend && air & cd frontend && pnpm dev

down:
	docker compose down

seed:                # Phase 1+
	python infra/seed-data/generate_ocsf.py

test: test-backend test-frontend

test-backend:
	cd backend && go test ./...

test-frontend:
	cd frontend && pnpm test

lint: lint-backend lint-frontend

lint-backend:
	cd backend && golangci-lint run

lint-frontend:
	cd frontend && pnpm lint

fmt:
	cd backend && gofmt -w .
	cd frontend && pnpm fmt

gen:
	cd protocols && ./gen-types.sh

verify: lint test       # the merge gate; pre-commit hook runs this

logs:
	docker compose logs -f $(SVC)
```

`make verify` is **the** gate. Pre-commit hook runs it. CI runs it.

### 9.2 Test strategy

| Layer | What | Tool |
|---|---|---|
| Backend unit | handlers, adapter logic, audit, auth | `go test` + `testify` |
| Backend conformance | scripted scenarios → every adapter emits same canonical events | `go test` + golden YAML files |
| Frontend unit | components, hooks, event-stream parser | Vitest + Testing Library |
| Contract | TS types in sync with `protocols/openapi.yaml` and `agent-events.schema.json` | typecheck after `make gen` |
| End-to-end | one golden path: login → session → message → see tool calls in UI | Playwright (1 test for MVP) |

Conformance test scenarios live in `backend/internal/adapters/conformance/scenarios/*.yaml` and double as documentation of the adapter contract.

### 9.3 Hot reload

| Service | Tool | Restart needed for routine edits? |
|---|---|---|
| Frontend | Vite | no — full HMR |
| Backend | `air` | no — Go save → rebuild + restart |
| Mock adapter scenarios | YAML watcher | no — reload per request |
| Postgres | — | only on schema migration |

### 9.4 CI

Three jobs in parallel on every PR (target <4 min):
- `backend` — lint, vet, `go test ./...`, conformance.
- `frontend` — lint, typecheck, vitest, build.
- `contract` — `make gen && git diff --exit-code` (fails if codegen out of sync).

One sequential e2e job after them passes (target <2 min more).

### 9.5 Codegen — narrow, one-directional

Source of truth: `protocols/openapi.yaml` + `protocols/agent-events.schema.json`.
Generated: TypeScript types (frontend), Go structs + validator (backend protocol package).

**Not generated:** Go server stubs from OpenAPI. With one backend developer, hand-written handlers + per-route integration tests are faster and clearer than wrestling generator output. Revisit when team grows.

### 9.6 Observability for MVP

- Structured JSON logging via `slog` (backend) and `pino` or `console` (frontend).
- Correlation ID `run_id` flows frontend → backend → adapter → audit log → response. **Single most useful debugging primitive in agent systems.**
- No tracing/metrics yet (Phase 2+).

---

## 10. First slice: Phase 0 (MVP) and Phase 1 (Trino POC)

> **Implementation-plan scope:** the *first* implementation plan derived from this spec covers **Phase 0 only**. Phase 1 is described here in enough detail to verify the architecture supports it, but it gets its own implementation plan when Phase 0 ships. Phases 2–6 in §12 are roadmap, not work-in-flight.

### 10.1 Phase 0 — MVP Skeleton (Mock-first, no external deps)

Reordered to be **protocol-first** per §4.

#### Phase 0.0 — Protocol definition (no service code)

**0.0.a — Research spike (timeboxed, 1–2 days):**
- Evaluate AG-UI, A2UI, MCP-shaped events, vendor-native (Dify SSE, Anthropic SSE) for: documentation maturity, community adoption, event vocabulary fit, transport flexibility, generative-UI readiness, translation distance to/from each other.
- Output: **ADR-0003** — "Wire protocol for agent event stream." Recommends v1 shape and migration target.

**0.0.b — Author v1 artifacts (informed by 0.0.a):**
- `protocols/openapi.yaml`, `protocols/agent-events.schema.json`, `protocols/agent-protocol.md`, `protocols/examples/*.json`, `protocols/VERSION` = `"v1"`.
- Run `make gen`; commit generated TS types and Go structs.

#### Phase 0.1 — Parallelizable dev (frontend & backend simultaneously)

The contract being frozen means the frontend and backend tracks can progress **independently** — the work is parallelizable whether it's done by one person switching hats or by separate developers.

- **Frontend track:** build the entire UI against `protocols/examples/*.json` fixtures. No backend needed.
- **Backend track:** build the Adapter framework + the Mock adapter with scripted scenarios.

Each track is blocked only by the contract, not by the other.

#### Phase 0.2 — Integration

- Frontend talks to Backend talks to Mock adapter end-to-end.
- One Playwright e2e test of the golden path passes.
- Conformance test suite runs against the Mock adapter.

#### Phase 0 — Definition of done

- `make up` brings the full stack up; browse to `localhost:5173`, log in as admin, see one available agent: `mock-trino-flavored`.
- Start a session, send a message. SSE stream returns canned events that *look like* a Trino agent ran (`tool_call_start{tool:"execute_query"}` → `tool_call_end{result_preview:"…"}` → `text_delta` → `run_finished`).
- Frontend renders user text **and** tool calls as collapsible cards.
- Audit log endpoint shows the same tool-call events, persisted in Postgres.
- CI green: lint, types, unit tests, conformance, contract (codegen-up-to-date), one Playwright e2e.
- Concept essays: `01-the-agent-loop.md` and `02-streaming-protocols.md`.
- ADRs: 0001 (three-tier-with-go-gateway), 0002 (mock-first-then-real-platform), 0003 (wire protocol from spike).

#### Phase 0 — explicitly out of scope

Real LLM, real Trino, real platform, multi-user, deployment.

### 10.2 Phase 1 — Trino tool-use POC (the first real concept)

#### What gets added on top of the MVP skeleton

1. **Local Trino + sample dataset** under `infra/trino/` and `infra/seed-data/`.
   - Hive connector + parquet files on local FS (or MinIO). Production-shaped — same pattern as real SIEM data lakes.
   - Bootstrapped with a single SQL/Python script. `make seed` regenerates.

2. **First real platform adapter: GoClaw.** Per ADR-0009, GoClaw is the first non-Mock adapter. Other platforms (Dify, NemoClaw, OpenClaw) become later additions that prove the abstraction holds across vendors.

3. **OCSF toy dataset** — see §11.

4. **Trino tool surface** (configured *in the platform*, not in our backend): five focused tools.
   - `list_catalogs() → [name]`
   - `list_schemas(catalog) → [name]`
   - `list_tables(catalog, schema) → [name]`
   - `describe_table(catalog, schema, table) → columns`
   - `execute_query(sql, max_rows=100) → rows` *(read-only, hard row cap, query timeout, blocked DDL/DML)*

   Five narrow tools — not one open `execute_sql` — so the agent's reasoning shows up as schema-discovery → query-formation → execution. That's the actual tool-use loop the user is learning.

5. **Adapter conformance tests** — scripted scenarios from MVP now also run against the real platform adapter (with Trino mocked at the SQL level via fixtures).

#### Phase 1 — Definition of done

- User asks: *"There's an alert from last night about suspicious PowerShell on `WIN-WS-014` around 02:14 UTC. Investigate."*
- Stream shows: `describe_table(detection_finding)` → `execute_query` → `describe_table(process_activity)` → `execute_query` → `describe_table(authentication)` → `execute_query` → `describe_table(network_activity)` → `execute_query` → natural-language timeline answer.
- Audit log records every SQL query executed, by whom, and the row count returned.
- A "bad" question (e.g. "delete all orders") is refused — read-only enforcement holds.
- Concept essays: `03-tool-use-and-narrow-tools.md`, `04-the-adapter-pattern-for-platforms.md`, `05-ocsf-as-an-agent-data-shape.md`.

---

## 11. OCSF toy dataset (Phase 1)

### 11.1 Tables (4 OCSF event classes — narrow on purpose)

| Table | OCSF class | What's in it |
|---|---|---|
| `authentication` | 3002 — Authentication | logins/logouts, success/failure, MFA, src IP, user, host |
| `process_activity` | 1007 — Process Activity | process executions with parent/child PIDs, cmdline, hash, host |
| `network_activity` | 4001 — Network Activity | TCP/UDP connections: src/dst IP, port, bytes, duration |
| `detection_finding` | 2004 — Detection Finding | pre-existing alerts/findings the agent can pivot from |

(DNS Activity 4003 is a tempting fifth — held for a later OCSF experiment.)

### 11.2 Schema fidelity

Real OCSF column names, nested as Trino struct types where OCSF is nested. `describe_table` returns `time`, `class_uid`, `class_name`, `severity_id`, `actor.user.name`, `device.hostname`, `process.cmd_line`, `src_endpoint.ip`, `dst_endpoint.ip`, etc. The agent learns to read OCSF schema, not toy schema.

### 11.3 Storage

Parquet files on local FS, exposed via Trino's **Hive connector** with a local Hive metastore. The exact pattern real SIEM data lakes use (Parquet on S3/MinIO, Hive metastore, queried by Trino/Athena/Starburst).

### 11.4 Volume

~30 days of events, ~100k rows total, 5–10 MB on disk. Small enough to commit-or-regenerate; large enough that filters and JOINs matter.

### 11.5 The synthetic incident woven through the data

```
[T-0]   external IP 185.x.x.x → 14 failed auths against user `j.smith` over 4 min
[T+5m]  one successful auth from same IP — credential stuffing succeeded
[T+12m] j.smith's host (WIN-WS-014) launches:
        powershell.exe -EncodedCommand <base64-suspicious> (parent: explorer.exe)
[T+13m] same host → outbound connection to 185.y.y.y:443, 8 MB upload
[T+1h]  detection_finding row exists: "suspicious_powershell_execution" on WIN-WS-014
        — but no link to the auth events; the agent has to make the connection
```

Plus background noise: ~50 normal users, normal logins from corporate IPs, normal process trees, benign network traffic. Incident is ~0.1% of events.

The user has flagged that the dataset may evolve in later experiments (different OCSF classes, different incident scenarios, different scale).

---

## 12. Concept roadmap

```
Phase 0  ─  MVP Skeleton                    Mock adapter, frontend, audit, e2e
            Concepts: agent loop · streaming · canonical events

Phase 1  ─  Trino Tool-Use POC  ★            GoClaw adapter + OCSF + 5 narrow tools
            Concepts: tool use · adapter pattern · OCSF as agent data shape

Phase 2  ─  RAG / Knowledge-Augmented        Vector store + corpus of detection runbooks
            Concepts: embeddings · retrieval · grounded citations

Phase 3  ─  Multi-Agent Orchestration        Supervisor → {investigator, responder, reporter}
            Concepts: agent specialization · routing · inter-agent contracts

Phase 4  ─  Memory / Persistent Agent State  Cross-session "case notes"
            Concepts: short-term vs long-term memory · what to persist

Phase 5  ─  Evals Harness                    Canonical investigations + golden traces
            Concepts: evaluating agents · regression vs quality
```

**Why this ordering:** RAG is just another tool — comes after tool-use is solid. Multi-agent has more value once individual agents do something useful — comes after. Memory is intentionally late because the simple version is "Postgres table the backend manages" and you understand what to persist only after watching dozens of runs. Evals are last because without varied real behavior they're theatrical.

**Note on the previously-planned AG-UI / A2UI phase:** The wire-protocol decision is now made upfront in **Phase 0.0.a** (the research spike) — adopting AG-UI / A2UI / a custom shape is a Phase 0 deliverable, not a future experiment. Generative-UI rendering (the *UX* layer that builds on top of `state_update` events) remains a possible future addition; if you want it as an explicit phase later, we can slot it in as a new Phase 6.

**Each phase carries the same definition of done:**
- Plugs into the existing skeleton without breaking earlier phases.
- Conformance tests still pass against Mock scenarios (and any new ones).
- 1–2 concept essays land in `docs/concepts/`.
- An ADR captures any non-obvious decision.

**Architecture leaves room for, but does not pre-build:** real OAuth, multi-tenant, cloud deployment, telemetry/tracing pipeline. These come later if the project graduates from B-scope to C-scope.

---

## 13. Open questions

1. **Wire protocol selection** (§4.5) — AG-UI vs A2UI vs custom-translatable. Resolved by Phase 0.0.a research spike → ADR-0003.
2. **Dataset evolution** — current OCSF toy dataset may be replaced/extended for later experiments per user request.

### 13.1 Resolved during this spec

- **First real platform (Phase 1):** **GoClaw** — committed 2026-05-01. See ADR-0009.

---

## 14. ADRs to extract from this spec

These decisions in this spec deserve to be split into individual ADR files when the repo is bootstrapped:

- **ADR-0001** — Three-tier with Go gateway (rationale: §3, why three services not two).
- **ADR-0002** — Mock-first, real platform later (rationale: §10.1, lightweight start).
- **ADR-0003** — Wire protocol for agent event stream (output of Phase 0.0.a research spike).
- **ADR-0004** — Stateful agent platform; Backend stores only audit + session mapping (rationale: §5.2).
- **ADR-0005** — JSON Schema for events (not AsyncAPI yet) (rationale: §4.3).
- **ADR-0006** — Hand-written Go handlers, not generated from OpenAPI (rationale: §9.5).
- **ADR-0007** — `docs/dictionary.md` as a first-class shared-vocabulary doc (rationale: §7.4).
- **ADR-0008** — OCSF toy dataset, narrow 4 tables, Hive+Parquet (rationale: §11).
- **ADR-0009** — Phase 1 platform = GoClaw (rationale: §10.2; user decision 2026-05-01).

ADR-0001, 0002, 0004–0009 can be authored from this spec directly. ADR-0003 follows the research spike.
