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
