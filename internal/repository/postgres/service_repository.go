package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/iluha481/shieldgate/internal/domain"
	"github.com/iluha481/shieldgate/pkg/postgres"
)

type ServiceRepository struct {
	pg *postgres.DB
}

func NewServiceRepository(pg *postgres.DB) *ServiceRepository {
	return &ServiceRepository{pg: pg}
}

func (r *ServiceRepository) Create(ctx context.Context, svc *domain.Service) error {
	sql := `
		INSERT INTO services (
			user_id, name, path_prefix, target_url, 
			rate_limit_rps, burst_capacity, is_active, 
			created_at, updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, NOW(), NOW())
		RETURNING id, created_at, updated_at
	`

	err := r.pg.Pool.QueryRow(ctx, sql,
		svc.UserID,
		svc.Name,
		svc.PathPrefix,
		svc.TargetURL,
		svc.RateLimitRPS,
		svc.BurstCapacity,
		svc.IsActive,
	).Scan(&svc.ID, &svc.CreatedAt, &svc.UpdatedAt)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return domain.ErrPathPrefixAlreadyExists
		}
		return fmt.Errorf("service_repo.Create: %w", err)
	}

	return nil
}

func (r *ServiceRepository) GetByID(ctx context.Context, id int) (*domain.Service, error) {
	sql := `
		SELECT id, user_id, name, path_prefix, target_url, rate_limit_rps, burst_capacity, is_active, created_at, updated_at
		FROM services
		WHERE id = $1
	`

	var s domain.Service
	err := r.pg.Pool.QueryRow(ctx, sql, id).Scan(
		&s.ID, &s.UserID, &s.Name, &s.PathPrefix, &s.TargetURL,
		&s.RateLimitRPS, &s.BurstCapacity, &s.IsActive, &s.CreatedAt, &s.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrServiceNotFound
		}
		return nil, fmt.Errorf("service_repo.GetByID: %w", err)
	}

	return &s, nil
}

func (r *ServiceRepository) GetByPathPrefix(ctx context.Context, pathPrefix string) (*domain.Service, error) {
	sql := `
		SELECT id, user_id, name, path_prefix, target_url, rate_limit_rps, burst_capacity, is_active, created_at, updated_at
		FROM services
		WHERE path_prefix = $1 AND is_active = true
	`

	var s domain.Service
	err := r.pg.Pool.QueryRow(ctx, sql, pathPrefix).Scan(
		&s.ID, &s.UserID, &s.Name, &s.PathPrefix, &s.TargetURL,
		&s.RateLimitRPS, &s.BurstCapacity, &s.IsActive, &s.CreatedAt, &s.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrServiceNotFound
		}
		return nil, fmt.Errorf("service_repo.GetByPathPrefix: %w", err)
	}

	return &s, nil
}

func (r *ServiceRepository) GetByUserID(ctx context.Context, userID int) ([]domain.Service, error) {
	sql := `
		SELECT id, user_id, name, path_prefix, target_url, rate_limit_rps, burst_capacity, is_active, created_at, updated_at
		FROM services
		WHERE user_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.pg.Pool.Query(ctx, sql, userID)
	if err != nil {
		return nil, fmt.Errorf("service_repo.GetByUserID: %w", err)
	}
	defer rows.Close()

	var services []domain.Service
	for rows.Next() {
		var s domain.Service
		err := rows.Scan(
			&s.ID, &s.UserID, &s.Name, &s.PathPrefix, &s.TargetURL,
			&s.RateLimitRPS, &s.BurstCapacity, &s.IsActive, &s.CreatedAt, &s.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("service_repo.GetByUserID: scan row failed: %w", err)
		}
		services = append(services, s)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("service_repo.GetByUserID: rows iteration failed: %w", err)
	}

	return services, nil
}

func (r *ServiceRepository) Update(ctx context.Context, svc *domain.Service) error {
	sql := `
		UPDATE services
		SET name = $1, path_prefix = $2, target_url = $3, 
		    rate_limit_rps = $4, burst_capacity = $5, is_active = $6, updated_at = NOW()
		WHERE id = $7
	`

	tag, err := r.pg.Pool.Exec(ctx, sql,
		svc.Name, svc.PathPrefix, svc.TargetURL,
		svc.RateLimitRPS, svc.BurstCapacity, svc.IsActive, svc.ID,
	)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return domain.ErrPathPrefixAlreadyExists
		}
		return fmt.Errorf("service_repo.Update: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return domain.ErrServiceNotFound
	}

	return nil
}

func (r *ServiceRepository) Delete(ctx context.Context, id int) error {
	sql := `DELETE FROM services WHERE id = $1`

	tag, err := r.pg.Pool.Exec(ctx, sql, id)
	if err != nil {
		return fmt.Errorf("service_repo.Delete: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return domain.ErrServiceNotFound
	}

	return nil
}
