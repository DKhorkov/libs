package http

import (
	"net/http"

	"github.com/DKhorkov/libs/tracing"
	"google.golang.org/grpc/metadata"
)

// TracingMiddleware creates root span of request and logs its Start and End events.
func TracingMiddleware(
	tp tracing.Provider,
	spanConfig tracing.SpanConfig,
) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == MetricsURLPath {
				next.ServeHTTP(w, r)

				return
			}

			ctx, span := tp.Span(
				r.Context(),
				spanConfig.Name,
				spanConfig.Opts...,
			)

			defer span.End()

			span.AddEvent(spanConfig.Events.Start.Name, spanConfig.Events.Start.Opts...)
			defer span.AddEvent(spanConfig.Events.End.Name, spanConfig.Events.End.Opts...)

			traceID := span.SpanContext().TraceID().String()
			ctx = metadata.AppendToOutgoingContext(
				ctx,
				tracing.Key,
				traceID,
			) // setting for cross-service usage
			r = r.WithContext(ctx)

			// Create new newInterceptingResponseWriter for response intercepting purpose:
			trw := newInterceptingResponseWriter(w)
			next.ServeHTTP(trw, r)

			if trw.StatusCode >= http.StatusBadRequest {
				span.SetStatus(tracing.StatusError, string(trw.Body))
			}
		})
	}
}
