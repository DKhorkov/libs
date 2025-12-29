package http

import "net/http"

func newInterceptingResponseWriter(w http.ResponseWriter) *interceptingResponseWriter {
	return &interceptingResponseWriter{ResponseWriter: w, StatusCode: http.StatusOK}
}

// interceptingResponseWriter intercepts response from GraphQL for checking errors.
type interceptingResponseWriter struct {
	http.ResponseWriter

	StatusCode int
	Body       []byte
}

// WriteHeader intercepts response body for later usage in trace.Span.
func (trw *interceptingResponseWriter) WriteHeader(statusCode int) {
	trw.StatusCode = statusCode
	trw.ResponseWriter.WriteHeader(statusCode)
}

// Write intercepts response body for later usage in trace.Span.
func (trw *interceptingResponseWriter) Write(body []byte) (int, error) {
	trw.Body = body

	return trw.ResponseWriter.Write(body)
}
