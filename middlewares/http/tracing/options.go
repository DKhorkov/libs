package tracing

import (
	"net/http"

	"github.com/DKhorkov/libs/middlewares/http/route"
)

// Option настраивает трассировку запроса.
type Option func(*options)

// options — собранные настройки.
type options struct {
	// route — шаблон маршрута: из него имя спана и атрибут http.route.
	route func(*http.Request) string

	// skip — запросы без спана.
	skip func(*http.Request) bool
}

func defaultOptions() options {
	return options{
		route: route.FromPattern,
		skip:  func(*http.Request) bool { return false },
	}
}

// WithRoute задаёт шаблон маршрута. Функция обязана возвращать шаблон, а не
// путь: имя спана — операция в Jaeger и измерение span_metrics.
func WithRoute(routeOf func(*http.Request) string) Option {
	return func(o *options) { o.route = routeOf }
}

// WithSkip исключает запросы из трассировки.
func WithSkip(skip func(*http.Request) bool) Option {
	return func(o *options) { o.skip = skip }
}
