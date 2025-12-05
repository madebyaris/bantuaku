package admin

import (
	"net/http"
	"strconv"
	"time"

	"github.com/bantuaku/backend/errors"
	"github.com/bantuaku/backend/services/tokenusage"
)

// GetTokenUsage retrieves token usage statistics
func (h *AdminHandler) GetTokenUsage(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse query parameters
	var companyID *string
	if cid := r.URL.Query().Get("company_id"); cid != "" {
		companyID = &cid
	}

	var model *string
	if m := r.URL.Query().Get("model"); m != "" {
		model = &m
	}

	var startDate, endDate *time.Time
	if sd := r.URL.Query().Get("start_date"); sd != "" {
		if t, err := time.Parse("2006-01-02", sd); err == nil {
			startDate = &t
		}
	}
	if ed := r.URL.Query().Get("end_date"); ed != "" {
		if t, err := time.Parse("2006-01-02", ed); err == nil {
			endDate = &t
		}
	}

	// Default to last 30 days if no dates provided
	if startDate == nil && endDate == nil {
		now := time.Now()
		thirtyDaysAgo := now.AddDate(0, 0, -30)
		startDate = &thirtyDaysAgo
		endDate = &now
	}

	// Parse pagination
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit < 1 || limit > 100 {
		limit = 50
	}

	tokenService := tokenusage.NewService(h.db)

	// Get aggregate stats
	stats, err := tokenService.GetUsageStats(ctx, companyID, model, startDate, endDate)
	if err != nil {
		appErr := errors.NewDatabaseError(err, "get token usage stats")
		h.respondError(w, appErr, r)
		return
	}

	// Get detailed usage records
	usages, total, err := tokenService.GetTokenUsage(ctx, companyID, model, startDate, endDate, page, limit)
	if err != nil {
		appErr := errors.NewDatabaseError(err, "get token usage records")
		h.respondError(w, appErr, r)
		return
	}

	h.respondJSON(w, http.StatusOK, map[string]interface{}{
		"stats": stats,
		"usage": usages,
		"pagination": map[string]interface{}{
			"page":  page,
			"limit": limit,
			"total": total,
		},
	})
}
