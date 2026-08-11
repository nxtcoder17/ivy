package requestid

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/nxtcoder17/ivy"
)

func TestMiddleware_SetsRequestID(t *testing.T) {
	r := ivy.NewRouter()
	r.Use(New(func() string { return "test-id-123" }))
	r.Get("/ping", func(c *ivy.Context) error {
		return c.SendString(c.GetRequestID())
	})

	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if got, want := rec.Body.String(), "test-id-123"; got != want {
		t.Errorf("request id: got %q, want %q", got, want)
	}
}

func TestMiddleware_DefaultGenerator(t *testing.T) {
	r := ivy.NewRouter()
	r.Use(New())
	r.Get("/ping", func(c *ivy.Context) error {
		return c.SendString(c.GetRequestID())
	})

	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if got := rec.Body.String(); got == "" {
		t.Error("expected a generated request id, got empty")
	}
}

func TestMiddleware_PreservesExistingRequestID(t *testing.T) {
	r := ivy.NewRouter()
	r.Get("/ping",
		func(c *ivy.Context) error {
			c.SetRequestID("preset")
			return c.Next()
		},
		New(func() string { return "should-not-be-used" }),
		func(c *ivy.Context) error {
			return c.SendString(c.GetRequestID())
		},
	)

	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if got, want := rec.Body.String(), "preset"; got != want {
		t.Errorf("request id: got %q, want %q", got, want)
	}
}
