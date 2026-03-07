# Phase 9: Performance Baseline (Not Premature Optimization)

> Outcome: You have a measured baseline and known bottlenecks for the MVP.

## Learning Objectives

- Benchmark with intent and interpret key metrics.
- Profile only after behavior is correct.

## Tasks

1. Add benchmark for `SendSpanBatch` ingest throughput.
2. Capture baseline metrics:
   - spans/sec
   - p95 ingest latency
   - memory allocations/op
3. Use pprof to identify top CPU and allocation hotspots.
4. Apply one targeted optimization and measure delta.
5. Document current limits and next constraints.

## Exit Criteria

- You can state realistic throughput for this MVP.
- Optimization work is justified by measurements.
