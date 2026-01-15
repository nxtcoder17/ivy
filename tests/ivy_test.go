package ivy_test

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/nxtcoder17/ivy"
)

type Request struct {
	Method  string
	Route   string
	Body    []byte
	Headers map[string]string
}

type Response struct {
	StatusCode    int
	Body          []byte
	Headers       map[string]string
	cookie        []*http.Cookie
	cookieExpired []string
}

type Test struct{}

func TestHTTPPrimitives(t *testing.T) {
	tests := []struct {
		name string

		request Request

		handler ivy.Handler

		response Response
	}{
		{
			name: "1. [HTTP request] url path",

			request: Request{
				Method:  http.MethodGet,
				Route:   "/",
				Headers: map[string]string{},
			},

			handler: func(c *ivy.Context) error {
				return c.SendString(c.URL().Path)
			},

			response: Response{
				StatusCode: http.StatusOK,
				Body:       []byte("/"),
				Headers:    map[string]string{},
			},
		},

		{
			name: "2. [HTTP request] query params",

			request: Request{
				Method:  http.MethodGet,
				Route:   "/?msg=hi",
				Headers: map[string]string{},
			},

			handler: func(c *ivy.Context) error {
				return c.SendString(c.QueryParam("msg"))
			},

			response: Response{
				StatusCode: http.StatusOK,
				Body:       []byte("hi"),
				Headers:    map[string]string{},
			},
		},

		{
			name: "3. [HTTP request] headers",

			request: Request{
				Method: http.MethodGet,
				Route:  "/",
				Headers: map[string]string{
					"Authorization": "sample",
				},
			},

			handler: func(c *ivy.Context) error {
				return c.SendJSON(c.GetHeaders()["Authorization"])
			},

			response: Response{
				StatusCode: http.StatusOK,
				Body:       []byte(`["sample"]`),
				Headers:    map[string]string{},
			},
		},

		{
			name: "4. [HTTP Response] sending string",

			request: Request{
				Method: http.MethodGet,
				Route:  "/",
			},

			handler: func(c *ivy.Context) error {
				return c.SendString("OK")
			},

			response: Response{
				StatusCode: http.StatusOK,
				Body:       []byte("OK"),
			},
		},

		{
			name: "5. [HTTP Response] sending bytes",

			request: Request{
				Method: http.MethodGet,
				Route:  "/",
			},

			handler: func(c *ivy.Context) error {
				return c.SendBytes([]byte("hello world"))
			},

			response: Response{
				StatusCode: http.StatusOK,
				Body:       []byte("hello world"),
			},
		},

		{
			name: "6. [HTTP Response] sending json",

			request: Request{
				Method: http.MethodGet,
				Route:  "/",
			},

			handler: func(c *ivy.Context) error {
				return c.JSON(map[string]string{"message": "hello"})
			},

			response: Response{
				StatusCode: http.StatusOK,
				Body:       []byte("{\"message\":\"hello\"}"),
			},
		},

		{
			name: "7. [HTTP Response] other status codes",

			request: Request{
				Method: http.MethodGet,
				Route:  "/",
			},

			handler: func(c *ivy.Context) error {
				return c.SendStatus(205)
			},

			response: Response{
				StatusCode: 205,
				Body:       nil,
			},
		},

		{
			name: "8. [HTTP Response] sending cookie",

			request: Request{
				Method: http.MethodGet,
				Route:  "/",
			},

			handler: func(c *ivy.Context) error {
				c.SetCookie(&http.Cookie{
					Name:  "hello",
					Value: "world",
				})

				return c.SendString("OK")
			},

			response: Response{
				StatusCode: http.StatusOK,
				Body:       []byte("OK"),
				cookie: []*http.Cookie{
					{
						Name:  "hello",
						Value: "world",
					},
				},
			},
		},

		{
			name: "9. [HTTP Request] sending cookie",

			request: Request{
				Method: http.MethodGet,
				Route:  "/",
				Headers: map[string]string{
					"cookie": "hello=world",
				},
			},

			handler: func(c *ivy.Context) error {
				cookie, err := c.GetCookie("hello")
				if err != nil {
					return err
				}
				return c.SendString(cookie.Value)
			},

			response: Response{
				StatusCode: http.StatusOK,
				Body:       []byte("world"),
			},
		},

		{
			name: "10. [HTTP Request] clearing cookie",

			request: Request{
				Method: http.MethodGet,
				Route:  "/",
				Headers: map[string]string{
					"cookie": "hello=world",
				},
			},

			handler: func(c *ivy.Context) error {
				c.ClearCookie("hello")
				return c.SendStatus(http.StatusOK)
			},

			response: Response{
				StatusCode:    http.StatusOK,
				Body:          nil,
				cookieExpired: []string{"hello"},
			},
		},

		{
			name: "10. [HTTP Response] sending headers",

			request: Request{
				Method: http.MethodGet,
				Route:  "/",
			},

			handler: func(c *ivy.Context) error {
				c.SetHeader("from", "testing router")
				return c.SendString("hello")
			},

			response: Response{
				StatusCode: http.StatusOK,
				Body:       []byte("hello"),
			},
		},

		{
			name: "11. [HTTP Request] reading header values from request",

			request: Request{
				Method: http.MethodGet,
				Route:  "/",
				Headers: map[string]string{
					"hi": "hello",
				},
			},

			handler: func(c *ivy.Context) error {
				return c.SendString(c.GetHeaders().Get("hi"))
			},

			response: Response{
				StatusCode: http.StatusOK,
				Body:       []byte("hello"),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.request.Method, tt.request.Route, bytes.NewReader(tt.request.Body))

			for k, v := range tt.request.Headers {
				req.Header.Add(k, v)
			}

			w := httptest.NewRecorder()

			tt.handler.ServeHTTP(w, req)

			res := w.Result()
			defer res.Body.Close()

			if res.StatusCode != tt.response.StatusCode {
				t.Errorf("expected status code\n\t got: %d\n\twant: %d\n", res.StatusCode, tt.response.StatusCode)
			}

			data, err := io.ReadAll(res.Body)
			if err != nil {
				t.Errorf("expected error to be nil got %v", err)
			}

			if string(data) != string(tt.response.Body) {
				t.Errorf("expected response body\n\t got: %s\n\twant: %s\n", data, tt.response.Body)
			}

			if tt.response.cookie != nil {
				for _, ce := range tt.response.cookieExpired {
					c, err := req.Cookie(ce)
					if err != nil {
						panic(err)
					}

					if time.Now().Before(c.Expires) {
						t.Errorf("expected cookie (%s), to be expired, but it is not", ce)
					}
				}

				if fmt.Sprintf("%+v", res.Cookies()) != fmt.Sprintf("%+v", tt.response.cookie) {
					t.Errorf("expected cookie body\n\t got: %+v\n\twant: %+v\n", res.Cookies(), tt.response.cookie)
				}
			}
		})
	}
}

func TestRouter(t *testing.T) {
	tests := []struct {
		name string

		request Request

		router func(r *ivy.Router)

		errorHandler ivy.ErrorHandler

		response Response
	}{
		{
			name: "1. [GET /] no middlewares",
			request: Request{
				Method: http.MethodGet,
				Route:  "/",
			},
			router: func(r *ivy.Router) {
				r.Get("/", func(c *ivy.Context) error {
					return c.SendString("OK")
				})
			},
			response: Response{
				StatusCode: 200,
				Body:       []byte("OK"),
			},
		},

		{
			name: "2. [GET /] with middleware setting a key into kv-store",
			request: Request{
				Method: http.MethodGet,
				Route:  "/",
			},
			router: func(r *ivy.Router) {
				r.Get("/",
					func(c *ivy.Context) error {
						c.KV.Set("hello", "world")
						return c.Next()
					},
					func(c *ivy.Context) error {
						return c.SendString(c.KV.Get("hello").(string))
					})
			},
			response: Response{
				StatusCode: 200,
				Body:       []byte("world"),
			},
		},

		{
			name: "3. [GET /]with default error handler",
			request: Request{
				Method: http.MethodGet,
				Route:  "/",
			},
			router: func(r *ivy.Router) {
				r.Get("/",
					func(c *ivy.Context) error {
						return fmt.Errorf("this is an error")
					})
			},
			response: Response{
				StatusCode: 500,
				Body:       []byte("this is an error\n"),
			},
		},

		{
			name: "4. [GET /] with custom error handler",
			request: Request{
				Method: http.MethodGet,
				Route:  "/",
			},
			router: func(r *ivy.Router) {
				r.Get("/",
					func(c *ivy.Context) error {
						return fmt.Errorf("this is an error")
					})
			},
			errorHandler: func(c *ivy.Context, err error) {
				c.Status(http.StatusInternalServerError).SendString("| ERROR | " + err.Error())
			},

			response: Response{
				StatusCode: 500,
				Body:       []byte("| ERROR | this is an error"),
			},
		},

		{
			name: "5. [GET /] with path params",
			request: Request{
				Method: http.MethodGet,
				Route:  "/hello",
			},
			router: func(r *ivy.Router) {
				r.Get("/{p}",
					func(c *ivy.Context) error {
						return c.SendString(c.PathParam("p"))
					})
			},
			response: Response{
				StatusCode: 200,
				Body:       []byte("hello"),
			},
		},

		{
			name: "6. [POST /] sending string",

			request: Request{
				Method: http.MethodPost,
				Route:  "/",
				Headers: map[string]string{
					"Content-Type": "application/json",
				},
				Body: []byte("{\"message\":\"hello\"}"),
			},

			router: func(r *ivy.Router) {
				r.Post("/", func(c *ivy.Context) error {
					var x struct {
						Message string `json:"message"`
					}
					if err := c.ParseBodyInto(&x); err != nil {
						return err
					}
					return c.SendString(x.Message)
				})
			},

			response: Response{
				StatusCode: http.StatusOK,
				Body:       []byte("hello"),
			},
		},

		{
			name: "7. [PUT /] sending string",

			request: Request{
				Method: http.MethodPut,
				Route:  "/",
				Headers: map[string]string{
					"Content-Type": "application/json",
				},
				Body: []byte("{\"message\":\"hello\"}"),
			},

			router: func(r *ivy.Router) {
				r.Put("/", func(c *ivy.Context) error {
					var x struct {
						Message string `json:"message"`
					}
					if err := c.ParseBodyInto(&x); err != nil {
						return err
					}
					return c.SendString(x.Message)
				})
			},

			response: Response{
				StatusCode: http.StatusOK,
				Body:       []byte("hello"),
			},
		},

		{
			name: "8. [DELETE /] sending string",

			request: Request{
				Method: http.MethodDelete,
				Route:  "/",
				Headers: map[string]string{
					"Content-Type": "application/json",
				},
				Body: []byte("{\"message\":\"hello\"}"),
			},

			router: func(r *ivy.Router) {
				r.Delete("/", func(c *ivy.Context) error {
					var x struct {
						Message string `json:"message"`
					}
					if err := c.ParseBodyInto(&x); err != nil {
						return err
					}
					return c.SendString(x.Message)
				})
			},

			response: Response{
				StatusCode: http.StatusOK,
				Body:       []byte("hello"),
			},
		},

		{
			name: "9. [MY_HTTP_METHOD /] sending string",

			request: Request{
				Method: "MY_HTTP_METHOD",
				Route:  "/",
				Headers: map[string]string{
					"Content-Type": "application/json",
				},
				Body: []byte("{\"message\":\"hello\"}"),
			},

			router: func(r *ivy.Router) {
				r.Method("MY_HTTP_METHOD", "/", func(c *ivy.Context) error {
					var x struct {
						Message string `json:"message"`
					}
					if err := c.ParseBodyInto(&x); err != nil {
						return err
					}
					return c.SendString(x.Message)
				})
			},

			response: Response{
				StatusCode: http.StatusOK,
				Body:       []byte("hello"),
			},
		},

		{
			name: "10. [Mount another ivy.Router] sending string",

			request: Request{
				Method: http.MethodGet,
				Route:  "/v2/",
				Headers: map[string]string{
					"Content-Type": "application/json",
				},
				Body: []byte("{\"message\":\"hello\"}"),
			},

			router: func(r *ivy.Router) {
				r2 := ivy.NewRouter()
				r2.Get("/", func(c *ivy.Context) error {
					var x struct {
						Message string `json:"message"`
					}
					if err := c.ParseBodyInto(&x); err != nil {
						return err
					}
					return c.SendString(x.Message)
				})

				r.Mount("/v2", r2)
			},

			response: Response{
				StatusCode: http.StatusOK,
				Body:       []byte("hello"),
			},
		},

		{
			name: "11. [Mount another http.Handler] sending string",

			request: Request{
				Method: http.MethodGet,
				Route:  "/v2",
				Headers: map[string]string{
					"Content-Type": "application/json",
				},
				Body: []byte("{\"message\":\"hello\"}"),
			},

			router: func(r *ivy.Router) {
				r.Mount("/v2", http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
					w.Write([]byte("hello world"))
				}))
			},

			response: Response{
				StatusCode: http.StatusOK,
				Body:       []byte("hello world"),
			},
		},

		{
			name: "12. [primitive http.HandleFunc] send string",

			request: Request{
				Method: http.MethodGet,
				Route:  "/v2",
				Headers: map[string]string{
					"Content-Type": "application/json",
				},
				Body: []byte("{\"message\":\"hello\"}"),
			},

			router: func(r *ivy.Router) {
				r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
					w.Write([]byte("ok"))
				})
			},

			response: Response{
				StatusCode: http.StatusOK,
				Body:       []byte("ok"),
			},
		},

		{
			name: "13. [primitive http.Handle] send string",

			request: Request{
				Method: http.MethodGet,
				Route:  "/v2",
				Headers: map[string]string{
					"Content-Type": "application/json",
				},
			},

			router: func(r *ivy.Router) {
				r.Handle("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.Write([]byte("ok"))
				}))
			},

			response: Response{
				StatusCode: http.StatusOK,
				Body:       []byte("ok"),
			},
		},

		{
			name: "14. [middleware propagation] from parent to child router",

			request: Request{
				Method: http.MethodGet,
				Route:  "/v2/test",
				Headers: map[string]string{
					"Content-Type": "application/json",
				},
			},

			router: func(r *ivy.Router) {
				r.Use(func(c *ivy.Context) error {
					c.KV.Set("hello", "world")
					return c.Next()
				})

				r2 := ivy.NewRouter()
				r2.Get("/test", func(c *ivy.Context) error {
					return c.SendString(c.KV.Get("hello").(string))
				})

				r.Mount("/v2", r2)
			},

			response: Response{
				StatusCode: http.StatusOK,
				// Body:       []byte("{\"hello\":\"world\"}"),
				Body: []byte("world"),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := ivy.NewRouter()
			if tt.errorHandler != nil {
				r.ErrorHandler = tt.errorHandler
			}

			tt.router(r)

			s := httptest.NewServer(r)
			defer s.Close()

			req := httptest.NewRequest(tt.request.Method, s.URL+tt.request.Route, bytes.NewReader(tt.request.Body))

			for k, v := range tt.request.Headers {
				req.Header.Add(k, v)
			}

			w := httptest.NewRecorder()

			r.ServeHTTP(w, req)

			res := w.Result()
			defer res.Body.Close()

			if res.StatusCode != tt.response.StatusCode {
				t.Errorf("expected status code\n\t got: %d\n\twant: %d\n", res.StatusCode, tt.response.StatusCode)
			}

			data, err := io.ReadAll(res.Body)
			if err != nil {
				t.Errorf("expected error to be nil got %v", err)
			}

			if string(data) != string(tt.response.Body) {
				t.Errorf("expected response body\n\t got: %q\n\twant: %q\n", data, tt.response.Body)
			}

			if tt.response.cookie != nil {
				if fmt.Sprintf("%+v", res.Cookies()) != fmt.Sprintf("%+v", tt.response.cookie) {
					t.Errorf("expected cookie body\n\t got: %+v\n\twant: %+v\n", res.Cookies(), tt.response.cookie)
				}
			}
		})
	}
}

func TestMiddlewareChaining(t *testing.T) {
	t.Run("final handler calling Next() should not panic", func(t *testing.T) {
		r := ivy.NewRouter()

		r.Get("/", func(c *ivy.Context) error {
			c.SendString("OK")
			return c.Next() // Final handler calling Next() - should be safe
		})

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		res := w.Result()
		defer res.Body.Close()

		if res.StatusCode != http.StatusOK {
			t.Errorf("status code: got %d, want %d", res.StatusCode, http.StatusOK)
		}

		data, _ := io.ReadAll(res.Body)
		if string(data) != "OK" {
			t.Errorf("body: got %q, want %q", data, "OK")
		}
	})

	t.Run("middleware chain execution order", func(t *testing.T) {
		r := ivy.NewRouter()

		var order []int

		r.Use(func(c *ivy.Context) error {
			order = append(order, 1)
			return c.Next()
		})

		r.Use(func(c *ivy.Context) error {
			order = append(order, 2)
			return c.Next()
		})

		r.Get("/",
			func(c *ivy.Context) error {
				order = append(order, 3)
				return c.Next()
			},
			func(c *ivy.Context) error {
				order = append(order, 4)
				return c.SendString("OK")
			},
		)

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		expected := []int{1, 2, 3, 4}
		if len(order) != len(expected) {
			t.Errorf("execution order length: got %d, want %d", len(order), len(expected))
		}
		for i, v := range expected {
			if order[i] != v {
				t.Errorf("execution order[%d]: got %d, want %d", i, order[i], v)
			}
		}
	})

	t.Run("middleware can short-circuit chain", func(t *testing.T) {
		r := ivy.NewRouter()

		handlerCalled := false

		r.Use(func(c *ivy.Context) error {
			return c.SendString("blocked") // Don't call Next()
		})

		r.Get("/", func(c *ivy.Context) error {
			handlerCalled = true
			return c.SendString("OK")
		})

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		res := w.Result()
		defer res.Body.Close()

		if handlerCalled {
			t.Error("handler should not have been called")
		}

		data, _ := io.ReadAll(res.Body)
		if string(data) != "blocked" {
			t.Errorf("body: got %q, want %q", data, "blocked")
		}
	})

	t.Run("all handlers calling Next() at end of chain", func(t *testing.T) {
		r := ivy.NewRouter()

		r.Use(func(c *ivy.Context) error {
			c.KV.Set("mw1", true)
			return c.Next()
		})

		r.Use(func(c *ivy.Context) error {
			c.KV.Set("mw2", true)
			return c.Next()
		})

		r.Get("/",
			func(c *ivy.Context) error {
				c.KV.Set("h1", true)
				return c.Next()
			},
			func(c *ivy.Context) error {
				c.KV.Set("h2", true)
				return c.Next()
			},
			func(c *ivy.Context) error {
				c.KV.Set("h3", true)
				c.SendString("done")
				return c.Next()
			},
		)

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		res := w.Result()
		defer res.Body.Close()

		if res.StatusCode != http.StatusOK {
			t.Errorf("status code: got %d, want %d", res.StatusCode, http.StatusOK)
		}

		data, _ := io.ReadAll(res.Body)
		if string(data) != "done" {
			t.Errorf("body: got %q, want %q", data, "done")
		}
	})
}
