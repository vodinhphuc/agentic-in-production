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
