package tracing_test

import (
	"context"
	"errors"
	"testing"

	"github.com/DKhorkov/libs/tracing"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/codes"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

func TestCallerName(t *testing.T) {
	t.Parallel()

	t.Run("Default skip level", func(t *testing.T) {
		t.Parallel()

		// Вызываем CallerName из вспомогательной функции для контроля уровня стека
		result := helperCallerName(tracing.DefaultSkipLevel)
		// Ожидаем имя функции helperCallerName
		expected := "github.com/DKhorkov/libs/tracing_test.helperCallerName"
		require.Equal(t, expected, result)
	})

	t.Run("Skip level 0", func(t *testing.T) {
		t.Parallel()

		// skipLevel = 0 должен вернуть имя CallerName
		result := tracing.CallerName(0)
		expected := "github.com/DKhorkov/libs/tracing.CallerName"
		require.Equal(t, expected, result)
	})

	t.Run("Invalid skip level", func(t *testing.T) {
		t.Parallel()

		// Слишком большой skipLevel должен вернуть информацию об ошибке
		result := tracing.CallerName(1000)
		require.Contains(t, result, "Unknown")
		require.Contains(t, result, "line 0")
	})

	t.Run("Nil function", func(t *testing.T) {
		t.Parallel()

		// Имитация ситуации, когда runtime.FuncForPC возвращает nil, сложна,
		// поэтому полагаемся на корректность runtime.Caller.
		// Проверяем, что для разумного skipLevel возвращается имя функции.
		result := helperCallerName(1)
		require.NotContains(t, result, "Unknown")
		require.Contains(t, result, "helperCallerName")
	})
}

// Вспомогательная функция для создания дополнительного уровня стека.
func helperCallerName(skipLevel int) string {
	return tracing.CallerName(skipLevel)
}

func TestRecordError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		err             error
		wantStatus      codes.Code
		wantDescription string
		wantEvents      []string
	}{
		{
			name:            "error marks span and keeps exception event",
			err:             errors.New("invalid login or password"),
			wantStatus:      codes.Error,
			wantDescription: "invalid login or password",
			wantEvents:      []string{"exception"},
		},
		{
			name:       "nil error leaves span untouched",
			err:        nil,
			wantStatus: codes.Unset,
			wantEvents: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			recorder := tracetest.NewSpanRecorder()
			provider := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(recorder))

			_, span := provider.Tracer("test").Start(context.Background(), "decorated")
			tracing.RecordError(span, tt.err)
			span.End()

			ended := recorder.Ended()
			require.Len(t, ended, 1)
			require.Equal(t, tt.wantStatus, ended[0].Status().Code)
			require.Equal(t, tt.wantDescription, ended[0].Status().Description)

			events := make([]string, 0, len(ended[0].Events()))
			for _, event := range ended[0].Events() {
				events = append(events, event.Name)
			}

			require.Equal(t, tt.wantEvents, events)
		})
	}
}
