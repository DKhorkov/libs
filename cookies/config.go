package cookies

import (
	"net/http"
	"time"
)

type Config struct {
	// See http.Cookie as reference.

	Path    string        // optional
	Domain  string        // optional
	Expires time.Duration // optional

	// MaxAge=0 means no 'Max-Age' attribute specified.
	// MaxAge<0 means delete cookie now, equivalently 'Max-Age: 0'
	// MaxAge>0 means Max-Age attribute present and given in seconds
	MaxAge   int
	Secure   bool
	HTTPOnly bool

	//	SameSiteDefaultMode SameSite = iota + 1 (1)
	//	SameSiteLaxMode (2)
	//	SameSiteStrictMode (3)
	//	SameSiteNoneMode (4)
	SameSite http.SameSite
}
