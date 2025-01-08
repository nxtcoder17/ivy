package ivy

import "net/http"

type ErrorHandler func(err error, w http.ResponseWriter, r *http.Request)

type JSONEncoder func(v any) ([]byte, error)

type JSONDecoder func(b []byte, v any) error

type Middleware func(next http.Handler) http.Handler
