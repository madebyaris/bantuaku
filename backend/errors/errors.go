package errors

import (
	"encoding/json"
	"fmt"
	"net/http"
	"runtime"
	"time"
)

// ErrorCode represents different types of errors
type ErrorCode string

const (
	// Validation errors
	ErrCodeValidation   ErrorCode = "validation_error"
	ErrCodeInvalidInput ErrorCode = "invalid_input"
	ErrCodeMissingInput ErrorCode = "missing_input"

	// Authentication errors
	ErrCodeUnauthorized ErrorCode = "unauthorized"
	ErrCodeForbidden    ErrorCode = "forbidden"
	ErrCodeInvalidToken ErrorCode = "invalid_token"
	ErrCodeTokenExpired ErrorCode = "token_expired"

	// Resource errors
	ErrCodeNotFound      ErrorCode = "not_found"
	ErrCodeConflict      ErrorCode = "conflict"
	ErrCodeLimitExceeded ErrorCode = "limit_exceeded"

	// System errors
	ErrCodeInternal ErrorCode = "internal_error"
	ErrCodeDatabase ErrorCode = "database_error"
	ErrCodeExternal ErrorCode = "external_service_error"

	// Business logic errors
	ErrCodeBusiness          ErrorCode = "business_rule_violation"
	ErrCodeInsufficientStock ErrorCode = "insufficient_stock"
)

// AppError represents an application error with structured information
type AppError struct {
	Code       ErrorCode `json:"code"`
	Message    string    `json:"message"`
	Details    string    `json:"details,omitempty"`
	Timestamp  string    `json:"timestamp"`
	StackTrace string    `json:"stack_trace,omitempty"`
}

// Error implements the error interface
func (e *AppError) Error() string {
	return e.Message
}

// NewAppError creates a new application error
func NewAppError(code ErrorCode, message, details string) *AppError {
	return &AppError{
		Code:      code,
		Message:   message,
		Details:   details,
		Timestamp: time.Now().Format(time.RFC3339),
	}
}

// NewValidationError creates a validation error
func NewValidationError(message, details string) *AppError {
	return NewAppError(ErrCodeValidation, message, details)
}

// NewUnauthorizedError creates an unauthorized error
func NewUnauthorizedError(message string) *AppError {
	if message == "" {
		message = "Unauthorized access"
	}
	return NewAppError(ErrCodeUnauthorized, message, "")
}

// NewForbiddenError creates a forbidden error
func NewForbiddenError(message string) *AppError {
	if message == "" {
		message = "Access forbidden"
	}
	return NewAppError(ErrCodeForbidden, message, "")
}

// NewNotFoundError creates a not found error
func NewNotFoundError(resource string) *AppError {
	message := fmt.Sprintf("%s not found", resource)
	return NewAppError(ErrCodeNotFound, message, "")
}

// NewConflictError creates a conflict error
func NewConflictError(message, details string) *AppError {
	return NewAppError(ErrCodeConflict, message, details)
}

// NewInternalError creates an internal server error
func NewInternalError(err error, message string) *AppError {
	details := err.Error()
	return NewAppError(ErrCodeInternal, message, details)
}

// NewDatabaseError creates a database error
func NewDatabaseError(err error, operation string) *AppError {
	message := fmt.Sprintf("Database operation failed: %s", operation)
	details := err.Error()
	appErr := NewAppError(ErrCodeDatabase, message, details)

	// Add stack trace for database errors
	appErr.StackTrace = getStackTrace()
	return appErr
}

// NewExternalServiceError creates an external service error
func NewExternalServiceError(service, message, details string) *AppError {
	errorMessage := fmt.Sprintf("External service error (%s): %s", service, message)
	return NewAppError(ErrCodeExternal, errorMessage, details)
}

// NewBusinessRuleError creates a business rule violation error
func NewBusinessRuleError(rule, message string) *AppError {
	errorMessage := fmt.Sprintf("Business rule violation (%s): %s", rule, message)
	return NewAppError(ErrCodeBusiness, errorMessage, "")
}

// NewInsufficientStockError creates an insufficient stock error
func NewInsufficientStockError(productID string, requested, available int) *AppError {
	message := "Insufficient stock"
	details := fmt.Sprintf("Product %s: requested %d, available %d", productID, requested, available)
	return NewAppError(ErrCodeInsufficientStock, message, details)
}

// IsAppError checks if an error is an AppError
func IsAppError(err error) bool {
	_, ok := err.(*AppError)
	return ok
}

// GetErrorCode returns the error code from an error
func GetErrorCode(err error) ErrorCode {
	if appErr, ok := err.(*AppError); ok {
		return appErr.Code
	}
	return ErrCodeInternal
}

// HTTPStatusFromErrorCode returns appropriate HTTP status code for an error code
func HTTPStatusFromErrorCode(code ErrorCode) int {
	switch code {
	case ErrCodeValidation, ErrCodeInvalidInput, ErrCodeMissingInput:
		return 400
	case ErrCodeUnauthorized:
		return 401
	case ErrCodeForbidden:
		return 403
	case ErrCodeNotFound:
		return 404
	case ErrCodeConflict:
		return 409
	case ErrCodeBusiness, ErrCodeInsufficientStock, ErrCodeLimitExceeded:
		return 422
	case ErrCodeTokenExpired:
		return 419
	default:
		return 500
	}
}

// getStackTrace generates the current stack trace
func getStackTrace() string {
	buf := make([]byte, 1024)
	n := runtime.Stack(buf, false)
	return string(buf[:n])
}

// WriteJSONError writes a proper JSON error response
func WriteJSONError(w http.ResponseWriter, err error, code ErrorCode) {
	w.Header().Set("Content-Type", "application/json")

	statusCode := HTTPStatusFromErrorCode(code)
	w.WriteHeader(statusCode)

	// If it's already an AppError, use it directly
	if appErr, ok := err.(*AppError); ok {
		if err := json.NewEncoder(w).Encode(appErr); err != nil {
			// Log encoding error but can't do much else
			http.Error(w, "Failed to encode error response", http.StatusInternalServerError)
		}
		return
	}

	// Otherwise create a generic error response
	appErr := NewAppError(code, err.Error(), "")
	if err := json.NewEncoder(w).Encode(appErr); err != nil {
		http.Error(w, "Failed to encode error response", http.StatusInternalServerError)
	}
}
