package queryparams

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/nxtcoder17/ivy"
)

func TestRequired_MissingParam(t *testing.T) {
	r := ivy.NewRouter()
	r.Get("/search", Required("q"), func(c *ivy.Context) error {
		return c.SendString("ok")
	})

	req := httptest.NewRequest(http.MethodGet, "/search", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), `missing query-param "q"`) {
		t.Errorf("expected missing-param error in body, got %q", rec.Body.String())
	}
}

func TestRequired_ParamPresent(t *testing.T) {
	r := ivy.NewRouter()
	r.Get("/search", Required("q"), func(c *ivy.Context) error {
		return c.SendString("ok")
	})

	req := httptest.NewRequest(http.MethodGet, "/search?q=hello", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
	if got := rec.Body.String(); got != "ok" {
		t.Errorf("expected body 'ok', got %q", got)
	}
}

func TestRequired_MultipleParams(t *testing.T) {
	r := ivy.NewRouter()
	r.Get("/search", Required("q", "page"), func(c *ivy.Context) error {
		return c.SendString("ok")
	})

	// one of the required params missing -> 400, naming the missing one
	req := httptest.NewRequest(http.MethodGet, "/search?q=hello", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400 when one param is missing, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), `missing query-param "page"`) {
		t.Errorf("expected missing 'page' error, got %q", rec.Body.String())
	}

	// all required params present -> 200
	req = httptest.NewRequest(http.MethodGet, "/search?q=hello&page=1", nil)
	rec = httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200 when all params are present, got %d", rec.Code)
	}
}

func TestRequired_EmptyValueCountsAsPresent(t *testing.T) {
	r := ivy.NewRouter()
	r.Get("/search", Required("q"), func(c *ivy.Context) error {
		return c.SendString("ok")
	})

	req := httptest.NewRequest(http.MethodGet, "/search?q=", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200 when param is present but empty, got %d", rec.Code)
	}
}
