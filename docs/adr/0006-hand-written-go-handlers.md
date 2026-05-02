# ADR-0006: Hand-written Go handlers, not generated from OpenAPI

**Status:** Accepted | **Date:** 2026-05-01

## Context
With one developer, fighting `oapi-codegen`'s generated middleware exceeds the duplication cost of writing handlers by hand.

## Decision
OpenAPI is the source of truth for **types only** (request/response models in Go and TS). HTTP handlers are written by hand. Per-route integration tests catch contract drift.

## Consequences
Reconsider when team grows past 2 backend devs or API surface exceeds ~30 routes.
