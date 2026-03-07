# Phase 4: In-Memory Time-Window Store + Query API

> Outcome: You can ask one useful question: current vs baseline latency for a model over a time window.

## Learning Objectives

- Model time-windowed storage for short-lived operational insight.
- Expose minimal HTTP query API in Go.
- Test concurrent reads/writes safely.

## Tasks

1. Build in-memory ring buffer storage (last 30-60 minutes).
2. Store only data required for latency analysis.
3. Add HTTP endpoint:
   - `GET /models/{name}/latency?window=5m`
4. Endpoint returns at least:
   - sample_count, p50, p95, p99, window_start, window_end
5. Add tests for empty data, single-point data, normal window queries.

## Exit Criteria

- Endpoint produces correct quantiles for known test fixtures.
- Concurrent ingestion + query tests pass without race conditions.
