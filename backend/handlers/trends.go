package handlers

import (
	"net/http"
	"time"

	"github.com/bantuaku/backend/errors"
	"github.com/bantuaku/backend/logger"
	"github.com/bantuaku/backend/services/audit"
	"github.com/bantuaku/backend/services/trends"
)

// CreateKeywordRequest represents a request to create a tracked keyword
type CreateKeywordRequest struct {
	Keyword  string  `json:"keyword" validate:"required"`
	Geo      string  `json:"geo" validate:"required,len=2"`
	Category *string  `json:"category,omitempty"`
}

// CreateKeyword creates a new tracked keyword for a company
func (h *Handler) CreateKeyword(w http.ResponseWriter, r *http.Request) {
	log := logger.With("request_id", r.Context().Value("request_id"))
	companyID := r.Context().Value("company_id").(string)
	if companyID == "" {
		h.respondError(w, errors.NewValidationError("company_id required", ""), r)
		return
	}

	var req CreateKeywordRequest
	if err := h.parseJSON(r, &req); err != nil {
		h.respondError(w, err, r)
		return
	}

	// Validate geo code
	if len(req.Geo) != 2 {
		h.respondError(w, errors.NewValidationError("geo must be 2-letter country code", ""), r)
		return
	}

	// Create store
	store := trends.NewStore(h.db.Pool())

	// Upsert keyword
	keywordID, err := store.UpsertKeyword(r.Context(), companyID, req.Keyword, req.Geo, req.Category)
	if err != nil {
		log.Error("Failed to create keyword", "error", err)
		h.respondError(w, errors.NewInternalError(err, "Failed to create keyword"), r)
		return
	}

	log.Info("Keyword created", "keyword_id", keywordID, "keyword", req.Keyword)

	h.respondJSON(w, http.StatusOK, map[string]interface{}{
		"keyword_id": keywordID,
		"keyword":    req.Keyword,
		"geo":       req.Geo,
		"category":  req.Category,
		"created_at": time.Now(),
	})
}

// ListKeywords lists all tracked keywords for a company
func (h *Handler) ListKeywords(w http.ResponseWriter, r *http.Request) {
	log := logger.With("request_id", r.Context().Value("request_id"))
	companyID := r.Context().Value("company_id").(string)
	if companyID == "" {
		h.respondError(w, errors.NewValidationError("company_id required", ""), r)
		return
	}

	store := trends.NewStore(h.db.Pool())
	keywords, err := store.GetKeywords(r.Context(), companyID)
	if err != nil {
		log.Error("Failed to list keywords", "error", err)
		h.respondError(w, errors.NewInternalError(err, "Failed to list keywords"), r)
		return
	}

	h.respondJSON(w, http.StatusOK, map[string]interface{}{
		"keywords": keywords,
		"count":   len(keywords),
	})
}

// GetTimeSeries retrieves time series data for a keyword
func (h *Handler) GetTimeSeries(w http.ResponseWriter, r *http.Request) {
	log := logger.With("request_id", r.Context().Value("request_id"))
	keywordID := r.URL.Query().Get("keyword_id")
	if keywordID == "" {
		h.respondError(w, errors.NewValidationError("keyword_id is required", ""), r)
		return
	}

	// Parse optional time range
	var startTime, endTime *time.Time
	if startStr := r.URL.Query().Get("start_time"); startStr != "" {
		if t, err := time.Parse(time.RFC3339, startStr); err == nil {
			startTime = &t
		}
	}
	if endStr := r.URL.Query().Get("end_time"); endStr != "" {
		if t, err := time.Parse(time.RFC3339, endStr); err == nil {
			endTime = &t
		}
	}

	store := trends.NewStore(h.db.Pool())
	points, err := store.GetTimeSeries(r.Context(), keywordID, startTime, endTime)
	if err != nil {
		log.Error("Failed to get time series", "error", err)
		h.respondError(w, errors.NewInternalError(err, "Failed to get time series"), r)
		return
	}

	h.respondJSON(w, http.StatusOK, map[string]interface{}{
		"keyword_id": keywordID,
		"time_series": points,
		"count":      len(points),
	})
}

// DeleteKeyword deletes a tracked keyword
func (h *Handler) DeleteKeyword(w http.ResponseWriter, r *http.Request) {
	log := logger.With("request_id", r.Context().Value("request_id"))
	keywordID := r.URL.Query().Get("keyword_id")
	if keywordID == "" {
		h.respondError(w, errors.NewValidationError("keyword_id is required", ""), r)
		return
	}

	// Soft delete by setting is_active = false
	pool := h.db.Pool()
	_, err := pool.Exec(r.Context(),
		"UPDATE trends_keywords SET is_active = false, updated_at = NOW() WHERE id = $1",
		keywordID,
	)
	if err != nil {
		log.Error("Failed to delete keyword", "error", err)
		h.respondError(w, errors.NewInternalError(err, "Failed to delete keyword"), r)
		return
	}

	log.Info("Keyword deleted", "keyword_id", keywordID)

	h.respondJSON(w, http.StatusOK, map[string]interface{}{
		"message": "Keyword deleted successfully",
		"keyword_id": keywordID,
	})
}

// TriggerIngestion triggers manual trends ingestion for a company
func (h *Handler) TriggerIngestion(w http.ResponseWriter, r *http.Request) {
	log := logger.With("request_id", r.Context().Value("request_id"))
	companyID := r.Context().Value("company_id").(string)
	if companyID == "" {
		h.respondError(w, errors.NewValidationError("company_id required", ""), r)
		return
	}

	// Create ingestion service
	ingestService := trends.NewIngestService(h.db.Pool())

	// Create scraper
	scraper := trends.NewGoogleTrendsScraper()

	// Run ingestion in goroutine
	go func() {
		ctx := r.Context()
		if err := ingestService.IngestCompanyKeywords(ctx, companyID, scraper); err != nil {
			log.Error("Trends ingestion failed", "error", err)
		} else {
			log.Info("Trends ingestion completed", "company_id", companyID)
		}
	}()

	log.Info("Trends ingestion triggered", "company_id", companyID)

	// Log audit event for trends ingestion
	if h.auditLogger != nil {
		h.auditLogger.LogAction(r.Context(), r, audit.ActionTrendsIngested, map[string]interface{}{
			"company_id": companyID,
		})
	}

	h.respondJSON(w, http.StatusAccepted, map[string]interface{}{
		"message":   "Trends ingestion started",
		"company_id": companyID,
		"status":    "running",
	})
}

