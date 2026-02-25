# Phase 9: Benchmarking + Performance Profiling

> You built a working system. Now you make it fast. This phase teaches you Go's built-in performance tooling — benchmarks and the `pprof` profiler. You'll measure the collector's throughput, find the bottlenecks, apply fixes, and prove the improvement with data. This is the difference between "it seems fast" and "it handles 500k spans/sec."

---

## Learning Objectives

By the end of this phase you will understand:

- How Go's benchmark framework works (`testing.B`, `b.RunParallel`, `b.ResetTimer`)
- How to interpret benchmark output (`ns/op`, `B/op`, `allocs/op`)
- How to profile CPU usage with `pprof` and read a flame graph
- How to profile memory allocations and identify what's causing GC pressure
- What `sync.Pool` buys you and how to measure the difference
- How to use `go test -race` to catch concurrency bugs

---

## Tasks

1. **Write the canonical collector benchmark** in `internal/collector/collector_bench_test.go` — if you haven't already from Phase 3:
   ```go
   func BenchmarkCollectorIngest_Sequential(b *testing.B) {
       c := NewCollector(8, 100000)
       c.Start()
       defer c.Stop()

       b.ResetTimer() // Don't count setup time
       for i := 0; i < b.N; i++ {
           span := &proto.Span{
               ServiceName:   "bench-service",
               OperationName: "infer",
           }
           c.Ingest(span)
       }
   }

   func BenchmarkCollectorIngest_Parallel(b *testing.B) {
       c := NewCollector(8, 100000)
       c.Start()
       defer c.Stop()

       b.ResetTimer()
       b.RunParallel(func(pb *testing.PB) {
           for pb.Next() {
               c.Ingest(&proto.Span{ServiceName: "bench-service"})
           }
       })
   }
   ```

2. **Run your baseline benchmark** and save the results:
   ```bash
   go test -bench=. -benchmem -count=5 ./internal/collector/ | tee bench_baseline.txt
   ```
   Record the `ns/op` and `allocs/op` numbers. This is your starting point.

3. **Profile CPU usage:**
   ```bash
   go test -bench=BenchmarkCollectorIngest_Parallel \
       -cpuprofile=cpu.prof \
       -benchtime=10s \
       ./internal/collector/

   go tool pprof -http=:8081 cpu.prof
   ```
   Open `http://localhost:8081` — you'll see a flame graph. The widest bars consume the most CPU. Identify the top 3 hotspots.

4. **Profile memory allocations:**
   ```bash
   go test -bench=BenchmarkCollectorIngest_Parallel \
       -memprofile=mem.prof \
       ./internal/collector/

   go tool pprof -http=:8082 mem.prof
   ```
   Switch to the "alloc_space" view. What's allocating the most memory? Protobuf deserialization? The `Span` struct itself?

5. **Apply Optimization 1 — `sync.Pool` for span reuse.** If you haven't wired the pool into `Ingest`:
   ```go
   func (c *Collector) Ingest(span *proto.Span) error {
       // Get a span from pool, copy fields, avoid new allocation
       pooled := c.spanPool.Get().(*proto.Span)
       proto.Merge(pooled, span) // copy fields in
       select {
       case c.ingestQueue <- pooled:
           return nil
       default:
           c.spanPool.Put(pooled) // return immediately if queue full
           return fmt.Errorf("queue full")
       }
   }
   ```
   Re-run the benchmark. How many `allocs/op` did you eliminate?

6. **Apply Optimization 2 — batch writes to storage.** Instead of `storage.Write(span)` per span in the worker, accumulate a batch and flush:
   ```go
   func (c *Collector) ingestWorker() {
       defer c.wg.Done()
       batch := make([]*proto.Span, 0, 256)
       ticker := time.NewTicker(50 * time.Millisecond)

       for {
           select {
           case span, ok := <-c.ingestQueue:
               if !ok {
                   c.storage.WriteBatch(batch) // flush remaining
                   return
               }
               batch = append(batch, span)
               if len(batch) >= 256 {
                   c.storage.WriteBatch(batch)
                   batch = batch[:0] // reset without reallocating
               }
           case <-ticker.C:
               if len(batch) > 0 {
                   c.storage.WriteBatch(batch)
                   batch = batch[:0]
               }
           }
       }
   }
   ```

7. **Run the benchmark again** and compare:
   ```bash
   go test -bench=. -benchmem -count=5 ./internal/collector/ | tee bench_optimized.txt
   ```
   Use `benchstat` to get a statistically rigorous comparison:
   ```bash
   go install golang.org/x/perf/cmd/benchstat@latest
   benchstat bench_baseline.txt bench_optimized.txt
   ```

8. **Run a sustained load test** against the real gRPC server — write a `cmd/loadtest/main.go`:
   ```go
   func main() {
       conn, _ := grpc.Dial("localhost:4317", grpc.WithInsecure())
       client := proto.NewCollectorServiceClient(conn)

       var sent int64
       start := time.Now()

       // Launch 50 goroutines sending spans as fast as possible
       var wg sync.WaitGroup
       for i := 0; i < 50; i++ {
           wg.Add(1)
           go func() {
               defer wg.Done()
               for time.Since(start) < 30*time.Second {
                   client.SendSpan(context.Background(), &proto.SendSpanRequest{
                       Span: &proto.Span{ServiceName: "loadtest"},
                   })
                   atomic.AddInt64(&sent, 1)
               }
           }()
       }
       wg.Wait()
       fmt.Printf("Sent %d spans in 30s = %.0f spans/sec\n", sent, float64(sent)/30)
   }
   ```

9. **Document your findings** in a `BENCHMARKS.md` file at the repo root:
   - Baseline numbers
   - Optimized numbers
   - Explanation of what each optimization did and why
   - Your target: 500k spans/sec (or as close as you can get)

   This file becomes part of your portfolio. Real companies want to see this kind of rigorous measurement.
