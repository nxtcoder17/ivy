package sse

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/nxtcoder17/ivy"
)

// serve issues a GET request against r and returns the recorded response.
func serve(t *testing.T, r *ivy.Router, path string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, path, nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	return rec
}

func TestDefaultAcceptHeaderValidator(t *testing.T) {
	tests := []struct {
		accept string
		want   bool
	}{
		{accept: "", want: true}, // no Accept header: accepts everything
		{accept: "*/*", want: true},
		{accept: "text/*", want: true},
		{accept: "text/event-stream", want: true},
		{accept: "text/event-stream; charset=utf-8", want: true},
		{accept: "text/html, text/event-stream", want: true},
		{accept: "text/event-stream, text/html", want: true},
		{accept: "  text/event-stream  ", want: true}, // whitespace is trimmed
		{accept: "TEXT/EVENT-STREAM", want: true},     // matching is case-insensitive
		{accept: "text/event-stream; charset=UTF-8", want: true},
		{accept: "text/html", want: false},
		{accept: "application/json", want: false},
		{accept: "application/*", want: false},
	}

	for _, tt := range tests {
		if got := defaultAcceptHeaderValidator(tt.accept); got != tt.want {
			t.Errorf("defaultAcceptHeaderValidator(%q) = %v, want %v", tt.accept, got, tt.want)
		}
	}
}

// TestNew_StashesStreamEagerly verifies that New creates the stream before the
// handler runs: it is present in the KV store and the SSE headers are already
// set, even when the handler never sends an event.
func TestNew_StashesStreamEagerly(t *testing.T) {
	var first, second *EventStream
	var kvVal any
	var headers http.Header

	r := ivy.NewRouter()
	r.Get("/events", New(), func(c *ivy.Context) error {
		var err error
		first, err = Stream(c)
		if err != nil {
			return err
		}
		second, err = Stream(c)
		if err != nil {
			return err
		}
		kvVal, _ = c.KV.Lookup(streamKey{})
		headers = c.ResponseWriter().Header()
		return nil
	})

	rec := serve(t, r, "/events")

	if rec.Code != http.StatusOK {
		t.Fatalf("status: got %d, want 200", rec.Code)
	}
	if first == nil {
		t.Fatal("expected a stream to be stashed by New")
	}
	if first != second {
		t.Error("Stream(c) should return the same stream on every call")
	}
	if kvStream, ok := kvVal.(*EventStream); !ok || kvStream != first {
		t.Error("KV should hold the same stream under the sse key")
	}
	for k, want := range map[string]string{
		"Content-Type":      "text/event-stream",
		"Cache-Control":     "no-cache",
		"Connection":        "keep-alive",
		"X-Accel-Buffering": "no",
	} {
		if got := headers.Get(k); got != want {
			t.Errorf("%s before handler runs: got %q, want %q", k, got, want)
		}
	}
}

// TestNew_AcceptHeaderValidatorOverride verifies that Options.AcceptHeaderValidator
// replaces the built-in check, and that an always-true validator disables
// validation entirely.
func TestNew_AcceptHeaderValidatorOverride(t *testing.T) {
	custom := func(accept string) bool { return accept == "application/json" }

	r := ivy.NewRouter()
	r.Get("/events", New(Options{AcceptHeaderValidator: custom}), func(c *ivy.Context) error {
		return c.SendString("ok")
	})

	req := httptest.NewRequest(http.MethodGet, "/events", nil)
	req.Header.Set("Accept", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("custom validator accepting application/json: got %d, want 200", rec.Code)
	}

	req = httptest.NewRequest(http.MethodGet, "/events", nil)
	req.Header.Set("Accept", "text/event-stream")
	rec = httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusNotAcceptable {
		t.Errorf("custom validator rejecting text/event-stream: got %d, want 406", rec.Code)
	}

	r2 := ivy.NewRouter()
	r2.Get("/events", New(Options{
		AcceptHeaderValidator: func(string) bool { return true },
	}), func(c *ivy.Context) error {
		return c.SendString("ok")
	})

	req = httptest.NewRequest(http.MethodGet, "/events", nil)
	req.Header.Set("Accept", "application/json")
	rec = httptest.NewRecorder()
	r2.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("validation disabled: got %d, want 200", rec.Code)
	}
}

// TestNew_RejectsNonSSEAccept verifies the middleware short-circuits with 406
// and the route handler never runs.
func TestNew_RejectsNonSSEAccept(t *testing.T) {
	handlerRan := false

	r := ivy.NewRouter()
	r.Get("/events", New(), func(c *ivy.Context) error {
		handlerRan = true
		return nil
	})

	req := httptest.NewRequest(http.MethodGet, "/events", nil)
	req.Header.Set("Accept", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotAcceptable {
		t.Errorf("status: got %d, want 406", rec.Code)
	}
	if handlerRan {
		t.Error("handler must not run when Accept header is rejected")
	}
}

// TestStream_AvailableFromMountedSubRouter verifies the stream survives the
// router mounting boundary: the KV store is re-attached from the request
// context, so a handler on a mounted sub-router can use the stream created by
// New on the parent router.
func TestStream_AvailableFromMountedSubRouter(t *testing.T) {
	sub := ivy.NewRouter()
	sub.Get("/events", func(c *ivy.Context) error {
		stream, err := Stream(c)
		if err != nil {
			return err
		}
		return stream.SendEvent("message", []byte("hello"))
	})

	parent := ivy.NewRouter()
	parent.Use(New())
	parent.Mount("/v2", sub)

	rec := serve(t, parent, "/v2/events")

	if rec.Code != http.StatusOK {
		t.Fatalf("status: got %d, want 200", rec.Code)
	}
	if got, want := rec.Body.String(), "event: message\ndata: hello\n\n"; got != want {
		t.Errorf("body: got %q, want %q", got, want)
	}
	if got, want := rec.Header().Get("Content-Type"), "text/event-stream"; got != want {
		t.Errorf("Content-Type: got %q, want %q", got, want)
	}
}
