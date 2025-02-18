package tracing

import (
	"context"

	"go.opentelemetry.io/otel/trace"
)

// Provider interface is created for usage in external application according to
// "dependency inversion principle" of SOLID due to working via abstractions.
//
//go:generate mockgen -source=interfaces.go -destination=mocks/providers.go -package=mocks
type Provider interface {
	Shutdown(ctx context.Context) error
	Span(ctx context.Context, name string, opts ...trace.SpanStartOption) (context.Context, trace.Span)
	TraceIDFromHex(traceHex string) (trace.TraceID, error)
	SpanFromTraceID(
		ctx context.Context,
		traceID trace.TraceID,
		name string,
		opts ...trace.SpanStartOption,
	) (context.Context, trace.Span)
}
