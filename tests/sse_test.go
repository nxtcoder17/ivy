package ivy_test

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/nxtcoder17/ivy"
)

func TestSSEvent(t *testing.T) {
	r := ivy.NewRouter()
	r.Get("/events", func(c *ivy.Context) error {
		c.SetSSEHeaders()
		return c.SendSSEvent("message", []byte(`{"message":"hello"}`))
	})

	req := httptest.NewRequest(http.MethodGet, "/events", nil)
	res := httptest.NewRecorder()

	r.ServeHTTP(res, req)

	received := res.Result()
	defer received.Body.Close()

	if got, want := received.Header.Get("Content-Type"), "text/event-stream"; got != want {
		t.Errorf("Content-Type: got %q, want %q", got, want)
	}
	if got, want := received.Header.Get("Cache-Control"), "no-cache"; got != want {
		t.Errorf("Cache-Control: got %q, want %q", got, want)
	}
	if got, want := received.Header.Get("Connection"), "keep-alive"; got != want {
		t.Errorf("Connection: got %q, want %q", got, want)
	}
	if got, want := received.Header.Get("X-Accel-Buffering"), "no"; got != want {
		t.Errorf("X-Accel-Buffering: got %q, want %q", got, want)
	}
	if !res.Flushed {
		t.Error("expected response to be flushed")
	}

	body, err := io.ReadAll(received.Body)
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Contains(body, []byte("event: message\n")) {
		t.Errorf("event type missing from body: %q", body)
	}
	if !bytes.Contains(body, []byte("data: {\"message\":\"hello\"}\n\n")) {
		t.Errorf("data missing from body: %q", body)
	}
}
