package ivy

import (
	"context"
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"
)

type Context struct {
	context.Context

	request  *http.Request
	response http.ResponseWriter

	jsonEncoder JSONEncoder
	jsonDecoder JSONDecoder

	next func(c *Context) error
}

func (c *Context) PathParam(key string) string {
	return chi.URLParam(c.request, key)
}

func (c *Context) QueryParam(key string) string {
	return c.request.URL.Query().Get(key)
}

func (c *Context) Next() error {
	return c.next(c)
}

func (c *Context) Writer() io.Writer {
	return c.response
}

func (c *Context) SendStatus(code int) error {
	c.response.WriteHeader(code)
	return nil
}

func (c *Context) SendString(s string) error {
	_, err := c.response.Write([]byte(s))
	return err
}

func (c *Context) SendJSON(s any) error {
	c.response.Header().Add("Content-Type", "application/json")
	b, err := c.jsonEncoder(s)
	if err != nil {
		// http.Error(c.response, err.Error(), 500)
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

// JSON is alias for SendJSON
func (c *Context) JSON(s any) error {
	return c.SendJSON(s)
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

// section: headers and cookies
func (c *Context) SetHeader(key, value string) {
	c.response.Header().Set(key, value)
}

func (c *Context) GetHeaders() http.Header {
	return c.request.Header
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
		Name:     key,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
	})
}
