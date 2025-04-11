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

type HTTPError struct {
	StatusCode int
	Message    string
}

// Error implements error.
func (h *HTTPError) Error() string {
	panic("unimplemented")
}

var _ error = (*HTTPError)(nil)
