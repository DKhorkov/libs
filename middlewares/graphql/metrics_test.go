package graphql

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	mocklogging "github.com/DKhorkov/libs/logging/mocks"
)

func TestMetricsMiddleware_RegularRequests(t *testing.T) {
	ctrl := gomock.NewController(t)
	logger := mocklogging.NewMockLogger(ctrl)

	successHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	errorHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})

	tests := []struct {
		name           string
		handler        http.Handler
		path           string
		expectedStatus int
		expectedLabels prometheus.Labels
	}{
		{
			name:           "successful request",
			handler:        successHandler,
			path:           "/api/users",
			expectedStatus: http.StatusOK,
			expectedLabels: prometheus.Labels{
				urlLabel:        "/api/users",
				statusLabel:     statusOK,
				statusCodeLabel: "200",
			},
		},
		{
			name:           "client error request",
			handler:        errorHandler,
			path:           "/api/users",
			expectedStatus: http.StatusNotFound,
			expectedLabels: prometheus.Labels{
				urlLabel:        "/api/users",
				statusLabel:     statusError,
				statusCodeLabel: "404",
			},
		},
		{
			name:           "metrics endpoint skipped",
			handler:        successHandler,
			path:           metricsURLPath,
			expectedStatus: http.StatusOK,
			expectedLabels: nil, // shouldn't be recorded
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Reset Metrics:
			requestsTotal.Reset()
			requestDuration.Reset()

			mw := MetricsMiddleware(tc.handler, logger)
			req := httptest.NewRequest("GET", tc.path, nil)
			rr := httptest.NewRecorder()

			mw.ServeHTTP(rr, req)

			require.Equal(t, tc.expectedStatus, rr.Code)

			if tc.expectedLabels != nil {
				count := testutil.ToFloat64(requestsTotal.With(tc.expectedLabels))
				require.Equal(t, 1.0, count, "metric count should be incremented")
			} else {
				// Verify no metrics were recorded for /metrics endpoint
				count := testutil.CollectAndCount(requestsTotal)
				require.Equal(t, 0, count, "no metrics should be recorded for /metrics endpoint")
			}
		})
	}
}

func TestMetricsMiddleware_GraphQLRequests(t *testing.T) {
	ctrl := gomock.NewController(t)
	logger := mocklogging.NewMockLogger(ctrl)

	tests := []struct {
		name                string
		setupMocks          func(logger *mocklogging.MockLogger)
		body                string
		mockStatus          int
		expectedURL         string
		expectedStatus      string
		expectedMetricCount float64
	}{
		{
			name:                "successful graphql query",
			body:                `{"query": "query GetUser { user(id: 1) { name } }"}`,
			mockStatus:          http.StatusOK,
			expectedURL:         "user",
			expectedStatus:      statusOK,
			expectedMetricCount: 1,
		},
		{
			name:                "graphql client error",
			body:                `{"query": "query GetUser { user(id: 1) { name } }"}`,
			mockStatus:          http.StatusBadRequest,
			expectedURL:         "user",
			expectedStatus:      statusError,
			expectedMetricCount: 1,
		},
		{
			name: "invalid graphql query",
			body: `{"query": "invalid query"}`,
			setupMocks: func(logger *mocklogging.MockLogger) {
				logger.
					EXPECT().
					ErrorContext(gomock.Any(), gomock.Any(), gomock.Any()).
					Times(1)
			},
			mockStatus:     http.StatusBadRequest,
			expectedURL:    graphqlURLPath, // fallback to original path
			expectedStatus: statusError,
		},
		{
			name: "invalid json",
			body: `invalid json`,
			setupMocks: func(logger *mocklogging.MockLogger) {
				logger.
					EXPECT().
					ErrorContext(gomock.Any(), gomock.Any(), gomock.Any()).
					Times(1)
			},
			mockStatus:     http.StatusBadRequest,
			expectedURL:    graphqlURLPath, // fallback to original path
			expectedStatus: statusError,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Reset Metrics:
			requestsTotal.Reset()
			requestDuration.Reset()

			if tc.setupMocks != nil {
				tc.setupMocks(logger)
			}

			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tc.mockStatus)
			})

			mw := MetricsMiddleware(handler, logger)
			req := httptest.NewRequest("POST", graphqlURLPath, bytes.NewBufferString(tc.body))
			rr := httptest.NewRecorder()

			mw.ServeHTTP(rr, req)

			require.Equal(t, tc.mockStatus, rr.Code)

			expectedLabels := prometheus.Labels{
				urlLabel:        tc.expectedURL,
				statusLabel:     tc.expectedStatus,
				statusCodeLabel: strconv.Itoa(tc.mockStatus),
			}

			count := testutil.ToFloat64(requestsTotal.With(expectedLabels))
			require.Equal(t, tc.expectedMetricCount, count, "metric count should be incremented")
		})
	}
}

func TestMetricsMiddleware_RequestBodyError(t *testing.T) {
	// Setup
	requestsTotal.Reset()
	requestDuration.Reset()

	ctrl := gomock.NewController(t)
	logger := mocklogging.NewMockLogger(ctrl)
	logger.
		EXPECT().
		ErrorContext(gomock.Any(), gomock.Any(), gomock.Any()).
		Times(1)

	// Create a request with a body that will fail to read
	req := httptest.NewRequest("POST", graphqlURLPath, &errorReader{})
	rr := httptest.NewRecorder()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	mw := MetricsMiddleware(handler, logger)
	mw.ServeHTTP(rr, req)

	// Should fall back to original path
	expectedLabels := prometheus.Labels{
		urlLabel:        graphqlURLPath,
		statusLabel:     statusOK,
		statusCodeLabel: "200",
	}

	var expectedMetricCount float64 = 0
	count := testutil.ToFloat64(requestsTotal.With(expectedLabels))
	require.Equal(t, expectedMetricCount, count, "metric count should be incremented")
}

// errorReader is an io.Reader that always returns an error
type errorReader struct{}

func (er *errorReader) Read(_ []byte) (n int, err error) {
	return 0, errors.New("mock read error")
}

func TestMetricsResponseWriter(t *testing.T) {
	rr := httptest.NewRecorder()
	mrw := newMetricsResponseWriter(rr)

	// Test WriteHeader
	mrw.WriteHeader(http.StatusNotFound)
	require.Equal(t, http.StatusNotFound, mrw.StatusCode)
	require.Equal(t, http.StatusNotFound, rr.Code)

	// Test Write
	body := "test body"
	n, err := mrw.Write([]byte(body))
	require.NoError(t, err)
	require.Equal(t, len(body), n)
	require.Equal(t, body, string(mrw.Body))
	require.Equal(t, body, rr.Body.String())
}
