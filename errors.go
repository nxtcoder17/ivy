package ivy

type joinErrors interface {
	Unwrap() []error
}

func ErrorFormatJSON(err error) map[string]any {
	je, ok := err.(joinErrors)
	if ok {
		errors := je.Unwrap()
		errs := make([]string, 0, len(errors))
		for i := range errors {
			errs = append(errs, errors[i].Error())
		}
		return map[string]any{"errors": errs}
	}

	return map[string]any{"errors": err}
}

type HTTPError interface {
	error
	Code() int
	Message() string
}

type httpError struct {
	code    int
	message string
}

// Error implements error.
func (h *httpError) Error() string {
	return h.message
}

func (h *httpError) Code() int {
	return h.code
}

func (h *httpError) Message() string {
	return h.message
}

var _ HTTPError = (*httpError)(nil)

func NewHTTPError(code int, msg string) HTTPError {
	return &httpError{code, msg}
}
