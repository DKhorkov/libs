package metrics

import (
	"bufio"
	"net"
	"net/http"
)

// responseRecorder запоминает код ответа и число записанных байт.
//
// Тело не копируется: метрикам нужен только размер, а копия держала бы
// ответ в памяти до конца запроса.
type responseRecorder struct {
	http.ResponseWriter

	// status — код ответа; без явного WriteHeader — 200.
	status int

	// written — сколько байт тела ушло клиенту.
	written int

	// wroteHeader — код уже зафиксирован.
	wroteHeader bool
}

// WriteHeader фиксирует первый окончательный код и передаёт его дальше.
//
// Информационный ответ (1xx, например 103 Early Hints) окончательным не
// является: обработчик шлёт его до настоящего кода, и метка несла бы 103 вместо
// итогового. 101 Switching Protocols — исключение, это и есть ответ.
func (r *responseRecorder) WriteHeader(code int) {
	if !r.wroteHeader && !informational(code) {
		r.status = code
		r.wroteHeader = true
	}

	r.ResponseWriter.WriteHeader(code)
}

// informational — код 1xx, после которого последует окончательный ответ.
func informational(code int) bool {
	return code >= http.StatusContinue &&
		code < http.StatusOK &&
		code != http.StatusSwitchingProtocols
}

// Write считает байты тела.
func (r *responseRecorder) Write(body []byte) (int, error) {
	r.wroteHeader = true

	n, err := r.ResponseWriter.Write(body)
	r.written += n

	return n, err
}

// Unwrap отдаёт обёрнутый writer http.ResponseController'у.
func (r *responseRecorder) Unwrap() http.ResponseWriter {
	return r.ResponseWriter
}

// Flush передаёт сброс буфера, если обёрнутый writer его умеет.
func (r *responseRecorder) Flush() {
	if flusher, ok := r.ResponseWriter.(http.Flusher); ok {
		flusher.Flush()
	}
}

// Hijack передаёт захват соединения, если обёрнутый writer его умеет.
func (r *responseRecorder) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	hijacker, ok := r.ResponseWriter.(http.Hijacker)
	if !ok {
		return nil, nil, ErrHijackUnsupported
	}

	return hijacker.Hijack()
}
