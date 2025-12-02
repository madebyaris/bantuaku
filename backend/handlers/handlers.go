package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/bantuaku/backend/config"
	"github.com/bantuaku/backend/errors"
	"github.com/bantuaku/backend/logger"
	"github.com/bantuaku/backend/services/storage"
)

// Handler holds dependencies for HTTP handlers
type Handler struct {
	db     *storage.Postgres
	redis  *storage.Redis
	config *config.Config
}

// New creates a new Handler with dependencies
func New(db *storage.Postgres, redis *storage.Redis, cfg *config.Config) *Handler {
	return &Handler{
		db:     db,
		redis:  redis,
		config: cfg,
	}
}

// HealthCheck returns the health status of the API
func (h *Handler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	// Create contextual logger
	log := logger.With("request_id", r.Context().Value("request_id"))

	log.Info("Health check requested")

	respondJSON(w, http.StatusOK, map[string]string{
		"status":  "ok",
		"service": "bantuaku-api",
	})
}

// respondJSON sends a JSON response
func (h *Handler) respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if data != nil {
		json.NewEncoder(w).Encode(data)
	}
}

// respondError sends an error response with proper logging
func (h *Handler) respondError(w http.ResponseWriter, err error, r *http.Request) {
	// Create contextual logger
	log := logger.With("request_id", r.Context().Value("request_id"))

	// Log the error
	log.LogError(err, "Handler error", r.Context())

	// Write JSON error response
	errors.WriteJSONError(w, err, errors.GetErrorCode(err))
}

// parseJSON parses JSON request body with error handling
func (h *Handler) parseJSON(r *http.Request, v interface{}) error {
	if err := json.NewDecoder(r.Body).Decode(v); err != nil {
		return errors.NewValidationError("Invalid JSON format", err.Error())
	}
	return nil
}

// Mock respondJSON 函数以保持向后兼容性
func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if data != nil {
		json.NewEncoder(w).Encode(data)
	}
}

// Mock respondError 函数以保持向后兼容性
func respondError(w http.ResponseWriter, status int, message string) {
	errors.WriteJSONError(w, errors.NewAppError(errors.ErrCodeInternal, message, ""), errors.ErrCodeInternal)
}

// Mock parseJSON 函数以保持向后兼容性
func parseJSON(r *http.Request, v interface{}) error {
	return json.NewDecoder(r.Body).Decode(v)
}
