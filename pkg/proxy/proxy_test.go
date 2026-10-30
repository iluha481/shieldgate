package proxy

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestForward_RewritesPathAndForwards(t *testing.T) {
	var gotPath, gotRealIP, gotForwardedFor string

	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotRealIP = r.Header.Get("X-Real-IP")
		gotForwardedFor = r.Header.Get("X-Forwarded-For")
		_, _ = io.WriteString(w, "upstream-ok")
	}))
	defer upstream.Close()

	req := httptest.NewRequest(http.MethodGet, "/api/users/1?active=true", nil)
	req.RemoteAddr = "203.0.113.9:4321"
	rec := httptest.NewRecorder()

	if err := NewProxyManager().Forward(rec, req, upstream.URL, "/api"); err != nil {
		t.Fatalf("Forward() error = %v", err)
	}

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if body := rec.Body.String(); body != "upstream-ok" {
		t.Errorf("body = %q, want %q", body, "upstream-ok")
	}
	if gotPath != "/users/1" {
		t.Errorf("upstream path = %q, want %q", gotPath, "/users/1")
	}
	if gotRealIP != "203.0.113.9" {
		t.Errorf("X-Real-IP = %q, want %q", gotRealIP, "203.0.113.9")
	}
	if gotForwardedFor != "203.0.113.9" {
		t.Errorf("X-Forwarded-For = %q, want %q", gotForwardedFor, "203.0.113.9")
	}
}

func TestForward_InvalidTargetURL(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/x", nil)

	err := NewProxyManager().Forward(rec, req, "://bad-url", "/api")
	if err == nil {
		t.Fatal("Forward() expected error for invalid target URL")
	}
}

func TestForward_UpstreamUnreachable(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/x", nil)

	// Port 1 is not expected to be listening.
	if err := NewProxyManager().Forward(rec, req, "http://127.0.0.1:1", "/api"); err != nil {
		t.Fatalf("Forward() error = %v", err)
	}

	if rec.Code != http.StatusBadGateway {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusBadGateway)
	}
}
