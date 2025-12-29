package http

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/DKhorkov/libs/logging"
)

func LoggingMiddleware(
	logger logging.Logger,
	sensitiveFields ...string,
) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Не логгируем сбор метрик:
			if r.URL.Path == MetricsURLPath {
				next.ServeHTTP(w, r)

				return
			}

			ctx := r.Context()

			// Reading request body:
			body, err := io.ReadAll(r.Body)
			if err != nil {
				logging.LogErrorContext(
					ctx,
					logger,
					"Failed to log request due to reading request body failure",
					err,
				)
			}

			// Restoring request body for later usage due to the fact that io.Reader can be read only once:
			r.Body = io.NopCloser(bytes.NewBuffer(body))

			var payload map[string]any

			if len(body) > 0 {
				if err = json.Unmarshal(body, &payload); err != nil {
					logging.LogErrorContext(
						ctx,
						logger,
						"Failed to log request due to reading request body failure",
						err,
					)
				}

				for _, field := range sensitiveFields {
					delete(payload, field)
				}
			}

			// Logging request info:
			logging.LogInfoContext(
				ctx,
				logger,
				fmt.Sprintf(
					"Received new request: Method=%s, URL=%v, Headers=%v, Cookies=%v, Payload=%v\n",
					r.Method,
					r.URL,
					r.Header,
					r.Cookies(),
					payload,
				),
			)

			// Create new newInterceptingResponseWriter for response intercepting purpose:
			trw := newInterceptingResponseWriter(w)
			next.ServeHTTP(trw, r)

			// Обнуляем payload для логгирования ответа:
			payload = map[string]any{}

			if len(trw.Body) > 0 {
				switch {
				case trw.StatusCode >= http.StatusBadRequest:
					if err = json.Unmarshal(trw.Body, &payload); err != nil {
						logging.LogErrorContext(
							ctx,
							logger,
							"Failed to log response body due to reading body failure",
							err,
						)
					}
				default:
					// Ошибки пишутся как обычные строки в тело ответа:
					payload["error"] = string(
						trw.Body,
					)
				}
			}

			// Logging response info:
			logging.LogInfoContext(
				ctx,
				logger,
				fmt.Sprintf(
					"Received response: Method=%s, URL=%v, StatusCode=%d, Headers=%v, Payload=%v\n",
					r.Method,
					r.URL,
					trw.StatusCode,
					trw.Header(),
					payload,
				),
			)
		})
	}
}
