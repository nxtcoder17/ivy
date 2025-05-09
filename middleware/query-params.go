package middleware

import (
	"fmt"
	"net/http"

	"github.com/nxtcoder17/ivy"
)

func RequiredQueryParams(params ...string) ivy.Handler {
	return func(c *ivy.Context) error {
		q := c.URL().Query()
		for i := range params {
			if !q.Has(params[i]) {
				return &ivy.HTTPError{
					Code: http.StatusBadRequest,
					Message:    fmt.Sprintf("missing query-param %q", params[i]),
				}
			}
		}

		return c.Next()
	}
}
