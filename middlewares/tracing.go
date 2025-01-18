package middlewares

import (
	"net/http"

	"github.com/DKhorkov/libs/tracing"
	"google.golang.org/grpc/metadata"
)

// TracingMiddleware creates root span of request and logs its Start and End events.
func TracingMiddleware(next http.Handler, tp tracing.TraceProvider, spanConfig tracing.SpanConfig) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, span := tp.Span(
			r.Context(),
			spanConfig.Name,
			spanConfig.Opts...,
		)

		defer span.End()
		span.AddEvent(spanConfig.Events.Start.Name, spanConfig.Events.Start.Opts...)

		traceID := span.SpanContext().TraceID().String()
		ctx = metadata.AppendToOutgoingContext(ctx, tracing.Key, traceID) // setting for cross-service usage
		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)

		span.AddEvent(spanConfig.Events.End.Name, spanConfig.Events.End.Opts...)
	})
}
