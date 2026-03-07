# Phase 8: Packaging + Local Operability

> Outcome: Running the MVP locally is easy and repeatable.

## Learning Objectives

- Package multi-service local environments.
- Write docs that enable reuse by others.

## Tasks

1. Add `docker-compose.yml` for collector + optional detector service.
2. Add Dockerfiles needed for local startup.
3. Write a practical README with:
   - project purpose (single painful problem)
   - quick start
   - demo run steps
   - expected alert output
4. Add troubleshooting notes for common failures (ports, protoc, missing Go).

## Exit Criteria

- `docker compose up --build` brings up a runnable local demo.
- README is enough for another engineer to reproduce results.
