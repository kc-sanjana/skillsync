package config

import (
	"os"
	"time"
)

type Config struct {
	Port           string
	DatabaseURL    string
	JWTSecret      string
	JWTExpiry      time.Duration
	ClaudeAPIKey   string
	AllowedOrigins []string
	LogLevel       string
	Environment    string
}

func Load() *Config {
	return &Config{
		Port:           getEnv("PORT", "8080"),
		DatabaseURL:    getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/skillsync?sslmode=disable"),
		JWTSecret:      getEnv("JWT_SECRET", "change-me-in-production"),
		JWTExpiry:      parseDuration(getEnv("JWT_EXPIRY", "24h")),
		ClaudeAPIKey:   getEnv("CLAUDE_API_KEY", ""),
		AllowedOrigins: []string{getEnv("ALLOWED_ORIGINS", "http://localhost:3000")},
		LogLevel:       getEnv("LOG_LEVEL", "info"),
		Environment:    getEnv("ENVIRONMENT", "development"),
	}
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}

func parseDuration(s string) time.Duration {
	d, err := time.ParseDuration(s)
	if err != nil {
		return 24 * time.Hour
	}
	return d
}
