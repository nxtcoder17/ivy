package sse

import (
	"net/http"
	"strings"

	"github.com/nxtcoder17/ivy"
)

type Options struct {
	// AcceptHeaderValidator overrides the built-in Accept header check.
	// Pass an always-true function to skip validation entirely.
	AcceptHeaderValidator func(accept string) bool
}

// New returns the SSE middleware. It validates the client's Accept header,
// sets the SSE response headers, and creates the request's stream in the KV
// store, retrievable with sse.Stream(c).
//
// The middleware is required for sending events: register it on every route
// that streams events.
//
// Example:
//
//	r.Get("/events", sse.New(), func(c *ivy.Context) error {
//		stream, err := sse.Stream(c)
//		if err != nil {
//			return err
//		}
//		return stream.SendEvent("message", []byte("hello"))
//	})
func New(opts ...Options) ivy.Handler {
	var o Options
	if len(opts) > 0 {
		o = opts[0]
	}

	validate := o.AcceptHeaderValidator
	if validate == nil {
		validate = defaultAcceptHeaderValidator
	}

	return func(c *ivy.Context) error {
		if !validate(c.GetHeaders().Get("Accept")) {
			return ivy.NewHTTPError(http.StatusNotAcceptable, "client must accept text/event-stream")
		}

		setHeaders(c)

		// Eagerly create the stream so the KV write happens before the handler
		// runs, keeping the KV store read-only during request handling.
		c.KV.Set(streamKey{}, &EventStream{c: c})

		return c.Next()
	}
}

// setHeaders sets the response headers required for an SSE stream.
func setHeaders(c *ivy.Context) {
	c.SetHeader("Content-Type", "text/event-stream")
	c.SetHeader("Cache-Control", "no-cache")
	c.SetHeader("Connection", "keep-alive")

	// INFO: Disable response buffering (required for nginx streaming).
	c.SetHeader("X-Accel-Buffering", "no")
}

// defaultAcceptHeaderValidator validates client's Accept header for text/event-stream
// for common usual ways, clients use SSE.
// For Anything Fancy, use Options.AcceptHeaderValidator
func defaultAcceptHeaderValidator(accept string) bool {
	accept = strings.ToLower(strings.TrimSpace(accept))

	switch accept {
	case "", "*/*", "text/*", "text/event-stream", "text/event-stream; charset=utf-8":
		return true
	}

	return strings.Contains(accept, "text/event-stream")
}
