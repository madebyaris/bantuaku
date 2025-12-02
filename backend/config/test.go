package config

import (
	"os"
)

// LoadTest loads test configuration
func LoadTest() *Config {
	return &Config{
		DatabaseURL:   getTestDatabaseURL(),
		RedisURL:      getTestRedisURL(),
		JWTSecret:     "test-jwt-secret",
		KolosalAPIKey: "", // Disable Kolosal.ai in tests
		CORSOrigin:    "http://localhost:3000",
		Port:          "8080",
		LogLevel:      "debug",
	}
}

// getTestDatabaseURL returns the test database URL
func getTestDatabaseURL() string {
	if url := os.Getenv("TEST_DATABASE_URL"); url != "" {
		return url
	}
	return "postgres://bantuaku:bantuaku_secret@localhost:5432/bantuaku_test?sslmode=disable"
}

// getTestRedisURL returns the test Redis URL
func getTestRedisURL() string {
	if url := os.Getenv("TEST_REDIS_URL"); url != "" {
		return url
	}
	return "redis://localhost:6379/1" // Use database 1 for tests
}
