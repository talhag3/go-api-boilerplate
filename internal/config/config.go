package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)
// Config is a struct that holds all the configuration values for our app.
// I listed all of them here so they are easy to find!
type Config struct {
	AppPort    string // The port our web server will listen on (default is 3000)
	AppEnv     string // development, production, etc.
	DBHost     string // Where the database is hosted (like localhost or a remote IP)
	DBPort     string // Database port (usually 5432 for Postgres)
	DBUser     string // Username for the database login
	DBPassword string // Password for the database login
	DBName     string // The name of the database we want to connect to
	DBSSLMode  string // disable, require, etc. for security
	DBMaxConns int32  // Maximum connections in the database pool
	LogLevel   string // log level (debug, info, warn, error)
}

// LoadConfig loads the settings from the environment or .env file.
func LoadConfig() (*Config, error) {
	// Let's try to load the .env file. If it fails, that's fine (maybe we are running in Docker/prod)
	// so we ignore the error using "_"
	_ = godotenv.Load()

	// DB_MAX_CONNS in env is a string, so we need to convert it to an integer.
	// strconv is the package for string conversions. ParseInt takes the string, base 10, and bit size 32.
	maxConns, err := strconv.ParseInt(envOr("DB_MAX_CONNS", "10"), 10, 32)
	if err != nil {
		// If someone entered a bad number like "hello", we return an error!
		return nil, fmt.Errorf("invalid DB_MAX_CONNS: %w", err)
	}

	// Create and return the Config struct with all values loaded.
	// If the environment variables are not set, we use fallback values.
	return &Config{
		AppPort:    envOr("APP_PORT", "3000"),
		AppEnv:     envOr("APP_ENV", "DEBUG"),
		DBHost:     envOr("DB_HOST", "localhost"),
		DBPort:     envOr("DB_PORT", "5432"),
		DBUser:     envOr("DB_USER", "postgres"),
		DBPassword: envOr("DB_PASSWORD", "postgres"),
		DBName:     envOr("DB_NAME", "userapi"),
		DBSSLMode:  envOr("DB_SSLMODE", "disable"),
		DBMaxConns: int32(maxConns),
		LogLevel:   envOr("LOG_LEVEL", "info"),
	}, nil
}

// envOr checks if an environment variable exists.
// If it exists and is not empty, we return it. Otherwise, we return the fallback string.
func envOr(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return fallback
}

// DSN returns the Data Source Name string which is the connection string for PostgreSQL.
// It format is: postgres://username:password@host:port/database_name?sslmode=disable
func (c *Config) DSN() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		c.DBUser, c.DBPassword, c.DBHost, c.DBPort, c.DBName, c.DBSSLMode,
	)
}
