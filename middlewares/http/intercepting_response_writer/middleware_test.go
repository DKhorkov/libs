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

	// Вызывающий вправе переиспользовать буфер сразу после Write: так делает
	// fmt.Fprintln внутри http.Error, возвращая буфер в пул. Захват по ссылке
	// отдавал бы трассировке то, что в буфер написали следующим.
	t.Run("Write copies body instead of aliasing caller buffer", func(t *testing.T) {
		t.Parallel()

		rw := New(httptest.NewRecorder())

		buffer := []byte("invalid login or password\n")
		_, err := rw.Write(buffer)
		require.NoError(t, err)

		copy(buffer, "/Users/someone/go/pkg/mod/")
		require.Equal(t, "invalid login or password\n", string(rw.Body))
	})

	t.Run("Write accumulates body written in several chunks", func(t *testing.T) {
		t.Parallel()

		rr := httptest.NewRecorder()
		rw := New(rr)

		for _, chunk := range []string{`{"data":`, `"test"`, `}`} {
			_, err := rw.Write([]byte(chunk))
			require.NoError(t, err)
		}

		require.Equal(t, `{"data":"test"}`, string(rw.Body))
		require.False(t, rw.BodyTruncated)
		require.Equal(t, `{"data":"test"}`, rr.Body.String())
	})

	// Перехват нужен логам и трассировке, а не отдаче: ответ клиенту уходит
	// целиком, копия держится не длиннее MaxCapturedBodySize.
	t.Run("Write caps captured body but passes response through", func(t *testing.T) {
		t.Parallel()

		rr := httptest.NewRecorder()
		rw := New(rr)

		first := strings.Repeat("a", MaxCapturedBodySize-1)
		_, err := rw.Write([]byte(first))
		require.NoError(t, err)
		require.False(t, rw.BodyTruncated)

		n, err := rw.Write([]byte("bcd"))
		require.NoError(t, err)
		require.Equal(t, 3, n)

		require.Len(t, rw.Body, MaxCapturedBodySize)
		require.Equal(t, first+"b", string(rw.Body))
		require.True(t, rw.BodyTruncated)
		require.Equal(t, MaxCapturedBodySize+2, rr.Body.Len())
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
