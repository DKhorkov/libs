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

// MaxCapturedBodySize — сколько байт тела ответа держит перехват.
//
// Копия тела нужна логам и статусу спана, а не отдаче: клиенту ответ уходит
// целиком. Без потолка отдача файла держала бы его копию в памяти до конца
// запроса.
const MaxCapturedBodySize = 1 << 20

func New(w http.ResponseWriter) *InterceptingResponseWriter {
	return &InterceptingResponseWriter{ResponseWriter: w, StatusCode: http.StatusOK}
}

// InterceptingResponseWriter intercepts response from GraphQL for checking errors.
type InterceptingResponseWriter struct {
	http.ResponseWriter

	// StatusCode — код ответа; без явного WriteHeader остаётся 200.
	StatusCode int

	// Body — копия тела ответа, склеенная из всех вызовов Write, не длиннее
	// MaxCapturedBodySize.
	Body []byte

	// BodyTruncated — тело длиннее MaxCapturedBodySize, и Body хранит только
	// его начало.
	BodyTruncated bool
}

// WriteHeader intercepts response body for later usage in trace.Span.
func (rw *InterceptingResponseWriter) WriteHeader(statusCode int) {
	rw.StatusCode = statusCode
	rw.ResponseWriter.WriteHeader(statusCode)
}

// Write intercepts response body for later usage in trace.Span.
//
// Тело копируется, а не запоминается ссылкой: вызывающий вправе
// переиспользовать буфер сразу после Write. Так делает fmt.Fprintln внутри
// http.Error — буфер возвращается в пул fmt, и следующий fmt.Sprintf (например,
// traceback логгера) переписывал перехваченное тело до того, как его читала
// трассировка.
func (rw *InterceptingResponseWriter) Write(body []byte) (int, error) {
	if room := MaxCapturedBodySize - len(rw.Body); room < len(body) {
		rw.Body = append(rw.Body, body[:max(room, 0)]...)
		rw.BodyTruncated = true
	} else {
		rw.Body = append(rw.Body, body...)
	}

	return rw.ResponseWriter.Write(body)
}

func (rw *InterceptingResponseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	h, ok := rw.ResponseWriter.(http.Hijacker)
	if !ok {
		return nil, nil, ErrResponseDoesNotImplementHijacker
	}

	return h.Hijack()
}
