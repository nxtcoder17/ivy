package ivy

import "io"

func (c *Context) Body() io.ReadCloser {
	return c.request.Body
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
