# Phase 2: Protocol Buffers + gRPC in Go

> This phase replaces JSON-over-HTTP with a faster, contract-first protocol. You'll define the shape of a `Span` in a `.proto` file, generate Go code from it, and build a gRPC server that actually receives spans. You already have a `.proto` file in `grpc/` — this phase formalizes and extends that work.

---

## Learning Objectives

By the end of this phase you will understand:

- Why Protocol Buffers are faster and smaller than JSON (binary encoding, no field names on the wire)
- The difference between gRPC and REST (streaming, generated clients, strict contracts)
- How `.proto` files define messages and services
- How `protoc` generates Go code from a `.proto` file
- The difference between unary and streaming RPCs
- How to implement a gRPC server interface in Go

---

## Tasks

1. **Install the toolchain:**
   ```bash
   # Install protoc (the protobuf compiler)
   brew install protobuf

   # Install Go plugins for protoc
   go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
   go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
   ```
   Add `$(go env GOPATH)/bin` to your `$PATH` if it isn't already.

2. **Create `proto/span.proto`** with the core span schema:
   ```protobuf
   syntax = "proto3";
   package infertrace;
   option go_package = "github.com/yourusername/infertrace/proto";

   message Span {
     bytes trace_id = 1;
     bytes span_id = 2;
     bytes parent_span_id = 3;
     string service_name = 4;
     string operation_name = 5;
     int64 start_time_unix_nano = 6;
     int64 duration_nanos = 7;
     string status = 8;
     map<string, string> tags = 9;
   }
   ```

3. **Add `InferenceContext` and `ResourceMetrics` messages** to the same file:
   ```protobuf
   message InferenceContext {
     string model_name = 1;
     string model_version = 2;
     int32 batch_size = 3;
     int32 input_tokens = 4;
     int32 output_tokens = 5;
   }

   message ResourceMetrics {
     double gpu_utilization_percent = 1;
     int64 gpu_memory_used_bytes = 2;
     int64 gpu_memory_total_bytes = 3;
   }
   ```
   Add `InferenceContext inference = 10` and `ResourceMetrics resources = 11` to the `Span` message.

4. **Define the `CollectorService`** in the same `.proto` file:
   ```protobuf
   message SendSpanRequest {
     Span span = 1;
   }

   message SendSpanResponse {
     bool accepted = 1;
   }

   message SendSpanBatchRequest {
     repeated Span spans = 1;
   }

   service CollectorService {
     rpc SendSpan(SendSpanRequest) returns (SendSpanResponse);
     rpc SendSpanBatch(SendSpanBatchRequest) returns (SendSpanResponse);
   }
   ```
   Think about why `SendSpanBatch` is useful — sending 100 spans in one call is much more efficient than 100 individual calls.

5. **Generate Go code:**
   ```bash
   protoc --go_out=. --go-grpc_out=. proto/span.proto
   ```
   Open the generated files. You'll see Go structs for every message and an interface you must implement to create the server.

6. **Implement the gRPC server** in `internal/collector/server.go`:
   - Create a `Server` struct that implements the generated `CollectorServiceServer` interface
   - In `SendSpan`: log the received span's service name and model name, return `&proto.SendSpanResponse{Accepted: true}`
   - In `SendSpanBatch`: loop over spans, log each one

7. **Wire it up in `cmd/collector/main.go`:**
   ```go
   lis, _ := net.Listen("tcp", ":4317")
   grpcServer := grpc.NewServer()
   proto.RegisterCollectorServiceServer(grpcServer, &collector.Server{})
   grpcServer.Serve(lis)
   ```

8. **Write a test client** in `cmd/testclient/main.go` that:
   - Connects to `localhost:4317`
   - Sends a single `Span` with `service_name = "test-service"` and `model_name = "gpt-4"`
   - Prints whether it was accepted

   Run the server in one terminal, the client in another. You should see the span logged.
