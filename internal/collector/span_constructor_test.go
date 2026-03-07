package collector

import "testing"

func TestNewSpan_SetsRequiredFields(t *testing.T) {
	span, err := NewSpan("inference-api", "predict", "gpt-4o-mini")
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if span.SpanID == "" {
		t.Fatalf("expected non-empty SpanID")
	}
	if span.ServiceName != "inference-api" {
		t.Fatalf("got %q, want inference-api", span.ServiceName)
	}
	if span.OperationName != "predict" {
		t.Fatalf("got %q, want predict", span.OperationName)
	}
	if span.ModelName != "gpt-4o-mini" {
		t.Fatalf("got %q, want gpt-4o-mini", span.ModelName)
	}
	if span.StartTimeUnixNano <= 0 {
		t.Fatalf("expected StartTimeUnixNano > 0")
	}
}

func TestNewSpan_RejectsMissingRequiredFields(t *testing.T) {
	_, err := NewSpan("", "predict", "gpt-4o-mini")
	if err == nil {
		t.Fatalf("expected error for missing service name")
	}

	_, err = NewSpan("inference-api", "", "gpt-4o-mini")
	if err == nil {
		t.Fatalf("expected error for missing operation name")
	}

	_, err = NewSpan("inference-api", "predict", "")
	if err == nil {
		t.Fatalf("expected error for missing model name")
	}
}
