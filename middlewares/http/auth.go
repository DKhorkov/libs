package http

import (
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"slices"

	"github.com/DKhorkov/libs/contextlib"
	"github.com/DKhorkov/libs/security"
)

const (
	UserIDContextKey = "userID"
)

var ErrInvalidJWT = errors.New("invalid jwt token")

type IgnoreURL struct {
	Methods []string       `json:"methods"`
	Path    *regexp.Regexp `json:"path"`
}

func AuthMiddleware(
	accessTokenCookieName string,
	securityConfig security.Config,
	ignoreURLs ...IgnoreURL,
) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Не проверяем OPTIONS на аутентификацию
			if r.Method == http.MethodOptions {
				next.ServeHTTP(w, r)

				return
			}

			// Не проверяем метрики на аутентификацию
			if r.URL.Path == MetricsURLPath {
				next.ServeHTTP(w, r)

				return
			}

			// Если URL с вызванным методом должен игнорироваться - просто вызываем следующий хэндлер:
			for _, ignoreURL := range ignoreURLs {
				if ignoreURL.Path.MatchString(r.URL.Path) &&
					slices.Contains(ignoreURL.Methods, r.Method) {
					next.ServeHTTP(w, r)

					return
				}
			}

			accessTokenCookie, err := r.Cookie(accessTokenCookieName)
			if err != nil {
				http.Error(
					w,
					accessTokenCookieName+" cookie not provided",
					http.StatusUnauthorized,
				)

				return
			}

			accessTokenPayload, err := security.ParseJWT(
				accessTokenCookie.Value,
				securityConfig.JWT.SecretKey,
			)
			if err != nil {
				http.Error(
					w,
					fmt.Errorf("%w: %w", ErrInvalidJWT, err).Error(),
					http.StatusUnauthorized,
				)

				return
			}

			floatUserID, ok := accessTokenPayload.(float64)
			if !ok {
				http.Error(
					w,
					fmt.Errorf("%w: failed to parse userID", ErrInvalidJWT).Error(),
					http.StatusUnauthorized,
				)

				return
			}

			userID := uint64(floatUserID)

			ctx := contextlib.WithValue(
				r.Context(),
				UserIDContextKey,
				userID,
			) // Устанавливаем значения для испоьзования в хэндлерах

			r = r.WithContext(ctx)
			next.ServeHTTP(w, r)
		})
	}
}
