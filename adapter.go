package ivy

import "net/http"

func ToIvyHandler(h http.Handler) Handler {
	return func(c *Context) error {
		h.ServeHTTP(c.response, c.request)
		return nil
	}
}

func ToHTTPHandler(h Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h.ServeHTTP(w, r)
	})
}
