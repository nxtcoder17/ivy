package ivy

import "net/http"

type (
	ErrorHandler func(err error, w http.ResponseWriter, r *http.Request)
)
