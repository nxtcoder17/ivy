package logger

import (
	"net/http"
)

type ResponseWriter struct {
	StatusCode int
	HttpRW     http.ResponseWriter
}

// Header implements http.ResponseWriter.
func (rw *ResponseWriter) Header() http.Header {
	return rw.HttpRW.Header()
}

// Write implements http.ResponseWriter.
func (rw *ResponseWriter) Write(b []byte) (int, error) {
	if rw.StatusCode == 0 {
		// means, it is not set
		rw.StatusCode = http.StatusOK
	}
	return rw.HttpRW.Write(b)
}

// WriteHeader implements http.ResponseWriter.
func (rw *ResponseWriter) WriteHeader(statusCode int) {
	rw.StatusCode = statusCode
	rw.HttpRW.WriteHeader(statusCode)
}

// Flush implements http.Flusher.
func (rw *ResponseWriter) Flush() {
	flusher := rw.HttpRW.(http.Flusher)
	flusher.Flush()
}

var (
	_ http.Flusher        = (*ResponseWriter)(nil)
	_ http.ResponseWriter = (*ResponseWriter)(nil)
)
