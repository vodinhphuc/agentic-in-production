# ADR-0004: Stateful agent platform; backend stores only audit + session mapping

**Status:** Accepted | **Date:** 2026-05-01

## Context
Real Agent Platforms own conversation history. Duplicating in our backend would cause drift.

## Decision
Platform owns conversation state. Backend stores: `session_id ↔ platform_conversation_id` mapping; its own audit log of every tool call observed; per-platform connection settings. Backend sends only the new user message on each Run.

## Consequences
- Memory (Phase 4) becomes a backend concern that injects synthesised history at Run time, leaving the platform contract unchanged.
- Mock adapter must also be stateful to faithfully simulate real platforms.
