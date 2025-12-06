package admin

import (
	"net/http"
	"strconv"
	"time"

	"github.com/bantuaku/backend/errors"
	"github.com/bantuaku/backend/services/chatlogs"
)

// GetChatUsage retrieves aggregate chat usage statistics
func (h *AdminHandler) GetChatUsage(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse query parameters
	var companyID *string
	if cid := r.URL.Query().Get("company_id"); cid != "" {
		companyID = &cid
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

	// Parse pagination for daily logs
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit < 1 || limit > 100 {
		limit = 30
	}

	chatLogsService := chatlogs.NewService(h.db)

	// Get aggregate stats
	stats, err := chatLogsService.GetUsageStats(ctx, companyID, startDate, endDate)
	if err != nil {
		appErr := errors.NewDatabaseError(err, "get chat usage stats")
		h.respondError(w, appErr, r)
		return
	}

	// Get daily logs
	logs, total, err := chatLogsService.GetDailyLogs(ctx, companyID, startDate, endDate, page, limit)
	if err != nil {
		appErr := errors.NewDatabaseError(err, "get daily logs")
		h.respondError(w, appErr, r)
		return
	}

	h.respondJSON(w, http.StatusOK, map[string]interface{}{
		"stats":      stats,
		"daily_logs": logs,
		"pagination": map[string]interface{}{
			"page":  page,
			"limit": limit,
			"total": total,
		},
	})
}
