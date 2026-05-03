# Learning index

Read this when you've lost the thread. It maps the project's concepts to the
experiments that drove them and the code that implements them.

## Foundational reading
- [design spec](superpowers/specs/2026-05-01-agentic-platform-design.md)
- [dictionary](dictionary.md) — shared vocabulary
- [onboarding](onboarding.md) — clone → first green Playwright on a new dev box
- [dev-environment](dev-environment.md) — toolchain rationale + install

## Concepts, in the order they were learned
*(outlines authored at Phase 0 close; prose follows in a retrospective pass)*
1. [The agent loop](concepts/01-the-agent-loop.md) — Phase 0
2. [Streaming protocols](concepts/02-streaming-protocols.md) — Phase 0

## Decisions
- [ADR-0001 — Three-tier with Go gateway](adr/0001-three-tier-with-go-gateway.md)
- [ADR-0002 — Mock-first MVP](adr/0002-mock-first-then-real-platform.md)
- [ADR-0003 — Wire protocol](adr/0003-wire-protocol-for-agent-event-stream.md)
- [ADR-0004 — Stateful agent platform](adr/0004-stateful-agent-platform.md)
- [ADR-0005 — JSON Schema (not AsyncAPI yet)](adr/0005-json-schema-not-asyncapi-yet.md)
- [ADR-0006 — Hand-written Go handlers](adr/0006-hand-written-go-handlers.md)
- [ADR-0007 — dictionary.md as first-class doc](adr/0007-dictionary-md-as-first-class-doc.md)
- [ADR-0008 — OCSF toy dataset](adr/0008-ocsf-toy-dataset.md)
- [ADR-0009 — Phase 1 platform = GoClaw](adr/0009-phase1-platform-goclaw.md)

## Concepts → code (lookup table — grows as code lands)
- "How does streaming work?" → `backend/internal/httpapi/messages_handler.go`
- "Where do tool calls get audited?" → `backend/internal/audit/`
- "What's the canonical event shape?" → `protocols/agent-events.schema.json`
