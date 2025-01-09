package ivy

import "net/http"

func ToIvyHandler(h http.Handler) Handler {
	return func(c *Context) error {
		h.ServeHTTP(c.response, c.request)
		return nil
	}
}
