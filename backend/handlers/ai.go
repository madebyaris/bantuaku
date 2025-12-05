package handlers

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/bantuaku/backend/services/kolosal"

	"github.com/bantuaku/backend/middleware"
	"github.com/bantuaku/backend/models"
)

// AIAnalyze handles AI analysis questions
func (h *Handler) AIAnalyze(w http.ResponseWriter, r *http.Request) {
	storeID := middleware.GetStoreID(r.Context())
	if storeID == "" {
		respondError(w, http.StatusUnauthorized, "Store not found in context")
		return
	}

	var req models.AIAnalyzeRequest
	if err := parseJSON(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if strings.TrimSpace(req.Question) == "" {
		respondError(w, http.StatusBadRequest, "Question is required")
		return
	}

	// Check cache
	cacheKey := fmt.Sprintf("ai:%s:%s", storeID, hashQuestion(req.Question))
	cached, err := h.redis.Get(r.Context(), cacheKey)
	if err == nil && cached != "" {
		var response models.AIAnalyzeResponse
		if json.Unmarshal([]byte(cached), &response) == nil {
			respondJSON(w, http.StatusOK, response)
			return
		}
	}

	// Gather context data
	ctx := r.Context()
	storeContext := h.gatherStoreContext(ctx, storeID)

	// Build prompt
	systemPrompt := buildSystemPrompt()
	userPrompt := buildUserPrompt(req.Question, storeContext)

	// Call Kolosal.ai (or return mock response if no API key)
	var answer string
	var confidence float64
	dataSources := []string{}

	if h.config.KolosalAPIKey != "" {
		client := kolosal.NewClient(h.config.KolosalAPIKey)

		resp, err := client.CreateChatCompletion(ctx, kolosal.ChatCompletionRequest{
			Model: "default", // Use default model from Kolosal.ai
			Messages: []kolosal.ChatCompletionMessage{
				{Role: "system", Content: systemPrompt},
				{Role: "user", Content: userPrompt},
			},
			MaxTokens:   1000,
			Temperature: 0.7,
		})

		if err != nil {
			// Fallback to mock response on error
			answer, confidence, dataSources = generateMockResponse(req.Question, storeContext)
		} else {
			answer = resp.Choices[0].Message.Content
			confidence = 0.85
			dataSources = []string{"sales_history", "forecasts"}
		}
	} else {
		// No API key, use mock response
		answer, confidence, dataSources = generateMockResponse(req.Question, storeContext)
	}

	response := models.AIAnalyzeResponse{
		Answer:      answer,
		Confidence:  confidence,
		DataSources: dataSources,
	}

	// Cache for 24 hours
	cacheData, _ := json.Marshal(response)
	h.redis.Set(ctx, cacheKey, string(cacheData), 24*time.Hour)

	respondJSON(w, http.StatusOK, response)
}

// StoreContext holds contextual data for AI
type StoreContext struct {
	StoreName     string
	TotalProducts int
	TopProducts   []ProductSummary
	RecentRevenue float64
	ForecastData  string
}

type ProductSummary struct {
	Name        string
	Sales30d    int
	Forecast30d int
}

func (h *Handler) gatherStoreContext(ctx context.Context, storeID string) StoreContext {
	sc := StoreContext{}

	// Get store name
	h.db.Pool().QueryRow(ctx, `SELECT name FROM companies WHERE id = $1`, storeID).Scan(&sc.StoreName)

	// Get total products
	h.db.Pool().QueryRow(ctx, `SELECT COUNT(*) FROM products WHERE company_id = $1`, storeID).Scan(&sc.TotalProducts)

	// Get top products with sales
	rows2, _ := h.db.Pool().Query(ctx, `
		SELECT p.name, COALESCE(SUM(s.quantity), 0) as sales
		FROM products p
		LEFT JOIN sales_history s ON p.id = s.product_id AND s.sale_date >= $2
		WHERE p.company_id = $1
		GROUP BY p.id, p.name
		ORDER BY sales DESC
		LIMIT 5
	`, storeID, time.Now().AddDate(0, 0, -30))
	if rows2 != nil {
		defer rows2.Close()
		for rows2.Next() {
			var ps ProductSummary
			if rows2.Scan(&ps.Name, &ps.Sales30d) == nil {
				ps.Forecast30d = int(float64(ps.Sales30d) * 1.1) // Simple projection
				sc.TopProducts = append(sc.TopProducts, ps)
			}
		}
	}

	// Get recent revenue
	h.db.Pool().QueryRow(ctx, `
		SELECT COALESCE(SUM(quantity * price), 0)
		FROM sales_history
		WHERE company_id = $1 AND sale_date >= $2
	`, storeID, time.Now().AddDate(0, 0, -30)).Scan(&sc.RecentRevenue)

	return sc
}

func buildSystemPrompt() string {
	return `Kamu adalah Asisten Bantuaku, AI assistant untuk membantu UMKM Indonesia membuat keputusan bisnis berbasis data.

Panduan:
1. SELALU jawab dalam Bahasa Indonesia yang natural dan ramah
2. Berikan saran yang praktis dan actionable
3. Gunakan data yang tersedia untuk mendukung rekomendasi
4. Jika tidak yakin, sampaikan dengan jujur
5. Format jawaban dengan bullet points untuk kemudahan baca
6. Akhiri dengan satu pertanyaan follow-up untuk membantu lebih lanjut

Konteks: Kamu membantu pemilik UMKM dengan:
- Forecasting permintaan produk berdasarkan data penjualan
- Analisis penjualan dan tren
- Insight pasar dan sentiment`
}

func buildUserPrompt(question string, sc StoreContext) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("Toko: %s\n", sc.StoreName))
	sb.WriteString(fmt.Sprintf("Total Produk: %d\n", sc.TotalProducts))
	sb.WriteString(fmt.Sprintf("Revenue 30 hari: Rp %.0f\n", sc.RecentRevenue))

	if len(sc.TopProducts) > 0 {
		sb.WriteString("\nTop Produk (30 hari terakhir):\n")
		for _, p := range sc.TopProducts {
			sb.WriteString(fmt.Sprintf("- %s: Terjual %d, Proyeksi 30 hari %d\n", p.Name, p.Sales30d, p.Forecast30d))
		}
	}

	sb.WriteString(fmt.Sprintf("\nPertanyaan: %s", question))

	return sb.String()
}

func generateMockResponse(question string, sc StoreContext) (string, float64, []string) {
	questionLower := strings.ToLower(question)

	var answer string
	confidence := 0.75
	dataSources := []string{"sales_history"}

	if strings.Contains(questionLower, "order") || strings.Contains(questionLower, "beli") || strings.Contains(questionLower, "permintaan") {
		answer = fmt.Sprintf(`Berdasarkan analisis data penjualan toko %s:

**Proyeksi Permintaan Bulan Depan:**

`, sc.StoreName)

		if len(sc.TopProducts) > 0 {
			for _, p := range sc.TopProducts {
				answer += fmt.Sprintf("• **%s**: Proyeksi permintaan %d unit (berdasarkan penjualan 30 hari: %d unit)\n", p.Name, p.Forecast30d, p.Sales30d)
			}
		} else {
			answer += "Belum ada data penjualan yang cukup untuk membuat proyeksi.\n"
		}

		answer += "\nApakah ada produk tertentu yang ingin Anda analisis lebih detail?"
		dataSources = append(dataSources, "forecasts")

	} else if strings.Contains(questionLower, "turun") || strings.Contains(questionLower, "menurun") || strings.Contains(questionLower, "kenapa") {
		answer = fmt.Sprintf(`Analisis penurunan penjualan untuk %s:

**Kemungkinan Penyebab:**
• Faktor musiman - cek apakah ini periode off-season untuk kategori produk Anda
• Kompetitor - mungkin ada promo dari pesaing
• Perubahan preferensi konsumen - cek tren pasar terkini

**Rekomendasi:**
1. Analisis produk yang mengalami penurunan paling signifikan
2. Pertimbangkan promo untuk produk yang lambat terjual
3. Cek feedback pelanggan di social media

Revenue 30 hari terakhir: Rp %.0f

Mau saya bantu analisis produk tertentu lebih detail?`, sc.StoreName, sc.RecentRevenue)
		dataSources = append(dataSources, "sentiment")

	} else if strings.Contains(questionLower, "trending") || strings.Contains(questionLower, "trend") || strings.Contains(questionLower, "populer") {
		answer = fmt.Sprintf(`Trend pasar untuk toko %s:

**Kategori Trending Saat Ini:**
• Produk ramah lingkungan (+25%% growth)
• Fashion lokal Indonesia (+18%% growth)
• Skincare natural (+32%% growth)

**Saran untuk Toko Anda:**
1. Pertimbangkan menambah produk eco-friendly
2. Highlight "Made in Indonesia" untuk produk lokal
3. Gunakan hashtag trending di promosi

Apakah ada kategori tertentu yang ingin Anda eksplorasi?`, sc.StoreName)
		dataSources = append(dataSources, "market_trends")
		confidence = 0.7

	} else {
		answer = fmt.Sprintf(`Terima kasih atas pertanyaan Anda!

Berikut ringkasan toko %s:
• Total produk: %d
• Revenue 30 hari: Rp %.0f

Saya bisa membantu Anda dengan:
1. **Proyeksi permintaan** - "Berapa proyeksi permintaan bulan depan?"
2. **Analisis penjualan** - "Mengapa penjualan turun?"
3. **Trend pasar** - "Produk apa yang sedang trending?"
4. **Rekomendasi bisnis** - "Apa yang harus saya lakukan untuk meningkatkan penjualan?"

Apa yang ingin Anda ketahui lebih lanjut?`, sc.StoreName, sc.TotalProducts, sc.RecentRevenue)
		confidence = 0.6
	}

	return answer, confidence, dataSources
}

func hashQuestion(question string) string {
	h := sha256.Sum256([]byte(strings.ToLower(strings.TrimSpace(question))))
	return hex.EncodeToString(h[:8])
}
