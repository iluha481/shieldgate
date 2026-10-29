package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"

	"github.com/iluha481/shieldgate/internal/domain"
)

var (
	ErrRouteNotFound = errors.New("route not found for this path")
	ErrInvalidKey    = errors.New("invalid or inactive api key")
	ErrRateExceeded  = errors.New("rate limit exceeded (429)")
)

type GatewayService struct {
	serviceRepo ServiceRepository
	apiKeyRepo  APIKeyRepository
	limiter     RateLimiter
}

func NewGatewayService(
	sRepo ServiceRepository,
	kRepo APIKeyRepository,
	limiter RateLimiter,
) *GatewayService {
	return &GatewayService{
		serviceRepo: sRepo,
		apiKeyRepo:  kRepo,
		limiter:     limiter,
	}
}

type ResolutionResult struct {
	Service *domain.Service
	APIKey  *domain.APIKey
}

func (s *GatewayService) HandleIncomingRequest(
	ctx context.Context,
	path string,
	rawAPIKey string,
) (*ResolutionResult, error) {
	prefix := extractPrefix(path)
	if prefix == "" {
		return nil, ErrRouteNotFound
	}

	svc, err := s.serviceRepo.GetByPathPrefix(ctx, prefix)
	if err != nil {
		if errors.Is(err, domain.ErrServiceNotFound) {
			return nil, ErrRouteNotFound
		}
		return nil, fmt.Errorf("failed to get service: %w", err)
	}

	if rawAPIKey == "" {
		return nil, ErrInvalidKey
	}

	hash := sha256.Sum256([]byte(rawAPIKey))
	keyHash := hex.EncodeToString(hash[:])

	apiKey, err := s.apiKeyRepo.GetByKeyHash(ctx, keyHash)
	if err != nil {
		return nil, ErrInvalidKey
	}

	if apiKey.ServiceID != svc.ID {
		return nil, ErrInvalidKey
	}

	limiterKey := fmt.Sprintf("apikey:%d", apiKey.ID)
	allowed, err := s.limiter.Allow(ctx, limiterKey, svc.RateLimitRPS)
	if err != nil {
		return nil, fmt.Errorf("rate limiter error: %w", err)
	}

	if !allowed {
		return nil, ErrRateExceeded
	}

	return &ResolutionResult{
		Service: svc,
		APIKey:  apiKey,
	}, nil
}

func extractPrefix(path string) string {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) == 0 || parts[0] == "" {
		return ""
	}
	return "/" + parts[0]
}
