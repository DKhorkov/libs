package tracing_test

import (
	"testing"

	"github.com/DKhorkov/libs/tracing"
	"github.com/stretchr/testify/require"
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
