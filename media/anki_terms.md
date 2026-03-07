# Anki Terms — InferTrace / Go Concurrency

## Spans

A {{c1::span}} is the smallest unit of data in InferTrace — it records one inference event including model name, start time, duration, and service name.

A span is used to detect {{c1::latency regressions}} by comparing current duration data against a baseline.

---

## gRPC Handler

The {{c1::gRPC handler}} receives requests in binary protobuf format, validates each span, converts it to an internal domain type, and enqueues it for processing.

The gRPC handler is the {{c1::entry gate}} — it rejects spans with missing required fields before they ever reach the queue.

---

## Goroutines

A {{c1::goroutine}} is Go's lightweight thread. You can run thousands of them cheaply because they are managed by the Go runtime, not the OS.

Goroutines allow the gRPC handler to {{c1::return immediately}} while span processing continues in the background.

---

## Channels

A {{c1::channel}} is Go's mechanism for passing data safely between goroutines. Only one goroutine touches the data at a time — the sender gives up ownership on send.

The Go concurrency philosophy is: don't communicate by {{c1::sharing memory}}, share memory by {{c1::communicating (through channels)}}.

---

## Buffered vs Unbuffered Channels

An {{c1::unbuffered}} channel has zero capacity — the sender blocks until a receiver is ready. It is a direct handoff.

A {{c1::buffered}} channel has fixed capacity. The sender can add items up to that capacity before it {{c1::blocks}}.

`make(chan Span, 100)` creates a buffered channel with capacity {{c1::100}}.

---

## Data Race

A {{c1::data race}} occurs when two goroutines access the same memory concurrently and at least one is writing, with no synchronization between them.

The danger of a data race is that updates can be {{c1::silently lost}} — the result depends on thread timing and is non-deterministic.

Example: two goroutines both read `counter = 5`, both add 1, both write `6`. The counter should be {{c1::7}} but ends up as {{c1::6}}.

---

## Atomic Operations

An {{c1::atomic operation}} is enforced at the hardware level — the CPU executes the read-modify-write as a single uninterruptible instruction, with no gap for another core to intervene.

Atomics are {{c1::lock-free}} because no goroutine is suspended — the hardware serializes the memory access, not the OS scheduler.

`atomic.Int64.Add(1)` compiles to a single {{c1::atomic CPU instruction}} (e.g. `LOCK XADD` on x86).

Atomics only work for {{c1::simple operations on a single value}} (increment, swap, compare-and-swap). Multi-step operations still require a {{c1::mutex or channel}}.

---

## Mutex (Lock)

A {{c1::mutex}} enforces exclusivity at the software level — goroutines that cannot acquire the lock are {{c1::suspended by the OS}} and put to sleep until it is released.

The cost of a mutex over an atomic is {{c1::context switching}} and OS scheduler involvement, which adds overhead under contention.

---

## Worker Pool

A {{c1::worker pool}} is a fixed number of goroutines all reading from the same channel. It bounds the maximum parallelism in the system.

Worker pools provide {{c1::bounded concurrency}} — unlike spawning a new goroutine per request, memory and CPU usage stay predictable under load.

In InferTrace, each worker loops with `for span := range c.queue`, which {{c1::blocks}} until a span is available, then processes it.

---

## Backpressure

{{c1::Backpressure}} is the policy that determines what happens when a producer sends work faster than consumers can handle it.

The three backpressure options are: {{c1::block}} (caller waits), {{c1::drop}} (reject new work immediately), and {{c1::fail}} (return error to caller).

InferTrace uses {{c1::drop-on-full}} — when the queue is full, new spans are rejected immediately via a `select` with a `default` branch, keeping ingest latency bounded.

---

## select with default (Non-Blocking Send)

A `select` with a `default` branch performs a {{c1::non-blocking}} channel send — if the channel is full, it falls through to `default` instantly instead of blocking.

```go
select {
case queue <- span:   // succeeds if channel has room
    accepted++
default:              // runs immediately if channel is full
    dropped++
}
```

Without `default`, a full channel send would {{c1::block the caller}} until a worker frees a slot.

---

## Graceful Shutdown

In InferTrace, `Stop()` shuts down in three steps: (1) set `closed = true` to reject new enqueues, (2) {{c1::close the channel}} to signal workers to exit after draining, (3) call {{c1::workerWg.Wait()}} to block until all workers finish.

`sync.Once` is used in `Stop()` to ensure the channel is {{c1::only closed once}} — closing an already-closed channel in Go causes a {{c1::panic}}.

When a channel is closed, a `for range` loop over it {{c1::exits cleanly}} after processing all remaining items.

---

## Go Execution Model

In Go, you usually run a {{c1::package}} (for example `./cmd/collector`) rather than a single file.

`go run ./cmd/collector` does two things: {{c1::compiles}} the `main` package and then {{c1::executes}} the resulting temporary binary.

A runnable Go program must be in `package {{c1::main}}` and define `func {{c1::main()}}`.

Compared to Python's `python3 main.py`, Go is {{c1::package-first}} while Python is commonly {{c1::file/script-first}}.

---

## Build vs Run

`go build -o bin/collector ./cmd/collector` creates a persistent {{c1::executable binary}} you can run later with `./bin/collector`.

`go run` creates a {{c1::temporary binary}}, while `go build` creates a {{c1::saved artifact}} for deployment.

A Go binary is machine code for a specific {{c1::OS/CPU target}}.

---

## Why Compile

One major advantage of compiled Go is shifting many failures to {{c1::compile time}} instead of runtime.

Compiled executables often have {{c1::faster startup and runtime}} than interpreted scripts because they run native machine instructions.

Tradeoff: binaries are usually {{c1::larger on disk}} than source scripts.

Larger binary size mostly affects {{c1::storage/download/cold-start I/O}}, while CPU execution is often {{c1::faster}}.

---

## Go Test Commands

Run all tests in the module with `go test {{c1::./...}}`.

Run tests for a specific folder/package with `go test {{c1::./internal/store}}` (or another package path).

Run a single named test with `go test ./internal/store -run {{c1::TestLatencyStore_ConcurrentAddAndQuery}}`.

Use `-count={{c1::1}}` to bypass test caching and force a fresh run.

Use `go test -race ./...` to detect {{c1::data races}} in concurrent code.
