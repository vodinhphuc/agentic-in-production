# ADR-0005: JSON Schema for events (not AsyncAPI yet)

**Status:** Accepted | **Date:** 2026-05-01

## Context
For the streaming event envelope: AsyncAPI vs plain JSON Schema. AsyncAPI has more ceremony than we currently need.

## Decision
v1 of the event envelope is described by JSON Schema 2020-12. OpenAPI 3.1 covers the REST endpoints. Reconsider AsyncAPI when there is a second non-frontend consumer.

## Consequences
- Codegen uses `json-schema-to-typescript` (TS) and `omissis/go-jsonschema` (Go).
- Migrating to AsyncAPI later is a translation, not a re-architecture.
