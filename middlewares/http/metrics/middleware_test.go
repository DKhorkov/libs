package metrics

import (
	"bufio"
	"context"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newInstrumenter(t *testing.T, opts ...Option) *Instrumenter {
	t.Helper()

	instrumenter, err := New(prometheus.NewRegistry(), opts...)
	require.NoError(t, err)

	return instrumenter
}

func serve(handler http.Handler, method, path string) {
	handler.ServeHTTP(
		httptest.NewRecorder(),
		httptest.NewRequestWithContext(context.Background(), method, path, http.NoBody),
	)
}

func TestRouteLabelFromOption(t *testing.T) {
	t.Parallel()

	instrumenter := newInstrumenter(t, WithRouteLabel(func(*http.Request) string {
		return "/web/battle/{id}"
	}))

	handler := instrumenter.Middleware(
		http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		}),
	)

	serve(handler, http.MethodGet, "/web/battle/4182")
	serve(handler, http.MethodGet, "/web/battle/77")

	assert.InDelta(t, 2, testutil.ToFloat64(
		instrumenter.requests.WithLabelValues("/web/battle/{id}", http.MethodGet, "404"),
	), 0)
	assert.Equal(t, 1, testutil.CollectAndCount(instrumenter.requests))
}

func TestDefaultRouteLabelFromServeMuxPattern(t *testing.T) {
	t.Parallel()

	instrumenter := newInstrumenter(t)

	// Шаблон ставит ServeMux на запросе, который передаёт обработчику, —
	// поэтому инструментатор стоит внутри маршрута, а не снаружи mux.
	mux := http.NewServeMux()
	mux.Handle("GET /items/{id}", instrumenter.Middleware(
		http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
		}),
	))

	serve(mux, http.MethodGet, "/items/42")

	assert.InDelta(t, 1, testutil.ToFloat64(
		instrumenter.requests.WithLabelValues("/items/{id}", http.MethodGet, "200"),
	), 0)
}

func TestNoPatternBecomesOther(t *testing.T) {
	t.Parallel()

	instrumenter := newInstrumenter(t)
	handler := instrumenter.Middleware(
		http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}),
	)

	serve(handler, http.MethodGet, "/raw/path/1")
	serve(handler, http.MethodGet, "/raw/path/2")

	assert.InDelta(t, 2, testutil.ToFloat64(
		instrumenter.requests.WithLabelValues("other", http.MethodGet, "200"),
	), 0)
}

func TestUnknownMethodBecomesOTHER(t *testing.T) {
	t.Parallel()

	instrumenter := newInstrumenter(t, WithRouteLabel(func(*http.Request) string { return "/x" }))
	handler := instrumenter.Middleware(
		http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}),
	)

	serve(handler, "PROPFIND", "/x")
	serve(handler, "BREW", "/x")

	assert.InDelta(t, 2, testutil.ToFloat64(
		instrumenter.requests.WithLabelValues("/x", "OTHER", "200"),
	), 0)
}

func TestMiddlewareForUsesFixedRoute(t *testing.T) {
	t.Parallel()

	instrumenter := newInstrumenter(t)
	handler := instrumenter.MiddlewareFor("unmatched")(http.NotFoundHandler())

	for _, path := range []string{"/a", "/b", "/wp-login.php"} {
		serve(handler, http.MethodGet, path)
	}

	assert.InDelta(t, 3, testutil.ToFloat64(
		instrumenter.requests.WithLabelValues("unmatched", http.MethodGet, "404"),
	), 0)
	assert.Equal(t, 1, testutil.CollectAndCount(instrumenter.requests))
}

func TestUpgradeRequestsAreSkipped(t *testing.T) {
	t.Parallel()

	instrumenter := newInstrumenter(t)
	handler := instrumenter.Middleware(
		http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusSwitchingProtocols)
		}),
	)

	request := httptest.NewRequestWithContext(
		context.Background(),
		http.MethodGet,
		"/api/ws",
		http.NoBody,
	)
	request.Header.Set("Connection", "keep-alive, Upgrade")
	request.Header.Set("Upgrade", "websocket")
	handler.ServeHTTP(httptest.NewRecorder(), request)

	assert.Equal(t, 0, testutil.CollectAndCount(instrumenter.requests))
	assert.Equal(t, 0, testutil.CollectAndCount(instrumenter.duration))
	assert.InDelta(t, 0, testutil.ToFloat64(instrumenter.inFlight), 0)
}

func TestWithSkip(t *testing.T) {
	t.Parallel()

	instrumenter := newInstrumenter(t, WithSkip(func(r *http.Request) bool {
		return r.URL.Path == "/healthcheck"
	}))
	handler := instrumenter.Middleware(
		http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}),
	)

	serve(handler, http.MethodGet, "/healthcheck")

	assert.Equal(t, 0, testutil.CollectAndCount(instrumenter.requests))
}

func TestInFlightReturnsToZero(t *testing.T) {
	t.Parallel()

	instrumenter := newInstrumenter(t)

	var during float64

	handler := instrumenter.Middleware(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		during = testutil.ToFloat64(instrumenter.inFlight)
	}))

	serve(handler, http.MethodGet, "/")

	assert.InDelta(t, 1, during, 0)
	assert.InDelta(t, 0, testutil.ToFloat64(instrumenter.inFlight), 0)
}

func TestResponseSizeAndDurationObserved(t *testing.T) {
	t.Parallel()

	instrumenter := newInstrumenter(t, WithRouteLabel(func(*http.Request) string { return "/x" }))
	handler := instrumenter.Middleware(
		http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			_, _ = w.Write([]byte("abc"))
			_, _ = w.Write([]byte("de"))
		}),
	)

	serve(handler, http.MethodPost, "/x")

	expected := `
# HELP http_response_size_bytes Size of HTTP responses in bytes by route.
# TYPE http_response_size_bytes histogram
http_response_size_bytes_bucket{method="POST",route="/x",le="256"} 1
http_response_size_bytes_bucket{method="POST",route="/x",le="1024"} 1
http_response_size_bytes_bucket{method="POST",route="/x",le="4096"} 1
http_response_size_bytes_bucket{method="POST",route="/x",le="16384"} 1
http_response_size_bytes_bucket{method="POST",route="/x",le="65536"} 1
http_response_size_bytes_bucket{method="POST",route="/x",le="262144"} 1
http_response_size_bytes_bucket{method="POST",route="/x",le="1.048576e+06"} 1
http_response_size_bytes_bucket{method="POST",route="/x",le="4.194304e+06"} 1
http_response_size_bytes_bucket{method="POST",route="/x",le="+Inf"} 1
http_response_size_bytes_sum{method="POST",route="/x"} 5
http_response_size_bytes_count{method="POST",route="/x"} 1
`
	require.NoError(t, testutil.CollectAndCompare(
		instrumenter.size, stringsReader(expected), "http_response_size_bytes",
	))
	assert.Equal(t, 1, testutil.CollectAndCount(instrumenter.duration))
}

func TestDuplicateRegistrationFails(t *testing.T) {
	t.Parallel()

	registry := prometheus.NewRegistry()

	_, err := New(registry)
	require.NoError(t, err)

	_, err = New(registry)
	require.Error(t, err)
}

// hijackable — ResponseWriter, умеющий Hijack, как у net/http.
type hijackable struct {
	*httptest.ResponseRecorder

	// hijacked — Hijack был вызван.
	hijacked bool
}

func (h *hijackable) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	h.hijacked = true

	return nil, nil, nil
}

func TestRecorderPassesHijackThrough(t *testing.T) {
	t.Parallel()

	inner := &hijackable{ResponseRecorder: httptest.NewRecorder()}
	recorder := &responseRecorder{ResponseWriter: inner, status: http.StatusOK}

	_, _, err := recorder.Hijack()
	require.NoError(t, err)
	assert.True(t, inner.hijacked)

	plain := &responseRecorder{ResponseWriter: httptest.NewRecorder(), status: http.StatusOK}
	_, _, err = plain.Hijack()
	require.ErrorIs(t, err, ErrHijackUnsupported)
}

func stringsReader(s string) *strings.Reader { return strings.NewReader(s) }

// TestInformationalStatusIsNotFinal — 1xx (103 Early Hints) не окончательный
// ответ: обработчик пишет его до настоящего кода, и метка обязана нести
// итоговый 200, а не 103.
func TestInformationalStatusIsNotFinal(t *testing.T) {
	t.Parallel()

	instrumenter := newInstrumenter(t, WithRouteLabel(func(*http.Request) string { return "/x" }))
	handler := instrumenter.Middleware(
		http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Link", "</style.css>; rel=preload; as=style")
			w.WriteHeader(http.StatusEarlyHints)
			w.WriteHeader(http.StatusOK)
		}),
	)

	serve(handler, http.MethodGet, "/x")

	assert.InDelta(t, 1, testutil.ToFloat64(
		instrumenter.requests.WithLabelValues("/x", http.MethodGet, "200"),
	), 0)
	assert.Equal(t, 1, testutil.CollectAndCount(instrumenter.requests))
}

// TestPanicIsCountedAsServerError — паника обработчика — самая важная
// пятисотка, и метрика обязана её увидеть: net/http перехватит панику и
// закроет соединение, а без учёта запрос пропал бы из rps и из доли 5xx.
// Паника при этом не глотается — её ловит тот, кто стоит выше.
func TestPanicIsCountedAsServerError(t *testing.T) {
	t.Parallel()

	instrumenter := newInstrumenter(t, WithRouteLabel(func(*http.Request) string { return "/x" }))
	handler := instrumenter.Middleware(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		panic("boom")
	}))

	assert.PanicsWithValue(t, "boom", func() { serve(handler, http.MethodGet, "/x") })

	assert.InDelta(t, 1, testutil.ToFloat64(
		instrumenter.requests.WithLabelValues("/x", http.MethodGet, "500"),
	), 0)
	assert.InDelta(t, 0, testutil.ToFloat64(instrumenter.inFlight), 0)
}

// TestAbortHandlerIsNotCounted — http.ErrAbortHandler — штатный способ
// оборвать ответ (клиент ушёл, прокси прервал поток), а не сбой: в 5xx он не
// попадает.
func TestAbortHandlerIsNotCounted(t *testing.T) {
	t.Parallel()

	instrumenter := newInstrumenter(t)
	handler := instrumenter.Middleware(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		panic(http.ErrAbortHandler)
	}))

	assert.Panics(t, func() { serve(handler, http.MethodGet, "/x") })
	assert.Equal(t, 0, testutil.CollectAndCount(instrumenter.requests))
}
