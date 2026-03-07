# Phase 3 Learnings: Collector Pipeline + Backpressure

## Backpressure Policy Chosen

Policy: `drop-on-full` (non-blocking enqueue).

When the in-memory queue is full, new spans are rejected immediately instead of blocking RPC handlers.

## Why This Fits MVP

- Keeps ingest latency predictable under burst traffic.
- Avoids turning collector overload into cascading request timeouts.
- Makes overload visible via counters (`dropped`) and batch response rejected counts.

For this MVP, fast failure is more useful than buffering everything, because the core goal is detecting regressions quickly with clear system behavior.

## Collector Design

- Bounded channel queue.
- Fixed worker pool reading from the queue.
- Atomic counters for `accepted`, `dropped`, and `invalid`.
- `Stop(ctx)` closes queue once and waits for workers to drain queued spans.

## Behavioral Guarantees

- Queue full behavior is deterministic: enqueue returns `ErrQueueFull`.
- After stop, enqueue returns `ErrCollectorClosed`.
- Graceful stop drains already accepted spans before workers exit (unless context deadline is exceeded).
