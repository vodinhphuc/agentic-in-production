# ADR-0002: Mock-first MVP, real platform deferred to Phase 1

**Status:** Accepted | **Date:** 2026-05-01

## Context
Real platforms add hosting, auth credentials, and platform-specific quirks — friction that would slow down validation of the architecture.

## Decision
Phase 0 ships with **only** the Mock adapter (in-process, scripted scenarios). Phase 1 adds the first real platform (GoClaw — see ADR-0009).

## Consequences
- The mock adapter is permanent infrastructure. It runs in CI forever as the conformance baseline.
- Frontend can be built fully against scripted events.
- "Done" for Phase 0 does not require a live LLM.
