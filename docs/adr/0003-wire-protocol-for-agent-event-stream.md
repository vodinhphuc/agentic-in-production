# ADR-0003: Wire protocol for agent event stream

**Status:** Accepted | **Date:** 2026-05-02

## Context

We need to commit to a wire protocol for the canonical streaming-event envelope (Backend ↔ Frontend, Adapter ↔ Backend). Candidates considered in the Phase 0.0.a research spike:

- **AG-UI** (CopilotKit) — open-source agent-UI event protocol, published under github.com/copilotkit
- **A2UI** — agent-to-UI protocol; less mature documentation surface
- **MCP-shaped events** — Model Context Protocol; designed for tool/resource exchange, not for streaming agent reasoning
- **OpenAI Realtime events** — vendor-specific (OpenAI), well-specified but tied to OpenAI's roadmap
- **Vendor-native** — GoClaw native, Dify SSE, etc., per-platform shapes
- **Custom v1** — the AG-UI-aligned minimal envelope drafted in [the design spec §4.2](../superpowers/specs/2026-05-01-agentic-platform-design.md)

Scoring (1–5; for "translation distance" lower is better):

| Candidate | Doc maturity | Adoption | Vocabulary fit | Transport flexibility | Generative-UI ready | Translation distance from v1 |
|---|---|---|---|---|---|---|
| AG-UI (direct) | 4 | 3 | 5 | 5 | 5 | 1 |
| A2UI | 2 | 1 | 3 (best-guess) | ? | ? | 3 |
| MCP-shaped | 5 | 5 | 1 (wrong shape) | 4 | 1 | 5 |
| OpenAI Realtime | 5 | 5 | 4 | 3 (WS-only) | 3 | 4 |
| Vendor-native | 3 | n/a | 4 | varies | varies | 4–5 |
| **Custom v1 (this repo)** | n/a | n/a | 5 | 5 | 4 | **0** |

The scores reflect what is publicly documented and our project's needs at Phase 0; they are not the output of a deep multi-day investigation but a calibrated read of the candidates against [the design spec](../superpowers/specs/2026-05-01-agentic-platform-design.md). Re-score before any breaking protocol change.

## Decision

**Outcome A — keep the AG-UI-aligned custom v1 as the wire format.**

The schema in [`protocols/agent-events.schema.json`](../../protocols/agent-events.schema.json) is v1, frozen until a v2 ADR. Events: `run_started`, `text_delta`, `tool_call_start`, `tool_call_end`, `state_update`, `error`, `run_finished`.

AG-UI direct adoption is the **migration target**: translation distance is ~1 (rename a few fields, add optional metadata). A2UI requires further investigation before we can evaluate it; until that investigation lands we stay translation-ready and avoid platform-specific semantics in the canonical schema.

## Consequences

- The schema is the source of truth. Adapters translate platform-native events into it.
- Migration to AG-UI direct (or A2UI) becomes a labelled v2 if/when triggered, with v1 retained on `/api/v1/...` for at least one phase ([protocol versioning policy](../../protocols/README.md)).
- Vendor-native and MCP-shaped paths are not pursued — the former locks the frontend to a platform; the latter is the wrong shape for streaming reasoning.
- Future re-investigation is gated on a concrete trigger (e.g., a second non-frontend consumer, or a CopilotKit-built UI that we want to drop in directly).
