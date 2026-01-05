package http

import (
	"bufio"
	"errors"
	"net"
	"net/http"
)

var ErrResponseDoesNotImplementHijacker = errors.New(
	"websocket: response does not implement http.Hijacker",
)

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
func (rw *interceptingResponseWriter) WriteHeader(statusCode int) {
	rw.StatusCode = statusCode
	rw.ResponseWriter.WriteHeader(statusCode)
}

// Write intercepts response body for later usage in trace.Span.
func (rw *interceptingResponseWriter) Write(body []byte) (int, error) {
	rw.Body = body

	return rw.ResponseWriter.Write(body)
}

func (rw *interceptingResponseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	h, ok := rw.ResponseWriter.(http.Hijacker)
	if !ok {
		return nil, nil, ErrResponseDoesNotImplementHijacker
	}

	return h.Hijack()
}
