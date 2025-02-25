package cookies

import (
	"net/http"
	"time"
)

// Set sets http.Cookie to provided http.ResponseWriter.
func Set(
	writer http.ResponseWriter,
	name string,
	value string,
	config Config,
) {
	http.SetCookie(
		writer,
		&http.Cookie{
			Name:     name,
			Value:    value,
			HttpOnly: config.HTTPOnly,
			Path:     config.Path,
			Domain:   config.Domain,
			Expires:  time.Now().UTC().Add(config.Expires),
			MaxAge:   config.MaxAge,
			SameSite: config.SameSite,
			Secure:   config.Secure,
		},
	)
}
