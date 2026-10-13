package logging

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/DKhorkov/libs/logging"
	mocklogging "github.com/DKhorkov/libs/logging/mocks"
	"github.com/DKhorkov/libs/middlewares/http/intercepting_response_writer"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

// TestMiddleware тестирует middleware логирования в table-driven стиле.
func TestMiddleware(t *testing.T) {
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
			name:            "success with json answer",
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
			name:            "success with array answer",
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

				if _, err = w.Write([]byte(`[{"test": "data"}]`)); err != nil {
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

				if string(body) != `[{"test": "data"}]` {
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
				// Не-JSON тело — строка на уровне запроса, а не ошибка:
				logger.EXPECT().InfoContext(
					gomock.Any(),
					gomock.Eq(nonJSONBodyMessage),
					gomock.Any(),
				).Times(1)

				logger.EXPECT().InfoContext(
					gomock.Any(), gomock.Not(gomock.Eq(nonJSONBodyMessage)),
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
				// Не-JSON тело — строка на уровне запроса, а не ошибка:
				logger.EXPECT().InfoContext(
					gomock.Any(),
					gomock.Eq(nonJSONBodyMessage),
					gomock.Any(),
				).Times(1)

				logger.EXPECT().InfoContext(
					gomock.Any(),
					gomock.Not(gomock.Eq(nonJSONBodyMessage)),
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
		{
			// Перехват держит только начало длинного тела: разбирать обрезанный
			// JSON — значит писать в лог ошибку, которой не было.
			name:            "truncated json answer is not parsed",
			path:            "/api/large",
			method:          "GET",
			requestBody:     "",
			sensitiveFields: []string{},
			handler: func(w http.ResponseWriter, r *http.Request) {
				body := `["` + strings.Repeat(
					"a",
					intercepting_response_writer.MaxCapturedBodySize,
				) + `"]`
				if _, err := w.Write([]byte(body)); err != nil {
					t.Fatalf("handler failed to write body: %v", err)
				}
			},
			setupMockLogger: func(logger *mocklogging.MockLogger) {
				logger.EXPECT().InfoContext(gomock.Any(), gomock.Any(), gomock.Any()).Times(2)
				logger.EXPECT().ErrorContext(gomock.Any(), gomock.Any(), gomock.Any()).Times(0)
			},
			expectations: func(t *testing.T, r *http.Request, rr *httptest.ResponseRecorder) {
				require.Equal(t, intercepting_response_writer.MaxCapturedBodySize+4, rr.Body.Len())
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
			middleware := Middleware(logger, tt.sensitiveFields...)

			// Create handler with middleware
			handler := middleware(tt.handler)

			// Create test request
			var body io.Reader
			if tt.requestBody != "" {
				body = bytes.NewBufferString(tt.requestBody)
			}

			req := httptest.NewRequestWithContext(context.Background(), tt.method, tt.path, body)

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

// TestMiddleware_RedactsCredentials — значения учётных данных не попадают в
// лог: ни куки и заголовки авторизации запроса, ни Set-Cookie ответа.
//
// Лог писал заголовки и куки целиком, и у кого был доступ к файлу журнала,
// тот входил чужой сессией по accessToken/refreshToken из строки лога.
// Логгер — настоящий slog в буфер: секрет ищется в той самой строке, что
// уходит в файл, а не в аргументах мока.
func TestMiddleware_RedactsCredentials(t *testing.T) {
	t.Parallel()

	//nolint:gosec // тестовые значения: тест проверяет, что их нет в логе
	const (
		accessToken  = "access-token-secret-value"
		refreshToken = "refresh-token-secret-value"
		bearer       = "Bearer bearer-secret-value"
		proxyAuth    = "Basic proxy-secret-value"
		issuedToken  = "issued-token-secret-value"
		csrfToken    = "csrf-token-secret-value"
		refererCode  = "987654"
	)

	var buffer bytes.Buffer

	logger := slog.New(slog.NewJSONHandler(&buffer, nil))

	handler := Middleware(logger)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Хендлер видит настоящие значения: редакция касается только лога.
		cookie, err := r.Cookie("accessToken")
		require.NoError(t, err)
		require.Equal(t, accessToken, cookie.Value)
		require.Equal(t, bearer, r.Header.Get("Authorization"))

		http.SetCookie(w, &http.Cookie{
			Name:     "accessToken",
			Value:    issuedToken,
			Path:     "/",
			Secure:   true,
			HttpOnly: true,
			SameSite: http.SameSiteLaxMode,
		})
		w.WriteHeader(http.StatusOK)
	}))

	request := httptest.NewRequestWithContext(
		context.Background(),
		http.MethodGet,
		"/api/me",
		http.NoBody,
	)
	// Срез, а не map: порядок кук в логе проверяется ниже.
	for _, cookie := range [][2]string{{"accessToken", accessToken}, {"refreshToken", refreshToken}} {
		request.AddCookie(&http.Cookie{
			Name:     cookie[0],
			Value:    cookie[1],
			Secure:   true,
			HttpOnly: true,
			SameSite: http.SameSiteLaxMode,
		})
	}

	request.Header.Set("Authorization", bearer)
	request.Header.Set("Proxy-Authorization", proxyAuth)
	// Double-submit: значение заголовка равно куке csrfToken, и заголовок,
	// записанный как есть, возвращал бы в лог спрятанную куку.
	request.Header.Set("X-Csrf-Token", csrfToken)
	// Браузер без Referrer-Policy шлёт своему же сайту полный адрес страницы,
	// и код со страницы подтверждения почты ехал в лог с каждой её статикой.
	request.Header.Set(
		"Referer",
		"https://evolify.example/web/verify-email/"+refererCode+"?email=x",
	)

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)

	logged := buffer.String()
	for _, secret := range []string{
		accessToken,
		refreshToken,
		"bearer-secret-value",
		"proxy-secret-value",
		issuedToken,
		csrfToken,
		refererCode,
	} {
		require.NotContains(t, logged, secret)
	}

	// Имена остаются: видно, с какими куками пришёл запрос и какую выдал ответ.
	require.Contains(t, logged, `"Cookies":["accessToken","refreshToken"]`)
	require.Contains(t, logged, "accessToken="+redactedValue)
	require.Contains(t, logged, redactedValue)
	// От Referer остаётся сайт: откуда пришёл запрос, видно, с какой страницы — нет.
	require.Contains(t, logged, `"Referer":["https://evolify.example"]`)

	// Клиент получает настоящую куку, а запрос после middleware не изменён.
	require.Contains(t, recorder.Header().Get("Set-Cookie"), issuedToken)
	require.Equal(t, bearer, request.Header.Get("Authorization"))
	require.Contains(t, request.Header.Get("Cookie"), accessToken)
}

// TestMiddleware_RedactsNonCanonicalHeaderKeys — заголовок, записанный в map
// мимо канонизации, редактируется так же.
//
// net/http канонизирует имена входящих заголовков сам, но хендлер может
// записать заголовок ответа прямо в map (`w.Header()["set-cookie"]`), и
// такой ключ уходит клиенту как есть. Сравнение по каноническим именам его не
// видело бы, и токен попал бы в лог.
func TestMiddleware_RedactsNonCanonicalHeaderKeys(t *testing.T) {
	t.Parallel()

	//nolint:gosec // тестовые значения: тест проверяет, что их нет в логе
	const (
		requestSecret  = "request-secret-value"
		responseSecret = "response-secret-value"
	)

	var buffer bytes.Buffer

	logger := slog.New(slog.NewJSONHandler(&buffer, nil))

	handler := Middleware(logger)(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header()["set-cookie"] = []string{"session=" + responseSecret}
		w.WriteHeader(http.StatusOK)
	}))

	request := httptest.NewRequestWithContext(
		context.Background(),
		http.MethodGet,
		"/api/me",
		http.NoBody,
	)
	request.Header["authorization"] = []string{"Bearer " + requestSecret}

	handler.ServeHTTP(httptest.NewRecorder(), request)

	logged := buffer.String()
	require.NotContains(t, logged, requestSecret)
	require.NotContains(t, logged, responseSecret)
	require.Contains(t, logged, "session="+redactedValue)
}

// TestMiddleware_MultipleSensitiveFields тестирует множественные чувствительные поля.
func TestMiddleware_MultipleSensitiveFields(t *testing.T) {
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

			middleware := Middleware(logger, tt.sensitiveFields...)

			handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			}))

			req := httptest.NewRequestWithContext(context.Background(),
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

// TestMiddleware_HandlerError тестирует обработку ошибок в хендлере.
func TestMiddleware_HandlerError(t *testing.T) {
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
				// Ответ 5xx — строка ERROR сверх строк запроса и ответа.
				logger.EXPECT().
					ErrorContext(gomock.Any(), gomock.Eq(serverErrorMessage), gomock.Any()).
					Times(1)
			},
			expectCode: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			logger := mocklogging.NewMockLogger(ctrl)
			tt.setupMock(logger)

			middleware := Middleware(logger)
			handler := middleware(tt.handler)

			req := httptest.NewRequestWithContext(
				context.Background(),
				http.MethodGet,
				"/api/test",
				http.NoBody,
			)
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

// TestNew_WithRequestURL — адрес запроса в логе берётся из функции
// приложения, а не из r.URL целиком.
//
// В пути и query бывают секреты: токен сброса пароля в пути — это смена
// чужого пароля для любого, кто читает журнал, пока токен жив; почта в query —
// персональные данные. Middleware о них не знает, знает приложение: у него
// есть шаблон маршрута.
func TestNew_WithRequestURL(t *testing.T) {
	t.Parallel()

	//nolint:gosec // тестовые значения: тест проверяет, что их нет в логе
	const (
		pathSecret     = "path-secret-123456"
		querySecret    = "player@example.com"
		passwordSecret = "body-password-value"
		loggedURL      = "/api/users/password/forget/{token}?email"
	)

	var buffer bytes.Buffer

	logger := slog.New(slog.NewJSONHandler(&buffer, nil))

	handler := New(
		logger,
		WithSensitiveFields("password"),
		WithRequestURL(func(*http.Request) string { return loggedURL }),
	)(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))

	request := httptest.NewRequestWithContext(
		context.Background(),
		http.MethodPost,
		"/api/users/password/forget/"+pathSecret+"?email="+url.QueryEscape(querySecret),
		bytes.NewBufferString(`{"password":"`+passwordSecret+`","name":"player"}`),
	)

	handler.ServeHTTP(httptest.NewRecorder(), request)

	logged := buffer.String()
	for _, secret := range []string{
		pathSecret,
		querySecret,
		url.QueryEscape(querySecret),
		passwordSecret,
	} {
		require.NotContains(t, logged, secret)
	}

	// Обе строки — запрос и ответ — с адресом из функции, без поля Query.
	require.Equal(t, 2, strings.Count(logged, `"URL":"`+loggedURL+`"`))
	require.NotContains(t, logged, `"Query"`)
	require.Contains(t, logged, `"name":"player"`)
}

// TestNew_LogsResponseBodyOnlyOnErrors — тело ответа пишется в лог только при
// ошибке (статус 400 и выше).
//
// Успешный ответ несёт данные игрока: почту в профиле, код приватного лобби.
// Фильтр полей работал только для тела запроса, и строка ответа выдавала
// их целиком. Тело ошибки — текст ошибки, его в лог и нужно.
func TestNew_LogsResponseBodyOnlyOnErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		status  int
		body    string
		present bool
	}{
		{
			name:    "успешный ответ",
			status:  http.StatusOK,
			body:    `{"email":"player@example.com","lobby":{"code":"ABC123"}}`,
			present: false,
		},
		{
			name:    "ошибка",
			status:  http.StatusBadRequest,
			body:    "lobby is full",
			present: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var buffer bytes.Buffer

			logger := slog.New(slog.NewJSONHandler(&buffer, nil))

			handler := New(logger)(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(tc.status)

				_, err := w.Write([]byte(tc.body))
				require.NoError(t, err)
			}))

			recorder := httptest.NewRecorder()
			handler.ServeHTTP(recorder, httptest.NewRequestWithContext(
				context.Background(),
				http.MethodGet,
				"/api/me",
				http.NoBody,
			))

			// Клиент получает тело целиком: перехват — только для лога.
			require.Equal(t, tc.body, recorder.Body.String())

			logged := buffer.String()
			if tc.present {
				require.Contains(t, logged, tc.body)

				return
			}

			require.NotContains(t, logged, "player@example.com")
			require.NotContains(t, logged, "ABC123")
		})
	}
}

// TestNew_WithLevel — строки запроса и ответа пишутся на уровне WithLevel.
//
// Приложение держит уровень INFO ради редких событий (старт и остановка
// серверов, снятие с очереди), а две строки на каждый запрос — это почти
// весь объём журнала. На DEBUG их нет, пока уровень приложения не понизят для
// отладки.
func TestNew_WithLevel(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		loggerLevel slog.Level
		written     bool
	}{
		{name: "логгер INFO — строк нет", loggerLevel: slog.LevelInfo, written: false},
		{name: "логгер DEBUG — строки есть", loggerLevel: slog.LevelDebug, written: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var buffer bytes.Buffer

			logger := slog.New(
				slog.NewJSONHandler(&buffer, &slog.HandlerOptions{Level: tc.loggerLevel}),
			)

			handler := New(logger, WithLevel(logging.Levels.DEBUG))(
				http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
					w.WriteHeader(http.StatusOK)
				}),
			)

			handler.ServeHTTP(httptest.NewRecorder(), httptest.NewRequestWithContext(
				context.Background(),
				http.MethodGet,
				"/api/me",
				http.NoBody,
			))

			logged := buffer.String()
			if !tc.written {
				require.Empty(t, logged)

				return
			}

			require.Equal(t, 2, strings.Count(logged, `"level":"DEBUG"`))
			require.Contains(t, logged, `"msg":"Received new request"`)
			require.Contains(t, logged, `"msg":"Received response"`)
		})
	}
}

// TestNew_SkipsWorkWhenLevelDisabled — на выключенном уровне middleware сразу
// передаёт запрос дальше: без чтения тела, без перехвата ответа, без строк.
//
// Строки запроса ушли на DEBUG, но работа ради них шла на каждом запросе:
// загрузка аватара до 20 МиБ читалась в память целиком, тело каждого ответа
// копировалось до 1 МиБ, а не-JSON тело давало строку уровня ERROR — ложную
// ошибку в журнале, где должны быть только ошибки и редкие события.
func TestNew_SkipsWorkWhenLevelDisabled(t *testing.T) {
	t.Parallel()

	var buffer bytes.Buffer

	logger := slog.New(slog.NewJSONHandler(&buffer, &slog.HandlerOptions{Level: slog.LevelInfo}))

	handler := New(logger, WithLevel(logging.Levels.DEBUG))(
		http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			// Перехват — только ради статуса (5xx): тело не копируется.
			if intercepting, ok := w.(*intercepting_response_writer.InterceptingResponseWriter); ok {
				defer func() {
					require.Empty(
						t,
						intercepting.Body,
						"тело скопировано ради строки, которую не напишут",
					)
				}()
			}

			w.WriteHeader(http.StatusOK)

			_, err := w.Write([]byte("PNG"))
			require.NoError(t, err)
		}),
	)

	handler.ServeHTTP(httptest.NewRecorder(), httptest.NewRequestWithContext(
		context.Background(),
		http.MethodPost,
		"/api/users/avatar",
		bytes.NewBufferString(
			"--boundary\r\nContent-Disposition: form-data; name=\"avatar\"\r\n\r\nPNG\r\n",
		),
	))

	require.Empty(t, buffer.String())
}

// TestNew_NonJSONBodyIsNotAnError — не-JSON тело запроса (загрузка файла)
// отмечается на уровне строк запроса, а не ошибкой.
func TestNew_NonJSONBodyIsNotAnError(t *testing.T) {
	t.Parallel()

	var buffer bytes.Buffer

	logger := slog.New(slog.NewJSONHandler(&buffer, &slog.HandlerOptions{Level: slog.LevelDebug}))

	handler := New(logger, WithLevel(logging.Levels.DEBUG))(
		http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
		}),
	)

	handler.ServeHTTP(httptest.NewRecorder(), httptest.NewRequestWithContext(
		context.Background(),
		http.MethodPost,
		"/api/users/avatar",
		bytes.NewBufferString("not json"),
	))

	logged := buffer.String()
	require.NotContains(t, logged, `"level":"ERROR"`)
	require.Contains(t, logged, `"msg":"`+nonJSONBodyMessage+`"`)
}

// TestNew_LogsServerErrorsAtErrorLevel — ответ 5xx даёт строку ERROR на любом
// уровне строк запроса: упавший запрос виден в журнале и тогда, когда строки
// запросов ушли на DEBUG.
//
// WriteInternalError в приложении ошибку не пишет — отмечает её в спане, — и
// без этой строки пятисотка на уровне INFO не оставляла следа в журнале.
func TestNew_LogsServerErrorsAtErrorLevel(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		status     int
		errorLines int
	}{
		{name: "500", status: http.StatusInternalServerError, errorLines: 1},
		{name: "502", status: http.StatusBadGateway, errorLines: 1},
		{name: "404 — отказ, а не сбой", status: http.StatusNotFound, errorLines: 0},
		{name: "499 — клиент ушёл", status: 499, errorLines: 0},
		{name: "200", status: http.StatusOK, errorLines: 0},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var buffer bytes.Buffer

			logger := slog.New(
				slog.NewJSONHandler(&buffer, &slog.HandlerOptions{Level: slog.LevelInfo}),
			)

			handler := New(
				logger,
				WithLevel(logging.Levels.DEBUG),
				WithRequestURL(func(*http.Request) string { return "/api/things/{token}" }),
			)(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				// Тело не копируется: на выключенном уровне строк его не пишут.
				if intercepting, ok := w.(*intercepting_response_writer.InterceptingResponseWriter); ok {
					defer func() { require.Empty(t, intercepting.Body) }()
				}

				w.WriteHeader(tc.status)

				_, err := w.Write([]byte("internal error"))
				require.NoError(t, err)
			}))

			handler.ServeHTTP(httptest.NewRecorder(), httptest.NewRequestWithContext(
				context.Background(),
				http.MethodGet,
				"/api/things/s3cr3t",
				http.NoBody,
			))

			logged := buffer.String()
			require.Equal(t, tc.errorLines, strings.Count(logged, `"level":"ERROR"`))
			require.NotContains(t, logged, "s3cr3t")

			if tc.errorLines > 0 {
				require.Contains(t, logged, `"msg":"`+serverErrorMessage+`"`)
				require.Contains(t, logged, `"URL":"/api/things/{token}"`)
			}
		})
	}
}
