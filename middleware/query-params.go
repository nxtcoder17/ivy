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
				return ivy.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("missing query-param %q", params[i]))
			}
		}

		return c.Next()
	}
}
