# Phase 6: End-to-End Demo Workflow

> Outcome: You can demonstrate the full user workflow in under 5 minutes.

## Learning Objectives

- Build a realistic but minimal demo harness.
- Validate system behavior end-to-end, not by component only.

## Tasks

1. Create demo client (`cmd/demo` or `scripts/demo`) that sends batches to gRPC collector.
2. Simulate two traffic modes for one model:
   - normal baseline latency
   - regression spike period
3. Run collector + query API + detector locally.
4. Capture a scripted demo flow:
   - start services
   - send baseline traffic
   - inject regression
   - show alert output
   - query latency endpoint
5. Add a single command or short script to run the demo reliably.

## Exit Criteria

- Fresh clone user can run the demo and observe a real alert.
- Workflow matches target user pain point exactly.
