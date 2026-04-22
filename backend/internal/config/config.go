// Package config handles application configuration from environment variables.
package config

import (
	"log"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

// Config holds the application configuration loaded from environment variables.
type Config struct {
	Port             string
	GinMode          string
	CORSAllowOrigin  string
	DatabaseURL      string
	JWTSecret        string
	AuthCookieName   string
	AuthCookieSecure bool
	AuthTokenTTL     time.Duration
}

// Load reads configuration from environment variables with sensible defaults.
// It prefers loading from `.env.local` first, then falls back to `.env`.
// If neither file is present, it will use the system environment variables.
func Load() *Config {
	// Try to load .env.local first, then .env.
	// godotenv.Load returns an error if none of the provided files are found.
	if err := godotenv.Load(".env.local", ".env"); err != nil {
		log.Println("No .env.local or .env file found, using system environment variables")
	} else {
		log.Println("Environment variables loaded from .env.local or .env")
	}

	return &Config{
		Port:             getEnv("PORT", "8080"),
		GinMode:          getEnv("GIN_MODE", "debug"),
		CORSAllowOrigin:  getEnv("CORS_ALLOW_ORIGIN", "*"),
		DatabaseURL:      getEnv("DATABASE_URL", ""),
		JWTSecret:        getEnv("JWT_SECRET", "dev-secret-change-me"),
		AuthCookieName:   getEnv("AUTH_COOKIE_NAME", "auth_token"),
		AuthCookieSecure: getEnvBool("AUTH_COOKIE_SECURE", false),
		AuthTokenTTL:     getEnvDuration("AUTH_TOKEN_TTL", 24*time.Hour),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvBool(key string, fallback bool) bool {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}

	parsed, err := strconv.ParseBool(v)
	if err != nil {
		return fallback
	}

	return parsed
}

func getEnvDuration(key string, fallback time.Duration) time.Duration {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}

	parsed, err := time.ParseDuration(v)
	if err != nil {
		return fallback
	}

	return parsed
}
