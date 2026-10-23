package repository

import (
	"context"
	"errors"
	"fmt"
	"shieldgate/internal/domain"
	"shieldgate/pkg/postgres"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type UserRepository struct {
	pg *postgres.DB
}

func NewUserRepository(pg *postgres.DB) *UserRepository {
	return &UserRepository{pg: pg}
}

func (r *UserRepository) Create(ctx context.Context, user *domain.User) error {
	sql := `
		INSERT INTO users (email, password_hash, created_at, updated_at)
		VALUES ($1, $2, NOW(), NOW())
		RETURNING id, created_at, updated_at
	`
	err := r.pg.Pool.QueryRow(ctx, sql, user.Email, user.PasswordHash).Scan(
		&user.ID,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return domain.ErrUserAlreadyExists
		}
		return fmt.Errorf("user_repo.Create: %w", err)
	}
	return nil
}

func (r *UserRepository) GetByID(ctx context.Context, id int) (*domain.User, error) {
	sql := `
		SELECT id, email, password_hash, created_at, updated_at
		FROM users
		WHERE id = $1
	`
	var user domain.User
	err := r.pg.Pool.QueryRow(ctx, sql, id).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrUserNotFound
		}
		return nil, fmt.Errorf("user_repo.GetByID: %w", err)
	}
	return &user, nil
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	sql := `
		SELECT id, email, password_hash, created_at, updated_at
		FROM users
		WHERE email = $1
	`
	var user domain.User
	err := r.pg.Pool.QueryRow(ctx, sql, email).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrUserNotFound
		}
		return nil, fmt.Errorf("user_repo.GetByEmail: %w", err)
	}
	return &user, nil
}

func (r *UserRepository) Update(ctx context.Context, user *domain.User) error {
	sql := `
		UPDATE users
		SET email = $1, password_hash = $2, updated_at = NOW()
		WHERE id = $3
	`
	tag, err := r.pg.Pool.Exec(ctx, sql, user.Email, user.PasswordHash, user.ID)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return domain.ErrUserAlreadyExists
		}
		return fmt.Errorf("user_repo.Update: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrUserNotFound
	}
	return nil
}

func (r *UserRepository) Delete(ctx context.Context, id int) error {
	sql := `DELETE FROM users WHERE id = $1`
	tag, err := r.pg.Pool.Exec(ctx, sql, id)
	if err != nil {
		return fmt.Errorf("user_repo.Delete: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrUserNotFound
	}
	return nil
}
