package ivy

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
)

func (c *Context) SetSSEHeaders() {
	// INFO: SSE required headers
	c.response.Header().Set("Content-Type", "text/event-stream")
	c.response.Header().Set("Cache-Control", "no-cache")
	c.response.Header().Set("Connection", "keep-alive")

	// INFO: Disable response buffering (required for nginx streaming)
	c.response.Header().Set("X-Accel-Buffering", "no")
}

// SendSSEvent writes a standard SSE event frame
func (c *Context) SendSSEvent(eventType string, data []byte) error {
	if strings.ContainsAny(eventType, "\n\r") {
		return errors.New("sse: event type must not contain newlines")
	}

	var buf bytes.Buffer
	fmt.Fprintf(&buf, "event: %s\n", eventType)

	for _, line := range bytes.Split(data, []byte("\n")) {
		line = bytes.TrimSuffix(line, []byte("\r"))
		fmt.Fprintf(&buf, "data: %s\n", line)
	}

	buf.WriteByte('\n')

	if _, err := c.response.Write(buf.Bytes()); err != nil {
		return err
	}
	c.Flush()
	return nil
}
