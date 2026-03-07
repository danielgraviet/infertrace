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

func TestSpanValidateForIngest(t *testing.T) {
	tests := []struct {
		name    string
		span    Span
		wantErr bool
	}{
		{
			name: "valid span",
			span: Span{
				ServiceName:       "inference-api",
				ModelName:         "gpt-4o-mini",
				StartTimeUnixNano: 1,
				DurationNanos:     10,
			},
			wantErr: false,
		},
		{
			name: "missing service name",
			span: Span{
				ModelName:         "gpt-4o-mini",
				StartTimeUnixNano: 1,
				DurationNanos:     10,
			},
			wantErr: true,
		},
		{
			name: "missing model name",
			span: Span{
				ServiceName:       "inference-api",
				StartTimeUnixNano: 1,
				DurationNanos:     10,
			},
			wantErr: true,
		},
		{
			name: "missing start time",
			span: Span{
				ServiceName:   "inference-api",
				ModelName:     "gpt-4o-mini",
				DurationNanos: 10,
			},
			wantErr: true,
		},
		{
			name: "missing duration",
			span: Span{
				ServiceName:       "inference-api",
				ModelName:         "gpt-4o-mini",
				StartTimeUnixNano: 1,
			},
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.span.ValidateForIngest()
			if tc.wantErr && err == nil {
				t.Fatalf("expected error, got nil")
			}
			if !tc.wantErr && err != nil {
				t.Fatalf("expected nil error, got %v", err)
			}
		})
	}
}
