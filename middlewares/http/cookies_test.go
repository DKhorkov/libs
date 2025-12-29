package http_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DKhorkov/libs/contextlib"
	http2 "github.com/DKhorkov/libs/middlewares/http"
	"github.com/stretchr/testify/require"
)

func TestCookiesMiddleware(t *testing.T) {
	t.Parallel()

	t.Run("Adds cookies to context", func(t *testing.T) {
		t.Parallel()

		cookieNames := []string{"session_id", "user_token"}
		req := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
		req.AddCookie(&http.Cookie{Name: "session_id", Value: "abc123"})
		req.AddCookie(&http.Cookie{Name: "user_token", Value: "xyz789"})

		var capturedCookies map[string]*http.Cookie

		nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			capturedCookies = make(map[string]*http.Cookie)

			for _, name := range cookieNames {
				cookie, err := contextlib.ValueFromContext[*http.Cookie](r.Context(), name)
				if err == nil {
					capturedCookies[name] = cookie
				}
			}

			w.WriteHeader(http.StatusOK)
		})

		middleware := http2.CookiesMiddleware(nextHandler, cookieNames)

		rr := httptest.NewRecorder()
		middleware.ServeHTTP(rr, req)

		require.Equal(t, http.StatusOK, rr.Code)
		require.Len(t, capturedCookies, 2)
		require.Equal(t, "abc123", capturedCookies["session_id"].Value)
		require.Equal(t, "xyz789", capturedCookies["user_token"].Value)
	})

	t.Run("Handles missing cookies", func(t *testing.T) {
		t.Parallel()

		cookieNames := []string{"session_id", "user_token"}
		req := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
		// Добавляем только одну cookie
		req.AddCookie(&http.Cookie{Name: "session_id", Value: "abc123"})

		var capturedCookies map[string]*http.Cookie

		nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			capturedCookies = make(map[string]*http.Cookie)

			for _, name := range cookieNames {
				cookie, err := contextlib.ValueFromContext[*http.Cookie](r.Context(), name)
				if err == nil {
					capturedCookies[name] = cookie
				}
			}

			w.WriteHeader(http.StatusOK)
		})

		middleware := http2.CookiesMiddleware(nextHandler, cookieNames)

		rr := httptest.NewRecorder()
		middleware.ServeHTTP(rr, req)

		require.Equal(t, http.StatusOK, rr.Code)
		require.Len(t, capturedCookies, 1)
		require.Equal(t, "abc123", capturedCookies["session_id"].Value)
		require.NotContains(t, capturedCookies, "user_token")
	})

	t.Run("Adds ResponseWriter to context", func(t *testing.T) {
		t.Parallel()

		cookieNames := []string{"session_id"}
		req := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
		req.AddCookie(&http.Cookie{Name: "session_id", Value: "abc123"})

		var capturedWriter http.ResponseWriter

		nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			writer, err := contextlib.ValueFromContext[http.ResponseWriter](
				r.Context(),
				http2.CookiesWriterName,
			)
			require.NoError(t, err)

			capturedWriter = writer

			w.WriteHeader(http.StatusCreated)
		})

		middleware := http2.CookiesMiddleware(nextHandler, cookieNames)

		rr := httptest.NewRecorder()
		middleware.ServeHTTP(rr, req)

		require.Equal(t, http.StatusCreated, rr.Code)
		require.NotNil(t, capturedWriter)
		// Проверяем, что capturedWriter — это rr
		require.Equal(t, rr, capturedWriter)
	})

	t.Run("Calls next handler with empty cookie names", func(t *testing.T) {
		t.Parallel()

		cookieNames := []string{}
		req := httptest.NewRequest(http.MethodGet, "/", http.NoBody)

		var handlerCalled bool

		nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			handlerCalled = true
			// Проверяем, что ResponseWriter всё равно добавлен
			writer, err := contextlib.ValueFromContext[http.ResponseWriter](
				r.Context(),
				http2.CookiesWriterName,
			)
			require.NoError(t, err)
			require.NotNil(t, writer)
			w.WriteHeader(http.StatusOK)
		})

		middleware := http2.CookiesMiddleware(nextHandler, cookieNames)

		rr := httptest.NewRecorder()
		middleware.ServeHTTP(rr, req)

		require.True(t, handlerCalled)
		require.Equal(t, http.StatusOK, rr.Code)
	})

	t.Run("Preserves original context values", func(t *testing.T) {
		t.Parallel()

		cookieNames := []string{"session_id"}
		ctx := contextlib.WithValue(context.Background(), "custom-key", "custom-value")
		req := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
		req = req.WithContext(ctx)
		req.AddCookie(&http.Cookie{Name: "session_id", Value: "abc123"})

		nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Проверяем оригинальное значение контекста
			value, err := contextlib.ValueFromContext[string](r.Context(), "custom-key")
			require.NoError(t, err)
			require.Equal(t, "custom-value", value)
			// Проверяем cookie
			cookie, err := contextlib.ValueFromContext[*http.Cookie](r.Context(), "session_id")
			require.NoError(t, err)
			require.Equal(t, "abc123", cookie.Value)
			// Проверяем ResponseWriter
			writer, err := contextlib.ValueFromContext[http.ResponseWriter](
				r.Context(),
				http2.CookiesWriterName,
			)
			require.NoError(t, err)
			require.NotNil(t, writer)
			w.WriteHeader(http.StatusOK)
		})

		middleware := http2.CookiesMiddleware(nextHandler, cookieNames)

		rr := httptest.NewRecorder()
		middleware.ServeHTTP(rr, req)

		require.Equal(t, http.StatusOK, rr.Code)
	})
}
