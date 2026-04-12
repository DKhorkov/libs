package intercepting_response_writer

import (
	"bufio"
	"errors"
	"net"
	"net/http"
)

var ErrResponseDoesNotImplementHijacker = errors.New(
	"websocket: response does not implement http.Hijacker",
)

func New(w http.ResponseWriter) *InterceptingResponseWriter {
	return &InterceptingResponseWriter{ResponseWriter: w, StatusCode: http.StatusOK}
}

// InterceptingResponseWriter intercepts response from GraphQL for checking errors.
type InterceptingResponseWriter struct {
	http.ResponseWriter

	StatusCode int
	Body       []byte
}

// WriteHeader intercepts response body for later usage in trace.Span.
func (rw *InterceptingResponseWriter) WriteHeader(statusCode int) {
	rw.StatusCode = statusCode
	rw.ResponseWriter.WriteHeader(statusCode)
}

// Write intercepts response body for later usage in trace.Span.
func (rw *InterceptingResponseWriter) Write(body []byte) (int, error) {
	rw.Body = body

	return rw.ResponseWriter.Write(body)
}

func (rw *InterceptingResponseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	h, ok := rw.ResponseWriter.(http.Hijacker)
	if !ok {
		return nil, nil, ErrResponseDoesNotImplementHijacker
	}

	return h.Hijack()
}
