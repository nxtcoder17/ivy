package logger

import (
	"fmt"
	"log/slog"
	"net/http"
	"time"
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

type HttpLogger struct {
	*options
}

type options struct {
	Logger      *slog.Logger
	ShowQuery   bool
	ShowHeaders bool
	SilentPaths []string
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

func WithSilentPaths(paths []string) option {
	return func(opts *options) {
		opts.SilentPaths = paths
	}
}

func WithLogger(logger *slog.Logger) option {
	return func(opts *options) {
		opts.Logger = logger
	}
}

func New(opts ...option) func(next http.Handler) http.Handler {
	dopts := &options{
		Logger:      slog.Default(),
		ShowQuery:   true,
		ShowHeaders: false,
		SilentPaths: []string{},
	}

	for i := range opts {
		opts[i](dopts)
	}

	h := &HttpLogger{options: dopts}
	return h.Middleware
}

func (h *HttpLogger) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		for i := range h.SilentPaths {
			if r.URL.Path == h.SilentPaths[i] {
				next.ServeHTTP(w, r)
				return
			}
		}

		lrw := &ResponseWriter{
			statusCode: 0,
			httpRW:     w,
		}

		timestart := time.Now()

		route := r.URL.Path
		if r.URL.RawQuery != "" {
			route = fmt.Sprintf("%s?%s", route, r.URL.RawQuery)
		}

		h.Logger.Debug(fmt.Sprintf("❯❯ %s %s", r.Method, route))
		defer func() {
			h.Logger.Info(fmt.Sprintf("❮❮ %d %s %s took %.2fs", lrw.statusCode, r.Method, route, time.Since(timestart).Seconds()))
		}()

		next.ServeHTTP(lrw, r)
	})
}
