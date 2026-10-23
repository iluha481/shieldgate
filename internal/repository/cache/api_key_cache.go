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

type CachedAPIKeyRepository struct {
	pgRepo service.APIKeyRepository
	rdb    *redis.Client
	ttl    time.Duration
}

func NewCachedAPIKeyRepository(
	pgRepo service.APIKeyRepository,
	rdb *redis.Client,
	ttl time.Duration,
) *CachedAPIKeyRepository {
	return &CachedAPIKeyRepository{
		pgRepo: pgRepo,
		rdb:    rdb,
		ttl:    ttl,
	}
}

func (r *CachedAPIKeyRepository) GetByKeyHash(ctx context.Context, keyHash string) (*domain.APIKey, error) {
	cacheKey := fmt.Sprintf("cache:apikey:%s", keyHash)

	val, err := r.rdb.Get(ctx, cacheKey).Result()
	if err == nil {
		var key domain.APIKey
		if err := json.Unmarshal([]byte(val), &key); err == nil {
			return &key, nil
		}
	}

	key, err := r.pgRepo.GetByKeyHash(ctx, keyHash)
	if err != nil {
		return nil, err
	}

	jitter := time.Duration(rand.Intn(30)) * time.Second
	if data, err := json.Marshal(key); err == nil {
		if err := r.rdb.Set(ctx, cacheKey, data, r.ttl+jitter).Err(); err != nil {
			slog.Warn("failed to cache api key", "error", err)
		}
	}

	return key, nil
}

func (r *CachedAPIKeyRepository) SetActive(ctx context.Context, id int, isActive bool) error {
	key, err := r.pgRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if err := r.pgRepo.SetActive(ctx, id, isActive); err != nil {
		return err
	}
	cacheKey := fmt.Sprintf("cache:apikey:%s", key.KeyHash)
	r.rdb.Del(ctx, cacheKey)
	return nil
}
func (r *CachedAPIKeyRepository) Create(ctx context.Context, key *domain.APIKey) error {
	return r.pgRepo.Create(ctx, key)
}

func (r *CachedAPIKeyRepository) GetByServiceID(ctx context.Context, serviceID int) ([]domain.APIKey, error) {
	return r.pgRepo.GetByServiceID(ctx, serviceID)
}

func (r *CachedAPIKeyRepository) Delete(ctx context.Context, id int) error {
	key, err := r.pgRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if err := r.pgRepo.Delete(ctx, id); err != nil {
		return err
	}
	cacheKey := fmt.Sprintf("cache:apikey:%s", key.KeyHash)
	r.rdb.Del(ctx, cacheKey)
	return nil
}

func (r *CachedAPIKeyRepository) GetByID(ctx context.Context, id int) (*domain.APIKey, error) {
	return r.pgRepo.GetByID(ctx, id)
}
