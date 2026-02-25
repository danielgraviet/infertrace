package main

import (
	"errors"
	"fmt"
	"github.com/danielgraviet/infertrace/internal/collector"
)

func ParseTraceID(raw string) (string, error) {
	if raw == "" {
		return "", errors.New("Error parsing trace ID")
	}

	finalTraceID := raw
	return finalTraceID, nil
}

// create a new struct mock object
// pass in the trace ID to my function
// make sure it is robust. 

func main() {
	span := collector.NewSpan("auth-service", "validate-token")

	traceID, err := ParseTraceID("abc-123") // important to understand what the function purpose is. I thought we were parsing an existing one and validating. 
	if err != nil {
		fmt.Println("error: ", err)
		return
	}

	span.TraceID = traceID
	fmt.Println("Created span: ", span.SpanID, span.ServiceName)
}