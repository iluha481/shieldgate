package repository

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/iluha481/shieldgate/internal/domain"
	"github.com/iluha481/shieldgate/pkg/postgres"
)

type APIKeyRepository struct {
	pg *postgres.DB
}

func NewAPIKeyRepository(pg *postgres.DB) *APIKeyRepository {
	return &APIKeyRepository{pg: pg}
}

func (r *APIKeyRepository) Create(ctx context.Context, key *domain.APIKey) error {
	sql := `
		INSERT INTO api_keys (service_id, name, key_hash, key_prefix, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, NOW(), NOW())
		RETURNING id, created_at, updated_at
	`
	log.Printf("Creating API key for service_id: %d, key_prefix: %s", key.ServiceID, key.KeyPrefix)
	err := r.pg.Pool.QueryRow(ctx, sql,
		key.ServiceID,
		key.Name,
		key.KeyHash,
		key.KeyPrefix,
		key.IsActive,
	).Scan(&key.ID, &key.CreatedAt, &key.UpdatedAt)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return domain.ErrAPIKeyAlreadyExists
		}
		return fmt.Errorf("api_key_repo.Create: %w", err)
	}

	return nil
}

func (r *APIKeyRepository) GetByID(ctx context.Context, id int) (*domain.APIKey, error) {
	sql := `
		SELECT id, service_id, name, key_hash, key_prefix, is_active, created_at, updated_at
		FROM api_keys
		WHERE id = $1
	`

	var k domain.APIKey
	err := r.pg.Pool.QueryRow(ctx, sql, id).Scan(
		&k.ID, &k.ServiceID, &k.Name, &k.KeyHash, &k.KeyPrefix, &k.IsActive, &k.CreatedAt, &k.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrAPIKeyNotFound
		}
		return nil, fmt.Errorf("api_key_repo.GetByID: %w", err)
	}

	return &k, nil
}

func (r *APIKeyRepository) GetByKeyHash(ctx context.Context, keyHash string) (*domain.APIKey, error) {
	sql := `
		SELECT id, service_id, name, key_hash, key_prefix, is_active, created_at, updated_at
		FROM api_keys
		WHERE key_hash = $1 AND is_active = true
	`

	var k domain.APIKey
	err := r.pg.Pool.QueryRow(ctx, sql, keyHash).Scan(
		&k.ID, &k.ServiceID, &k.Name, &k.KeyHash, &k.KeyPrefix, &k.IsActive, &k.CreatedAt, &k.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrAPIKeyNotFound
		}
		return nil, fmt.Errorf("api_key_repo.GetByKeyHash: %w", err)
	}

	return &k, nil
}

func (r *APIKeyRepository) GetByServiceID(ctx context.Context, serviceID int) ([]domain.APIKey, error) {
	sql := `
		SELECT id, service_id, name, key_hash, key_prefix, is_active, created_at, updated_at
		FROM api_keys
		WHERE service_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.pg.Pool.Query(ctx, sql, serviceID)
	if err != nil {
		return nil, fmt.Errorf("api_key_repo.GetByServiceID: %w", err)
	}
	defer rows.Close()

	var keys []domain.APIKey
	for rows.Next() {
		var k domain.APIKey
		err := rows.Scan(
			&k.ID, &k.ServiceID, &k.Name, &k.KeyHash, &k.KeyPrefix, &k.IsActive, &k.CreatedAt, &k.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("api_key_repo.GetByServiceID: scan failed: %w", err)
		}
		keys = append(keys, k)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("api_key_repo.GetByServiceID: rows iteration failed: %w", err)
	}

	return keys, nil
}

func (r *APIKeyRepository) SetActive(ctx context.Context, id int, isActive bool) error {
	sql := `
		UPDATE api_keys
		SET is_active = $1, updated_at = NOW()
		WHERE id = $2
	`

	tag, err := r.pg.Pool.Exec(ctx, sql, isActive, id)
	if err != nil {
		return fmt.Errorf("api_key_repo.SetActive: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return domain.ErrAPIKeyNotFound
	}

	return nil
}

func (r *APIKeyRepository) Delete(ctx context.Context, id int) error {
	sql := `DELETE FROM api_keys WHERE id = $1`

	tag, err := r.pg.Pool.Exec(ctx, sql, id)
	if err != nil {
		return fmt.Errorf("api_key_repo.Delete: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return domain.ErrAPIKeyNotFound
	}

	return nil
}
