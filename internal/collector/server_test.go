package collector

import (
	"context"
	"testing"
	"time"

	infertracepb "github.com/danielgraviet/infertrace/proto"
)

func TestSendSpanBatch_AllValid(t *testing.T) {
	pipeline := NewCollector(1, 10, nil)
	t.Cleanup(func() {
		_ = pipeline.Stop(context.Background())
	})
	server := NewServer(pipeline)

	req := &infertracepb.SendSpanBatchRequest{
		Spans: []*infertracepb.Span{
			{
				TraceId:           "trace-1",
				SpanId:            "span-1",
				ServiceName:       "auth-service",
				OperationName:     "validate-token",
				ModelName:         "gpt-4o-mini",
				StartTimeUnixNano: 100,
				DurationNanos:     10,
				Status:            "ok",
			},
			{
				TraceId:           "trace-2",
				SpanId:            "span-2",
				ServiceName:       "billing-service",
				OperationName:     "score-risk",
				ModelName:         "gpt-4.1-mini",
				StartTimeUnixNano: 200,
				DurationNanos:     20,
				Status:            "ok",
			},
		},
	}

	resp, err := server.SendSpanBatch(context.Background(), req)
	if err != nil {
		t.Fatalf("SendSpanBatch returned error: %v", err)
	}
	if got, want := resp.GetAcceptedCount(), int32(2); got != want {
		t.Fatalf("accepted_count mismatch: got=%d want=%d", got, want)
	}
	if got, want := resp.GetRejectedCount(), int32(0); got != want {
		t.Fatalf("rejected_count mismatch: got=%d want=%d", got, want)
	}
}

func TestSendSpanBatch_EmptyBatch(t *testing.T) {
	pipeline := NewCollector(1, 10, nil)
	t.Cleanup(func() {
		_ = pipeline.Stop(context.Background())
	})
	server := NewServer(pipeline)

	resp, err := server.SendSpanBatch(context.Background(), &infertracepb.SendSpanBatchRequest{})
	if err != nil {
		t.Fatalf("SendSpanBatch returned error: %v", err)
	}
	if got, want := resp.GetAcceptedCount(), int32(0); got != want {
		t.Fatalf("accepted_count mismatch: got=%d want=%d", got, want)
	}
	if got, want := resp.GetRejectedCount(), int32(0); got != want {
		t.Fatalf("rejected_count mismatch: got=%d want=%d", got, want)
	}
}

func TestSendSpanBatch_MalformedSpanRejected(t *testing.T) {
	pipeline := NewCollector(1, 10, nil)
	t.Cleanup(func() {
		_ = pipeline.Stop(context.Background())
	})
	server := NewServer(pipeline)

	req := &infertracepb.SendSpanBatchRequest{
		Spans: []*infertracepb.Span{
			{
				TraceId:           "trace-valid",
				SpanId:            "span-valid",
				ServiceName:       "auth-service",
				OperationName:     "validate-token",
				ModelName:         "gpt-4o-mini",
				StartTimeUnixNano: 100,
				DurationNanos:     10,
				Status:            "ok",
			},
			{
				TraceId:           "trace-invalid",
				SpanId:            "span-invalid",
				ServiceName:       "",
				OperationName:     "validate-token",
				ModelName:         "gpt-4o-mini",
				StartTimeUnixNano: 100,
				DurationNanos:     10,
				Status:            "error",
			},
		},
	}

	resp, err := server.SendSpanBatch(context.Background(), req)
	if err != nil {
		t.Fatalf("SendSpanBatch returned error: %v", err)
	}
	if got, want := resp.GetAcceptedCount(), int32(1); got != want {
		t.Fatalf("accepted_count mismatch: got=%d want=%d", got, want)
	}
	if got, want := resp.GetRejectedCount(), int32(1); got != want {
		t.Fatalf("rejected_count mismatch: got=%d want=%d", got, want)
	}
}

func TestSendSpanBatch_QueueFullRejected(t *testing.T) {
	pipeline := NewCollector(0, 1, nil)
	server := NewServer(pipeline)

	req := &infertracepb.SendSpanBatchRequest{
		Spans: []*infertracepb.Span{
			{
				TraceId:           "trace-1",
				SpanId:            "span-1",
				ServiceName:       "auth-service",
				OperationName:     "validate-token",
				ModelName:         "gpt-4o-mini",
				StartTimeUnixNano: 100,
				DurationNanos:     10,
				Status:            "ok",
			},
			{
				TraceId:           "trace-2",
				SpanId:            "span-2",
				ServiceName:       "auth-service",
				OperationName:     "validate-token",
				ModelName:         "gpt-4o-mini",
				StartTimeUnixNano: 100,
				DurationNanos:     10,
				Status:            "ok",
			},
		},
	}

	resp, err := server.SendSpanBatch(context.Background(), req)
	if err != nil {
		t.Fatalf("SendSpanBatch returned error: %v", err)
	}
	if got, want := resp.GetAcceptedCount(), int32(1); got != want {
		t.Fatalf("accepted_count mismatch: got=%d want=%d", got, want)
	}
	if got, want := resp.GetRejectedCount(), int32(1); got != want {
		t.Fatalf("rejected_count mismatch: got=%d want=%d", got, want)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := pipeline.Stop(ctx); err != nil {
		t.Fatalf("pipeline stop failed: %v", err)
	}
}
