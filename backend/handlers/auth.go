package handlers

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/bantuaku/backend/errors"
	"github.com/bantuaku/backend/logger"
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

	// Insert company (stores table was renamed to companies in migration 003)
	_, err = tx.Exec(ctx, `
		INSERT INTO companies (id, owner_user_id, name, industry, subscription_plan, status, created_at)
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

	// Generate JWT token (default role is 'user')
	token, err := h.generateToken(userID, storeID, "user")
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

	// Get user by email (including role and status for JWT)
	// Use exact match since email is already normalized
	var userID, passwordHash, role, status string
	err := h.db.Pool().QueryRow(ctx, `
		SELECT id, password_hash, COALESCE(role, 'user'), COALESCE(status, 'active') 
		FROM users 
		WHERE email = $1
	`, req.Email).Scan(&userID, &passwordHash, &role, &status)
	if err != nil {
		// Log the actual error for debugging (but don't expose it to user)
		log := logger.With("request_id", r.Context().Value("request_id"))
		log.Debug("Login failed - user not found", "email", req.Email, "error", err.Error())
		appErr := errors.NewUnauthorizedError("Invalid email or password")
		h.respondError(w, appErr, r)
		return
	}

	// Check if user is active
	if status != "active" {
		appErr := errors.NewUnauthorizedError("Account is suspended. Please contact support.")
		h.respondError(w, appErr, r)
		return
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(req.Password)); err != nil {
		appErr := errors.NewUnauthorizedError("Invalid email or password")
		h.respondError(w, appErr, r)
		return
	}

	// Get company for this user (stores table was renamed to companies in migration 003)
	log := logger.With("request_id", r.Context().Value("request_id"))
	
	// Try to get active company first, then fall back to any company
	var storeID, storeName, plan string
	err = h.db.Pool().QueryRow(ctx, `
		SELECT id, name, subscription_plan 
		FROM companies 
		WHERE owner_user_id = $1 AND status = 'active' 
		LIMIT 1
	`, userID).Scan(&storeID, &storeName, &plan)
	
	// If no active company, try to get any company for this user (including inactive)
	if err != nil {
		log.Debug("No active company found, trying any company", "user_id", userID, "error", err.Error())
		err = h.db.Pool().QueryRow(ctx, `
			SELECT id, name, subscription_plan 
			FROM companies 
			WHERE owner_user_id = $1 
			ORDER BY created_at DESC
			LIMIT 1
		`, userID).Scan(&storeID, &storeName, &plan)
		
		// If we found an inactive company, reactivate it
		if err == nil {
			var companyStatus string
			statusErr := h.db.Pool().QueryRow(ctx, `SELECT status FROM companies WHERE id = $1`, storeID).Scan(&companyStatus)
			if statusErr == nil && companyStatus != "active" {
				log.Info("Reactivating inactive company for user", "user_id", userID, "company_id", storeID)
				_, _ = h.db.Pool().Exec(ctx, `UPDATE companies SET status = 'active' WHERE id = $1`, storeID)
			}
		}
	}
	
	// If still no company found, this should not happen with transaction-based user creation
	// Return an error instead of auto-creating with generic name
	if err != nil {
		log.Error("CRITICAL: Company not found for user - user was created without company!", "user_id", userID, "error", err.Error())
		appErr := errors.NewInternalError(err, "User account is missing company information. Please contact support to fix this issue.")
		h.respondError(w, appErr, r)
		return
	}

	// Generate JWT token (include role)
	token, err := h.generateToken(userID, storeID, role)
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

func (h *Handler) generateToken(userID, storeID, role string) (string, error) {
	claims := jwt.MapClaims{
		"user_id":  userID,
		"store_id": storeID,
		"role":     role,
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
