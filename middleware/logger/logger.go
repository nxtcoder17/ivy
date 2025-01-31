package logger

import (
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/nxtcoder17/ivy"
)

type ResponseWriter struct {
	statusCode int
	httpRW     http.ResponseWriter
}

// Header implements http.ResponseWriter.
func (rw *ResponseWriter) Header() http.Header {
	return rw.httpRW.Header()
}

// Write implements http.ResponseWriter.
func (rw *ResponseWriter) Write(b []byte) (int, error) {
	if rw.statusCode == 0 {
		// means, it is not set
		rw.statusCode = http.StatusOK
	}
	return rw.httpRW.Write(b)
}

// WriteHeader implements http.ResponseWriter.
func (rw *ResponseWriter) WriteHeader(statusCode int) {
	rw.statusCode = statusCode
	rw.httpRW.WriteHeader(statusCode)
}

// Flush implements http.Flusher.
func (rw *ResponseWriter) Flush() {
	flusher := rw.httpRW.(http.Flusher)
	flusher.Flush()
}

var (
	_ http.Flusher        = (*ResponseWriter)(nil)
	_ http.ResponseWriter = (*ResponseWriter)(nil)
)

type options struct {
	Logger      *slog.Logger
	ShowQuery   bool
	ShowHeaders bool
	RouteFilter func(path string) bool
}

type option func(opts *options)

func ShowQuery(v bool) option {
	return func(opts *options) {
		opts.ShowQuery = v
	}
}

func ShowHeaders(v bool) option {
	return func(opts *options) {
		opts.ShowHeaders = v
	}
}

func WithRouteFilter(filter func(p string) bool) option {
	return func(opts *options) {
		opts.RouteFilter = filter
	}
}

func WithLogger(logger *slog.Logger) option {
	return func(opts *options) {
		opts.Logger = logger
	}
}

func New(opts ...option) func(c *ivy.Context) error {
	opt := &options{
		Logger:      slog.Default(),
		ShowQuery:   true,
		ShowHeaders: false,
		RouteFilter: nil,
	}

	for i := range opts {
		opts[i](opt)
	}

	return func(c *ivy.Context) error {
		route := c.URL().Path

		if opt.RouteFilter != nil && !opt.RouteFilter(route) {
			return c.Next()
		}

		if c.URL().RawQuery != "" {
			route = fmt.Sprintf("%s?%s", route, c.URL().RawQuery)
		}

		start := time.Now()

		rw := &ResponseWriter{
			statusCode: 0,
			httpRW:     c.ResponseWriter(),
		}

		c.SetResponseWriter(rw)

		opt.Logger.Debug(fmt.Sprintf("❯❯ %s %s", c.Request().Method, route))
		defer func() {
			opt.Logger.Info(fmt.Sprintf("❮❮ %d %s %s took %.2fs", rw.statusCode, c.Request().Method, route, time.Since(start).Seconds()))
		}()

		return c.Next()
	}
}
