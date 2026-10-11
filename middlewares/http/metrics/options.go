package metrics

import (
	"net/http"

	"github.com/DKhorkov/libs/middlewares/http/route"
	"github.com/prometheus/client_golang/prometheus"
)

// Option настраивает инструментатор.
type Option func(*options)

// options — собранные настройки инструментатора.
type options struct {
	// routeLabel — метка маршрута запроса.
	routeLabel func(*http.Request) string

	// skip — запросы, которые не меряются вовсе.
	skip func(*http.Request) bool

	// durationBuckets — границы гистограммы длительности, секунды.
	durationBuckets []float64

	// sizeBuckets — границы гистограммы размера ответа, байты.
	sizeBuckets []float64
}

// Границы гистограммы размера по умолчанию: от 256 Б с шагом ×4, восемь
// границ — до 4 МиБ.
const (
	sizeBucketStart  = 256
	sizeBucketFactor = 4
	sizeBucketCount  = 8
)

// defaultOptions — шаблон ServeMux, ничего не пропускается, стандартные
// границы длительности и размеры от 256 Б до 4 МиБ.
func defaultOptions() options {
	return options{
		routeLabel:      route.FromPattern,
		skip:            func(*http.Request) bool { return false },
		durationBuckets: prometheus.DefBuckets,
		sizeBuckets: prometheus.ExponentialBuckets(
			sizeBucketStart,
			sizeBucketFactor,
			sizeBucketCount,
		),
	}
}

// WithRouteLabel задаёт метку маршрута. Функция обязана возвращать шаблон, а
// не путь: каждое уникальное значение — отдельная серия.
func WithRouteLabel(routeLabel func(*http.Request) string) Option {
	return func(o *options) { o.routeLabel = routeLabel }
}

// WithSkip исключает запросы из измерения. Апгрейды соединения исключены
// всегда.
func WithSkip(skip func(*http.Request) bool) Option {
	return func(o *options) { o.skip = skip }
}

// WithDurationBuckets задаёт границы гистограммы длительности.
func WithDurationBuckets(buckets []float64) Option {
	return func(o *options) { o.durationBuckets = buckets }
}

// WithSizeBuckets задаёт границы гистограммы размера ответа.
func WithSizeBuckets(buckets []float64) Option {
	return func(o *options) { o.sizeBuckets = buckets }
}
