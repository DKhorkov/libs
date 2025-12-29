package http

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestInterceptingResponseWriter(t *testing.T) {
	t.Parallel()

	t.Run("WriteHeader captures status code", func(t *testing.T) {
		t.Parallel()

		rr := httptest.NewRecorder()
		trw := &interceptingResponseWriter{ResponseWriter: rr}

		trw.WriteHeader(http.StatusCreated)
		require.Equal(t, http.StatusCreated, trw.StatusCode)
		require.Equal(t, http.StatusCreated, rr.Code)
	})

	t.Run("Write captures body", func(t *testing.T) {
		t.Parallel()

		rr := httptest.NewRecorder()
		trw := &interceptingResponseWriter{ResponseWriter: rr}

		body := []byte(`{"data":"test"}`)
		n, err := trw.Write(body)
		require.NoError(t, err)
		require.Equal(t, len(body), n)
		require.Equal(t, body, trw.Body)
		require.Equal(t, string(body), rr.Body.String())
	})
}
