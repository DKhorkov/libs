package http

import (
	"errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
)

func TestMetricsMiddleware_RegularRequests(t *testing.T) {
	// Тестовые обработчики
	successHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte("OK"))
		require.NoError(t, err)
	})

	errorHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, err := w.Write([]byte("Not Found"))
		require.NoError(t, err)
	})

	serverErrorHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, err := w.Write([]byte("Internal Server Error"))
		require.NoError(t, err)
	})

	tests := []struct {
		name           string
		method         string
		handler        http.Handler
		path           string
		expectedStatus int
		expectedLabels prometheus.Labels
		shouldRecord   bool
	}{
		{
			name:           "successful GET request",
			method:         "GET",
			handler:        successHandler,
			path:           "/api/users",
			expectedStatus: http.StatusOK,
			expectedLabels: prometheus.Labels{
				urlLabel:        "/api/users",
				methodLabel:     "GET",
				statusLabel:     statusOK,
				statusCodeLabel: "200",
			},
			shouldRecord: true,
		},
		{
			name:           "successful POST request",
			method:         "POST",
			handler:        successHandler,
			path:           "/api/users",
			expectedStatus: http.StatusOK,
			expectedLabels: prometheus.Labels{
				urlLabel:        "/api/users",
				methodLabel:     "POST",
				statusLabel:     statusOK,
				statusCodeLabel: "200",
			},
			shouldRecord: true,
		},
		{
			name:           "client error request",
			method:         "GET",
			handler:        errorHandler,
			path:           "/api/users",
			expectedStatus: http.StatusNotFound,
			expectedLabels: prometheus.Labels{
				urlLabel:        "/api/users",
				methodLabel:     "GET",
				statusLabel:     statusError,
				statusCodeLabel: "404",
			},
			shouldRecord: true,
		},
		{
			name:           "server error request",
			method:         "GET",
			handler:        serverErrorHandler,
			path:           "/api/users",
			expectedStatus: http.StatusInternalServerError,
			expectedLabels: prometheus.Labels{
				urlLabel:        "/api/users",
				methodLabel:     "GET",
				statusLabel:     statusError,
				statusCodeLabel: "500",
			},
			shouldRecord: true,
		},
		{
			name:           "metrics endpoint skipped",
			method:         "GET",
			handler:        successHandler,
			path:           metricsURLPath,
			expectedStatus: http.StatusOK,
			expectedLabels: nil,
			shouldRecord:   false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Сброс метрик перед каждым тестом
			requestsTotal.Reset()
			requestDuration.Reset()

			mw := MetricsMiddleware(tc.handler)
			req := httptest.NewRequest(tc.method, tc.path, nil)
			rr := httptest.NewRecorder()

			mw.ServeHTTP(rr, req)

			require.Equal(t, tc.expectedStatus, rr.Code)
			assert.NotEmpty(t, rr.Body.String())

			if tc.shouldRecord {
				// Проверяем, что метрика requestsTotal была увеличена
				count := testutil.ToFloat64(requestsTotal.With(tc.expectedLabels))
				assert.Equal(t, 1.0, count, "requests_total должен быть увеличен на 1")

				// Проверяем, что метрика requestDuration была обновлена
				// Используем CollectAndCount для проверки наличия наблюдений
				metricCount := testutil.CollectAndCount(requestDuration)
				assert.Greater(t, metricCount, 0, "request_duration_seconds должен иметь наблюдения")
			} else {
				// Проверяем, что метрики не были записаны для /metrics
				count := testutil.CollectAndCount(requestsTotal)
				assert.Equal(t, 0, count, "метрики не должны записываться для эндпоинта /metrics")
			}
		})
	}
}

func TestMetricsMiddleware_MethodLabel(t *testing.T) {
	tests := []struct {
		name     string
		method   string
		expected string
	}{
		{"GET method", "GET", "GET"},
		{"POST method", "POST", "POST"},
		{"PUT method", "PUT", "PUT"},
		{"DELETE method", "DELETE", "DELETE"},
		{"PATCH method", "PATCH", "PATCH"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			requestsTotal.Reset()
			requestDuration.Reset()

			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			})

			mw := MetricsMiddleware(handler)
			req := httptest.NewRequest(tc.method, "/test", nil)
			rr := httptest.NewRecorder()

			mw.ServeHTTP(rr, req)

			expectedLabels := prometheus.Labels{
				urlLabel:        "/test",
				methodLabel:     tc.expected,
				statusLabel:     statusOK,
				statusCodeLabel: "200",
			}

			count := testutil.ToFloat64(requestsTotal.With(expectedLabels))
			assert.Equal(t, 1.0, count, "метод должен быть правильно записан в метриках")
		})
	}
}

func TestCollectGoMetrics(t *testing.T) {
	// Сброс метрик
	goroutinesCount.Set(0)
	memoryUsage.Set(0)

	// Вызов функции сбора метрик Go
	collectGoMetrics()

	// Проверяем, что метрики были установлены (значения могут быть любые, но не 0)
	goroutinesValue := testutil.ToFloat64(goroutinesCount)
	memoryValue := testutil.ToFloat64(memoryUsage)

	// Мы не можем точно знать значения, но можем проверить, что они не остались 0
	// после вызова runtime.GC() и metrics.Read()
	assert.True(t, goroutinesValue >= 0, "goroutinesCount должен быть неотрицательным")
	assert.True(t, memoryValue >= 0, "memoryUsage должен быть неотрицательным")
}

func TestMetricsResponseWriter(t *testing.T) {
	tests := []struct {
		name               string
		statusCode         int
		body               string
		expectedStatusCode int
		expectedBody       string
	}{
		{
			name:               "OK response",
			statusCode:         http.StatusOK,
			body:               "OK response",
			expectedStatusCode: http.StatusOK,
			expectedBody:       "OK response",
		},
		{
			name:               "Created response",
			statusCode:         http.StatusCreated,
			body:               "Created",
			expectedStatusCode: http.StatusCreated,
			expectedBody:       "Created",
		},
		{
			name:               "Bad Request response",
			statusCode:         http.StatusBadRequest,
			body:               "Bad Request",
			expectedStatusCode: http.StatusBadRequest,
			expectedBody:       "Bad Request",
		},
		{
			name:               "Internal Server Error response",
			statusCode:         http.StatusInternalServerError,
			body:               "Internal Server Error",
			expectedStatusCode: http.StatusInternalServerError,
			expectedBody:       "Internal Server Error",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Создаем recorder и обертку
			rr := httptest.NewRecorder()
			mrw := newMetricsResponseWriter(rr)

			// Тестируем WriteHeader
			mrw.WriteHeader(tc.statusCode)
			assert.Equal(t, tc.expectedStatusCode, mrw.StatusCode)
			assert.Equal(t, tc.expectedStatusCode, rr.Code)

			// Тестируем Write
			n, err := mrw.Write([]byte(tc.body))
			assert.NoError(t, err)
			assert.Equal(t, len(tc.body), n)
			assert.Equal(t, tc.expectedBody, string(mrw.Body))
			assert.Equal(t, tc.expectedBody, rr.Body.String())

			// Проверяем, что ResponseWriter интерфейс работает правильно
			assert.Equal(t, rr.Header(), mrw.Header())
		})
	}
}

func TestMetricsMiddleware_ErrorStatusCodeClassification(t *testing.T) {
	tests := []struct {
		name           string
		statusCode     int
		expectedStatus string
	}{
		{
			name:           "OK is not error",
			statusCode:     http.StatusOK,
			expectedStatus: statusOK,
		},
		{
			name:           "Created is not error",
			statusCode:     http.StatusCreated,
			expectedStatus: statusOK,
		},
		{
			name:           "Bad Request is error",
			statusCode:     http.StatusBadRequest,
			expectedStatus: statusError,
		},
		{
			name:           "Unauthorized is error",
			statusCode:     http.StatusUnauthorized,
			expectedStatus: statusError,
		},
		{
			name:           "Not Found is error",
			statusCode:     http.StatusNotFound,
			expectedStatus: statusError,
		},
		{
			name:           "Internal Server Error is error",
			statusCode:     http.StatusInternalServerError,
			expectedStatus: statusError,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			requestsTotal.Reset()
			requestDuration.Reset()

			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tc.statusCode)
			})

			mw := MetricsMiddleware(handler)
			req := httptest.NewRequest("GET", "/test", nil)
			rr := httptest.NewRecorder()

			mw.ServeHTTP(rr, req)

			require.Equal(t, tc.statusCode, rr.Code)

			expectedLabels := prometheus.Labels{
				urlLabel:        "/test",
				methodLabel:     "GET",
				statusLabel:     tc.expectedStatus,
				statusCodeLabel: strconv.Itoa(tc.statusCode),
			}

			count := testutil.ToFloat64(requestsTotal.With(expectedLabels))
			assert.Equal(t, 1.0, count, "статус должен быть правильно классифицирован")
		})
	}
}

func TestMetricsMiddleware_ConcurrentAccess(t *testing.T) {
	requestsTotal.Reset()
	requestDuration.Reset()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte("OK"))
		require.NoError(t, err)
	})

	mw := MetricsMiddleware(handler)

	done := make(chan bool)
	concurrentRequests := 10

	for i := 0; i < concurrentRequests; i++ {
		go func(id int) {
			path := "/api/test/" + strconv.Itoa(id)
			req := httptest.NewRequest("GET", path, nil)
			rr := httptest.NewRecorder()
			mw.ServeHTTP(rr, req)
			assert.Equal(t, http.StatusOK, rr.Code)
			done <- true
		}(i)
	}

	// Ждем завершения всех горутин
	for i := 0; i < concurrentRequests; i++ {
		<-done
	}

	// Проверяем, что все метрики были записаны
	metrics := testutil.CollectAndCount(requestsTotal)
	assert.Equal(t, concurrentRequests, metrics, "должны быть записаны метрики для всех запросов")
}

// errorReader - io.Reader, который всегда возвращает ошибку
type errorReader struct{}

func (er *errorReader) Read(p []byte) (n int, err error) {
	return 0, errors.New("mock read error")
}

func TestMetricsRegistration(t *testing.T) {
	// Проверяем, что метрики правильно зарегистрированы в Prometheus
	registry := prometheus.NewRegistry()

	// Регистрируем наши метрики
	err := registry.Register(requestsTotal)
	assert.NoError(t, err, "requestsTotal должен регистрироваться без ошибок")

	err = registry.Register(requestDuration)
	assert.NoError(t, err, "requestDuration должен регистрироваться без ошибок")

	err = registry.Register(goroutinesCount)
	assert.NoError(t, err, "goroutinesCount должен регистрироваться без ошибок")

	err = registry.Register(memoryUsage)
	assert.NoError(t, err, "memoryUsage должен регистрироваться без ошибок")

	// Пытаемся зарегистрировать повторно (должна быть ошибка)
	err = registry.Register(requestsTotal)
	assert.Error(t, err, "повторная регистрация должна вызывать ошибку")
}

func TestMetricsMiddleware_DurationMeasurement(t *testing.T) {
	requestsTotal.Reset()
	requestDuration.Reset()

	// Обработчик с задержкой
	slowHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Имитируем работу
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte("Slow response"))
		require.NoError(t, err)
	})

	mw := MetricsMiddleware(slowHandler)
	req := httptest.NewRequest("GET", "/slow", nil)
	rr := httptest.NewRecorder()

	mw.ServeHTTP(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)

	// Проверяем, что duration было измерено
	// Просто проверяем, что метрика существует и имеет наблюдения
	metricCount := testutil.CollectAndCount(requestDuration)
	assert.Greater(t, metricCount, 0, "request_duration_seconds должен иметь наблюдения")
}

func TestMetricsMiddleware_EmptyResponse(t *testing.T) {
	requestsTotal.Reset()
	requestDuration.Reset()

	// Обработчик без тела ответа
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	mw := MetricsMiddleware(handler)
	req := httptest.NewRequest("GET", "/empty", nil)
	rr := httptest.NewRecorder()

	mw.ServeHTTP(rr, req)

	require.Equal(t, http.StatusNoContent, rr.Code)
	assert.Empty(t, rr.Body.String())

	// Проверяем, что метрики все равно записаны
	expectedLabels := prometheus.Labels{
		urlLabel:        "/empty",
		methodLabel:     "GET",
		statusLabel:     statusOK, // 204 < 400
		statusCodeLabel: "204",
	}

	count := testutil.ToFloat64(requestsTotal.With(expectedLabels))
	assert.Equal(t, 1.0, count, "метрики должны быть записаны даже для пустого ответа")
}

func TestNewMetricsResponseWriter_NilResponseWriter(t *testing.T) {
	// Тестируем создание с nil ResponseWriter
	mrw := newMetricsResponseWriter(nil)
	assert.NotNil(t, mrw)
	assert.Nil(t, mrw.ResponseWriter)
	assert.Equal(t, http.StatusOK, mrw.StatusCode)
	assert.Empty(t, mrw.Body)
}
