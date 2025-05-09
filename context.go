package ivy

import (
	"context"
	"io"
	"io/fs"
	"log/slog"
	"net/http"
	"net/url"
	"time"
)

type Context struct {
	request  *http.Request
	response http.ResponseWriter

	// Alias to request.Context()
	// It allows user to pass ivy Context in place of context.Context
	context.Context

	// for middleware management
	handlerIdx int
	next       func(c *Context) error

	// Logger is in context to allow middlewares to add extra key value pairs to the logging context
	Logger *slog.Logger

	// per-request key value store
	// useful to put arbitrary authentication constants, or user information or requestID like fields
	KV *KV
}

type ivyContextKey string

func newContext(r *http.Request, w http.ResponseWriter) *Context {
	ctx := &Context{
		Context:    r.Context(),
		request:    r,
		response:   w,
		handlerIdx: 0,
		next:       nil,
		KV:         &KV{},
		Logger:     Logger,
	}

	kvCtxKey := ivyContextKey("ivy.ctx.kv")

	// INFO: this is needed to ensure that when we are converting from ivy.Handler to http.HandlerFunc, or vice-versa, we can have KV store set by previous middlewares
	kv := r.Context().Value(kvCtxKey)
	if kv != nil {
		switch v := kv.(type) {
		case *KV:
			ctx.KV = v
		default:
			panic("it must not have happened, unknown type")
		}
	}

	vctx := context.WithValue(ctx.Context, kvCtxKey, ctx.KV)

	ctx.request = r.WithContext(vctx)
	ctx.Context = vctx

	return ctx
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

// SetResponseWriter must only be used when you want a custom response writer instead of normal http.ResponseWriter
// e.g use cases such as request logger
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
		Name:    key,
		Expires: time.Now().Add(-1 * time.Minute),
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

func (c *Context) SendFileFS(fs fs.FS, fp string) error {
	http.ServeFileFS(c.response, c.request, fs, fp)
	return nil
}
