import grpc

_stub = None
_service_name = None


def init(endpoint: str, service: str) -> None:
    global _stub, _service_name

    from infertrace._proto import span_pb2_grpc

    channel = grpc.insecure_channel(endpoint)
    _stub = span_pb2_grpc.CollectorServiceStub(channel)
    _service_name = service


def send_span(model_name: str, start_time_unix_nano: int, duration_nanos: int) -> None:
    if _stub is None:
        raise RuntimeError("infertrace not initialized — call infertrace.init() first")

    from infertrace._proto import span_pb2

    span = span_pb2.Span(
        service_name=_service_name,
        model_name=model_name,
        start_time_unix_nano=start_time_unix_nano,
        duration_nanos=duration_nanos,
    )
    _stub.SendSpanBatch(span_pb2.SendSpanBatchRequest(spans=[span]))
