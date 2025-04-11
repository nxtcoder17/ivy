package ivy

type (
	ErrorHandler func(c *Context, err error)
)

func Ptr[T any](v T) *T {
	return &v
}
