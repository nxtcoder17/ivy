package logger2

import "net/http"

type logger struct{}

// ServeHTTP implements http.Handler.
func (l *logger) ServeHTTP(http.ResponseWriter, *http.Request) {
	panic("unimplemented")
}

var _ http.Handler = (*logger)(nil)
