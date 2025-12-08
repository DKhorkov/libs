package http

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/mock/gomock"
	"google.golang.org/grpc/metadata"

	mocklogging "github.com/DKhorkov/libs/logging/mocks"
	"github.com/DKhorkov/libs/tracing"
	mocktracing "github.com/DKhorkov/libs/tracing/mocks"
)

func TestTracingMiddleware(t *testing.T) {
	t.Run("Creates span and adds traceID to metadata", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		provider := mocktracing.NewMockProvider(ctrl)
		span := mocktracing.NewMockSpan()
		logger := mocklogging.NewMockLogger(ctrl)

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
			Span(ctx, spanConfig.Name, gomock.Any()).
			Return(ctx, span).
			Times(1)

		logger.
			EXPECT().
			InfoContext(gomock.Any(), gomock.Any(), gomock.Any()).
			Times(1)

		nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Проверяем метаданные
			md, ok := metadata.FromOutgoingContext(r.Context())
			require.True(t, ok)
			require.Contains(t, md, tracing.Key)
			require.NotZero(t, md.Get(tracing.Key))
			w.WriteHeader(http.StatusOK)
		})

		middleware := TracingMiddleware(nextHandler, logger, provider, spanConfig)

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rr := httptest.NewRecorder()

		middleware.ServeHTTP(rr, req)

		require.Equal(t, http.StatusOK, rr.Code)
	})

	t.Run("Ignores metrics url", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		provider := mocktracing.NewMockProvider(ctrl)
		logger := mocklogging.NewMockLogger(ctrl)
		spanConfig := tracing.SpanConfig{}

		nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		middleware := TracingMiddleware(nextHandler, logger, provider, spanConfig)

		req := httptest.NewRequest(http.MethodGet, metricsURLPath, nil)
		rr := httptest.NewRecorder()

		middleware.ServeHTTP(rr, req)

		require.Equal(t, http.StatusOK, rr.Code)
	})

	t.Run("Handles response with errors", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		provider := mocktracing.NewMockProvider(ctrl)
		span := mocktracing.NewMockSpan()
		logger := mocklogging.NewMockLogger(ctrl)

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
			Span(ctx, spanConfig.Name, gomock.Any()).
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

		middleware := TracingMiddleware(nextHandler, logger, provider, spanConfig)

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rr := httptest.NewRecorder()

		middleware.ServeHTTP(rr, req)

		require.Equal(t, http.StatusOK, rr.Code)
		require.JSONEq(t, string(responseBody), rr.Body.String())
	})

	t.Run("Handles invalid JSON response", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		provider := mocktracing.NewMockProvider(ctrl)
		span := mocktracing.NewMockSpan()
		logger := mocklogging.NewMockLogger(ctrl)

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
			Span(ctx, spanConfig.Name, gomock.Any()).
			Return(ctx, span).
			Times(1)

		logger.
			EXPECT().
			InfoContext(gomock.Any(), gomock.Any(), gomock.Any()).
			Times(1)

		nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, err := w.Write([]byte("invalid json"))
			require.NoError(t, err)
		})

		middleware := TracingMiddleware(nextHandler, logger, provider, spanConfig)

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rr := httptest.NewRecorder()

		middleware.ServeHTTP(rr, req)

		require.Equal(t, http.StatusOK, rr.Code)
	})

	t.Run("Handles response without errors", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		provider := mocktracing.NewMockProvider(ctrl)
		span := mocktracing.NewMockSpan()
		logger := mocklogging.NewMockLogger(ctrl)

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
			Span(ctx, spanConfig.Name, gomock.Any()).
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

		middleware := TracingMiddleware(nextHandler, logger, provider, spanConfig)

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rr := httptest.NewRecorder()

		middleware.ServeHTTP(rr, req)

		require.Equal(t, http.StatusOK, rr.Code)
		require.JSONEq(t, string(responseBody), rr.Body.String())
	})
}

func TestTracingResponseWriter(t *testing.T) {
	t.Run("WriteHeader captures status code", func(t *testing.T) {
		rr := httptest.NewRecorder()
		trw := &tracingResponseWriter{ResponseWriter: rr}

		trw.WriteHeader(http.StatusCreated)
		require.Equal(t, http.StatusCreated, trw.StatusCode)
		require.Equal(t, http.StatusCreated, rr.Code)
	})

	t.Run("Write captures body", func(t *testing.T) {
		rr := httptest.NewRecorder()
		trw := &tracingResponseWriter{ResponseWriter: rr}

		body := []byte(`{"data":"test"}`)
		n, err := trw.Write(body)
		require.NoError(t, err)
		require.Equal(t, len(body), n)
		require.Equal(t, body, trw.Body)
		require.Equal(t, string(body), rr.Body.String())
	})
}
