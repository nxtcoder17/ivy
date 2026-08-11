package sse

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/nxtcoder17/ivy"
)

// streamKey is the unexported KV key under which ivy stores the stream, so it
// cannot collide with any user-set keys.
type streamKey struct{}

// EventStream is the request-scoped SSE stream created by New and stored in
// the request's KV store. It serializes writes so concurrent SendEvent calls
// (e.g. from a broadcast hub) cannot interleave frames.
//
// Only valid on routes where sse.New() is registered; using it elsewhere makes
// Stream return ErrNoStream.
type EventStream struct {
	c  *ivy.Context
	mu sync.Mutex
}

// ErrNoStream is returned by Stream when sse.New() was not registered on the
// route.
var ErrNoStream = errors.New("sse: no stream for this request; register sse.New() on the route")

// Stream returns the EventStream created by New for the current request, or
// ErrNoStream when the middleware is not registered on the route.
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
func Stream(c *ivy.Context) (*EventStream, error) {
	s, ok := c.KV.Get(streamKey{}).(*EventStream)
	if !ok {
		return nil, ErrNoStream
	}
	return s, nil
}

// SendEvent writes a standard SSE event frame and flushes it to the client.
// An empty eventType omits the event: field. Safe for concurrent use.
func (s *EventStream) SendEvent(eventType string, data []byte) error {
	if s == nil {
		return ErrNoStream
	}

	if strings.ContainsAny(eventType, "\n\r") {
		return errors.New("sse: event type must not contain newlines")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	var buf bytes.Buffer
	if eventType != "" {
		fmt.Fprintf(&buf, "event: %s\n", eventType)
	}

	for _, line := range bytes.Split(data, []byte("\n")) {
		line = bytes.TrimSuffix(line, []byte("\r"))
		fmt.Fprintf(&buf, "data: %s\n", line)
	}

	buf.WriteByte('\n')

	if _, err := s.c.Write(buf.Bytes()); err != nil {
		return err
	}
	s.c.Flush()
	return nil
}
