package logger

import (
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/nxtcoder17/ivy"
)

// newTestRouter returns a router with the logger middleware installed, and a
// buffer capturing the logs (at Debug level, so both log lines are emitted).
func newTestRouter(t *testing.T, opts ...Options) (*ivy.Router, *strings.Builder) {
	t.Helper()

	var logs strings.Builder
	oldLogger := ivy.Logger
	ivy.Logger = slog.New(slog.NewTextHandler(&logs, &slog.HandlerOptions{Level: slog.LevelDebug}))
	t.Cleanup(func() { ivy.Logger = oldLogger })

	r := ivy.NewRouter()
	r.Use(New(opts...))
	return r, &logs
}

func TestMiddleware_LogsRequestAndResponse(t *testing.T) {
	r, logs := newTestRouter(t)
	r.Get("/hello", func(c *ivy.Context) error {
		return c.SendString("world")
	})

	req := httptest.NewRequest(http.MethodGet, "/hello", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	out := logs.String()
	if !strings.Contains(out, "❯❯ GET /hello") {
		t.Errorf("missing request log line, got: %q", out)
	}
	if !strings.Contains(out, "❮❮ 200 GET /hello") {
		t.Errorf("missing response log line with status, got: %q", out)
	}
}

func TestMiddleware_LogsHandlerStatus(t *testing.T) {
	r, logs := newTestRouter(t)
	r.Get("/not-found", func(c *ivy.Context) error {
		return c.Status(http.StatusNotFound).SendString("nope")
	})

	req := httptest.NewRequest(http.MethodGet, "/not-found", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}

	if !strings.Contains(logs.String(), "❮❮ 404 GET /not-found") {
		t.Errorf("expected 404 in response log, got: %q", logs.String())
	}
}

func TestMiddleware_RouteFilterSkipsLogging(t *testing.T) {
	r, logs := newTestRouter(t, Options{RouteFilter: func(path string) bool {
		return path != "/skip"
	}})
	r.Get("/skip", func(c *ivy.Context) error {
		return c.SendString("ok")
	})

	req := httptest.NewRequest(http.MethodGet, "/skip", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if got := logs.String(); got != "" {
		t.Errorf("expected no logs for filtered route, got: %q", got)
	}
}

func TestMiddleware_StripsQueryByDefault(t *testing.T) {
	// default: query string is stripped from the logged route
	r, logs := newTestRouter(t)
	r.Get("/search", func(c *ivy.Context) error {
		return c.SendString("ok")
	})

	req := httptest.NewRequest(http.MethodGet, "/search?q=hello&page=2", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	out := logs.String()
	if !strings.Contains(out, "❯❯ GET /search") {
		t.Errorf("expected query-stripped request log, got: %q", out)
	}
	if strings.Contains(out, "q=hello") {
		t.Errorf("query string should be stripped by default, got: %q", out)
	}
}

func TestMiddleware_ShowQueryEnabled(t *testing.T) {
	r, logs := newTestRouter(t, Options{ShowQuery: ivy.Ptr(true)})
	r.Get("/search", func(c *ivy.Context) error {
		return c.SendString("ok")
	})

	req := httptest.NewRequest(http.MethodGet, "/search?q=hello", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if !strings.Contains(logs.String(), "❯❯ GET /search?q=hello") {
		t.Errorf("expected full query string in request log, got: %q", logs.String())
	}
}
