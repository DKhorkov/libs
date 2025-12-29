package cookies_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/DKhorkov/libs/cookies"
	"github.com/stretchr/testify/assert"
)

func TestSetCookie(t *testing.T) {
	t.Parallel()

	t.Run("Set cookie successfully", func(t *testing.T) {
		t.Parallel()

		var (
			name   = "test"
			value  = name
			w      = httptest.NewRecorder()
			config = cookies.Config{
				Path:     "/",
				Domain:   "",
				MaxAge:   0,
				Expires:  time.Minute * time.Duration(15),
				Secure:   false,
				HTTPOnly: false,
				SameSite: http.SameSite(1),
			}
			expectedTime = time.Now().UTC().Add(config.Expires).Format(http.TimeFormat)
			expected     = http.Header{
				"Set-Cookie": []string{"test=test; Path=/; Expires=" + expectedTime},
			}
		)

		cookies.Set(w, name, value, config)
		assert.Equal(t, expected, w.Header())
	})
}
