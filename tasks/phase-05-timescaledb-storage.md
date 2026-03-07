# Phase 5: Regression Detector (Baseline vs Current)

> Outcome: System emits an alert when model latency regresses beyond threshold.

## Learning Objectives

- Build practical anomaly logic without overfitting complexity.
- Separate detection policy from data retrieval.
- Define actionable alert payloads.

## Tasks

1. Implement detector module (Go or Python; pick one and stay consistent) that compares:
   - baseline window (example: previous 30m)
   - current window (example: last 5m)
2. Default trigger rule:
   - alert if `current_p95 >= baseline_p95 * 1.8` and minimum sample count met
3. Emit structured alert with:
   - model_name, current_p95, baseline_p95, multiplier, first_seen_at
4. Add unit tests for:
   - no baseline
   - insufficient samples
   - threshold breach
5. Add a simple sink for alerts (stdout log is enough for MVP).

## Exit Criteria

- Synthetic regression causes alert within 1-2 polling intervals.
- Alert message is specific enough to drive action.
