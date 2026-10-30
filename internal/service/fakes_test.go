package service

import (
	"context"

	"github.com/iluha481/shieldgate/internal/domain"
)

type stubUserRepo struct {
	user          *domain.User
	createErr     error
	getByIDErr    error
	getByEmailErr error
	created       *domain.User
}

func (r *stubUserRepo) Create(_ context.Context, user *domain.User) error {
	r.created = user
	return r.createErr
}

func (r *stubUserRepo) GetByID(_ context.Context, _ int) (*domain.User, error) {
	return r.user, r.getByIDErr
}

func (r *stubUserRepo) GetByEmail(_ context.Context, _ string) (*domain.User, error) {
	return r.user, r.getByEmailErr
}

func (r *stubUserRepo) Update(_ context.Context, _ *domain.User) error { return nil }

func (r *stubUserRepo) Delete(_ context.Context, _ int) error { return nil }

type stubServiceRepo struct {
	svc          *domain.Service
	services     []domain.Service
	createErr    error
	getByIDErr   error
	getByPathErr error
	getByUserErr error
	created      *domain.Service
}

func (r *stubServiceRepo) Create(_ context.Context, svc *domain.Service) error {
	r.created = svc
	return r.createErr
}

func (r *stubServiceRepo) GetByID(_ context.Context, _ int) (*domain.Service, error) {
	return r.svc, r.getByIDErr
}

func (r *stubServiceRepo) GetByPathPrefix(_ context.Context, _ string) (*domain.Service, error) {
	return r.svc, r.getByPathErr
}

func (r *stubServiceRepo) GetByUserID(_ context.Context, _ int) ([]domain.Service, error) {
	return r.services, r.getByUserErr
}

func (r *stubServiceRepo) Update(_ context.Context, _ *domain.Service) error { return nil }

func (r *stubServiceRepo) Delete(_ context.Context, _ int) error { return nil }

type stubAPIKeyRepo struct {
	key             *domain.APIKey
	keys            []domain.APIKey
	createErr       error
	getByHashErr    error
	getByServiceErr error
	getByIDErr      error
	setActiveErr    error
	deleteErr       error

	created        *domain.APIKey
	gotKeyHash     string
	activeID       int
	activeValue    bool
	setActiveCalls int
}

func (r *stubAPIKeyRepo) Create(_ context.Context, key *domain.APIKey) error {
	r.created = key
	return r.createErr
}

func (r *stubAPIKeyRepo) GetByKeyHash(_ context.Context, keyHash string) (*domain.APIKey, error) {
	r.gotKeyHash = keyHash
	return r.key, r.getByHashErr
}

func (r *stubAPIKeyRepo) GetByServiceID(_ context.Context, _ int) ([]domain.APIKey, error) {
	return r.keys, r.getByServiceErr
}

func (r *stubAPIKeyRepo) GetByID(_ context.Context, _ int) (*domain.APIKey, error) {
	return r.key, r.getByIDErr
}

func (r *stubAPIKeyRepo) SetActive(_ context.Context, id int, isActive bool) error {
	r.activeID = id
	r.activeValue = isActive
	r.setActiveCalls++
	return r.setActiveErr
}

func (r *stubAPIKeyRepo) Delete(_ context.Context, _ int) error { return r.deleteErr }

type stubLimiter struct {
	allowed  bool
	err      error
	gotKey   string
	gotLimit int
	calls    int
}

func (l *stubLimiter) Allow(_ context.Context, key string, limit int) (bool, error) {
	l.calls++
	l.gotKey = key
	l.gotLimit = limit
	return l.allowed, l.err
}

type stubJWT struct {
	token    string
	genErr   error
	parseErr error
	gotID    string
}

func (j *stubJWT) GenerateToken(userID string) (string, error) {
	j.gotID = userID
	return j.token, j.genErr
}

func (j *stubJWT) ParseToken(_ string) (string, error) {
	return "", j.parseErr
}
