package domain

import (
	"errors"
	"time"
)

var (
	ErrServiceNotFound         = errors.New("service not found")
	ErrPathPrefixAlreadyExists = errors.New("service with this path prefix already exists")
)

type Service struct {
	ID            int       `json:"id"`
	UserID        int       `json:"user_id"`
	Name          string    `json:"name"`
	PathPrefix    string    `json:"path_prefix"`
	TargetURL     string    `json:"target_url"`
	RateLimitRPS  int       `json:"rate_limit_rps"`
	BurstCapacity int       `json:"burst_capacity"`
	IsActive      bool      `json:"is_active"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}
