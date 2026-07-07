package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	AppPort string
	AppEnv  string
}

func LoadConfig() (*Config, error) {
	_ = godotenv.Load()

	return &Config{
		AppPort: envOr("APP_PORT", "3000"),
		AppEnv:  envOr("APP_ENVIRONMENT", "DEBUG"),
	}, nil
}

func envOr(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return fallback
}
