package tracing

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	"go.opentelemetry.io/otel/trace"
)

const (
	Key = "x-trace-id"
)

// New create new *CommonProvider for creating spans for traces.
func New(config Config, opts ...trace.TracerOption) (*CommonProvider, error) {
	// Setting jaeger endpoint for viewing traces:
	exporter, err := jaeger.New(jaeger.WithCollectorEndpoint(
		jaeger.WithEndpoint(config.JaegerURL),
	))

	if err != nil {
		return nil, err
	}

	// Setting service info:
	res := resource.NewWithAttributes(
		semconv.SchemaURL,
		semconv.ServiceNameKey.String(config.ServiceName),
		semconv.ServiceVersionKey.String(config.ServiceVersion),
	)

	// Creating provider:
	traceProvider := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sdktrace.AlwaysSample()), // Configuring, how much traces will be saved
	)

	otel.SetTracerProvider(traceProvider) // Set provider as global

	provider := &CommonProvider{
		traceProvider: traceProvider,
		tracer:        otel.Tracer(config.ServiceName, opts...),
	}

	return provider, nil
}

// CommonProvider provides the ability to easily create spans for tracing.
type CommonProvider struct {
	traceProvider *sdktrace.TracerProvider
	tracer        trace.Tracer
}

// Shutdown correctly shutdowns CommonProvider's inner logic.
func (tp *CommonProvider) Shutdown(ctx context.Context) error {
	return tp.traceProvider.Shutdown(ctx)
}

// Span creates new span.
func (tp *CommonProvider) Span(
	ctx context.Context,
	name string,
	opts ...trace.SpanStartOption,
) (context.Context, trace.Span) {
	return tp.tracer.Start(ctx, name, opts...)
}

// SpanFromTraceID creates new span on base of provided trace.TraceID.
func (tp *CommonProvider) SpanFromTraceID(
	ctx context.Context,
	traceID trace.TraceID,
	name string,
	opts ...trace.SpanStartOption,
) (context.Context, trace.Span) {
	// Creating a span context with a predefined trace-id:
	spanContext := trace.NewSpanContext(trace.SpanContextConfig{
		TraceID: traceID,
	})

	// Embedding span config into the context:
	ctx = trace.ContextWithSpanContext(ctx, spanContext)

	return tp.Span(ctx, name, opts...)
}

// TraceIDFromHex decodes trace.TraceID from hash.
func (tp *CommonProvider) TraceIDFromHex(traceHex string) (trace.TraceID, error) {
	return trace.TraceIDFromHex(traceHex)
}
