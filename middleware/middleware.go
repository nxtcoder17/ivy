package middleware

import (
	"fmt"

	"github.com/nxtcoder17/ivy"
	"github.com/nxtcoder17/ivy/middleware/logger"
)

var Logger = logger.New()

func MustHaveQueryParams(params ...string) ivy.Handler {
	return func(c *ivy.Context) error {
		q := c.URL().Query()
		for i := range params {
			if !q.Has(params[i]) {
				return fmt.Errorf("INVALID request missing query-param (%q)", params[i])
			}
		}

		return c.Next()
	}
}
