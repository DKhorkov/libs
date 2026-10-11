package tracing

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DKhorkov/libs/tracing"
	mocktracing "github.com/DKhorkov/libs/tracing/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/mock/gomock"
	"google.golang.org/grpc/metadata"
)

func TestMiddleware(t *testing.T) {
	t.Parallel()

	t.Run("Creates span and adds traceID to metadata", func(t *testing.T) {
		t.Parallel()

		ctrl := gomock.NewController(t)
		provider := mocktracing.NewMockProvider(ctrl)
		span := mocktracing.NewMockSpan()

		spanConfig := tracing.SpanConfig{
			Name: "test-span",
			Events: tracing.SpanEventsConfig{
				Start: tracing.SpanEventConfig{Name: "start"},
				End:   tracing.SpanEventConfig{Name: "end"},
			},
			Opts: []trace.SpanStartOption{},
		}
		ctx := context.Background()

		provider.
			EXPECT().
			Span(ctx, "GET", gomock.Any()).
			Return(ctx, span).
			Times(1)

		nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Проверяем метаданные
			md, ok := metadata.FromOutgoingContext(r.Context())
			require.True(t, ok)
			require.Contains(t, md, tracing.Key)
			require.NotZero(t, md.Get(tracing.Key))
			w.WriteHeader(http.StatusOK)
		})

		middleware := Middleware(provider, spanConfig)(nextHandler)

		req := httptest.NewRequestWithContext(
			context.Background(),
			http.MethodGet,
			"/",
			http.NoBody,
		)
		rr := httptest.NewRecorder()

		middleware.ServeHTTP(rr, req)

		require.Equal(t, http.StatusOK, rr.Code)
	})

	t.Run("Handles response with errors", func(t *testing.T) {
		t.Parallel()

		ctrl := gomock.NewController(t)
		provider := mocktracing.NewMockProvider(ctrl)
		span := mocktracing.NewMockSpan()

		spanConfig := tracing.SpanConfig{
			Name: "test-span",
			Events: tracing.SpanEventsConfig{
				Start: tracing.SpanEventConfig{Name: "start"},
				End:   tracing.SpanEventConfig{Name: "end"},
			},
		}
		ctx := context.Background()

		provider.
			EXPECT().
			Span(ctx, "GET", gomock.Any()).
			Return(ctx, span).
			Times(1)

		response := map[string]any{
			"errors": []any{
				map[string]any{"message": "error1", "path": "/path1"},
				map[string]any{"message": "error2", "path": "/path2"},
			},
		}
		responseBody, _ := json.Marshal(response)

		nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, err := w.Write(responseBody)
			require.NoError(t, err)
		})

		middleware := Middleware(provider, spanConfig)(nextHandler)

		req := httptest.NewRequestWithContext(
			context.Background(),
			http.MethodGet,
			"/",
			http.NoBody,
		)
		rr := httptest.NewRecorder()

		middleware.ServeHTTP(rr, req)

		require.Equal(t, http.StatusOK, rr.Code)
		require.JSONEq(t, string(responseBody), rr.Body.String())
	})

	t.Run("Handles invalid JSON response", func(t *testing.T) {
		t.Parallel()

		ctrl := gomock.NewController(t)
		provider := mocktracing.NewMockProvider(ctrl)
		span := mocktracing.NewMockSpan()

		spanConfig := tracing.SpanConfig{
			Name: "test-span",
			Events: tracing.SpanEventsConfig{
				Start: tracing.SpanEventConfig{Name: "start"},
				End:   tracing.SpanEventConfig{Name: "end"},
			},
		}
		ctx := context.Background()

		provider.
			EXPECT().
			Span(ctx, "GET", gomock.Any()).
			Return(ctx, span).
			Times(1)

		nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, err := w.Write([]byte("invalid json"))
			require.NoError(t, err)
		})

		middleware := Middleware(provider, spanConfig)(nextHandler)

		req := httptest.NewRequestWithContext(
			context.Background(),
			http.MethodGet,
			"/",
			http.NoBody,
		)
		rr := httptest.NewRecorder()

		middleware.ServeHTTP(rr, req)

		require.Equal(t, http.StatusOK, rr.Code)
	})

	t.Run("Handles response without errors", func(t *testing.T) {
		t.Parallel()

		ctrl := gomock.NewController(t)
		provider := mocktracing.NewMockProvider(ctrl)
		span := mocktracing.NewMockSpan()

		spanConfig := tracing.SpanConfig{
			Name: "test-span",
			Events: tracing.SpanEventsConfig{
				Start: tracing.SpanEventConfig{Name: "start"},
				End:   tracing.SpanEventConfig{Name: "end"},
			},
		}
		ctx := context.Background()

		provider.
			EXPECT().
			Span(ctx, "GET", gomock.Any()).
			Return(ctx, span).
			Times(1)

		response := map[string]any{
			"data": "success",
		}
		responseBody, _ := json.Marshal(response)

		nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, err := w.Write(responseBody)
			require.NoError(t, err)
		})

		middleware := Middleware(provider, spanConfig)(nextHandler)

		req := httptest.NewRequestWithContext(
			context.Background(),
			http.MethodGet,
			"/",
			http.NoBody,
		)
		rr := httptest.NewRecorder()

		middleware.ServeHTTP(rr, req)

		require.Equal(t, http.StatusOK, rr.Code)
		require.JSONEq(t, string(responseBody), rr.Body.String())
	})
}

func recordingProvider(
	t *testing.T,
	ctrl *gomock.Controller,
) (*mocktracing.MockProvider, *tracetest.SpanRecorder) {
	t.Helper()

	recorder := tracetest.NewSpanRecorder()
	tracer := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(recorder)).Tracer("test")

	provider := mocktracing.NewMockProvider(ctrl)
	provider.EXPECT().
		Span(gomock.Any(), gomock.Any(), gomock.Any()).
		DoAndReturn(func(
			ctx context.Context,
			name string,
			opts ...trace.SpanStartOption,
		) (context.Context, trace.Span) {
			return tracer.Start(ctx, name, opts...) //nolint:spancheck // спан завершает middleware
		}).
		AnyTimes()

	return provider, recorder
}

func TestSpanNamedByRoute(t *testing.T) {
	t.Parallel()

	provider, recorder := recordingProvider(t, gomock.NewController(t))

	handler := Middleware(
		provider,
		tracing.SpanConfig{},
		WithRoute(func(*http.Request) string { return "/web/battle/{id}" }),
	)(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusTeapot)
	}))

	handler.ServeHTTP(
		httptest.NewRecorder(),
		httptest.NewRequestWithContext(
			context.Background(),
			http.MethodGet,
			"/web/battle/4182",
			http.NoBody,
		),
	)

	spans := recorder.Ended()
	require.Len(t, spans, 1)
	assert.Equal(t, "GET /web/battle/{id}", spans[0].Name())

	attributes := map[attribute.Key]attribute.Value{}
	for _, kv := range spans[0].Attributes() {
		attributes[kv.Key] = kv.Value
	}

	assert.Equal(t, "GET", attributes["http.request.method"].AsString())
	assert.Equal(t, "/web/battle/{id}", attributes["http.route"].AsString())
	assert.Equal(t, int64(http.StatusTeapot), attributes["http.response.status_code"].AsInt64())
}

func TestDefaultRouteFromServeMuxPattern(t *testing.T) {
	t.Parallel()

	provider, recorder := recordingProvider(t, gomock.NewController(t))

	inner := http.HandlerFunc(func(http.ResponseWriter, *http.Request) {})
	mux := http.NewServeMux()
	mux.Handle("GET /items/{id}", Middleware(provider, tracing.SpanConfig{})(inner))

	mux.ServeHTTP(
		httptest.NewRecorder(),
		httptest.NewRequestWithContext(
			context.Background(),
			http.MethodGet,
			"/items/7",
			http.NoBody,
		),
	)

	require.Len(t, recorder.Ended(), 1)
	assert.Equal(t, "GET /items/{id}", recorder.Ended()[0].Name())
}

func TestWithSkipCreatesNoSpan(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	provider := mocktracing.NewMockProvider(ctrl) // ни одного ожидания: Span не зовётся

	handler := Middleware(
		provider,
		tracing.SpanConfig{},
		WithSkip(func(r *http.Request) bool { return r.URL.Path == "/healthcheck" }),
	)(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) }))

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(
		recorder,
		httptest.NewRequestWithContext(
			context.Background(),
			http.MethodGet,
			"/healthcheck",
			http.NoBody,
		),
	)

	assert.Equal(t, http.StatusOK, recorder.Code)
}

// TestUnknownMethodSpanName — метод вне стандартного набора в имени спана —
// OTHER: имя спана — операция Jaeger и измерение span_metrics, и сканер,
// перебирающий методы, иначе плодил бы их без предела.
func TestUnknownMethodSpanName(t *testing.T) {
	t.Parallel()

	provider, recorder := recordingProvider(t, gomock.NewController(t))

	handler := Middleware(
		provider,
		tracing.SpanConfig{},
		WithRoute(func(*http.Request) string { return "method_not_allowed" }),
	)(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusMethodNotAllowed)
	}))

	for _, method := range []string{"PROPFIND", "BREW"} {
		handler.ServeHTTP(
			httptest.NewRecorder(),
			httptest.NewRequestWithContext(context.Background(), method, "/x", http.NoBody),
		)
	}

	spans := recorder.Ended()
	require.Len(t, spans, 2)

	for _, span := range spans {
		assert.Equal(t, "OTHER method_not_allowed", span.Name())
	}
}

// TestPanicMarksRootSpanAsError — паника обработчика — ошибка корневого спана:
// иначе спан закрывался бы со статусом Unset, и доля ошибок span_metrics её
// не видела. Паника не глотается.
func TestPanicMarksRootSpanAsError(t *testing.T) {
	t.Parallel()

	provider, recorder := recordingProvider(t, gomock.NewController(t))

	handler := Middleware(
		provider,
		tracing.SpanConfig{},
		WithRoute(func(*http.Request) string { return "/x" }),
	)(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		panic("boom")
	}))

	assert.Panics(t, func() {
		handler.ServeHTTP(
			httptest.NewRecorder(),
			httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/x", http.NoBody),
		)
	})

	spans := recorder.Ended()
	require.Len(t, spans, 1)
	assert.Equal(t, codes.Error, spans[0].Status().Code)
}
