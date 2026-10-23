package service

import (
	"context"
	"shieldgate/internal/domain"
)

type (
	APIKeyRepository interface {
		Create(ctx context.Context, key *domain.APIKey) error
		GetByKeyHash(ctx context.Context, keyHash string) (*domain.APIKey, error)
		GetByServiceID(ctx context.Context, serviceID int) ([]domain.APIKey, error)
		GetByID(ctx context.Context, id int) (*domain.APIKey, error)
		SetActive(ctx context.Context, id int, isActive bool) error
		Delete(ctx context.Context, id int) error
	}

	ServiceRepository interface {
		Create(ctx context.Context, svc *domain.Service) error
		GetByID(ctx context.Context, id int) (*domain.Service, error)
		GetByPathPrefix(ctx context.Context, pathPrefix string) (*domain.Service, error)
		GetByUserID(ctx context.Context, userID int) ([]domain.Service, error)
		Update(ctx context.Context, svc *domain.Service) error
		Delete(ctx context.Context, id int) error
	}

	UserRepository interface {
		Create(ctx context.Context, user *domain.User) error
		GetByID(ctx context.Context, id int) (*domain.User, error)
		GetByEmail(ctx context.Context, email string) (*domain.User, error)
		Update(ctx context.Context, user *domain.User) error
		Delete(ctx context.Context, id int) error
	}
	JwtManager interface {
		GenerateToken(userID string) (string, error)
		ParseToken(tokenString string) (string, error)
	}
	RateLimiter interface {
		Allow(ctx context.Context, key string, limit int) (bool, error)
	}
)
