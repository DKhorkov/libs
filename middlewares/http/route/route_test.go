package route

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFromPattern(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		pattern string
		want    string
	}{
		{"шаблона нет", "", Other},
		{"метод и путь", "GET /items/{id}", "/items/{id}"},
		{"метод через табуляцию", "GET\t/a", "/a"},
		{"несколько пробелов", "POST   /b", "/b"},
		{"хост без метода", "example.com/items/{id}", "/items/{id}"},
		{"метод и хост", "POST example.com/x/{rest...}", "/x/{rest...}"},
		{"только путь", "/static/", "/static/"},
		{"корень", "GET /", "/"},
		{"без слэша", "GET example.com", Other},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			r := httptest.NewRequestWithContext(
				context.Background(),
				http.MethodGet,
				"/whatever/42",
				http.NoBody,
			)
			r.Pattern = tc.pattern

			assert.Equal(t, tc.want, FromPattern(r))
		})
	}
}

// TestFromPatternThroughServeMux — шаблон ставит сам ServeMux, а не тест.
func TestFromPatternThroughServeMux(t *testing.T) {
	t.Parallel()

	var got string

	mux := http.NewServeMux()
	mux.HandleFunc("GET /items/{id}", func(_ http.ResponseWriter, r *http.Request) {
		got = FromPattern(r)
	})

	mux.ServeHTTP(
		httptest.NewRecorder(),
		httptest.NewRequestWithContext(
			context.Background(),
			http.MethodGet,
			"/items/42",
			http.NoBody,
		),
	)

	assert.Equal(t, "/items/{id}", got)
}

// TestMethod — стандартный метод как есть, прочие — OtherMethod: метод задаёт
// клиент, и без нормализации он был бы неограниченной меткой и именем спана.
func TestMethod(t *testing.T) {
	t.Parallel()

	for _, method := range []string{
		http.MethodGet, http.MethodHead, http.MethodPost, http.MethodPut, http.MethodPatch,
		http.MethodDelete, http.MethodOptions, http.MethodConnect, http.MethodTrace,
	} {
		assert.Equal(
			t,
			method,
			Method(httptest.NewRequestWithContext(context.Background(), method, "/", http.NoBody)),
			method,
		)
	}

	for _, method := range []string{"PROPFIND", "BREW", "get"} {
		assert.Equal(
			t,
			OtherMethod,
			Method(httptest.NewRequestWithContext(context.Background(), method, "/", http.NoBody)),
			method,
		)
	}
}
