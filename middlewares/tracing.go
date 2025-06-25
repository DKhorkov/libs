package middlewares

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"google.golang.org/grpc/metadata"

	"github.com/DKhorkov/libs/logging"
	"github.com/DKhorkov/libs/tracing"
)

// TracingMiddleware creates root span of request and logs its Start and End events.
func TracingMiddleware(
	next http.Handler,
	logger logging.Logger,
	tp tracing.Provider,
	spanConfig tracing.SpanConfig,
) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == metricsURLPath {
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

		// Parsing response body for investigating errors:
		var response map[string]any
		if err := json.Unmarshal(trw.Body, &response); err != nil {
			logger.InfoContext(
				r.Context(),
				"Failed to parse response body for tracing errors purpose",
				err,
			)

			return
		}

		// Check errors section in response body:
		if errorsSection, ok := response["errors"].([]any); ok && len(errorsSection) > 0 {
			concatenatedErrBuilder := strings.Builder{}
			concatenatedErrBuilder.WriteString(
				"Next errors were received during processing request:\n",
			)

			for i, errInfo := range errorsSection {
				errInfo, ok := errInfo.(map[string]any)
				if !ok {
					logger.InfoContext(
						r.Context(),
						"Failed to parse response body for tracing errors purpose\n",
					)

					continue
				}

				concatenatedErrBuilder.WriteString(
					fmt.Sprintf(
						"%d) Message: %s; Path: %s\n",
						i+1,
						errInfo["message"],
						errInfo["path"],
					),
				)
			}

			span.SetStatus(tracing.StatusError, concatenatedErrBuilder.String())
		}
	})
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
