# Phase 8: Analysis Engine — Anomaly Detection + Insights

> Raw trace data is useful, but actionable insights are valuable. This phase builds the Python analysis engine that continuously reads spans and surfaces problems: latency spikes, GPU underutilization, poor batching. You're back in Python here — this is where Python's data tooling shines.

---

## Learning Objectives

By the end of this phase you will understand:

- How rolling window statistics work (mean, percentiles over a sliding time window)
- How threshold-based anomaly detection works — simple but effective
- How to run a Python service as a background loop with configurable polling
- How to separate detection logic (is this anomalous?) from reporting logic (what should we say?)
- Why baselines matter — a 200ms latency spike means nothing without a baseline

---

## Tasks

1. **Create the analysis engine directory:**
   ```
   analysis/
   ├── __init__.py
   ├── main.py              ← entry point / polling loop
   ├── windows.py           ← rolling window statistics
   ├── detectors.py         ← anomaly detector implementations
   ├── insights.py          ← human-readable recommendation generation
   └── client.py            ← HTTP client to fetch spans from Query API
   ```

2. **Implement `RollingWindow`** in `windows.py`:
   ```python
   from collections import deque
   import statistics

   class RollingWindow:
       """Tracks the last N values and computes statistics."""
       def __init__(self, max_size: int = 1000):
           self._data = deque(maxlen=max_size)

       def add(self, value: float):
           self._data.append(value)

       def mean(self) -> float:
           return statistics.mean(self._data) if self._data else 0.0

       def p99(self) -> float:
           if not self._data:
               return 0.0
           sorted_data = sorted(self._data)
           idx = int(len(sorted_data) * 0.99)
           return sorted_data[idx]

       def count(self) -> int:
           return len(self._data)
   ```

3. **Define the `Anomaly` dataclass** in `detectors.py`:
   ```python
   from dataclasses import dataclass

   @dataclass
   class Anomaly:
       kind: str           # "latency_spike", "gpu_underutil", "poor_batching"
       model_name: str
       severity: str       # "warning" or "critical"
       value: float        # the observed value
       baseline: float     # what we expected
       details: dict
   ```

4. **Implement the `LatencySpikeDetector`:**
   ```python
   class LatencySpikeDetector:
       def __init__(self, threshold_multiplier: float = 3.0):
           self._windows: dict[str, RollingWindow] = {}
           self._threshold = threshold_multiplier

       def observe(self, model_name: str, duration_ms: float):
           if model_name not in self._windows:
               self._windows[model_name] = RollingWindow()
           self._windows[model_name].add(duration_ms)

       def detect(self, model_name: str, current_p99: float) -> Anomaly | None:
           window = self._windows.get(model_name)
           if not window or window.count() < 10:  # need baseline
               return None
           baseline = window.mean()
           if current_p99 > baseline * self._threshold:
               return Anomaly(
                   kind="latency_spike",
                   model_name=model_name,
                   severity="critical",
                   value=current_p99,
                   baseline=baseline,
                   details={"multiplier": current_p99 / baseline},
               )
           return None
   ```

5. **Implement `GPUUnderutilizationDetector`:**
   ```python
   class GPUUnderutilizationDetector:
       def __init__(self, util_threshold: float = 40.0):
           self._threshold = util_threshold

       def detect(self, model_name: str, gpu_util: float, queue_depth: int) -> Anomaly | None:
           if gpu_util < self._threshold and queue_depth > 0:
               return Anomaly(
                   kind="gpu_underutil",
                   model_name=model_name,
                   severity="warning",
                   value=gpu_util,
                   baseline=self._threshold,
                   details={"queue_depth": queue_depth},
               )
           return None
   ```

6. **Implement `BatchingEfficiencyDetector`:**
   ```python
   class BatchingEfficiencyDetector:
       def __init__(self, min_fill_ratio: float = 0.3):
           self._min_ratio = min_fill_ratio

       def detect(self, model_name: str, batch_fill_ratio: float) -> Anomaly | None:
           if 0 < batch_fill_ratio < self._min_ratio:
               return Anomaly(
                   kind="poor_batching",
                   model_name=model_name,
                   severity="warning",
                   value=batch_fill_ratio,
                   baseline=self._min_ratio,
                   details={},
               )
           return None
   ```

7. **Implement `generate_recommendation`** in `insights.py` — convert anomalies to human-readable text:
   ```python
   def generate_recommendation(anomaly: Anomaly) -> str:
       if anomaly.kind == "latency_spike":
           return (
               f"Model '{anomaly.model_name}' p99 latency is "
               f"{anomaly.value:.1f}ms — {anomaly.details['multiplier']:.1f}x above baseline "
               f"({anomaly.baseline:.1f}ms). Check for memory pressure or batch size changes."
           )
       elif anomaly.kind == "gpu_underutil":
           return (
               f"Model '{anomaly.model_name}' GPU utilization is {anomaly.value:.1f}% "
               f"with a queue depth of {anomaly.details['queue_depth']}. "
               f"Batch timeout may be too aggressive."
           )
       elif anomaly.kind == "poor_batching":
           return (
               f"Model '{anomaly.model_name}' average batch fill ratio is "
               f"{anomaly.value:.0%}. Consider increasing the batch wait timeout."
           )
       return f"Anomaly detected for model '{anomaly.model_name}': {anomaly.kind}"
   ```

8. **Build the polling loop** in `main.py`:
   ```python
   import time
   import httpx

   def run_analysis_loop(api_url: str, poll_interval: int = 30):
       latency_detector = LatencySpikeDetector()
       gpu_detector = GPUUnderutilizationDetector()
       batch_detector = BatchingEfficiencyDetector()

       while True:
           spans = httpx.get(f"{api_url}/spans").json()
           for span in spans:
               model = span.get("model_name")
               if not model:
                   continue
               latency_detector.observe(model, span["duration_ms"])
               anomaly = latency_detector.detect(model, span["duration_ms"])
               if anomaly:
                   print(f"[ALERT] {generate_recommendation(anomaly)}")
           time.sleep(poll_interval)

   if __name__ == "__main__":
       run_analysis_loop("http://localhost:8080")
   ```

9. **Test the detectors in isolation** — write unit tests for each detector:
   - Feed 100 normal values (50ms), then one spike (500ms) — verify `LatencySpikeDetector` fires
   - Feed a span with `gpu_util=20%` and `queue_depth=5` — verify `GPUUnderutilizationDetector` fires
   - These should be pure Python unit tests with no network calls

10. **Run it end-to-end:** Start your collector, send spans with the Python SDK demo, start the analysis engine, then manually inject a latency spike (sleep longer in the demo). Watch the alert appear.
