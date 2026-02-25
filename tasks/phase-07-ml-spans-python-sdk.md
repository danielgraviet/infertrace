# Phase 7: ML Span Attributes + Python SDK

> This phase adds the ML-specific intelligence that separates InferTrace from generic tracers. You'll extend the proto schema with inference metadata, then build the Python SDK that ML engineers will use to instrument their models. You already have gRPC Python stubs in `grpc/` — this builds directly on that work.

---

## Learning Objectives

By the end of this phase you will understand:

- How to extend a protobuf schema without breaking existing clients (field numbering rules)
- How the decorator pattern works in Python for wrapping functions transparently
- How Python context managers (`with` statement) are implemented with `__enter__`/`__exit__`
- How `pynvml` reads GPU metrics from NVIDIA hardware
- How to design a clean SDK API that hides gRPC complexity from the user
- What graceful degradation means — the SDK should work even with no GPU present

---

## Tasks

1. **Extend `proto/span.proto`** — your `InferenceContext` and `ResourceMetrics` messages already exist from Phase 2. Add these fields:
   ```protobuf
   message InferenceContext {
     // existing fields...
     string framework = 6;          // "pytorch", "tensorflow", "onnx"
     float sequence_length = 7;
     float cache_hit_rate = 8;
     string pipeline_stage = 9;     // "preprocessing", "inference", "postprocessing"
     int32 queue_depth = 10;
     float queue_wait_time_ms = 11;
     float batch_fill_ratio = 12;
   }

   message ResourceMetrics {
     // existing fields...
     string gpu_device_id = 4;
     double cpu_utilization_percent = 5;
     int64 memory_heap_bytes = 6;
   }
   ```
   **Important:** Never reuse a field number. Removed fields should be marked `reserved`. Adding new fields with new numbers is always safe.

2. **Regenerate both Go and Python stubs:**
   ```bash
   # Go
   protoc --go_out=. --go-grpc_out=. proto/span.proto

   # Python (you'll need grpcio-tools)
   python -m grpc_tools.protoc -I. --python_out=sdk/python/ --grpc_python_out=sdk/python/ proto/span.proto
   ```

3. **Create the SDK directory structure:**
   ```
   sdk/
   └── python/
       ├── infertrace/
       │   ├── __init__.py
       │   ├── tracer.py       ← Tracer class
       │   ├── span.py         ← Span context manager
       │   └── gpu.py          ← GPU metrics via pynvml
       └── pyproject.toml
   ```

4. **Implement `gpu.py`** — wrap pynvml with graceful fallback:
   ```python
   try:
       import pynvml
       pynvml.nvmlInit()
       _GPU_AVAILABLE = True
   except Exception:
       _GPU_AVAILABLE = False

   def get_gpu_metrics(device_index: int = 0) -> dict:
       if not _GPU_AVAILABLE:
           return {}
       handle = pynvml.nvmlDeviceGetHandleByIndex(device_index)
       mem = pynvml.nvmlDeviceGetMemoryInfo(handle)
       util = pynvml.nvmlDeviceGetUtilizationRates(handle)
       return {
           "gpu_utilization_percent": float(util.gpu),
           "gpu_memory_used_bytes": mem.used,
           "gpu_memory_total_bytes": mem.total,
           "gpu_device_id": str(device_index),
       }
   ```

5. **Implement the `Span` class** as a context manager in `span.py`:
   ```python
   class Span:
       def __init__(self, tracer, model_name: str, service_name: str):
           self._tracer = tracer
           self._pb_span = proto.Span(
               trace_id=...,
               span_id=...,
               service_name=service_name,
               inference=proto.InferenceContext(model_name=model_name),
           )
           self._start = time.time_ns()

       def set_batch_size(self, size: int):
           self._pb_span.inference.batch_size = size

       def set_tokens(self, input: int, output: int):
           self._pb_span.inference.input_tokens = input
           self._pb_span.inference.output_tokens = output

       def __enter__(self):
           return self

       def __exit__(self, *_):
           self._pb_span.duration_nanos = time.time_ns() - self._start
           # Capture GPU metrics automatically on span end
           gpu = get_gpu_metrics()
           if gpu:
               self._pb_span.resources.CopyFrom(proto.ResourceMetrics(**gpu))
           self._tracer._send(self._pb_span)
   ```

6. **Implement the `Tracer` class** in `tracer.py`:
   ```python
   import grpc
   from . import span as span_module
   from .generated import span_pb2_grpc

   class Tracer:
       def __init__(self, endpoint: str = "localhost:4317"):
           channel = grpc.insecure_channel(endpoint)
           self._stub = span_pb2_grpc.CollectorServiceStub(channel)
           self._service_name = "unknown"

       def configure(self, service_name: str):
           self._service_name = service_name
           return self

       def start_span(self, model_name: str) -> span_module.Span:
           return span_module.Span(self, model_name, self._service_name)

       def _send(self, pb_span):
           try:
               self._stub.SendSpan(SendSpanRequest(span=pb_span), timeout=1.0)
           except grpc.RpcError:
               pass  # Never let tracing crash the user's code
   ```

7. **Implement the `@trace_inference` decorator:**
   ```python
   def trace_inference(model: str):
       def decorator(fn):
           def wrapper(*args, **kwargs):
               with tracer.start_span(model) as span:
                   result = fn(*args, **kwargs)
                   return result
           return wrapper
       return decorator
   ```

8. **Write a demo script** `sdk/python/demo.py` that simulates an inference call:
   ```python
   from infertrace import Tracer

   tracer = Tracer(endpoint="localhost:4317").configure(service_name="demo-model-server")

   def run_batch(inputs):
       with tracer.start_span("gpt-4") as span:
           span.set_batch_size(len(inputs))
           import time; time.sleep(0.05)  # simulate inference
           span.set_tokens(input=sum(len(i) for i in inputs), output=100)
           return ["result"] * len(inputs)

   for _ in range(10):
       run_batch(["prompt 1", "prompt 2", "prompt 3"])
       time.sleep(0.1)
   ```
   Run this while your Go collector is running and verify the spans appear in TimescaleDB.

9. **Verify end-to-end:** Run collector → run `demo.py` → `curl http://localhost:8080/spans?model=gpt-4`. You should see real spans with ML metadata.
