# InferTrace

InferTrace is a learning-first observability project for LLM inference services.

## MVP Pivot

The project was narrowed from a broad tracing platform into one concrete workflow:

1. Ingest span batches from one inference service.
2. Compare current latency against a baseline.
3. Alert when latency regresses after a model deploy.

This keeps scope small enough to build quickly while still solving a real operational problem.

## Current State

Implemented so far:

- Go span model with ingest validation (`service_name`, `model_name`, `start_time_unix_nano`, `duration_nanos`).
- Protobuf + gRPC ingestion contract (`SendSpanBatch`).
- Collector gRPC server that returns accepted/rejected counts.
- Bounded in-memory pipeline with worker pool and explicit backpressure policy.

Backpressure policy for MVP:

- `drop-on-full` (non-blocking enqueue).
- If the queue is full, spans are rejected quickly instead of blocking callers.

## Architecture (Current)

- `cmd/collector/main.go`: app bootstrap, gRPC server startup, graceful shutdown.
- `internal/collector/span.go`: internal span model + validation.
- `internal/collector/server.go`: gRPC handler, protobuf-to-domain conversion, enqueue logic.
- `internal/collector/collector.go`: bounded queue, workers, counters, graceful stop.
- `proto/span.proto`: ingestion API contract.
- `proto/*.pb.go`: generated Go protobuf/gRPC code.

## Run Locally

Start collector server:

```bash
go run ./cmd/collector
```

It listens on `:50051`.

## Run Tests

```bash
go test ./...
```

## Project Phases

- Phase 1: Go fundamentals and span modeling.
- Phase 2: Protobuf contract and gRPC ingestion.
- Phase 3: Collector pipeline and backpressure.
- Next: in-memory/query/storage phases from `tasks/`.

## Notes

- Learning notes live in `learnings/`.
- Phase tasks are in `tasks/`.
