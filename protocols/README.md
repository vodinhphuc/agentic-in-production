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
