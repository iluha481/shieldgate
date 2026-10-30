package v1

import (
	"context"

	"github.com/iluha481/shieldgate/internal/domain"
)

type fakeServiceRepo struct {
	svc          *domain.Service
	svcByPath    *domain.Service
	services     []domain.Service
	createErr    error
	getByIDErr   error
	getByPathErr error
	getByUserErr error
	created      *domain.Service
}

func (r *fakeServiceRepo) Create(_ context.Context, svc *domain.Service) error {
	r.created = svc
	return r.createErr
}

func (r *fakeServiceRepo) GetByID(_ context.Context, _ int) (*domain.Service, error) {
	return r.svc, r.getByIDErr
}

func (r *fakeServiceRepo) GetByPathPrefix(_ context.Context, _ string) (*domain.Service, error) {
	if r.svcByPath != nil {
		return r.svcByPath, r.getByPathErr
	}
	return r.svc, r.getByPathErr
}

func (r *fakeServiceRepo) GetByUserID(_ context.Context, _ int) ([]domain.Service, error) {
	return r.services, r.getByUserErr
}

func (r *fakeServiceRepo) Update(_ context.Context, _ *domain.Service) error { return nil }

func (r *fakeServiceRepo) Delete(_ context.Context, _ int) error { return nil }

type fakeAPIKeyRepo struct {
	key       *domain.APIKey
	createErr error
	getErr    error
	setErr    error
	created   *domain.APIKey
}

func (r *fakeAPIKeyRepo) Create(_ context.Context, key *domain.APIKey) error {
	r.created = key
	return r.createErr
}

func (r *fakeAPIKeyRepo) GetByKeyHash(_ context.Context, _ string) (*domain.APIKey, error) {
	return r.key, r.getErr
}

func (r *fakeAPIKeyRepo) GetByServiceID(_ context.Context, _ int) ([]domain.APIKey, error) {
	return nil, r.getErr
}

func (r *fakeAPIKeyRepo) GetByID(_ context.Context, _ int) (*domain.APIKey, error) {
	return r.key, r.getErr
}

func (r *fakeAPIKeyRepo) SetActive(_ context.Context, _ int, _ bool) error { return r.setErr }

func (r *fakeAPIKeyRepo) Delete(_ context.Context, _ int) error { return nil }

type fakeUserRepo struct {
	user      *domain.User
	createErr error
	getErr    error
	created   *domain.User
}

func (r *fakeUserRepo) Create(_ context.Context, user *domain.User) error {
	r.created = user
	return r.createErr
}

func (r *fakeUserRepo) GetByID(_ context.Context, _ int) (*domain.User, error) {
	return r.user, r.getErr
}

func (r *fakeUserRepo) GetByEmail(_ context.Context, _ string) (*domain.User, error) {
	return r.user, r.getErr
}

func (r *fakeUserRepo) Update(_ context.Context, _ *domain.User) error { return nil }

func (r *fakeUserRepo) Delete(_ context.Context, _ int) error { return nil }

type fakeJWT struct {
	token string
	err   error
}

func (j *fakeJWT) GenerateToken(_ string) (string, error) { return j.token, j.err }

func (j *fakeJWT) ParseToken(_ string) (string, error) { return "", j.err }

type fakeLimiter struct {
	allowed bool
	err     error
}

func (l *fakeLimiter) Allow(_ context.Context, _ string, _ int) (bool, error) {
	return l.allowed, l.err
}
