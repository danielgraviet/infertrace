# Phase 1: Go Foundations for the MVP

> Outcome: You can build and test a small Go service confidently before adding gRPC and analysis logic.

## MVP Scope Anchor

- Painful problem: latency regressions after deploy are noticed too late.
- User: ML/platform engineer operating one inference service.
- Workflow: ingest spans -> compute baseline/current latency -> emit alert.

## Learning Objectives

- Use Go modules, packages, and exported types intentionally.
- Write and run unit tests with table-driven style.
- Handle errors via `(value, error)` consistently.
- Build simple domain types for tracing data.

## Tasks

1. Install and verify Go (`go version`) and ensure module builds locally.
2. Keep `internal/collector/span.go` as your core domain model and remove unused fields only if they distract from MVP.
3. Implement strict constructors/validators for required fields (`service_name`, `model_name`, `duration_nanos`, `start_time`).
4. Add unit tests for constructor validation and timestamp/duration edge cases.
5. Add a tiny `cmd/collector/main.go` bootstrap that starts and stops cleanly (no gRPC yet required in this phase).

## Exit Criteria

- `go test ./...` passes.
- Span model enforces required fields relevant to latency detection.
- You can explain where each core type lives and why.
