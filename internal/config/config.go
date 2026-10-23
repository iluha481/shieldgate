package config

import (
	"os"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	DBHost       string
	DBPort       string
	DBUser       string
	DBPass       string
	DBName       string
	RedisHost    string
	AppPort      string
	JWTSecret    string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
}

func Load() (Config, error) {
	err := godotenv.Load()
	if err != nil {
		return Config{}, err
	}
	readTimeout, err := time.ParseDuration(os.Getenv("READ_TIMEOUT"))
	if err != nil {
		return Config{}, err
	}
	writeTimeout, err := time.ParseDuration(os.Getenv("WRITE_TIMEOUT"))
	if err != nil {
		return Config{}, err
	}
	idleTimeout, err := time.ParseDuration(os.Getenv("IDLE_TIMEOUT"))
	if err != nil {
		return Config{}, err
	}

	return Config{
		DBHost:       os.Getenv("DB_HOST"),
		DBPort:       os.Getenv("DB_PORT"),
		DBUser:       os.Getenv("DB_USER"),
		DBPass:       os.Getenv("DB_PASSWORD"),
		DBName:       os.Getenv("DB_NAME"),
		RedisHost:    os.Getenv("REDIS_HOST"),
		AppPort:      os.Getenv("APP_PORT"),
		JWTSecret:    os.Getenv("JWT_SECRET"),
		ReadTimeout:  readTimeout,
		WriteTimeout: writeTimeout,
		IdleTimeout:  idleTimeout,
	}, nil
}

func MustLoad() Config {
	config, err := Load()
	if err != nil {
		panic(err)
	}
	return config
}
