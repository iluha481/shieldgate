package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func writeEnv(t *testing.T, content string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(".", ".env"), []byte(content), 0o600); err != nil {
		t.Fatalf("failed to write .env: %v", err)
	}
}

func TestLoad(t *testing.T) {
	t.Chdir(t.TempDir())
	writeEnv(t, "")

	env := map[string]string{
		"DB_HOST":       "db.example",
		"DB_PORT":       "5432",
		"DB_USER":       "user",
		"DB_PASSWORD":   "pass",
		"DB_NAME":       "shieldgate",
		"REDIS_HOST":    "redis.example:6379",
		"APP_PORT":      "9090",
		"JWT_SECRET":    "secret",
		"READ_TIMEOUT":  "5s",
		"WRITE_TIMEOUT": "10s",
		"IDLE_TIMEOUT":  "2m",
	}
	for k, v := range env {
		t.Setenv(k, v)
	}

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.DBHost != "db.example" || cfg.DBPort != "5432" || cfg.DBUser != "user" ||
		cfg.DBPass != "pass" || cfg.DBName != "shieldgate" {
		t.Errorf("unexpected DB config: %+v", cfg)
	}
	if cfg.RedisHost != "redis.example:6379" || cfg.AppPort != "9090" || cfg.JWTSecret != "secret" {
		t.Errorf("unexpected service config: %+v", cfg)
	}
	if cfg.ReadTimeout != 5*time.Second {
		t.Errorf("ReadTimeout = %v, want 5s", cfg.ReadTimeout)
	}
	if cfg.WriteTimeout != 10*time.Second {
		t.Errorf("WriteTimeout = %v, want 10s", cfg.WriteTimeout)
	}
	if cfg.IdleTimeout != 2*time.Minute {
		t.Errorf("IdleTimeout = %v, want 2m", cfg.IdleTimeout)
	}
}

func TestLoad_ReadsDotEnv(t *testing.T) {
	t.Chdir(t.TempDir())
	writeEnv(t, "DB_HOST=from-dotenv\nREAD_TIMEOUT=1s\nWRITE_TIMEOUT=2s\nIDLE_TIMEOUT=3s\n")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.DBHost != "from-dotenv" {
		t.Errorf("DBHost = %q, want %q", cfg.DBHost, "from-dotenv")
	}
}

func TestLoad_MissingDotEnv(t *testing.T) {
	t.Chdir(t.TempDir())

	if _, err := Load(); err == nil {
		t.Fatal("Load() expected error when .env is missing")
	}
}

func TestLoad_InvalidTimeout(t *testing.T) {
	tests := []struct {
		name string
		key  string
	}{
		{name: "read", key: "READ_TIMEOUT"},
		{name: "write", key: "WRITE_TIMEOUT"},
		{name: "idle", key: "IDLE_TIMEOUT"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Chdir(t.TempDir())
			writeEnv(t, "")
			t.Setenv("READ_TIMEOUT", "1s")
			t.Setenv("WRITE_TIMEOUT", "1s")
			t.Setenv("IDLE_TIMEOUT", "1s")
			t.Setenv(tt.key, "not-a-duration")

			if _, err := Load(); err == nil {
				t.Fatalf("Load() expected error for invalid %s", tt.key)
			}
		})
	}
}

func TestMustLoad_PanicsOnError(t *testing.T) {
	t.Chdir(t.TempDir())

	defer func() {
		if recover() == nil {
			t.Fatal("MustLoad() expected panic when config cannot be loaded")
		}
	}()

	MustLoad()
}
