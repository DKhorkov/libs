package intercepting_response_writer

import (
	"bufio"
	"errors"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestInterceptingResponseWriter(t *testing.T) {
	t.Parallel()

	t.Run("WriteHeader captures status code", func(t *testing.T) {
		t.Parallel()

		rr := httptest.NewRecorder()
		rw := &InterceptingResponseWriter{ResponseWriter: rr}

		rw.WriteHeader(http.StatusCreated)
		require.Equal(t, http.StatusCreated, rw.StatusCode)
		require.Equal(t, http.StatusCreated, rr.Code)
	})

	t.Run("Write captures body", func(t *testing.T) {
		t.Parallel()

		rr := httptest.NewRecorder()
		rw := &InterceptingResponseWriter{ResponseWriter: rr}

		body := []byte(`{"data":"test"}`)
		n, err := rw.Write(body)
		require.NoError(t, err)
		require.Equal(t, len(body), n)
		require.Equal(t, body, rw.Body)
		require.Equal(t, string(body), rr.Body.String())
	})
}

// MockHijacker — мок для http.ResponseWriter, реализующий http.Hijacker.
type MockHijacker struct {
	http.ResponseWriter

	conn       net.Conn
	readWriter *bufio.ReadWriter
	hijackErr  error
}

func (m *MockHijacker) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return m.conn, m.readWriter, m.hijackErr
}

func TestHijack_TableDriven(t *testing.T) {
	t.Parallel()
	// Определяем тестовые случаи
	tests := []struct {
		name           string
		responseWriter http.ResponseWriter
		expectConn     bool   // true, если ожидаем не-nil conn
		expectRW       bool   // true, если ожидаем не-nil ReadWriter
		expectErr      bool   // true, если ожидаем ошибку
		errMsg         string // подстрока для проверки сообщения ошибки
	}{
		{
			name: "Success: Hijacker returns valid conn and rw",
			responseWriter: &MockHijacker{
				conn:       &net.TCPConn{},
				readWriter: &bufio.ReadWriter{},
				hijackErr:  nil,
			},
			expectConn: true,
			expectRW:   true,
			expectErr:  false,
		},
		{
			name: "Hijacker returns error",
			responseWriter: &MockHijacker{
				hijackErr: errors.New("hijack failed"),
			},
			expectConn: false,
			expectRW:   false,
			expectErr:  true,
			errMsg:     "hijack failed",
		},
		{
			name:           "ResponseWriter does not implement Hijacker",
			responseWriter: http.ResponseWriter(nil), // или httptest.NewRecorder()
			expectConn:     false,
			expectRW:       false,
			expectErr:      true,
			errMsg:         "websocket: response does not implement http.Hijacker",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Создаём экземпляр InterceptingResponseWriter
			irw := &InterceptingResponseWriter{
				ResponseWriter: tt.responseWriter,
			}

			// Вызываем метод
			conn, rw, err := irw.Hijack()

			// Проверки
			if tt.expectErr {
				if err == nil {
					t.Fatal("Expected an error, got nil")
				}

				if tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("Expected error containing '%s', got '%s'", tt.errMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error, got %v", err)
				}
			}

			if tt.expectConn {
				if conn == nil {
					t.Error("Expected non-nil conn")
				}
			} else {
				if conn != nil {
					t.Error("Expected nil conn")
				}
			}

			if tt.expectRW {
				if rw == nil {
					t.Error("Expected non-nil ReadWriter")
				}
			} else {
				if rw != nil {
					t.Error("Expected nil ReadWriter")
				}
			}
		})
	}
}
