package admin

import (
	"net/http"
	"strconv"
	"time"

	"github.com/bantuaku/backend/errors"
)

// GetTokenUsageAggregates returns aggregated token usage with filters.
// Filters: company_id, user_id, model, provider, start_date, end_date. Pagination supported.
func (h *AdminHandler) GetTokenUsageAggregates(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var companyID, userID, model, provider *string
	if v := r.URL.Query().Get("company_id"); v != "" {
		companyID = &v
	}
	if v := r.URL.Query().Get("user_id"); v != "" {
		userID = &v
	}
	if v := r.URL.Query().Get("model"); v != "" {
		model = &v
	}
	if v := r.URL.Query().Get("provider"); v != "" {
		provider = &v
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
	// Default to last 30 days
	if startDate == nil && endDate == nil {
		now := time.Now()
		thirty := now.AddDate(0, 0, -30)
		startDate = &thirty
		endDate = &now
	}

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit < 1 || limit > 200 {
		limit = 50
	}
	offset := (page - 1) * limit

	query := `
		SELECT date, user_id, company_id, model, provider,
			   prompt_tokens, completion_tokens, total_tokens
		FROM token_usage_aggregates
		WHERE 1=1
	`
	args := []interface{}{}
	arg := 1
	if companyID != nil {
		query += ` AND company_id = $` + strconv.Itoa(arg)
		args = append(args, *companyID)
		arg++
	}
	if userID != nil {
		query += ` AND user_id = $` + strconv.Itoa(arg)
		args = append(args, *userID)
		arg++
	}
	if model != nil {
		query += ` AND model = $` + strconv.Itoa(arg)
		args = append(args, *model)
		arg++
	}
	if provider != nil {
		query += ` AND provider = $` + strconv.Itoa(arg)
		args = append(args, *provider)
		arg++
	}
	if startDate != nil {
		query += ` AND date >= $` + strconv.Itoa(arg)
		args = append(args, startDate.Format("2006-01-02"))
		arg++
	}
	if endDate != nil {
		query += ` AND date <= $` + strconv.Itoa(arg)
		args = append(args, endDate.Format("2006-01-02"))
		arg++
	}
	query += ` ORDER BY date DESC LIMIT $` + strconv.Itoa(arg) + ` OFFSET $` + strconv.Itoa(arg+1)
	args = append(args, limit, offset)

	rows, err := h.db.Pool().Query(ctx, query, args...)
	if err != nil {
		h.respondError(w, errors.NewDatabaseError(err, "query token usage aggregates"), r)
		return
	}
	defer rows.Close()

	type row struct {
		Date            string  `json:"date"`
		UserID          *string `json:"user_id,omitempty"`
		CompanyID       *string `json:"company_id,omitempty"`
		Model           string  `json:"model"`
		Provider        string  `json:"provider"`
		PromptTokens    int     `json:"prompt_tokens"`
		CompletionTokens int    `json:"completion_tokens"`
		TotalTokens     int     `json:"total_tokens"`
	}
	var data []row
	for rows.Next() {
		var rec row
		if err := rows.Scan(
			&rec.Date, &rec.UserID, &rec.CompanyID, &rec.Model, &rec.Provider,
			&rec.PromptTokens, &rec.CompletionTokens, &rec.TotalTokens,
		); err == nil {
			data = append(data, rec)
		}
	}

	// Count query
	countQuery := `SELECT COUNT(*) FROM token_usage_aggregates WHERE 1=1`
	countArgs := []interface{}{}
	carg := 1
	if companyID != nil {
		countQuery += ` AND company_id = $` + strconv.Itoa(carg)
		countArgs = append(countArgs, *companyID)
		carg++
	}
	if userID != nil {
		countQuery += ` AND user_id = $` + strconv.Itoa(carg)
		countArgs = append(countArgs, *userID)
		carg++
	}
	if model != nil {
		countQuery += ` AND model = $` + strconv.Itoa(carg)
		countArgs = append(countArgs, *model)
		carg++
	}
	if provider != nil {
		countQuery += ` AND provider = $` + strconv.Itoa(carg)
		countArgs = append(countArgs, *provider)
		carg++
	}
	if startDate != nil {
		countQuery += ` AND date >= $` + strconv.Itoa(carg)
		countArgs = append(countArgs, startDate.Format("2006-01-02"))
		carg++
	}
	if endDate != nil {
		countQuery += ` AND date <= $` + strconv.Itoa(carg)
		countArgs = append(countArgs, endDate.Format("2006-01-02"))
		carg++
	}

	var total int
	if err := h.db.Pool().QueryRow(ctx, countQuery, countArgs...).Scan(&total); err != nil {
		total = 0
	}

	h.respondJSON(w, http.StatusOK, map[string]interface{}{
		"data": data,
		"pagination": map[string]interface{}{
			"page":  page,
			"limit": limit,
			"total": total,
		},
	})
}

