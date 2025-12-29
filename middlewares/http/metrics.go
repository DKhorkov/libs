package http

import (
	"net/http"
	"runtime"
	"runtime/metrics"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

const (
	MetricsURLPath = "/metrics"

	urlLabel        = "url"
	methodLabel     = "method"
	statusLabel     = "status"
	statusCodeLabel = "status_code"

	statusOK    = "ok"
	statusError = "error"
)

var (
	// requestsTotal PROMQL => rate(requests_total[30s]).
	requestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "requests_total",
			Help: "Number of HTTP requests by url.",
		},
		[]string{
			urlLabel,
			methodLabel,
			statusLabel,
			statusCodeLabel,
		},
	)

	// requestDuration PROMQL => rate(request_duration_seconds_sum[30s]) / rate(request_duration_seconds_count[30s]).
	requestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name: "request_duration_seconds",
			Help: "Response time of HTTP request.",
		},
		[]string{
			urlLabel,
			methodLabel,
			statusLabel,
			statusCodeLabel,
		},
	)

	goroutinesNumber = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "goroutines_number",
			Help: "Number of goroutines that currently exist.",
		},
	)

	memoryUsage = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "memory_usage",
			Help: "Current memory usage.",
		},
	)

	metricsToCollect = map[string]prometheus.Metric{
		goroutinesNumberMetricName: goroutinesNumber,
		memoryUsageMetricName:      memoryUsage,
	}
)

const (
	goroutinesNumberMetricName = "/sched/goroutines:goroutines"
	memoryUsageMetricName      = "/memory/classes/heap/free:bytes"
)

func init() {
	prometheus.MustRegister(requestDuration)
	prometheus.MustRegister(requestsTotal)
	prometheus.MustRegister(goroutinesNumber)
	prometheus.MustRegister(memoryUsage)
}

// MetricsMiddleware collect metrics.
func MetricsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == MetricsURLPath {
			next.ServeHTTP(w, r)

			return
		}

		collectGoMetrics()

		now := time.Now()

		// Create new metricsResponseWriter for response intercepting purpose:
		mrw := newInterceptingResponseWriter(w)
		next.ServeHTTP(mrw, r)

		status := statusOK
		if mrw.StatusCode >= http.StatusBadRequest {
			status = statusError
		}

		requestsTotal.With(
			prometheus.Labels{
				urlLabel:        r.URL.Path,
				methodLabel:     r.Method,
				statusLabel:     status,
				statusCodeLabel: strconv.Itoa(mrw.StatusCode),
			},
		).Inc()

		requestDuration.With(
			prometheus.Labels{
				urlLabel:        r.URL.Path,
				methodLabel:     r.Method,
				statusLabel:     status,
				statusCodeLabel: strconv.Itoa(mrw.StatusCode),
			},
		).Observe(time.Since(now).Seconds())
	})
}

func collectGoMetrics() {
	runtime.GC()

	metricsSample := make([]metrics.Sample, 0, len(metricsToCollect))
	for metric := range metricsToCollect {
		metricsSample = append(metricsSample, metrics.Sample{Name: metric})
	}

	metrics.Read(metricsSample)

	for _, m := range metricsSample {
		switch m.Name {
		case goroutinesNumberMetricName:
			goroutinesNumber.Set(float64(m.Value.Uint64()))
		case memoryUsageMetricName:
			memoryUsage.Set(float64(m.Value.Uint64()))
		}
	}
}
