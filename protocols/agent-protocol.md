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
