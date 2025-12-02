package config

import "os"

// Config holds all configuration for the application
type Config struct {
	Port          string
	DatabaseURL   string
	RedisURL      string
	JWTSecret     string
	KolosalAPIKey string // Using Kolosal.ai instead of OpenAI
	CORSOrigin    string
	LogLevel      string
}

// Load reads configuration from environment variables
func Load() *Config {
	return &Config{
		Port:          getEnv("PORT", "8080"),
		DatabaseURL:   getEnv("DATABASE_URL", "postgres://bantuaku:bantuaku_secret@localhost:5432/bantuaku_dev?sslmode=disable"),
		RedisURL:      getEnv("REDIS_URL", "redis://localhost:6379"),
		JWTSecret:     getEnv("JWT_SECRET", "dev-jwt-secret-change-in-production"),
		KolosalAPIKey: getEnv("KOLOSAL_API_KEY", ""),
		CORSOrigin:    getEnv("CORS_ORIGIN", "http://localhost:3000"),
		LogLevel:      getEnv("LOG_LEVEL", "info"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
