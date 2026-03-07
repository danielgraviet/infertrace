# Phase 7: Test Coverage + Reliability Hardening

> Outcome: MVP is trustworthy enough to iterate on.

## Learning Objectives

- Write focused tests that protect behavior, not implementation details.
- Add operational guardrails (timeouts, validation, shutdown).

## Tasks

1. Add integration tests covering gRPC ingest -> store -> query -> detector.
2. Add input validation limits (batch size, required fields, max duration bounds).
3. Add context timeouts for outbound/inbound operations.
4. Add graceful shutdown handling for SIGINT/SIGTERM.
5. Run `go test -race ./...` and fix discovered races.

## Exit Criteria

- Critical path has automated coverage.
- Service exits cleanly without data corruption/panics.
