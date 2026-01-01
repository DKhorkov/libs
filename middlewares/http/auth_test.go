package http_test

import (
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"
	"time"

	"github.com/DKhorkov/libs/contextlib"
	http2 "github.com/DKhorkov/libs/middlewares/http"
	"github.com/DKhorkov/libs/security"
	"github.com/stretchr/testify/require"
)

// TestAuthMiddleware табличные тесты с реальным JWT.
func TestAuthMiddleware(t *testing.T) {
	t.Parallel()

	// Создаем реальный JWT конфиг для тестов
	securityConfig := security.Config{
		JWT: security.JWTConfig{
			SecretKey:       "test-secret-key-12345",
			Algorithm:       "HS256",
			RefreshTokenTTL: 24 * time.Hour,
			AccessTokenTTL:  1 * time.Hour,
		},
	}

	// Вспомогательная функция для генерации тестовых токенов
	generateTestToken := func(userID uint64) (string, error) {
		return security.GenerateJWT(
			userID,
			securityConfig.JWT.SecretKey,
			securityConfig.JWT.AccessTokenTTL,
			securityConfig.JWT.Algorithm,
		)
	}

	generateExpiredToken := func(userID uint64) (string, error) {
		return security.GenerateJWT(
			userID,
			securityConfig.JWT.SecretKey,
			-1*time.Hour, // Прошедшее время
			securityConfig.JWT.Algorithm,
		)
	}

	tests := []struct {
		name           string
		requestMethod  string
		requestPath    string
		cookieValue    string
		cookiePresent  bool
		cookieName     string
		ignoreURLs     []http2.IgnoreURL
		expectedStatus int
		expectUserID   uint64
		setupToken     func() string // Функция для генерации токена
	}{
		{
			name:           "Успешная аутентификация с валидным токеном",
			requestMethod:  "GET",
			requestPath:    "/api/v1/protected",
			cookieName:     "access_token",
			cookiePresent:  true,
			expectedStatus: http.StatusOK,
			expectUserID:   123,
			setupToken: func() string {
				token, err := generateTestToken(123)
				if err != nil {
					t.Fatalf("Не удалось сгенерировать тестовый токен: %v", err)
				}

				return token
			},
		},
		{
			name:           "Куки отсутствует",
			requestMethod:  "GET",
			requestPath:    "/api/v1/protected",
			cookieName:     "access_token",
			cookiePresent:  false,
			expectedStatus: http.StatusUnauthorized,
			setupToken:     func() string { return "" },
		},
		{
			name:           "Пустая строка в куки",
			requestMethod:  "POST",
			requestPath:    "/api/v1/protected",
			cookieName:     "access_token",
			cookiePresent:  true,
			cookieValue:    "",
			expectedStatus: http.StatusUnauthorized,
			setupToken:     func() string { return "" },
		},
		{
			name:           "Невалидный JWT токен (поддельный)",
			requestMethod:  "POST",
			requestPath:    "/api/v1/protected",
			cookieName:     "access_token",
			cookiePresent:  true,
			expectedStatus: http.StatusUnauthorized,
			expectUserID:   0,
			setupToken:     func() string { return "invalid.jwt.token" },
		},
		{
			name:           "Истёкший JWT токен",
			requestMethod:  "PUT",
			requestPath:    "/api/v1/protected",
			cookieName:     "access_token",
			cookiePresent:  true,
			expectedStatus: http.StatusUnauthorized,
			setupToken: func() string {
				token, err := generateExpiredToken(456)
				if err != nil {
					t.Fatalf("Не удалось сгенерировать просроченный токен: %v", err)
				}

				return token
			},
		},
		{
			name:           "JWT с другим секретным ключом",
			requestMethod:  "GET",
			requestPath:    "/api/v1/protected",
			cookieName:     "access_token",
			cookiePresent:  true,
			expectedStatus: http.StatusUnauthorized,
			setupToken: func() string {
				// Генерируем токен с другим ключом
				token, err := security.GenerateJWT(
					789,
					"different-secret-key",
					securityConfig.JWT.AccessTokenTTL,
					securityConfig.JWT.Algorithm,
				)
				if err != nil {
					t.Fatalf("Не удалось сгенерировать токен с другим ключом: %v", err)
				}

				return token
			},
		},
		{
			name:          "URL игнорируется по методу и пути",
			requestMethod: "GET",
			requestPath:   "/api/v1/public/users",
			cookieName:    "access_token",
			cookiePresent: false,
			ignoreURLs: []http2.IgnoreURL{
				{
					Methods: []string{"GET", "POST"},
					Path:    regexp.MustCompile(`^/api/v1/public/.*$`),
				},
			},
			expectedStatus: http.StatusOK,
			setupToken:     func() string { return "" },
		},
		{
			name:          "URL игнорируется по точному совпадению пути",
			requestMethod: "POST",
			requestPath:   "/api/v1/auth/login",
			cookieName:    "access_token",
			cookiePresent: false,
			ignoreURLs: []http2.IgnoreURL{
				{
					Methods: []string{"POST"},
					Path:    regexp.MustCompile(`^/api/v1/auth/login$`),
				},
			},
			expectedStatus: http.StatusOK,
			setupToken:     func() string { return "" },
		},
		{
			name:          "URL НЕ игнорируется - метод не совпадает",
			requestMethod: "DELETE",
			requestPath:   "/api/v1/public/users",
			cookieName:    "access_token",
			cookiePresent: false,
			ignoreURLs: []http2.IgnoreURL{
				{
					Methods: []string{"GET", "POST"},
					Path:    regexp.MustCompile(`^/api/v1/public/.*$`),
				},
			},
			expectedStatus: http.StatusUnauthorized,
			setupToken:     func() string { return "" },
		},
		{
			name:           "Другое имя куки",
			requestMethod:  "GET",
			requestPath:    "/api/v1/protected",
			cookieName:     "auth_token", // Другое имя куки
			cookiePresent:  false,
			ignoreURLs:     []http2.IgnoreURL{},
			expectedStatus: http.StatusUnauthorized,
			setupToken:     func() string { return "" },
		},
		{
			name:          "Несколько ignoreURLs - совпадает второй",
			requestMethod: "GET",
			requestPath:   "/health",
			cookieName:    "access_token",
			cookiePresent: false,
			ignoreURLs: []http2.IgnoreURL{
				{
					Methods: []string{"POST"},
					Path:    regexp.MustCompile(`^/api/.*$`),
				},
				{
					Methods: []string{"GET"},
					Path:    regexp.MustCompile(`^/health$`),
				},
			},
			expectedStatus: http.StatusOK,
			setupToken:     func() string { return "" },
		},
		{
			name:          "Метод OPTIONS игнорируется",
			requestMethod: "OPTIONS",
			requestPath:   "/api/v1/cors",
			cookieName:    "access_token",
			cookiePresent: false,
			ignoreURLs: []http2.IgnoreURL{
				{
					Methods: []string{"OPTIONS"},
					Path:    regexp.MustCompile(`.*`),
				},
			},
			expectedStatus: http.StatusOK,
			setupToken:     func() string { return "" },
		},
		{
			name:           "Некорректное значение в JWT (не число)",
			requestMethod:  "GET",
			requestPath:    "/api/v1/protected",
			cookieName:     "access_token",
			cookiePresent:  true,
			expectedStatus: http.StatusUnauthorized,
			setupToken: func() string {
				// Генерируем токен со строкой вместо числа
				token, err := security.GenerateJWT(
					"not-a-number",
					securityConfig.JWT.SecretKey,
					securityConfig.JWT.AccessTokenTTL,
					securityConfig.JWT.Algorithm,
				)
				if err != nil {
					t.Fatalf("Не удалось сгенерировать токен со строкой: %v", err)
				}

				return token
			},
		},
		{
			name:           "Пустой ignoreURLs срез",
			requestMethod:  "GET",
			requestPath:    "/api/v1/protected",
			cookieName:     "access_token",
			cookiePresent:  false,
			ignoreURLs:     []http2.IgnoreURL{},
			expectedStatus: http.StatusUnauthorized,
			setupToken:     func() string { return "" },
		},
		{
			name:           "Nil ignoreURLs (varargs пустой)",
			requestMethod:  "GET",
			requestPath:    "/api/v1/protected",
			cookieName:     "access_token",
			cookiePresent:  false,
			ignoreURLs:     nil,
			expectedStatus: http.StatusUnauthorized,
			setupToken:     func() string { return "" },
		},
		{
			name:           "JWT с нулевым TTL (мгновенно истекает)",
			requestMethod:  "GET",
			requestPath:    "/api/v1/protected",
			cookieName:     "access_token",
			cookiePresent:  true,
			ignoreURLs:     []http2.IgnoreURL{},
			expectedStatus: http.StatusUnauthorized,
			setupToken: func() string {
				// Генерируем токен с нулевым TTL
				token, err := security.GenerateJWT(
					777,
					securityConfig.JWT.SecretKey,
					0, // Нулевой TTL
					securityConfig.JWT.Algorithm,
				)
				if err != nil {
					t.Fatalf("Не удалось сгенерировать токен с нулевым TTL: %v", err)
				}

				return token
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Генерируем токен если требуется
			var token string

			if tt.cookiePresent {
				if tt.cookieValue != "" {
					token = tt.cookieValue
				} else {
					token = tt.setupToken()
				}
			}

			// Создаем middleware с тестовыми параметрами
			middleware := http2.AuthMiddleware(
				tt.cookieName,
				securityConfig,
				tt.ignoreURLs...,
			)

			// Создаем тестовый хендлер для проверки контекста
			var capturedUserID uint64

			testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Проверяем, что userID установлен в контексте
				capturedUserID, _ = contextlib.ValueFromContext[uint64](
					r.Context(),
					http2.UserIDContextKey,
				)

				w.WriteHeader(http.StatusOK)

				_, err := w.Write([]byte("OK"))
				require.NoError(t, err)
			})

			// Создаем запрос
			req := httptest.NewRequest(tt.requestMethod, tt.requestPath, http.NoBody)

			// Добавляем куки если требуется
			if tt.cookiePresent && token != "" {
				req.AddCookie(&http.Cookie{
					Name:  tt.cookieName,
					Value: token,
				})
			}

			// Создаем ResponseRecorder
			rr := httptest.NewRecorder()

			// Выполняем middleware
			middleware(testHandler).ServeHTTP(rr, req)

			require.Equal(t, tt.expectedStatus, rr.Code)
			require.Equal(t, tt.expectUserID, capturedUserID)
		})
	}
}

// TestAuthMiddlewareContext проверяет корректность установки контекста.
func TestAuthMiddlewareContext(t *testing.T) {
	t.Parallel()

	securityConfig := security.Config{
		JWT: security.JWTConfig{
			SecretKey:       "test-secret-key",
			Algorithm:       "HS256",
			RefreshTokenTTL: 24 * time.Hour,
			AccessTokenTTL:  1 * time.Hour,
		},
	}

	// Генерируем валидный токен
	token, err := security.GenerateJWT(
		12345,
		securityConfig.JWT.SecretKey,
		securityConfig.JWT.AccessTokenTTL,
		securityConfig.JWT.Algorithm,
	)
	if err != nil {
		t.Fatalf("Не удалось сгенерировать токен: %v", err)
	}

	middleware := http2.AuthMiddleware("access_token", securityConfig)

	// Хендлер для проверки контекста
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Проверяем что контекст содержит userID
		ctx := r.Context()
		userID, err := contextlib.ValueFromContext[uint64](ctx, http2.UserIDContextKey)
		require.NoError(t, err)

		if userID != 12345 {
			t.Errorf("wrong userID in context: got %v want 12345", userID)
			http.Error(w, "wrong userID", http.StatusInternalServerError)

			return
		}

		w.WriteHeader(http.StatusOK)
		_, err = w.Write([]byte("OK"))
		require.NoError(t, err)
	})

	// Создаем запрос с куки
	req := httptest.NewRequest(http.MethodGet, "/test", http.NoBody)
	req.AddCookie(&http.Cookie{
		Name:  "access_token",
		Value: token,
	})

	rr := httptest.NewRecorder()
	middleware(testHandler).ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}
}

// TestAuthMiddlewareIgnoreURLMultipleMethods проверяет несколько методов в ignoreURL.
func TestAuthMiddlewareIgnoreURLMultipleMethods(t *testing.T) {
	t.Parallel()

	securityConfig := security.Config{
		JWT: security.JWTConfig{
			SecretKey:       "test-secret-key",
			Algorithm:       "HS256",
			RefreshTokenTTL: 24 * time.Hour,
			AccessTokenTTL:  1 * time.Hour,
		},
	}

	ignoreURL := http2.IgnoreURL{
		Methods: []string{"GET", "POST", "PUT", "DELETE"},
		Path:    regexp.MustCompile(`^/api/v1/public/.*$`),
	}

	middleware := http2.AuthMiddleware(
		"access_token",
		securityConfig,
		ignoreURL,
	)

	tests := []struct {
		method       string
		path         string
		expectAuth   bool
		expectStatus int
	}{
		{"GET", "/api/v1/public/users", false, http.StatusOK},
		{"POST", "/api/v1/public/users", false, http.StatusOK},
		{"PUT", "/api/v1/public/users", false, http.StatusOK},
		{"DELETE", "/api/v1/public/users", false, http.StatusOK},
		{"PATCH", "/api/v1/public/users", true, http.StatusUnauthorized},
		{"GET", "/api/v1/private/users", true, http.StatusUnauthorized},
	}

	for _, tt := range tests {
		t.Run(tt.method+" "+tt.path, func(t *testing.T) {
			t.Parallel()

			var handlerCalled bool

			testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				handlerCalled = true

				w.WriteHeader(http.StatusOK)
			})

			req := httptest.NewRequest(tt.method, tt.path, http.NoBody)
			rr := httptest.NewRecorder()

			middleware(testHandler).ServeHTTP(rr, req)

			if status := rr.Code; status != tt.expectStatus {
				t.Errorf("handler returned wrong status code: got %v want %v",
					status, tt.expectStatus)
			}

			// Если ожидаем аутентификацию, хендлер не должен быть вызван
			if tt.expectAuth && handlerCalled {
				t.Error("handler was called when authentication was expected")
			}
		})
	}
}
