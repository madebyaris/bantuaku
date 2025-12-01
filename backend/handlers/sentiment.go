package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/bantuaku/backend/middleware"
	"github.com/bantuaku/backend/models"
)

// GetSentiment returns sentiment data for a product
func (h *Handler) GetSentiment(w http.ResponseWriter, r *http.Request) {
	storeID := middleware.GetStoreID(r.Context())
	productID := r.PathValue("product_id")

	if productID == "" {
		respondError(w, http.StatusBadRequest, "Product ID is required")
		return
	}

	// Check cache first
	cacheKey := fmt.Sprintf("sentiment:%s", productID)
	cached, err := h.redis.Get(r.Context(), cacheKey)
	if err == nil && cached != "" {
		var sentiment models.SentimentData
		if json.Unmarshal([]byte(cached), &sentiment) == nil {
			respondJSON(w, http.StatusOK, sentiment)
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

	// For MVP, return curated sample data
	// In production, this would integrate with social media APIs
	sentiment := generateSampleSentiment(productID, productName)

	// Cache for 6 hours
	cacheData, _ := json.Marshal(sentiment)
	h.redis.Set(r.Context(), cacheKey, string(cacheData), 6*time.Hour)

	respondJSON(w, http.StatusOK, sentiment)
}

// GetMarketTrends returns market trend data
func (h *Handler) GetMarketTrends(w http.ResponseWriter, r *http.Request) {
	storeID := middleware.GetStoreID(r.Context())
	if storeID == "" {
		respondError(w, http.StatusUnauthorized, "Store not found in context")
		return
	}

	// Check cache
	cacheKey := fmt.Sprintf("trends:%s", storeID)
	cached, err := h.redis.Get(r.Context(), cacheKey)
	if err == nil && cached != "" {
		var trends []models.MarketTrend
		if json.Unmarshal([]byte(cached), &trends) == nil {
			respondJSON(w, http.StatusOK, trends)
			return
		}
	}

	// Get store's product categories
	rows, err := h.db.Pool().Query(r.Context(), `
		SELECT DISTINCT category FROM products WHERE store_id = $1 AND category != ''
	`, storeID)

	categories := []string{}
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var cat string
			if rows.Scan(&cat) == nil {
				categories = append(categories, cat)
			}
		}
	}

	// Generate sample trends based on categories
	trends := generateSampleTrends(categories)

	// Cache for 24 hours
	cacheData, _ := json.Marshal(trends)
	h.redis.Set(r.Context(), cacheKey, string(cacheData), 24*time.Hour)

	respondJSON(w, http.StatusOK, trends)
}

// Sample data generators for MVP demo

func generateSampleSentiment(productID, productName string) models.SentimentData {
	// Deterministic-ish sample data based on product ID
	hashVal := 0
	for _, c := range productID {
		hashVal += int(c)
	}

	// Generate sentiment score between 0.2 and 0.9
	baseScore := 0.2 + (float64(hashVal%70) / 100.0)

	positiveCount := 15 + (hashVal % 20)
	negativeCount := 2 + (hashVal % 5)
	neutralCount := 5 + (hashVal % 10)

	mentions := []models.Mention{
		{
			Source:    "instagram",
			Text:      fmt.Sprintf("Produk %s bagus banget! Worth it 🔥", productName),
			Sentiment: 0.8,
			Date:      time.Now().AddDate(0, 0, -1),
		},
		{
			Source:    "tiktok",
			Text:      fmt.Sprintf("Review jujur %s - recommended buat yang cari kualitas!", productName),
			Sentiment: 0.7,
			Date:      time.Now().AddDate(0, 0, -2),
		},
		{
			Source:    "review",
			Text:      "Pengiriman cepat, barang sesuai deskripsi. Mantap!",
			Sentiment: 0.9,
			Date:      time.Now().AddDate(0, 0, -3),
		},
		{
			Source:    "instagram",
			Text:      "Lumayan sih, tapi harga agak mahal ya",
			Sentiment: 0.3,
			Date:      time.Now().AddDate(0, 0, -4),
		},
		{
			Source:    "tiktok",
			Text:      fmt.Sprintf("Unboxing %s! Packaging rapi dan aman 📦", productName),
			Sentiment: 0.75,
			Date:      time.Now().AddDate(0, 0, -5),
		},
	}

	return models.SentimentData{
		ProductID:      productID,
		SentimentScore: baseScore,
		PositiveCount:  positiveCount,
		NegativeCount:  negativeCount,
		NeutralCount:   neutralCount,
		RecentMentions: mentions,
	}
}

func generateSampleTrends(categories []string) []models.MarketTrend {
	trends := []models.MarketTrend{
		{
			Name:       "Produk Ramah Lingkungan",
			Category:   "eco-friendly",
			TrendScore: 0.85,
			GrowthRate: 25.5,
			Source:     "google_trends",
		},
		{
			Name:       "Fashion Lokal Indonesia",
			Category:   "fashion",
			TrendScore: 0.78,
			GrowthRate: 18.3,
			Source:     "social_media",
		},
		{
			Name:       "Makanan Sehat",
			Category:   "food",
			TrendScore: 0.72,
			GrowthRate: 15.2,
			Source:     "marketplace",
		},
		{
			Name:       "Skincare Natural",
			Category:   "beauty",
			TrendScore: 0.88,
			GrowthRate: 32.1,
			Source:     "instagram",
		},
		{
			Name:       "Gadget Accessories",
			Category:   "electronics",
			TrendScore: 0.65,
			GrowthRate: 12.4,
			Source:     "marketplace",
		},
	}

	// Add category-specific trends
	for _, cat := range categories {
		trends = append(trends, models.MarketTrend{
			Name:       fmt.Sprintf("Trend %s", cat),
			Category:   cat,
			TrendScore: 0.6 + (float64(len(cat)%30) / 100.0),
			GrowthRate: 10.0 + float64(len(cat)%20),
			Source:     "analysis",
		})
	}

	return trends
}
