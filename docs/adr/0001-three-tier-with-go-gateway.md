# ADR-0001: Three-tier architecture with a Go gateway

**Status:** Accepted | **Date:** 2026-05-01

## Context
We want frontend stable across changing AI agent platforms, server-side audit of every tool call, centralised auth. Two-tier (browser → platform) puts LLM credentials in the client, has no audit trail, and forces a frontend redeploy whenever the platform changes.

## Decision
Three services: React frontend, Go backend (gateway with adapter layer), external Agent Platform. Frontend never talks to the platform directly.

## Consequences
- Backend owns auth, audit, rate limiting, agent registry, adapter translation.
- Switching platforms = new adapter folder; no frontend code moves.
- One extra hop per request — acceptable for a single-tenant local app.
