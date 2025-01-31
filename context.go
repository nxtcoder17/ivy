package ivy

import (
	"context"
	"io"
	"net/http"
	"net/url"
)

type Context struct {
	request  *http.Request
	response http.ResponseWriter

	context.Context

	handlerIdx int
	next       func(c *Context) error

	// per-request key value store
	KV KV

	statusCode int
}

func NewContext(r *http.Request, w http.ResponseWriter) *Context {
	return &Context{
		Context:    r.Context(),
		request:    r,
		response:   w,
		handlerIdx: 0,
		next:       nil,
		KV:         KV{m: nil},
	}
}

// Calling Next() calls the next middleware in request handler chain
func (c *Context) Next() error {
	if c.next != nil {
		c.handlerIdx += 1
		return c.next(c)
	}
	return nil
}

// PathParam is like this `id` in this route path `/resource/{id}`
func (c *Context) PathParam(key string) string {
	return c.request.PathValue(key)
}

// QueryParam is like this `id` in this route path `/resource?id=hello-world`
func (c *Context) QueryParam(key string) string {
	return c.request.URL.Query().Get(key)
}

// Request() returns original http request
// for feature parity, until ivy gets a rigid API design
func (c *Context) Request() *http.Request {
	return c.request
}

// ResponseWriter() returns original http response writer
// for feature parity, until ivy gets a rigid API design
func (c *Context) ResponseWriter() http.ResponseWriter {
	return c.response
}

func (c *Context) SetResponseWriter(rw http.ResponseWriter) {
	c.response = rw
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

// :SECTION: request writers

func (c *Context) Body() io.ReadCloser {
	return c.request.Body
}

func (c *Context) ParseBodyInto(v any) error {
	b, err := io.ReadAll(c.request.Body)
	if err != nil {
		return err
	}

	return JSONDecoder(b, v)
}

// BodyParser is alias for ParseBodyInto
func (c *Context) BodyParser(v any) error {
	return c.ParseBodyInto(v)
}

// Write implements io.Writer.
func (c *Context) Write(p []byte) (n int, err error) {
	return c.response.Write(p)
}

var _ io.Writer = (*Context)(nil)

func (c *Context) Flush() {
	if v, ok := c.response.(http.Flusher); ok {
		v.Flush()
	}
}

var _ http.Flusher = (*Context)(nil)

// ----------- SECTION: response writers -------------------

// Status(int) let's you set status codes for your responses
// use it like
// `c.Status(201).SendString("OK")`
func (c *Context) Status(code int) *Context {
	c.response.WriteHeader(code)
	return c
}

func (c *Context) SendStatus(code int) error {
	c.Status(code)
	return nil
}

func (c *Context) SendBytes(b []byte) error {
	_, err := c.response.Write(b)
	return err
}

func (c *Context) SendString(s string) error {
	_, err := c.response.Write([]byte(s))
	return err
}

func (c *Context) SendJSON(s any) error {
	c.response.Header().Add("Content-Type", "application/json")
	b, err := JSONEncoder(s)
	if err != nil {
		return err
	}
	_, err = c.response.Write(b)
	return err
}

// JSON is alias for SendJSON
func (c *Context) JSON(s any) error {
	return c.SendJSON(s)
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
