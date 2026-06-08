// Package config handles application configuration from environment variables.
package config

import (
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
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
	GeminiAPIKey     string
	GeminiModel      string
}

// Load reads configuration from environment variables with sensible defaults.
// It prefers loading from `.env.local` first, then falls back to `.env` and `backend/.env`.
// If neither file is present, it will use the system environment variables.
func Load() *Config {
	loadedEnvFiles := loadEnvFiles()
	if len(loadedEnvFiles) == 0 {
		log.Println("No .env.local, .env or backend/.env file found, using system environment variables")
	} else {
		log.Printf("Environment variables loaded from %s", strings.Join(loadedEnvFiles, ", "))
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
		GeminiAPIKey:     getEnv("GEMINI_API_KEY", ""),
		GeminiModel:      getEnv("GEMINI_MODEL", "gemini-2.5-flash"),
	}
}

func loadEnvFiles() []string {
	candidates := envFileCandidates()
	loaded := make([]string, 0, len(candidates))

	for _, candidate := range candidates {
		if _, err := os.Stat(candidate); err != nil {
			continue
		}
		if err := godotenv.Load(candidate); err != nil {
			log.Printf("Failed to load environment file %s: %v", candidate, err)
			continue
		}
		loaded = append(loaded, candidate)
	}

	return loaded
}

func envFileCandidates() []string {
	workingDir, err := os.Getwd()
	if err == nil && filepath.Base(workingDir) == "backend" {
		return []string{"../.env.local", "../.env", ".env"}
	}

	return []string{".env.local", ".env", "backend/.env"}
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
