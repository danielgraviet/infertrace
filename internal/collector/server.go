package collector

import (
	"context"

	infertracepb "github.com/danielgraviet/infertrace/proto"
)

type Server struct {
	infertracepb.UnimplementedCollectorServiceServer
	collector *Collector
}

func NewServer(collector *Collector) *Server {
	if collector == nil {
		collector = NewCollector(1, 1, nil)
	}
	return &Server{
		collector: collector,
	}
}

func (s *Server) SendSpanBatch(
	_ context.Context,
	req *infertracepb.SendSpanBatchRequest,
) (*infertracepb.SendSpanBatchResponse, error) {
	var accepted int32
	var rejected int32

	for _, pbSpan := range req.GetSpans() {
		span := Span{
			TraceID:           pbSpan.GetTraceId(),
			SpanID:            pbSpan.GetSpanId(),
			ServiceName:       pbSpan.GetServiceName(),
			OperationName:     pbSpan.GetOperationName(),
			ModelName:         pbSpan.GetModelName(),
			StartTimeUnixNano: pbSpan.GetStartTimeUnixNano(),
			DurationNanos:     pbSpan.GetDurationNanos(),
			Status:            pbSpan.GetStatus(),
		}

		if err := span.ValidateForIngest(); err != nil {
			s.collector.AddInvalid(1)
			rejected++
			continue
		}

		if err := s.collector.Enqueue(span); err != nil {
			rejected++
			continue
		}

		accepted++
	}

	return &infertracepb.SendSpanBatchResponse{
		AcceptedCount: accepted,
		RejectedCount: rejected,
	}, nil
}
