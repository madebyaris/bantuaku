package handlers

import (
	"net/http"
	"time"

	"github.com/bantuaku/backend/middleware"
	"github.com/bantuaku/backend/models"
)

// DashboardSummary returns the main dashboard KPIs
func (h *Handler) DashboardSummary(w http.ResponseWriter, r *http.Request) {
	storeID := middleware.GetStoreID(r.Context())
	if storeID == "" {
		respondError(w, http.StatusUnauthorized, "Store not found in context")
		return
	}

	ctx := r.Context()
	summary := models.DashboardSummary{}

	// Total products
	h.db.Pool().QueryRow(ctx, `
		SELECT COUNT(*) FROM products WHERE store_id = $1
	`, storeID).Scan(&summary.TotalProducts)

	// Low stock count (< 10 units)
	h.db.Pool().QueryRow(ctx, `
		SELECT COUNT(*) FROM products WHERE store_id = $1 AND stock < 10
	`, storeID).Scan(&summary.LowStockCount)

	// Revenue this month
	firstOfMonth := time.Now().AddDate(0, 0, -time.Now().Day()+1)
	h.db.Pool().QueryRow(ctx, `
		SELECT COALESCE(SUM(quantity * price), 0)
		FROM sales_history
		WHERE store_id = $1 AND sale_date >= $2
	`, storeID, firstOfMonth).Scan(&summary.RevenueThisMonth)

	// Revenue last month for trend calculation
	firstOfLastMonth := firstOfMonth.AddDate(0, -1, 0)
	var lastMonthRevenue float64
	h.db.Pool().QueryRow(ctx, `
		SELECT COALESCE(SUM(quantity * price), 0)
		FROM sales_history
		WHERE store_id = $1 AND sale_date >= $2 AND sale_date < $3
	`, storeID, firstOfLastMonth, firstOfMonth).Scan(&lastMonthRevenue)

	// Calculate trend
	if lastMonthRevenue > 0 {
		summary.RevenueTrend = ((summary.RevenueThisMonth - lastMonthRevenue) / lastMonthRevenue) * 100
	}

	// Top selling product this month
	h.db.Pool().QueryRow(ctx, `
		SELECT p.product_name
		FROM products p
		JOIN sales_history s ON p.id = s.product_id
		WHERE p.store_id = $1 AND s.sale_date >= $2
		GROUP BY p.id, p.product_name
		ORDER BY SUM(s.quantity) DESC
		LIMIT 1
	`, storeID, firstOfMonth).Scan(&summary.TopSellingProduct)

	// Forecast accuracy (simplified: compare last month's forecast vs actual)
	// For MVP, use a placeholder value
	summary.ForecastAccuracy = 78.5

	respondJSON(w, http.StatusOK, summary)
}
