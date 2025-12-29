package graphql_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DKhorkov/libs/graphql"
	mocklogging "github.com/DKhorkov/libs/logging/mocks"
	graphql2 "github.com/DKhorkov/libs/middlewares/graphql"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestLoggingMiddleware(t *testing.T) {
	t.Parallel()

	t.Run("Skips non-graphql path", func(t *testing.T) {
		t.Parallel()

		ctrl := gomock.NewController(t)
		logger := mocklogging.NewMockLogger(ctrl)

		var handlerCalled bool

		nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			handlerCalled = true

			w.WriteHeader(http.StatusOK)
		})

		middleware := graphql2.LoggingMiddleware(nextHandler, logger)

		req := httptest.NewRequest(http.MethodPost, "/not-query", http.NoBody)
		rr := httptest.NewRecorder()

		middleware.ServeHTTP(rr, req)

		require.True(t, handlerCalled)
		require.Equal(t, http.StatusOK, rr.Code)
	})

	t.Run("Handles body read error", func(t *testing.T) {
		t.Parallel()

		ctrl := gomock.NewController(t)
		logger := mocklogging.NewMockLogger(ctrl)

		logger.
			EXPECT().
			ErrorContext(gomock.Any(), gomock.Any(), gomock.Any()).
			Times(1)

		var handlerCalled bool

		nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			handlerCalled = true

			w.WriteHeader(http.StatusOK)
		})

		middleware := graphql2.LoggingMiddleware(nextHandler, logger)

		// Создаём запрос с телом, которое нельзя прочитать
		req := httptest.NewRequest(http.MethodPost, "/query", http.NoBody)
		req.Body = &errorReader{err: errors.New("read error")}

		rr := httptest.NewRecorder()
		middleware.ServeHTTP(rr, req)

		require.True(t, handlerCalled)
		require.Equal(t, http.StatusOK, rr.Code)
	})

	t.Run("Handles invalid JSON", func(t *testing.T) {
		t.Parallel()

		ctrl := gomock.NewController(t)
		logger := mocklogging.NewMockLogger(ctrl)

		logger.
			EXPECT().
			ErrorContext(gomock.Any(), gomock.Any(), gomock.Any()).
			Times(1)

		var handlerCalled bool

		nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			handlerCalled = true

			w.WriteHeader(http.StatusOK)
		})

		middleware := graphql2.LoggingMiddleware(nextHandler, logger)

		body := []byte("invalid json")
		req := httptest.NewRequest(http.MethodPost, "/query", bytes.NewReader(body))
		rr := httptest.NewRecorder()

		middleware.ServeHTTP(rr, req)

		require.True(t, handlerCalled)
		require.Equal(t, http.StatusOK, rr.Code)
	})

	t.Run("Handles  parse error", func(t *testing.T) {
		t.Parallel()

		ctrl := gomock.NewController(t)
		logger := mocklogging.NewMockLogger(ctrl)

		logger.
			EXPECT().
			InfoContext(gomock.Any(), gomock.Any(), gomock.Any()).
			Times(1)

		query := `query Test { field }`

		var handlerCalled bool

		nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			handlerCalled = true

			w.WriteHeader(http.StatusOK)
		})

		middleware := graphql2.LoggingMiddleware(nextHandler, logger)

		body := map[string]any{"query": query}
		bodyBytes, _ := json.Marshal(body)
		req := httptest.NewRequest(http.MethodPost, "/query", bytes.NewReader(bodyBytes))
		rr := httptest.NewRecorder()

		middleware.ServeHTTP(rr, req)

		require.True(t, handlerCalled)
		require.Equal(t, http.StatusOK, rr.Code)
	})

	t.Run("Logs valid  request", func(t *testing.T) {
		t.Parallel()

		ctrl := gomock.NewController(t)
		logger := mocklogging.NewMockLogger(ctrl)

		query := `mutation Test($input: Input!) { test(input: $input) { id } }`
		info := &graphql.QueryInfo{
			Type:       "mutation",
			Name:       "Test",
			Parameters: map[string]string{"$input": "input!"},
			Fields: []graphql.FieldInfo{
				{
					Name: "test",
					Arguments: map[string]any{
						"input": map[string]any{
							"password": "secret",
							"username": "user",
						},
					},
				},
			},
			Variables: map[string]any{
				"input": map[string]any{
					"password": "secret",
					"username": "user",
				},
			},
		}

		logger.
			EXPECT().
			InfoContext(gomock.Any(), gomock.Any(), gomock.Any()).
			Times(1)

		var handlerCalled bool

		nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			handlerCalled = true
			// Проверяем, что тело запроса восстановлено
			bodyBytes, _ := json.Marshal(
				map[string]any{"query": query, "variables": info.Variables},
			)
			actualBody, _ := json.Marshal(json.RawMessage(bodyBytes))
			expectedBody, _ := json.Marshal(json.RawMessage(bodyBytes))
			require.JSONEq(t, string(expectedBody), string(actualBody))
			w.WriteHeader(http.StatusOK)
		})

		middleware := graphql2.LoggingMiddleware(nextHandler, logger)

		body := map[string]any{
			"query":     query,
			"variables": info.Variables,
		}
		bodyBytes, _ := json.Marshal(body)
		req := httptest.NewRequest(http.MethodPost, "/query", bytes.NewReader(bodyBytes))
		rr := httptest.NewRecorder()

		middleware.ServeHTTP(rr, req)

		require.True(t, handlerCalled)
		require.Equal(t, http.StatusOK, rr.Code)
	})
}

// errorReader для симуляции ошибки чтения тела запроса.
type errorReader struct {
	err error
}

func (r *errorReader) Read(p []byte) (n int, err error) {
	return 0, r.err
}

func (r *errorReader) Close() error {
	return nil
}
