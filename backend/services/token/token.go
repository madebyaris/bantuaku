package token

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/bantuaku/backend/errors"
	"github.com/bantuaku/backend/logger"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Service handles OTP and token generation, validation, and management
type Service struct {
	pool *pgxpool.Pool
	log  logger.Logger
}

// NewService creates a new token service
func NewService(pool *pgxpool.Pool) *Service {
	return &Service{
		pool: pool,
		log:  *logger.Default(),
	}
}

// CodeType represents the type of verification code
type CodeType string

const (
	CodeTypeEmailVerification CodeType = "email_verification"
	CodeTypePasswordReset     CodeType = "password_reset"
)

// VerificationCode represents a verification code or token
type VerificationCode struct {
	ID        string
	UserID    string
	Code      string // 5-digit OTP for email verification
	Token     string // Secure token for password reset
	CodeType  CodeType
	ExpiresAt time.Time
	UsedAt    *time.Time
	CreatedAt time.Time
}

// GenerateOTP generates a 5-digit OTP (00000-99999)
func GenerateOTP() (string, error) {
	// Generate 5 random digits
	bytes := make([]byte, 5)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}

	// Convert to 5-digit number (00000-99999)
	otp := ""
	for _, b := range bytes {
		otp += fmt.Sprintf("%d", int(b)%10)
	}

	return otp, nil
}

// GenerateToken generates a secure random token (32 bytes, hex encoded)
func GenerateToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}
	return hex.EncodeToString(bytes), nil
}

// CreateVerificationOTP creates a new email verification OTP
func (s *Service) CreateVerificationOTP(ctx context.Context, userID string) (string, error) {
	// Check rate limit: max 3 OTP requests per user per hour
	count, err := s.countRecentOTPs(ctx, userID, CodeTypeEmailVerification, time.Hour)
	if err != nil {
		return "", fmt.Errorf("failed to check rate limit: %w", err)
	}
	if count >= 3 {
		return "", errors.NewRateLimitError("Too many verification requests. Please try again later.", "Maximum 3 OTP requests per hour")
	}

	// Invalidate existing OTPs for this user
	if err := s.invalidateExistingCodes(ctx, userID, CodeTypeEmailVerification); err != nil {
		s.log.Warn("Failed to invalidate existing OTPs", "error", err)
		// Continue anyway
	}

	// Generate OTP
	otp, err := GenerateOTP()
	if err != nil {
		return "", fmt.Errorf("failed to generate OTP: %w", err)
	}

	// Create verification code
	codeID := uuid.New().String()
	expiresAt := time.Now().Add(1 * time.Hour)

	_, err = s.pool.Exec(ctx, `
		INSERT INTO verification_codes (id, user_id, code, token, code_type, expires_at, created_at)
		VALUES ($1, $2, $3, NULL, $4, $5, NOW())
	`, codeID, userID, otp, CodeTypeEmailVerification, expiresAt)
	if err != nil {
		return "", fmt.Errorf("failed to create verification code: %w", err)
	}

	s.log.Info("Verification OTP created", "user_id", userID, "expires_at", expiresAt)
	return otp, nil
}

// CreatePasswordResetToken creates a new password reset token
func (s *Service) CreatePasswordResetToken(ctx context.Context, userID string) (string, error) {
	// Invalidate existing tokens for this user
	if err := s.invalidateExistingCodes(ctx, userID, CodeTypePasswordReset); err != nil {
		s.log.Warn("Failed to invalidate existing tokens", "error", err)
		// Continue anyway
	}

	// Generate token
	token, err := GenerateToken()
	if err != nil {
		return "", fmt.Errorf("failed to generate token: %w", err)
	}

	// Create verification code
	codeID := uuid.New().String()
	expiresAt := time.Now().Add(1 * time.Hour)

	_, err = s.pool.Exec(ctx, `
		INSERT INTO verification_codes (id, user_id, code, token, code_type, expires_at, created_at)
		VALUES ($1, $2, NULL, $3, $4, $5, NOW())
	`, codeID, userID, token, CodeTypePasswordReset, expiresAt)
	if err != nil {
		return "", fmt.Errorf("failed to create password reset token: %w", err)
	}

	s.log.Info("Password reset token created", "user_id", userID, "expires_at", expiresAt)
	return token, nil
}

// ValidateAndConsumeOTP validates and consumes an OTP for email verification
func (s *Service) ValidateAndConsumeOTP(ctx context.Context, userID, otp string) error {
	// Normalize OTP (trim whitespace, case-insensitive)
	otp = strings.TrimSpace(otp)
	if len(otp) != 5 {
		return errors.NewValidationError("Invalid OTP format", "OTP must be exactly 5 digits")
	}

	var code VerificationCode
	var usedAt sql.NullTime
	err := s.pool.QueryRow(ctx, `
		SELECT id, user_id, code, token, code_type, expires_at, used_at, created_at
		FROM verification_codes
		WHERE user_id = $1 
			AND code = $2 
			AND code_type = $3
			AND used_at IS NULL
		ORDER BY created_at DESC
		LIMIT 1
	`, userID, otp, CodeTypeEmailVerification).Scan(
		&code.ID, &code.UserID, &code.Code, &code.Token, &code.CodeType,
		&code.ExpiresAt, &usedAt, &code.CreatedAt,
	)

	if err == pgx.ErrNoRows {
		return errors.NewValidationError("Invalid or expired OTP", "The verification code is invalid or has already been used")
	}
	if err != nil {
		return fmt.Errorf("failed to query verification code: %w", err)
	}

	// Check if expired
	if time.Now().After(code.ExpiresAt) {
		return errors.NewValidationError("OTP expired", "The verification code has expired. Please request a new one")
	}

	// Mark as used
	_, err = s.pool.Exec(ctx, `
		UPDATE verification_codes
		SET used_at = NOW()
		WHERE id = $1
	`, code.ID)
	if err != nil {
		return fmt.Errorf("failed to mark OTP as used: %w", err)
	}

	s.log.Info("OTP validated and consumed", "user_id", userID)
	return nil
}

// ValidateAndConsumeToken validates and consumes a password reset token
func (s *Service) ValidateAndConsumeToken(ctx context.Context, token string) (string, error) {
	var code VerificationCode
	var usedAt sql.NullTime
	err := s.pool.QueryRow(ctx, `
		SELECT id, user_id, code, token, code_type, expires_at, used_at, created_at
		FROM verification_codes
		WHERE token = $1 
			AND code_type = $2
			AND used_at IS NULL
		ORDER BY created_at DESC
		LIMIT 1
	`, token, CodeTypePasswordReset).Scan(
		&code.ID, &code.UserID, &code.Code, &code.Token, &code.CodeType,
		&code.ExpiresAt, &usedAt, &code.CreatedAt,
	)

	if err == pgx.ErrNoRows {
		return "", errors.NewValidationError("Invalid or expired token", "The reset token is invalid or has already been used")
	}
	if err != nil {
		return "", fmt.Errorf("failed to query reset token: %w", err)
	}

	// Check if expired
	if time.Now().After(code.ExpiresAt) {
		return "", errors.NewValidationError("Token expired", "The reset token has expired. Please request a new one")
	}

	// Mark as used
	_, err = s.pool.Exec(ctx, `
		UPDATE verification_codes
		SET used_at = NOW()
		WHERE id = $1
	`, code.ID)
	if err != nil {
		return "", fmt.Errorf("failed to mark token as used: %w", err)
	}

	s.log.Info("Password reset token validated and consumed", "user_id", code.UserID)
	return code.UserID, nil
}

// countRecentOTPs counts recent OTP requests for rate limiting
func (s *Service) countRecentOTPs(ctx context.Context, userID string, codeType CodeType, duration time.Duration) (int, error) {
	since := time.Now().Add(-duration)
	var count int
	err := s.pool.QueryRow(ctx, `
		SELECT COUNT(*)
		FROM verification_codes
		WHERE user_id = $1 
			AND code_type = $2
			AND created_at >= $3
	`, userID, codeType, since).Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}

// invalidateExistingCodes invalidates existing unused codes for a user
func (s *Service) invalidateExistingCodes(ctx context.Context, userID string, codeType CodeType) error {
	_, err := s.pool.Exec(ctx, `
		UPDATE verification_codes
		SET used_at = NOW()
		WHERE user_id = $1 
			AND code_type = $2
			AND used_at IS NULL
			AND expires_at > NOW()
	`, userID, codeType)
	return err
}

// CleanupExpiredCodes removes expired codes (can be called periodically)
func (s *Service) CleanupExpiredCodes(ctx context.Context) error {
	result, err := s.pool.Exec(ctx, `
		DELETE FROM verification_codes
		WHERE expires_at < NOW() - INTERVAL '7 days'
	`)
	if err != nil {
		return fmt.Errorf("failed to cleanup expired codes: %w", err)
	}
	deleted := result.RowsAffected()
	if deleted > 0 {
		s.log.Info("Cleaned up expired verification codes", "count", deleted)
	}
	return nil
}

