# Changelog

## [Unreleased]

### Maintenance

- Upgraded the HTTP server from Fiber v2 to v3 — no API behavior changes.

## [0.0.3] - 2026-08-30

### Changed

- Address normalization now uses the shared `address-parser-go` library instead of the in-repo parser — the embedded streets database moves with it, and `PARSER__ADDRESS_DB_PATH` still accepts an external database.

## [0.0.2] - 2026-07-21

### Bug Fixes

- Fixed unknown street names being dropped from parsed outage details — the original name is now kept.

## [0.0.1] - 2026-07-14

### New Features

#### Outage pipeline

- **Scheduled pipeline** — runs scrape → parse → diff → publish on a configurable interval.
- **Scraper** — fetches the outage schedule with ETag change detection and Windows-1251 decoding.
- **Parser** — extracts organization, address, outage and Russian date data.
- **Storage** — Redis-backed deduplication of parsed records with configurable retention.
- **Publisher** — publishes new outage records as JSON to a Redis Pub/Sub channel.

#### API and observability

- **HTTP API** — health probes, Swagger UI, Prometheus metrics, and endpoints to trigger and inspect each pipeline stage.
- **Metrics** — Prometheus counters and latency histograms for pipeline operations.
- **Configuration validation** — invalid Redis/scraper URLs, non-positive intervals or TTLs, and empty prefixes fail at startup.
