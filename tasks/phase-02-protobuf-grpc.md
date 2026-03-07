# Phase 2: Protobuf Contract + gRPC Ingestion

> Outcome: A single, clear ingestion contract exists and a Go gRPC server accepts batched spans.

## Learning Objectives

- Design a protobuf schema for one workflow, not a platform.
- Generate Go stubs and implement server interfaces.
- Understand batch RPC tradeoffs (`SendSpanBatch` only for MVP).

## Tasks

1. Create `proto/span.proto` for MVP fields only:
   - trace_id, span_id, service_name, operation_name, model_name
   - start_time_unix_nano, duration_nanos, status
2. Define `CollectorService` with one RPC:
   - `rpc SendSpanBatch(SendSpanBatchRequest) returns (SendSpanBatchResponse)`
3. Generate Go code via `protoc` and check in generated files.
4. Implement gRPC server in `internal/collector/server.go`.
5. Return accepted count + rejected count in response for visibility.
6. Add server tests for valid batch, empty batch, malformed span.

## Exit Criteria

- Local client can send one batch and receive accepted/rejected counts.
- Invalid spans do not crash server and are reported as rejected.
