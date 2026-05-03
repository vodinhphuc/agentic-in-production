# Onboarding — clone to first green Playwright

What to do on a fresh dev box (or a new contributor's machine) to get from `git clone` to a passing end-to-end run. Read [dev-environment.md](dev-environment.md) for the tool-install rationale; this doc is the *sequence*.

## 0. Prerequisites — install the toolchain

Follow [dev-environment.md](dev-environment.md) sections 1–7 first. Verify:

```bash
make --version           # 4.x
go version               # 1.22+
node --version           # v20+
pnpm --version
docker --version
golangci-lint --version
air -v
```

If any of those fail, fix the install before continuing.

## 1. Clone

```bash
git clone git@github.com:vodinhphuc/agentic-in-production.git
cd agentic-in-production
git checkout phase-0     # or `git checkout v0.0.0` for the tagged Phase-0 snapshot
```

## 2. Secrets — make `.env` from the example

`.env` is gitignored (it holds the bcrypt admin hash and the JWT signing key). Copy the template, then bake a real password hash:

```bash
cp .env.example .env
# rotate JWT_SIGNING_KEY to a fresh 32+ byte random value:
sed -i.bak "s|^JWT_SIGNING_KEY=.*|JWT_SIGNING_KEY=$(openssl rand -hex 32)|" .env && rm .env.bak

cd backend && go run ./cmd/admin-password    # type a password (e.g. hunter2). It writes ADMIN_PASSWORD_HASH='...' to ../.env
cd ..
```

The `admin-password` helper single-quotes the bcrypt hash so it survives `set -a; source .env`. If you skip this, login returns 401.

## 3. Install deps + hooks

```bash
make install   # go mod download + pnpm install (a few minutes the first time)
make hooks     # copies scripts/pre-commit into .git/hooks (opt-in per clone)

cd frontend && pnpm playwright install chromium && cd ..    # ~110MB browser download
```

## 4. Stack up

```bash
make up                  # docker compose up -d postgres (applies infra/postgres/init.sql on first run)
docker compose ps        # postgres should be (healthy)
```

If you ever need to reseed the database (e.g. you changed `init.sql`):

```bash
docker compose down -v && docker compose up -d postgres
```

## 5. First green verify

```bash
make verify   # lint + tests + codegen drift check
```

This is the merge gate. It should be green on a fresh clone.

## 6. Run the stack manually

Three processes — easiest in three terminals (or background them with `nohup`):

```bash
# Terminal 1 — backend (must run from backend/ for mock scenario_dir to resolve)
cd backend
set -a; source ../.env; set +a
make dev-backend          # uses air for hot reload, OR: go run ./cmd/server

# Terminal 2 — frontend
cd frontend
pnpm dev                  # http://localhost:5173

# Terminal 3 — browse
xdg-open http://localhost:5173   # log in with admin / <the password you typed in step 2>
```

## 7. Run the Playwright golden-path

With backend + Vite up:

```bash
cd frontend
AIP_E2E_ADMIN_PASSWORD=<the password you typed in step 2> pnpm playwright test
```

Expected: 1 passed in ~10–20s.

## What's *not* in the repo and you'll need to recreate

| File / dir | Why missing | How to recreate |
|---|---|---|
| `.env` | gitignored — secrets | Step 2 above |
| `backend/go.sum` cache | local | `go mod download` (in `make install`) |
| `frontend/node_modules/` | gitignored | `pnpm install` (in `make install`) |
| `frontend/dist/` | gitignored — build artifact | `pnpm build` |
| `backend/tmp/` | gitignored — air's hot-reload binary | regenerated when `air` runs |
| `tmp/aip-server` (operator artifact) | local | `cd backend && go build -o /tmp/aip-server ./cmd/server` |
| Docker images | per-host cache | `make up` pulls them |
| Playwright Chromium | `~/.cache/ms-playwright/` | `pnpm playwright install chromium` |
| `.git/hooks/pre-commit` | hooks aren't cloned by git | `make hooks` |

## Where to find what

- [README.md (root)](../README.md) — project elevator pitch
- [docs/README.md](README.md) — the learning index (start here for concepts)
- [docs/dictionary.md](dictionary.md) — terms with project-specific meaning
- [docs/dev-environment.md](dev-environment.md) — toolchain rationale + install
- [docs/adr/](adr/) — every load-bearing decision
- [docs/superpowers/specs/](superpowers/specs/) — the foundational design spec
- [docs/superpowers/plans/](superpowers/plans/) — the Phase 0 implementation plan
- [protocols/](../protocols/) — wire contract (JSON Schema + OpenAPI). Source of truth.
- [backend/CLAUDE.md](../backend/CLAUDE.md), [frontend/CLAUDE.md](../frontend/CLAUDE.md) — sub-tree conventions

## When something breaks on the new machine

- **Login returns 401** → the bcrypt hash in `.env` got mangled by shell expansion. Re-run `make admin-password`; it now writes single-quoted values. Confirm with `set -a; source .env; set +a; echo "$ADMIN_PASSWORD_HASH"` — should still start with `$2a$10$`.
- **Backend errors `no scenarios in internal/adapters/mock/scenarios`** → server was started from the wrong cwd. The mock adapter's `scenario_dir` config is relative; always launch from `backend/`.
- **`go mod tidy` bumps dep versions or the `go` directive** → re-pin to the plan's versions: `go get github.com/jackc/pgx/v5@v5.7.0 github.com/stretchr/testify@v1.9.0 golang.org/x/text@v0.18.0 golang.org/x/sync@v0.7.0` then `go mod edit -go=1.22 -toolchain=none && go mod tidy`.
- **`pkill aip-server` then `kill $!` doesn't free port 8080** → `go run ./cmd/server &` spawns a child binary; `kill $!` kills the `go` wrapper, not the server. Build to `/tmp/aip-server` and run the binary, then `pkill -x aip-server`.
- **Vitest collects the Playwright `e2e/` files and fails** → `vite.config.ts` already excludes `e2e/**`; if you add a new e2e dir, exclude it too.
- **`pnpm install` warns about ignored build scripts (esbuild)** → run `pnpm rebuild esbuild` once to drop the platform binary; otherwise `pnpm dev` won't start.
