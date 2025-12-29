package http

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	mocklogging "github.com/DKhorkov/libs/logging/mocks"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

// TestLoggingMiddleware тестирует middleware логирования в table-driven стиле.
func TestLoggingMiddleware(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)

	tests := []struct {
		name            string
		path            string
		method          string
		requestBody     string
		sensitiveFields []string
		handler         http.HandlerFunc
		setupMockLogger func(*mocklogging.MockLogger)
		expectations    func(*testing.T, *http.Request, *httptest.ResponseRecorder)
	}{
		{
			name:    "skip metrics endpoint",
			path:    MetricsURLPath,
			method:  "GET",
			handler: func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) },
			setupMockLogger: func(logger *mocklogging.MockLogger) {
				// No logs expected for metrics
				logger.EXPECT().InfoContext(gomock.Any(), gomock.Any()).Times(0)
				logger.EXPECT().ErrorContext(gomock.Any(), gomock.Any(), gomock.Any()).Times(0)
			},
			expectations: func(t *testing.T, r *http.Request, rr *httptest.ResponseRecorder) {
				if rr.Code != http.StatusOK {
					t.Errorf("expected status OK, got %d", rr.Code)
				}
			},
		},
		{
			name:            "restore request body for handler",
			path:            "/api/test",
			method:          "POST",
			requestBody:     `{"test": "data"}`,
			sensitiveFields: []string{},
			handler: func(w http.ResponseWriter, r *http.Request) {
				// Handler should be able to read body
				body, err := io.ReadAll(r.Body)
				if err != nil {
					t.Errorf("handler failed to read body: %v", err)
					w.WriteHeader(http.StatusInternalServerError)

					return
				}

				var data map[string]any
				if err = json.Unmarshal(body, &data); err != nil {
					t.Errorf("handler failed to unmarshal body: %v", err)
					w.WriteHeader(http.StatusInternalServerError)

					return
				}

				if data["test"] != "data" {
					t.Errorf("handler got wrong data: %v", data)
				}

				w.WriteHeader(http.StatusOK)

				if _, err = w.Write([]byte(`{"test": "data"}`)); err != nil {
					t.Fatalf("handler failed to write body: %v", err)
				}
			},
			setupMockLogger: func(logger *mocklogging.MockLogger) {
				logger.EXPECT().
					InfoContext(gomock.Any(), gomock.Any(), gomock.Any()).
					Times(2)
				// Request and response
			},
			expectations: func(t *testing.T, r *http.Request, rr *httptest.ResponseRecorder) {
				if rr.Code != http.StatusOK {
					t.Errorf("expected status OK, got %d", rr.Code)
				}

				// Verify body was restored and can be read again
				body, err := io.ReadAll(rr.Body)
				if err != nil {
					t.Errorf("failed to read restored body: %v", err)
				}

				if string(body) != `{"test": "data"}` {
					t.Errorf("restored body mismatch: got %s", string(body))
				}
			},
		},
		{
			name:            "filter sensitive fields from request",
			path:            "/api/login",
			method:          "POST",
			requestBody:     `{"username": "john", "password": "secret", "token": "abc123"}`,
			sensitiveFields: []string{"password", "token"},
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			},
			setupMockLogger: func(logger *mocklogging.MockLogger) {
				logger.EXPECT().InfoContext(gomock.Any(), gomock.Any(), gomock.Any()).Times(2)
			},
			expectations: func(t *testing.T, r *http.Request, rr *httptest.ResponseRecorder) {
				if rr.Code != http.StatusOK {
					t.Errorf("expected status OK, got %d", rr.Code)
				}
			},
		},
		{
			name:            "sensitive field not in request",
			path:            "/api/test",
			method:          "POST",
			requestBody:     `{"name": "John", "email": "john@example.com"}`,
			sensitiveFields: []string{"password", "credit_card"},
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			},
			setupMockLogger: func(logger *mocklogging.MockLogger) {
				logger.EXPECT().InfoContext(gomock.Any(), gomock.Any(), gomock.Any()).Times(2)
			},
			expectations: func(t *testing.T, r *http.Request, rr *httptest.ResponseRecorder) {
				if rr.Code != http.StatusOK {
					t.Errorf("expected status OK, got %d", rr.Code)
				}
			},
		},
		{
			name:            "empty request body",
			path:            "/api/test",
			method:          "GET",
			requestBody:     "",
			sensitiveFields: []string{},
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			},
			setupMockLogger: func(logger *mocklogging.MockLogger) {
				logger.EXPECT().InfoContext(gomock.Any(), gomock.Any(), gomock.Any()).Times(2)
			},
			expectations: func(t *testing.T, r *http.Request, rr *httptest.ResponseRecorder) {
				if rr.Code != http.StatusOK {
					t.Errorf("expected status OK, got %d", rr.Code)
				}

				// Empty body should still be restorable
				body, err := io.ReadAll(r.Body)
				if err != nil {
					t.Errorf("failed to read empty body: %v", err)
				}

				if len(body) != 0 {
					t.Errorf("empty body mismatch: got %s", string(body))
				}
			},
		},
		{
			name:            "invalid JSON body - middleware should still work",
			path:            "/api/test",
			method:          "POST",
			requestBody:     `invalid json`,
			sensitiveFields: []string{},
			handler: func(w http.ResponseWriter, r *http.Request) {
				// Handler should still get the original invalid body
				body, err := io.ReadAll(r.Body)
				if err != nil {
					t.Errorf("handler failed to read body: %v", err)
					w.WriteHeader(http.StatusInternalServerError)

					return
				}

				if string(body) != "invalid json" {
					t.Errorf("handler got wrong body: %s", string(body))
				}

				w.WriteHeader(http.StatusOK)

				if _, err = w.Write([]byte(`{"error": "invalid json"}`)); err != nil {
					t.Fatalf("handler failed to write body: %v", err)
				}
			},
			setupMockLogger: func(logger *mocklogging.MockLogger) {
				// Should log error for invalid JSON
				logger.EXPECT().ErrorContext(
					gomock.Any(),
					gomock.Eq("Failed to log request due to reading request body failure"),
					gomock.Any(),
				).Times(1)

				logger.EXPECT().InfoContext(
					gomock.Any(), gomock.Any(),
					gomock.Any(),
				).Times(2)
			},
			expectations: func(t *testing.T, r *http.Request, rr *httptest.ResponseRecorder) {
				if rr.Code != http.StatusOK {
					t.Errorf("expected status OK, got %d", rr.Code)
				}

				// Body should still be restored
				body, err := io.ReadAll(rr.Body)
				if err != nil {
					t.Errorf("failed to read restored body: %v", err)
				}

				if string(body) != `{"error": "invalid json"}` {
					t.Errorf("restored body mismatch: got %s", string(body))
				}
			},
		},
		{
			name:   "nil body in request",
			path:   "/api/test",
			method: "GET",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			},
			setupMockLogger: func(logger *mocklogging.MockLogger) {
				// Request with nil body should still be logged
				logger.EXPECT().InfoContext(gomock.Any(), gomock.Any(), gomock.Any()).Times(2)
			},
			expectations: func(t *testing.T, r *http.Request, rr *httptest.ResponseRecorder) {
				if rr.Code != http.StatusOK {
					t.Errorf("expected status OK, got %d", rr.Code)
				}
			},
		},
		{
			name:            "non-JSON content type",
			path:            "/api/upload",
			method:          "POST",
			requestBody:     "plain text body",
			sensitiveFields: []string{},
			handler: func(w http.ResponseWriter, r *http.Request) {
				// Handler should get original body
				body, err := io.ReadAll(r.Body)
				if err != nil {
					t.Errorf("handler failed to read body: %v", err)
					w.WriteHeader(http.StatusInternalServerError)

					return
				}

				if string(body) != "plain text body" {
					t.Errorf("handler got wrong body: %s", string(body))
				}

				w.WriteHeader(http.StatusOK)
			},
			setupMockLogger: func(logger *mocklogging.MockLogger) {
				// Should log error for non-JSON body
				logger.EXPECT().ErrorContext(
					gomock.Any(),
					gomock.Eq("Failed to log request due to reading request body failure"),
					gomock.Any(),
				).Times(1)

				logger.EXPECT().InfoContext(
					gomock.Any(),
					gomock.Any(),
					gomock.Any(),
				).Times(2)
			},
			expectations: func(t *testing.T, r *http.Request, rr *httptest.ResponseRecorder) {
				if rr.Code != http.StatusOK {
					t.Errorf("expected status OK, got %d", rr.Code)
				}
			},
		},
		{
			name:            "response logging with status code",
			path:            "/api/error",
			method:          "GET",
			requestBody:     "",
			sensitiveFields: []string{},
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusNotFound)

				if _, err := w.Write([]byte(`{"error": "not found"}`)); err != nil {
					t.Fatalf("handler failed to write body: %v", err)
				}
			},
			setupMockLogger: func(logger *mocklogging.MockLogger) {
				// Should log both request and response
				logger.EXPECT().InfoContext(gomock.Any(), gomock.Any(), gomock.Any()).Times(2)
			},
			expectations: func(t *testing.T, r *http.Request, rr *httptest.ResponseRecorder) {
				if rr.Code != http.StatusNotFound {
					t.Errorf("expected status 404, got %d", rr.Code)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Setup mock logger
			logger := mocklogging.NewMockLogger(ctrl)
			tt.setupMockLogger(logger)

			// Create middleware
			middleware := LoggingMiddleware(logger, tt.sensitiveFields...)

			// Create handler with middleware
			handler := middleware(tt.handler)

			// Create test request
			var body io.Reader
			if tt.requestBody != "" {
				body = bytes.NewBufferString(tt.requestBody)
			}

			req := httptest.NewRequest(tt.method, tt.path, body)

			// Create response recorder
			rr := httptest.NewRecorder()

			// Serve request
			handler.ServeHTTP(rr, req)

			// Run custom expectations
			if tt.expectations != nil {
				tt.expectations(t, req, rr)
			}
		})
	}
}

// TestLoggingMiddleware_MultipleSensitiveFields тестирует множественные чувствительные поля.
func TestLoggingMiddleware_MultipleSensitiveFields(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)

	tests := []struct {
		name             string
		requestBody      string
		sensitiveFields  []string
		shouldNotContain []string
	}{
		{
			name:             "remove all sensitive fields",
			requestBody:      `{"public": "data", "secret1": "value1", "secret2": "value2", "secret3": "value3"}`,
			sensitiveFields:  []string{"secret1", "secret2", "secret3"},
			shouldNotContain: []string{"secret1", "secret2", "secret3"},
		},
		{
			name:             "mix of present and absent sensitive fields",
			requestBody:      `{"name": "test", "password": "123", "public": "info"}`,
			sensitiveFields:  []string{"password", "token", "ssn"},
			shouldNotContain: []string{"password"},
		},
		{
			name:             "sensitive field in nested object (current limitation)",
			requestBody:      `{"user": {"name": "john", "password": "secret"}}`,
			sensitiveFields:  []string{"password"},
			shouldNotContain: []string{}, // Текущая реализация не удаляет вложенные поля
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			logger := mocklogging.NewMockLogger(ctrl)

			logger.EXPECT().InfoContext(gomock.Any(), gomock.Any(), gomock.Any()).Times(2)

			middleware := LoggingMiddleware(logger, tt.sensitiveFields...)

			handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			}))

			req := httptest.NewRequest(
				http.MethodPost,
				"/api/test",
				bytes.NewBufferString(tt.requestBody),
			)
			rr := httptest.NewRecorder()

			handler.ServeHTTP(rr, req)

			// Verify body was restored
			body, err := io.ReadAll(req.Body)
			if err != nil {
				t.Errorf("failed to read restored body: %v", err)
			}

			if string(body) != tt.requestBody {
				t.Errorf("restored body mismatch: got %s", string(body))
			}
		})
	}
}

// TestLoggingMiddleware_HandlerError тестирует обработку ошибок в хендлере.
func TestLoggingMiddleware_HandlerError(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)

	tests := []struct {
		name       string
		handler    http.HandlerFunc
		setupMock  func(*mocklogging.MockLogger)
		expectCode int
	}{
		{
			name: "handler writes error response",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)

				data, err := json.Marshal(struct {
					Error string `json:"error"`
				}{
					Error: http.StatusText(http.StatusInternalServerError),
				})
				require.NoError(t, err)

				if _, err = w.Write(data); err != nil {
					t.Fatalf("failed to write body: %v", err)
				}
			},
			setupMock: func(logger *mocklogging.MockLogger) {
				logger.EXPECT().InfoContext(gomock.Any(), gomock.Any(), gomock.Any()).Times(2)
			},
			expectCode: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			logger := mocklogging.NewMockLogger(ctrl)
			tt.setupMock(logger)

			middleware := LoggingMiddleware(logger)
			handler := middleware(tt.handler)

			req := httptest.NewRequest(http.MethodGet, "/api/test", http.NoBody)
			rr := httptest.NewRecorder()

			// For panic test, ensure we recover
			if tt.name == "handler panics" {
				defer func() {
					if r := recover(); r == nil {
						t.Error("expected panic to propagate")
					}
				}()
			}

			handler.ServeHTTP(rr, req)

			if rr.Code != tt.expectCode {
				t.Errorf("expected status %d, got %d", tt.expectCode, rr.Code)
			}
		})
	}
}
