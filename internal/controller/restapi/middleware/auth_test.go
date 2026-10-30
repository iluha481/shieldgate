package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/iluha481/shieldgate/pkg/jwt"
)

func TestAuth_MissingHeader(t *testing.T) {
	handler := Auth(jwt.New("secret", time.Hour), okHandler(t))

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestAuth_InvalidHeaderFormat(t *testing.T) {
	handler := Auth(jwt.New("secret", time.Hour), okHandler(t))

	for _, header := range []string{"Token abc", "Bearer", "Bearer a b"} {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Authorization", header)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusUnauthorized {
			t.Errorf("header %q: status = %d, want %d", header, rec.Code, http.StatusUnauthorized)
		}
	}
}

func TestAuth_InvalidToken(t *testing.T) {
	handler := Auth(jwt.New("secret", time.Hour), okHandler(t))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer not-a-token")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestAuth_NonNumericSubject(t *testing.T) {
	manager := jwt.New("secret", time.Hour)
	token, err := manager.GenerateToken("not-a-number")
	if err != nil {
		t.Fatalf("GenerateToken() error = %v", err)
	}
	handler := Auth(manager, okHandler(t))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestAuth_ValidToken(t *testing.T) {
	manager := jwt.New("secret", time.Hour)
	token, err := manager.GenerateToken("42")
	if err != nil {
		t.Fatalf("GenerateToken() error = %v", err)
	}

	var gotUserID int
	var ok bool
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotUserID, ok = GetUserID(r.Context())
		w.WriteHeader(http.StatusNoContent)
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	Auth(manager, next).ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusNoContent)
	}
	if !ok || gotUserID != 42 {
		t.Errorf("GetUserID() = (%d, %v), want (42, true)", gotUserID, ok)
	}
}

func TestGetUserID_Missing(t *testing.T) {
	if _, ok := GetUserID(httptest.NewRequest(http.MethodGet, "/", nil).Context()); ok {
		t.Error("GetUserID() returned ok for a context without a user ID")
	}
}

func okHandler(t *testing.T) http.Handler {
	t.Helper()
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
}
