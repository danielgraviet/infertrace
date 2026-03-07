import time
import functools
from infertrace import _client


def trace(model: str):
    def decorator(fn):
        @functools.wraps(fn)
        def wrapper(*args, **kwargs):
            start_ns = time.time_ns()
            result = fn(*args, **kwargs)
            duration_ns = time.time_ns() - start_ns
            _client.send_span(model, start_ns, duration_ns)
            return result
        return wrapper
    return decorator
