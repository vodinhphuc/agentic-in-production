# ADR-0007: docs/dictionary.md as a first-class shared-vocabulary doc

**Status:** Accepted | **Date:** 2026-05-01

## Context
Silent vocabulary mismatch (user and AI coding agent meaning different things by "agent", "session", "adapter") is one of the most expensive classes of misunderstanding.

## Decision
`docs/dictionary.md` has both user and AI coding agents as primary audience. Root `CLAUDE.md` requires consulting it. Entries are added the moment a misunderstanding surfaces.

## Consequences
- Maintenance is reactive — entries land at the moment of evidence, not from a guessed-up-front list.
- All AI agent guidance files (CLAUDE.md, AGENTS.md) must point to it.
