# Phase 3: Collector Pipeline + Backpressure

> Outcome: Ingestion is processed through a bounded worker pipeline that behaves predictably under load.

## Learning Objectives

- Use goroutines/channels for bounded concurrency.
- Apply backpressure policies explicitly (drop vs block vs fail).
- Shut down gracefully with `context` and `WaitGroup`.

## Tasks

1. Implement `Collector` with:
   - bounded queue
   - worker pool
   - explicit backpressure behavior when queue is full
2. Wire gRPC handler to enqueue validated spans.
3. Add counters for accepted, dropped, invalid spans.
4. Add tests for queue-full behavior and graceful stop.
5. Document chosen backpressure policy in code comments + phase notes.

## Exit Criteria

- Behavior is deterministic under burst load.
- You can explain why the chosen backpressure policy matches MVP goals.
