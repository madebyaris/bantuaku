package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/bantuaku/backend/config"
	"github.com/bantuaku/backend/services/storage"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// TestMain sets up the test environment
func TestMain(m *testing.M) {
	// Setup test database
	if err := setupTestDatabase(); err != nil {
		log.Printf("Warning: Failed to setup test database: %v", err)
	}

	// Run tests
	code := m.Run()

	// Cleanup test database
	if err := cleanupTestDatabase(); err != nil {
		log.Printf("Warning: Failed to cleanup test database: %v", err)
	}

	os.Exit(code)
}

// setupTestDatabase creates the test database schema
func setupTestDatabase() error {
	// Get test database URL from environment
	testDBURL := os.Getenv("TEST_DATABASE_URL")
	if testDBURL == "" {
		testDBURL = "postgres://bantuaku:bantuaku_secret@localhost:5432/bantuaku_test?sslmode=disable"
	}

	// Connect to PostgreSQL (without specifying database to create it)
	masterDBURL := testDBURL[:len(testDBURL)-len("bantuaku_test?sslmode=disable")] + "postgres?sslmode=disable"

	db, err := storage.NewPostgres(masterDBURL)
	if err != nil {
		return fmt.Errorf("failed to connect to master database: %v", err)
	}
	defer db.Close()

	// Create test database if it doesn't exist
	_, err = db.Pool().Exec(context.Background(),
		"CREATE DATABASE bantuaku_test")
	if err != nil {
		// Ignore error if database already exists
		if !strings.Contains(err.Error(), "already exists") {
			return fmt.Errorf("failed to create test database: %v", err)
		}
	}

	// Connect to test database
	testDB, err := storage.NewPostgres(testDBURL)
	if err != nil {
		return fmt.Errorf("failed to connect to test database: %v", err)
	}
	defer testDB.Close()

	// Read and execute schema migration
	schemaPath := "../database/migrations/001_init_schema.sql"
	schemaSQL, err := os.ReadFile(schemaPath)
	if err != nil {
		return fmt.Errorf("failed to read schema file: %v", err)
	}

	_, err = testDB.Pool().Exec(context.Background(), string(schemaSQL))
	if err != nil {
		return fmt.Errorf("failed to execute schema migration: %v", err)
	}

	return nil
}

// cleanupTestDatabase drops all tables from the test database
func cleanupTestDatabase() error {
	testDBURL := os.Getenv("TEST_DATABASE_URL")
	if testDBURL == "" {
		testDBURL = "postgres://bantuaku:bantuaku_secret@localhost:5432/bantuaku_test?sslmode=disable"
	}

	db, err := storage.NewPostgres(testDBURL)
	if err != nil {
		return fmt.Errorf("failed to connect to test database: %v", err)
	}
	defer db.Close()

	// List all tables and drop them
	rows, err := db.Pool().Query(context.Background(),
		"SELECT tablename FROM pg_tables WHERE schemaname = 'public'")
	if err != nil {
		return fmt.Errorf("failed to list tables: %v", err)
	}
	defer rows.Close()

	var tables []string
	for rows.Next() {
		var table string
		if err := rows.Scan(&table); err != nil {
			return fmt.Errorf("failed to scan table name: %v", err)
		}
		tables = append(tables, table)
	}

	// Drop all tables
	for _, table := range tables {
		_, err := db.Pool().Exec(context.Background(),
			fmt.Sprintf("DROP TABLE IF EXISTS %s CASCADE", table))
		if err != nil {
			return fmt.Errorf("failed to drop table %s: %v", table, err)
		}
	}

	return nil
}

// CreateTestConfig creates a test configuration
func CreateTestConfig() *config.Config {
	return &config.Config{
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

// CreateTestDB creates a test database connection
func CreateTestDB(t *testing.T) *storage.PostgresDB {
	t.Helper()

	db, err := storage.NewPostgres(getTestDatabaseURL())
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	return db
}

// CreateTestRedis creates a test Redis connection
func CreateTestRedis(t *testing.T) *storage.RedisDB {
	t.Helper()

	redis, err := storage.NewRedis(getTestRedisURL())
	if err != nil {
		t.Logf("Warning: Failed to connect to test Redis: %v", err)
		// Return nil to allow tests to run without Redis
		return nil
	}

	return redis
}

// CleanupTestData cleans up test data from all tables
func CleanupTestData(t *testing.T, db *storage.PostgresDB) {
	t.Helper()

	// Clean up tables in reverse order of dependencies
	tables := []string{
		"sales_history", "recommendations", "forecasts", "sentiment_data",
		"market_trends", "integrations", "products", "stores", "users",
	}

	for _, table := range tables {
		_, err := db.Pool().Exec(context.Background(), "DELETE FROM "+table)
		if err != nil {
			t.Logf("Warning: Failed to clean up table %s: %v", table, err)
		}
	}
}

// CreateTestUser creates a test user in the database
func CreateTestUser(t *testing.T, db *storage.PostgresDB, email, password string) (string, error) {
	t.Helper()

	userID := uuid.New().String()
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %v", err)
	}

	_, err = db.Pool().Exec(context.Background(), `
		INSERT INTO users (id, email, password_hash, created_at)
		VALUES ($1, $2, $3, $4)
	`, userID, email, string(hashedPassword), time.Now())

	return userID, err
}

// CreateTestStore creates a test store in the database
func CreateTestStore(t *testing.T, db *storage.PostgresDB, userID, storeName string) (string, error) {
	t.Helper()

	storeID := uuid.New().String()

	_, err := db.Pool().Exec(context.Background(), `
		INSERT INTO stores (id, user_id, store_name, subscription_plan, status, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, storeID, userID, storeName, "free", "active", time.Now())

	return storeID, err
}

// CreateTestProduct creates a test product in the database
func CreateTestProduct(t *testing.T, db *storage.PostgresDB, storeID string, productName string) (string, error) {
	t.Helper()

	productID := uuid.New().String()

	_, err := db.Pool().Exec(context.Background(), `
		INSERT INTO products (id, store_id, product_name, sku, category, unit_price, cost, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`, productID, storeID, productName, "TEST-SKU", "Test Category", 99.99, 50.00, time.Now(), time.Now())

	return productID, err
}

// CreateTestSale creates a test sale in the database
func CreateTestSale(t *testing.T, db *storage.PostgresDB, storeID, productID string, quantity int, price float64) error {
	t.Helper()

	_, err := db.Pool().Exec(context.Background(), `
		INSERT INTO sales_history (store_id, product_id, quantity, price, sale_date, source, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, storeID, productID, quantity, price, time.Now().AddDate(0, 0, -1), "manual", time.Now())

	return err
}
