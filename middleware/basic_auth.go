// Ported from chi's basic_auth middleware implementation
// [Source](https://github.com/go-chi/chi/blob/v1.5.5/middleware/basic_auth.go)
package middleware

import (
	"crypto/subtle"
	"fmt"
	"net/http"

	"github.com/nxtcoder17/ivy"
)

// BasicAuth implements a simple middleware for HTTP Basic Authentication.
// It takes a realm (displayed in browser prompt) and a map of username -> password.
//
// Example:
//
//	r.Use(middleware.BasicAuth("Restricted", map[string]string{
//	    "admin": "secret",
//	    "user":  "password",
//	}))
func BasicAuth(realm string, creds map[string]string) ivy.Handler {
	return func(c *ivy.Context) error {
		user, pass, ok := c.Request().BasicAuth()
		if !ok {
			basicAuthFailed(c, realm)
			return nil
		}

		credPass, credUserOk := creds[user]
		if !credUserOk || subtle.ConstantTimeCompare([]byte(pass), []byte(credPass)) != 1 {
			basicAuthFailed(c, realm)
			return nil
		}

		return c.Next()
	}
}

func basicAuthFailed(c *ivy.Context, realm string) {
	c.SetHeader("WWW-Authenticate", fmt.Sprintf(`Basic realm="%s"`, realm))
	c.ResponseWriter().WriteHeader(http.StatusUnauthorized)
}
