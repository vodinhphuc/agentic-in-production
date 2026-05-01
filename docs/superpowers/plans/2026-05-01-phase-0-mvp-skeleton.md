# Phase 0 — MVP Skeleton Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use `superpowers:subagent-driven-development` (recommended) or `superpowers:executing-plans` to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Stand up the MVP skeleton from [the design spec](../specs/2026-05-01-agentic-platform-design.md): protocol committed and validated, frontend renders against fixtures, backend serves the Mock adapter end-to-end with audit logging, single Playwright e2e test green, CI on parallel jobs in <4 min.

**Architecture:** Three-tier (React + Go gateway + external Agent Platform). Phase 0 ships only the Mock adapter — no real LLM, no real Trino. Frontend and backend tracks (Phase 0.1) develop in parallel against the protocol frozen in Phase 0.0.

**Tech stack:**

| Layer | Choice | Why |
|---|---|---|
| Go router | [chi](https://github.com/go-chi/chi) | smallest, stdlib-shaped |
| Postgres driver | [pgx/v5](https://github.com/jackc/pgx) | PG-native, fast |
| Go test | [testify](https://github.com/stretchr/testify) | standard |
| Go schema validator | [santhosh-tekuri/jsonschema/v5](https://github.com/santhosh-tekuri/jsonschema) | JSON Schema 2020-12 |
| Go hot reload | [air](https://github.com/cosmtrek/air) | dev loop |
| Frontend | React 18 + TypeScript + [Vite](https://vitejs.dev) | standard |
| State | [Zustand](https://github.com/pmndrs/zustand) | minimal global state |
| Test (unit) | [Vitest](https://vitest.dev) + Testing Library | native Vite |
| Test (e2e) | [Playwright](https://playwright.dev) | golden-path coverage |
| TS types from OpenAPI | [openapi-typescript](https://github.com/drwpow/openapi-typescript) | types-only |
| TS types from JSON Schema | [json-schema-to-typescript](https://github.com/bcherny/json-schema-to-typescript) | discriminated unions |
| Go types from OpenAPI | [oapi-codegen](https://github.com/deepmap/oapi-codegen) types-only | not handlers (per ADR-0006) |
| Go types from JSON Schema | [omissis/go-jsonschema](https://github.com/omissis/go-jsonschema) | structs |

**Phase structure (checkpoint commits between phases):**

- **Phase 0.0** — Foundation. Repo bootstrap + ADRs + protocol artifacts + codegen pipeline. Tasks 1–8.
- **Phase 0.1a** — Backend track. Postgres + auth + sessions + audit + adapter framework + Mock adapter + SSE endpoint + conformance tests. Tasks 9–17.
- **Phase 0.1b** — Frontend track (parallelizable with 0.1a). Vite project + generated types + auth UI + sessions UI + chat UI with tool-call cards + audit view. Tasks 18–24.
- **Phase 0.2** — Integration + DoD. docker-compose full stack + Playwright e2e + CI workflows + pre-commit hook + concept essays. Tasks 25–30.

---

## Conventions used throughout

- **Exact file paths.** Every step names files explicitly.
- **Complete code blocks** for project-specific logic. For well-known boilerplate (e.g. `vite create` output), only the project-specific edits are shown.
- **Test-first** for backend and frontend logic. Tests authored before implementation, run to verify failure, then implementation.
- **One commit per task** unless a task explicitly directs otherwise. Conventional Commits style.
- **All commands run from repo root** unless otherwise noted.

---

## Phase 0.0 — Foundation

### Task 1: Bootstrap repo skeleton

Get the root files in place so `make`, `docker compose config`, and `git status` all work cleanly before any code lands.

**Files:**
- Create: `.gitignore`, `.env.example`, `Makefile`, `docker-compose.yml`, `README.md`

- [ ] **Step 1: Create `.gitignore`**

```gitignore
# Editor
.vscode/
.idea/
*.swp
# OS
.DS_Store
Thumbs.db
# Go
backend/bin/
backend/tmp/
*.test
*.out
*.coverprofile
# Node
frontend/node_modules/
frontend/dist/
frontend/.vite/
frontend/playwright-report/
frontend/test-results/
# Env (template is committed; real .env is not)
.env
.env.local
# Phase 1+
infra/seed-data/output/
infra/trino/data/
```

- [ ] **Step 2: Create `.env.example`**

```bash
# Postgres (used by docker-compose + backend)
POSTGRES_USER=aip
POSTGRES_PASSWORD=aip
POSTGRES_DB=aip
POSTGRES_PORT=5432

# Backend
BACKEND_PORT=8080
JWT_SIGNING_KEY=dev-only-replace-32-bytes-random-please
ADMIN_USERNAME=admin
ADMIN_PASSWORD_HASH=  # set via `make admin-password` after Task 11

# Frontend (Vite reads VITE_-prefixed only)
VITE_API_BASE_URL=http://localhost:8080
```

- [ ] **Step 3: Create `Makefile`**

```makefile
.PHONY: up down dev-backend dev-frontend install hooks gen \
        test test-backend test-frontend test-e2e \
        lint lint-backend lint-frontend fmt verify logs admin-password

up:
	docker compose up -d postgres

down:
	docker compose down

dev-backend:
	cd backend && air

dev-frontend:
	cd frontend && pnpm dev

install:
	cd backend && go mod download
	cd frontend && pnpm install

hooks:
	cp scripts/pre-commit .git/hooks/pre-commit
	chmod +x .git/hooks/pre-commit

gen:
	cd protocols && bash gen-types.sh

test: test-backend test-frontend
test-backend:
	cd backend && go test ./...
test-frontend:
	cd frontend && pnpm test
test-e2e:
	cd frontend && pnpm playwright test

lint: lint-backend lint-frontend
lint-backend:
	cd backend && go vet ./... && golangci-lint run
lint-frontend:
	cd frontend && pnpm lint && pnpm typecheck

fmt:
	cd backend && gofmt -w . && goimports -w .
	cd frontend && pnpm fmt

# the merge gate: lint + tests + codegen sync check
verify: lint test
	cd protocols && bash gen-types.sh
	@git diff --exit-code -- '*.gen.ts' '*.gen.go' 2>/dev/null || \
	  (echo "Generated files out of sync. Run 'make gen' and commit."; exit 1)

logs:
	docker compose logs -f $(SVC)

admin-password:
	@cd backend && go run ./cmd/admin-password
```

- [ ] **Step 4: Create `docker-compose.yml`** (Postgres only for Phase 0)

```yaml
services:
  postgres:
    image: postgres:16-alpine
    environment:
      POSTGRES_USER: ${POSTGRES_USER}
      POSTGRES_PASSWORD: ${POSTGRES_PASSWORD}
      POSTGRES_DB: ${POSTGRES_DB}
    ports:
      - "${POSTGRES_PORT}:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data
      - ./infra/postgres/init.sql:/docker-entrypoint-initdb.d/00-init.sql:ro
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U ${POSTGRES_USER} -d ${POSTGRES_DB}"]
      interval: 5s
      timeout: 5s
      retries: 10

volumes:
  postgres_data:
```

- [ ] **Step 5: Create root `README.md`**

```markdown
# agentic-in-production

Three-tier learning platform for AI agent concepts on cybersecurity (OCSF) data.

See [docs/README.md](docs/README.md) for the learning index, or jump to the
foundational design spec: [docs/superpowers/specs/2026-05-01-agentic-platform-design.md](docs/superpowers/specs/2026-05-01-agentic-platform-design.md).

## Quickstart

    cp .env.example .env
    make install
    make hooks
    make up
    # in another terminal:
    make dev-backend
    # in yet another:
    make dev-frontend
    # browse to http://localhost:5173

## The merge gate

    make verify   # lint + tests + codegen sync check; must pass before commit

## Repo orientation

- [CLAUDE.md](CLAUDE.md) — guidance for AI coding agents
- [docs/dictionary.md](docs/dictionary.md) — shared vocabulary
- [protocols/](protocols/) — wire contracts (frozen, versioned)
- [docs/adr/](docs/adr/) — architecture decisions
```

- [ ] **Step 6: Verify env interpolation**

Run: `cp .env.example .env && docker compose config | grep POSTGRES_USER`
Expected: `      POSTGRES_USER: aip`

- [ ] **Step 7: Commit**

```bash
git add .gitignore .env.example Makefile docker-compose.yml README.md
git commit -m "chore: bootstrap repo skeleton"
```

---

### Task 2: Author 9 ADRs

Authoring ADRs upfront means anyone reading the repo can find the *why* before the *what*. ADR-0003 is intentionally a stub; its content lands when the Phase 0.0.a research spike (Task 8 of this plan) completes.

**Files:** `docs/adr/0001..0009-*.md`

- [ ] **Step 1: Create `docs/adr/0001-three-tier-with-go-gateway.md`**

```markdown
# ADR-0001: Three-tier architecture with a Go gateway

**Status:** Accepted | **Date:** 2026-05-01

## Context
We want frontend stable across changing AI agent platforms, server-side audit of every tool call, centralised auth. Two-tier (browser → platform) puts LLM credentials in the client, has no audit trail, and forces a frontend redeploy whenever the platform changes.

## Decision
Three services: React frontend, Go backend (gateway with adapter layer), external Agent Platform. Frontend never talks to the platform directly.

## Consequences
- Backend owns auth, audit, rate limiting, agent registry, adapter translation.
- Switching platforms = new adapter folder; no frontend code moves.
- One extra hop per request — acceptable for a single-tenant local app.
```

- [ ] **Step 2: Create `docs/adr/0002-mock-first-then-real-platform.md`**

```markdown
# ADR-0002: Mock-first MVP, real platform deferred to Phase 1

**Status:** Accepted | **Date:** 2026-05-01

## Context
Real platforms add hosting, auth credentials, and platform-specific quirks — friction that would slow down validation of the architecture.

## Decision
Phase 0 ships with **only** the Mock adapter (in-process, scripted scenarios). Phase 1 adds the first real platform (GoClaw — see ADR-0009).

## Consequences
- The mock adapter is permanent infrastructure. It runs in CI forever as the conformance baseline.
- Frontend can be built fully against scripted events.
- "Done" for Phase 0 does not require a live LLM.
```

- [ ] **Step 3: Create `docs/adr/0003-wire-protocol-for-agent-event-stream.md`** (stub)

```markdown
# ADR-0003: Wire protocol for agent event stream

**Status:** Proposed (filled in by Phase 0.0.a research spike — Task 8) | **Date:** 2026-05-01

## Context
Need to commit to a wire protocol for the canonical event stream. Candidates: AG-UI (CopilotKit), A2UI, MCP-shaped events, OpenAI Realtime events, vendor-native, custom.

## Decision
**TO BE COMPLETED** by the Phase 0.0.a research spike. Until then v1 follows the AG-UI-aligned shape from spec §4.2: `run_started`, `text_delta`, `tool_call_start`, `tool_call_end`, `state_update`, `error`, `run_finished`.

## Consequences
(filled in by spike output)
```

- [ ] **Step 4: Create `docs/adr/0004-stateful-agent-platform.md`**

```markdown
# ADR-0004: Stateful agent platform; backend stores only audit + session mapping

**Status:** Accepted | **Date:** 2026-05-01

## Context
Real Agent Platforms own conversation history. Duplicating in our backend would cause drift.

## Decision
Platform owns conversation state. Backend stores: `session_id ↔ platform_conversation_id` mapping; its own audit log of every tool call observed; per-platform connection settings. Backend sends only the new user message on each Run.

## Consequences
- Memory (Phase 4) becomes a backend concern that injects synthesised history at Run time, leaving the platform contract unchanged.
- Mock adapter must also be stateful to faithfully simulate real platforms.
```

- [ ] **Step 5: Create `docs/adr/0005-json-schema-not-asyncapi-yet.md`**

```markdown
# ADR-0005: JSON Schema for events (not AsyncAPI yet)

**Status:** Accepted | **Date:** 2026-05-01

## Context
For the streaming event envelope: AsyncAPI vs plain JSON Schema. AsyncAPI has more ceremony than we currently need.

## Decision
v1 of the event envelope is described by JSON Schema 2020-12. OpenAPI 3.1 covers the REST endpoints. Reconsider AsyncAPI when there is a second non-frontend consumer.

## Consequences
- Codegen uses `json-schema-to-typescript` (TS) and `omissis/go-jsonschema` (Go).
- Migrating to AsyncAPI later is a translation, not a re-architecture.
```

- [ ] **Step 6: Create `docs/adr/0006-hand-written-go-handlers.md`**

```markdown
# ADR-0006: Hand-written Go handlers, not generated from OpenAPI

**Status:** Accepted | **Date:** 2026-05-01

## Context
With one developer, fighting `oapi-codegen`'s generated middleware exceeds the duplication cost of writing handlers by hand.

## Decision
OpenAPI is the source of truth for **types only** (request/response models in Go and TS). HTTP handlers are written by hand. Per-route integration tests catch contract drift.

## Consequences
Reconsider when team grows past 2 backend devs or API surface exceeds ~30 routes.
```

- [ ] **Step 7: Create `docs/adr/0007-dictionary-md-as-first-class-doc.md`**

```markdown
# ADR-0007: docs/dictionary.md as a first-class shared-vocabulary doc

**Status:** Accepted | **Date:** 2026-05-01

## Context
Silent vocabulary mismatch (user and AI coding agent meaning different things by "agent", "session", "adapter") is one of the most expensive classes of misunderstanding.

## Decision
`docs/dictionary.md` has both user and AI coding agents as primary audience. Root `CLAUDE.md` requires consulting it. Entries are added the moment a misunderstanding surfaces.

## Consequences
- Maintenance is reactive — entries land at the moment of evidence, not from a guessed-up-front list.
- All AI agent guidance files (CLAUDE.md, AGENTS.md) must point to it.
```

- [ ] **Step 8: Create `docs/adr/0008-ocsf-toy-dataset.md`**

```markdown
# ADR-0008: OCSF toy dataset (4 narrow tables, Hive+Parquet)

**Status:** Accepted (Phase 1 implementation; recorded now for traceability) | **Date:** 2026-05-01

## Context
For the Trino tool-use POC we need a dataset that's realistic enough for real-shaped agent reasoning, small enough to commit-or-regenerate, and aligned with the user's cybersecurity domain.

## Decision
Four OCSF event classes: `authentication` (3002), `process_activity` (1007), `network_activity` (4001), `detection_finding` (2004). ~30 days, ~100k rows total, real OCSF column names (nested struct types preserved). Parquet on local FS via Trino's Hive connector with a local metastore. One synthetic incident woven through (credential stuffing → suspicious PowerShell → outbound exfil).

## Consequences
- Phase 1 brings up Trino + Hive metastore in docker-compose.
- Dataset generator is regenerable, not committed binary.
```

- [ ] **Step 9: Create `docs/adr/0009-phase1-platform-goclaw.md`**

```markdown
# ADR-0009: Phase 1 platform = GoClaw

**Status:** Accepted | **Date:** 2026-05-01

## Context
First non-Mock adapter (Phase 1). Candidate list was Dify, GoClaw, NemoClaw, OpenClaw.

## Decision
GoClaw is the first real Agent Platform integration in Phase 1. Other platforms become later additions (each must pass the Phase 0 conformance suite).

## Consequences
- `backend/internal/adapters/goclaw/` is built in Phase 1.
- Other adapter folders remain placeholders.
- GoClaw-vs-alternatives rationale captured here when implementation lands.
```

- [ ] **Step 10: Commit**

```bash
git add docs/adr/
git commit -m "docs(adr): author ADRs 0001-0009 (0003 stubbed for spike)"
```

---

### Task 3: Bootstrap docs/ — learning index, dictionary

**Files:**
- Create: `docs/README.md`, `docs/dictionary.md`
- Create: `docs/concepts/.gitkeep`, `docs/diagrams/.gitkeep`

- [ ] **Step 1: Create `docs/README.md`** (the learning index)

```markdown
# Learning index

Read this when you've lost the thread. It maps the project's concepts to the
experiments that drove them and the code that implements them.

## Foundational reading
- [design spec](superpowers/specs/2026-05-01-agentic-platform-design.md)
- [dictionary](dictionary.md) — shared vocabulary

## Concepts, in the order they were learned
*(empty until Phase 0 finishes — concept essays are written **after** the experiment ships)*
1. The agent loop — Phase 0 (TODO)
2. Streaming protocols — Phase 0 (TODO)

## Decisions
- [ADR-0001 — Three-tier with Go gateway](adr/0001-three-tier-with-go-gateway.md)
- [ADR-0002 — Mock-first MVP](adr/0002-mock-first-then-real-platform.md)
- [ADR-0003 — Wire protocol](adr/0003-wire-protocol-for-agent-event-stream.md)
- [ADR-0004 — Stateful agent platform](adr/0004-stateful-agent-platform.md)
- [ADR-0005 — JSON Schema (not AsyncAPI yet)](adr/0005-json-schema-not-asyncapi-yet.md)
- [ADR-0006 — Hand-written Go handlers](adr/0006-hand-written-go-handlers.md)
- [ADR-0007 — dictionary.md as first-class doc](adr/0007-dictionary-md-as-first-class-doc.md)
- [ADR-0008 — OCSF toy dataset](adr/0008-ocsf-toy-dataset.md)
- [ADR-0009 — Phase 1 platform = GoClaw](adr/0009-phase1-platform-goclaw.md)

## Concepts → code (lookup table — grows as code lands)
- "How does streaming work?" → `backend/internal/httpapi/messages_handler.go`
- "Where do tool calls get audited?" → `backend/internal/audit/`
- "What's the canonical event shape?" → `protocols/agent-events.schema.json`
```

- [ ] **Step 2: Create `docs/dictionary.md`** with 12 starter entries

```markdown
# Dictionary

> **Equally for the user and AI coding agents.** Always check this file before
> using project-specific vocabulary. Add an entry the moment a misunderstanding
> surfaces — that is when it is most valuable.

## Entry format
\`\`\`markdown
### <Term>
**Aliases:** <other names>
**Definition:** <1–3 sentences>
**Don't confuse with:** <similar terms with different meaning>
**Code:** <pointer>
\`\`\`
---

### Agent
**Aliases:** none — qualify when you use it
**Definition:** Triple-meaning. Always qualify in writing:
1. **AI coding agent** — Claude Code, Cursor, etc., editing this codebase.
2. **Agent Platform** — external system (Mock for MVP, GoClaw for Phase 1) that runs the LLM tool-using loop.
3. **Agent service / loop** — informal; usually means whichever platform is currently active.
**Don't confuse with:** any of the above with each other.

### Agent Platform
**Aliases:** "the platform"
**Definition:** External system that owns the LLM agent loop, tool execution, conversation history, and platform-side authentication. Mock for Phase 0; GoClaw for Phase 1; later: Dify, NemoClaw, OpenClaw, custom.
**Don't confuse with:** AI coding agent.
**Code:** `backend/internal/adapters/<platform>/`

### Adapter
**Aliases:** `AgentPlatformAdapter`, "platform adapter"
**Definition:** Go implementation that translates between our **canonical event envelope** and a specific Agent Platform's native API. One folder per adapter. All adapters must pass the conformance test suite.
**Don't confuse with:** Trino's "connector" (Trino's vocabulary).
**Code:** `backend/internal/adapters/adapter.go`

### Canonical event envelope
**Aliases:** "canonical events", "the event protocol"
**Definition:** Frozen JSON-Schema-defined event shape used between Backend↔Frontend and Adapter↔Backend. v1 events: `run_started`, `text_delta`, `tool_call_start`, `tool_call_end`, `state_update`, `error`, `run_finished`. Adapters translate platform-native streams INTO this.
**Don't confuse with:** the platform's native event format.
**Code:** `protocols/agent-events.schema.json`

### Session
**Aliases:** none
**Definition:** **Our** concept — Postgres row owned by the backend that binds a logged-in user to a specific agent and to a `platform_conversation_id`.
**Don't confuse with:** Conversation (the platform's concept).
**Code:** `backend/internal/sessions/`

### Conversation
**Aliases:** `platform_conversation_id`
**Definition:** **The platform's** concept — owned and managed externally, holds message history, mapped 1:1 to one of our sessions.
**Don't confuse with:** Session (ours).

### Run
**Aliases:** none
**Definition:** One execution of the agent against one user message. Bracketed by `run_started`/`run_finished`. Has a `run_id` that serves as correlation ID frontend→backend→adapter→audit.
**Don't confuse with:** Session (which contains many runs).
**Code:** `protocols/agent-events.schema.json`

### Tool call
**Aliases:** none
**Definition:** One invocation of a registered tool by the agent during a Run. Has a `call_id` linking `tool_call_start` to `tool_call_end`.
**Don't confuse with:** the function definition (a "tool"); a "tool call" is an invocation.
**Code:** `protocols/agent-events.schema.json`

### Experiment
**Aliases:** none
**Definition:** Self-contained slice that demonstrates one agent concept (Trino tool-use, RAG, …). Lives under `agents/<name>/` (config) with adapter code under `backend/internal/adapters/<platform>/`.
**Don't confuse with:** Phase (a roadmap milestone; Phase 1 *delivers* the Trino experiment).

### Phase
**Aliases:** none
**Definition:** Numbered milestone in the project roadmap. Phase 0 = MVP skeleton; Phase 1 = Trino tool-use POC; Phase 2 = RAG; Phase 3 = multi-agent; Phase 4 = memory; Phase 5 = evals.
**Don't confuse with:** Experiment.

### Detection Finding
**Aliases:** "finding", "alert"
**Definition:** OCSF class 2004. Pre-existing alert/finding row in the toy dataset that the agent can pivot from. Not the agent's output.
**Don't confuse with:** the agent's investigation result.
**Code:** `infra/seed-data/` (Phase 1)

### OCSF class
**Aliases:** "OCSF event class"
**Definition:** Specific event type in the Open Cybersecurity Schema Framework, identified by `class_uid` (e.g. 3002 = Authentication).
**Don't confuse with:** an OOP class. Always say "OCSF class" when ambiguity is possible.
```

- [ ] **Step 3: Create empty `docs/concepts/.gitkeep` and `docs/diagrams/.gitkeep`**

- [ ] **Step 4: Commit**

```bash
git add docs/README.md docs/dictionary.md docs/concepts/.gitkeep docs/diagrams/.gitkeep
git commit -m "docs: bootstrap learning index, dictionary, concept/diagram folders"
```

---

### Task 4: Bootstrap CLAUDE.md hierarchy + AGENTS.md + .claude/settings.json

**Files:**
- Create: `CLAUDE.md`, `AGENTS.md`, `.claude/settings.json`
- Create: `backend/CLAUDE.md`, `frontend/CLAUDE.md`, `agents/CLAUDE.md` (placeholders, filled by later tasks)

- [ ] **Step 1: Create root `CLAUDE.md`**

```markdown
# agentic-in-production — guidance for AI coding agents

Three-tier learning platform: React → Go gateway (Adapter pattern) → external Agent Platform.

## Read these before using project vocabulary
- [docs/dictionary.md](docs/dictionary.md) — terms with project-specific meaning. **Always** check before using "agent", "session", "adapter", "experiment", "run".
- [protocols/](protocols/) — wire contracts. Source of truth.

## Where docs live
- [docs/README.md](docs/README.md) — learning index (start here)
- [docs/concepts/](docs/concepts/) — conceptual essays (the "why")
- [docs/adr/](docs/adr/) — architecture decisions (read before changing them)
- [docs/superpowers/specs/](docs/superpowers/specs/) — design specs
- [docs/superpowers/plans/](docs/superpowers/plans/) — implementation plans

## Sub-agent guidance (load in addition to this file)
- Working in `backend/` ? Read [backend/CLAUDE.md](backend/CLAUDE.md).
- Working in `frontend/` ? Read [frontend/CLAUDE.md](frontend/CLAUDE.md).
- Working in `agents/` ? Read [agents/CLAUDE.md](agents/CLAUDE.md) — that folder is platform configuration, not code.

## Hard rules — do NOT do these without asking
- Add a top-level dependency to any service.
- Modify `protocols/` (wire contract) without an accompanying ADR.
- Add a platform adapter that doesn't pass the conformance suite.
- Bypass the backend's audit log.
- Skip the concept essay when finishing an experiment.
- Edit generated files (`*.gen.ts`, `*.gen.go`) by hand — run `make gen`.

## Workflow skills
- Starting any feature → `superpowers:brainstorming`
- Implementing → `superpowers:writing-plans` then `superpowers:executing-plans`
- Bug or unexpected behaviour → `superpowers:systematic-debugging` first
- Before "done" → `superpowers:verification-before-completion`
- Major step done → `superpowers:requesting-code-review`

## Common dev commands
- `make verify` — lint + tests + codegen sync (the merge gate)
- `make up` / `make down` — stack up/down
- `make gen` — regenerate types from `protocols/`
- `make test` — all unit tests
- `make test-e2e` — Playwright e2e
```

- [ ] **Step 2: Create `AGENTS.md`**

```markdown
# AGENTS.md

This project's AI coding agent guidance lives in [CLAUDE.md](CLAUDE.md). Tools that read AGENTS.md (Cursor, Aider, etc.) should treat that file as authoritative.
```

- [ ] **Step 3: Create `.claude/settings.json`**

```json
{
  "permissions": {
    "allow": [
      "Bash(make *)",
      "Bash(go test:*)",
      "Bash(go build:*)",
      "Bash(go vet:*)",
      "Bash(go mod:*)",
      "Bash(gofmt:*)",
      "Bash(goimports:*)",
      "Bash(golangci-lint *)",
      "Bash(pnpm install)",
      "Bash(pnpm run *)",
      "Bash(pnpm test:*)",
      "Bash(pnpm typecheck)",
      "Bash(pnpm lint)",
      "Bash(pnpm fmt)",
      "Bash(pnpm playwright *)",
      "Bash(docker compose ps)",
      "Bash(docker compose logs *)",
      "Bash(docker compose config)"
    ]
  }
}
```

- [ ] **Step 4: Create per-service CLAUDE.md placeholders**

`backend/CLAUDE.md`:
```markdown
# Backend (Go) — guidance for AI coding agents

Filled in by Task 9 of the Phase 0 plan when the Go service is initialised. See root [../CLAUDE.md](../CLAUDE.md) until then.
```

`frontend/CLAUDE.md`:
```markdown
# Frontend (React + TypeScript) — guidance for AI coding agents

Filled in by Task 18 of the Phase 0 plan when the Vite project is initialised. See root [../CLAUDE.md](../CLAUDE.md) until then.
```

`agents/CLAUDE.md`:
```markdown
# agents/ — platform configuration, NOT code

This folder holds **platform-side configuration** for each agent experiment: prompts, tool definitions, exported workflow files. There is no Python/Node code here, no service deps.

When adding an agent experiment:
- Create `agents/<name>/` with: `README.md`, `prompts/`, `tools.yaml`, and any platform-export files (e.g. a GoClaw workflow file).
- The corresponding **adapter code** lives in `backend/internal/adapters/<platform>/`, not here.
- See [../docs/adr/0009-phase1-platform-goclaw.md](../docs/adr/0009-phase1-platform-goclaw.md).
```

- [ ] **Step 5: Commit**

```bash
mkdir -p .claude backend frontend agents
git add CLAUDE.md AGENTS.md .claude/settings.json backend/CLAUDE.md frontend/CLAUDE.md agents/CLAUDE.md
git commit -m "docs: bootstrap CLAUDE.md hierarchy + AGENTS.md + .claude/settings"
```

---

### Task 5: Author the canonical event JSON Schema

This is the most important file in the repo. It defines the wire format that every adapter translates into and every frontend renders. Discriminated union on `type`.

**Files:**
- Create: `protocols/agent-events.schema.json`
- Test: `protocols/examples/*.json` will validate against this in Task 6

- [ ] **Step 1: Create `protocols/agent-events.schema.json`**

```json
{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "$id": "https://agentic-in-production.local/schemas/agent-events.json",
  "title": "Agent Event",
  "description": "Canonical event envelope. Every event on the wire (Backend↔Frontend and Adapter↔Backend) is one of these. v1.",
  "oneOf": [
    { "$ref": "#/$defs/RunStarted" },
    { "$ref": "#/$defs/TextDelta" },
    { "$ref": "#/$defs/ToolCallStart" },
    { "$ref": "#/$defs/ToolCallEnd" },
    { "$ref": "#/$defs/StateUpdate" },
    { "$ref": "#/$defs/Error" },
    { "$ref": "#/$defs/RunFinished" }
  ],
  "$defs": {
    "RunStarted": {
      "type": "object",
      "required": ["type", "run_id"],
      "properties": {
        "type":   { "const": "run_started" },
        "run_id": { "type": "string", "minLength": 1 },
        "metadata": { "$ref": "#/$defs/Metadata" }
      },
      "additionalProperties": false
    },
    "TextDelta": {
      "type": "object",
      "required": ["type", "text"],
      "properties": {
        "type": { "const": "text_delta" },
        "text": { "type": "string" }
      },
      "additionalProperties": false
    },
    "ToolCallStart": {
      "type": "object",
      "required": ["type", "call_id", "tool"],
      "properties": {
        "type":    { "const": "tool_call_start" },
        "call_id": { "type": "string", "minLength": 1 },
        "tool":    { "type": "string", "minLength": 1 },
        "args":    { "type": "object" }
      },
      "additionalProperties": false
    },
    "ToolCallEnd": {
      "type": "object",
      "required": ["type", "call_id", "ok"],
      "properties": {
        "type":           { "const": "tool_call_end" },
        "call_id":        { "type": "string", "minLength": 1 },
        "ok":             { "type": "boolean" },
        "result_preview": { "type": "string" },
        "error_message":  { "type": "string" }
      },
      "additionalProperties": false
    },
    "StateUpdate": {
      "type": "object",
      "required": ["type", "key", "value"],
      "properties": {
        "type":  { "const": "state_update" },
        "key":   { "type": "string", "minLength": 1 },
        "value": {}
      },
      "additionalProperties": false
    },
    "Error": {
      "type": "object",
      "required": ["type", "code", "message"],
      "properties": {
        "type":    { "const": "error" },
        "code":    { "type": "string", "minLength": 1 },
        "message": { "type": "string" }
      },
      "additionalProperties": false
    },
    "RunFinished": {
      "type": "object",
      "required": ["type", "reason"],
      "properties": {
        "type":   { "const": "run_finished" },
        "reason": { "enum": ["done", "stopped", "error"] }
      },
      "additionalProperties": false
    },
    "Metadata": {
      "type": "object",
      "description": "Opaque pass-through. Platform-specific keys go under metadata.platform; never expose to clients with semantic meaning.",
      "properties": {
        "platform": { "type": "object", "additionalProperties": true }
      },
      "additionalProperties": true
    }
  }
}
```

- [ ] **Step 2: Verify it parses as valid JSON Schema**

Run: `python -c "import json,jsonschema; jsonschema.Draft202012Validator.check_schema(json.load(open('protocols/agent-events.schema.json')))"`

Or, if Python+jsonschema isn't installed, defer this verification to Step 4 of Task 6 (where examples validate against it).

- [ ] **Step 3: Commit**

```bash
git add protocols/agent-events.schema.json
git commit -m "feat(protocols): add v1 canonical event JSON Schema"
```

---

### Task 6: Author protocol README, VERSION, agent-protocol.md, examples

**Files:**
- Create: `protocols/VERSION`, `protocols/README.md`, `protocols/agent-protocol.md`
- Create: `protocols/examples/text-only-stream.json`, `tool-call-success.json`, `tool-call-failure.json`, `multi-step-investigation.json`

- [ ] **Step 1: Create `protocols/VERSION`** (single-line file)

```
v1
```

- [ ] **Step 2: Create `protocols/README.md`**

```markdown
# protocols/ — frozen wire contracts

This folder is the single source of truth for the wires between services. **Treat it as authoritative.**

- [VERSION](VERSION) — current major version (currently `v1`)
- [openapi.yaml](openapi.yaml) — REST contract (Frontend ↔ Backend)
- [agent-events.schema.json](agent-events.schema.json) — JSON Schema for the canonical streaming event envelope
- [agent-protocol.md](agent-protocol.md) — human-readable lifecycle, ordering invariants, error semantics
- [examples/](examples/) — canonical event sequences, used as fixtures by frontend and as conformance baselines by backend
- [gen-types.sh](gen-types.sh) — codegen entrypoint (regenerate types in `frontend/src/api/`, `frontend/src/events/`, `backend/internal/protocol/`)

## Versioning policy

| Change | Treatment |
|---|---|
| Add an optional field to an event | additive — no version label change |
| Add a new event type | additive |
| Make a field required, rename, change type, remove a field, remove an event | **breaking — requires v2** |
| Breaking rollout | v1 and v2 coexist on `/api/v1/...` and `/api/v2/...` for at least one phase |

**Every protocol change requires an ADR.**

## Hard rules
- Do not edit generated files (`*.gen.ts`, `*.gen.go`) by hand. Run `make gen`.
- Do not bypass schema validation in adapters or backend.
- Frontend team owns nothing here — backend proposes; ADR consensus changes.
```

- [ ] **Step 3: Create `protocols/agent-protocol.md`** (lifecycle doc)

```markdown
# Canonical event protocol — lifecycle and ordering

Every Run produces a stream of events conforming to [agent-events.schema.json](agent-events.schema.json). The same envelope is used Backend ↔ Frontend and Adapter ↔ Backend.

## Lifecycle

A Run is bracketed by exactly one `run_started` and exactly one `run_finished`. Between them, any number of `text_delta`, `tool_call_start`/`tool_call_end` pairs, `state_update`, and at most one terminal `error` may occur.

```
run_started ── ( text_delta | tool_call_start ... tool_call_end | state_update | error )* ── run_finished
```

## Invariants

1. **Bracketing.** A stream that emits any event before `run_started` or after `run_finished` is malformed. Backend rejects malformed streams (drops them and emits a synthetic `error` + `run_finished{reason:"error"}`).
2. **Tool-call pairing.** Every `tool_call_start{call_id:X}` MUST be followed by exactly one `tool_call_end{call_id:X}` before `run_finished`. Adapters that lose a `tool_call_end` synthesise one with `ok:false`, `error_message:"adapter_lost_end_event"`.
3. **Run termination on error.** A terminal `error` event MUST be followed by `run_finished{reason:"error"}` and then the stream closes.
4. **Run ID stability.** The `run_id` in `run_started` is the correlation ID for the entire run — used in audit logs, frontend keying, and platform debugging.
5. **No interleaving across runs.** A given SSE connection carries exactly one Run. New Run = new connection.

## Reconnection

If the SSE connection drops mid-Run, the client SHOULD treat the Run as abandoned. v1 does not support resumption. Phase 5 (memory) may revisit.

## Error semantics

`error.code` is a stable string identifier (lowercase snake_case) with a fixed set of values:

| Code | Meaning |
|---|---|
| `platform_unavailable` | upstream Agent Platform unreachable |
| `tool_execution_failed` | a tool call raised — payload in `error_message` |
| `auth_failed` | upstream platform rejected backend's API key |
| `internal_error` | unclassified backend or adapter bug |
| `adapter_lost_end_event` | synthesised when a tool_call_end is missing |
| `rate_limited` | upstream rate limit hit |
```

- [ ] **Step 4: Create the four example fixtures**

`protocols/examples/text-only-stream.json`:
```json
[
  { "type": "run_started", "run_id": "run_e1" },
  { "type": "text_delta", "text": "Hello, " },
  { "type": "text_delta", "text": "world!" },
  { "type": "run_finished", "reason": "done" }
]
```

`protocols/examples/tool-call-success.json`:
```json
[
  { "type": "run_started", "run_id": "run_e2" },
  { "type": "text_delta", "text": "Looking at the orders table." },
  { "type": "tool_call_start", "call_id": "c1", "tool": "execute_query", "args": { "sql": "SELECT count(*) FROM orders" } },
  { "type": "tool_call_end", "call_id": "c1", "ok": true, "result_preview": "1 row, 1 column: count=12483" },
  { "type": "text_delta", "text": " There are 12,483 orders." },
  { "type": "run_finished", "reason": "done" }
]
```

`protocols/examples/tool-call-failure.json`:
```json
[
  { "type": "run_started", "run_id": "run_e3" },
  { "type": "tool_call_start", "call_id": "c1", "tool": "execute_query", "args": { "sql": "SELECT * FROM nope" } },
  { "type": "tool_call_end", "call_id": "c1", "ok": false, "error_message": "Table 'nope' does not exist" },
  { "type": "text_delta", "text": "That table doesn't exist; let me check the schema." },
  { "type": "tool_call_start", "call_id": "c2", "tool": "list_tables", "args": {} },
  { "type": "tool_call_end", "call_id": "c2", "ok": true, "result_preview": "tables: orders, customers, products" },
  { "type": "text_delta", "text": " I can see orders, customers, products." },
  { "type": "run_finished", "reason": "done" }
]
```

`protocols/examples/multi-step-investigation.json` (Trino-flavored — used by Mock adapter to simulate a Phase-1-style investigation):
```json
[
  { "type": "run_started", "run_id": "run_e4" },
  { "type": "text_delta", "text": "Investigating the alert on WIN-WS-014." },
  { "type": "state_update", "key": "current_table", "value": "detection_finding" },
  { "type": "tool_call_start", "call_id": "c1", "tool": "describe_table", "args": { "catalog": "ocsf", "schema": "events", "table": "detection_finding" } },
  { "type": "tool_call_end", "call_id": "c1", "ok": true, "result_preview": "columns: time, class_uid, severity_id, finding_info.title, device.hostname" },
  { "type": "tool_call_start", "call_id": "c2", "tool": "execute_query", "args": { "sql": "SELECT * FROM ocsf.events.detection_finding WHERE device.hostname = 'WIN-WS-014' AND time > now() - interval '24' hour" } },
  { "type": "tool_call_end", "call_id": "c2", "ok": true, "result_preview": "1 row: suspicious_powershell_execution at 03:14 UTC" },
  { "type": "state_update", "key": "current_table", "value": "process_activity" },
  { "type": "tool_call_start", "call_id": "c3", "tool": "execute_query", "args": { "sql": "SELECT process.cmd_line, process.parent_process.name FROM ocsf.events.process_activity WHERE device.hostname = 'WIN-WS-014' AND time BETWEEN ..." } },
  { "type": "tool_call_end", "call_id": "c3", "ok": true, "result_preview": "1 row: powershell.exe -EncodedCommand <base64>; parent: explorer.exe" },
  { "type": "text_delta", "text": "Found the suspicious PowerShell — encoded command launched from explorer.exe at 03:14 UTC. Checking the user's auth history next." },
  { "type": "run_finished", "reason": "done" }
]
```

- [ ] **Step 5: Validate every example against the schema**

Add a small Python validation script (one-shot, used here and again in Task 8's `gen-types.sh`):

Run inline:
```bash
python3 -c "
import json, sys
from pathlib import Path
try:
    import jsonschema
except ImportError:
    print('Install: pip install jsonschema'); sys.exit(1)
schema = json.load(open('protocols/agent-events.schema.json'))
v = jsonschema.Draft202012Validator(schema)
ok = True
for p in sorted(Path('protocols/examples').glob('*.json')):
    events = json.load(open(p))
    for i, ev in enumerate(events):
        for err in v.iter_errors(ev):
            print(f'{p.name}[{i}]: {err.message}'); ok = False
print('PASS' if ok else 'FAIL'); sys.exit(0 if ok else 1)
"
```

Expected: `PASS`

- [ ] **Step 6: Commit**

```bash
git add protocols/VERSION protocols/README.md protocols/agent-protocol.md protocols/examples/
git commit -m "feat(protocols): add VERSION, README, lifecycle doc, 4 example fixtures"
```

---

### Task 7: Author the OpenAPI spec

**Files:**
- Create: `protocols/openapi.yaml`

- [ ] **Step 1: Create `protocols/openapi.yaml`**

```yaml
openapi: 3.1.0
info:
  title: agentic-in-production API
  version: "1.0.0"
  description: REST contract between Frontend and Backend. SSE stream for messages.
servers:
  - url: http://localhost:8080
paths:
  /api/healthz:
    get:
      summary: Liveness check
      operationId: healthz
      responses:
        '200':
          description: OK
          content:
            application/json:
              schema: { $ref: '#/components/schemas/Healthz' }

  /api/auth/login:
    post:
      summary: Log in (single admin in MVP)
      operationId: login
      requestBody:
        required: true
        content:
          application/json:
            schema: { $ref: '#/components/schemas/LoginRequest' }
      responses:
        '200':
          description: token issued (also set as HttpOnly cookie)
          content:
            application/json:
              schema: { $ref: '#/components/schemas/LoginResponse' }
        '401':
          description: Bad credentials

  /api/agents:
    get:
      summary: List enabled agents
      operationId: listAgents
      security: [{ cookieAuth: [] }]
      responses:
        '200':
          description: list of agents
          content:
            application/json:
              schema:
                type: array
                items: { $ref: '#/components/schemas/Agent' }

  /api/sessions:
    post:
      summary: Create a session bound to an agent
      operationId: createSession
      security: [{ cookieAuth: [] }]
      requestBody:
        required: true
        content:
          application/json:
            schema: { $ref: '#/components/schemas/CreateSessionRequest' }
      responses:
        '201':
          description: session created
          content:
            application/json:
              schema: { $ref: '#/components/schemas/Session' }

  /api/sessions/{id}/messages:
    get:
      summary: List messages persisted for a session
      operationId: listMessages
      security: [{ cookieAuth: [] }]
      parameters:
        - { name: id, in: path, required: true, schema: { type: string } }
      responses:
        '200':
          description: history
          content:
            application/json:
              schema:
                type: array
                items: { $ref: '#/components/schemas/Message' }
    post:
      summary: Send a user message and stream agent events
      operationId: sendMessage
      security: [{ cookieAuth: [] }]
      parameters:
        - { name: id, in: path, required: true, schema: { type: string } }
      requestBody:
        required: true
        content:
          application/json:
            schema: { $ref: '#/components/schemas/SendMessageRequest' }
      responses:
        '200':
          description: |
            Server-Sent Events stream. Each event's `data` field is a JSON object
            conforming to agent-events.schema.json. Stream ends when a
            `run_finished` event is delivered.
          content:
            text/event-stream:
              schema:
                type: string
                description: see agent-events.schema.json

  /api/sessions/{id}/audit:
    get:
      summary: Audit trail of tool calls for a session
      operationId: getAudit
      security: [{ cookieAuth: [] }]
      parameters:
        - { name: id, in: path, required: true, schema: { type: string } }
      responses:
        '200':
          description: audit entries
          content:
            application/json:
              schema:
                type: array
                items: { $ref: '#/components/schemas/AuditEntry' }

components:
  securitySchemes:
    cookieAuth:
      type: apiKey
      in: cookie
      name: aip_session
  schemas:
    Healthz:
      type: object
      required: [ok]
      properties:
        ok: { type: boolean }
        version: { type: string }
    LoginRequest:
      type: object
      required: [username, password]
      properties:
        username: { type: string }
        password: { type: string }
    LoginResponse:
      type: object
      required: [token]
      properties:
        token: { type: string, description: "JWT; also set as HttpOnly cookie aip_session" }
    Agent:
      type: object
      required: [name, version, enabled]
      properties:
        name: { type: string }
        version: { type: string }
        description: { type: string }
        enabled: { type: boolean }
    CreateSessionRequest:
      type: object
      required: [agent_name]
      properties:
        agent_name: { type: string }
    Session:
      type: object
      required: [id, agent_name, created_at]
      properties:
        id: { type: string }
        agent_name: { type: string }
        created_at: { type: string, format: date-time }
    Message:
      type: object
      required: [id, role, text, created_at]
      properties:
        id: { type: string }
        role: { type: string, enum: [user, assistant] }
        text: { type: string }
        created_at: { type: string, format: date-time }
    SendMessageRequest:
      type: object
      required: [text]
      properties:
        text: { type: string, minLength: 1 }
    AuditEntry:
      type: object
      required: [id, run_id, kind, occurred_at, payload]
      properties:
        id: { type: string }
        run_id: { type: string }
        kind: { type: string, description: "matches event type from agent-events.schema.json" }
        occurred_at: { type: string, format: date-time }
        payload: { type: object, description: "the event JSON" }
```

- [ ] **Step 2: Validate the OpenAPI spec**

Run: `npx @redocly/cli lint protocols/openapi.yaml`
Expected: `Woohoo! Your API description is valid. 🎉` (or zero errors).

- [ ] **Step 3: Commit**

```bash
git add protocols/openapi.yaml
git commit -m "feat(protocols): add OpenAPI 3.1 REST contract"
```

---

### Task 8: Codegen pipeline + Phase 0.0.a research spike

This task does two things atomically because they're entangled: it sets up `make gen` (which generates types from the protocol files) AND it captures the research spike's recommendation in ADR-0003. If the spike's outcome is "stick with the current shape", no schema edits land. If it's "adopt AG-UI directly", the schema is edited and ADR-0003 is updated *before* `make gen` runs for the first time.

**Files:**
- Create: `protocols/gen-types.sh`
- Modify: `docs/adr/0003-wire-protocol-for-agent-event-stream.md`
- Generated (committed): `frontend/src/api/types.gen.ts`, `frontend/src/events/types.gen.ts`, `backend/internal/protocol/api.gen.go`, `backend/internal/protocol/events.gen.go`

- [ ] **Step 1: Conduct the research spike (timeboxed: 1 working day)**

Read the official docs / repos for each candidate and record findings in a scratch buffer (notes, not committed):
- AG-UI (CopilotKit) — github.com/copilotkit/CopilotKit — does the event vocabulary fit our `text_delta`/`tool_call_*`/`state_update` shape? what's the transport assumption? generative-UI integration cost?
- A2UI — find the canonical reference (search "A2UI agent UI protocol"); evaluate maturity, adoption, vendor backing, schema clarity.
- MCP-shaped events — modelcontextprotocol.io — designed for tool/resource exchange, not streaming agent events. Likely a "no" but worth noting why.
- OpenAI Realtime events — platform.openai.com/docs/guides/realtime — vendor-specific but well-specified.
- GoClaw native, Dify SSE — vendor-native shapes (note where they fit / don't fit).

Score each on:
1. Documentation maturity (1–5)
2. Community/vendor adoption (1–5)
3. Event vocabulary fit for our agent loop (1–5)
4. Transport flexibility (1–5)
5. Generative-UI readiness (1–5)
6. Translation distance from our v1 (1–5; lower is better)

- [ ] **Step 2: Write up the decision in ADR-0003**

Replace the placeholder content with concrete prose. Three plausible outcomes (pick one based on the spike):

**Outcome A — keep current AG-UI-aligned custom v1 (most likely):**
> "We adopt our current AG-UI-aligned minimal schema as v1. AG-UI direct adoption is the migration target; translation distance is ~1 (rename a few fields). A2UI requires further investigation before we can evaluate it as a target. Until that investigation lands, we stay translation-ready and avoid platform-specific semantics in the schema."

**Outcome B — adopt AG-UI directly:**
> "We adopt AG-UI's official schema as v1. Our existing schema is replaced with AG-UI's; field renames are applied (e.g. … → …). Trade-off accepted: tighter coupling to CopilotKit's release cadence in exchange for off-the-shelf client libraries and ecosystem alignment."

**Outcome C — adopt A2UI:** *(only if A2UI investigation concludes it's mature and a strict superset of our needs)*

Whichever is chosen, list the score table from Step 1 in the Consequences section so future readers can see what was weighed.

- [ ] **Step 3: If the decision changes the schema, edit `protocols/agent-events.schema.json` BEFORE running codegen**

If Outcome A — no schema change.
If Outcome B/C — apply the field renames and commit them as a separate commit BEFORE Step 4 below: `git commit -m "feat(protocols)!: adopt <protocol> as v1 wire format"`.

- [ ] **Step 4: Author `protocols/gen-types.sh`**

```bash
#!/usr/bin/env bash
# protocols/gen-types.sh
# Regenerates types into frontend/ and backend/ from the protocol files.
# Idempotent. Invoked by `make gen`. CI uses git diff to detect drift.
set -euo pipefail

cd "$(dirname "$0")"
ROOT="$(cd .. && pwd)"

echo "==> validating examples against schema"
python3 - <<'PY'
import json, sys
from pathlib import Path
import jsonschema
schema = json.load(open('agent-events.schema.json'))
jsonschema.Draft202012Validator.check_schema(schema)
v = jsonschema.Draft202012Validator(schema)
ok = True
for p in sorted(Path('examples').glob('*.json')):
    events = json.load(open(p))
    for i, ev in enumerate(events):
        for err in v.iter_errors(ev):
            print(f"{p.name}[{i}]: {err.message}"); ok = False
sys.exit(0 if ok else 1)
PY

echo "==> generating frontend TS event types from JSON Schema"
npx --yes json-schema-to-typescript@15 \
    --input agent-events.schema.json \
    --output "$ROOT/frontend/src/events/types.gen.ts" \
    --bannerComment "// GENERATED by protocols/gen-types.sh — do not edit by hand"

echo "==> generating frontend TS API types from OpenAPI"
npx --yes openapi-typescript@7 \
    openapi.yaml \
    -o "$ROOT/frontend/src/api/types.gen.ts" \
    --root-types

# Add banner to OpenAPI-generated TS (the generator doesn't take one)
{ echo "// GENERATED by protocols/gen-types.sh — do not edit by hand"; cat "$ROOT/frontend/src/api/types.gen.ts"; } > /tmp/_aip.ts && mv /tmp/_aip.ts "$ROOT/frontend/src/api/types.gen.ts"

echo "==> generating Go event structs from JSON Schema"
go run github.com/atombender/go-jsonschema@v0.18.0 \
    -p protocol \
    --tags json \
    --output "$ROOT/backend/internal/protocol/events.gen.go" \
    agent-events.schema.json

echo "==> generating Go API types from OpenAPI (types-only per ADR-0006)"
go run github.com/deepmap/oapi-codegen/v2/cmd/oapi-codegen@v2.4.1 \
    -generate types \
    -package protocol \
    -o "$ROOT/backend/internal/protocol/api.gen.go" \
    openapi.yaml

echo "==> done"
```

- [ ] **Step 5: Make it executable and run it**

```bash
chmod +x protocols/gen-types.sh
make gen
```

Expected: prints validation pass, then four "generating" lines, then "done". Creates four new generated files.

If a CLI is missing: install Python+jsonschema (`pip install jsonschema`), Node 20+ with pnpm/npx, Go 1.22+. The script uses `npx --yes` and `go run github.com/...@version` so no global installs are required.

- [ ] **Step 6: Author `backend/internal/protocol/validate.go`** (the runtime validator)

This file is hand-written (not generated); it wraps the schema validator so backend code can call `protocol.ValidateEvent(b []byte) error` against the real JSON Schema at runtime. It uses the schema as an embedded resource so deployments don't need protocols/ on disk.

```go
// backend/internal/protocol/validate.go
package protocol

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/santhosh-tekuri/jsonschema/v5"
)

//go:embed events.schema.embed.json
var eventsSchemaBytes []byte

var (
	once     sync.Once
	compiled *jsonschema.Schema
	loadErr  error
)

func compileOnce() {
	once.Do(func() {
		c := jsonschema.NewCompiler()
		var v interface{}
		if err := json.Unmarshal(eventsSchemaBytes, &v); err != nil {
			loadErr = fmt.Errorf("embedded schema: %w", err); return
		}
		if err := c.AddResource("events.schema.json", v); err != nil {
			loadErr = err; return
		}
		s, err := c.Compile("events.schema.json")
		if err != nil {
			loadErr = err; return
		}
		compiled = s
	})
}

// ValidateEvent returns nil if b is a valid canonical agent event.
func ValidateEvent(b []byte) error {
	compileOnce()
	if loadErr != nil {
		return loadErr
	}
	var v interface{}
	if err := json.Unmarshal(b, &v); err != nil {
		return fmt.Errorf("event not valid JSON: %w", err)
	}
	return compiled.Validate(v)
}
```

The script must also copy `agent-events.schema.json` to `backend/internal/protocol/events.schema.embed.json` so `go:embed` works. Add this line to `gen-types.sh` near the bottom (before "done"):

```bash
cp agent-events.schema.json "$ROOT/backend/internal/protocol/events.schema.embed.json"
```

Re-run `make gen`.

- [ ] **Step 7: Test the runtime validator** (the first Go test in the repo — establishes the test pattern)

Create `backend/internal/protocol/validate_test.go`:

```go
package protocol

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestValidateEvent_AcceptsAllExamples(t *testing.T) {
	matches, err := filepath.Glob("../../../protocols/examples/*.json")
	require.NoError(t, err)
	require.NotEmpty(t, matches, "no example files found")

	for _, path := range matches {
		path := path
		t.Run(filepath.Base(path), func(t *testing.T) {
			b, err := os.ReadFile(path)
			require.NoError(t, err)

			var events []json.RawMessage
			require.NoError(t, json.Unmarshal(b, &events))
			require.NotEmpty(t, events)

			for i, ev := range events {
				if err := ValidateEvent(ev); err != nil {
					t.Fatalf("event %d failed: %v", i, err)
				}
			}
		})
	}
}

func TestValidateEvent_RejectsMalformed(t *testing.T) {
	cases := map[string]string{
		"unknown_type":       `{"type":"flibbertigibbet"}`,
		"missing_run_id":     `{"type":"run_started"}`,
		"unpaired_call_id":   `{"type":"tool_call_end","ok":true}`,
		"bad_finish_reason":  `{"type":"run_finished","reason":"banana"}`,
		"extra_property":     `{"type":"text_delta","text":"hi","extra":1}`,
	}
	for name, raw := range cases {
		t.Run(name, func(t *testing.T) {
			require.Error(t, ValidateEvent([]byte(raw)))
		})
	}
}
```

Run: `cd backend && go test ./internal/protocol/...`
Expected: PASS for both tests. *(Task 9 sets up the Go module first; if running before Task 9, defer this test run to after Task 9.)*

- [ ] **Step 8: Commit**

```bash
git add protocols/gen-types.sh backend/internal/protocol/
git commit -m "feat(protocols): add codegen pipeline + runtime validator + ADR-0003 outcome"
```

If ADR-0003 was rewritten to reflect spike findings, include it in this commit (or stage it as a separate prior commit before Step 8).

---

### Phase 0.0 checkpoint

At this point the protocol is **frozen**. Tasks 9–17 (backend) and 18–24 (frontend) develop in parallel against this contract. Smoke-test before moving on:

```bash
make verify  # will fail until backend/frontend exist; that's expected.
ls protocols/        # should show README, VERSION, openapi.yaml, schema, agent-protocol.md, examples/, gen-types.sh
ls backend/internal/protocol/  # events.gen.go, api.gen.go, validate.go, events.schema.embed.json
ls frontend/src/api/ frontend/src/events/  # types.gen.ts in each (frontend dir is created by Task 18)
```

---

## Phase 0.1a — Backend track

### Task 9: Initialize Go module + chi router + healthz

**Files:**
- Create: `backend/go.mod`, `backend/cmd/server/main.go`, `backend/internal/httpapi/router.go`, `backend/internal/httpapi/healthz_handler.go`, `backend/.air.toml`, `backend/.golangci.yaml`
- Modify: `backend/CLAUDE.md` (replace placeholder)

- [ ] **Step 1: Initialize the module**

```bash
cd backend
go mod init github.com/phucvd2512/agentic-in-production/backend
go get github.com/go-chi/chi/v5@v5.1.0
go get github.com/jackc/pgx/v5@v5.7.0
go get github.com/santhosh-tekuri/jsonschema/v5@v5.3.1
go get github.com/stretchr/testify@v1.9.0
go get github.com/golang-jwt/jwt/v5@v5.2.1
go get golang.org/x/crypto@v0.27.0
go get gopkg.in/yaml.v3@v3.0.1
```

(Adjust the module path if you prefer a different GitHub username; everything else is unaffected.)

- [ ] **Step 2: Write the failing test for healthz**

`backend/internal/httpapi/healthz_handler_test.go`:
```go
package httpapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestHealthz_ReturnsOK(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/api/healthz", nil)
	w := httptest.NewRecorder()
	HealthzHandler{Version: "test"}.ServeHTTP(w, r)
	require.Equal(t, http.StatusOK, w.Code)
	var body struct {
		Ok      bool   `json:"ok"`
		Version string `json:"version"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	require.True(t, body.Ok)
	require.Equal(t, "test", body.Version)
}
```

- [ ] **Step 3: Run test — expect FAIL**

```bash
cd backend && go test ./internal/httpapi/...
```
Expected: build error (`HealthzHandler` undefined).

- [ ] **Step 4: Implement `healthz_handler.go`**

```go
package httpapi

import (
	"encoding/json"
	"net/http"
)

type HealthzHandler struct {
	Version string
}

func (h HealthzHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"ok":      true,
		"version": h.Version,
	})
}
```

- [ ] **Step 5: Run test — expect PASS**

`go test ./internal/httpapi/...` → PASS.

- [ ] **Step 6: Author the router** (`backend/internal/httpapi/router.go`)

```go
package httpapi

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type Deps struct {
	Version string
	// More deps wired in later tasks (sessions store, adapter registry, audit, auth)
}

func NewRouter(d Deps) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)
	r.Get("/api/healthz", HealthzHandler{Version: d.Version}.ServeHTTP)
	return r
}
```

- [ ] **Step 7: Author `backend/cmd/server/main.go`**

```go
package main

import (
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/phucvd2512/agentic-in-production/backend/internal/httpapi"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	port := os.Getenv("BACKEND_PORT")
	if port == "" {
		port = "8080"
	}

	srv := &http.Server{
		Addr:              ":" + port,
		Handler:           httpapi.NewRouter(httpapi.Deps{Version: "0.0.0-dev"}),
		ReadHeaderTimeout: 5 * time.Second,
	}
	slog.Info("listening", "addr", srv.Addr)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		slog.Error("server", "err", err); os.Exit(1)
	}
}
```

- [ ] **Step 8: Author `backend/.air.toml`** (hot reload config)

```toml
root = "."
tmp_dir = "tmp"

[build]
bin = "tmp/server"
cmd = "go build -o ./tmp/server ./cmd/server"
delay = 500
exclude_dir = ["tmp", "vendor"]
include_ext = ["go"]
log = "build.log"

[log]
time = false
```

- [ ] **Step 9: Author `backend/.golangci.yaml`**

```yaml
run:
  timeout: 3m
linters:
  enable:
    - errcheck
    - gosimple
    - govet
    - ineffassign
    - staticcheck
    - unused
    - gofmt
    - goimports
```

- [ ] **Step 10: Update `backend/CLAUDE.md`**

```markdown
# Backend (Go) — guidance for AI coding agents

Go gateway service. Receives requests from the React frontend, calls into an
**Adapter** (per-platform Go implementation under `internal/adapters/`), and
streams canonical events back via SSE.

## Conventions
- One package per concern. `internal/auth/`, `internal/sessions/`, `internal/audit/`, etc.
- Hand-written HTTP handlers (see ADR-0006). OpenAPI is types-only.
- All adapter outputs go through `internal/protocol.ValidateEvent` in dev/test.
- Errors: return `error` from internal funcs; only HTTP handlers convert to status codes.
- Logging: `slog`. Always include `run_id` when in a Run context.

## Hot reload
`make dev-backend` runs `air` which rebuilds on file save.

## Tests
- Unit tests next to the file: `foo.go` + `foo_test.go`.
- testify/require for assertions; testify/mock kept minimal — prefer real types where possible.
- The conformance suite (`internal/adapters/conformance/`) is the load-bearing test for adapters.

## When making changes
- Touching `internal/protocol/*.gen.go`? Don't — run `make gen`.
- Touching `internal/adapters/adapter.go` (the interface)? Update every adapter; conformance must still pass.
- Adding an HTTP route? Update `protocols/openapi.yaml` first, run `make gen`, then implement the handler.
```

- [ ] **Step 11: Manual smoke test**

```bash
cd backend && go run ./cmd/server &
sleep 1
curl -s http://localhost:8080/api/healthz
kill %1
```
Expected: `{"ok":true,"version":"0.0.0-dev"}`

- [ ] **Step 12: Commit**

```bash
git add backend/
git commit -m "feat(backend): init Go module, chi router, /api/healthz, air config"
```

---

### Task 10: Postgres init.sql + connection pool

**Files:**
- Create: `infra/postgres/init.sql`, `backend/internal/db/db.go`, `backend/internal/db/db_test.go`

- [ ] **Step 1: Author `infra/postgres/init.sql`**

```sql
-- All Phase 0 tables. Created by docker-entrypoint on first container start.

CREATE TABLE IF NOT EXISTS sessions (
    id                       text         PRIMARY KEY,
    user_id                  text         NOT NULL,
    agent_name               text         NOT NULL,
    platform_conversation_id text,                                 -- nullable: Mock has no remote conv
    created_at               timestamptz  NOT NULL DEFAULT now(),
    ended_at                 timestamptz
);

CREATE INDEX IF NOT EXISTS idx_sessions_user ON sessions (user_id, created_at DESC);

CREATE TABLE IF NOT EXISTS messages (
    id          text         PRIMARY KEY,
    session_id  text         NOT NULL REFERENCES sessions(id) ON DELETE CASCADE,
    role        text         NOT NULL CHECK (role IN ('user','assistant')),
    text        text         NOT NULL,
    created_at  timestamptz  NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_messages_session ON messages (session_id, created_at);

CREATE TABLE IF NOT EXISTS audit_log (
    id           bigserial    PRIMARY KEY,
    session_id   text         NOT NULL REFERENCES sessions(id) ON DELETE CASCADE,
    run_id       text         NOT NULL,
    kind         text         NOT NULL,                    -- canonical event type
    occurred_at  timestamptz  NOT NULL DEFAULT now(),
    payload      jsonb        NOT NULL                     -- the event JSON
);

CREATE INDEX IF NOT EXISTS idx_audit_session ON audit_log (session_id, occurred_at);
CREATE INDEX IF NOT EXISTS idx_audit_run     ON audit_log (run_id);

CREATE TABLE IF NOT EXISTS agent_registry (
    name         text   PRIMARY KEY,
    version      text   NOT NULL,
    description  text,
    enabled      boolean NOT NULL DEFAULT true,
    adapter_kind text   NOT NULL,             -- e.g. 'mock', 'goclaw'
    config       jsonb  NOT NULL DEFAULT '{}'  -- per-adapter settings (base_url, api_key_ref, ...)
);

-- Phase 0 seed: one Mock-backed agent
INSERT INTO agent_registry (name, version, description, enabled, adapter_kind, config) VALUES
    ('mock-trino-flavored', '0.1.0', 'Scripted Trino-flavored conversations for MVP smoke', true, 'mock', '{}')
ON CONFLICT (name) DO NOTHING;
```

- [ ] **Step 2: Bring up Postgres and verify tables exist**

```bash
docker compose up -d postgres
sleep 3
docker compose exec postgres psql -U aip -d aip -c "\dt"
```
Expected: lists `agent_registry`, `audit_log`, `messages`, `sessions`.

- [ ] **Step 3: Author `backend/internal/db/db.go`** (thin pgxpool wrapper)

```go
package db

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type DB struct {
	Pool *pgxpool.Pool
}

func Open(ctx context.Context) (*DB, error) {
	user := getenv("POSTGRES_USER", "aip")
	pass := getenv("POSTGRES_PASSWORD", "aip")
	name := getenv("POSTGRES_DB", "aip")
	host := getenv("POSTGRES_HOST", "localhost")
	port := getenv("POSTGRES_PORT", "5432")
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", user, pass, host, port, name)

	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, err
	}
	cfg.MaxConns = 8
	cfg.MaxConnIdleTime = 30 * time.Second

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, err
	}
	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("ping: %w", err)
	}
	return &DB{Pool: pool}, nil
}

func (d *DB) Close() { d.Pool.Close() }

func getenv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
```

- [ ] **Step 4: Write a smoke test that depends on a live Postgres**

`backend/internal/db/db_test.go`:
```go
package db

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// Skip when Postgres isn't reachable so unit tests still run on a bare workstation.
func TestOpen_PingsLivePostgres(t *testing.T) {
	if os.Getenv("AIP_SKIP_DB_TESTS") == "1" {
		t.Skip("AIP_SKIP_DB_TESTS=1")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	d, err := Open(ctx)
	if err != nil {
		t.Skipf("postgres not reachable; skipping (%v)", err)
	}
	defer d.Close()
	require.NoError(t, d.Pool.Ping(ctx))
}
```

- [ ] **Step 5: Run the test against the running Postgres**

```bash
cd backend && go test ./internal/db/...
```
Expected: PASS (or `Skipped` if Postgres isn't up — both are acceptable; CI runs with Postgres up).

- [ ] **Step 6: Wire `db.Open` into `cmd/server/main.go`**

Replace the body of `main()` with:

```go
func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	ctx := context.Background()
	d, err := db.Open(ctx)
	if err != nil {
		slog.Error("db.Open", "err", err); os.Exit(1)
	}
	defer d.Close()

	port := os.Getenv("BACKEND_PORT")
	if port == "" {
		port = "8080"
	}
	srv := &http.Server{
		Addr:              ":" + port,
		Handler:           httpapi.NewRouter(httpapi.Deps{Version: "0.0.0-dev", DB: d}),
		ReadHeaderTimeout: 5 * time.Second,
	}
	slog.Info("listening", "addr", srv.Addr)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		slog.Error("server", "err", err); os.Exit(1)
	}
}
```

Add `DB *db.DB` to `httpapi.Deps`. Add the `context` and `db` imports.

- [ ] **Step 7: Commit**

```bash
git add infra/postgres/ backend/internal/db/ backend/cmd/server/main.go backend/internal/httpapi/router.go
git commit -m "feat(backend): add Postgres init.sql + pgx pool, wire into server"
```

---

### Task 11: Auth — JWT cookie + admin-password CLI

**Files:**
- Create: `backend/internal/auth/auth.go`, `backend/internal/auth/auth_test.go`
- Create: `backend/internal/httpapi/auth_handler.go`, `..._test.go`
- Create: `backend/cmd/admin-password/main.go`

- [ ] **Step 1: Write failing tests for `auth` package**

`backend/internal/auth/auth_test.go`:
```go
package auth

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestVerifyPassword_AcceptsCorrect(t *testing.T) {
	hash, err := HashPassword("hunter2")
	require.NoError(t, err)
	require.True(t, VerifyPassword(hash, "hunter2"))
	require.False(t, VerifyPassword(hash, "wrong"))
}

func TestIssueAndParseToken_RoundTrips(t *testing.T) {
	a := NewAuthenticator([]byte("0123456789abcdef0123456789abcdef"), "admin", "ignored", 1*time.Hour)
	tok, err := a.IssueToken("admin")
	require.NoError(t, err)
	claims, err := a.ParseToken(tok)
	require.NoError(t, err)
	require.Equal(t, "admin", claims.Subject)
}

func TestParseToken_RejectsTampered(t *testing.T) {
	a := NewAuthenticator([]byte("0123456789abcdef0123456789abcdef"), "admin", "ignored", 1*time.Hour)
	tok, _ := a.IssueToken("admin")
	_, err := a.ParseToken(tok + "x")
	require.Error(t, err)
}
```

Run: `go test ./internal/auth/...` → expect FAIL (build).

- [ ] **Step 2: Implement `auth.go`**

```go
package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type Claims struct {
	Subject string `json:"sub"`
	jwt.RegisteredClaims
}

type Authenticator struct {
	signingKey   []byte
	adminUser    string
	adminPwHash  string
	tokenTTL     time.Duration
}

func NewAuthenticator(key []byte, adminUser, adminPwHash string, ttl time.Duration) *Authenticator {
	return &Authenticator{signingKey: key, adminUser: adminUser, adminPwHash: adminPwHash, tokenTTL: ttl}
}

func HashPassword(plain string) (string, error) {
	b, err := bcrypt.GenerateFromPassword([]byte(plain), bcrypt.DefaultCost)
	return string(b), err
}

func VerifyPassword(hash, plain string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(plain)) == nil
}

func (a *Authenticator) Login(user, pw string) (string, error) {
	if user != a.adminUser || !VerifyPassword(a.adminPwHash, pw) {
		return "", errors.New("bad credentials")
	}
	return a.IssueToken(user)
}

func (a *Authenticator) IssueToken(sub string) (string, error) {
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		Subject: sub,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(a.tokenTTL)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	})
	return t.SignedString(a.signingKey)
}

func (a *Authenticator) ParseToken(raw string) (*Claims, error) {
	tok, err := jwt.ParseWithClaims(raw, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return a.signingKey, nil
	})
	if err != nil {
		return nil, err
	}
	c, ok := tok.Claims.(*Claims)
	if !ok || !tok.Valid {
		return nil, errors.New("invalid token")
	}
	return c, nil
}
```

Run: `go test ./internal/auth/...` → PASS.

- [ ] **Step 3: Build the auth HTTP handler + middleware**

`backend/internal/httpapi/auth_handler.go`:
```go
package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/phucvd2512/agentic-in-production/backend/internal/auth"
)

const cookieName = "aip_session"

type AuthHandler struct {
	Auth *auth.Authenticator
}

func (h AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Username, Password string
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest); return
	}
	tok, err := h.Auth.Login(req.Username, req.Password)
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized); return
	}
	http.SetCookie(w, &http.Cookie{
		Name:     cookieName,
		Value:    tok,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Expires:  time.Now().Add(time.Hour),
	})
	w.Header().Set("content-type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"token": tok})
}

type ctxKey int
const userKey ctxKey = 1

func RequireAuth(a *auth.Authenticator) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			c, err := r.Cookie(cookieName)
			if err != nil {
				http.Error(w, "unauthorized", http.StatusUnauthorized); return
			}
			claims, err := a.ParseToken(c.Value)
			if err != nil {
				http.Error(w, "unauthorized", http.StatusUnauthorized); return
			}
			ctx := context.WithValue(r.Context(), userKey, claims.Subject)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func UserFromCtx(ctx context.Context) string {
	v, _ := ctx.Value(userKey).(string)
	return v
}
```

- [ ] **Step 4: Test the login endpoint**

`backend/internal/httpapi/auth_handler_test.go`:
```go
package httpapi

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/phucvd2512/agentic-in-production/backend/internal/auth"
	"github.com/stretchr/testify/require"
)

func TestLogin_GoodCredentials_SetsCookie(t *testing.T) {
	hash, _ := auth.HashPassword("hunter2")
	a := auth.NewAuthenticator([]byte("0123456789abcdef0123456789abcdef"), "admin", hash, time.Hour)
	h := AuthHandler{Auth: a}

	req := httptest.NewRequest(http.MethodPost, "/api/auth/login",
		bytes.NewReader([]byte(`{"username":"admin","password":"hunter2"}`)))
	w := httptest.NewRecorder()
	h.Login(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	cookies := w.Result().Cookies()
	require.Len(t, cookies, 1)
	require.Equal(t, "aip_session", cookies[0].Name)
}

func TestLogin_BadCredentials_401(t *testing.T) {
	hash, _ := auth.HashPassword("hunter2")
	a := auth.NewAuthenticator([]byte("0123456789abcdef0123456789abcdef"), "admin", hash, time.Hour)
	h := AuthHandler{Auth: a}

	req := httptest.NewRequest(http.MethodPost, "/api/auth/login",
		bytes.NewReader([]byte(`{"username":"admin","password":"wrong"}`)))
	w := httptest.NewRecorder()
	h.Login(w, req)
	require.Equal(t, http.StatusUnauthorized, w.Code)
}
```

Run: `go test ./internal/httpapi/...` → PASS.

- [ ] **Step 5: Author `backend/cmd/admin-password/main.go`** (helper to bcrypt the admin password and rewrite `.env`)

```go
// admin-password reads ADMIN_PASSWORD from stdin and writes its bcrypt hash to .env.
package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/phucvd2512/agentic-in-production/backend/internal/auth"
)

func main() {
	fmt.Print("admin password (will not echo to .env literally; we hash it): ")
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	pw := strings.TrimSpace(scanner.Text())
	if pw == "" {
		fmt.Fprintln(os.Stderr, "empty password"); os.Exit(1)
	}
	hash, err := auth.HashPassword(pw)
	if err != nil {
		fmt.Fprintln(os.Stderr, err); os.Exit(1)
	}
	envPath := "../.env" // run from backend/
	updateEnv(envPath, "ADMIN_PASSWORD_HASH", hash)
	fmt.Println("OK — ADMIN_PASSWORD_HASH written to", envPath)
}

func updateEnv(path, key, val string) {
	b, _ := os.ReadFile(path)
	lines := strings.Split(string(b), "\n")
	found := false
	for i, ln := range lines {
		if strings.HasPrefix(ln, key+"=") {
			lines[i] = key + "=" + val
			found = true; break
		}
	}
	if !found {
		lines = append(lines, key+"="+val)
	}
	_ = os.WriteFile(path, []byte(strings.Join(lines, "\n")), 0644)
}
```

- [ ] **Step 6: Wire login route into the router**

In `router.go`, extend `Deps` with `Auth *auth.Authenticator` and inside `NewRouter`:
```go
r.Post("/api/auth/login", AuthHandler{Auth: d.Auth}.Login)
```

In `main.go`, construct `auth.NewAuthenticator` from env vars (`JWT_SIGNING_KEY`, `ADMIN_USERNAME`, `ADMIN_PASSWORD_HASH`) with a 1-hour TTL, pass it into `Deps`.

- [ ] **Step 7: Manual smoke test**

```bash
make admin-password   # set the hash
cd backend && go run ./cmd/server &
sleep 1
curl -s -i -X POST http://localhost:8080/api/auth/login \
  -H 'content-type: application/json' \
  -d '{"username":"admin","password":"hunter2"}'
kill %1
```
Expected: `HTTP/1.1 200 OK`, `Set-Cookie: aip_session=...`, body `{"token":"..."}`.

- [ ] **Step 8: Commit**

```bash
git add backend/internal/auth/ backend/internal/httpapi/auth_handler*.go backend/cmd/admin-password/ backend/internal/httpapi/router.go backend/cmd/server/main.go
git commit -m "feat(backend): JWT cookie auth + login endpoint + admin-password helper"
```

---

### Task 12: Sessions store

**Files:**
- Create: `backend/internal/sessions/store.go`, `..._test.go`

- [ ] **Step 1: Write failing tests** (require live Postgres; skipped if absent)

`backend/internal/sessions/store_test.go`:
```go
package sessions

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/phucvd2512/agentic-in-production/backend/internal/db"
	"github.com/stretchr/testify/require"
)

func openDB(t *testing.T) *db.DB {
	t.Helper()
	if os.Getenv("AIP_SKIP_DB_TESTS") == "1" {
		t.Skip("AIP_SKIP_DB_TESTS=1")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	d, err := db.Open(ctx)
	if err != nil {
		t.Skipf("postgres not reachable (%v)", err)
	}
	t.Cleanup(d.Close)
	return d
}

func TestCreateAndGet(t *testing.T) {
	d := openDB(t)
	ctx := context.Background()
	s := NewStore(d)

	created, err := s.Create(ctx, "admin", "mock-trino-flavored")
	require.NoError(t, err)
	require.NotEmpty(t, created.ID)

	got, err := s.Get(ctx, created.ID)
	require.NoError(t, err)
	require.Equal(t, created.ID, got.ID)
	require.Equal(t, "mock-trino-flavored", got.AgentName)
}

func TestSetPlatformConvID(t *testing.T) {
	d := openDB(t)
	ctx := context.Background()
	s := NewStore(d)

	created, _ := s.Create(ctx, "admin", "mock-trino-flavored")
	require.NoError(t, s.SetPlatformConvID(ctx, created.ID, "conv_xyz"))
	got, _ := s.Get(ctx, created.ID)
	require.NotNil(t, got.PlatformConvID)
	require.Equal(t, "conv_xyz", *got.PlatformConvID)
}
```

Run: `go test ./internal/sessions/...` → expect FAIL (build).

- [ ] **Step 2: Implement `store.go`**

```go
package sessions

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"time"

	"github.com/phucvd2512/agentic-in-production/backend/internal/db"
)

type Session struct {
	ID             string
	UserID         string
	AgentName      string
	PlatformConvID *string
	CreatedAt      time.Time
	EndedAt        *time.Time
}

type Store struct{ DB *db.DB }

func NewStore(d *db.DB) *Store { return &Store{DB: d} }

func newID(prefix string) string {
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	return prefix + "_" + hex.EncodeToString(b)
}

func (s *Store) Create(ctx context.Context, user, agentName string) (Session, error) {
	id := newID("sess")
	row := s.DB.Pool.QueryRow(ctx,
		`INSERT INTO sessions (id, user_id, agent_name) VALUES ($1, $2, $3)
		 RETURNING id, user_id, agent_name, platform_conversation_id, created_at, ended_at`,
		id, user, agentName)
	var out Session
	err := row.Scan(&out.ID, &out.UserID, &out.AgentName, &out.PlatformConvID, &out.CreatedAt, &out.EndedAt)
	return out, err
}

func (s *Store) Get(ctx context.Context, id string) (Session, error) {
	row := s.DB.Pool.QueryRow(ctx,
		`SELECT id, user_id, agent_name, platform_conversation_id, created_at, ended_at
		 FROM sessions WHERE id = $1`, id)
	var out Session
	err := row.Scan(&out.ID, &out.UserID, &out.AgentName, &out.PlatformConvID, &out.CreatedAt, &out.EndedAt)
	return out, err
}

func (s *Store) SetPlatformConvID(ctx context.Context, id, convID string) error {
	_, err := s.DB.Pool.Exec(ctx,
		`UPDATE sessions SET platform_conversation_id = $1 WHERE id = $2`, convID, id)
	return err
}

func (s *Store) ListMessages(ctx context.Context, sessionID string) ([]Message, error) {
	rows, err := s.DB.Pool.Query(ctx,
		`SELECT id, role, text, created_at FROM messages WHERE session_id=$1 ORDER BY created_at`,
		sessionID)
	if err != nil { return nil, err }
	defer rows.Close()
	var out []Message
	for rows.Next() {
		var m Message
		if err := rows.Scan(&m.ID, &m.Role, &m.Text, &m.CreatedAt); err != nil { return nil, err }
		m.SessionID = sessionID
		out = append(out, m)
	}
	return out, rows.Err()
}

func (s *Store) AppendMessage(ctx context.Context, sessionID, role, text string) (Message, error) {
	id := newID("msg")
	row := s.DB.Pool.QueryRow(ctx,
		`INSERT INTO messages (id, session_id, role, text) VALUES ($1, $2, $3, $4)
		 RETURNING id, session_id, role, text, created_at`, id, sessionID, role, text)
	var m Message
	err := row.Scan(&m.ID, &m.SessionID, &m.Role, &m.Text, &m.CreatedAt)
	return m, err
}

type Message struct {
	ID        string
	SessionID string
	Role      string // "user" | "assistant"
	Text      string
	CreatedAt time.Time
}
```

Run tests → PASS.

- [ ] **Step 3: Commit**

```bash
git add backend/internal/sessions/
git commit -m "feat(backend): sessions/messages store backed by Postgres"
```

---

### Task 13: Audit log store

**Files:**
- Create: `backend/internal/audit/audit.go`, `..._test.go`

- [ ] **Step 1: Write failing test**

`backend/internal/audit/audit_test.go`:
```go
package audit

import (
	"context"
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/phucvd2512/agentic-in-production/backend/internal/db"
	"github.com/phucvd2512/agentic-in-production/backend/internal/sessions"
	"github.com/stretchr/testify/require"
)

func TestAppendAndList(t *testing.T) {
	if os.Getenv("AIP_SKIP_DB_TESTS") == "1" { t.Skip() }
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	d, err := db.Open(ctx)
	if err != nil { t.Skipf("postgres not reachable (%v)", err) }
	defer d.Close()

	sess, err := sessions.NewStore(d).Create(context.Background(), "admin", "mock-trino-flavored")
	require.NoError(t, err)

	a := NewLog(d)
	payload, _ := json.Marshal(map[string]any{"type": "tool_call_start", "call_id": "c1", "tool": "execute_query", "args": map[string]any{"sql": "select 1"}})
	require.NoError(t, a.Append(ctx, sess.ID, "run_1", "tool_call_start", payload))

	entries, err := a.List(ctx, sess.ID)
	require.NoError(t, err)
	require.Len(t, entries, 1)
	require.Equal(t, "run_1", entries[0].RunID)
	require.Equal(t, "tool_call_start", entries[0].Kind)
}
```

- [ ] **Step 2: Implement `audit.go`**

```go
package audit

import (
	"context"
	"encoding/json"
	"time"

	"github.com/phucvd2512/agentic-in-production/backend/internal/db"
)

type Entry struct {
	ID         int64           `json:"id"`
	SessionID  string          `json:"session_id"`
	RunID      string          `json:"run_id"`
	Kind       string          `json:"kind"`
	OccurredAt time.Time       `json:"occurred_at"`
	Payload    json.RawMessage `json:"payload"`
}

type Log struct{ DB *db.DB }

func NewLog(d *db.DB) *Log { return &Log{DB: d} }

func (l *Log) Append(ctx context.Context, sessionID, runID, kind string, payload json.RawMessage) error {
	_, err := l.DB.Pool.Exec(ctx,
		`INSERT INTO audit_log (session_id, run_id, kind, payload) VALUES ($1, $2, $3, $4)`,
		sessionID, runID, kind, payload)
	return err
}

func (l *Log) List(ctx context.Context, sessionID string) ([]Entry, error) {
	rows, err := l.DB.Pool.Query(ctx,
		`SELECT id, session_id, run_id, kind, occurred_at, payload
		 FROM audit_log WHERE session_id=$1 ORDER BY occurred_at, id`, sessionID)
	if err != nil { return nil, err }
	defer rows.Close()
	var out []Entry
	for rows.Next() {
		var e Entry
		if err := rows.Scan(&e.ID, &e.SessionID, &e.RunID, &e.Kind, &e.OccurredAt, &e.Payload); err != nil { return nil, err }
		out = append(out, e)
	}
	return out, rows.Err()
}
```

Run tests → PASS.

- [ ] **Step 3: Commit**

```bash
git add backend/internal/audit/
git commit -m "feat(backend): audit log store"
```

---

### Task 14: Adapter interface + registry + agent registry store

**Files:**
- Create: `backend/internal/adapters/adapter.go`, `..._test.go`
- Create: `backend/internal/adapters/registry.go`
- Create: `backend/internal/agentregistry/store.go`, `..._test.go`

- [ ] **Step 1: Author the adapter interface and canonical event type**

`backend/internal/adapters/adapter.go`:
```go
package adapters

import (
	"context"
	"encoding/json"
)

// AgentEvent is one canonical event yielded by an adapter. Payload is the raw
// JSON object conforming to protocols/agent-events.schema.json.
type AgentEvent struct {
	Type    string          // "run_started" | "text_delta" | ...
	RunID   string          // copied from run_started; available for downstream correlation
	Payload json.RawMessage // the full event JSON, schema-valid
}

type Capabilities struct {
	Name    string   `json:"name"`
	Model   string   `json:"model,omitempty"`
	Tools   []string `json:"tools,omitempty"`
}

type Session struct {
	ID        string
	UserID    string
	AgentName string
}

type RunRequest struct {
	RunID            string
	Session          Session
	PlatformConvID   string // empty for first message; adapters may create one
	UserMessage      string
	ToolAllowlist    []string // backend-enforced; adapter must honor
}

type AgentPlatformAdapter interface {
	Name() string
	Capabilities(ctx context.Context) (Capabilities, error)

	// StartConversation may be a no-op (Mock) or a remote create call (real platforms).
	StartConversation(ctx context.Context, sess Session) (platformConvID string, err error)
	EndConversation(ctx context.Context, platformConvID string) error

	// Run yields canonical AgentEvents on the returned channel and closes it on completion.
	// The channel must NOT be returned in an error state; errors are emitted as Error events
	// followed by RunFinished{reason:"error"}.
	Run(ctx context.Context, req RunRequest) (<-chan AgentEvent, error)
}
```

- [ ] **Step 2: Author the in-process adapter registry**

`backend/internal/adapters/registry.go`:
```go
package adapters

import (
	"errors"
	"sync"
)

// Factory builds an adapter from per-instance configuration JSON.
type Factory func(configJSON []byte) (AgentPlatformAdapter, error)

var (
	regMu      sync.RWMutex
	factories  = map[string]Factory{} // keyed by adapter_kind, e.g. "mock"
)

func Register(kind string, f Factory) {
	regMu.Lock(); defer regMu.Unlock()
	factories[kind] = f
}

func Build(kind string, configJSON []byte) (AgentPlatformAdapter, error) {
	regMu.RLock(); defer regMu.RUnlock()
	f, ok := factories[kind]
	if !ok {
		return nil, errors.New("unknown adapter kind: " + kind)
	}
	return f(configJSON)
}
```

- [ ] **Step 3: Author the agent_registry store**

`backend/internal/agentregistry/store.go`:
```go
package agentregistry

import (
	"context"
	"encoding/json"

	"github.com/phucvd2512/agentic-in-production/backend/internal/db"
)

type Agent struct {
	Name        string          `json:"name"`
	Version     string          `json:"version"`
	Description string          `json:"description"`
	Enabled     bool            `json:"enabled"`
	AdapterKind string          `json:"adapter_kind"`
	Config      json.RawMessage `json:"config"`
}

type Store struct{ DB *db.DB }

func NewStore(d *db.DB) *Store { return &Store{DB: d} }

func (s *Store) ListEnabled(ctx context.Context) ([]Agent, error) {
	rows, err := s.DB.Pool.Query(ctx,
		`SELECT name, version, description, enabled, adapter_kind, config
		 FROM agent_registry WHERE enabled = true ORDER BY name`)
	if err != nil { return nil, err }
	defer rows.Close()
	var out []Agent
	for rows.Next() {
		var a Agent
		if err := rows.Scan(&a.Name, &a.Version, &a.Description, &a.Enabled, &a.AdapterKind, &a.Config); err != nil { return nil, err }
		out = append(out, a)
	}
	return out, rows.Err()
}

func (s *Store) GetByName(ctx context.Context, name string) (Agent, error) {
	row := s.DB.Pool.QueryRow(ctx,
		`SELECT name, version, description, enabled, adapter_kind, config
		 FROM agent_registry WHERE name = $1`, name)
	var a Agent
	err := row.Scan(&a.Name, &a.Version, &a.Description, &a.Enabled, &a.AdapterKind, &a.Config)
	return a, err
}
```

- [ ] **Step 4: Smoke test the registry against seeded Postgres**

`backend/internal/agentregistry/store_test.go`:
```go
package agentregistry

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/phucvd2512/agentic-in-production/backend/internal/db"
	"github.com/stretchr/testify/require"
)

func TestListEnabled_FindsSeed(t *testing.T) {
	if os.Getenv("AIP_SKIP_DB_TESTS") == "1" { t.Skip() }
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	d, err := db.Open(ctx)
	if err != nil { t.Skipf("postgres not reachable (%v)", err) }
	defer d.Close()

	agents, err := NewStore(d).ListEnabled(ctx)
	require.NoError(t, err)
	require.NotEmpty(t, agents)

	var found bool
	for _, a := range agents {
		if a.Name == "mock-trino-flavored" {
			found = true; require.Equal(t, "mock", a.AdapterKind); break
		}
	}
	require.True(t, found, "expected mock-trino-flavored agent in registry")
}
```

Run: `go test ./internal/agentregistry/...` → PASS.

- [ ] **Step 5: Commit**

```bash
git add backend/internal/adapters/adapter.go backend/internal/adapters/registry.go backend/internal/agentregistry/
git commit -m "feat(backend): adapter interface, in-process registry, agent_registry store"
```

---

### Task 15: Mock adapter with scripted YAML scenarios

The Mock adapter is permanent infrastructure — it is the conformance baseline that every future adapter is compared against. Scenarios are YAML files that name a trigger pattern and an event sequence.

**Files:**
- Create: `backend/internal/adapters/mock/mock.go`, `..._test.go`
- Create: `backend/internal/adapters/mock/scenarios/default.yaml`
- Create: `backend/internal/adapters/mock/scenarios/trino-investigation.yaml`

- [ ] **Step 1: Define the scenario file format**

`backend/internal/adapters/mock/scenarios/default.yaml`:
```yaml
# Triggered by any user message that doesn't match a more specific scenario.
name: default
match:
  any: true
events:
  - { type: run_started, run_id: __RUN_ID__ }
  - { type: text_delta, text: "(mock) I received: " }
  - { type: text_delta, text: "__USER_MESSAGE__" }
  - { type: run_finished, reason: done }
```

`backend/internal/adapters/mock/scenarios/trino-investigation.yaml`:
```yaml
# Triggered by messages mentioning "WIN-WS-014" or "powershell" — Phase-0 demo of
# the multi-step investigation shape that the real Trino agent will produce in Phase 1.
name: trino-investigation
match:
  contains_any: ["WIN-WS-014", "powershell", "investigate"]
events:
  - { type: run_started, run_id: __RUN_ID__ }
  - { type: text_delta, text: "(mock) Investigating the alert on WIN-WS-014." }
  - { type: state_update, key: current_table, value: detection_finding }
  - { type: tool_call_start, call_id: c1, tool: describe_table, args: { catalog: ocsf, schema: events, table: detection_finding } }
  - { type: tool_call_end, call_id: c1, ok: true, result_preview: "columns: time, class_uid, severity_id, finding_info.title, device.hostname" }
  - { type: tool_call_start, call_id: c2, tool: execute_query, args: { sql: "SELECT * FROM ocsf.events.detection_finding WHERE device.hostname='WIN-WS-014'" } }
  - { type: tool_call_end, call_id: c2, ok: true, result_preview: "1 row: suspicious_powershell_execution at 03:14 UTC" }
  - { type: text_delta, text: "Found suspicious PowerShell at 03:14 UTC. (Phase 1 will continue to process_activity, authentication, network_activity.)" }
  - { type: run_finished, reason: done }
```

- [ ] **Step 2: Write failing tests for the mock adapter**

`backend/internal/adapters/mock/mock_test.go`:
```go
package mock

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/phucvd2512/agentic-in-production/backend/internal/adapters"
	"github.com/stretchr/testify/require"
)

func TestRun_DefaultScenario_EchoesMessage(t *testing.T) {
	a, err := New(Config{ScenarioDir: "scenarios"})
	require.NoError(t, err)

	req := adapters.RunRequest{
		RunID:       "run_1",
		Session:     adapters.Session{ID: "s1", UserID: "admin", AgentName: "mock-trino-flavored"},
		UserMessage: "hello there",
	}
	ch, err := a.Run(context.Background(), req)
	require.NoError(t, err)

	var got []map[string]any
	for ev := range ch {
		var m map[string]any
		require.NoError(t, json.Unmarshal(ev.Payload, &m))
		got = append(got, m)
	}
	require.NotEmpty(t, got)
	require.Equal(t, "run_started", got[0]["type"])
	require.Equal(t, "run_finished", got[len(got)-1]["type"])

	// At least one text_delta should contain the user message verbatim.
	var saw bool
	for _, m := range got {
		if m["type"] == "text_delta" && strings.Contains(m["text"].(string), "hello there") {
			saw = true; break
		}
	}
	require.True(t, saw, "expected user message to appear in echoed text")
}

func TestRun_TrinoInvestigationScenario_TriggersOnKeyword(t *testing.T) {
	a, err := New(Config{ScenarioDir: "scenarios"})
	require.NoError(t, err)

	req := adapters.RunRequest{
		RunID:       "run_2",
		Session:     adapters.Session{ID: "s1", UserID: "admin", AgentName: "mock-trino-flavored"},
		UserMessage: "investigate WIN-WS-014",
	}
	ch, _ := a.Run(context.Background(), req)
	var types []string
	for ev := range ch {
		var m map[string]any
		_ = json.Unmarshal(ev.Payload, &m)
		types = append(types, m["type"].(string))
	}
	// expect at least one tool_call_start in the trino scenario
	var sawTool bool
	for _, t := range types { if t == "tool_call_start" { sawTool = true; break } }
	require.True(t, sawTool, "trino-investigation scenario should emit tool calls")
}
```

Run: `go test ./internal/adapters/mock/...` → expect FAIL (build).

- [ ] **Step 3: Implement `mock.go`**

```go
package mock

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/phucvd2512/agentic-in-production/backend/internal/adapters"
	"gopkg.in/yaml.v3"
)

type Config struct {
	ScenarioDir string `json:"scenario_dir"`
}

type Adapter struct {
	scenarios []scenario
}

type scenario struct {
	Name   string         `yaml:"name"`
	Match  matchRule      `yaml:"match"`
	Events []rawEvent     `yaml:"events"`
}

type matchRule struct {
	Any          bool     `yaml:"any"`
	ContainsAny  []string `yaml:"contains_any"`
}

type rawEvent map[string]any

func New(c Config) (*Adapter, error) {
	dir := c.ScenarioDir
	if dir == "" { dir = "scenarios" }
	matches, err := filepath.Glob(filepath.Join(dir, "*.yaml"))
	if err != nil { return nil, err }
	if len(matches) == 0 { return nil, errors.New("no scenarios in " + dir) }

	var scenarios []scenario
	var defaultScenario *scenario
	for _, m := range matches {
		b, err := os.ReadFile(m)
		if err != nil { return nil, err }
		var s scenario
		if err := yaml.Unmarshal(b, &s); err != nil { return nil, fmt.Errorf("%s: %w", m, err) }
		if s.Match.Any {
			scenarios = append(scenarios, s)
			defaultScenario = &scenarios[len(scenarios)-1]
		} else {
			scenarios = append(scenarios, s)
		}
	}
	if defaultScenario == nil {
		return nil, errors.New("at least one scenario must have match: any: true (the default)")
	}
	return &Adapter{scenarios: scenarios}, nil
}

func init() {
	adapters.Register("mock", func(configJSON []byte) (adapters.AgentPlatformAdapter, error) {
		var c Config
		if len(configJSON) > 0 {
			if err := json.Unmarshal(configJSON, &c); err != nil { return nil, err }
		}
		return New(c)
	})
}

func (a *Adapter) Name() string { return "mock" }

func (a *Adapter) Capabilities(_ context.Context) (adapters.Capabilities, error) {
	return adapters.Capabilities{Name: "mock", Tools: []string{"describe_table", "execute_query", "list_tables", "list_schemas", "list_catalogs"}}, nil
}

func (a *Adapter) StartConversation(_ context.Context, sess adapters.Session) (string, error) {
	return "mock_conv_" + sess.ID, nil
}

func (a *Adapter) EndConversation(_ context.Context, _ string) error { return nil }

func (a *Adapter) Run(ctx context.Context, req adapters.RunRequest) (<-chan adapters.AgentEvent, error) {
	s := a.pick(req.UserMessage)
	out := make(chan adapters.AgentEvent, len(s.Events))
	go func() {
		defer close(out)
		for _, raw := range s.Events {
			ev := substitute(raw, req)
			b, err := json.Marshal(ev)
			if err != nil { return }
			t, _ := ev["type"].(string)
			runID, _ := ev["run_id"].(string)
			select {
			case <-ctx.Done(): return
			case out <- adapters.AgentEvent{Type: t, RunID: runID, Payload: b}:
			}
		}
	}()
	return out, nil
}

func (a *Adapter) pick(userMsg string) scenario {
	lower := strings.ToLower(userMsg)
	var def *scenario
	for i := range a.scenarios {
		s := &a.scenarios[i]
		if s.Match.Any { def = s; continue }
		for _, kw := range s.Match.ContainsAny {
			if strings.Contains(lower, strings.ToLower(kw)) { return *s }
		}
	}
	return *def
}

// substitute replaces __RUN_ID__ and __USER_MESSAGE__ tokens in event values.
func substitute(in rawEvent, req adapters.RunRequest) rawEvent {
	out := make(rawEvent, len(in))
	for k, v := range in {
		out[k] = sub(v, req)
	}
	return out
}

func sub(v any, req adapters.RunRequest) any {
	switch t := v.(type) {
	case string:
		s := strings.ReplaceAll(t, "__RUN_ID__", req.RunID)
		s = strings.ReplaceAll(s, "__USER_MESSAGE__", req.UserMessage)
		return s
	case map[string]any:
		o := make(map[string]any, len(t))
		for k, vv := range t { o[k] = sub(vv, req) }
		return o
	case []any:
		o := make([]any, len(t))
		for i, vv := range t { o[i] = sub(vv, req) }
		return o
	default:
		return v
	}
}
```

Run tests → PASS.

- [ ] **Step 4: Cross-check Mock-emitted events against the schema**

Add to `mock_test.go`:
```go
func TestRun_AllEmittedEventsAreSchemaValid(t *testing.T) {
	a, err := New(Config{ScenarioDir: "scenarios"})
	require.NoError(t, err)

	for _, msg := range []string{"hello", "investigate WIN-WS-014", "show me powershell"} {
		req := adapters.RunRequest{RunID: "r", Session: adapters.Session{ID: "s"}, UserMessage: msg}
		ch, _ := a.Run(context.Background(), req)
		for ev := range ch {
			require.NoError(t, protocol.ValidateEvent(ev.Payload), "schema violation for: %s", string(ev.Payload))
		}
	}
}
```

Add the import: `"github.com/phucvd2512/agentic-in-production/backend/internal/protocol"`.

Run again → PASS.

- [ ] **Step 5: Commit**

```bash
git add backend/internal/adapters/mock/
git commit -m "feat(backend): mock adapter with YAML-scripted scenarios"
```

---

### Task 16: SSE streaming endpoint + sessions/agents handlers

**Files:**
- Create: `backend/internal/httpapi/sessions_handler.go`, `..._test.go`
- Create: `backend/internal/httpapi/messages_handler.go`, `..._test.go`
- Create: `backend/internal/httpapi/agents_handler.go`
- Create: `backend/internal/httpapi/audit_handler.go`
- Modify: `backend/internal/httpapi/router.go`, `backend/cmd/server/main.go`

- [ ] **Step 1: Implement `agents_handler.go`** (simple list)

```go
package httpapi

import (
	"encoding/json"
	"net/http"

	"github.com/phucvd2512/agentic-in-production/backend/internal/agentregistry"
)

type AgentsHandler struct{ Registry *agentregistry.Store }

func (h AgentsHandler) List(w http.ResponseWriter, r *http.Request) {
	agents, err := h.Registry.ListEnabled(r.Context())
	if err != nil { http.Error(w, err.Error(), http.StatusInternalServerError); return }
	w.Header().Set("content-type", "application/json")
	_ = json.NewEncoder(w).Encode(agents)
}
```

- [ ] **Step 2: Implement `sessions_handler.go`**

```go
package httpapi

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/phucvd2512/agentic-in-production/backend/internal/sessions"
)

type SessionsHandler struct{ Sessions *sessions.Store }

func (h SessionsHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req struct{ AgentName string `json:"agent_name"` }
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil { http.Error(w, "bad request", 400); return }
	user := UserFromCtx(r.Context())
	sess, err := h.Sessions.Create(r.Context(), user, req.AgentName)
	if err != nil { http.Error(w, err.Error(), 500); return }
	w.Header().Set("content-type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"id":         sess.ID,
		"agent_name": sess.AgentName,
		"created_at": sess.CreatedAt,
	})
}

func (h SessionsHandler) ListMessages(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	msgs, err := h.Sessions.ListMessages(r.Context(), id)
	if err != nil { http.Error(w, err.Error(), 500); return }
	w.Header().Set("content-type", "application/json")
	_ = json.NewEncoder(w).Encode(msgs)
}
```

- [ ] **Step 3: Implement `audit_handler.go`**

```go
package httpapi

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/phucvd2512/agentic-in-production/backend/internal/audit"
)

type AuditHandler struct{ Log *audit.Log }

func (h AuditHandler) Get(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	entries, err := h.Log.List(r.Context(), id)
	if err != nil { http.Error(w, err.Error(), 500); return }
	w.Header().Set("content-type", "application/json")
	_ = json.NewEncoder(w).Encode(entries)
}
```

- [ ] **Step 4: Implement the SSE message handler — the centrepiece**

`backend/internal/httpapi/messages_handler.go`:
```go
package httpapi

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/phucvd2512/agentic-in-production/backend/internal/adapters"
	"github.com/phucvd2512/agentic-in-production/backend/internal/agentregistry"
	"github.com/phucvd2512/agentic-in-production/backend/internal/audit"
	"github.com/phucvd2512/agentic-in-production/backend/internal/protocol"
	"github.com/phucvd2512/agentic-in-production/backend/internal/sessions"
)

type MessagesHandler struct {
	Sessions *sessions.Store
	Registry *agentregistry.Store
	Audit    *audit.Log
}

func newRunID() string {
	b := make([]byte, 8); _, _ = rand.Read(b); return "run_" + hex.EncodeToString(b)
}

func (h MessagesHandler) Send(w http.ResponseWriter, r *http.Request) {
	sessionID := chi.URLParam(r, "id")
	var body struct{ Text string `json:"text"` }
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil { http.Error(w, "bad request", 400); return }
	if body.Text == "" { http.Error(w, "empty text", 400); return }

	ctx := r.Context()

	// 1. Look up session + agent.
	sess, err := h.Sessions.Get(ctx, sessionID)
	if err != nil { http.Error(w, "session not found", 404); return }
	agent, err := h.Registry.GetByName(ctx, sess.AgentName)
	if err != nil { http.Error(w, "agent not found", 404); return }

	// 2. Persist user message.
	if _, err := h.Sessions.AppendMessage(ctx, sessionID, "user", body.Text); err != nil {
		http.Error(w, "persist failed", 500); return
	}

	// 3. Build adapter on demand (Phase 0: cheap; Phase 1+ may want pooling).
	ad, err := adapters.Build(agent.AdapterKind, agent.Config)
	if err != nil { http.Error(w, err.Error(), 500); return }

	// 4. Ensure platform conversation id (StartConversation if absent).
	convID := ""
	if sess.PlatformConvID != nil { convID = *sess.PlatformConvID }
	if convID == "" {
		convID, err = ad.StartConversation(ctx, adapters.Session{ID: sess.ID, UserID: sess.UserID, AgentName: sess.AgentName})
		if err != nil { http.Error(w, err.Error(), 500); return }
		_ = h.Sessions.SetPlatformConvID(ctx, sess.ID, convID)
	}

	// 5. Open SSE stream to client.
	w.Header().Set("content-type", "text/event-stream")
	w.Header().Set("cache-control", "no-cache")
	w.Header().Set("connection", "keep-alive")
	flusher, ok := w.(http.Flusher)
	if !ok { http.Error(w, "streaming unsupported", 500); return }

	runID := newRunID()
	req := adapters.RunRequest{
		RunID:          runID,
		Session:        adapters.Session{ID: sess.ID, UserID: sess.UserID, AgentName: sess.AgentName},
		PlatformConvID: convID,
		UserMessage:    body.Text,
	}
	ch, err := ad.Run(ctx, req)
	if err != nil {
		writeEvent(w, errEvent("internal_error", err.Error()))
		writeEvent(w, finishEvent("error"))
		flusher.Flush(); return
	}

	// 6. Forward + audit.
	var assistantText string
	for ev := range ch {
		// Validate before forwarding (per ADR-0005 / spec §4.4).
		if vErr := protocol.ValidateEvent(ev.Payload); vErr != nil {
			slog.Warn("adapter emitted invalid event", "err", vErr, "raw", string(ev.Payload))
			writeEvent(w, errEvent("internal_error", "adapter emitted invalid event: "+vErr.Error()))
			writeEvent(w, finishEvent("error"))
			flusher.Flush(); return
		}
		writeEvent(w, ev.Payload)
		flusher.Flush()
		_ = h.Audit.Append(ctx, sess.ID, runID, ev.Type, ev.Payload)

		if ev.Type == "text_delta" {
			var td struct{ Text string `json:"text"` }
			_ = json.Unmarshal(ev.Payload, &td)
			assistantText += td.Text
		}
	}

	if assistantText != "" {
		_, _ = h.Sessions.AppendMessage(ctx, sessionID, "assistant", assistantText)
	}
}

func writeEvent(w io.Writer, payload []byte) {
	_, _ = fmt.Fprintf(w, "data: %s\n\n", payload)
}

func errEvent(code, msg string) []byte {
	b, _ := json.Marshal(map[string]any{"type": "error", "code": code, "message": msg})
	return b
}

func finishEvent(reason string) []byte {
	b, _ := json.Marshal(map[string]any{"type": "run_finished", "reason": reason})
	return b
}
```

- [ ] **Step 5: Wire all handlers into the router**

Update `backend/internal/httpapi/router.go` to:
```go
package httpapi

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/phucvd2512/agentic-in-production/backend/internal/agentregistry"
	"github.com/phucvd2512/agentic-in-production/backend/internal/audit"
	"github.com/phucvd2512/agentic-in-production/backend/internal/auth"
	"github.com/phucvd2512/agentic-in-production/backend/internal/db"
	"github.com/phucvd2512/agentic-in-production/backend/internal/sessions"
)

type Deps struct {
	Version  string
	DB       *db.DB
	Auth     *auth.Authenticator
	Sessions *sessions.Store
	Registry *agentregistry.Store
	Audit    *audit.Log
}

func NewRouter(d Deps) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)

	r.Get("/api/healthz", HealthzHandler{Version: d.Version}.ServeHTTP)
	r.Post("/api/auth/login", AuthHandler{Auth: d.Auth}.Login)

	r.Group(func(r chi.Router) {
		r.Use(RequireAuth(d.Auth))
		r.Get("/api/agents", AgentsHandler{Registry: d.Registry}.List)
		r.Post("/api/sessions", SessionsHandler{Sessions: d.Sessions}.Create)
		r.Get("/api/sessions/{id}/messages", SessionsHandler{Sessions: d.Sessions}.ListMessages)
		r.Post("/api/sessions/{id}/messages",
			MessagesHandler{Sessions: d.Sessions, Registry: d.Registry, Audit: d.Audit}.Send)
		r.Get("/api/sessions/{id}/audit", AuditHandler{Log: d.Audit}.Get)
	})
	return r
}
```

Update `backend/cmd/server/main.go` to construct each dependency from env vars and the DB pool, and pass them all into `Deps`. Also add a side-effect import to register the mock adapter:

```go
import (
	_ "github.com/phucvd2512/agentic-in-production/backend/internal/adapters/mock"
)
```

The mock factory's `Config.ScenarioDir` defaults to `"scenarios"`, so the mock adapter must be initialised with `ScenarioDir: "internal/adapters/mock/scenarios"` *relative to the working directory at server start*. Set it via the `agent_registry.config` JSON when seeding (or override in `mock.New` to a path resolved against `os.Args[0]` + module root). Easiest in Phase 0: set the `config` to `{"scenario_dir":"internal/adapters/mock/scenarios"}` and run the server from `backend/`.

Update the seed in `infra/postgres/init.sql`:
```sql
INSERT INTO agent_registry (name, version, description, enabled, adapter_kind, config) VALUES
    ('mock-trino-flavored', '0.1.0', 'Scripted Trino-flavored conversations for MVP smoke',
     true, 'mock', '{"scenario_dir":"internal/adapters/mock/scenarios"}')
ON CONFLICT (name) DO NOTHING;
```

(re-running `docker compose down -v && docker compose up -d postgres` reseeds)

- [ ] **Step 6: Manual end-to-end smoke test**

```bash
docker compose up -d postgres
make admin-password   # sets ADMIN_PASSWORD_HASH
cd backend && go run ./cmd/server &
sleep 1

# login
curl -s -c /tmp/aip.cookies -X POST http://localhost:8080/api/auth/login \
  -H 'content-type: application/json' \
  -d '{"username":"admin","password":"hunter2"}' | jq .

# create session
SID=$(curl -s -b /tmp/aip.cookies -X POST http://localhost:8080/api/sessions \
  -H 'content-type: application/json' -d '{"agent_name":"mock-trino-flavored"}' | jq -r .id)
echo "session=$SID"

# send message + read SSE
curl -N -s -b /tmp/aip.cookies -X POST "http://localhost:8080/api/sessions/$SID/messages" \
  -H 'content-type: application/json' -d '{"text":"investigate WIN-WS-014"}'

# audit
curl -s -b /tmp/aip.cookies "http://localhost:8080/api/sessions/$SID/audit" | jq '.[0]'
kill %1
```

Expected: SSE stream with `data: {"type":"run_started",...}`, several `data: {"type":"text_delta",...}` and `tool_call_*` events, ending with `data: {"type":"run_finished","reason":"done"}`. Audit endpoint returns at least the same number of entries.

- [ ] **Step 7: Commit**

```bash
git add backend/internal/httpapi/ backend/cmd/server/main.go infra/postgres/init.sql
git commit -m "feat(backend): SSE streaming endpoint, sessions/agents/audit handlers, full router"
```

---

### Task 17: Adapter conformance test framework

The conformance suite is the load-bearing test for the platform-agnostic abstraction. Phase 0 ships the framework and a Mock-only suite; Phase 1+ adds runs against real platforms.

**Files:**
- Create: `backend/internal/adapters/conformance/runner.go`, `runner_test.go`
- Create: `backend/internal/adapters/conformance/scenarios/01-text-only.yaml`
- Create: `backend/internal/adapters/conformance/scenarios/02-tool-call-success.yaml`
- Create: `backend/internal/adapters/conformance/scenarios/03-state-update.yaml`

- [ ] **Step 1: Define the conformance scenario format** (different from Mock's: input + expected event-type sequence)

`scenarios/01-text-only.yaml`:
```yaml
name: text-only
input: "hello there"
expect_types:
  - run_started
  - text_delta
  - run_finished
expect_min_events: 3
```

`scenarios/02-tool-call-success.yaml`:
```yaml
name: tool-call-success
input: "investigate WIN-WS-014"
expect_types_contains:
  - tool_call_start
  - tool_call_end
expect_starts_with: run_started
expect_ends_with: run_finished
```

`scenarios/03-state-update.yaml`:
```yaml
name: state-update
input: "investigate WIN-WS-014"
expect_types_contains:
  - state_update
expect_ends_with: run_finished
```

- [ ] **Step 2: Implement the runner**

`backend/internal/adapters/conformance/runner.go`:
```go
package conformance

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/phucvd2512/agentic-in-production/backend/internal/adapters"
	"github.com/phucvd2512/agentic-in-production/backend/internal/protocol"
	"gopkg.in/yaml.v3"
)

type Scenario struct {
	Name                 string   `yaml:"name"`
	Input                string   `yaml:"input"`
	ExpectTypes          []string `yaml:"expect_types"`
	ExpectTypesContains  []string `yaml:"expect_types_contains"`
	ExpectStartsWith     string   `yaml:"expect_starts_with"`
	ExpectEndsWith       string   `yaml:"expect_ends_with"`
	ExpectMinEvents      int      `yaml:"expect_min_events"`
}

func LoadScenarios(dir string) ([]Scenario, error) {
	matches, err := filepath.Glob(filepath.Join(dir, "*.yaml"))
	if err != nil { return nil, err }
	var out []Scenario
	for _, m := range matches {
		b, err := os.ReadFile(m)
		if err != nil { return nil, err }
		var s Scenario
		if err := yaml.Unmarshal(b, &s); err != nil { return nil, fmt.Errorf("%s: %w", m, err) }
		out = append(out, s)
	}
	return out, nil
}

// Run executes a single scenario against the given adapter and returns
// the collected event types (already schema-validated) or an error
// describing the first invariant violation.
func Run(ctx context.Context, ad adapters.AgentPlatformAdapter, s Scenario) ([]string, error) {
	convID, err := ad.StartConversation(ctx, adapters.Session{ID: "conf_sess", UserID: "conf"})
	if err != nil { return nil, fmt.Errorf("start_conv: %w", err) }
	defer ad.EndConversation(ctx, convID)

	ch, err := ad.Run(ctx, adapters.RunRequest{
		RunID: "conf_run", PlatformConvID: convID,
		Session: adapters.Session{ID: "conf_sess", UserID: "conf"},
		UserMessage: s.Input,
	})
	if err != nil { return nil, err }

	var got []string
	for ev := range ch {
		if err := protocol.ValidateEvent(ev.Payload); err != nil {
			return got, fmt.Errorf("schema: %w (raw=%s)", err, string(ev.Payload))
		}
		var m map[string]any
		_ = json.Unmarshal(ev.Payload, &m)
		got = append(got, m["type"].(string))
	}
	return got, nil
}

func Verify(s Scenario, got []string) error {
	if s.ExpectMinEvents > 0 && len(got) < s.ExpectMinEvents {
		return fmt.Errorf("expected at least %d events, got %d", s.ExpectMinEvents, len(got))
	}
	if s.ExpectStartsWith != "" && (len(got) == 0 || got[0] != s.ExpectStartsWith) {
		return fmt.Errorf("expected first event %q, got %v", s.ExpectStartsWith, got)
	}
	if s.ExpectEndsWith != "" && (len(got) == 0 || got[len(got)-1] != s.ExpectEndsWith) {
		return fmt.Errorf("expected last event %q, got %v", s.ExpectEndsWith, got)
	}
	for _, t := range s.ExpectTypesContains {
		if !contains(got, t) {
			return fmt.Errorf("expected to contain %q, got %v", t, got)
		}
	}
	if len(s.ExpectTypes) > 0 {
		// strict in-order subset check
		i := 0
		for _, want := range s.ExpectTypes {
			for i < len(got) && got[i] != want { i++ }
			if i >= len(got) { return fmt.Errorf("expected sequence %v missing %q after position %d", s.ExpectTypes, want, i) }
			i++
		}
	}
	return nil
}

func contains(xs []string, x string) bool {
	for _, v := range xs { if v == x { return true } }
	return false
}
```

- [ ] **Step 3: Run the conformance suite against the Mock adapter**

`backend/internal/adapters/conformance/runner_test.go`:
```go
package conformance

import (
	"context"
	"testing"

	"github.com/phucvd2512/agentic-in-production/backend/internal/adapters/mock"
	"github.com/stretchr/testify/require"
)

func TestMockAdapter_PassesAllScenarios(t *testing.T) {
	scenarios, err := LoadScenarios("scenarios")
	require.NoError(t, err)
	require.NotEmpty(t, scenarios)

	ad, err := mock.New(mock.Config{ScenarioDir: "../mock/scenarios"})
	require.NoError(t, err)

	for _, s := range scenarios {
		s := s
		t.Run(s.Name, func(t *testing.T) {
			got, err := Run(context.Background(), ad, s)
			require.NoError(t, err)
			require.NoError(t, Verify(s, got), "got events: %v", got)
		})
	}
}
```

Run: `cd backend && go test ./internal/adapters/conformance/...`
Expected: all three scenarios PASS.

- [ ] **Step 4: Commit**

```bash
git add backend/internal/adapters/conformance/
git commit -m "feat(backend): adapter conformance framework + Mock-passing scenarios"
```

---

### Phase 0.1a checkpoint

Backend track is feature-complete for MVP. Smoke-test:

```bash
make test-backend             # all unit + conformance tests pass
docker compose up -d postgres
cd backend && go run ./cmd/server &
sleep 1
curl -s http://localhost:8080/api/healthz
kill %1
```

---

## Phase 0.1b — Frontend track

> **Parallelizable with Phase 0.1a.** Until backend is wired, the frontend works against `protocols/examples/*.json` fixtures.

### Task 18: Initialize Vite + React + TypeScript + Vitest

**Files:**
- Create: `frontend/package.json`, `frontend/vite.config.ts`, `frontend/tsconfig.json`, `frontend/.eslintrc.cjs`, `frontend/index.html`, `frontend/src/main.tsx`, `frontend/src/App.tsx`
- Modify: `frontend/CLAUDE.md` (replace placeholder)

- [ ] **Step 1: Scaffold the Vite project**

```bash
cd frontend
pnpm create vite@5 . --template react-ts
# accept overwrite for the README and any other conflicts; we keep our own README
pnpm install
```

This creates `package.json`, `vite.config.ts`, `tsconfig.json`, `index.html`, `src/main.tsx`, `src/App.tsx`, etc. Replace generated `App.tsx` with a minimal shell:

```tsx
// frontend/src/App.tsx
export function App() {
  return <main className="p-6"><h1>agentic-in-production</h1></main>;
}
export default App;
```

- [ ] **Step 2: Add dev dependencies**

```bash
pnpm add zustand
pnpm add -D vitest @vitest/ui @testing-library/react @testing-library/dom @testing-library/jest-dom jsdom \
            @types/node \
            eslint @typescript-eslint/parser @typescript-eslint/eslint-plugin eslint-plugin-react eslint-plugin-react-hooks \
            prettier eslint-config-prettier eslint-plugin-prettier \
            @playwright/test
```

- [ ] **Step 3: Author `frontend/vite.config.ts`** to add Vitest config

```ts
import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";

export default defineConfig({
  plugins: [react()],
  server: {
    port: 5173,
    proxy: {
      "/api": "http://localhost:8080",
    },
  },
  test: {
    environment: "jsdom",
    globals: true,
    setupFiles: ["./src/test-setup.ts"],
  },
});
```

Create `frontend/src/test-setup.ts`:
```ts
import "@testing-library/jest-dom";
```

- [ ] **Step 4: Author `frontend/tsconfig.json`**

Replace the generated tsconfig with one that's strict and Vitest-friendly:

```json
{
  "compilerOptions": {
    "target": "ES2022",
    "lib": ["ES2022", "DOM", "DOM.Iterable"],
    "module": "ESNext",
    "moduleResolution": "bundler",
    "jsx": "react-jsx",
    "strict": true,
    "noUncheckedIndexedAccess": true,
    "noImplicitOverride": true,
    "esModuleInterop": true,
    "skipLibCheck": true,
    "forceConsistentCasingInFileNames": true,
    "resolveJsonModule": true,
    "types": ["vitest/globals", "@testing-library/jest-dom"],
    "baseUrl": ".",
    "paths": {
      "@/*": ["src/*"],
      "@protocols/*": ["../protocols/*"]
    }
  },
  "include": ["src", "vite.config.ts"]
}
```

- [ ] **Step 5: Author `frontend/.eslintrc.cjs`**

```js
module.exports = {
  root: true,
  parser: "@typescript-eslint/parser",
  plugins: ["@typescript-eslint", "react", "react-hooks", "prettier"],
  extends: [
    "eslint:recommended",
    "plugin:@typescript-eslint/recommended",
    "plugin:react/recommended",
    "plugin:react-hooks/recommended",
    "prettier",
  ],
  settings: { react: { version: "detect" } },
  ignorePatterns: ["dist", "*.gen.ts"],
  rules: {
    "react/react-in-jsx-scope": "off",
    "@typescript-eslint/no-unused-vars": ["error", { argsIgnorePattern: "^_" }],
  },
};
```

- [ ] **Step 6: Add scripts to `frontend/package.json`**

Edit the `"scripts"` block:
```json
"scripts": {
  "dev": "vite",
  "build": "tsc -b && vite build",
  "preview": "vite preview",
  "test": "vitest run",
  "test:watch": "vitest",
  "lint": "eslint src --ext .ts,.tsx",
  "typecheck": "tsc -b --noEmit",
  "fmt": "prettier --write 'src/**/*.{ts,tsx,css}'"
}
```

- [ ] **Step 7: Update `frontend/CLAUDE.md`**

```markdown
# Frontend (React + TypeScript) — guidance for AI coding agents

Vite + React 18 + TypeScript. Talks to the Go backend at `/api/*`. SSE for the message stream.

## Conventions
- Functional components only. Hooks for local state.
- Global state via [Zustand](https://github.com/pmndrs/zustand) — minimal. Each store is a small file in `src/store/`.
- Path alias `@/` → `src/`, `@protocols/` → `../protocols/`. Both resolve at build and test time.
- Generated types live in `src/api/types.gen.ts` (REST) and `src/events/types.gen.ts` (events). **Don't edit by hand** — `make gen` regenerates them from `protocols/`.
- SSE parser is in `src/events/parser.ts`. New event types are not added by hand — they appear when the JSON Schema changes and codegen runs.

## Testing
- Unit: Vitest + Testing Library. Test file lives next to the component (`Foo.test.tsx`).
- E2E: Playwright (one golden-path test for MVP, in `e2e/`). Driven from `make test-e2e`.

## Offline dev (no backend)
The `src/fixtures/` folder imports from `../../protocols/examples/*.json` so the UI can be developed against canonical event sequences before the backend exists. See `docs/superpowers/plans/2026-05-01-phase-0-mvp-skeleton.md` Task 24.

## When making changes
- Touching `*.gen.ts`? Don't — run `make gen` from repo root.
- Adding an HTTP call? Update `protocols/openapi.yaml` first, regenerate, then call.
- Adding a new event type? Update `protocols/agent-events.schema.json` first; the parser becomes a compile error until you handle it.
```

- [ ] **Step 8: Smoke test build + dev server**

```bash
cd frontend
pnpm typecheck     # should pass
pnpm test          # zero tests, exits 0
pnpm dev &
sleep 2
curl -s http://localhost:5173 | head -1   # expect <!doctype html>
kill %1
```

- [ ] **Step 9: Commit**

```bash
git add frontend/
git commit -m "feat(frontend): scaffold Vite + React + TS + Vitest + ESLint"
```

---

### Task 19: Run codegen → frontend types

Re-runs `make gen` from Task 8 now that `frontend/` exists. After this task, `frontend/src/api/types.gen.ts` and `frontend/src/events/types.gen.ts` exist as committed files.

- [ ] **Step 1: Run codegen**

```bash
cd .. && make gen   # from repo root
```

Expected: creates/updates `frontend/src/api/types.gen.ts` and `frontend/src/events/types.gen.ts`.

- [ ] **Step 2: Sanity-check the generated files**

```bash
head -3 frontend/src/api/types.gen.ts
head -3 frontend/src/events/types.gen.ts
```
Expected: both start with `// GENERATED by protocols/gen-types.sh — do not edit by hand`.

- [ ] **Step 3: Verify TS compiles with the generated types**

```bash
cd frontend && pnpm typecheck
```
Expected: PASS.

- [ ] **Step 4: Commit**

```bash
git add frontend/src/api/types.gen.ts frontend/src/events/types.gen.ts \
        backend/internal/protocol/events.gen.go backend/internal/protocol/api.gen.go
git commit -m "build: commit initial generated types from protocols/"
```

---

### Task 20: API client + auth store + login UI

**Files:**
- Create: `frontend/src/api/client.ts`, `client.test.ts`
- Create: `frontend/src/store/auth.ts`
- Create: `frontend/src/components/LoginForm.tsx`, `LoginForm.test.tsx`
- Create: `frontend/src/pages/Login.tsx`

- [ ] **Step 1: Author `api/client.ts`** (typed fetch wrapper)

```ts
// frontend/src/api/client.ts
import type { components } from "./types.gen";

type LoginRequest = components["schemas"]["LoginRequest"];
type LoginResponse = components["schemas"]["LoginResponse"];
type Agent = components["schemas"]["Agent"];
type Session = components["schemas"]["Session"];
type Message = components["schemas"]["Message"];
type AuditEntry = components["schemas"]["AuditEntry"];

const BASE = import.meta.env.VITE_API_BASE_URL ?? "";

async function req<T>(path: string, init?: RequestInit): Promise<T> {
  const res = await fetch(BASE + path, {
    credentials: "include",
    headers: { "content-type": "application/json", ...(init?.headers ?? {}) },
    ...init,
  });
  if (!res.ok) throw new Error(`${res.status} ${res.statusText}: ${await res.text()}`);
  return res.json() as Promise<T>;
}

export const api = {
  healthz: () => req<{ ok: boolean; version?: string }>("/api/healthz"),
  login:   (body: LoginRequest) => req<LoginResponse>("/api/auth/login", { method: "POST", body: JSON.stringify(body) }),
  agents:  () => req<Agent[]>("/api/agents"),
  createSession: (agent_name: string) => req<Session>("/api/sessions", { method: "POST", body: JSON.stringify({ agent_name }) }),
  listMessages:  (id: string) => req<Message[]>(`/api/sessions/${id}/messages`),
  audit:         (id: string) => req<AuditEntry[]>(`/api/sessions/${id}/audit`),
};

// SSE message-send returns a ReadableStream of decoded text lines.
export function sendMessage(sessionID: string, text: string, signal: AbortSignal) {
  return fetch(`${BASE}/api/sessions/${sessionID}/messages`, {
    method: "POST",
    credentials: "include",
    headers: { "content-type": "application/json", accept: "text/event-stream" },
    body: JSON.stringify({ text }),
    signal,
  });
}
```

- [ ] **Step 2: Test `api.login` against a mocked fetch**

`frontend/src/api/client.test.ts`:
```ts
import { describe, it, expect, vi } from "vitest";
import { api } from "./client";

describe("api.login", () => {
  it("posts JSON to /api/auth/login", async () => {
    const json = vi.fn().mockResolvedValue({ token: "tok" });
    const fetchMock = vi.fn().mockResolvedValue({ ok: true, json });
    vi.stubGlobal("fetch", fetchMock);
    const res = await api.login({ username: "admin", password: "x" });
    expect(res).toEqual({ token: "tok" });
    expect(fetchMock).toHaveBeenCalledWith(
      expect.stringContaining("/api/auth/login"),
      expect.objectContaining({ method: "POST" })
    );
  });

  it("throws on non-2xx", async () => {
    vi.stubGlobal("fetch", vi.fn().mockResolvedValue({ ok: false, status: 401, statusText: "Unauthorized", text: () => Promise.resolve("bad") }));
    await expect(api.login({ username: "x", password: "x" })).rejects.toThrow(/401/);
  });
});
```

Run: `pnpm test` → PASS.

- [ ] **Step 3: Author `store/auth.ts`** (Zustand)

```ts
// frontend/src/store/auth.ts
import { create } from "zustand";

type AuthState = {
  loggedIn: boolean;
  username?: string;
  setLoggedIn: (username: string) => void;
  setLoggedOut: () => void;
};

export const useAuthStore = create<AuthState>((set) => ({
  loggedIn: false,
  setLoggedIn: (username) => set({ loggedIn: true, username }),
  setLoggedOut: () => set({ loggedIn: false, username: undefined }),
}));
```

- [ ] **Step 4: Author `components/LoginForm.tsx` + test**

`frontend/src/components/LoginForm.tsx`:
```tsx
import { useState } from "react";
import { api } from "@/api/client";
import { useAuthStore } from "@/store/auth";

export function LoginForm() {
  const setLoggedIn = useAuthStore((s) => s.setLoggedIn);
  const [username, setUsername] = useState("admin");
  const [password, setPassword] = useState("");
  const [err, setErr] = useState<string | null>(null);
  const [busy, setBusy] = useState(false);

  async function submit(e: React.FormEvent) {
    e.preventDefault();
    setErr(null); setBusy(true);
    try {
      await api.login({ username, password });
      setLoggedIn(username);
    } catch (e) {
      setErr(e instanceof Error ? e.message : "login failed");
    } finally {
      setBusy(false);
    }
  }

  return (
    <form onSubmit={submit} aria-label="login form" style={{ display: "grid", gap: 8, maxWidth: 320 }}>
      <label>Username<input value={username} onChange={(e) => setUsername(e.target.value)} /></label>
      <label>Password<input type="password" value={password} onChange={(e) => setPassword(e.target.value)} /></label>
      <button type="submit" disabled={busy}>{busy ? "Signing in..." : "Sign in"}</button>
      {err && <div role="alert" style={{ color: "crimson" }}>{err}</div>}
    </form>
  );
}
```

`frontend/src/components/LoginForm.test.tsx`:
```tsx
import { describe, it, expect, vi } from "vitest";
import { render, screen, fireEvent, waitFor } from "@testing-library/react";
import { LoginForm } from "./LoginForm";

describe("LoginForm", () => {
  it("calls /api/auth/login on submit", async () => {
    const fetchMock = vi.fn().mockResolvedValue({
      ok: true, json: () => Promise.resolve({ token: "tok" }),
    });
    vi.stubGlobal("fetch", fetchMock);
    render(<LoginForm />);
    fireEvent.change(screen.getByLabelText(/password/i), { target: { value: "hunter2" } });
    fireEvent.click(screen.getByRole("button", { name: /sign in/i }));
    await waitFor(() => expect(fetchMock).toHaveBeenCalled());
  });
});
```

Run: `pnpm test` → PASS.

- [ ] **Step 5: Author `pages/Login.tsx`**

```tsx
import { LoginForm } from "@/components/LoginForm";
export function Login() {
  return (
    <main style={{ padding: 24 }}>
      <h1>agentic-in-production</h1>
      <LoginForm />
    </main>
  );
}
```

- [ ] **Step 6: Wire login into `App.tsx`** (minimal — full routing comes in Task 22)

```tsx
import { Login } from "@/pages/Login";
import { useAuthStore } from "@/store/auth";

export function App() {
  const loggedIn = useAuthStore((s) => s.loggedIn);
  if (!loggedIn) return <Login />;
  return <main style={{ padding: 24 }}><h1>agentic-in-production</h1><p>(home — implemented in Task 22)</p></main>;
}
export default App;
```

- [ ] **Step 7: Commit**

```bash
git add frontend/src/api/client*.ts frontend/src/store/auth.ts frontend/src/components/LoginForm*.tsx frontend/src/pages/Login.tsx frontend/src/App.tsx
git commit -m "feat(frontend): API client, auth store, login form"
```

---

### Task 21: SSE event-stream parser

The parser reads the SSE stream from the backend and emits typed events conforming to the JSON Schema.

**Files:**
- Create: `frontend/src/events/parser.ts`, `parser.test.ts`

- [ ] **Step 1: Write failing tests**

`frontend/src/events/parser.test.ts`:
```ts
import { describe, it, expect } from "vitest";
import { parseSSE, type AgentEvent } from "./parser";

function streamFrom(chunks: string[]): ReadableStream<Uint8Array> {
  const enc = new TextEncoder();
  return new ReadableStream({
    start(controller) {
      for (const c of chunks) controller.enqueue(enc.encode(c));
      controller.close();
    },
  });
}

describe("parseSSE", () => {
  it("parses two events split across chunks", async () => {
    const got: AgentEvent[] = [];
    await parseSSE(streamFrom([
      'data: {"type":"run_started","run_id":"r1"}\n\n',
      'data: {"type":"text_delta","text":"hi"}\n\n',
    ]), (ev) => got.push(ev));
    expect(got).toHaveLength(2);
    expect(got[0]).toMatchObject({ type: "run_started", run_id: "r1" });
    expect(got[1]).toMatchObject({ type: "text_delta", text: "hi" });
  });

  it("handles a JSON message split across chunk boundaries", async () => {
    const got: AgentEvent[] = [];
    await parseSSE(streamFrom([
      'data: {"type":"run_st',
      'arted","run_id":"r2"}\n\n',
    ]), (ev) => got.push(ev));
    expect(got).toEqual([{ type: "run_started", run_id: "r2" }]);
  });
});
```

Run: `pnpm test` → expect FAIL (no parser yet).

- [ ] **Step 2: Implement `parser.ts`**

```ts
// frontend/src/events/parser.ts
import type { AgentEvent as Generated } from "./types.gen";

// Re-export under a friendlier name. The generated type is a discriminated
// union on `type`, which is exactly what callers want.
export type AgentEvent = Generated;

/**
 * Parses an SSE stream where every event is `data: <json>\n\n`. Calls `onEvent`
 * for each successfully-parsed canonical event. Skips comment/heartbeat lines.
 */
export async function parseSSE(
  stream: ReadableStream<Uint8Array>,
  onEvent: (ev: AgentEvent) => void
): Promise<void> {
  const reader = stream.getReader();
  const decoder = new TextDecoder();
  let buf = "";

  for (;;) {
    const { value, done } = await reader.read();
    if (done) break;
    buf += decoder.decode(value, { stream: true });

    let sep: number;
    while ((sep = buf.indexOf("\n\n")) >= 0) {
      const block = buf.slice(0, sep);
      buf = buf.slice(sep + 2);
      const data = block
        .split("\n")
        .filter((l) => l.startsWith("data:"))
        .map((l) => l.slice(5).trimStart())
        .join("");
      if (!data) continue;
      try {
        onEvent(JSON.parse(data) as AgentEvent);
      } catch {
        // ignore unparseable frames; backend validates server-side
      }
    }
  }
}
```

Run tests → PASS.

- [ ] **Step 3: Commit**

```bash
git add frontend/src/events/parser.ts frontend/src/events/parser.test.ts
git commit -m "feat(frontend): SSE event-stream parser with chunk-boundary handling"
```

---

### Task 22: Sessions list, create, and routing

**Files:**
- Create: `frontend/src/store/session.ts`
- Create: `frontend/src/components/SessionList.tsx`, `..._test.tsx`
- Create: `frontend/src/pages/Sessions.tsx`
- Modify: `frontend/src/App.tsx`

- [ ] **Step 1: Author the session store**

```ts
// frontend/src/store/session.ts
import { create } from "zustand";
import type { components } from "@/api/types.gen";

type Session = components["schemas"]["Session"];

type State = {
  sessions: Session[];
  current: Session | null;
  setSessions: (s: Session[]) => void;
  setCurrent: (s: Session | null) => void;
};

export const useSessionStore = create<State>((set) => ({
  sessions: [],
  current: null,
  setSessions: (sessions) => set({ sessions }),
  setCurrent: (current) => set({ current }),
}));
```

- [ ] **Step 2: Author `components/SessionList.tsx`**

```tsx
import { useEffect, useState } from "react";
import { api } from "@/api/client";
import { useSessionStore } from "@/store/session";
import type { components } from "@/api/types.gen";

type Agent = components["schemas"]["Agent"];

export function SessionList() {
  const { sessions, setSessions, setCurrent } = useSessionStore();
  const [agents, setAgents] = useState<Agent[]>([]);
  const [busy, setBusy] = useState(false);

  useEffect(() => { void api.agents().then(setAgents).catch(() => setAgents([])); }, []);

  async function newSession(name: string) {
    setBusy(true);
    try {
      const s = await api.createSession(name);
      setSessions([s, ...sessions]);
      setCurrent(s);
    } finally { setBusy(false); }
  }

  return (
    <section>
      <h2>Sessions</h2>
      <div style={{ display: "flex", gap: 8 }}>
        {agents.map((a) => (
          <button key={a.name} onClick={() => newSession(a.name)} disabled={busy || !a.enabled}>
            New session: {a.name}
          </button>
        ))}
      </div>
      <ul>
        {sessions.map((s) => (
          <li key={s.id}>
            <button onClick={() => setCurrent(s)}>{s.id} · {s.agent_name} · {new Date(s.created_at).toLocaleString()}</button>
          </li>
        ))}
      </ul>
    </section>
  );
}
```

`frontend/src/components/SessionList.test.tsx`:
```tsx
import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, fireEvent, waitFor } from "@testing-library/react";
import { SessionList } from "./SessionList";

describe("SessionList", () => {
  beforeEach(() => {
    vi.stubGlobal("fetch", vi.fn().mockImplementation((url: string, init?: RequestInit) => {
      if (url.endsWith("/api/agents")) {
        return Promise.resolve({ ok: true, json: () => Promise.resolve([{ name: "mock-trino-flavored", version: "0.1.0", enabled: true }]) });
      }
      if (url.endsWith("/api/sessions") && init?.method === "POST") {
        return Promise.resolve({ ok: true, json: () => Promise.resolve({ id: "sess_1", agent_name: "mock-trino-flavored", created_at: new Date().toISOString() }) });
      }
      return Promise.reject(new Error("unexpected " + url));
    }));
  });

  it("creates a new session when an agent button is clicked", async () => {
    render(<SessionList />);
    const btn = await screen.findByRole("button", { name: /new session: mock/i });
    fireEvent.click(btn);
    await waitFor(() => expect(screen.getByText(/sess_1/)).toBeInTheDocument());
  });
});
```

Run: `pnpm test` → PASS.

- [ ] **Step 3: Author `pages/Sessions.tsx`**

```tsx
import { SessionList } from "@/components/SessionList";
import { useSessionStore } from "@/store/session";
import { ChatView } from "@/components/ChatView";

export function Sessions() {
  const current = useSessionStore((s) => s.current);
  return (
    <main style={{ display: "grid", gridTemplateColumns: "320px 1fr", gap: 16, padding: 16 }}>
      <SessionList />
      {current ? <ChatView session={current} /> : <p>Select or create a session.</p>}
    </main>
  );
}
```

(`ChatView` is created in Task 23. The `Sessions` page will fail to compile until then — that's fine, we'll typecheck after Task 23.)

- [ ] **Step 4: Update `App.tsx` to switch on auth**

```tsx
import { Login } from "@/pages/Login";
import { Sessions } from "@/pages/Sessions";
import { useAuthStore } from "@/store/auth";

export function App() {
  const loggedIn = useAuthStore((s) => s.loggedIn);
  return loggedIn ? <Sessions /> : <Login />;
}
export default App;
```

- [ ] **Step 5: Commit (typecheck deferred to Task 23)**

```bash
git add frontend/src/store/session.ts frontend/src/components/SessionList* frontend/src/pages/Sessions.tsx frontend/src/App.tsx
git commit -m "feat(frontend): session list/create + sessions page wiring"
```

---

### Task 23: Chat view + tool-call cards

**Files:**
- Create: `frontend/src/components/MessageBubble.tsx`
- Create: `frontend/src/components/ToolCallCard.tsx`, `..._test.tsx`
- Create: `frontend/src/components/ChatView.tsx`, `..._test.tsx`

- [ ] **Step 1: Author `MessageBubble.tsx`**

```tsx
type Props = { role: "user" | "assistant"; text: string };
export function MessageBubble({ role, text }: Props) {
  const align = role === "user" ? "flex-end" : "flex-start";
  const bg = role === "user" ? "#dbeafe" : "#f3f4f6";
  return (
    <div style={{ display: "flex", justifyContent: align, margin: "4px 0" }}>
      <div style={{ background: bg, padding: "8px 12px", borderRadius: 8, maxWidth: "70%" }}>{text}</div>
    </div>
  );
}
```

- [ ] **Step 2: Author `ToolCallCard.tsx` + test**

```tsx
import { useState } from "react";

type Props = {
  tool: string;
  args?: unknown;
  status: "pending" | "ok" | "error";
  resultPreview?: string;
  errorMessage?: string;
};

export function ToolCallCard({ tool, args, status, resultPreview, errorMessage }: Props) {
  const [open, setOpen] = useState(false);
  const colour = status === "pending" ? "#f59e0b" : status === "ok" ? "#10b981" : "#ef4444";
  return (
    <div style={{ border: `1px solid ${colour}`, borderRadius: 6, margin: "6px 0", padding: "6px 10px", background: "#fafafa" }}>
      <button onClick={() => setOpen(!open)} aria-expanded={open}
              style={{ background: "none", border: "none", padding: 0, fontWeight: 600, cursor: "pointer" }}>
        {open ? "▼" : "▶"} {tool} <span style={{ color: colour }}>· {status}</span>
      </button>
      {open && (
        <div style={{ marginTop: 6 }}>
          {args !== undefined && (
            <details open><summary>args</summary><pre style={{ fontSize: 12 }}>{JSON.stringify(args, null, 2)}</pre></details>
          )}
          {resultPreview && (
            <details open><summary>result</summary><pre style={{ fontSize: 12 }}>{resultPreview}</pre></details>
          )}
          {errorMessage && <div style={{ color: "crimson" }}>{errorMessage}</div>}
        </div>
      )}
    </div>
  );
}
```

`frontend/src/components/ToolCallCard.test.tsx`:
```tsx
import { describe, it, expect } from "vitest";
import { render, screen, fireEvent } from "@testing-library/react";
import { ToolCallCard } from "./ToolCallCard";

describe("ToolCallCard", () => {
  it("renders collapsed by default and expands on click", () => {
    render(<ToolCallCard tool="execute_query" status="ok" args={{ sql: "SELECT 1" }} resultPreview="1 row" />);
    const btn = screen.getByRole("button", { expanded: false });
    expect(screen.queryByText(/SELECT 1/)).toBeNull();
    fireEvent.click(btn);
    expect(screen.getByText(/SELECT 1/)).toBeInTheDocument();
    expect(screen.getByText(/1 row/)).toBeInTheDocument();
  });
});
```

Run tests → PASS.

- [ ] **Step 3: Author `ChatView.tsx`** — the integration component

```tsx
import { useState, useRef } from "react";
import { sendMessage } from "@/api/client";
import { parseSSE, type AgentEvent } from "@/events/parser";
import { MessageBubble } from "./MessageBubble";
import { ToolCallCard } from "./ToolCallCard";
import type { components } from "@/api/types.gen";

type Session = components["schemas"]["Session"];

type Item =
  | { kind: "text"; role: "user" | "assistant"; text: string }
  | { kind: "tool"; call_id: string; tool: string; args?: unknown; status: "pending" | "ok" | "error"; resultPreview?: string; errorMessage?: string };

export function ChatView({ session }: { session: Session }) {
  const [items, setItems] = useState<Item[]>([]);
  const [input, setInput] = useState("");
  const [busy, setBusy] = useState(false);
  const lastAssistantText = useRef<number | null>(null);

  function appendText(role: "user" | "assistant", text: string) {
    setItems((xs) => [...xs, { kind: "text", role, text }]);
    if (role === "assistant") lastAssistantText.current = items.length; // index of new last
  }

  function appendOrUpdateAssistantDelta(delta: string) {
    setItems((xs) => {
      const last = xs[xs.length - 1];
      if (last && last.kind === "text" && last.role === "assistant") {
        return [...xs.slice(0, -1), { ...last, text: last.text + delta }];
      }
      return [...xs, { kind: "text", role: "assistant", text: delta }];
    });
  }

  function applyEvent(ev: AgentEvent) {
    switch (ev.type) {
      case "run_started":  /* noop in UI */ break;
      case "text_delta":   appendOrUpdateAssistantDelta(ev.text); break;
      case "tool_call_start":
        setItems((xs) => [...xs, { kind: "tool", call_id: ev.call_id, tool: ev.tool, args: ev.args, status: "pending" }]);
        break;
      case "tool_call_end":
        setItems((xs) => xs.map((it) => it.kind === "tool" && it.call_id === ev.call_id
          ? { ...it, status: ev.ok ? "ok" : "error", resultPreview: ev.result_preview, errorMessage: ev.error_message }
          : it));
        break;
      case "state_update": /* could surface in a side panel later */ break;
      case "error":        appendText("assistant", `(error: ${ev.code}) ${ev.message}`); break;
      case "run_finished": /* end of stream */ break;
    }
  }

  async function send() {
    if (!input.trim() || busy) return;
    const text = input; setInput(""); setBusy(true);
    appendText("user", text);
    const ctrl = new AbortController();
    try {
      const res = await sendMessage(session.id, text, ctrl.signal);
      if (!res.ok || !res.body) throw new Error(`HTTP ${res.status}`);
      await parseSSE(res.body, applyEvent);
    } catch (e) {
      appendText("assistant", `(stream error: ${(e as Error).message})`);
    } finally {
      setBusy(false);
    }
  }

  return (
    <section style={{ display: "flex", flexDirection: "column", height: "calc(100vh - 32px)" }}>
      <h2>{session.agent_name} — {session.id}</h2>
      <div style={{ flex: 1, overflow: "auto", padding: 4 }} aria-label="message list">
        {items.map((it, i) => it.kind === "text"
          ? <MessageBubble key={i} role={it.role} text={it.text} />
          : <ToolCallCard key={i} tool={it.tool} args={it.args} status={it.status} resultPreview={it.resultPreview} errorMessage={it.errorMessage} />)}
      </div>
      <form onSubmit={(e) => { e.preventDefault(); void send(); }} style={{ display: "flex", gap: 8 }}>
        <input value={input} onChange={(e) => setInput(e.target.value)} placeholder="Ask the agent..." disabled={busy} style={{ flex: 1 }} />
        <button type="submit" disabled={busy || !input.trim()}>{busy ? "..." : "Send"}</button>
      </form>
    </section>
  );
}
```

- [ ] **Step 4: Test ChatView with a fixture stream**

`frontend/src/components/ChatView.test.tsx`:
```tsx
import { describe, it, expect, vi } from "vitest";
import { render, screen, fireEvent, waitFor } from "@testing-library/react";
import { ChatView } from "./ChatView";

function makeStream(events: object[]): Response {
  const body = events.map((e) => `data: ${JSON.stringify(e)}\n\n`).join("");
  const enc = new TextEncoder();
  const stream = new ReadableStream({
    start(c) { c.enqueue(enc.encode(body)); c.close(); },
  });
  return new Response(stream, { status: 200 });
}

describe("ChatView", () => {
  it("renders user message, tool call card, and assistant text from a fixture stream", async () => {
    vi.stubGlobal("fetch", vi.fn().mockImplementation(() => Promise.resolve(makeStream([
      { type: "run_started", run_id: "r1" },
      { type: "text_delta", text: "Looking at " },
      { type: "tool_call_start", call_id: "c1", tool: "execute_query", args: { sql: "SELECT 1" } },
      { type: "tool_call_end", call_id: "c1", ok: true, result_preview: "1 row" },
      { type: "text_delta", text: "the orders table." },
      { type: "run_finished", reason: "done" },
    ]))));

    render(<ChatView session={{ id: "s1", agent_name: "mock-trino-flavored", created_at: new Date().toISOString() } as any} />);
    fireEvent.change(screen.getByPlaceholderText(/ask the agent/i), { target: { value: "go" } });
    fireEvent.click(screen.getByRole("button", { name: /send/i }));

    await waitFor(() => expect(screen.getByText(/the orders table/)).toBeInTheDocument());
    expect(screen.getByRole("button", { name: /execute_query/ })).toBeInTheDocument();
  });
});
```

Run tests → PASS.

- [ ] **Step 5: Typecheck the whole project**

```bash
pnpm typecheck
```
Expected: PASS (Sessions page from Task 22 now compiles since `ChatView` exists).

- [ ] **Step 6: Commit**

```bash
git add frontend/src/components/MessageBubble.tsx frontend/src/components/ToolCallCard* frontend/src/components/ChatView*
git commit -m "feat(frontend): chat view with streamed text + collapsible tool-call cards"
```

---

### Task 24: Audit log view + offline fixture mode

**Files:**
- Create: `frontend/src/components/AuditLogView.tsx`, `..._test.tsx`
- Create: `frontend/src/fixtures/index.ts`
- Modify: `frontend/src/pages/Sessions.tsx` (add audit panel toggle)

- [ ] **Step 1: Author `AuditLogView.tsx`**

```tsx
import { useEffect, useState } from "react";
import { api } from "@/api/client";
import type { components } from "@/api/types.gen";

type AuditEntry = components["schemas"]["AuditEntry"];

export function AuditLogView({ sessionID }: { sessionID: string }) {
  const [rows, setRows] = useState<AuditEntry[]>([]);
  const [err, setErr] = useState<string | null>(null);

  useEffect(() => {
    let alive = true;
    void api.audit(sessionID)
      .then((r) => { if (alive) setRows(r); })
      .catch((e) => { if (alive) setErr(String(e)); });
    return () => { alive = false; };
  }, [sessionID]);

  if (err) return <div role="alert">Audit error: {err}</div>;

  return (
    <details>
      <summary>Audit log ({rows.length})</summary>
      <table style={{ fontSize: 12, width: "100%" }}>
        <thead><tr><th>time</th><th>kind</th><th>run_id</th><th>payload</th></tr></thead>
        <tbody>
          {rows.map((r) => (
            <tr key={r.id}>
              <td>{new Date(r.occurred_at).toLocaleTimeString()}</td>
              <td>{r.kind}</td>
              <td>{r.run_id}</td>
              <td><code>{JSON.stringify(r.payload).slice(0, 80)}…</code></td>
            </tr>
          ))}
        </tbody>
      </table>
    </details>
  );
}
```

`frontend/src/components/AuditLogView.test.tsx`:
```tsx
import { describe, it, expect, vi } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import { AuditLogView } from "./AuditLogView";

describe("AuditLogView", () => {
  it("loads and renders audit entries", async () => {
    vi.stubGlobal("fetch", vi.fn().mockResolvedValue({
      ok: true,
      json: () => Promise.resolve([
        { id: 1, run_id: "r1", kind: "tool_call_start", occurred_at: new Date().toISOString(), payload: { x: 1 } },
      ]),
    }));
    render(<AuditLogView sessionID="s1" />);
    await waitFor(() => expect(screen.getByText("tool_call_start")).toBeInTheDocument());
  });
});
```

Run tests → PASS.

- [ ] **Step 2: Add audit panel to `Sessions.tsx`**

Replace the `Sessions` body with:
```tsx
import { SessionList } from "@/components/SessionList";
import { useSessionStore } from "@/store/session";
import { ChatView } from "@/components/ChatView";
import { AuditLogView } from "@/components/AuditLogView";

export function Sessions() {
  const current = useSessionStore((s) => s.current);
  return (
    <main style={{ display: "grid", gridTemplateColumns: "320px 1fr", gap: 16, padding: 16 }}>
      <SessionList />
      {current ? (
        <div style={{ display: "flex", flexDirection: "column", gap: 8 }}>
          <ChatView session={current} />
          <AuditLogView sessionID={current.id} />
        </div>
      ) : <p>Select or create a session.</p>}
    </main>
  );
}
```

- [ ] **Step 3: Add offline fixture loader (optional dev convenience)**

`frontend/src/fixtures/index.ts`:
```ts
// Imports raw event sequences from the protocol's example fixtures.
// Lets the UI run against canonical event traces without a backend.
import textOnly from "@protocols/examples/text-only-stream.json";
import toolCallSuccess from "@protocols/examples/tool-call-success.json";
import toolCallFailure from "@protocols/examples/tool-call-failure.json";
import multiStep from "@protocols/examples/multi-step-investigation.json";

export const fixtures = {
  textOnly,
  toolCallSuccess,
  toolCallFailure,
  multiStep,
} as const;
```

This is consumed only by ad-hoc local dev scripts; the main app talks to the real backend. Leaving the loader present makes it trivial for future-you to add an "offline mode" flag without scaffolding it.

- [ ] **Step 4: Typecheck and run all tests**

```bash
pnpm typecheck && pnpm test
```
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add frontend/src/components/AuditLogView* frontend/src/fixtures/ frontend/src/pages/Sessions.tsx
git commit -m "feat(frontend): audit log view + offline fixtures import"
```

---

### Phase 0.1b checkpoint

Frontend track is feature-complete for MVP. Smoke-test against the live backend:

```bash
docker compose up -d postgres
cd backend && go run ./cmd/server &
cd ../frontend && pnpm dev &
sleep 2
# In a browser: http://localhost:5173 — log in, click "New session: mock-trino-flavored",
# type "investigate WIN-WS-014", confirm tool-call cards appear and the audit log expands.
kill %1 %2
```

---

## Phase 0.2 — Integration, CI, concept essays, DoD

### Task 25: Playwright golden-path e2e test

**Files:**
- Create: `frontend/playwright.config.ts`, `frontend/e2e/golden-path.spec.ts`
- Modify: `frontend/package.json` (already has `@playwright/test` from Task 18)

- [ ] **Step 1: Install Playwright browsers**

```bash
cd frontend && pnpm playwright install chromium
```

- [ ] **Step 2: Author `playwright.config.ts`**

```ts
import { defineConfig } from "@playwright/test";

export default defineConfig({
  testDir: "./e2e",
  timeout: 30_000,
  fullyParallel: false,
  reporter: [["list"]],
  use: {
    baseURL: "http://localhost:5173",
    trace: "retain-on-failure",
  },
  // CI runs the dev server externally (Postgres + backend + Vite dev) — see Task 27.
  // For local dev you can `pnpm dev` in another terminal and just run `pnpm playwright test`.
});
```

- [ ] **Step 3: Author the golden-path test**

`frontend/e2e/golden-path.spec.ts`:
```ts
import { test, expect } from "@playwright/test";

test("login → create session → send message → see tool-call card → audit", async ({ page }) => {
  await page.goto("/");

  // 1. Login (admin password from env, set by `make admin-password` before running).
  const password = process.env.AIP_E2E_ADMIN_PASSWORD;
  test.skip(!password, "AIP_E2E_ADMIN_PASSWORD not set");
  await page.getByLabel(/password/i).fill(password!);
  await page.getByRole("button", { name: /sign in/i }).click();

  // 2. Create a session against the mock agent.
  await page.getByRole("button", { name: /new session: mock-trino-flavored/i }).click();
  await expect(page.getByPlaceholder(/ask the agent/i)).toBeVisible();

  // 3. Send a message that triggers the trino-investigation scenario.
  await page.getByPlaceholder(/ask the agent/i).fill("investigate WIN-WS-014");
  await page.getByRole("button", { name: /^send$/i }).click();

  // 4. Tool-call card appears.
  const toolBtn = page.getByRole("button", { name: /describe_table/ }).first();
  await expect(toolBtn).toBeVisible({ timeout: 10_000 });

  // 5. Final assistant text mentions the suspicious PowerShell.
  await expect(page.getByText(/suspicious PowerShell|03:14/i)).toBeVisible({ timeout: 10_000 });

  // 6. Audit log expands and shows entries.
  await page.getByText(/audit log/i).click();
  await expect(page.locator("table tbody tr").first()).toBeVisible();
});
```

- [ ] **Step 4: Run it locally**

In separate terminals (or background) run Postgres, backend, and the dev frontend. Then:

```bash
export AIP_E2E_ADMIN_PASSWORD=hunter2
cd frontend && pnpm playwright test
```

Expected: 1 passed.

- [ ] **Step 5: Commit**

```bash
git add frontend/playwright.config.ts frontend/e2e/
git commit -m "test(e2e): add Playwright golden-path test"
```

---

### Task 26: Pre-commit hook + scripts/

**Files:**
- Create: `scripts/pre-commit`

- [ ] **Step 1: Author the hook**

`scripts/pre-commit`:
```bash
#!/usr/bin/env bash
# Pre-commit: lint + tests + codegen sync. Runs `make verify`.
set -e
echo "[pre-commit] make verify"
make verify
```

- [ ] **Step 2: Install it**

```bash
make hooks
```

(That target copies `scripts/pre-commit` to `.git/hooks/pre-commit` and makes it executable. The Makefile target was authored in Task 1.)

- [ ] **Step 3: Smoke-test by attempting an empty-change commit**

```bash
git commit --allow-empty -m "chore: smoke pre-commit hook"
```

Expected: prints `[pre-commit] make verify`, runs lint+tests, and either creates the commit or fails the commit if `make verify` fails. Either is acceptable proof the hook is wired; if it fails, fix the underlying issue before continuing.

- [ ] **Step 4: Commit the hook script**

```bash
git add scripts/pre-commit
git commit -m "chore: add pre-commit hook script (invokes make verify)"
```

---

### Task 27: GitHub Actions CI

**Files:**
- Create: `.github/workflows/ci.yml`

- [ ] **Step 1: Author `.github/workflows/ci.yml`** (parallel jobs + sequential e2e)

```yaml
name: CI

on:
  pull_request:
    branches: [main]
  push:
    branches: [main]

jobs:
  backend:
    runs-on: ubuntu-latest
    services:
      postgres:
        image: postgres:16-alpine
        env:
          POSTGRES_USER: aip
          POSTGRES_PASSWORD: aip
          POSTGRES_DB: aip
        ports: ["5432:5432"]
        options: >-
          --health-cmd "pg_isready -U aip"
          --health-interval 5s --health-timeout 5s --health-retries 10
    env:
      POSTGRES_HOST: localhost
      POSTGRES_USER: aip
      POSTGRES_PASSWORD: aip
      POSTGRES_DB: aip
      POSTGRES_PORT: "5432"
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with: { go-version: "1.22.x" }
      - name: Apply schema
        run: psql "postgres://aip:aip@localhost:5432/aip" -f infra/postgres/init.sql
      - name: Install golangci-lint
        run: |
          curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | \
            sh -s -- -b $(go env GOPATH)/bin v1.61.0
      - name: Lint
        working-directory: backend
        run: go vet ./... && golangci-lint run
      - name: Test
        working-directory: backend
        run: go test ./...

  frontend:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: pnpm/action-setup@v4
        with: { version: 9 }
      - uses: actions/setup-node@v4
        with: { node-version: "20", cache: "pnpm", cache-dependency-path: frontend/pnpm-lock.yaml }
      - name: Install
        working-directory: frontend
        run: pnpm install --frozen-lockfile
      - name: Lint + typecheck
        working-directory: frontend
        run: pnpm lint && pnpm typecheck
      - name: Test
        working-directory: frontend
        run: pnpm test
      - name: Build
        working-directory: frontend
        run: pnpm build

  contract:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with: { go-version: "1.22.x" }
      - uses: pnpm/action-setup@v4
        with: { version: 9 }
      - uses: actions/setup-node@v4
        with: { node-version: "20" }
      - uses: actions/setup-python@v5
        with: { python-version: "3.12" }
      - name: Install jsonschema
        run: pip install jsonschema
      - name: Install frontend deps (for codegen)
        working-directory: frontend
        run: pnpm install --frozen-lockfile
      - name: Run codegen
        run: bash protocols/gen-types.sh
      - name: Detect drift
        run: |
          git diff --exit-code -- '*.gen.ts' '*.gen.go' \
            backend/internal/protocol/events.schema.embed.json \
            || (echo "Generated files out of sync. Run 'make gen' and commit."; exit 1)

  e2e:
    needs: [backend, frontend, contract]
    runs-on: ubuntu-latest
    services:
      postgres:
        image: postgres:16-alpine
        env: { POSTGRES_USER: aip, POSTGRES_PASSWORD: aip, POSTGRES_DB: aip }
        ports: ["5432:5432"]
        options: >-
          --health-cmd "pg_isready -U aip"
          --health-interval 5s --health-timeout 5s --health-retries 10
    env:
      POSTGRES_HOST: localhost
      POSTGRES_USER: aip
      POSTGRES_PASSWORD: aip
      POSTGRES_DB: aip
      POSTGRES_PORT: "5432"
      JWT_SIGNING_KEY: dev-only-replace-32-bytes-random-please
      ADMIN_USERNAME: admin
      AIP_E2E_ADMIN_PASSWORD: hunter2
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with: { go-version: "1.22.x" }
      - uses: pnpm/action-setup@v4
        with: { version: 9 }
      - uses: actions/setup-node@v4
        with: { node-version: "20", cache: "pnpm", cache-dependency-path: frontend/pnpm-lock.yaml }
      - run: psql "postgres://aip:aip@localhost:5432/aip" -f infra/postgres/init.sql
      - name: Compute admin password hash
        run: |
          cp .env.example .env
          (cd backend && go run ./cmd/admin-password <<< "$AIP_E2E_ADMIN_PASSWORD")
          # admin-password wrote ADMIN_PASSWORD_HASH=<hash> into ../.env; export it for the
          # subsequent `go run ./cmd/server` step.
          set -a; . ./.env; set +a
          echo "ADMIN_PASSWORD_HASH=$ADMIN_PASSWORD_HASH" >> "$GITHUB_ENV"
      - name: Start backend
        working-directory: backend
        run: |
          go run ./cmd/server &
          for i in $(seq 1 30); do curl -fs http://localhost:8080/api/healthz && break; sleep 1; done
      - name: Install frontend
        working-directory: frontend
        run: pnpm install --frozen-lockfile
      - name: Install Playwright browsers
        working-directory: frontend
        run: pnpm playwright install --with-deps chromium
      - name: Start dev server
        working-directory: frontend
        run: |
          pnpm dev &
          for i in $(seq 1 30); do curl -fs http://localhost:5173 && break; sleep 1; done
      - name: Run Playwright
        working-directory: frontend
        run: pnpm playwright test
```

The CI step writes a fresh `.env` from `.env.example`, runs the `admin-password` helper (which writes `ADMIN_PASSWORD_HASH=<hash>` into that `.env`), then sources `.env` and exports `ADMIN_PASSWORD_HASH` to `GITHUB_ENV` so the next step (which starts the backend) sees it as an environment variable.

- [ ] **Step 2: Push to a branch and verify CI**

If working in a git remote, push and confirm all jobs go green:

```bash
git checkout -b ci/initial
git add .github/workflows/ci.yml
git commit -m "ci: initial parallel + e2e workflow"
git push -u origin ci/initial
# open a PR, watch the checks
```

If working purely locally for now, just commit and validate later.

- [ ] **Step 3: Commit**

```bash
git add .github/workflows/ci.yml
git commit -m "ci: parallel backend/frontend/contract jobs + sequential e2e"
```

---

### Task 28: Concept essay 01 — the agent loop

> **Per spec §7.3:** the user writes the prose. AI assistants may write outlines and pose questions. This task lays the outline; the user fills in the experience-grounded prose after the implementation works end-to-end.

**Files:**
- Create: `docs/concepts/01-the-agent-loop.md`
- Modify: `docs/README.md` (un-stub the concept link)

- [ ] **Step 1: Author the outline of `docs/concepts/01-the-agent-loop.md`**

```markdown
# 01 — The agent loop

> Written *after* shipping Phase 0. This essay is for future-you. It is NOT an
> introduction to LLMs; it's "what I now understand about the loop after seeing
> it run end-to-end in our system."

## What is the agent loop, in one sentence?

*(your one-sentence answer)*

## Where it shows up in this codebase

- `backend/internal/adapters/adapter.go` — the `AgentPlatformAdapter.Run` method is the type-level expression of the loop: input messages in, stream of typed events out.
- `backend/internal/adapters/mock/mock.go` — a *scripted* loop. No LLM, no decision-making, but the *shape* is exactly what real platforms produce.
- `backend/internal/httpapi/messages_handler.go` — the orchestration around one Run: persist input, open SSE, forward events, audit, persist output.

## Three things that surprised me

1. *(your observation 1)*
2. *(your observation 2)*
3. *(your observation 3)*

## How "tool use" sits inside the loop

*(short paragraph on `tool_call_start`/`tool_call_end` pairing, what `result_preview` is for, why the LLM doesn't see the full result, …)*

## Open questions I still have

- *(question 1)*
- *(question 2)*
```

- [ ] **Step 2: Update `docs/README.md`** to link the essay

Replace the stub line with a real link in the "Concepts, in the order they were learned" section:
```markdown
1. [The agent loop](concepts/01-the-agent-loop.md) — Phase 0
```

- [ ] **Step 3: Commit (the prose can land in a follow-up if you want time to think)**

```bash
git add docs/concepts/01-the-agent-loop.md docs/README.md
git commit -m "docs(concepts): outline 01-the-agent-loop (prose to follow)"
```

---

### Task 29: Concept essay 02 — streaming protocols

**Files:**
- Create: `docs/concepts/02-streaming-protocols.md`
- Modify: `docs/README.md` (link the concept)

- [ ] **Step 1: Author the outline**

```markdown
# 02 — Streaming protocols

> Written *after* shipping Phase 0. Covers what I learned about streaming agent
> events end-to-end through a real system.

## SSE versus WebSockets, restated for our case

*(one paragraph on why SSE fits one-way agent → client streams; what we'd lose if we'd picked WebSockets)*

## Why a discriminated union of events, not raw text

- `text_delta` for text
- `tool_call_start`/`tool_call_end` for tool use (the user *sees what the agent did*)
- `state_update` for everything else

*(your prose on what changes when the wire is typed events vs unstructured tokens)*

## The contract chain

```
adapter (translates) ──► canonical event ──► backend (validates + audits) ──► SSE wire ──► frontend parser ──► UI
```

Each hop is a chance for the contract to drift. We chose:

- JSON Schema validation in the backend (`backend/internal/protocol/validate.go`) at runtime in dev/test — adapters that emit malformed events fail loudly.
- Code-generated TS types from the same schema — the frontend can't accidentally mis-handle a field.
- Conformance YAML scenarios in `backend/internal/adapters/conformance/` — every adapter must pass them.

*(your prose on which of these caught a real bug during Phase 0)*

## What was harder than expected

- *(observation)*

## What would I have done differently knowing what I know now?

*(your reflection)*
```

- [ ] **Step 2: Update `docs/README.md`**

Replace the stub line with:
```markdown
2. [Streaming protocols](concepts/02-streaming-protocols.md) — Phase 0
```

- [ ] **Step 3: Commit**

```bash
git add docs/concepts/02-streaming-protocols.md docs/README.md
git commit -m "docs(concepts): outline 02-streaming-protocols (prose to follow)"
```

---

### Task 30: Definition-of-done verification + tag v0.0.0

**Files:** none (verification + tag)

- [ ] **Step 1: Run the merge gate**

```bash
make verify
```
Expected: zero errors. Lint passes for both services. All Go and Vitest tests pass. Codegen produces no diff.

- [ ] **Step 2: Run e2e**

```bash
docker compose up -d postgres
make admin-password   # ensures ADMIN_PASSWORD_HASH is set in .env
cd backend && go run ./cmd/server &
cd ../frontend && pnpm dev &
sleep 3
export AIP_E2E_ADMIN_PASSWORD=hunter2
cd frontend && pnpm playwright test
kill %2 %1
```
Expected: 1 passed.

- [ ] **Step 3: Manual DoD walkthrough** (matches spec §10.1 Phase 0 DoD verbatim)

Open a browser, log in, and verify each line of the spec's Phase 0 Definition-of-Done:

- [ ] `make up` brings the stack up; browse to `http://localhost:5173`, log in as admin, see `mock-trino-flavored` listed.
- [ ] Start a session, send a message. SSE stream returns canned events that look like a Trino agent ran (`tool_call_start{tool:"execute_query"}` → `tool_call_end{result_preview:"…"}` → `text_delta` → `run_finished`).
- [ ] Frontend renders user text **and** tool calls as collapsible cards.
- [ ] Audit log endpoint shows the same tool-call events, persisted in Postgres.
- [ ] CI green: lint, types, unit tests, conformance, contract (codegen-up-to-date), one Playwright e2e.
- [ ] Concept essay outlines at `docs/concepts/01-the-agent-loop.md` and `docs/concepts/02-streaming-protocols.md`.
- [ ] ADR files: 0001 (three-tier-with-go-gateway), 0002 (mock-first-then-real-platform), 0003 (wire protocol from spike) all present.

- [ ] **Step 4: Tag v0.0.0**

```bash
git tag -a v0.0.0 -m "Phase 0 — MVP skeleton"
git push origin v0.0.0   # if a remote is configured
```

- [ ] **Step 5: Final commit (concept-essay prose)** *(optional, may be done in a separate session)*

Replace the placeholder `(your one-sentence answer)` / `(your prose on …)` blocks in `01-the-agent-loop.md` and `02-streaming-protocols.md` with the user's actual reflections. When both have substantive prose:

```bash
git add docs/concepts/
git commit -m "docs(concepts): write essays 01 and 02 — Phase 0 retrospective"
```

---

## Phase 0 — done

The repo now satisfies every line of spec §10.1 Phase 0 Definition-of-Done. Phase 1 (Trino tool-use POC + GoClaw adapter) gets its own implementation plan when the user is ready to start it; nothing in Phase 0 is undone or revisited along the way.