package middleware

import (
	"fmt"
	"strings"
	"time"

	"github.com/nxtcoder17/ivy"
	"github.com/nxtcoder17/ivy/middleware/internal/logger"
)

type LoggerOptions struct {
	ShowQuery   *bool
	ShowHeaders *bool
	RouteFilter func(path string) bool
}

func (c *LoggerOptions) withDefaultsIfMissing() {
	if c.ShowQuery == nil {
		c.ShowQuery = ivy.Ptr(true)
	}

	if c.ShowHeaders == nil {
		c.ShowHeaders = ivy.Ptr(false)
	}

	if c.RouteFilter == nil {
		c.RouteFilter = nil
	}
}

func Logger(loggerOpts ...LoggerOptions) ivy.Handler {
	var opts LoggerOptions
	if len(loggerOpts) > 0 {
		opts = loggerOpts[0]
	}

	opts.withDefaultsIfMissing()

	return func(c *ivy.Context) error {
		route := c.Request().RequestURI

		if opts.RouteFilter != nil && !opts.RouteFilter(route) {
			return c.Next()
		}

		if *opts.ShowQuery {
			if idx := strings.IndexByte(route, '?'); idx != -1 {
				route = route[:idx]
			}
		}

		start := time.Now()

		rw := &logger.ResponseWriter{
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
