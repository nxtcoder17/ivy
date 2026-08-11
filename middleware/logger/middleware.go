package logger

import (
	"fmt"
	"strings"
	"time"

	"github.com/nxtcoder17/ivy"
)

type Options struct {
	// ShowQuery controls whether the query string is included in the logged
	// route. When false (the default), the query string is stripped from the
	// log line.
	ShowQuery   *bool
	ShowHeaders *bool
	RouteFilter func(path string) bool
}

func (c *Options) withDefaultsIfMissing() {
	if c.ShowQuery == nil {
		c.ShowQuery = ivy.Ptr(false)
	}

	if c.ShowHeaders == nil {
		c.ShowHeaders = ivy.Ptr(false)
	}

	if c.RouteFilter == nil {
		c.RouteFilter = nil
	}
}

// New returns a request logging middleware.
func New(opts ...Options) ivy.Handler {
	var o Options
	if len(opts) > 0 {
		o = opts[0]
	}

	o.withDefaultsIfMissing()

	return func(c *ivy.Context) error {
		route := c.Request().RequestURI

		if o.RouteFilter != nil && !o.RouteFilter(route) {
			return c.Next()
		}

		if !*o.ShowQuery {
			if idx := strings.IndexByte(route, '?'); idx != -1 {
				route = route[:idx]
			}
		}

		start := time.Now()

		rw := &responseWriter{
			StatusCode: 0,
			HttpRW:     c.ResponseWriter(),
		}

		c.SetResponseWriter(rw)

		c.Logger.Debug(fmt.Sprintf("❯❯ %s %s", c.Request().Method, route))
		defer func() {
			c.Logger.Info(fmt.Sprintf("❮❮ %d %s %s took %s", rw.StatusCode, c.Request().Method, route, time.Since(start).String()))
		}()

		return c.Next()
	}
}
