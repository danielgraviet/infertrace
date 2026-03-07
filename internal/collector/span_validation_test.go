package collector

import "testing"

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
