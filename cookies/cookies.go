package cookies

import (
	"net/http"
	"time"
)

func SetCookie(
	writer http.ResponseWriter,
	name string,
	value string,
	cookieConfig CookieConfig,
) {
	http.SetCookie(
		writer,
		&http.Cookie{
			Name:     name,
			Value:    value,
			HttpOnly: cookieConfig.HTTPOnly,
			Path:     cookieConfig.Path,
			Domain:   cookieConfig.Domain,
			Expires:  time.Now().Add(cookieConfig.Expires),
			MaxAge:   cookieConfig.MaxAge,
			SameSite: cookieConfig.SameSite,
			Secure:   cookieConfig.Secure,
		},
	)
}
