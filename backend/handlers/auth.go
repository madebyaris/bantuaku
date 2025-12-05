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

	// Insert user with email_verified = false
	_, err = tx.Exec(ctx, `
		INSERT INTO users (id, email, password_hash, email_verified, created_at)
		VALUES ($1, $2, $3, false, $4)
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

	// Generate OTP and send verification email
	otp, err := h.tokenSvc.CreateVerificationOTP(ctx, userID)
	if err != nil {
		appErr := errors.NewInternalError(err, "Failed to create verification code")
		h.respondError(w, appErr, r)
		return
	}

	// Send verification email
	if err := h.emailSvc.SendVerificationEmail(req.Email, otp); err != nil {
		log := logger.With("request_id", r.Context().Value("request_id"))
		log.Warn("Failed to send verification email", "error", err, "user_id", userID)
		// Don't fail registration if email fails, but log it
	}

	// Return success without auto-login (user must verify email first)
	h.respondJSON(w, http.StatusCreated, map[string]interface{}{
		"message": "Registration successful. Please check your email for verification code.",
		"email":   req.Email,
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

	// Get user by email (including role, status, and email_verified for JWT)
	// Use exact match since email is already normalized
	var userID, passwordHash, role, status string
	var emailVerified bool
	err := h.db.Pool().QueryRow(ctx, `
		SELECT id, password_hash, COALESCE(role, 'user'), COALESCE(status, 'active'), COALESCE(email_verified, false)
		FROM users 
		WHERE email = $1
	`, req.Email).Scan(&userID, &passwordHash, &role, &status, &emailVerified)
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

	// Check if email is verified
	if !emailVerified {
		appErr := errors.NewUnauthorizedError("Email not verified. Please check your email for the verification code.")
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

// VerifyEmailRequest represents email verification request
type VerifyEmailRequest struct {
	Email string `json:"email" validate:"required,email"`
	OTP   string `json:"otp" validate:"required,len=5"`
}

// ResendVerificationRequest represents resend verification request
type ResendVerificationRequest struct {
	Email string `json:"email" validate:"required,email"`
}

// ForgotPasswordRequest represents forgot password request
type ForgotPasswordRequest struct {
	Email string `json:"email" validate:"required,email"`
}

// ResetPasswordRequest represents password reset request
type ResetPasswordRequest struct {
	Token       string `json:"token" validate:"required"`
	NewPassword string `json:"new_password" validate:"required,min:6"`
}

// VerifyEmail handles email verification with OTP
func (h *Handler) VerifyEmail(w http.ResponseWriter, r *http.Request) {
	var req VerifyEmailRequest
	if err := h.parseJSON(r, &req); err != nil {
		h.respondError(w, err, r)
		return
	}

	// Validate input
	if err := validation.Validate(req); err != nil {
		h.respondError(w, err, r)
		return
	}

	// Normalize email
	req.Email = strings.TrimSpace(strings.ToLower(req.Email))

	ctx := r.Context()

	// Get user ID by email
	var userID string
	err := h.db.Pool().QueryRow(ctx, "SELECT id FROM users WHERE email = $1", req.Email).Scan(&userID)
	if err != nil {
		appErr := errors.NewNotFoundError("User")
		h.respondError(w, appErr, r)
		return
	}

	// Validate and consume OTP
	if err := h.tokenSvc.ValidateAndConsumeOTP(ctx, userID, req.OTP); err != nil {
		h.respondError(w, err, r)
		return
	}

	// Mark email as verified
	_, err = h.db.Pool().Exec(ctx, `
		UPDATE users 
		SET email_verified = true, email_verified_at = NOW()
		WHERE id = $1
	`, userID)
	if err != nil {
		appErr := errors.NewDatabaseError(err, "update email verified")
		h.respondError(w, appErr, r)
		return
	}

	h.respondJSON(w, http.StatusOK, map[string]string{
		"message": "Email verified successfully",
	})
}

// ResendVerification handles resending verification email
func (h *Handler) ResendVerification(w http.ResponseWriter, r *http.Request) {
	var req ResendVerificationRequest
	if err := h.parseJSON(r, &req); err != nil {
		h.respondError(w, err, r)
		return
	}

	// Validate input
	if err := validation.Validate(req); err != nil {
		h.respondError(w, err, r)
		return
	}

	// Normalize email
	req.Email = strings.TrimSpace(strings.ToLower(req.Email))

	ctx := r.Context()

	// Get user ID by email
	var userID string
	var emailVerified bool
	err := h.db.Pool().QueryRow(ctx, `
		SELECT id, COALESCE(email_verified, false) 
		FROM users 
		WHERE email = $1
	`, req.Email).Scan(&userID, &emailVerified)
	if err != nil {
		appErr := errors.NewNotFoundError("User")
		h.respondError(w, appErr, r)
		return
	}

	// Check if already verified
	if emailVerified {
		appErr := errors.NewConflictError("Email already verified", "This email address has already been verified")
		h.respondError(w, appErr, r)
		return
	}

	// Generate new OTP
	otp, err := h.tokenSvc.CreateVerificationOTP(ctx, userID)
	if err != nil {
		h.respondError(w, err, r)
		return
	}

	// Send verification email
	if err := h.emailSvc.SendVerificationEmail(req.Email, otp); err != nil {
		appErr := errors.NewExternalServiceError("Mailjet", "Failed to send verification email", err.Error())
		h.respondError(w, appErr, r)
		return
	}

	h.respondJSON(w, http.StatusOK, map[string]string{
		"message": "Verification email sent successfully",
	})
}

// RequestPasswordReset handles password reset request
func (h *Handler) RequestPasswordReset(w http.ResponseWriter, r *http.Request) {
	var req ForgotPasswordRequest
	if err := h.parseJSON(r, &req); err != nil {
		h.respondError(w, err, r)
		return
	}

	// Validate input
	if err := validation.Validate(req); err != nil {
		h.respondError(w, err, r)
		return
	}

	// Normalize email
	req.Email = strings.TrimSpace(strings.ToLower(req.Email))

	ctx := r.Context()

	// Get user ID by email
	var userID string
	err := h.db.Pool().QueryRow(ctx, "SELECT id FROM users WHERE email = $1", req.Email).Scan(&userID)
	if err != nil {
		// Don't reveal if user exists or not (security best practice)
		// Return success even if user doesn't exist
		h.respondJSON(w, http.StatusOK, map[string]string{
			"message": "If the email exists, a password reset link has been sent",
		})
		return
	}

	// Generate reset token
	resetToken, err := h.tokenSvc.CreatePasswordResetToken(ctx, userID)
	if err != nil {
		appErr := errors.NewInternalError(err, "Failed to create reset token")
		h.respondError(w, appErr, r)
		return
	}

	// Send password reset email
	if err := h.emailSvc.SendPasswordResetEmail(req.Email, resetToken); err != nil {
		appErr := errors.NewExternalServiceError("Mailjet", "Failed to send password reset email", err.Error())
		h.respondError(w, appErr, r)
		return
	}

	h.respondJSON(w, http.StatusOK, map[string]string{
		"message": "If the email exists, a password reset link has been sent",
	})
}

// ResetPassword handles password reset with token
func (h *Handler) ResetPassword(w http.ResponseWriter, r *http.Request) {
	var req ResetPasswordRequest
	if err := h.parseJSON(r, &req); err != nil {
		h.respondError(w, err, r)
		return
	}

	// Validate input
	if err := validation.Validate(req); err != nil {
		h.respondError(w, err, r)
		return
	}

	ctx := r.Context()

	// Validate and consume token
	userID, err := h.tokenSvc.ValidateAndConsumeToken(ctx, req.Token)
	if err != nil {
		h.respondError(w, err, r)
		return
	}

	// Hash new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		appErr := errors.NewInternalError(err, "Failed to process password")
		h.respondError(w, appErr, r)
		return
	}

	// Update password
	_, err = h.db.Pool().Exec(ctx, `
		UPDATE users 
		SET password_hash = $1
		WHERE id = $2
	`, string(hashedPassword), userID)
	if err != nil {
		appErr := errors.NewDatabaseError(err, "update password")
		h.respondError(w, appErr, r)
		return
	}

	h.respondJSON(w, http.StatusOK, map[string]string{
		"message": "Password reset successfully",
	})
}
