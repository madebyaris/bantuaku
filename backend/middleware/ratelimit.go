package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	apperrors "github.com/bantuaku/backend/errors"
	"github.com/bantuaku/backend/logger"
	"github.com/bantuaku/backend/services/storage"
)

// RateLimitConfig defines rate limiting configuration
type RateLimitConfig struct {
	RequestsPerMinute int
	BurstSize         int
	KeyFunc           func(*http.Request) string // Function to generate rate limit key
}

// RateLimiter implements token bucket rate limiting using Redis
func RateLimiter(redis *storage.Redis, config RateLimitConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestID, _ := r.Context().Value(RequestIDKey).(string)
			log := logger.With("request_id", requestID)

			// Generate rate limit key
			key := config.KeyFunc(r)
			if key == "" {
				// If no key can be generated, allow the request
				next.ServeHTTP(w, r)
				return
			}

			// Check rate limit using Redis
			allowed, remaining, resetTime, err := checkRateLimit(r.Context(), redis, key, config.RequestsPerMinute, config.BurstSize)
			if err != nil {
				// If Redis is unavailable, log error but allow request (fail open)
				log.Warn("Rate limit check failed", "error", err)
				next.ServeHTTP(w, r)
				return
			}

			// Set rate limit headers
			w.Header().Set("X-RateLimit-Limit", strconv.Itoa(config.RequestsPerMinute))
			w.Header().Set("X-RateLimit-Remaining", strconv.Itoa(remaining))
			w.Header().Set("X-RateLimit-Reset", strconv.FormatInt(resetTime, 10))

			if !allowed {
				err := apperrors.NewAppError(
					apperrors.ErrCodeLimitExceeded,
					fmt.Sprintf("Rate limit exceeded. Maximum %d requests per minute.", config.RequestsPerMinute),
					"",
				)
				log.Warn("Rate limit exceeded", "key", key, "limit", config.RequestsPerMinute)
				apperrors.WriteJSONError(w, err, err.Code)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// checkRateLimit implements token bucket algorithm using Redis
func checkRateLimit(ctx context.Context, redis *storage.Redis, key string, requestsPerMinute, burstSize int) (allowed bool, remaining int, resetTime int64, err error) {
	if redis == nil {
		// If Redis is not available, allow all requests
		return true, requestsPerMinute, time.Now().Add(time.Minute).Unix(), nil
	}

	redisKey := fmt.Sprintf("ratelimit:%s", key)
	now := time.Now()
	windowStart := now.Truncate(time.Minute).Unix()

	// Use Redis INCR with expiration
	// First, increment the counter
	incrCmd := redis.Client().Incr(ctx, redisKey)
	if incrCmd.Err() != nil {
		return false, 0, 0, incrCmd.Err()
	}

	count := int(incrCmd.Val())

	// Set expiration if this is the first request in the window
	if count == 1 {
		redis.Client().Expire(ctx, redisKey, time.Minute)
	}

	// Check if limit exceeded
	limit := requestsPerMinute
	if count > limit {
		// Check burst allowance
		if count > limit+burstSize {
			return false, 0, windowStart + 60, nil
		}
	}

	remaining = limit - count
	if remaining < 0 {
		remaining = 0
	}

	resetTime = windowStart + 60

	return true, remaining, resetTime, nil
}

// KeyByIP generates rate limit key based on client IP
func KeyByIP(r *http.Request) string {
	ip := getClientIP(r)
	if ip == "" {
		return ""
	}
	return fmt.Sprintf("ip:%s", ip)
}

// KeyByUser generates rate limit key based on authenticated user ID
func KeyByUser(r *http.Request) string {
	userID, ok := r.Context().Value(UserIDKey).(string)
	if !ok || userID == "" {
		return ""
	}
	return fmt.Sprintf("user:%s", userID)
}

// KeyByUserAndPath generates rate limit key based on user ID and path
func KeyByUserAndPath(r *http.Request) string {
	userID, ok := r.Context().Value(UserIDKey).(string)
	if !ok || userID == "" {
		return KeyByIP(r)
	}
	return fmt.Sprintf("user:%s:path:%s", userID, r.URL.Path)
}

// DefaultRateLimitConfigs provides common rate limit configurations
var DefaultRateLimitConfigs = struct {
	Scraping    RateLimitConfig
	Trends      RateLimitConfig
	Admin       RateLimitConfig
	API         RateLimitConfig
}{
	Scraping: RateLimitConfig{
		RequestsPerMinute: 10,  // 10 scraping requests per minute
		BurstSize:         5,   // Allow 5 extra requests
		KeyFunc:           KeyByUser,
	},
	Trends: RateLimitConfig{
		RequestsPerMinute: 20,  // 20 trends requests per minute
		BurstSize:         10,  // Allow 10 extra requests
		KeyFunc:           KeyByUser,
	},
	Admin: RateLimitConfig{
		RequestsPerMinute: 60,  // 60 admin requests per minute
		BurstSize:         20,  // Allow 20 extra requests
		KeyFunc:           KeyByUser,
	},
	API: RateLimitConfig{
		RequestsPerMinute: 100, // 100 API requests per minute
		BurstSize:         50,  // Allow 50 extra requests
		KeyFunc:           KeyByUserAndPath,
	},
}

