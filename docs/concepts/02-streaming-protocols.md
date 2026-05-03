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
