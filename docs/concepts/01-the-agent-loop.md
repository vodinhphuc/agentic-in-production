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
