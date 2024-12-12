package middlewares

import (
	"context"
	"net/http"

	"github.com/DKhorkov/libs/contextlib"
)

var CookiesWriterName = "cookiesWriterName"

// CookiesMiddleware reads provided cookies from request and paste them into context for graphql purposes.
func CookiesMiddleware(handler http.Handler, cookieNames []string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		for _, cookieName := range cookieNames {
			cookie, err := r.Cookie(cookieName)
			if err != nil {
				continue
			}

			ctx := contextlib.SetValueToContext(context.Background(), CookiesWriterName, cookie)
			r = r.WithContext(ctx)
		}

		// Paste writer to context for writing cookies in resolvers purposes:
		ctx := contextlib.SetValueToContext(context.Background(), CookiesWriterName, w)
		r = r.WithContext(ctx)
		handler.ServeHTTP(w, r)
	})
}
