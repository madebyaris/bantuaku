package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/bantuaku/backend/config"
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
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "ok",
		"service": "bantuaku-api",
	})
}

// respondJSON sends a JSON response
func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if data != nil {
		json.NewEncoder(w).Encode(data)
	}
}

// respondError sends an error response
func respondError(w http.ResponseWriter, status int, message string) {
	respondJSON(w, status, map[string]string{"error": message})
}

// parseJSON parses JSON request body
func parseJSON(r *http.Request, v interface{}) error {
	return json.NewDecoder(r.Body).Decode(v)
}
