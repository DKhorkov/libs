package logging

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"

	"github.com/DKhorkov/libs/logging"
	"github.com/DKhorkov/libs/middlewares/http/intercepting_response_writer"
	"github.com/DKhorkov/libs/middlewares/http/metrics"
)

func Middleware(
	logger logging.Logger,
	sensitiveFields ...string,
) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Не логгируем сбор метрик:
			if r.URL.Path == metrics.MetricsURLPath {
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

			var payloadInput map[string]any

			if len(body) > 0 {
				if err = json.Unmarshal(body, &payloadInput); err != nil {
					logging.LogErrorContext(
						ctx,
						logger,
						"Failed to log request due to reading request body failure",
						err,
					)
				}

				for _, field := range sensitiveFields {
					delete(payloadInput, field)
				}
			}

			// Logging request info:
			logging.LogInfoContext(
				ctx,
				logger,
				"Received new request",
				[]any{
					"From", r.Host,
					"Method", r.Method,
					"URL", r.URL,
					"Headers", r.Header,
					"Query", r.URL.Query(),
					"Cookies", r.Cookies(),
					"Form", r.PostForm,
					"Payload", payloadInput,
				}...,
			)

			// Create new New for response intercepting purpose:
			rw := intercepting_response_writer.New(w)
			next.ServeHTTP(rw, r)

			// Делаем любого типа, чтобы обрабатывать и массивы, и словари:
			var payloadOutput any

			if len(rw.Body) > 0 {
				switch {
				case rw.StatusCode < http.StatusBadRequest:
					if err = json.Unmarshal(rw.Body, &payloadOutput); err != nil {
						logging.LogErrorContext(
							ctx,
							logger,
							"Failed to log response body due to reading body failure",
							err,
						)
					}
				default:
					// Ошибки пишутся как обычные строки в тело ответа:
					payloadOutput = map[string]string{"error": string(rw.Body)}
				}
			}

			// Logging response info:
			logging.LogInfoContext(
				ctx,
				logger,
				"Received response",
				[]any{
					"For", r.Host,
					"Method", r.Method,
					"URL", r.URL,
					"StatusCode", rw.StatusCode,
					"Headers", rw.Header(),
					"Payload", payloadOutput,
				}...,
			)
		})
	}
}
