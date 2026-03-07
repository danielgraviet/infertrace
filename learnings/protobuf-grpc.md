# Protocol Buffers + gRPC

## What We Built

A gRPC server that receives `Span` data over the network — replacing the in-memory struct work from phase 1 with a real network protocol.

---

## Why gRPC Instead of REST?

| | REST + JSON | gRPC + Protobuf |
|---|---|---|
| Format | Text (JSON) | Binary (protobuf) |
| Schema | Optional | Enforced by `.proto` |
| Speed | Slower (larger payloads) | Faster (smaller, no field names on wire) |
| Code gen | Manual | Auto-generated from `.proto` |
| Streaming | No | Yes |

For a tracing collector receiving thousands of spans per second, binary + strict contracts matter.

---

## How It Works

### 1. Define the contract in `.proto`
`proto/span.proto` defines the shape of every message and which RPCs exist:
```protobuf
service CollectorService {
    rpc SendSpan(SendSpanRequest) returns (SendSpanResponse);
    rpc SendSpanBatch(SendSpanBatchRequest) returns (SendSpanResponse);
}
```
This is the single source of truth. Both the server and client are generated from it.

### 2. Generate Go code
```bash
protoc --go_out=. --go-grpc_out=. proto/span.proto
```
This produces two files:
- `proto/span.pb.go` — Go structs for every message
- `proto/span_grpc.pb.go` — the server interface and client you must implement

### 3. Implement the server (`internal/collector/server.go`)
Your `Server` struct implements the generated `CollectorServiceServer` interface:
```go
type Server struct {
    proto.UnimplementedCollectorServiceServer // embed for forward compatibility
}

func (s *Server) SendSpan(ctx context.Context, req *proto.SendSpanRequest) (*proto.SendSpanResponse, error) {
    fmt.Printf("received span: service=%s model=%s\n", req.Span.ServiceName, req.Span.Inference.ModelName)
    return &proto.SendSpanResponse{Accepted: true}, nil
}
```
`UnimplementedCollectorServiceServer` pre-implements all interface methods so your code doesn't break if new RPCs are added later.

### 4. Wire it up (`cmd/collector/main.go`)
```go
lis, _ := net.Listen("tcp", ":4317")       // open TCP port
grpcServer := grpc.NewServer()              // create the gRPC engine
proto.RegisterCollectorServiceServer(       // tell gRPC: route CollectorService calls here
    grpcServer, &collector.Server{})
grpcServer.Serve(lis)                       // block and listen forever
```

### 5. Test client (`cmd/testclient/main.go`)
Connects and sends a real span over the wire:
```go
conn, _ := grpc.NewClient("localhost:4317", grpc.WithTransportCredentials(insecure.NewCredentials()))
client := proto.NewCollectorServiceClient(conn)
resp, _ := client.SendSpan(context.Background(), &proto.SendSpanRequest{
    Span: &proto.Span{
        ServiceName: "test-service",
        Inference:   &proto.InferenceContext{ModelName: "gpt-4"},
    },
})
```

---

## How to Run the Test

**Terminal 1 — start the server:**
```bash
go run cmd/collector/main.go
```
Wait for: `collector listening on :4317`

**Terminal 2 — send a test span:**
```bash
go run cmd/testclient/main.go
```

**Expected output:**

Terminal 1 (server):
```
received span: service=test-service model=gpt-4
```

Terminal 2 (client):
```
accepted: true
```

---

## Key Concepts

**Embedding** — putting one struct inside another to inherit its methods. `proto.UnimplementedCollectorServiceServer` gives you default "not implemented" responses for all RPCs so the interface is satisfied even if you only implement some methods.

**`&collector.Server{}`** — the `&` creates a pointer to a new empty `Server` struct. gRPC holds onto this pointer and calls its methods when requests arrive.

**Port 4317** — the standard OpenTelemetry collector port. Using it means any OTel-compatible SDK can send spans to your collector without custom configuration.

**`insecure.NewCredentials()`** — skips TLS for local development. In production you'd use real certificates.