package handlers

import (
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"time"

	"github.com/bantuaku/backend/middleware"
	"github.com/bantuaku/backend/models"
	"github.com/google/uuid"
)

// ForecastResponse represents a forecast response with additional context
type ForecastResponse struct {
	models.Forecast
	ProductName     string       `json:"product_name"`
	HistoricalSales []DailySales `json:"historical_sales,omitempty"`
}

// DailySales represents aggregated daily sales
type DailySales struct {
	Date     string `json:"date"`
	Quantity int    `json:"quantity"`
}

// GetForecast returns the forecast for a specific product
func (h *Handler) GetForecast(w http.ResponseWriter, r *http.Request) {
	storeID := middleware.GetStoreID(r.Context())
	productID := r.PathValue("product_id")

	if productID == "" {
		respondError(w, http.StatusBadRequest, "Product ID is required")
		return
	}

	// Check cache first
	cacheKey := fmt.Sprintf("forecast:%s", productID)
	cached, err := h.redis.Get(r.Context(), cacheKey)
	if err == nil && cached != "" {
		var forecast ForecastResponse
		if json.Unmarshal([]byte(cached), &forecast) == nil {
			respondJSON(w, http.StatusOK, forecast)
			return
		}
	}

	// Verify product belongs to store
	var productName string
	err = h.db.Pool().QueryRow(r.Context(), `
		SELECT product_name FROM products WHERE id = $1 AND store_id = $2
	`, productID, storeID).Scan(&productName)
	if err != nil {
		respondError(w, http.StatusNotFound, "Product not found")
		return
	}

	// Get historical sales (last 90 days)
	rows, err := h.db.Pool().Query(r.Context(), `
		SELECT sale_date, SUM(quantity) as total_qty
		FROM sales_history
		WHERE product_id = $1 AND store_id = $2 AND sale_date >= $3
		GROUP BY sale_date
		ORDER BY sale_date ASC
	`, productID, storeID, time.Now().AddDate(0, 0, -90))
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to fetch sales history")
		return
	}
	defer rows.Close()

	var salesData []float64
	var historicalSales []DailySales
	for rows.Next() {
		var date time.Time
		var qty int
		if rows.Scan(&date, &qty) == nil {
			salesData = append(salesData, float64(qty))
			historicalSales = append(historicalSales, DailySales{
				Date:     date.Format("2006-01-02"),
				Quantity: qty,
			})
		}
	}

	// Calculate forecast
	var forecast30d, forecast60d, forecast90d int
	var confidence float64
	algorithm := "ensemble"

	if len(salesData) >= 7 {
		// Simple Moving Average (7-day)
		sma := simpleMovingAverage(salesData, 7)

		// Exponential Smoothing
		es := exponentialSmoothing(salesData, 0.3)

		// Trend extraction
		trend := trendExtraction(salesData)

		// Ensemble prediction
		predicted := sma*0.4 + es*0.35 + trend*0.25

		forecast30d = int(math.Round(predicted * 30))
		forecast60d = int(math.Round(predicted * 60))
		forecast90d = int(math.Round(predicted * 90))

		confidence = calculateConfidence(salesData, predicted)
	} else if len(salesData) > 0 {
		// Simple average for limited data
		sum := 0.0
		for _, v := range salesData {
			sum += v
		}
		avg := sum / float64(len(salesData))

		forecast30d = int(math.Round(avg * 30))
		forecast60d = int(math.Round(avg * 60))
		forecast90d = int(math.Round(avg * 90))
		confidence = 0.5 // Lower confidence for limited data
		algorithm = "simple_average"
	} else {
		// No data
		forecast30d = 0
		forecast60d = 0
		forecast90d = 0
		confidence = 0
		algorithm = "no_data"
	}

	forecastResp := ForecastResponse{
		Forecast: models.Forecast{
			ID:          uuid.New().String(),
			ProductID:   productID,
			Forecast30d: forecast30d,
			Forecast60d: forecast60d,
			Forecast90d: forecast90d,
			Confidence:  confidence,
			Algorithm:   algorithm,
			GeneratedAt: time.Now(),
			ExpiresAt:   time.Now().Add(time.Hour),
		},
		ProductName:     productName,
		HistoricalSales: historicalSales,
	}

	// Cache the result
	cacheData, _ := json.Marshal(forecastResp)
	h.redis.Set(r.Context(), cacheKey, string(cacheData), time.Hour)

	respondJSON(w, http.StatusOK, forecastResp)
}

// GetRecommendations returns demand forecast recommendations for all products
func (h *Handler) GetRecommendations(w http.ResponseWriter, r *http.Request) {
	storeID := middleware.GetStoreID(r.Context())
	if storeID == "" {
		respondError(w, http.StatusUnauthorized, "Store not found in context")
		return
	}

	// Get all products with their sales data
	rows, err := h.db.Pool().Query(r.Context(), `
		SELECT p.id, p.product_name,
			COALESCE(SUM(s.quantity), 0) as total_sales,
			COUNT(DISTINCT s.sale_date) as days_with_sales
		FROM products p
		LEFT JOIN sales_history s ON p.id = s.product_id 
			AND s.sale_date >= $2
		WHERE p.store_id = $1
		GROUP BY p.id, p.product_name
		ORDER BY total_sales DESC
	`, storeID, time.Now().AddDate(0, 0, -30))
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to fetch recommendations")
		return
	}
	defer rows.Close()

	recommendations := []models.Recommendation{}
	for rows.Next() {
		var productID, productName string
		var totalSales, daysWithSales int

		if rows.Scan(&productID, &productName, &totalSales, &daysWithSales) != nil {
			continue
		}

		// Calculate projected demand
		avgDailySales := 0.0
		if daysWithSales > 0 {
			avgDailySales = float64(totalSales) / float64(daysWithSales)
		}

		// Projected 30-day demand
		projected30d := int(math.Ceil(avgDailySales * 30))

		// Determine risk level based on sales trend
		var riskLevel, reason string
		if projected30d == 0 {
			riskLevel = "low"
			reason = "Tidak ada proyeksi permintaan berdasarkan data penjualan."
		} else if projected30d < 10 {
			riskLevel = "low"
			reason = fmt.Sprintf("Proyeksi permintaan rendah: %d unit dalam 30 hari ke depan.", projected30d)
		} else if projected30d < 50 {
			riskLevel = "medium"
			reason = fmt.Sprintf("Proyeksi permintaan sedang: %d unit dalam 30 hari ke depan.", projected30d)
		} else {
			riskLevel = "high"
			reason = fmt.Sprintf("Proyeksi permintaan tinggi: %d unit dalam 30 hari ke depan. Pastikan ketersediaan produk.", projected30d)
		}

		recommendations = append(recommendations, models.Recommendation{
			ProductID:       productID,
			ProductName:     productName,
			ProjectedDemand: projected30d,
			Reason:          reason,
			RiskLevel:       riskLevel,
		})
	}

	respondJSON(w, http.StatusOK, recommendations)
}

// Forecasting helper functions

func simpleMovingAverage(data []float64, period int) float64 {
	if len(data) < period {
		period = len(data)
	}
	if period == 0 {
		return 0
	}

	sum := 0.0
	for i := len(data) - period; i < len(data); i++ {
		sum += data[i]
	}
	return sum / float64(period)
}

func exponentialSmoothing(data []float64, alpha float64) float64 {
	if len(data) == 0 {
		return 0
	}

	result := data[0]
	for i := 1; i < len(data); i++ {
		result = alpha*data[i] + (1-alpha)*result
	}
	return result
}

func trendExtraction(data []float64) float64 {
	n := float64(len(data))
	if n == 0 {
		return 0
	}

	// Calculate average
	sum := 0.0
	for _, v := range data {
		sum += v
	}
	avg := sum / n

	// Calculate trend
	sumNumerator := 0.0
	sumDenominator := 0.0

	for i, v := range data {
		x := float64(i) - (n-1)/2
		sumNumerator += x * (v - avg)
		sumDenominator += x * x
	}

	trend := 0.0
	if sumDenominator != 0 {
		trend = sumNumerator / sumDenominator
	}

	// Project forward (average + trend for 7 days)
	lastValue := data[len(data)-1]
	return lastValue + trend*7
}

func calculateConfidence(data []float64, forecast float64) float64 {
	if len(data) == 0 || forecast == 0 {
		return 0
	}

	// Calculate average
	sum := 0.0
	for _, v := range data {
		sum += v
	}
	avg := sum / float64(len(data))

	if avg == 0 {
		return 0.5
	}

	// Calculate standard deviation
	variance := 0.0
	for _, v := range data {
		variance += (v - avg) * (v - avg)
	}
	stdDev := math.Sqrt(variance / float64(len(data)))

	// Coefficient of variation
	cv := stdDev / avg

	// Confidence: lower CV = higher confidence
	confidence := math.Max(0, math.Min(1, 1-cv))

	return confidence
}
