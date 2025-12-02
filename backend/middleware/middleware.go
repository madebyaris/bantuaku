package middleware

import (
	"context"
	"net/http"
	"strings"
	"time"

	apperrors "github.com/bantuaku/backend/errors"
	"github.com/bantuaku/backend/logger"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type contextKey string

const (
	RequestIDKey contextKey = "request_id"
	UserIDKey    contextKey = "user_id"
	StoreIDKey   contextKey = "store_id"
)

// Chain applies multiple middleware to a handler
func Chain(h http.Handler, middlewares ...func(http.Handler) http.Handler) http.Handler {
	for i := len(middlewares) - 1; i >= 0; i-- {
		h = middlewares[i](h)
	}
	return h
}

// RequestID adds a unique request ID to each request
func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := r.Header.Get("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
		}
		ctx := context.WithValue(r.Context(), RequestIDKey, requestID)
		w.Header().Set("X-Request-ID", requestID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// StructuredLogger logs request details using the structured logger
func StructuredLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		next.ServeHTTP(wrapped, r)

		requestID, _ := r.Context().Value(RequestIDKey).(string)
		log := logger.With("request_id", requestID)

		// Log request with structured fields
		log.Info(
			"HTTP request completed",
			"method", r.Method,
			"path", r.URL.Path,
			"query", r.URL.RawQuery,
			"remote_addr", r.RemoteAddr,
			"user_agent", r.UserAgent(),
			"status", wrapped.statusCode,
			"duration_ms", time.Since(start).Milliseconds(),
		)
	})
}

// Logger logs request details (alias for StructuredLogger for backward compatibility)
func Logger(next http.Handler) http.Handler {
	return StructuredLogger(next)
}

// DebugLogger logs detailed request information for debugging
func DebugLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		next.ServeHTTP(wrapped, r)

		requestID, _ := r.Context().Value(RequestIDKey).(string)
		log := logger.With("request_id", requestID).RequestID(r.Context())

		// Debug level logging with more details
		log.Debug(
			"HTTP request debug",
			"method", r.Method,
			"path", r.URL.Path,
			"query", r.URL.RawQuery,
			"remote_addr", r.RemoteAddr,
			"user_agent", r.UserAgent(),
			"referer", r.Referer(),
			"content_length", r.ContentLength,
			"status", wrapped.statusCode,
			"duration_ms", time.Since(start).Milliseconds(),
			"headers", getRelevantHeaders(r),
		)
	})
}

// getRelevantHeaders extracts relevant headers for debugging
func getRelevantHeaders(r *http.Request) map[string]string {
	headers := make(map[string]string)
	relevantHeaders := []string{
		"Accept", "Accept-Encoding", "Accept-Language",
		"Authorization", "Content-Type", "Referer", "User-Agent",
	}

	for _, header := range relevantHeaders {
		if value := r.Header.Get(header); value != "" {
			headers[header] = value
		}
	}

	return headers
}

// ErrorHandler handles errors with proper structured logging
func ErrorHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Create a contextual logger for this request
		requestID, _ := r.Context().Value(RequestIDKey).(string)
		log := logger.With("request_id", requestID)

		// Create a custom response writer to capture errors
		wrapped := &errorResponseWriter{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}

		// Catch panics and convert them to proper errors
		defer func() {
			if err := recover(); err != nil {
				appErr := apperrors.NewAppError(apperrors.ErrCodeInternal, "panic recovered", "")

				log.LogError(appErr, "Panic recovered in HTTP handler", r.Context())
				apperrors.WriteJSONError(w, appErr, appErr.Code)
			}
		}()

		// Set error context for all handlers
		ctx := context.WithValue(r.Context(), "logger", log)
		next.ServeHTTP(wrapped, r.WithContext(ctx))

		// Handle any errors that might have been set
		if wrapped.err != nil {
			log.LogError(wrapped.err, "Error in HTTP handler", r.Context())
			apperrors.WriteJSONError(w, wrapped.err, wrapped.errCode)
		}
	})
}

// errorResponseWriter extends responseWriter to capture errors
type errorResponseWriter struct {
	http.ResponseWriter
	statusCode int
	err        error
	errCode    apperrors.ErrorCode
}

// WriteError sets the error in the response writer
func (erw *errorResponseWriter) WriteError(err error, code apperrors.ErrorCode) {
	erw.err = err
	erw.errCode = code
}

// WriteHeader overrides the default WriteHeader to capture status code
func (erw *errorResponseWriter) WriteHeader(code int) {
	erw.statusCode = code
	erw.ResponseWriter.WriteHeader(code)
}

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// CORS handles Cross-Origin Resource Sharing
func CORS(allowedOrigin string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")

			// Log CORS request
			requestID, _ := r.Context().Value(RequestIDKey).(string)
			log := logger.With("request_id", requestID)

			if origin != "" {
				log.Debug("CORS request", "origin", origin, "allowed_origin", allowedOrigin)
			}

			// Set CORS headers
			if allowedOrigin == "*" || origin == allowedOrigin {
				w.Header().Set("Access-Control-Allow-Origin", allowedOrigin)
			} else {
				// For production, you might want to implement more sophisticated origin checking
				w.Header().Set("Access-Control-Allow-Origin", allowedOrigin)
			}

			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Request-ID")
			w.Header().Set("Access-Control-Allow-Credentials", "true")

			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// Recover handles panics gracefully with proper logging
func Recover(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				// Create contextual logger
				requestID, _ := r.Context().Value(RequestIDKey).(string)
				log := logger.With("request_id", requestID)

				// Create proper error object
				appErr := apperrors.NewAppError(apperrors.ErrCodeInternal, "panic recovered", "")

				// Log the panic
				log.LogError(appErr, "Panic recovered in HTTP handler", r.Context())

				// Return proper JSON error response
				apperrors.WriteJSONError(w, appErr, appErr.Code)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

// RateLimiter implements basic rate limiting
func RateLimiter(requestsPerMinute int) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// In a real implementation, you would use a proper rate limiting library
			// like github.com/ulule/limiter or implement token bucket algorithm

			// For now, just log that rate limiting would be applied here
			requestID, _ := r.Context().Value(RequestIDKey).(string)
			log := logger.With("request_id", requestID)

			log.Debug(
				"Rate limiting middleware applied",
				"requests_per_minute", requestsPerMinute,
				"client_ip", getClientIP(r),
			)

			next.ServeHTTP(w, r)
		})
	}
}

// getClientIP extracts the real client IP from request
func getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header (for reverse proxies)
	xForwardedFor := r.Header.Get("X-Forwarded-For")
	if xForwardedFor != "" {
		ips := strings.Split(xForwardedFor, ",")
		return strings.TrimSpace(ips[0])
	}

	// Check X-Real-IP header
	xRealIP := r.Header.Get("X-Real-IP")
	if xRealIP != "" {
		return xRealIP
	}

	// Fall back to RemoteAddr
	return strings.Split(r.RemoteAddr, ":")[0]
}

// Auth validates JWT tokens and extracts user/store info
func Auth(jwtSecret string, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Create contextual logger
		requestID, _ := r.Context().Value(RequestIDKey).(string)
		log := logger.With("request_id", requestID)

		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			err := apperrors.NewUnauthorizedError("Missing authorization header")
			log.LogError(err, "Authentication failed - missing header", r.Context())
			apperrors.WriteJSONError(w, err, err.Code)
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			err := apperrors.NewUnauthorizedError("Invalid authorization header format")
			log.LogError(err, "Authentication failed - invalid format", r.Context())
			apperrors.WriteJSONError(w, err, err.Code)
			return
		}

		tokenString := parts[1]

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return []byte(jwtSecret), nil
		})

		if err != nil || !token.Valid {
			// Check if token is expired
			if err != nil {
				// Try to extract expiration error
				if strings.Contains(err.Error(), "expired") || strings.Contains(err.Error(), "exp") {
					appErr := apperrors.NewAppError(apperrors.ErrCodeTokenExpired, "Token has expired", "")
					log.LogError(appErr, "Authentication failed - expired token", r.Context())
					apperrors.WriteJSONError(w, appErr, appErr.Code)
				} else {
					appErr := apperrors.NewUnauthorizedError("Invalid token")
					log.LogError(appErr, "Authentication failed - invalid token", r.Context())
					apperrors.WriteJSONError(w, appErr, appErr.Code)
				}
			} else {
				appErr := apperrors.NewUnauthorizedError("Invalid token")
				log.LogError(appErr, "Authentication failed - invalid token", r.Context())
				apperrors.WriteJSONError(w, appErr, appErr.Code)
			}
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			err := apperrors.NewUnauthorizedError("Invalid token claims")
			log.LogError(err, "Authentication failed - invalid claims", r.Context())
			apperrors.WriteJSONError(w, err, err.Code)
			return
		}

		userID, _ := claims["user_id"].(string)
		storeID, _ := claims["store_id"].(string)

		ctx := r.Context()
		ctx = context.WithValue(ctx, UserIDKey, userID)
		ctx = context.WithValue(ctx, StoreIDKey, storeID)

		log.Debug(
			"Authentication successful",
			"user_id", userID,
			"store_id", storeID,
		)

		next.ServeHTTP(w, r.WithContext(ctx))
	}
}

// GetUserID extracts user ID from context
func GetUserID(ctx context.Context) string {
	userID, _ := ctx.Value(UserIDKey).(string)
	return userID
}

// GetStoreID extracts store ID from context
func GetStoreID(ctx context.Context) string {
	storeID, _ := ctx.Value(StoreIDKey).(string)
	return storeID
}

// GetCompanyID extracts company ID from context (same as store_id for now)
// TODO: Update JWT to use company_id instead of store_id
func GetCompanyID(ctx context.Context) string {
	// For now, store_id in JWT is actually company_id after migration
	return GetStoreID(ctx)
}
