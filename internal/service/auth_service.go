package service

import (
	"context"
	"fmt"
	"shieldgate/internal/domain"
	"strconv"

	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	userRepo UserRepository
	jwt      JwtManager
}

func NewAuthService(userRepo UserRepository, jwt JwtManager) *AuthService {
	return &AuthService{
		userRepo: userRepo,
		jwt:      jwt,
	}
}

func (s *AuthService) Register(ctx context.Context, email, password string) (*domain.User, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("AuthService - Register - bcrypt.GenerateFromPassword: %w", err)
	}

	user := domain.User{
		Email:        email,
		PasswordHash: string(hash),
	}

	err = s.userRepo.Create(ctx, &user)
	if err != nil {
		return nil, fmt.Errorf("AuthService - Register - userRepo.Create: %w", err)
	}

	return &user, nil
}

// Login -.
func (s *AuthService) Login(ctx context.Context, email, password string) (string, error) {
	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return "", domain.ErrUserNotFound
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		return "", domain.ErrUserNotFound
	}

	token, err := s.jwt.GenerateToken(strconv.Itoa(user.ID))
	if err != nil {
		return "", fmt.Errorf("AuthService - Login - s.jwt.GenerateToken: %w", err)
	}

	return token, nil
}

// GetUser -.
func (s *AuthService) GetUser(ctx context.Context, userID int) (*domain.User, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("AuthService - GetUser - s.userRepo.GetByID: %w", err)
	}

	return user, nil
}
