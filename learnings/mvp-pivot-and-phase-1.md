# MVP Pivot + Phase 1 Learnings

## Why We Pivoted

The original project vision was broad: full tracing platform, storage tiers, analysis engine, UI, and ML-specific extras.

That scope is valuable long-term, but too large for fast learning and real usefulness right now.

We narrowed to one painful problem:

- Problem: latency regressions after model deploy are noticed too late.
- User: ML/platform engineer running one inference service.
- Workflow: ingest spans -> compare current latency vs baseline -> alert.

This pivot improves two things at once:

- Learning quality: fewer moving parts means deeper understanding of Go + gRPC + system behavior.
- Practical utility: we can ship one real, testable outcome quickly.

## What We Changed in Phase 1

### 1. Added `ModelName` to `Span`

We added `ModelName` to the core span struct because regression detection is model-specific. Without it, we cannot answer "which model regressed?"

### 2. Added `ValidateForIngest()`

We added a guard method that validates required ingest fields:

- `service_name`
- `model_name`
- `start_time_unix_nano > 0`
- `duration_nanos > 0`

Why this matters:

- invalid data is rejected early
- downstream logic can assume required fields exist
- debugging gets easier because failures happen at boundaries

### 3. Upgraded constructor to enforce required fields

`NewSpan` now takes required MVP identity fields and returns `(*Span, error)`.

Why return an error from constructor:

- prevents silently creating invalid spans
- keeps invalid state from spreading through the system
- follows idiomatic Go error handling (`value, error`)

## Test-First Workflow We Used

We wrote/used tests to define behavior first, then implemented minimum code to satisfy tests.

Current tests cover:

- constructor sets key fields (`SpanID`, service/op/model, timestamp)
- constructor rejects missing required fields
- validation rejects missing service/model/start/duration

Why this was important:

- gave exact target behavior
- reduced guesswork while coding
- made refactoring safer

## Key Go Lessons Reinforced

- Guard clauses with `if` are the simplest validation style in Go.
- Go booleans are `true`/`false` (lowercase).
- Go has `if`, `else if`, `switch`; no Python `elif` or `match`.
- Small, explicit constructors + validation are better than permissive object creation.

## How This Sets Up Phase 2

Because spans are now validated and include `model_name`, Phase 2 can focus on gRPC ingestion contract design instead of cleaning bad input later.

That keeps the next step narrow:

- define one proto for batch span ingestion
- accept batches in collector
- report accepted vs rejected
