package sse

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/nxtcoder17/ivy"
)

// TestSendEvent_Frame verifies the exact wire format SendEvent produces for the
// common frame shapes.
func TestSendEvent_Frame(t *testing.T) {
	tests := []struct {
		name      string
		eventType string
		data      []byte
		want      string
	}{
		{
			name:      "simple message",
			eventType: "message",
			data:      []byte("hello"),
			want:      "event: message\ndata: hello\n\n",
		},
		{
			name:      "multiline data becomes multiple data lines",
			eventType: "message",
			data:      []byte("first\nsecond"),
			want:      "event: message\ndata: first\ndata: second\n\n",
		},
		{
			name:      "CRLF line endings are normalized",
			eventType: "message",
			data:      []byte("first\r\nsecond"),
			want:      "event: message\ndata: first\ndata: second\n\n",
		},
		{
			name:      "empty data",
			eventType: "message",
			data:      nil,
			want:      "event: message\ndata: \n\n",
		},
		{
			name:      "empty event type omits the event field",
			eventType: "",
			data:      []byte("hello"),
			want:      "data: hello\n\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := ivy.NewRouter()
			r.Get("/events", New(), func(c *ivy.Context) error {
				stream, err := Stream(c)
				if err != nil {
					return err
				}
				return stream.SendEvent(tt.eventType, tt.data)
			})

			rec := serve(t, r, "/events")

			if got := rec.Body.String(); got != tt.want {
				t.Errorf("frame: got %q, want %q", got, tt.want)
			}
			if !rec.Flushed {
				t.Error("expected the response to be flushed after SendEvent")
			}
		})
	}
}

// TestSendEvent_InvalidEventType verifies that event types containing newlines
// or carriage returns are rejected before anything is written.
func TestSendEvent_InvalidEventType(t *testing.T) {
	for _, tt := range []string{"bad\nname", "bad\rname"} {
		t.Run(fmt.Sprintf("%q", tt), func(t *testing.T) {
			var gotErr error

			r := ivy.NewRouter()
			r.ErrorHandler = func(c *ivy.Context, err error) { gotErr = err }
			r.Get("/events", New(), func(c *ivy.Context) error {
				stream, err := Stream(c)
				if err != nil {
					return err
				}
				return stream.SendEvent(tt, []byte("hello"))
			})

			rec := serve(t, r, "/events")

			if gotErr == nil {
				t.Fatal("expected an error for invalid event type")
			}
			if got, want := gotErr.Error(), "sse: event type must not contain newlines"; got != want {
				t.Errorf("error: got %q, want %q", got, want)
			}
			if body := rec.Body.String(); body != "" {
				t.Errorf("nothing should be written for invalid event type, got %q", body)
			}
		})
	}
}

// TestSendEvent_NilStream verifies that calling SendEvent on a nil stream fails
// with ErrNoStream, even when the caller discards the error from Stream.
func TestSendEvent_NilStream(t *testing.T) {
	var s *EventStream

	err := s.SendEvent("message", []byte("hello"))

	if !errors.Is(err, ErrNoStream) {
		t.Fatalf("expected ErrNoStream, got %v", err)
	}
}

// TestStream_WithoutMiddleware verifies that the stream accessor on a route
// without sse.New() returns a nil stream and ErrNoStream.
func TestStream_WithoutMiddleware(t *testing.T) {
	var gotStream *EventStream
	var gotErr error

	r := ivy.NewRouter()
	r.Get("/events", func(c *ivy.Context) error {
		gotStream, gotErr = Stream(c)
		return nil
	})

	rec := serve(t, r, "/events")

	if rec.Code != http.StatusOK {
		t.Fatalf("status: got %d, want 200", rec.Code)
	}
	if gotStream != nil {
		t.Errorf("expected nil stream, got %v", gotStream)
	}
	if !errors.Is(gotErr, ErrNoStream) {
		t.Errorf("expected ErrNoStream, got %v", gotErr)
	}
}

// failingWriter is an http.ResponseWriter whose Write always fails, simulating
// a client that went away mid-stream.
type failingWriter struct {
	header http.Header
	err    error
}

func (w *failingWriter) Header() http.Header       { return w.header }
func (w *failingWriter) Write([]byte) (int, error) { return 0, w.err }
func (w *failingWriter) WriteHeader(int)           {}

// TestSendEvent_WriteErrorPropagates verifies that a failed write surfaces as
// an error instead of being silently swallowed.
func TestSendEvent_WriteErrorPropagates(t *testing.T) {
	wantErr := errors.New("client disconnected")
	var gotErr error

	r := ivy.NewRouter()
	r.ErrorHandler = func(c *ivy.Context, err error) { gotErr = err }
	r.Get("/events", New(), func(c *ivy.Context) error {
		stream, err := Stream(c)
		if err != nil {
			return err
		}
		return stream.SendEvent("message", []byte("hello"))
	})

	req := httptest.NewRequest(http.MethodGet, "/events", nil)
	r.ServeHTTP(&failingWriter{header: http.Header{}, err: wantErr}, req)

	if !errors.Is(gotErr, wantErr) {
		t.Errorf("expected the write error to propagate, got %v", gotErr)
	}
}

// TestSendEvent_Concurrent verifies that SendEvent is safe for concurrent use:
// interleaved writes from multiple goroutines must not corrupt or drop frames.
func TestSendEvent_Concurrent(t *testing.T) {
	const goroutines, eventsPerGoroutine = 10, 50

	r := ivy.NewRouter()
	r.Get("/events", New(), func(c *ivy.Context) error {
		stream, err := Stream(c)
		if err != nil {
			return err
		}

		var wg sync.WaitGroup
		for i := 0; i < goroutines; i++ {
			wg.Add(1)
			go func(n int) {
				defer wg.Done()
				for j := 0; j < eventsPerGoroutine; j++ {
					if err := stream.SendEvent("message", []byte(fmt.Sprintf("%d-%d", n, j))); err != nil {
						t.Errorf("SendEvent: %v", err)
						return
					}
				}
			}(i)
		}
		wg.Wait()
		return nil
	})

	rec := serve(t, r, "/events")

	body := rec.Body.String()
	if got, want := bytes.Count([]byte(body), []byte("event: message\n")), goroutines*eventsPerGoroutine; got != want {
		t.Errorf("frame count: got %d, want %d", got, want)
	}

	for i := 0; i < goroutines; i++ {
		for j := 0; j < eventsPerGoroutine; j++ {
			if !strings.Contains(body, fmt.Sprintf("data: %d-%d\n", i, j)) {
				t.Errorf("missing frame data %d-%d", i, j)
			}
		}
	}
}
