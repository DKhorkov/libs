package http

import (
	"google.golang.org/grpc/metadata"
	"net/http"

	"github.com/DKhorkov/libs/tracing"
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

			// Create new traceResponseWriter for response intercepting purpose:
			trw := newTracingResponseWriter(w)
			next.ServeHTTP(trw, r)

			if trw.StatusCode >= http.StatusBadRequest {
				span.SetStatus(tracing.StatusError, string(trw.Body))
			}
		})
	}
}

func newTracingResponseWriter(w http.ResponseWriter) *tracingResponseWriter {
	return &tracingResponseWriter{ResponseWriter: w, StatusCode: http.StatusOK}
}

// tracingResponseWriter intercepts response from GraphQL for checking errors.
type tracingResponseWriter struct {
	http.ResponseWriter
	StatusCode int
	Body       []byte
}

// WriteHeader intercepts response body for later usage in trace.Span.
func (trw *tracingResponseWriter) WriteHeader(statusCode int) {
	trw.StatusCode = statusCode
	trw.ResponseWriter.WriteHeader(statusCode)
}

// Write intercepts response body for later usage in trace.Span.
func (trw *tracingResponseWriter) Write(body []byte) (int, error) {
	trw.Body = body

	return trw.ResponseWriter.Write(body)
}
