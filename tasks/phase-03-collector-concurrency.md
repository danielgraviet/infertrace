# Phase 3: Core Collector — Concurrency + Worker Pool

> This is the most important Go phase. The collector needs to handle 100k+ spans/second, which means you can't process each span synchronously — you need concurrency. Go's approach (goroutines + channels) is fundamentally different from Python's threading model. By the end of this phase you'll understand *why* Go is the right choice for this component.

---

## Learning Objectives

By the end of this phase you will understand:

- What goroutines are and how they differ from Python threads (lightweight, scheduled by the Go runtime)
- How channels work as typed, concurrency-safe message queues
- The worker pool pattern — a fixed number of goroutines draining a shared queue
- What `sync.Pool` is and why it reduces garbage collector (GC) pressure
- What backpressure means: when the queue is full, you must decide to block, drop, or error
- How to use `sync.WaitGroup` to wait for goroutines to finish

---

## Tasks

1. **Read "Goroutines" and "Channels"** from [A Tour of Go](https://go.dev/tour/concurrency/1) before writing any code. The mental model is: goroutines are cheap threads, channels are the pipes between them.

2. **Define the `Collector` struct** in `internal/collector/collector.go`:
   ```go
   type Collector struct {
       ingestQueue chan *proto.Span
       spanPool    *sync.Pool
       workers     int
       wg          sync.WaitGroup
   }

   func NewCollector(workers, queueSize int) *Collector {
       return &Collector{
           ingestQueue: make(chan *proto.Span, queueSize),
           workers:     workers,
           spanPool: &sync.Pool{
               New: func() any { return &proto.Span{} },
           },
       }
   }
   ```

3. **Implement the `Start` method** that launches N worker goroutines:
   ```go
   func (c *Collector) Start() {
       for i := 0; i < c.workers; i++ {
           c.wg.Add(1)
           go c.ingestWorker()
       }
   }
   ```

4. **Implement `ingestWorker`** — each worker reads from the channel in a loop:
   ```go
   func (c *Collector) ingestWorker() {
       defer c.wg.Done()
       for span := range c.ingestQueue {
           c.processSpan(span)
           // Return span to pool when done (Phase 4 will write it to storage instead)
           c.spanPool.Put(span)
       }
   }
   ```
   For now, `processSpan` just logs the span. You'll swap this out in Phase 4.

5. **Implement `Ingest` with backpressure** — this is the method the gRPC server calls:
   ```go
   func (c *Collector) Ingest(span *proto.Span) error {
       select {
       case c.ingestQueue <- span:
           return nil
       default:
           // Queue is full — drop the span and return an error
           return fmt.Errorf("ingest queue full, span dropped")
       }
   }
   ```
   The `select` + `default` pattern is how you do non-blocking channel sends in Go. Understand what happens if you remove the `default` case (it blocks — the gRPC handler would stall).

6. **Implement `Stop`** for graceful shutdown:
   ```go
   func (c *Collector) Stop() {
       close(c.ingestQueue) // signals workers to stop when queue is drained
       c.wg.Wait()          // waits for all workers to finish
   }
   ```

7. **Wire the Collector into your gRPC server** from Phase 2:
   - Store a `*Collector` on the `Server` struct
   - In `SendSpan`, call `c.collector.Ingest(req.Span)` instead of just logging
   - In `SendSpanBatch`, call `Ingest` for each span in the batch

8. **Write a benchmark** in `internal/collector/collector_bench_test.go`:
   ```go
   func BenchmarkCollectorIngest(b *testing.B) {
       c := NewCollector(8, 10000)
       c.Start()
       defer c.Stop()

       b.RunParallel(func(pb *testing.PB) {
           for pb.Next() {
               span := &proto.Span{ServiceName: "bench-service"}
               c.Ingest(span)
           }
       })
   }
   ```
   Run it: `go test -bench=BenchmarkCollectorIngest -benchmem ./internal/collector/`

   Note the `ns/op` and `allocs/op` numbers. You'll improve them in Phase 9.

9. **Experiment:** What happens to throughput if you change `workers` from 8 to 1? To 32? Run the benchmark each time and observe. Form a hypothesis about why.
