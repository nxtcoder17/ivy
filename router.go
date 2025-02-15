package ivy

import (
	"encoding/json"
	"net/http"
	"strings"
)

type Router struct {
	mux         *http.ServeMux
	middlewares []Handler

	ErrorHandler ErrorHandler
}

var DefaultErrorHandler ErrorHandler = func(c *Context, err error) {
	http.Error(c.ResponseWriter(), err.Error(), http.StatusInternalServerError)
}

// Package level variables, it mainly just introduces a constraint
// that there will be same encoders across the package in an application lifecycle
var (
	// JSONEncoder defaults to [encoding/json.Marshal](https://pkg.go.dev/encoding/json#Marshal)
	JSONEncoder func(v any) ([]byte, error) = json.Marshal

	// JSONDecoder defaults to [encoding/json.Unmarshal](https://pkg.go.dev/encoding/json#Unmarshal)
	JSONDecoder func(b []byte, v any) error = json.Unmarshal
)

// ServeHTTP implements http.Handler.
func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.mux.ServeHTTP(w, req)
}

var _ http.Handler = (*Router)(nil)

type Handler func(c *Context) error

// ServeHTTP implements http.Handler.
func (hf Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	hf(NewContext(r, w))
}

var _ http.Handler = (Handler)(nil)

func (r *Router) chainHandlers(handlers ...Handler) http.HandlerFunc {
	allHandlers := make([]Handler, 0, len(r.middlewares)+len(handlers))
	allHandlers = append(allHandlers, r.middlewares...)
	allHandlers = append(allHandlers, handlers...)

	next := func(c *Context) error {
		if c.handlerIdx > len(allHandlers) {
			return nil
		}

		return allHandlers[c.handlerIdx](c)
	}

	return func(w http.ResponseWriter, req *http.Request) {
		ctx := NewContext(req, w)
		ctx.next = next

		if kv := req.Context().Value(ivyRequestCtxKey); kv != nil {
			switch v := kv.(type) {
			case *KV:
				ctx.KV = v
			}
		}

		if err := next(ctx); err != nil {
			if r.ErrorHandler != nil {
				r.ErrorHandler(ctx, err)
				return
			}
			DefaultErrorHandler(ctx, err)
		}
	}
}

func (r *Router) register(method string, path string, handlers ...Handler) {
	if handlers == nil {
		return
	}

	r.mux.HandleFunc(method+" "+path, r.chainHandlers(handlers...))
}

func (r *Router) Get(path string, handlers ...Handler) {
	r.register(http.MethodGet, path, handlers...)
}

func (r *Router) Post(path string, handlers ...Handler) {
	r.register(http.MethodPost, path, handlers...)
}

func (r *Router) Put(path string, handlers ...Handler) {
	r.register(http.MethodPut, path, handlers...)
}

func (r *Router) Delete(path string, handlers ...Handler) {
	r.register(http.MethodDelete, path, handlers...)
}

func (r *Router) Head(path string, handlers ...Handler) {
	r.register(http.MethodHead, path, handlers...)
}

func (r *Router) Method(method string, path string, handlers ...Handler) {
	r.register(method, path, handlers...)
}

func (r *Router) Use(handlers ...Handler) {
	r.middlewares = append(r.middlewares, handlers...)
}

// INFO: when mouting another router / http.Handler, we need to ensure that middlewares defined on router (r), are also applied on new handlers

func (r *Router) Mount(path string, h http.Handler) {
	if !strings.HasSuffix(path, "/") {
		path = path + "/"
	}

	if anotherRouter, ok := h.(*Router); ok {
		if anotherRouter.ErrorHandler == nil {
			anotherRouter.ErrorHandler = r.ErrorHandler
		}
	}

	r.mux.Handle(path, http.StripPrefix(path[:len(path)-1], r.chainHandlers(ToIvyHandler(h))))
	r.mux.Handle(path[:len(path)-1], http.StripPrefix(path[:len(path)-1], r.chainHandlers(ToIvyHandler(h))))
}

func (r *Router) HandleFunc(path string, handle http.HandlerFunc) {
	r.mux.HandleFunc(path, r.chainHandlers(ToIvyHandler(handle)))
}

func (r *Router) Handle(path string, handler http.Handler) {
	r.mux.Handle(path, r.chainHandlers(ToIvyHandler(handler)))
}

var _ http.Handler = (*Router)(nil)

// NewRouter() creates an ivy.Router with some defaults
func NewRouter() *Router {
	return &Router{
		mux:          http.NewServeMux(),
		middlewares:  nil,
		ErrorHandler: nil,
	}
}
