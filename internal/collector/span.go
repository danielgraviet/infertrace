package collector

import (
	"time"
	"github.com/google/uuid"
)

type Span struct {
	TraceID string
	SpanID string
	ParentSpanID string
	ServiceName string
	OperationName string
	StartTimeUnixNano int64
	DurationNanos int64
	Status string
}

func NewSpan(serviceName, operationName string) *Span {
	return &Span{
		SpanID: uuid.New().String(),
		ServiceName: serviceName,
		OperationName: operationName,
		StartTimeUnixNano: time.Now().UnixNano(),
		// what happens to the rest of the fields I do not create?
	}
}

