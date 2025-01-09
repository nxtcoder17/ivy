package ivy

import (
	"context"
	"io"
	"net/http"
	"net/url"

	"github.com/go-chi/chi/v5"
)

type Context struct {
	context.Context

	request  *http.Request
	response http.ResponseWriter

	jsonEncoder JSONEncoder
	jsonDecoder JSONDecoder

	handlerIdx int
	next       func(c *Context) error

	kv map[string]any
}

// PathParam is like this `id` in this route path `/resource/{id}`
func (c *Context) PathParam(key string) string {
	return chi.URLParam(c.request, key)
}

// QueryParam is like this `id` in this route path `/resource?id=hello-world`
func (c *Context) QueryParam(key string) string {
	return c.request.URL.Query().Get(key)
}

// Calling Next() calls the next middleware in request handler chain
func (c *Context) Next() error {
	next := c.next
	return next(&Context{
		Context:     c.request.Context(),
		request:     c.request,
		response:    c.response,
		jsonEncoder: c.jsonEncoder,
		jsonDecoder: c.jsonDecoder,
		handlerIdx:  c.handlerIdx + 1,
		next:        next,
		kv:          c.kv,
	})
}

// SetHeaders() set http response headers
func (c *Context) SetHeader(key, value string) {
	c.response.Header().Set(key, value)
}

// GetHeaders() is http request headers
func (c *Context) GetHeaders() http.Header {
	return c.request.Header
}

func (c *Context) URL() *url.URL {
	return c.request.URL
}

func (c *Context) SetCookie(cookie *http.Cookie) {
	http.SetCookie(c.response, cookie)
}

func (c *Context) GetCookie(name string) (*http.Cookie, error) {
	return c.request.Cookie(name)
}

func (c *Context) AllCookies() []*http.Cookie {
	return c.request.Cookies()
}

func (c *Context) ClearCookie(key string) {
	http.SetCookie(c.response, &http.Cookie{
		Name:   key,
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	})
}

func (c *Context) Writer() io.Writer {
	return c.response
}

// JSON is alias for SendJSON
func (c *Context) JSON(s any) error {
	return c.SendJSON(s)
}

// Set sets a key into the request level Key-Value store
func (c *Context) Set(k string, v any) {
	if c.kv == nil {
		c.kv = make(map[string]any, 1)
	}
	c.kv[k] = v
}

// get fetches the value of key in request level KV store
// in case, key is not present default value is returned
func (c *Context) Get(k string) any {
	return c.kv[k]
}

// lookup fetches the value of key in request level KV store
// in case default value is not present, ok will be false
func (c *Context) Lookup(k string) (value any, ok bool) {
	v, ok := c.kv[k]
	return v, ok
}

// :SECTION: request writers

func (c *Context) Body() io.ReadCloser {
	return c.request.Body
}

func (c *Context) ParseBodyInto(v any) error {
	b, err := io.ReadAll(c.request.Body)
	if err != nil {
		return err
	}

	return c.jsonDecoder(b, v)
}

// BodyParser is alias for ParseBodyInto
func (c *Context) BodyParser(v any) error {
	return c.ParseBodyInto(v)
}

// :SECTION: response writers

// Write implements io.Writer.
func (c *Context) Write(p []byte) (n int, err error) {
	return c.response.Write(p)
}

var _ io.Writer = (*Context)(nil)

func (c *Context) SendStatus(code int) error {
	c.response.WriteHeader(code)
	return nil
}

func (c *Context) SendBytes(b []byte) error {
	_, err := c.response.Write(b)
	return err
}

func (c *Context) Flush() {
	if v, ok := c.response.(http.Flusher); ok {
		v.Flush()
	}
}

func (c *Context) SendString(s string) error {
	_, err := c.response.Write([]byte(s))
	return err
}

func (c *Context) SendJSON(s any) error {
	c.response.Header().Add("Content-Type", "application/json")
	b, err := c.jsonEncoder(s)
	if err != nil {
		return err
	}
	_, err = c.response.Write(b)
	return err
}

func (c *Context) SendHTML(s []byte) error {
	c.response.Header().Add("Content-Type", "text/html")
	_, err := c.response.Write(s)
	return err
}

func (c *Context) SendFile(fp string) error {
	http.ServeFile(c.response, c.request, fp)
	return nil
}
