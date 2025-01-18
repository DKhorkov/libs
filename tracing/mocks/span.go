package mocks

import (
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"go.opentelemetry.io/otel/trace/noop"
)

func NewMockSpan() MockSpan {
	return MockSpan{}
}

// MockSpan is a mock for testing in other projects.
type MockSpan struct{}

func (ms *MockSpan) End(options ...trace.SpanEndOption) {}

func (ms *MockSpan) AddEvent(name string, options ...trace.EventOption) {}

func (ms *MockSpan) AddLink(link trace.Link) {}

func (ms *MockSpan) IsRecording() bool {
	return true
}

func (ms *MockSpan) RecordError(err error, options ...trace.EventOption) {}

func (ms *MockSpan) SpanContext() trace.SpanContext {
	return trace.SpanContext{}
}

func (ms *MockSpan) SetStatus(code codes.Code, description string) {}

func (ms *MockSpan) SetName(name string) {}

func (ms *MockSpan) SetAttributes(kv ...attribute.KeyValue) {}

func (ms *MockSpan) TracerProvider() trace.TracerProvider {
	return noop.NewTracerProvider()
}
