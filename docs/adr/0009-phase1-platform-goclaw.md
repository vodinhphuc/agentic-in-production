# ADR-0009: Phase 1 platform = GoClaw

**Status:** Accepted | **Date:** 2026-05-01

## Context
First non-Mock adapter (Phase 1). Candidate list was Dify, GoClaw, NemoClaw, OpenClaw.

## Decision
GoClaw is the first real Agent Platform integration in Phase 1. Other platforms become later additions (each must pass the Phase 0 conformance suite).

## Consequences
- `backend/internal/adapters/goclaw/` is built in Phase 1.
- Other adapter folders remain placeholders.
- GoClaw-vs-alternatives rationale captured here when implementation lands.
