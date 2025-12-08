package graphql

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strconv"

	"github.com/prometheus/client_golang/prometheus"

	graphqlparser "github.com/DKhorkov/libs/graphql"
	"github.com/DKhorkov/libs/logging"
)

const (
	metricsURLPath = "/metrics"

	urlLabel        = "url"
	statusLabel     = "status"
	statusCodeLabel = "status_code"

	statusOK    = "ok"
	statusError = "error"
)

var (
	// requestsTotal PROMQL => rate(requests_total{}[5m]).
	requestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "requests_total",
			Help: "Number of HTTP requests by url.",
		},
		[]string{
			urlLabel,
			statusLabel,
			statusCodeLabel,
		},
	)

	// requestDuration PROMQL => rate(request_duration_seconds_sum{}[5m]) / rate(request_duration_seconds_count{}[5m]).
	requestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name: "request_duration_seconds",
			Help: "Response time of HTTP request.",
		},
		[]string{
			urlLabel,
		},
	)
)

func init() {
	prometheus.MustRegister(requestDuration)
	prometheus.MustRegister(requestsTotal)
}

// MetricsMiddleware collect metrics.
func MetricsMiddleware(
	next http.Handler,
	logger logging.Logger,
) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == metricsURLPath {
			next.ServeHTTP(w, r)

			return
		}

		path := r.URL.Path

		if r.URL.Path == graphqlURLPath {
			ctx := r.Context()

			// Reading request body:
			body, err := io.ReadAll(r.Body)
			if err != nil {
				logging.LogErrorContext(
					ctx,
					logger,
					"Failed to collect metrics due to reading request body failure",
					err,
				)

				next.ServeHTTP(w, r)

				return
			}

			// Restoring request body for later usage due to the fact that io.Reader can be read only once:
			r.Body = io.NopCloser(bytes.NewBuffer(body))

			// Parsing request body:
			var requestBody struct {
				Query     string         `json:"query"`
				Variables map[string]any `json:"variables"`
			}

			if err = json.Unmarshal(body, &requestBody); err != nil {
				logging.LogErrorContext(
					ctx,
					logger,
					"Failed to collect metrics due to invalid JSON",
					err,
				)

				next.ServeHTTP(w, r)

				return
			}

			// Retrieving request info:
			info, err := graphqlparser.ParseQuery(requestBody.Query)
			if err != nil {
				logging.LogErrorContext(
					ctx,
					logger,
					"Failed to collect metrics due to GraphQL query parse failure",
					err,
				)

				next.ServeHTTP(w, r)

				return
			}

			path = info.Name
		}

		timer := prometheus.NewTimer(
			requestDuration.With(
				prometheus.Labels{
					urlLabel: path,
				},
			),
		)
		defer timer.ObserveDuration()

		// Create new metricsResponseWriter for response intercepting purpose:
		mrw := newMetricsResponseWriter(w)
		next.ServeHTTP(mrw, r)

		status := statusOK
		if mrw.StatusCode >= http.StatusBadRequest {
			status = statusError
		}

		requestsTotal.With(
			prometheus.Labels{
				urlLabel:        path,
				statusLabel:     status,
				statusCodeLabel: strconv.Itoa(mrw.StatusCode),
			},
		).Inc()
	})
}

func newMetricsResponseWriter(w http.ResponseWriter) *metricsResponseWriter {
	return &metricsResponseWriter{ResponseWriter: w, StatusCode: http.StatusOK}
}

// metricsResponseWriter intercepts response from GraphQL for MetricsMiddleware usage.
type metricsResponseWriter struct {
	http.ResponseWriter
	StatusCode int
	Body       []byte
}

// WriteHeader intercepts response body for later usage in MetricsMiddleware.
func (mrw *metricsResponseWriter) WriteHeader(statusCode int) {
	mrw.StatusCode = statusCode
	mrw.ResponseWriter.WriteHeader(statusCode)
}

// Write intercepts response body for later usage in MetricsMiddleware.
func (mrw *metricsResponseWriter) Write(body []byte) (int, error) {
	mrw.Body = body

	return mrw.ResponseWriter.Write(body)
}
