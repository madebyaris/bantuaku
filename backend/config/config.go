package config

import (
	"os"
	"strconv"
)

// Config holds all configuration for the application
type Config struct {
	Port                 string
	DatabaseURL          string
	RedisURL             string
	JWTSecret            string
	KolosalAPIKey        string // Using Kolosal.ai instead of OpenAI
	OpenRouterAPIKey     string // OpenRouter API key
	OpenRouterModelChat  string // OpenRouter model for chat (e.g., "openai/gpt-4o-mini")
	OpenRouterModelEmbed string // OpenRouter model for embeddings (e.g., "openai/text-embedding-3-small")
	CORSOrigin           string
	LogLevel             string

	// Regulations scraper configuration
	RegulationsScraperEnabled  bool
	RegulationsScraperSchedule string
	RegulationsBaseURL         string

	// Embedding configuration
	EmbeddingProvider string
	EmbeddingAPIKey   string

	// Forecasting service configuration
	ForecastingServiceURL string

	// Stripe billing configuration
	StripeSecretKey     string
	StripeWebhookSecret string

	// Mailjet email configuration
	MailjetAPIKey    string
	MailjetAPISecret string
	AppBaseURL       string

	// Exa.ai search API configuration
	ExaAPIKey string
}

// Load reads configuration from environment variables
func Load() *Config {
	return &Config{
		Port:                 getEnv("PORT", "8080"),
		DatabaseURL:          getEnv("DATABASE_URL", "postgres://bantuaku:bantuaku_secret@localhost:5432/bantuaku_dev?sslmode=disable"),
		RedisURL:             getEnv("REDIS_URL", "redis://localhost:6379"),
		JWTSecret:            getEnv("JWT_SECRET", "dev-jwt-secret-change-in-production"),
		KolosalAPIKey:        getEnv("KOLOSAL_API_KEY", ""),
		OpenRouterAPIKey:     getEnv("OPENROUTER_API_KEY", ""),
		OpenRouterModelChat:  getEnv("OPENROUTER_MODEL_CHAT", "openai/gpt-4o-mini"),
		OpenRouterModelEmbed: getEnv("OPENROUTER_MODEL_EMBED", "openai/text-embedding-3-small"),
		CORSOrigin:           getEnv("CORS_ORIGIN", "http://localhost:3000"),
		LogLevel:             getEnv("LOG_LEVEL", "info"),

		RegulationsScraperEnabled:  getEnvBool("REGULATIONS_SCRAPER_ENABLED", true),
		RegulationsScraperSchedule: getEnv("REGULATIONS_SCRAPER_SCHEDULE", "0 2 * * *"),
		RegulationsBaseURL:         getEnv("REGULATIONS_BASE_URL", "https://peraturan.go.id"),

		EmbeddingProvider: getEnv("EMBEDDING_PROVIDER", "kolosal"),
		EmbeddingAPIKey:   getEnv("EMBEDDING_API_KEY", ""), // Falls back to KolosalAPIKey if empty

		ForecastingServiceURL: getEnv("FORECASTING_SERVICE_URL", "http://localhost:8001"),

		StripeSecretKey:     getEnv("STRIPE_SECRET_KEY", ""),
		StripeWebhookSecret: getEnv("STRIPE_WEBHOOK_SECRET", ""),

		MailjetAPIKey:    getEnv("MAILJET_API_KEY", ""),
		MailjetAPISecret: getEnv("MAILJET_API_SECRET", ""),
		AppBaseURL:       getEnv("APP_BASE_URL", "http://localhost:3000"),

		ExaAPIKey: getEnv("EXA_API_KEY", ""),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	val := os.Getenv(key)
	if val == "" {
		return defaultValue
	}
	result, err := strconv.ParseBool(val)
	if err != nil {
		return defaultValue
	}
	return result
}
