// go test ./internal/collector/

package collector

import "testing"

func TestNewSpan_SetsSpanID(t *testing.T) {
	// call new span
	span := NewSpan("Testing", "validate")
	
	// check that span ID is not empty
	spanID := span.SpanID
	if spanID == "" {
		t.Errorf("got %q, want non-empty string", spanID) // %q does quotes
	}
}

func TestNewSpan_SetsServiceName(t *testing.T) {
	// call new span with "my-service"
	span := NewSpan("my-service", "validate")
	
	// check that span.servicename == "my_service"
	serviceName := span.ServiceName
	if serviceName != "my-service" {
		t.Errorf("got %q, want my-service", serviceName) // %q does quotes
	}
}