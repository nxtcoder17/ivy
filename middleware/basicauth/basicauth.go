// Ported from chi's basic_auth middleware implementation
// [Source](https://github.com/go-chi/chi/blob/v1.5.5/middleware/basic_auth.go)
package basicauth

import (
	"crypto/subtle"
	"fmt"
	"net/http"

	"github.com/nxtcoder17/ivy"
)

// New returns a middleware that enforces HTTP Basic Authentication.
// It takes a realm (displayed in browser prompt) and a map of username -> password.
//
// Example:
//
//	r.Use(basicauth.New("Restricted", map[string]string{
//	    "admin": "secret",
//	    "user":  "password",
//	}))
func New(realm string, creds map[string]string) ivy.Handler {
	return func(c *ivy.Context) error {
		user, pass, ok := c.Request().BasicAuth()
		if !ok {
			return basicAuthFailed(c, realm)
		}

		credPass, credUserOk := creds[user]
		if !credUserOk || subtle.ConstantTimeCompare([]byte(pass), []byte(credPass)) != 1 {
			return basicAuthFailed(c, realm)
		}

		return c.Next()
	}
}

func basicAuthFailed(c *ivy.Context, realm string) error {
	c.SetHeader("WWW-Authenticate", fmt.Sprintf(`Basic realm="%s"`, realm))
	return ivy.NewHTTPError(http.StatusUnauthorized, "Authentication Failed")
}
