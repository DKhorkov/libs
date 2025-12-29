package http

import (
	"net/http"

	"github.com/DKhorkov/libs/contextlib"
)

var CookiesWriterName = "cookiesWriterName"

// CookiesMiddleware reads provided cookies from request and paste them into context for graphql purposes.
// After all operations - calls next handler.
func CookiesMiddleware(next http.Handler, cookieNames []string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		for _, cookieName := range cookieNames {
			cookie, err := r.Cookie(cookieName)
			if err != nil {
				continue
			}

			// Если будет ctx := то будут падать тесты и не будет работать
			ctx = contextlib.WithValue(ctx, cookieName, cookie) //nolint:all
			r = r.WithContext(ctx)
		}

		// Paste writer to context for writing cookies in resolvers purposes:
		ctx = contextlib.WithValue(ctx, CookiesWriterName, w)
		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)
	})
}
