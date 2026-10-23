package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"math/rand"
	"time"

	"shieldgate/internal/domain"
	"shieldgate/internal/service"

	"github.com/redis/go-redis/v9"
)

type CachedServiceRepository struct {
	pgRepo service.ServiceRepository
	rdb    *redis.Client
	ttl    time.Duration
}

func NewCachedServiceRepository(
	pgRepo service.ServiceRepository,
	rdb *redis.Client,
	ttl time.Duration,
) *CachedServiceRepository {
	return &CachedServiceRepository{
		pgRepo: pgRepo,
		rdb:    rdb,
		ttl:    ttl,
	}
}

func (r *CachedServiceRepository) GetByPathPrefix(ctx context.Context, pathPrefix string) (*domain.Service, error) {
	cacheKey := fmt.Sprintf("cache:service:%s", pathPrefix)

	val, err := r.rdb.Get(ctx, cacheKey).Result()
	if err == nil {
		var svc domain.Service
		if err := json.Unmarshal([]byte(val), &svc); err == nil {
			return &svc, nil
		}
	}

	svc, err := r.pgRepo.GetByPathPrefix(ctx, pathPrefix)
	if err != nil {
		return nil, err
	}

	jitter := time.Duration(rand.Intn(30)) * time.Second
	if data, err := json.Marshal(svc); err == nil {
		if err := r.rdb.Set(ctx, cacheKey, data, r.ttl+jitter).Err(); err != nil {
			slog.Warn("failed to cache service", "error", err)
		}
	}

	return svc, nil
}

func (r *CachedServiceRepository) Create(ctx context.Context, svc *domain.Service) error {
	return r.pgRepo.Create(ctx, svc)
}

func (r *CachedServiceRepository) Update(ctx context.Context, svc *domain.Service) error {
	oldSvc, _ := r.pgRepo.GetByID(ctx, svc.ID)
	if err := r.pgRepo.Update(ctx, svc); err != nil {
		return err
	}
	if oldSvc != nil {
		r.rdb.Del(ctx, fmt.Sprintf("cache:service:%s", oldSvc.PathPrefix))
	}
	r.rdb.Del(ctx, fmt.Sprintf("cache:service:%s", svc.PathPrefix))
	return nil
}

func (r *CachedServiceRepository) Delete(ctx context.Context, id int) error {
	svc, err := r.pgRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if err := r.pgRepo.Delete(ctx, id); err != nil {
		return err
	}
	cacheKey := fmt.Sprintf("cache:service:%s", svc.PathPrefix)
	if err := r.rdb.Del(ctx, cacheKey).Err(); err != nil {
		slog.Error("failed to invalidate service cache on delete", "error", err)
	}
	return nil
}

func (r *CachedServiceRepository) GetByID(ctx context.Context, id int) (*domain.Service, error) {
	return r.pgRepo.GetByID(ctx, id)
}

func (r *CachedServiceRepository) GetByUserID(ctx context.Context, userID int) ([]domain.Service, error) {
	return r.pgRepo.GetByUserID(ctx, userID)
}
