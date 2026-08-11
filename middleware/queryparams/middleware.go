package queryparams

import (
	"fmt"
	"net/http"

	"github.com/nxtcoder17/ivy"
)

// Required returns a handler that rejects the request with 400 Bad Request when
// any of the given query params is missing.
//
// Example:
//
//	r.Get("/search", queryparams.Required("q"), func(c *ivy.Context) error { ... })
func Required(params ...string) ivy.Handler {
	return func(c *ivy.Context) error {
		q := c.URL().Query()
		for i := range params {
			if !q.Has(params[i]) {
				return ivy.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("missing query-param %q", params[i]))
			}
		}

		return c.Next()
	}
}
