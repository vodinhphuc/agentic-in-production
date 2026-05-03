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
