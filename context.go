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

	KV KV
}

type KV struct {
	m map[string]any
}

// Set sets a key into the request level Key-Value store
func (kv *KV) Set(k string, v any) {
	if kv.m == nil {
		kv.m = make(map[string]any, 1)
	}
	kv.m[k] = v
}

// Get fetches the value of key in request level KV store
// in case, key is not present default value is returned
func (kv *KV) Get(k string) any {
	return kv.m[k]
}

// Lookup fetches the value of key in request level KV store
// in case default value is not present, ok will be false
func (kv *KV) Lookup(k string) (any, bool) {
	v, ok := kv.m[k]
	return v, ok
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
	if c.next != nil {
		c.handlerIdx += 1
		return c.next(c)
	}
	return nil
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
