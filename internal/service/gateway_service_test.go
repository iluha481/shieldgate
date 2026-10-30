package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"testing"

	"github.com/iluha481/shieldgate/internal/domain"
)

func TestExtractPrefix(t *testing.T) {
	tests := []struct {
		path string
		want string
	}{
		{path: "/api/users", want: "/api"},
		{path: "api/users", want: "/api"},
		{path: "/api", want: "/api"},
		{path: "/", want: ""},
		{path: "", want: ""},
		{path: "///", want: ""},
	}

	for _, tt := range tests {
		if got := extractPrefix(tt.path); got != tt.want {
			t.Errorf("extractPrefix(%q) = %q, want %q", tt.path, got, tt.want)
		}
	}
}

func TestGatewayService_Success(t *testing.T) {
	const rawKey = "sg_live_test-key"
	svcRepo := &stubServiceRepo{svc: &domain.Service{ID: 3, PathPrefix: "/api", RateLimitRPS: 15}}
	keyRepo := &stubAPIKeyRepo{key: &domain.APIKey{ID: 7, ServiceID: 3}}
	limiter := &stubLimiter{allowed: true}
	svc := NewGatewayService(svcRepo, keyRepo, limiter)

	result, err := svc.HandleIncomingRequest(context.Background(), "/api/users", rawKey)
	if err != nil {
		t.Fatalf("HandleIncomingRequest() error = %v", err)
	}

	if result.Service != svcRepo.svc || result.APIKey != keyRepo.key {
		t.Errorf("unexpected result: %+v", result)
	}

	sum := sha256.Sum256([]byte(rawKey))
	if want := hex.EncodeToString(sum[:]); keyRepo.gotKeyHash != want {
		t.Errorf("GetByKeyHash() got %q, want %q", keyRepo.gotKeyHash, want)
	}
	if limiter.gotKey != "apikey:7" || limiter.gotLimit != 15 {
		t.Errorf("Allow() key=%q limit=%d, want apikey:7 / 15", limiter.gotKey, limiter.gotLimit)
	}
}

func TestGatewayService_RouteErrors(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		repoErr error
		wantErr error
	}{
		{
			name:    "empty path",
			path:    "/",
			wantErr: ErrRouteNotFound,
		},
		{
			name:    "service not found",
			path:    "/api/users",
			repoErr: domain.ErrServiceNotFound,
			wantErr: ErrRouteNotFound,
		},
		{
			name:    "repository failure",
			path:    "/api/users",
			repoErr: errors.New("db down"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := NewGatewayService(
				&stubServiceRepo{getByPathErr: tt.repoErr},
				&stubAPIKeyRepo{},
				&stubLimiter{allowed: true},
			)

			_, err := svc.HandleIncomingRequest(context.Background(), tt.path, "key")
			if err == nil {
				t.Fatal("HandleIncomingRequest() expected error")
			}
			if tt.wantErr != nil && !errors.Is(err, tt.wantErr) {
				t.Errorf("HandleIncomingRequest() error = %v, want %v", err, tt.wantErr)
			}
		})
	}
}

func TestGatewayService_KeyErrors(t *testing.T) {
	svcRepo := &stubServiceRepo{svc: &domain.Service{ID: 3}}

	t.Run("missing key", func(t *testing.T) {
		svc := NewGatewayService(svcRepo, &stubAPIKeyRepo{}, &stubLimiter{allowed: true})
		_, err := svc.HandleIncomingRequest(context.Background(), "/api/users", "")
		if !errors.Is(err, ErrInvalidKey) {
			t.Fatalf("error = %v, want ErrInvalidKey", err)
		}
	})

	t.Run("unknown key", func(t *testing.T) {
		svc := NewGatewayService(svcRepo, &stubAPIKeyRepo{getByHashErr: errors.New("not found")}, &stubLimiter{allowed: true})
		_, err := svc.HandleIncomingRequest(context.Background(), "/api/users", "key")
		if !errors.Is(err, ErrInvalidKey) {
			t.Fatalf("error = %v, want ErrInvalidKey", err)
		}
	})

	t.Run("key belongs to another service", func(t *testing.T) {
		keyRepo := &stubAPIKeyRepo{key: &domain.APIKey{ID: 1, ServiceID: 999}}
		svc := NewGatewayService(svcRepo, keyRepo, &stubLimiter{allowed: true})
		_, err := svc.HandleIncomingRequest(context.Background(), "/api/users", "key")
		if !errors.Is(err, ErrInvalidKey) {
			t.Fatalf("error = %v, want ErrInvalidKey", err)
		}
	})
}

func TestGatewayService_RateLimitErrors(t *testing.T) {
	svcRepo := &stubServiceRepo{svc: &domain.Service{ID: 3, RateLimitRPS: 10}}
	keyRepo := &stubAPIKeyRepo{key: &domain.APIKey{ID: 7, ServiceID: 3}}

	t.Run("limiter failure", func(t *testing.T) {
		limiterErr := errors.New("redis down")
		svc := NewGatewayService(svcRepo, keyRepo, &stubLimiter{err: limiterErr})

		_, err := svc.HandleIncomingRequest(context.Background(), "/api/users", "key")
		if !errors.Is(err, limiterErr) {
			t.Fatalf("error = %v, want %v", err, limiterErr)
		}
	})

	t.Run("limit exceeded", func(t *testing.T) {
		svc := NewGatewayService(svcRepo, keyRepo, &stubLimiter{allowed: false})

		_, err := svc.HandleIncomingRequest(context.Background(), "/api/users", "key")
		if !errors.Is(err, ErrRateExceeded) {
			t.Fatalf("error = %v, want ErrRateExceeded", err)
		}
	})
}
