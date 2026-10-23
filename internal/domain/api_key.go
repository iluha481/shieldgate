package domain

import (
	"errors"
	"time"
)

var (
	ErrAPIKeyNotFound      = errors.New("api key not found")
	ErrAPIKeyAlreadyExists = errors.New("api key hash already exists")
)

type APIKey struct {
	ID        int       `json:"id"`
	ServiceID int       `json:"service_id"`
	Name      string    `json:"name"`
	KeyHash   string    `json:"-"`
	KeyPrefix string    `json:"key_prefix"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
