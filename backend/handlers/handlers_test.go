package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/bantuaku/backend/config"
	"github.com/bantuaku/backend/models"
	"github.com/bantuaku/backend/services/storage"
	"github.com/google/uuid"
)

// testDB is a test database wrapper
type testDB struct {
	*storage.PostgresDB
}

// setupTestDB creates a test database connection
func setupTestDB(t *testing.T) *testDB {
	t.Helper()

	// Use test database URL from environment or default
	testDBURL := "postgres://bantuaku:bantuaku_secret@localhost:5432/bantuaku_test?sslmode=disable"

	db, err := storage.NewPostgres(testDBURL)
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	return &testDB{db}
}

// cleanupTestDB cleans up the test database
func cleanupTestDB(t *testing.T, db *testDB) {
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

	db.Close()
}

// setupTestHandler creates a test handler with test dependencies
func setupTestHandler(t *testing.T) (*Handler, *testDB) {
	t.Helper()

	db := setupTestDB(t)

	// Redis connection - use a test Redis instance or mock
	testRedisURL := "redis://localhost:6379/1" // Use database 1 for tests
	redis, err := storage.NewRedis(testRedisURL)
	if err != nil {
		t.Logf("Warning: Failed to connect to test Redis: %v", err)
		// Continue without Redis for tests
	}

	cfg := &config.Config{
		JWTSecret:     "test-jwt-secret",
		KolosalAPIKey: "",
		CORSOrigin:    "http://localhost:3000",
	}

	handler := New(db, redis, cfg)

	return handler, db
}

// TestRegisterHandler tests the user registration endpoint
func TestRegisterHandler(t *testing.T) {
	handler, db := setupTestHandler(t)
	defer cleanupTestDB(t, db)

	tests := []struct {
		name           string
		body           string
		expectedStatus int
		checkResponse  func(*testing.T, *http.Response)
	}{
		{
			name: "Valid registration",
			body: `{
				"email": "test@example.com",
				"password": "password123",
				"store_name": "Test Store"
			}`,
			expectedStatus: http.StatusCreated,
			checkResponse: func(t *testing.T, r *http.Response) {
				var response AuthResponse
				if err := json.NewDecoder(r.Body).Decode(&response); err != nil {
					t.Errorf("Failed to decode response: %v", err)
					return
				}

				if response.Token == "" {
					t.Error("Expected token in response")
				}

				if response.StoreName != "Test Store" {
					t.Errorf("Expected store name 'Test Store', got '%s'", response.StoreName)
				}
			},
		},
		{
			name: "Invalid email",
			body: `{
				"email": "invalid-email",
				"password": "password123",
				"store_name": "Test Store"
			}`,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Short password",
			body: `{
				"email": "test@example.com",
				"password": "123",
				"store_name": "Test Store"
			}`,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Missing store name",
			body: `{
				"email": "test@example.com",
				"password": "password123"
			}`,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Invalid JSON",
			body:           `{invalid json}`,
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", strings.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			handler.Register(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.checkResponse != nil {
				tt.checkResponse(t, w.Result())
			}
		})
	}
}

// TestLoginHandler tests the user login endpoint
func TestLoginHandler(t *testing.T) {
	handler, db := setupTestHandler(t)
	defer cleanupTestDB(t, db)

	// First, create a test user
	userID := uuid.New().String()
	storeID := uuid.New().String()

	// Insert test user directly into database
	_, err := db.Pool().Exec(context.Background(), `
		INSERT INTO users (id, email, password_hash, created_at)
		VALUES ($1, $2, $3, $4)
	`, userID, "test@example.com", "$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZRGdjGj/n3.RSIaYEG9sFKqFqz2Py", time.Now())
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	// Insert test store
	_, err = db.Pool().Exec(context.Background(), `
		INSERT INTO stores (id, user_id, store_name, subscription_plan, status, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, storeID, userID, "Test Store", "free", "active", time.Now())
	if err != nil {
		t.Fatalf("Failed to create test store: %v", err)
	}

	tests := []struct {
		name           string
		body           string
		expectedStatus int
		checkResponse  func(*testing.T, *http.Response)
	}{
		{
			name: "Valid login",
			body: `{
				"email": "test@example.com",
				"password": "demo123"
			}`,
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, r *http.Response) {
				var response AuthResponse
				if err := json.NewDecoder(r.Body).Decode(&response); err != nil {
					t.Errorf("Failed to decode response: %v", err)
					return
				}

				if response.Token == "" {
					t.Error("Expected token in response")
				}

				if response.StoreName != "Test Store" {
					t.Errorf("Expected store name 'Test Store', got '%s'", response.StoreName)
				}
			},
		},
		{
			name: "Invalid email",
			body: `{
				"email": "invalid-email",
				"password": "demo123"
			}`,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Non-existent email",
			body: `{
				"email": "nonexistent@example.com",
				"password": "demo123"
			}`,
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name: "Wrong password",
			body: `{
				"email": "test@example.com",
				"password": "wrongpassword"
			}`,
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "Invalid JSON",
			body:           `{invalid json}`,
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", strings.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			handler.Login(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.checkResponse != nil {
				tt.checkResponse(t, w.Result())
			}
		})
	}
}

// TestHealthCheck tests the health check endpoint
func TestHealthCheck(t *testing.T) {
	handler, db := setupTestHandler(t)
	defer cleanupTestDB(t, db)

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	w := httptest.NewRecorder()

	handler.HealthCheck(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var response map[string]string
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Errorf("Failed to decode response: %v", err)
		return
	}

	if response["status"] != "ok" {
		t.Errorf("Expected status 'ok', got '%s'", response["status"])
	}
}

// TestProductHandlers tests product-related endpoints
func TestProductHandlers(t *testing.T) {
	handler, db := setupTestHandler(t)
	defer cleanupTestDB(t, db)

	// Create test user and store
	userID := uuid.New().String()
	storeID := uuid.New().String()

	// Create test user
	_, err := db.Pool().Exec(context.Background(), `
		INSERT INTO users (id, email, password_hash, created_at)
		VALUES ($1, $2, $3, $4)
	`, userID, "test@example.com", "$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZRGdjGj/n3.RSIaYEG9sFKqFqz2Py", time.Now())
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	// Create test store
	_, err = db.Pool().Exec(context.Background(), `
		INSERT INTO stores (id, user_id, store_name, subscription_plan, status, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, storeID, userID, "Test Store", "free", "active", time.Now())
	if err != nil {
		t.Fatalf("Failed to create test store: %v", err)
	}

	// Generate JWT token
	token, err := handler.generateToken(userID, storeID)
	if err != nil {
		t.Fatalf("Failed to generate test token: %v", err)
	}

	// Test CreateProduct
	t.Run("CreateProduct", func(t *testing.T) {
		body := `{
			"product_name": "Test Product",
			"sku": "TEST-001",
			"category": "Test Category",
			"unit_price": 99.99,
			"cost": 50.00
		}`

		req := httptest.NewRequest(http.MethodPost, "/api/v1/products", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)

		w := httptest.NewRecorder()
		handler.CreateProduct(w, req)

		if w.Code != http.StatusCreated {
			t.Errorf("Expected status %d, got %d", http.StatusCreated, w.Code)
		}

		var product models.Product
		if err := json.NewDecoder(w.Body).Decode(&product); err != nil {
			t.Errorf("Failed to decode response: %v", err)
			return
		}

		if product.ProductName != "Test Product" {
			t.Errorf("Expected product name 'Test Product', got '%s'", product.ProductName)
		}
	})

	// Test ListProducts
	t.Run("ListProducts", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/products", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		w := httptest.NewRecorder()
		handler.ListProducts(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
		}

		var products []models.Product
		if err := json.NewDecoder(w.Body).Decode(&products); err != nil {
			t.Errorf("Failed to decode response: %v", err)
			return
		}

		if len(products) != 1 {
			t.Errorf("Expected 1 product, got %d", len(products))
		}
	})
}
