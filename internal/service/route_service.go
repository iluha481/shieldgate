package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"shieldgate/internal/domain"
)

var (
	ErrUnauthorizedServiceAccess = errors.New("you do not have access to this service")
)

type RouteService struct {
	serviceRepo ServiceRepository
	apiKeyRepo  APIKeyRepository
}

func NewRouteService(sRepo ServiceRepository, kRepo APIKeyRepository) *RouteService {
	return &RouteService{
		serviceRepo: sRepo,
		apiKeyRepo:  kRepo,
	}
}

// 1. Создание нового защищаемого сервиса
func (s *RouteService) CreateService(
	ctx context.Context,
	userID int,
	name, pathPrefix, targetURL string,
	rateLimitRPS, burstCapacity int,
) (*domain.Service, error) {
	// Базовые дефолтные значения, если не передали
	if rateLimitRPS <= 0 {
		rateLimitRPS = 10
	}
	if burstCapacity <= 0 {
		burstCapacity = 20
	}
	svc := &domain.Service{
		UserID:        userID,
		Name:          name,
		PathPrefix:    pathPrefix,
		TargetURL:     targetURL,
		RateLimitRPS:  rateLimitRPS,
		BurstCapacity: burstCapacity,
		IsActive:      true,
	}
	if err := s.serviceRepo.Create(ctx, svc); err != nil {
		return nil, fmt.Errorf("RouteService - CreateService: %w", err)
	}
	return svc, nil
}

func (s *RouteService) GetUserServices(ctx context.Context, userID int) ([]domain.Service, error) {
	services, err := s.serviceRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("RouteService - GetUserServices: %w", err)
	}
	return services, nil
}

func (s *RouteService) GenerateAPIKey(
	ctx context.Context,
	userID, serviceID int,
	keyName string,
) (plainKey string, apiKey *domain.APIKey, err error) {
	svc, err := s.serviceRepo.GetByID(ctx, serviceID)
	if err != nil {
		return "", nil, fmt.Errorf("RouteService - GenerateAPIKey - GetByID: %w", err)
	}
	if svc.UserID != userID {
		return "", nil, ErrUnauthorizedServiceAccess
	}
	randomBytes := make([]byte, 24)
	if _, err := rand.Read(randomBytes); err != nil {
		return "", nil, fmt.Errorf("RouteService - GenerateAPIKey - rand.Read: %w", err)
	}
	rawSecret := hex.EncodeToString(randomBytes)
	plainKey = fmt.Sprintf("sg_live_%s", rawSecret)
	keyPrefix := plainKey[:12]
	hash := sha256.Sum256([]byte(plainKey))
	keyHash := hex.EncodeToString(hash[:])
	apiKey = &domain.APIKey{
		ServiceID: serviceID,
		Name:      keyName,
		KeyHash:   keyHash,
		KeyPrefix: keyPrefix,
		IsActive:  true,
	}
	if err := s.apiKeyRepo.Create(ctx, apiKey); err != nil {
		return "", nil, fmt.Errorf("RouteService - GenerateAPIKey - Create: %w", err)
	}
	return plainKey, apiKey, nil
}

func (s *RouteService) RevokeAPIKey(ctx context.Context, userID, serviceID, keyID int) error {
	svc, err := s.serviceRepo.GetByID(ctx, serviceID)
	if err != nil {
		return fmt.Errorf("RouteService - RevokeAPIKey: %w", err)
	}
	if svc.UserID != userID {
		return ErrUnauthorizedServiceAccess
	}
	return s.apiKeyRepo.SetActive(ctx, keyID, false)
}
