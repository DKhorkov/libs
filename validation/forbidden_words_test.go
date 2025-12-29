package validation_test

import (
	"strings"
	"testing"

	"github.com/DKhorkov/libs/validation"
	"github.com/stretchr/testify/require"
)

func TestContainsForbiddenWords(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name:     "точное совпадение с основным матом",
			input:    "блядь",
			expected: true,
		},
		{
			name:     "регистронезависимая проверка",
			input:    "БлЯдЬ",
			expected: true,
		},
		{
			name:     "мат как часть слова",
			input:    "пиздобратия",
			expected: true,
		},
		{
			name:     "несколько матерных слов",
			input:    "хуй блядь ебать",
			expected: true,
		},
		{
			name:     "корректные слова без мата",
			input:    "привет мир",
			expected: false,
		},
		{
			name:     "пустая строка",
			input:    "",
			expected: false,
		},
		{
			name:     "похожие но допустимые слова",
			input:    "хлеб пистолет",
			expected: false,
		},
		{
			name:     "мат с пунктуацией",
			input:    "блядь!",
			expected: true,
		},
		{
			name:     "мат внутри длинного текста",
			input:    "Это текст со словом пизда посередине",
			expected: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			actual := validation.ContainsForbiddenWords(tc.input)
			require.Equal(t, tc.expected, actual, "Для входной строки: %q", tc.input)
		})
	}
}

func BenchmarkContainsForbiddenWords(b *testing.B) {
	benchmarks := []struct {
		name  string
		input string
	}{
		{
			name:  "короткий текст с матом",
			input: "блядь",
		},
		{
			name:  "длинный текст с матом",
			input: "Это длинный текст со словом пизда, которое нужно найти среди множества других слов",
		},
		{
			name:  "очень длинный текст без мата",
			input: strings.Repeat("нормальный текст ", 100),
		},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			for range b.N {
				validation.ContainsForbiddenWords(bm.input)
			}
		})
	}
}

func TestFalsePositives(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name  string
		input string
	}{
		{
			name:  "слово 'хлеб'",
			input: "хлеб",
		},
		{
			name:  "слово 'пистолет'",
			input: "пистолет",
		},
		{
			name:  "слово 'уютный'",
			input: "уютный",
		},
		{
			name:  "слово 'белый'",
			input: "белый",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			actual := validation.ContainsForbiddenWords(tc.input)
			require.False(t, actual, "Ожидалось отсутствие мата в слове: %q", tc.input)
		})
	}
}
