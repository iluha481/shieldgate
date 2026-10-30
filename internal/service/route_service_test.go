package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"strings"
	"testing"

	"github.com/iluha481/shieldgate/internal/domain"
)

func TestRouteService_CreateService_Defaults(t *testing.T) {
	repo := &stubServiceRepo{}
	svc := NewRouteService(repo, &stubAPIKeyRepo{})

	created, err := svc.CreateService(context.Background(), 1, "My API", "/api", "http://upstream", 0, 0)
	if err != nil {
		t.Fatalf("CreateService() error = %v", err)
	}

	if created.RateLimitRPS != 10 {
		t.Errorf("RateLimitRPS = %d, want 10", created.RateLimitRPS)
	}
	if created.BurstCapacity != 20 {
		t.Errorf("BurstCapacity = %d, want 20", created.BurstCapacity)
	}
	if !created.IsActive {
		t.Error("IsActive = false, want true")
	}
	if created.UserID != 1 || created.Name != "My API" || created.PathPrefix != "/api" || created.TargetURL != "http://upstream" {
		t.Errorf("unexpected service: %+v", created)
	}
	if repo.created != created {
		t.Error("CreateService() did not pass the service to the repository")
	}
}

func TestRouteService_CreateService_CustomValues(t *testing.T) {
	created, err := NewRouteService(&stubServiceRepo{}, &stubAPIKeyRepo{}).
		CreateService(context.Background(), 1, "API", "/api", "http://upstream", 50, 100)
	if err != nil {
		t.Fatalf("CreateService() error = %v", err)
	}

	if created.RateLimitRPS != 50 || created.BurstCapacity != 100 {
		t.Errorf("got RPS=%d burst=%d, want 50/100", created.RateLimitRPS, created.BurstCapacity)
	}
}

func TestRouteService_CreateService_RepoError(t *testing.T) {
	repoErr := errors.New("duplicate prefix")
	_, err := NewRouteService(&stubServiceRepo{createErr: repoErr}, &stubAPIKeyRepo{}).
		CreateService(context.Background(), 1, "API", "/api", "http://upstream", 10, 20)

	if !errors.Is(err, repoErr) {
		t.Fatalf("CreateService() error = %v, want %v", err, repoErr)
	}
}

func TestRouteService_GetUserServices(t *testing.T) {
	want := []domain.Service{{ID: 1}, {ID: 2}}
	svc := NewRouteService(&stubServiceRepo{services: want}, &stubAPIKeyRepo{})

	got, err := svc.GetUserServices(context.Background(), 1)
	if err != nil {
		t.Fatalf("GetUserServices() error = %v", err)
	}
	if len(got) != len(want) {
		t.Fatalf("GetUserServices() returned %d services, want %d", len(got), len(want))
	}
}

func TestRouteService_GetUserServices_Error(t *testing.T) {
	repoErr := errors.New("db down")
	_, err := NewRouteService(&stubServiceRepo{getByUserErr: repoErr}, &stubAPIKeyRepo{}).
		GetUserServices(context.Background(), 1)

	if !errors.Is(err, repoErr) {
		t.Fatalf("GetUserServices() error = %v, want %v", err, repoErr)
	}
}

func TestRouteService_GenerateAPIKey(t *testing.T) {
	svcRepo := &stubServiceRepo{svc: &domain.Service{ID: 5, UserID: 1}}
	keyRepo := &stubAPIKeyRepo{}
	svc := NewRouteService(svcRepo, keyRepo)

	plain, key, err := svc.GenerateAPIKey(context.Background(), 1, 5, "ci-key")
	if err != nil {
		t.Fatalf("GenerateAPIKey() error = %v", err)
	}

	if !strings.HasPrefix(plain, "sg_live_") {
		t.Errorf("plain key = %q, want sg_live_ prefix", plain)
	}
	if key.ServiceID != 5 || key.Name != "ci-key" || !key.IsActive {
		t.Errorf("unexpected api key: %+v", key)
	}
	if key.KeyPrefix != plain[:12] {
		t.Errorf("KeyPrefix = %q, want %q", key.KeyPrefix, plain[:12])
	}

	sum := sha256.Sum256([]byte(plain))
	if key.KeyHash != hex.EncodeToString(sum[:]) {
		t.Error("KeyHash does not match SHA-256 of the plain key")
	}
	if keyRepo.created != key {
		t.Error("GenerateAPIKey() did not pass the key to the repository")
	}
}

func TestRouteService_GenerateAPIKey_Errors(t *testing.T) {
	repoErr := errors.New("db down")

	tests := []struct {
		name    string
		svcRepo *stubServiceRepo
		keyRepo *stubAPIKeyRepo
		userID  int
		wantErr error
	}{
		{
			name:    "service lookup failed",
			svcRepo: &stubServiceRepo{getByIDErr: repoErr},
			keyRepo: &stubAPIKeyRepo{},
			userID:  1,
		},
		{
			name:    "foreign service",
			svcRepo: &stubServiceRepo{svc: &domain.Service{ID: 5, UserID: 2}},
			keyRepo: &stubAPIKeyRepo{},
			userID:  1,
			wantErr: ErrUnauthorizedServiceAccess,
		},
		{
			name:    "key creation failed",
			svcRepo: &stubServiceRepo{svc: &domain.Service{ID: 5, UserID: 1}},
			keyRepo: &stubAPIKeyRepo{createErr: repoErr},
			userID:  1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, err := NewRouteService(tt.svcRepo, tt.keyRepo).
				GenerateAPIKey(context.Background(), tt.userID, 5, "key")
			if err == nil {
				t.Fatal("GenerateAPIKey() expected error")
			}
			if tt.wantErr != nil && !errors.Is(err, tt.wantErr) {
				t.Errorf("GenerateAPIKey() error = %v, want %v", err, tt.wantErr)
			}
		})
	}
}

func TestRouteService_RevokeAPIKey(t *testing.T) {
	keyRepo := &stubAPIKeyRepo{}
	svc := NewRouteService(&stubServiceRepo{svc: &domain.Service{ID: 5, UserID: 1}}, keyRepo)

	if err := svc.RevokeAPIKey(context.Background(), 1, 5, 99); err != nil {
		t.Fatalf("RevokeAPIKey() error = %v", err)
	}
	if keyRepo.setActiveCalls != 1 || keyRepo.activeID != 99 || keyRepo.activeValue {
		t.Errorf("SetActive() calls=%d id=%d value=%v, want one call with id 99 and false",
			keyRepo.setActiveCalls, keyRepo.activeID, keyRepo.activeValue)
	}
}

func TestRouteService_RevokeAPIKey_Errors(t *testing.T) {
	t.Run("service lookup failed", func(t *testing.T) {
		repoErr := errors.New("db down")
		err := NewRouteService(&stubServiceRepo{getByIDErr: repoErr}, &stubAPIKeyRepo{}).
			RevokeAPIKey(context.Background(), 1, 5, 99)
		if !errors.Is(err, repoErr) {
			t.Fatalf("RevokeAPIKey() error = %v, want %v", err, repoErr)
		}
	})

	t.Run("foreign service", func(t *testing.T) {
		err := NewRouteService(&stubServiceRepo{svc: &domain.Service{ID: 5, UserID: 2}}, &stubAPIKeyRepo{}).
			RevokeAPIKey(context.Background(), 1, 5, 99)
		if !errors.Is(err, ErrUnauthorizedServiceAccess) {
			t.Fatalf("RevokeAPIKey() error = %v, want ErrUnauthorizedServiceAccess", err)
		}
	})
}
