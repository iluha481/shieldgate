package v1

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/iluha481/shieldgate/internal/domain"
	"github.com/iluha481/shieldgate/internal/service"
	"github.com/iluha481/shieldgate/pkg/proxy"
)

func TestGatewayController_ErrorMapping(t *testing.T) {
	tests := []struct {
		name     string
		svcRepo  *fakeServiceRepo
		keyRepo  *fakeAPIKeyRepo
		limiter  *fakeLimiter
		apiKey   string
		wantCode int
	}{
		{
			name:     "route not found",
			svcRepo:  &fakeServiceRepo{getByPathErr: domain.ErrServiceNotFound},
			keyRepo:  &fakeAPIKeyRepo{},
			limiter:  &fakeLimiter{allowed: true},
			apiKey:   "key",
			wantCode: http.StatusNotFound,
		},
		{
			name:     "invalid key",
			svcRepo:  &fakeServiceRepo{svc: &domain.Service{ID: 1}},
			keyRepo:  &fakeAPIKeyRepo{},
			limiter:  &fakeLimiter{allowed: true},
			apiKey:   "",
			wantCode: http.StatusUnauthorized,
		},
		{
			name:     "rate exceeded",
			svcRepo:  &fakeServiceRepo{svc: &domain.Service{ID: 1}},
			keyRepo:  &fakeAPIKeyRepo{key: &domain.APIKey{ID: 1, ServiceID: 1}},
			limiter:  &fakeLimiter{allowed: false},
			apiKey:   "key",
			wantCode: http.StatusTooManyRequests,
		},
		{
			name:     "internal error",
			svcRepo:  &fakeServiceRepo{getByPathErr: errInternal},
			keyRepo:  &fakeAPIKeyRepo{},
			limiter:  &fakeLimiter{allowed: true},
			apiKey:   "key",
			wantCode: http.StatusInternalServerError,
		},
		{
			name:     "upstream unreachable",
			svcRepo:  &fakeServiceRepo{svc: &domain.Service{ID: 1, TargetURL: "://bad", PathPrefix: "/api"}},
			keyRepo:  &fakeAPIKeyRepo{key: &domain.APIKey{ID: 1, ServiceID: 1}},
			limiter:  &fakeLimiter{allowed: true},
			apiKey:   "key",
			wantCode: http.StatusBadGateway,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := NewGatewayController(
				service.NewGatewayService(tt.svcRepo, tt.keyRepo, tt.limiter),
				proxy.NewProxyManager(),
			)

			req := httptest.NewRequest(http.MethodGet, "/api/users", nil)
			req.Header.Set("X-API-Key", tt.apiKey)
			rec := httptest.NewRecorder()

			ctrl.ServeHTTP(rec, req)

			if rec.Code != tt.wantCode {
				t.Errorf("status = %d, want %d; body: %s", rec.Code, tt.wantCode, rec.Body)
			}
			if tt.wantCode == http.StatusTooManyRequests && rec.Header().Get("Retry-After") != "1" {
				t.Errorf("Retry-After = %q, want %q", rec.Header().Get("Retry-After"), "1")
			}
		})
	}
}

func TestGatewayController_ForwardsToUpstream(t *testing.T) {
	var gotPath, gotAPIKey string
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotAPIKey = r.Header.Get("X-API-Key")
		_, _ = io.WriteString(w, "upstream response")
	}))
	defer upstream.Close()

	ctrl := NewGatewayController(
		service.NewGatewayService(
			&fakeServiceRepo{svcByPath: &domain.Service{ID: 1, TargetURL: upstream.URL, PathPrefix: "/api"}},
			&fakeAPIKeyRepo{key: &domain.APIKey{ID: 1, ServiceID: 1}},
			&fakeLimiter{allowed: true},
		),
		proxy.NewProxyManager(),
	)

	req := httptest.NewRequest(http.MethodGet, "/api/users", nil)
	req.Header.Set("X-API-Key", "secret-key")
	rec := httptest.NewRecorder()

	ctrl.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body: %s", rec.Code, http.StatusOK, rec.Body)
	}
	if gotPath != "/users" {
		t.Errorf("upstream path = %q, want %q", gotPath, "/users")
	}
	if gotAPIKey != "secret-key" {
		t.Errorf("upstream X-API-Key = %q, want %q", gotAPIKey, "secret-key")
	}
	if rec.Body.String() != "upstream response" {
		t.Errorf("body = %q, want %q", rec.Body.String(), "upstream response")
	}
}
