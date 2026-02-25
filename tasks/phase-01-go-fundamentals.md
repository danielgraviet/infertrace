# Phase 1: Go Fundamentals + Project Setup

> Before touching the collector, you need to feel comfortable in Go. This phase builds the language foundation you'll use for every subsequent phase. Coming from Python, Go will feel strict at first — that strictness is intentional and will serve you well.

---

## Learning Objectives

By the end of this phase you will understand:

- How Go's type system differs from Python's (static vs dynamic, zero values vs None)
- How Go handles errors — no exceptions, just `(value, error)` return pairs
- How Go modules work (`go.mod`, `go.sum`) vs Python's `pyproject.toml`
- What "exported" vs "unexported" means (capitalization = public/private)
- How Go structs and methods work vs Python classes
- How to write and run a basic Go test

---

## Tasks

1. **Install Go** — download from [go.dev/dl](https://go.dev/dl). Verify with `go version`.

2. **Initialize the Go module** at the project root:
   ```bash
   go mod init github.com/yourusername/infertrace
   ```
   Open `go.mod` and understand what it contains. Compare it mentally to your `pyproject.toml`.

3. **Create the project directory structure:**
   ```
   infertrace/
   ├── cmd/
   │   └── collector/
   │       └── main.go       ← collector entry point
   ├── internal/
   │   ├── collector/        ← collector logic (future phases)
   │   └── storage/          ← storage logic (future phases)
   ├── proto/                ← .proto files (Phase 2)
   ├── sdk/
   │   └── go/               ← Go SDK (Phase 7)
   └── go.mod
   ```

4. **Write a `Span` struct** in `internal/collector/span.go` with these fields:
   - `TraceID` (string)
   - `SpanID` (string)
   - `ParentSpanID` (string)
   - `ServiceName` (string)
   - `OperationName` (string)
   - `StartTimeUnixNano` (int64)
   - `DurationNanos` (int64)
   - `Status` (string)

   Notice: field names are capitalized because they need to be exported (accessible from other packages).

5. **Write a constructor function** `NewSpan(serviceName, operationName string) *Span` that sets `SpanID` to a random UUID and `StartTimeUnixNano` to the current time. Install the `google/uuid` package:
   ```bash
   go get github.com/google/uuid
   ```

6. **Practice Go error handling** — write a function `ParseTraceID(raw string) (string, error)` that returns an error if the string is empty. Call it from `main.go` and handle the error explicitly. Notice there are no try/catch blocks.

7. **Write your first Go test** in `internal/collector/span_test.go`:
   - Test that `NewSpan` sets a non-empty SpanID
   - Test that `ParseTraceID` returns an error for an empty string
   - Run with `go test ./...`

8. **Read:** Work through [A Tour of Go](https://go.dev/tour) — focus on: Basics, Types, Functions, Methods, and Interfaces. Skip concurrency for now (that's Phase 3). This should take 1–2 hours and will pay off immediately.
