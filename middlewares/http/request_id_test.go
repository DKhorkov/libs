package http_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DKhorkov/libs/contextlib"
	http2 "github.com/DKhorkov/libs/middlewares/http"
	"github.com/DKhorkov/libs/requestid"
	"github.com/stretchr/testify/require"
)

func TestRequestIDMiddleware(t *testing.T) {
	t.Parallel()

	t.Run("Generates and adds requestID to context", func(t *testing.T) {
		t.Parallel()

		// Создаём тестовый handler
		var capturedRequestID string

		nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Проверяем requestID в контексте
			ctxRequestID, err := contextlib.ValueFromContext[string](r.Context(), requestid.Key)
			require.NoError(t, err)

			capturedRequestID = ctxRequestID

			w.WriteHeader(http.StatusOK)
		})

		// Создаём middleware
		middleware := http2.RequestIDMiddleware(nextHandler)

		// Создаём тестовый запрос
		req := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
		rr := httptest.NewRecorder()

		// Выполняем запрос
		middleware.ServeHTTP(rr, req)

		// Проверяем результаты
		require.Equal(t, http.StatusOK, rr.Code)
		require.NotEmpty(t, capturedRequestID)
	})
}
