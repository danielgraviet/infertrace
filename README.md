# InferTrace

**Catch model latency regressions before users do.**

InferTrace is a high-performance observability collector for LLM inference services. It ingests spans from your model serving stack, tracks per-model latency, and detects regressions when you deploy a new model version.

---

## Why This Exists

Generic tracing tools (Jaeger, Tempo) were not built for ML inference. They do not know about model versions, batch sizes, or token counts. When your p99 latency doubles after a model swap, you should not have to dig through generic trace data to find out why.

InferTrace is purpose-built for LLM inference pipelines: ML-aware span schema, per-model latency windows, and regression detection out of the box.

## Who This Is For

- ML engineers running LLM inference in production
- Platform teams that own model serving infrastructure
- Anyone who has been surprised by a latency regression after a model deploy

---

## Architecture

```
Inference Service
      |
      | gRPC (SpanBatch)
      v
  [Collector]
      |
  Bounded Queue + Worker Pool
      |
      v
  In-Memory Store (per-model latency windows)
      |
      v
  HTTP Query API --> Regression Alerts
```

The collector is stateless and fast. The bounded queue uses drop-on-full backpressure so callers are never blocked. Workers drain the queue and write into per-model time windows. The query API answers: "how does current latency compare to baseline?"

---

## Demo

> Demo GIF coming in Phase 10. Will show: deploy new model -> latency regresses -> InferTrace detects and reports it.

---

## Quickstart

**Requirements:** Go 1.21+

```bash
git clone https://github.com/your-username/infertrace
cd infertrace
go run ./cmd/collector
```

The collector starts on `:50051` (gRPC) and is ready to accept span batches.

Run the full test suite:

```bash
go test ./...
```

---

## What Is Built

- Go span model with ingest validation (`service_name`, `model_name`, `start_time_unix_nano`, `duration_nanos`)
- Protobuf + gRPC ingestion contract (`SendSpanBatch`)
- Collector gRPC server with accepted/rejected counts per batch
- Bounded in-memory pipeline with worker pool and `drop-on-full` backpressure

---

## Roadmap

| Phase | Status | What It Delivers |
|-------|--------|-----------------|
| Span model + gRPC ingestion | Done | Core span struct, protobuf contract, gRPC collector |
| Collector pipeline | Done | Bounded queue, worker pool, backpressure |
| In-memory store + query API | Next | Per-model latency windows, p50/p95/p99 HTTP endpoint |
| Latency regression detection | Planned | Compare current vs baseline, surface regressions |
| Python SDK | Planned | Decorator-based tracing for PyTorch / HuggingFace |
| Demo + benchmarks | Planned | End-to-end demo with injected regression scenario |

---

## Project Structure

```
cmd/collector/main.go          -- app bootstrap, gRPC server, graceful shutdown
internal/collector/span.go     -- span model and validation
internal/collector/server.go   -- gRPC handler, protobuf-to-domain conversion
internal/collector/collector.go -- bounded queue, workers, counters
proto/span.proto               -- ingestion API contract
```

---

## Contributing

This project is under active development. Good places to start:

- Issues labeled `good first issue`
- Issues labeled `help wanted`

Feedback, bug reports, and pull requests are welcome. If you run LLM services, feedback on the design is especially valuable.
