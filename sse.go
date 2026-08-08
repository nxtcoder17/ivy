package ivy

import (
	"fmt"
	"time"
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
	payload := fmt.Sprintf("event: %s\nid: %d\ndata: %s\n\n", eventType, time.Now().UnixMilli(), data)
	_, err := c.response.Write([]byte(payload))
	if err != nil {
		return err
	}
	c.Flush()
	return nil
}
