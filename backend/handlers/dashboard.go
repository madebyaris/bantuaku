package handlers

import (
	"net/http"
	"time"

	"github.com/bantuaku/backend/middleware"
	"github.com/bantuaku/backend/models"
)

// DashboardSummary returns the main dashboard KPIs
func (h *Handler) DashboardSummary(w http.ResponseWriter, r *http.Request) {
	companyID := middleware.GetCompanyID(r.Context())
	if companyID == "" {
		respondError(w, http.StatusUnauthorized, "Company not found in context")
		return
	}

	ctx := r.Context()
	summary := models.DashboardSummary{}

	// Get company info
	h.db.Pool().QueryRow(ctx, `
		SELECT name, industry, 
		       COALESCE(city || ', ' || location_region, location_region, '') as location
		FROM companies 
		WHERE id = $1
	`, companyID).Scan(&summary.CompanyName, &summary.CompanyIndustry, &summary.CompanyLocation)

	// Revenue this month
	firstOfMonth := time.Now().AddDate(0, 0, -time.Now().Day()+1)
	h.db.Pool().QueryRow(ctx, `
		SELECT COALESCE(SUM(quantity * price), 0)
		FROM sales_history
		WHERE company_id = $1 AND sale_date >= $2
	`, companyID, firstOfMonth).Scan(&summary.RevenueThisMonth)

	// Revenue last month for trend calculation
	firstOfLastMonth := firstOfMonth.AddDate(0, -1, 0)
	var lastMonthRevenue float64
	h.db.Pool().QueryRow(ctx, `
		SELECT COALESCE(SUM(quantity * price), 0)
		FROM sales_history
		WHERE company_id = $1 AND sale_date >= $2 AND sale_date < $3
	`, companyID, firstOfLastMonth, firstOfMonth).Scan(&lastMonthRevenue)

	// Calculate trend
	if lastMonthRevenue > 0 {
		summary.RevenueTrend = ((summary.RevenueThisMonth - lastMonthRevenue) / lastMonthRevenue) * 100
	}

	// Top selling product this month
	h.db.Pool().QueryRow(ctx, `
		SELECT p.name
		FROM products p
		JOIN sales_history s ON p.id = s.product_id
		WHERE p.company_id = $1 AND s.sale_date >= $2
		GROUP BY p.id, p.name
		ORDER BY SUM(s.quantity) DESC
		LIMIT 1
	`, companyID, firstOfMonth).Scan(&summary.TopSellingProduct)

	// Total conversations
	h.db.Pool().QueryRow(ctx, `
		SELECT COUNT(*) FROM conversations WHERE company_id = $1
	`, companyID).Scan(&summary.TotalConversations)

	// Total insights
	h.db.Pool().QueryRow(ctx, `
		SELECT COUNT(*) FROM insights WHERE company_id = $1
	`, companyID).Scan(&summary.TotalInsights)

	// Total file uploads
	h.db.Pool().QueryRow(ctx, `
		SELECT COUNT(*) FROM file_uploads WHERE company_id = $1
	`, companyID).Scan(&summary.TotalFileUploads)

	// Insights summary by type
	h.db.Pool().QueryRow(ctx, `
		SELECT 
			COUNT(*) FILTER (WHERE type = 'forecast') as forecast_count,
			COUNT(*) FILTER (WHERE type = 'market_prediction') as market_count,
			COUNT(*) FILTER (WHERE type = 'marketing_recommendation') as marketing_count,
			COUNT(*) FILTER (WHERE type = 'gov_regulation') as regulation_count
		FROM insights
		WHERE company_id = $1
	`, companyID).Scan(
		&summary.InsightsSummary.Forecast,
		&summary.InsightsSummary.Market,
		&summary.InsightsSummary.Marketing,
		&summary.InsightsSummary.Regulation,
	)

	// Recent conversations (last 5)
	rows, err := h.db.Pool().Query(ctx, `
		SELECT id, COALESCE(title, 'Percakapan') as title, updated_at
		FROM conversations
		WHERE company_id = $1
		ORDER BY updated_at DESC
		LIMIT 5
	`, companyID)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var conv models.ConversationSummary
			var updatedAt time.Time
			if rows.Scan(&conv.ID, &conv.Title, &updatedAt) == nil {
				conv.UpdatedAt = updatedAt.Format(time.RFC3339)
				// Get last message preview
				var lastMessage string
				h.db.Pool().QueryRow(ctx, `
					SELECT content FROM messages 
					WHERE conversation_id = $1 
					ORDER BY created_at DESC 
					LIMIT 1
				`, conv.ID).Scan(&lastMessage)
				if len(lastMessage) > 50 {
					lastMessage = lastMessage[:50] + "..."
				}
				conv.LastMessage = lastMessage
				summary.RecentConversations = append(summary.RecentConversations, conv)
			}
		}
	}

	// Recent file uploads (last 5)
	rows2, err := h.db.Pool().Query(ctx, `
		SELECT id, original_filename, source_type, status, created_at
		FROM file_uploads
		WHERE company_id = $1
		ORDER BY created_at DESC
		LIMIT 5
	`, companyID)
	if err == nil {
		defer rows2.Close()
		for rows2.Next() {
			var file models.FileUploadSummary
			var createdAt time.Time
			if rows2.Scan(&file.ID, &file.OriginalFilename, &file.SourceType, &file.Status, &createdAt) == nil {
				file.CreatedAt = createdAt.Format(time.RFC3339)
				summary.RecentFileUploads = append(summary.RecentFileUploads, file)
			}
		}
	}

	respondJSON(w, http.StatusOK, summary)
}
