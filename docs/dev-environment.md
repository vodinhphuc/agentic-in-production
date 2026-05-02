# Development environment

What's installed on a working dev machine, why each piece is here, and how to rebuild from scratch.

Verified on **Ubuntu 24.04 (Noble Numbat)**, including WSL2. macOS and other Linux distros work but commands differ — adapt the package manager. The toolchain itself (Go, Node, Python) is portable.

## At a glance

| Tool | Version | Purpose | Source of truth |
|---|---|---|---|
| **Python 3** | 3.13+ | JSON Schema validation in `protocols/gen-types.sh`; small dev scripts. | system |
| `jsonschema` | 4.x | Validates protocol example fixtures and the schema itself. | pip |
| `pyyaml` | latest | Parses `protocols/openapi.yaml` for sanity checks. | pip |
| **Go** | 1.22+ | Backend service language. Plus `go run` invocations for codegen tools. | apt |
| **Node.js** | 20+ | Frontend runtime + ecosystem (Vite, Vitest, Playwright). Also hosts `npx`-based codegen tools. | NodeSource apt repo |
| **pnpm** | latest | Frontend package manager (chosen for disk efficiency + strict node_modules). | corepack (bundled with Node) |
| **golangci-lint** | 1.61.0 | Backend linter, runs in `make verify`. | github install script |
| **air** | latest | Hot-reload for `make dev-backend`. | `go install` |

Why these specific versions:

- **Go 1.22** — required by the project's Go modules and matches Ubuntu 24.04's apt default; no need to pin separately.
- **Node 20** — current LTS line; Vite 5+ and Playwright require Node 18+, but several of our codegen tools (`json-schema-to-typescript@15`, `openapi-typescript@7`) prefer Node 20+.
- **pnpm via corepack** — corepack ships with Node 20 and pins pnpm per-project via `package.json`'s `packageManager` field once we have a `frontend/package.json`. No separate `npm install -g` needed.

## Install steps (Ubuntu 24.04)

These are exactly the commands a fresh machine needs. Each block is independent — run them in order.

### 1. Python + JSON tooling

Python 3 ships with Ubuntu. Add the two libs:

```bash
pip install --user jsonschema pyyaml
```

Verify:

```bash
python3 -c "import jsonschema, yaml; print('ok')"
```

### 2. Go 1.22+

```bash
sudo apt update
sudo apt install -y golang-go
go version   # expect: go1.22.x or newer
```

### 3. Node.js 20+ via NodeSource

Ubuntu's apt-shipped `nodejs` is too old (v18 line). Use NodeSource's repo:

```bash
curl -fsSL https://deb.nodesource.com/setup_20.x | sudo -E bash -
sudo apt install -y nodejs
node --version   # expect: v20.x
npm --version
```

What `sudo -E bash -` does: `-E` preserves your env vars across the sudo boundary; the trailing `-` tells bash to read its script from stdin (which is where the piped `curl` output lands).

### 4. pnpm via corepack

```bash
sudo corepack enable
corepack prepare pnpm@latest --activate
pnpm --version
```

### 5. Verify the codegen toolchain

These probes don't install anything globally — they're the same `npx --yes` and `go run @version` invocations that `protocols/gen-types.sh` uses, so success here means the codegen pipeline will work:

```bash
npx --yes json-schema-to-typescript@15 --help | head -3
npx --yes openapi-typescript@7 --help | head -3
go run github.com/atombender/go-jsonschema@v0.18.0 --help | head -3
go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@v2.4.1 --help | head -3
```

### 6. Optional — only needed for `make verify` and `make dev-backend`

`golangci-lint` (used by `make lint-backend`):

```bash
curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh \
  | sh -s -- -b $(go env GOPATH)/bin v1.61.0
```

`air` (used by `make dev-backend` for hot reload):

```bash
go install github.com/air-verse/air@latest
```

Both install into `$(go env GOPATH)/bin`. Add it to your `PATH` if not already:

```bash
echo 'export PATH=$PATH:$(go env GOPATH)/bin' >> ~/.bashrc
source ~/.bashrc
```

## Tech stack — why each piece is here

Cross-references to the design spec and ADRs so you can drill in:

### Protocol layer (`protocols/`)
- **JSON Schema 2020-12** for the streaming event envelope ([ADR-0005](adr/0005-json-schema-not-asyncapi-yet.md)) — chosen over AsyncAPI because we have one consumer (the frontend) and don't need AsyncAPI's ceremony yet.
- **OpenAPI 3.1** for REST endpoints — types-only codegen ([ADR-0006](adr/0006-hand-written-go-handlers.md)); HTTP handlers are written by hand because fighting `oapi-codegen`'s generated middleware costs more than the duplication for a single-developer project.
- Codegen pipeline: `json-schema-to-typescript` → discriminated TS unions for events; `openapi-typescript` → TS types for REST; `go-jsonschema` → Go structs for events; `oapi-codegen` (types-only) → Go structs for REST.

### Backend (`backend/`, Go)
- **chi** router — smallest, stdlib-shaped HTTP router. Matches the team's preference for thin abstractions.
- **pgx/v5** — Postgres-native driver, faster and more featureful than `database/sql`'s lib/pq.
- **testify** — standard Go test assertion library.
- **santhosh-tekuri/jsonschema/v5** — runtime JSON Schema validator that supports 2020-12 (the draft we use).
- **air** — hot reload during `make dev-backend`.

### Frontend (`frontend/`, React + TypeScript)
- **Vite** — dev server + build tool. Native TS, native ESM, fast HMR.
- **Zustand** — minimal global state; no Redux ceremony for a Phase 0 app.
- **Vitest** — Vite-native unit test runner.
- **Playwright** — golden-path e2e (one test in Phase 0).

### Infrastructure
- **Postgres 16** in Docker — sessions, audit log, agent registry. Phase 1 adds Trino + Hive metastore.

## Rebuild from a clean machine

For a one-shot rebuild after wiping a dev box, run sections 1–5 above in order, then in the repo root:

```bash
cp .env.example .env
make install      # go mod download + pnpm install
make hooks        # installs the pre-commit hook
make up           # starts Postgres in Docker
make verify       # lint + tests + codegen sync — should pass green
```

## Adding a new tool to this list

When you install something the project depends on, append a row to the **At a glance** table and a short rationale paragraph below. The bar is "future me needs to rebuild this exact stack" — names + versions + one sentence of *why*.
