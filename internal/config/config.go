package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	AppPort    string
	AppEnv     string
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	DBSSLMode  string
	DBMaxConns int
	LogLevel   string // log level (debug, info, warn, error)
}

func LoadConfig() (*Config, error) {
	_ = godotenv.Load()

	maxConns, err := strconv.ParseInt(envOr("DB_MAX_CONNS", "10"), 10, 32)

	if err != nil {
		return nil, fmt.Errorf("invalid DB_MAX_CONNS: %w", err)
	}

	return &Config{
		AppPort:    envOr("APP_PORT", "3000"),
		AppEnv:     envOr("APP_ENV", "DEBUG"),
		DBHost:     envOr("DB_HOST", "localhost"),
		DBPort:     envOr("DB_PORT", "5432"),
		DBUser:     envOr("DB_USER", "postgres"),
		DBPassword: envOr("DB_PASSWORD", "postgres"),
		DBName:     envOr("DB_NAME", "userapi"),
		DBSSLMode:  envOr("DB_SSLMODE", "disable"),
		DBMaxConns: int(maxConns),
		LogLevel:   envOr("LOG_LEVEL", "info"),
	}, nil
}

func envOr(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return fallback
}
