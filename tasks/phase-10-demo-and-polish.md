# Phase 10: Demo App, Docker Compose + Polish

> A system no one can run locally is a system no one will adopt. This phase makes InferTrace easy to spin up in one command, adds a realistic demo that shows off every feature, and writes the documentation that turns a GitHub repo into a project people take seriously.

---

## Learning Objectives

By the end of this phase you will understand:

- How Docker Compose orchestrates multi-service applications
- How to write a realistic load generator that simulates production-like traffic
- How to expose Prometheus metrics from a Go service
- What makes a README excellent vs mediocre
- How to inject anomalies programmatically to create compelling demos

---

## Tasks

### Docker Compose

1. **Write `docker-compose.yml`** at the repo root:
   ```yaml
   version: "3.9"
   services:
     timescaledb:
       image: timescale/timescaledb:latest-pg16
       environment:
         POSTGRES_PASSWORD: infertrace
         POSTGRES_DB: infertrace
       ports: ["5432:5432"]
       volumes: ["pgdata:/var/lib/postgresql/data"]

     collector:
       build:
         context: .
         dockerfile: cmd/collector/Dockerfile
       ports:
         - "4317:4317"   # gRPC
         - "8080:8080"   # REST API
       depends_on: [timescaledb]
       environment:
         DB_URL: postgres://postgres:infertrace@timescaledb:5432/infertrace

     analysis:
       build:
         context: .
         dockerfile: analysis/Dockerfile
       depends_on: [collector]
       environment:
         API_URL: http://collector:8080

   volumes:
     pgdata:
   ```

2. **Write `cmd/collector/Dockerfile`:**
   ```dockerfile
   FROM golang:1.23-alpine AS builder
   WORKDIR /app
   COPY go.mod go.sum ./
   RUN go mod download
   COPY . .
   RUN go build -o collector ./cmd/collector

   FROM alpine:3.20
   COPY --from=builder /app/collector /collector
   EXPOSE 4317 8080
   CMD ["/collector"]
   ```
   This is a multi-stage build — the final image contains only the compiled binary, not the Go toolchain. Result: a tiny (~10MB) image.

3. **Write `analysis/Dockerfile`:**
   ```dockerfile
   FROM python:3.12-slim
   WORKDIR /app
   COPY analysis/pyproject.toml .
   RUN pip install .
   COPY analysis/ .
   CMD ["python", "main.py"]
   ```

4. **Verify it works:**
   ```bash
   docker compose up --build
   ```
   All services should start, and the collector should connect to TimescaleDB on startup.

---

### Load Generator + Anomaly Injection

5. **Write the load generator** in `cmd/demo/main.go`. It should simulate a realistic inference pipeline:
   ```go
   type ScenarioConfig struct {
       ModelName   string
       BaseLatency time.Duration
       MaxBatch    int
       RPS         int
   }

   scenarios := []ScenarioConfig{
       {ModelName: "gpt-4", BaseLatency: 80 * time.Millisecond, MaxBatch: 32, RPS: 50},
       {ModelName: "llama-70b", BaseLatency: 200 * time.Millisecond, MaxBatch: 8, RPS: 20},
       {ModelName: "clip-vit", BaseLatency: 15 * time.Millisecond, MaxBatch: 64, RPS: 200},
   }
   ```
   For each RPS target, launch a goroutine that sends spans at the correct rate using `time.Ticker`.

6. **Add anomaly injection** — every 100th request for `gpt-4`, inject a latency spike:
   ```go
   var requestCount int64
   func simulateLatency(base time.Duration, modelName string) time.Duration {
       count := atomic.AddInt64(&requestCount, 1)
       if modelName == "gpt-4" && count%100 == 0 {
           // Inject a 10x spike — this will trigger the analysis engine alert
           return base * 10
       }
       // Add ±20% jitter for realism
       jitter := float64(base) * (0.8 + rand.Float64()*0.4)
       return time.Duration(jitter)
   }
   ```

7. **Add GPU underutilization simulation** — every 50th span, send low GPU util with a non-zero queue:
   ```go
   if count%50 == 0 {
       span.Resources = &proto.ResourceMetrics{
           GpuUtilizationPercent: 22.0,  // triggers the GPU detector
       }
       span.Inference.QueueDepth = 5
   }
   ```

---

### Prometheus Metrics

8. **Add Prometheus metrics to the collector** — install the client:
   ```bash
   go get github.com/prometheus/client_golang/prometheus
   go get github.com/prometheus/client_golang/prometheus/promhttp
   ```
   Add these metrics to the `Collector`:
   ```go
   spansIngested = prometheus.NewCounterVec(
       prometheus.CounterOpts{Name: "infertrace_spans_ingested_total"},
       []string{"service_name"},
   )
   queueDepth = prometheus.NewGauge(
       prometheus.GaugeOpts{Name: "infertrace_queue_depth"},
   )
   ```
   Expose them at `GET /metrics` (Prometheus scrapes this endpoint).

---

### Documentation

9. **Write the README** — this is your first impression. Structure:
   ```markdown
   # InferTrace

   > High-performance distributed tracing for AI/ML inference pipelines.

   ## Why InferTrace?
   [2-3 sentences on the problem: existing tracers are generic, ML engineers need model-specific visibility]

   ## Architecture
   [Diagram: Instrumented Service → Agent → Collector (Go) → TimescaleDB → Query API → Analysis Engine]

   ## Quick Start
   docker compose up
   # That's it. Spans UI at http://localhost:8080

   ## Instrumenting Your Model Server
   [Python SDK example — 5 lines of code]

   ## Performance
   [Your benchmark results from Phase 9]

   ## Project Structure
   [Brief explanation of each directory]
   ```

10. **Add a `CONTRIBUTING.md`** with:
    - How to run the stack locally
    - How to run tests (`go test ./...`, `pytest`)
    - How to run the linter (`golangci-lint run`)

11. **Final integration test — run the full stack:**
    ```bash
    docker compose up --build -d
    python sdk/python/demo.py          # send real spans
    go run ./cmd/demo                  # send load with anomalies
    curl http://localhost:8080/spans?model=gpt-4   # query the API
    # Watch the analysis engine print alerts in its logs
    docker compose logs -f analysis
    ```

    If all pieces connect — you're done. You've built a production-grade distributed tracing system from scratch.
