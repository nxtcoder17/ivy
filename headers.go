package ivy

func (c *Context) SetRequestID(ID string) {
	c.request.Header.Set("X-Request-ID", ID)
}

func (c *Context) GetRequestID() string {
	return c.request.Header.Get("X-Request-ID")
}
