package v1

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/iluha481/shieldgate/internal/controller/restapi/middleware"
	"github.com/iluha481/shieldgate/internal/domain"
	"github.com/iluha481/shieldgate/internal/service"
	"github.com/iluha481/shieldgate/pkg/jwt"
)

// withUserID runs the request through the auth middleware so the controller
// can read the user ID from the request context.
func withUserID(t *testing.T, req *http.Request, id string) *http.Request {
	t.Helper()

	manager := jwt.New("test-secret", time.Hour)
	token, err := manager.GenerateToken(id)
	if err != nil {
		t.Fatalf("GenerateToken() error = %v", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)

	var authed *http.Request
	middleware.Auth(manager, http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		authed = r
	})).ServeHTTP(httptest.NewRecorder(), req)

	if authed == nil {
		t.Fatal("auth middleware did not call the next handler")
	}
	return authed
}

func decodeBody(t *testing.T, body []byte) map[string]any {
	t.Helper()
	var out map[string]any
	if err := json.Unmarshal(body, &out); err != nil {
		t.Fatalf("failed to decode response body %q: %v", body, err)
	}
	return out
}

func TestAuthController_Register(t *testing.T) {
	repo := &fakeUserRepo{}
	ctrl := NewAuthController(service.NewAuthService(repo, &fakeJWT{}))

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register",
		bytes.NewBufferString(`{"email":"user@example.com","password":"password123"}`))
	rec := httptest.NewRecorder()

	ctrl.Register(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d; body: %s", rec.Code, http.StatusCreated, rec.Body)
	}

	body := decodeBody(t, rec.Body.Bytes())
	if body["email"] != "user@example.com" {
		t.Errorf("email = %v, want user@example.com", body["email"])
	}
	if _, ok := body["password_hash"]; ok {
		t.Error("response must not expose the password hash")
	}
	if repo.created == nil {
		t.Error("user was not passed to the repository")
	}
}

func TestAuthController_Register_Errors(t *testing.T) {
	tests := []struct {
		name     string
		body     string
		repoErr  error
		wantCode int
	}{
		{name: "invalid json", body: `{`, wantCode: http.StatusBadRequest},
		{name: "duplicate email", body: `{"email":"a@b.c","password":"pass"}`, repoErr: domain.ErrUserAlreadyExists, wantCode: http.StatusConflict},
		{name: "internal error", body: `{"email":"a@b.c","password":"pass"}`, repoErr: errInternal, wantCode: http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := NewAuthController(service.NewAuthService(&fakeUserRepo{createErr: tt.repoErr}, &fakeJWT{}))
			req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewBufferString(tt.body))
			rec := httptest.NewRecorder()

			ctrl.Register(rec, req)

			if rec.Code != tt.wantCode {
				t.Errorf("status = %d, want %d", rec.Code, tt.wantCode)
			}
		})
	}
}

func TestAuthController_Login(t *testing.T) {
	hash, err := bcryptHash("password123")
	if err != nil {
		t.Fatalf("bcryptHash() error = %v", err)
	}

	repo := &fakeUserRepo{user: &domain.User{ID: 1, Email: "user@example.com", PasswordHash: hash}}
	ctrl := NewAuthController(service.NewAuthService(repo, &fakeJWT{token: "token-value"}))

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login",
		bytes.NewBufferString(`{"email":"user@example.com","password":"password123"}`))
	rec := httptest.NewRecorder()

	ctrl.Login(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body: %s", rec.Code, http.StatusOK, rec.Body)
	}
	if body := decodeBody(t, rec.Body.Bytes()); body["token"] != "token-value" {
		t.Errorf("token = %v, want token-value", body["token"])
	}
}

func TestAuthController_Login_Errors(t *testing.T) {
	ctrl := NewAuthController(service.NewAuthService(&fakeUserRepo{getErr: errInternal}, &fakeJWT{}))

	t.Run("invalid credentials", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login",
			bytes.NewBufferString(`{"email":"a@b.c","password":"pass"}`))
		rec := httptest.NewRecorder()

		ctrl.Login(rec, req)

		if rec.Code != http.StatusUnauthorized {
			t.Errorf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
		}
	})

	t.Run("invalid json", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewBufferString(`{`))
		rec := httptest.NewRecorder()

		ctrl.Login(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
		}
	})
}
