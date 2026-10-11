package metrics

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/DKhorkov/libs/middlewares/http/route"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	routeLabel  = "route"
	methodLabel = "method"
	codeLabel   = "code"
)

// Instrumenter меряет HTTP-запросы в переданный реестр.
type Instrumenter struct {
	// requests — число запросов по маршруту, методу и коду.
	requests *prometheus.CounterVec

	// duration — длительность обработки.
	duration *prometheus.HistogramVec

	// size — размер тела ответа.
	size *prometheus.HistogramVec

	// inFlight — запросы в обработке прямо сейчас.
	inFlight prometheus.Gauge

	// options — настройки из New.
	options options
}

// New регистрирует метрики HTTP в registerer.
//
// Глобального реестра и init() нет: тест получает свой реестр, а повторная
// регистрация в том же — ошибка, а не panic.
func New(registerer prometheus.Registerer, opts ...Option) (*Instrumenter, error) {
	o := defaultOptions()
	for _, opt := range opts {
		opt(&o)
	}

	instrumenter := &Instrumenter{
		requests: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Number of HTTP requests by route, method and status code.",
		}, []string{routeLabel, methodLabel, codeLabel}),
		duration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "Duration of HTTP requests by route, method and status code.",
			Buckets: o.durationBuckets,
		}, []string{routeLabel, methodLabel, codeLabel}),
		size: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "http_response_size_bytes",
			Help:    "Size of HTTP responses in bytes by route.",
			Buckets: o.sizeBuckets,
		}, []string{routeLabel, methodLabel}),
		inFlight: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "http_requests_in_flight",
			Help: "Number of HTTP requests being served right now.",
		}),
		options: o,
	}

	for _, collector := range []prometheus.Collector{
		instrumenter.requests,
		instrumenter.duration,
		instrumenter.size,
		instrumenter.inFlight,
	} {
		if err := registerer.Register(collector); err != nil {
			return nil, err
		}
	}

	return instrumenter, nil
}

// Middleware меряет запрос с меткой маршрута из WithRouteLabel.
func (i *Instrumenter) Middleware(next http.Handler) http.Handler {
	return i.instrument(next, i.options.routeLabel)
}

// MiddlewareFor меряет запрос с фиксированной меткой маршрута — для
// обработчиков без шаблона: 404, 405.
//
// Параметр назван fixedRoute, а не route: имя route занято импортом пакета
// разбора маршрута.
func (i *Instrumenter) MiddlewareFor(fixedRoute string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return i.instrument(next, func(*http.Request) string { return fixedRoute })
	}
}

func (i *Instrumenter) instrument(
	next http.Handler,
	routeOf func(*http.Request) string,
) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if isUpgrade(r) || i.options.skip(r) {
			next.ServeHTTP(w, r)

			return
		}

		i.inFlight.Inc()
		defer i.inFlight.Dec()

		started := time.Now()
		recorder := &responseRecorder{ResponseWriter: w, status: http.StatusOK}

		defer func() {
			// Паника — самая важная пятисотка: net/http перехватит её и
			// закроет соединение, а без учёта запрос пропал бы из rps и доли
			// 5xx. http.ErrAbortHandler — штатный обрыв ответа, не сбой.
			// Паника не глотается: её ловит тот, кто стоит выше.
			if recovered := recover(); recovered != nil {
				if err, isErr := recovered.(error); !isErr ||
					!errors.Is(err, http.ErrAbortHandler) {
					i.observe(
						routeOf(r),
						r,
						http.StatusInternalServerError,
						recorder.written,
						started,
					)
				}

				panic(recovered)
			}
		}()

		next.ServeHTTP(recorder, r)

		i.observe(routeOf(r), r, recorder.status, recorder.written, started)
	})
}

// observe записывает один обработанный запрос.
func (i *Instrumenter) observe(
	routeName string,
	r *http.Request,
	status, written int,
	started time.Time,
) {
	method := route.Method(r)
	code := strconv.Itoa(status)

	i.requests.WithLabelValues(routeName, method, code).Inc()
	i.duration.WithLabelValues(routeName, method, code).Observe(time.Since(started).Seconds())
	i.size.WithLabelValues(routeName, method).Observe(float64(written))
}

// isUpgrade — запрос на смену протокола (WebSocket и подобные).
//
// Его «длительность» — жизнь соединения, а не обработка запроса: часовой
// сокет лёг бы в +Inf и сделал бы среднее время ответа ложным.
func isUpgrade(r *http.Request) bool {
	if r.Header.Get("Upgrade") == "" {
		return false
	}

	for _, value := range r.Header.Values("Connection") {
		for token := range strings.SplitSeq(value, ",") {
			if strings.EqualFold(strings.TrimSpace(token), "upgrade") {
				return true
			}
		}
	}

	return false
}
