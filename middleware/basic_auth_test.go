package middleware

import (
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/nxtcoder17/ivy"
)

func TestBasicAuth_ValidCredentials(t *testing.T) {
	r := ivy.NewRouter()
	r.Use(BasicAuth("Test Realm", map[string]string{
		"admin": "secret",
	}))
	r.Get("/protected", func(c *ivy.Context) error {
		return c.SendString("ok")
	})

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.SetBasicAuth("admin", "secret")
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}
	if rec.Body.String() != "ok" {
		t.Errorf("expected body 'ok', got %q", rec.Body.String())
	}
}

func TestBasicAuth_InvalidPassword(t *testing.T) {
	r := ivy.NewRouter()
	r.Use(BasicAuth("Test Realm", map[string]string{
		"admin": "secret",
	}))
	r.Get("/protected", func(c *ivy.Context) error {
		return c.SendString("ok")
	})

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.SetBasicAuth("admin", "wrongpassword")
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", rec.Code)
	}
	if rec.Header().Get("WWW-Authenticate") != `Basic realm="Test Realm"` {
		t.Errorf("expected WWW-Authenticate header, got %q", rec.Header().Get("WWW-Authenticate"))
	}
}

func TestBasicAuth_InvalidUsername(t *testing.T) {
	r := ivy.NewRouter()
	r.Use(BasicAuth("Test Realm", map[string]string{
		"admin": "secret",
	}))
	r.Get("/protected", func(c *ivy.Context) error {
		return c.SendString("ok")
	})

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.SetBasicAuth("unknown", "secret")
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", rec.Code)
	}
}

func TestBasicAuth_NoCredentials(t *testing.T) {
	r := ivy.NewRouter()
	r.Use(BasicAuth("Test Realm", map[string]string{
		"admin": "secret",
	}))
	r.Get("/protected", func(c *ivy.Context) error {
		return c.SendString("ok")
	})

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", rec.Code)
	}
	if rec.Header().Get("WWW-Authenticate") != `Basic realm="Test Realm"` {
		t.Errorf("expected WWW-Authenticate header, got %q", rec.Header().Get("WWW-Authenticate"))
	}
}

func TestBasicAuth_MalformedHeader(t *testing.T) {
	r := ivy.NewRouter()
	r.Use(BasicAuth("Test Realm", map[string]string{
		"admin": "secret",
	}))
	r.Get("/protected", func(c *ivy.Context) error {
		return c.SendString("ok")
	})

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte("malformed")))
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", rec.Code)
	}
}

func TestBasicAuth_MultipleUsers(t *testing.T) {
	r := ivy.NewRouter()
	r.Use(BasicAuth("Test Realm", map[string]string{
		"admin": "adminpass",
		"user":  "userpass",
	}))
	r.Get("/protected", func(c *ivy.Context) error {
		return c.SendString("ok")
	})

	tests := []struct {
		user string
		pass string
		want int
	}{
		{"admin", "adminpass", http.StatusOK},
		{"user", "userpass", http.StatusOK},
		{"admin", "userpass", http.StatusUnauthorized},
		{"user", "adminpass", http.StatusUnauthorized},
	}

	for _, tt := range tests {
		req := httptest.NewRequest(http.MethodGet, "/protected", nil)
		req.SetBasicAuth(tt.user, tt.pass)
		rec := httptest.NewRecorder()

		r.ServeHTTP(rec, req)

		if rec.Code != tt.want {
			t.Errorf("user=%s pass=%s: expected status %d, got %d", tt.user, tt.pass, tt.want, rec.Code)
		}
	}
}
