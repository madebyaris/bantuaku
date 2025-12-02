package handlers

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/bantuaku/backend/errors"
	"github.com/bantuaku/backend/validation"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// RegisterRequest represents a registration request
type RegisterRequest struct {
	Email     string `json:"email" validate:"required,email"`
	Password  string `json:"password" validate:"required,min:6"`
	StoreName string `json:"store_name" validate:"required,max:255"`
	Industry  string `json:"industry,omitempty" validate:"max:100"`
}

// LoginRequest represents a login request
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
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
	if err := h.parseJSON(r, &req); err != nil {
		h.respondError(w, err, r)
		return
	}

	// Validate input using validation package
	if err := validation.Validate(req); err != nil {
		h.respondError(w, err, r)
		return
	}

	// Manually validate email format (in addition to tag validation)
	req.Email = strings.TrimSpace(strings.ToLower(req.Email))
	if req.Email == "" || !strings.Contains(req.Email, "@") {
		err := errors.NewValidationError("Valid email is required", "Email field is missing or invalid")
		h.respondError(w, err, r)
		return
	}

	ctx := r.Context()

	// Check if email already exists
	var existingID string
	err := h.db.Pool().QueryRow(ctx, "SELECT id FROM users WHERE email = $1", req.Email).Scan(&existingID)
	if err == nil {
		appErr := errors.NewConflictError("Email already registered", "A user with this email already exists")
		h.respondError(w, appErr, r)
		return
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		appErr := errors.NewInternalError(err, "Failed to process password")
		h.respondError(w, appErr, r)
		return
	}

	// Create user and store in transaction
	userID := uuid.New().String()
	storeID := uuid.New().String()

	tx, err := h.db.Pool().Begin(ctx)
	if err != nil {
		appErr := errors.NewDatabaseError(err, "begin transaction")
		h.respondError(w, appErr, r)
		return
	}
	defer tx.Rollback(ctx)

	// Insert user
	_, err = tx.Exec(ctx, `
		INSERT INTO users (id, email, password_hash, created_at)
		VALUES ($1, $2, $3, $4)
	`, userID, req.Email, string(hashedPassword), time.Now())
	if err != nil {
		appErr := errors.NewDatabaseError(err, "create user")
		h.respondError(w, appErr, r)
		return
	}

	// Insert store
	_, err = tx.Exec(ctx, `
		INSERT INTO stores (id, user_id, store_name, industry, subscription_plan, status, created_at)
		VALUES ($1, $2, $3, $4, 'free', 'active', $5)
	`, storeID, userID, req.StoreName, req.Industry, time.Now())
	if err != nil {
		appErr := errors.NewDatabaseError(err, "create store")
		h.respondError(w, appErr, r)
		return
	}

	if err := tx.Commit(ctx); err != nil {
		appErr := errors.NewDatabaseError(err, "commit transaction")
		h.respondError(w, appErr, r)
		return
	}

	// Generate JWT token
	token, err := h.generateToken(userID, storeID)
	if err != nil {
		appErr := errors.NewInternalError(err, "Failed to generate token")
		h.respondError(w, appErr, r)
		return
	}

	h.respondJSON(w, http.StatusCreated, AuthResponse{
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
	if err := h.parseJSON(r, &req); err != nil {
		h.respondError(w, err, r)
		return
	}

	// Validate input using validation package
	if err := validation.Validate(req); err != nil {
		h.respondError(w, err, r)
		return
	}

	// Manually validate email format (in addition to tag validation)
	req.Email = strings.TrimSpace(strings.ToLower(req.Email))
	if req.Email == "" {
		err := errors.NewValidationError("Email is required", "Email field is missing")
		h.respondError(w, err, r)
		return
	}

	ctx := r.Context()

	// Get user by email
	var userID, passwordHash string
	err := h.db.Pool().QueryRow(ctx, `
		SELECT id, password_hash FROM users WHERE email = $1
	`, req.Email).Scan(&userID, &passwordHash)
	if err != nil {
		appErr := errors.NewUnauthorizedError("Invalid email or password")
		h.respondError(w, appErr, r)
		return
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(req.Password)); err != nil {
		appErr := errors.NewUnauthorizedError("Invalid email or password")
		h.respondError(w, appErr, r)
		return
	}

	// Get store for this user
	var storeID, storeName, plan string
	err = h.db.Pool().QueryRow(ctx, `
		SELECT id, store_name, subscription_plan FROM stores WHERE user_id = $1 AND status = 'active' LIMIT 1
	`, userID).Scan(&storeID, &storeName, &plan)
	if err != nil {
		appErr := errors.NewInternalError(err, "Failed to fetch store")
		h.respondError(w, appErr, r)
		return
	}

	// Generate JWT token
	token, err := h.generateToken(userID, storeID)
	if err != nil {
		appErr := errors.NewInternalError(err, "Failed to generate token")
		h.respondError(w, appErr, r)
		return
	}

	h.respondJSON(w, http.StatusOK, AuthResponse{
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
