# ADR-0008: OCSF toy dataset (4 narrow tables, Hive+Parquet)

**Status:** Accepted (Phase 1 implementation; recorded now for traceability) | **Date:** 2026-05-01

## Context
For the Trino tool-use POC we need a dataset that's realistic enough for real-shaped agent reasoning, small enough to commit-or-regenerate, and aligned with the user's cybersecurity domain.

## Decision
Four OCSF event classes: `authentication` (3002), `process_activity` (1007), `network_activity` (4001), `detection_finding` (2004). ~30 days, ~100k rows total, real OCSF column names (nested struct types preserved). Parquet on local FS via Trino's Hive connector with a local metastore. One synthetic incident woven through (credential stuffing → suspicious PowerShell → outbound exfil).

## Consequences
- Phase 1 brings up Trino + Hive metastore in docker-compose.
- Dataset generator is regenerable, not committed binary.
