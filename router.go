package ivy

import (
	"log/slog"
	"net/http"

	"github.com/goccy/go-json"

	"github.com/go-chi/chi/v5"
)

type Router struct {
	mux *chi.Mux
	*options
}

// ServeHTTP implements http.Handler.
func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.mux.ServeHTTP(w, req)
}

var _ http.Handler = (*Router)(nil)

type Handler func(c *Context) error

// ServeHTTP implements http.Handler.
func (hf Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	hf(&Context{Context: r.Context(), request: r, response: w})
}

var _ http.Handler = (Handler)(nil)

func (r *Router) register(method string, path string, handlers ...Handler) {
	if handlers == nil {
		return
	}

	next := func(c *Context) error {
		// idx += 1
		// if idx == len(handlers) {
		// 	slog.Warn("middelwares Out-Of-Bounds | no response written", "method", method, "path", path, "len(handlers)", len(handlers))
		// 	return nil
		// }

		if c.handlerIdx > len(handlers) {
			slog.Warn("middelwares Out-Of-Bounds | no response written", "method", method, "path", path, "len(handlers)", len(handlers), "current.index", c.handlerIdx)
		}

		return handlers[c.handlerIdx](c)
	}

	r.mux.MethodFunc(method, path, func(w http.ResponseWriter, req *http.Request) {
		ctx := Context{
			Context: req.Context(),

			request:  req,
			response: w,

			jsonEncoder: r.JSONEncoder,
			jsonDecoder: r.JSONDecoder,

			handlerIdx: 0,
			next:       next,
		}

		if err := handlers[0](&ctx); err != nil {
			r.ErrorHandler(err, w, req)
		}
	})
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
	chi.RegisterMethod(method)
	r.register(method, path, handlers...)
}

func (r *Router) Mount(path string, h http.Handler) {
	r.mux.Mount(path, h)
}

func (r *Router) Use(handlers ...func(http.Handler) http.Handler) {
	r.mux.Use(handlers...)
}

func (r *Router) HandleFunc(path string, handle http.HandlerFunc) {
	r.mux.HandleFunc(path, handle)
}

func (r *Router) Handle(path string, handler http.Handler) {
	r.mux.Handle(path, handler)
}

var _ http.Handler = (*Router)(nil)

type options struct {
	// ErrorHandler handles error returned by an ivy.Handler
	ErrorHandler

	// JSONEncoder is for marshalling json
	JSONEncoder

	// JSONDecoder is for unmarshalling bytes into something
	JSONDecoder
}

func defaultOptions() *options {
	return &options{
		ErrorHandler: func(err error, w http.ResponseWriter, r *http.Request) {
			http.Error(w, err.Error(), 500)
		},
		JSONEncoder: json.Marshal,
		JSONDecoder: json.Unmarshal,
	}
}

type option func(o *options)

func WithErrorHandler(handler ErrorHandler) option {
	return func(o *options) {
		if handler != nil {
			o.ErrorHandler = handler
		}
	}
}

func WithJSONEncoder(encoder JSONEncoder) option {
	return func(o *options) {
		if encoder != nil {
			o.JSONEncoder = encoder
		}
	}
}

func WithJSONDecoder(decoder JSONDecoder) option {
	return func(o *options) {
		if decoder != nil {
			o.JSONDecoder = decoder
		}
	}
}

func NewRouter(opts ...option) *Router {
	r := chi.NewRouter()

	options := defaultOptions()
	for _, op := range opts {
		op(options)
	}

	return &Router{
		mux:     r,
		options: options,
	}
}
