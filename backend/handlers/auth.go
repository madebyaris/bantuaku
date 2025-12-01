package handlers

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// RegisterRequest represents a registration request
type RegisterRequest struct {
	Email     string `json:"email"`
	Password  string `json:"password"`
	StoreName string `json:"store_name"`
	Industry  string `json:"industry,omitempty"`
}

// LoginRequest represents a login request
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// AuthResponse represents authentication response with token
type AuthResponse struct {
	Token     string `json:"token"`
	UserID    string `json:"user_id"`
	StoreID   string `json:"store_id"`
	StoreName string `json:"store_name"`
	Plan      string `json:"plan"`
}

// Register handles user registration
func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := parseJSON(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate input
	req.Email = strings.TrimSpace(strings.ToLower(req.Email))
	if req.Email == "" || !strings.Contains(req.Email, "@") {
		respondError(w, http.StatusBadRequest, "Valid email is required")
		return
	}
	if len(req.Password) < 6 {
		respondError(w, http.StatusBadRequest, "Password must be at least 6 characters")
		return
	}
	if req.StoreName == "" {
		respondError(w, http.StatusBadRequest, "Store name is required")
		return
	}

	ctx := r.Context()

	// Check if email already exists
	var existingID string
	err := h.db.Pool().QueryRow(ctx, "SELECT id FROM users WHERE email = $1", req.Email).Scan(&existingID)
	if err == nil {
		respondError(w, http.StatusConflict, "Email already registered")
		return
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to process password")
		return
	}

	// Create user and store in transaction
	userID := uuid.New().String()
	storeID := uuid.New().String()

	tx, err := h.db.Pool().Begin(ctx)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to start transaction")
		return
	}
	defer tx.Rollback(ctx)

	// Insert user
	_, err = tx.Exec(ctx, `
		INSERT INTO users (id, email, password_hash, created_at)
		VALUES ($1, $2, $3, $4)
	`, userID, req.Email, string(hashedPassword), time.Now())
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to create user")
		return
	}

	// Insert store
	_, err = tx.Exec(ctx, `
		INSERT INTO stores (id, user_id, store_name, industry, subscription_plan, status, created_at)
		VALUES ($1, $2, $3, $4, 'free', 'active', $5)
	`, storeID, userID, req.StoreName, req.Industry, time.Now())
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to create store")
		return
	}

	if err := tx.Commit(ctx); err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to commit transaction")
		return
	}

	// Generate JWT token
	token, err := h.generateToken(userID, storeID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to generate token")
		return
	}

	respondJSON(w, http.StatusCreated, AuthResponse{
		Token:     token,
		UserID:    userID,
		StoreID:   storeID,
		StoreName: req.StoreName,
		Plan:      "free",
	})
}

// Login handles user authentication
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := parseJSON(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	req.Email = strings.TrimSpace(strings.ToLower(req.Email))
	if req.Email == "" {
		respondError(w, http.StatusBadRequest, "Email is required")
		return
	}

	ctx := r.Context()

	// Get user by email
	var userID, passwordHash string
	err := h.db.Pool().QueryRow(ctx, `
		SELECT id, password_hash FROM users WHERE email = $1
	`, req.Email).Scan(&userID, &passwordHash)
	if err != nil {
		respondError(w, http.StatusUnauthorized, "Invalid email or password")
		return
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(req.Password)); err != nil {
		respondError(w, http.StatusUnauthorized, "Invalid email or password")
		return
	}

	// Get store for this user
	var storeID, storeName, plan string
	err = h.db.Pool().QueryRow(ctx, `
		SELECT id, store_name, subscription_plan FROM stores WHERE user_id = $1 AND status = 'active' LIMIT 1
	`, userID).Scan(&storeID, &storeName, &plan)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to fetch store")
		return
	}

	// Generate JWT token
	token, err := h.generateToken(userID, storeID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to generate token")
		return
	}

	respondJSON(w, http.StatusOK, AuthResponse{
		Token:     token,
		UserID:    userID,
		StoreID:   storeID,
		StoreName: storeName,
		Plan:      plan,
	})
}

func (h *Handler) generateToken(userID, storeID string) (string, error) {
	claims := jwt.MapClaims{
		"user_id":  userID,
		"store_id": storeID,
		"exp":      time.Now().Add(24 * time.Hour).Unix(),
		"iat":      time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(h.config.JWTSecret))
}

// GetStoreIDFromContext extracts store ID from request context
func GetStoreIDFromContext(ctx context.Context) string {
	if storeID, ok := ctx.Value("store_id").(string); ok {
		return storeID
	}
	return ""
}
