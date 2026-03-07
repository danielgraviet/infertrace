package collector

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

type Span struct {
	TraceID           string
	SpanID            string
	ParentSpanID      string
	ServiceName       string
	OperationName     string
	ModelName         string
	StartTimeUnixNano int64
	DurationNanos     int64
	Status            string
}

func NewSpan(serviceName, operationName, modelName string) (*Span, error) {
	span := &Span{
		SpanID:            uuid.New().String(),
		ServiceName:       serviceName,
		OperationName:     operationName,
		ModelName:         modelName,
		StartTimeUnixNano: time.Now().UnixNano(),
	}

	if span.ServiceName == "" {
		return nil, errors.New("service_name is required")
	}
	if span.OperationName == "" {
		return nil, errors.New("operation_name is required")
	}
	if span.ModelName == "" {
		return nil, errors.New("model_name is required")
	}

	return span, nil
}

func (s Span) ValidateForIngest() error {
	if s.ServiceName == "" {
		return errors.New("service_name is required")
	}
	if s.ModelName == "" {
		return errors.New("model_name is required")
	}
	if s.StartTimeUnixNano <= 0 {
		return errors.New("start_time_unix_nano must be > 0")
	}
	if s.DurationNanos <= 0 {
		return errors.New("duration_nanos must be > 0")
	}
	return nil
}
