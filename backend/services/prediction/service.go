package prediction

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/bantuaku/backend/logger"
	"github.com/bantuaku/backend/services/chat"
	"github.com/bantuaku/backend/services/embedding"
	"github.com/bantuaku/backend/services/exa"
	"github.com/bantuaku/backend/services/forecast"
	"github.com/bantuaku/backend/services/usage"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// JobStatus represents the status of a prediction job
type JobStatus string

const (
	StatusPending    JobStatus = "pending"
	StatusProcessing JobStatus = "processing"
	StatusCompleted  JobStatus = "completed"
	StatusFailed     JobStatus = "failed"
)

// JobStep represents a step in the prediction job
type JobStep string

const (
	StepKeywords    JobStep = "keywords"
	StepSocialMedia JobStep = "social_media"
	StepForecast    JobStep = "forecast"
	StepMarket      JobStep = "market_prediction"
	StepMarketing   JobStep = "marketing"
	StepRegulations JobStep = "regulations"
)

// Progress tracks completion of each step
type Progress struct {
	Keywords    bool `json:"keywords"`
	SocialMedia bool `json:"social_media"`
	Forecast    bool `json:"forecast"`
	Market      bool `json:"market_prediction"`
	Marketing   bool `json:"marketing"`
	Regulations bool `json:"regulations"`
}

// Results contains the output from each research task
type Results struct {
	Keywords          []string               `json:"keywords,omitempty"`
	SocialMediaTrends map[string]interface{} `json:"social_media_trends,omitempty"`
	ForecastSummary   string                 `json:"forecast_summary,omitempty"`
	MarketPrediction  string                 `json:"market_prediction,omitempty"`
	MarketingRecs     string                 `json:"marketing_recommendations,omitempty"`
	Regulations       string                 `json:"regulations,omitempty"`
}

// Job represents a prediction job
type Job struct {
	ID           string     `json:"id"`
	CompanyID    string     `json:"company_id"`
	Status       JobStatus  `json:"status"`
	Progress     Progress   `json:"progress"`
	Results      Results    `json:"results"`
	ErrorMessage string     `json:"error_message,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
	StartedAt    *time.Time `json:"started_at,omitempty"`
	CompletedAt  *time.Time `json:"completed_at,omitempty"`
}

// CompletenessResult shows what's complete and what's missing
type CompletenessResult struct {
	IsComplete  bool     `json:"is_complete"`
	HasIndustry bool     `json:"has_industry"`
	HasCity     bool     `json:"has_city"`
	HasProducts bool     `json:"has_products"`
	HasSocial   bool     `json:"has_social"`
	Missing     []string `json:"missing,omitempty"`
}

// Service handles prediction job orchestration
type Service struct {
	pool            *pgxpool.Pool
	chatProvider    chat.ChatProvider
	forecastService *forecast.Service
	usageService    *usage.Service
	exaClient       *exa.Client
	embedder        embedding.Embedder
	chatModel       string
	log             logger.Logger
}

// NewService creates a new prediction service
func NewService(
	pool *pgxpool.Pool,
	chatProvider chat.ChatProvider,
	forecastService *forecast.Service,
	usageService *usage.Service,
	exaClient *exa.Client,
	embedder embedding.Embedder,
	chatModel string,
) *Service {
	log := logger.Default()
	return &Service{
		pool:            pool,
		chatProvider:    chatProvider,
		forecastService: forecastService,
		usageService:    usageService,
		exaClient:       exaClient,
		embedder:        embedder,
		chatModel:       chatModel,
		log:             *log,
	}
}

// CheckCompleteness validates if company profile is complete enough for predictions
func (s *Service) CheckCompleteness(ctx context.Context, companyID string) (*CompletenessResult, error) {
	result := &CompletenessResult{}

	// Check company fields
	var industry, city, socialMediaStr string
	err := s.pool.QueryRow(ctx, `
		SELECT COALESCE(industry, ''), COALESCE(city, ''), COALESCE(social_media_handles::text, '{}')
		FROM companies WHERE id = $1
	`, companyID).Scan(&industry, &city, &socialMediaStr)
	if err != nil {
		return nil, fmt.Errorf("failed to get company: %w", err)
	}

	result.HasIndustry = industry != ""
	result.HasCity = city != ""

	// Check social media
	var socialMap map[string]interface{}
	if err := json.Unmarshal([]byte(socialMediaStr), &socialMap); err == nil && len(socialMap) > 0 {
		result.HasSocial = true
	}

	// Check products
	var productCount int
	err = s.pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM products WHERE company_id = $1 AND is_active = true
	`, companyID).Scan(&productCount)
	if err != nil {
		return nil, fmt.Errorf("failed to count products: %w", err)
	}
	result.HasProducts = productCount > 0

	// Build missing list
	if !result.HasIndustry {
		result.Missing = append(result.Missing, "industry")
	}
	if !result.HasCity {
		result.Missing = append(result.Missing, "city")
	}
	if !result.HasProducts {
		result.Missing = append(result.Missing, "products")
	}
	if !result.HasSocial {
		result.Missing = append(result.Missing, "social_media")
	}

	result.IsComplete = result.HasIndustry && result.HasCity && result.HasProducts && result.HasSocial

	return result, nil
}

// GetActiveJob returns the current active job for a company (if any)
func (s *Service) GetActiveJob(ctx context.Context, companyID string) (*Job, error) {
	var job Job
	var progressJSON, resultsJSON []byte

	err := s.pool.QueryRow(ctx, `
		SELECT id, company_id, status, progress, results, COALESCE(error_message, ''), 
		       created_at, started_at, completed_at
		FROM prediction_jobs
		WHERE company_id = $1 AND status IN ('pending', 'processing')
		ORDER BY created_at DESC
		LIMIT 1
	`, companyID).Scan(
		&job.ID, &job.CompanyID, &job.Status, &progressJSON, &resultsJSON,
		&job.ErrorMessage, &job.CreatedAt, &job.StartedAt, &job.CompletedAt,
	)
	if err != nil {
		if err.Error() == "no rows in result set" {
			return nil, nil // No active job
		}
		return nil, fmt.Errorf("failed to get active job: %w", err)
	}

	json.Unmarshal(progressJSON, &job.Progress)
	json.Unmarshal(resultsJSON, &job.Results)

	return &job, nil
}

// GetJob returns a specific job by ID
func (s *Service) GetJob(ctx context.Context, jobID string) (*Job, error) {
	var job Job
	var progressJSON, resultsJSON []byte

	err := s.pool.QueryRow(ctx, `
		SELECT id, company_id, status, progress, results, COALESCE(error_message, ''), 
		       created_at, started_at, completed_at
		FROM prediction_jobs
		WHERE id = $1
	`, jobID).Scan(
		&job.ID, &job.CompanyID, &job.Status, &progressJSON, &resultsJSON,
		&job.ErrorMessage, &job.CreatedAt, &job.StartedAt, &job.CompletedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get job: %w", err)
	}

	json.Unmarshal(progressJSON, &job.Progress)
	json.Unmarshal(resultsJSON, &job.Results)

	return &job, nil
}

// PredictionUsage represents usage stats for predictions
type PredictionUsage struct {
	Used      int  `json:"used"`
	Limit     int  `json:"limit"`
	Remaining int  `json:"remaining"`
	Unlimited bool `json:"unlimited"`
}

// GetPredictionUsage returns the prediction usage for a company
func (s *Service) GetPredictionUsage(ctx context.Context, companyID string) (*PredictionUsage, error) {
	limits, err := s.usageService.GetPlanLimits(ctx, companyID)
	if err != nil {
		return nil, err
	}

	// Count prediction jobs this month
	now := time.Now()
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)

	var used int
	err = s.pool.QueryRow(ctx, `
		SELECT COUNT(*) 
		FROM prediction_jobs 
		WHERE company_id = $1 
		AND created_at >= $2
	`, companyID, monthStart).Scan(&used)
	if err != nil {
		used = 0
	}

	// If no limit set, it's unlimited
	if limits.MaxForecastRefreshesPerMonth == nil {
		return &PredictionUsage{
			Used:      used,
			Limit:     0,
			Remaining: 999, // Large number for UI
			Unlimited: true,
		}, nil
	}

	limit := *limits.MaxForecastRefreshesPerMonth
	remaining := limit - used
	if remaining < 0 {
		remaining = 0
	}

	return &PredictionUsage{
		Used:      used,
		Limit:     limit,
		Remaining: remaining,
		Unlimited: false,
	}, nil
}

// StartJob creates and starts a new prediction job
func (s *Service) StartJob(ctx context.Context, companyID string) (*Job, error) {
	// Check usage limits
	allowed, message, err := s.usageService.CheckForecastLimit(ctx, companyID)
	if err != nil {
		return nil, fmt.Errorf("failed to check usage: %w", err)
	}
	if !allowed {
		return nil, fmt.Errorf(message)
	}

	// Check for existing active job
	activeJob, err := s.GetActiveJob(ctx, companyID)
	if err != nil {
		return nil, err
	}
	if activeJob != nil {
		return nil, fmt.Errorf("a prediction job is already in progress")
	}

	// Create new job
	job := &Job{
		ID:        uuid.New().String(),
		CompanyID: companyID,
		Status:    StatusPending,
		Progress:  Progress{},
		Results:   Results{},
		CreatedAt: time.Now(),
	}

	progressJSON, _ := json.Marshal(job.Progress)
	resultsJSON, _ := json.Marshal(job.Results)

	_, err = s.pool.Exec(ctx, `
		INSERT INTO prediction_jobs (id, company_id, status, progress, results, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, job.ID, job.CompanyID, job.Status, progressJSON, resultsJSON, job.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to create job: %w", err)
	}

	// Start background processing
	go s.processJob(job.ID, companyID)

	return job, nil
}

// processJob runs the prediction job in the background
func (s *Service) processJob(jobID, companyID string) {
	ctx := context.Background()
	log := s.log.With("job_id", jobID, "company_id", companyID)

	log.Info("========== PREDICTION JOB STARTED ==========")
	log.Info("Configuration status",
		"has_exa_client", s.exaClient != nil,
		"has_embedder", s.embedder != nil,
		"chat_model", s.chatModel,
	)

	startTime := time.Now()

	// Update status to processing
	now := time.Now()
	s.updateJobStatus(ctx, jobID, StatusProcessing, &now, nil)

	var results Results
	var progress Progress

	// Step 1: Generate keywords
	log.Info(">>> STEP 1/6: Generating keywords...")
	step1Start := time.Now()
	keywords, err := s.generateKeywords(ctx, companyID)
	if err != nil {
		log.Error("STEP 1 FAILED", "error", err, "duration_ms", time.Since(step1Start).Milliseconds())
		s.failJob(ctx, jobID, "Failed to generate keywords: "+err.Error())
		return
	}
	results.Keywords = keywords
	progress.Keywords = true
	s.updateJobProgress(ctx, jobID, progress, results)
	log.Info("<<< STEP 1 COMPLETE", "keywords_count", len(keywords), "keywords", keywords, "duration_ms", time.Since(step1Start).Milliseconds())

	// Step 2: Social media research
	log.Info(">>> STEP 2/6: Researching social media trends...")
	step2Start := time.Now()
	socialTrends, err := s.researchSocialMedia(ctx, companyID, keywords)
	if err != nil {
		log.Warn("STEP 2 FAILED (non-fatal)", "error", err, "duration_ms", time.Since(step2Start).Milliseconds())
		// Non-fatal, continue
	} else {
		results.SocialMediaTrends = socialTrends
		dataSource := "unknown"
		if ds, ok := socialTrends["data_source"].(string); ok {
			dataSource = ds
		}
		sourcesCount := 0
		if sources, ok := socialTrends["sources"].([]string); ok {
			sourcesCount = len(sources)
		}
		log.Info("<<< STEP 2 COMPLETE", "data_source", dataSource, "sources_count", sourcesCount, "duration_ms", time.Since(step2Start).Milliseconds())
	}
	progress.SocialMedia = true
	s.updateJobProgress(ctx, jobID, progress, results)

	// Step 3: Generate forecast
	log.Info(">>> STEP 3/6: Generating forecast...")
	step3Start := time.Now()
	forecastSummary, err := s.generateForecastSummary(ctx, companyID)
	if err != nil {
		log.Warn("STEP 3 FAILED (non-fatal)", "error", err, "duration_ms", time.Since(step3Start).Milliseconds())
		// Non-fatal, continue
	} else {
		results.ForecastSummary = forecastSummary
		log.Info("<<< STEP 3 COMPLETE", "forecast_length", len(forecastSummary), "duration_ms", time.Since(step3Start).Milliseconds())
	}
	progress.Forecast = true
	s.updateJobProgress(ctx, jobID, progress, results)

	// Step 4: Market prediction
	log.Info(">>> STEP 4/6: Generating market prediction...")
	step4Start := time.Now()
	marketPred, err := s.generateMarketPrediction(ctx, companyID, keywords)
	if err != nil {
		log.Warn("STEP 4 FAILED (non-fatal)", "error", err, "duration_ms", time.Since(step4Start).Milliseconds())
	} else {
		results.MarketPrediction = marketPred
		hasExaSources := false
		if len(marketPred) > 0 && (len(marketPred) > 100 && marketPred[len(marketPred)-50:] != "") {
			hasExaSources = true // Rough check for sources section
		}
		log.Info("<<< STEP 4 COMPLETE", "prediction_length", len(marketPred), "has_sources", hasExaSources, "duration_ms", time.Since(step4Start).Milliseconds())
	}
	progress.Market = true
	s.updateJobProgress(ctx, jobID, progress, results)

	// Step 5: Marketing recommendations
	log.Info(">>> STEP 5/6: Generating marketing recommendations...")
	step5Start := time.Now()
	marketingRecs, err := s.generateMarketingRecs(ctx, companyID, keywords)
	if err != nil {
		log.Warn("STEP 5 FAILED (non-fatal)", "error", err, "duration_ms", time.Since(step5Start).Milliseconds())
	} else {
		results.MarketingRecs = marketingRecs
		log.Info("<<< STEP 5 COMPLETE", "recs_length", len(marketingRecs), "duration_ms", time.Since(step5Start).Milliseconds())
	}
	progress.Marketing = true
	s.updateJobProgress(ctx, jobID, progress, results)

	// Step 6: Government regulations
	log.Info(">>> STEP 6/6: Fetching relevant regulations...")
	step6Start := time.Now()
	regulations, err := s.fetchRegulations(ctx, companyID, keywords)
	if err != nil {
		log.Warn("STEP 6 FAILED (non-fatal)", "error", err, "duration_ms", time.Since(step6Start).Milliseconds())
	} else {
		results.Regulations = regulations
		hasRAGSources := false
		if len(regulations) > 0 && (len(regulations) > 20) {
			hasRAGSources = true
		}
		log.Info("<<< STEP 6 COMPLETE", "regulations_length", len(regulations), "has_rag_sources", hasRAGSources, "duration_ms", time.Since(step6Start).Milliseconds())
	}
	progress.Regulations = true
	s.updateJobProgress(ctx, jobID, progress, results)

	// Complete job
	completedAt := time.Now()
	s.updateJobStatus(ctx, jobID, StatusCompleted, nil, &completedAt)

	// Send notification
	s.sendCompletionNotification(ctx, companyID, jobID)

	totalDuration := time.Since(startTime)
	log.Info("========== PREDICTION JOB COMPLETED ==========",
		"total_duration_ms", totalDuration.Milliseconds(),
		"total_duration_sec", totalDuration.Seconds(),
		"steps_completed", 6,
	)
}

// generateKeywords uses AI to generate relevant keywords for research
func (s *Service) generateKeywords(ctx context.Context, companyID string) ([]string, error) {
	s.log.Debug("[generateKeywords] Starting", "company_id", companyID)

	// Get company info
	var name, industry, city string
	err := s.pool.QueryRow(ctx, `
		SELECT name, COALESCE(industry, ''), COALESCE(city, '')
		FROM companies WHERE id = $1
	`, companyID).Scan(&name, &industry, &city)
	if err != nil {
		return nil, err
	}
	s.log.Debug("[generateKeywords] Company info", "name", name, "industry", industry, "city", city)

	// Get products
	rows, err := s.pool.Query(ctx, `
		SELECT name FROM products WHERE company_id = $1 AND is_active = true LIMIT 5
	`, companyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var products []string
	for rows.Next() {
		var productName string
		rows.Scan(&productName)
		products = append(products, productName)
	}
	s.log.Debug("[generateKeywords] Products found", "count", len(products), "products", products)

	prompt := fmt.Sprintf(`Berdasarkan informasi bisnis berikut, buatkan 5-10 keyword untuk riset pasar dan tren:
- Nama Perusahaan: %s
- Industri: %s
- Lokasi: %s
- Produk/Layanan: %s

Berikan output dalam format JSON array of strings, contoh: ["keyword1", "keyword2", "keyword3"]
Hanya berikan JSON array, tanpa penjelasan tambahan.`, name, industry, city, products)

	s.log.Debug("[generateKeywords] Calling AI", "model", s.chatModel)
	resp, err := s.chatProvider.CreateChatCompletion(ctx, chat.ChatCompletionRequest{
		Model: s.chatModel,
		Messages: []chat.ChatCompletionMessage{
			{Role: "user", Content: prompt},
		},
		MaxTokens:   500,
		Temperature: 0.7,
	})
	if err != nil {
		s.log.Error("[generateKeywords] AI call failed", "error", err)
		return nil, err
	}

	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("no response from AI")
	}

	s.log.Debug("[generateKeywords] AI response", "content_length", len(resp.Choices[0].Message.Content))

	var keywords []string
	if err := json.Unmarshal([]byte(resp.Choices[0].Message.Content), &keywords); err != nil {
		s.log.Warn("[generateKeywords] Failed to parse JSON, using fallback", "error", err, "raw_content", resp.Choices[0].Message.Content[:min(200, len(resp.Choices[0].Message.Content))])
		// Try to extract from response
		keywords = []string{industry, city, name}
	}

	return keywords, nil
}

// researchSocialMedia researches social media trends using Exa.ai + AI analysis
func (s *Service) researchSocialMedia(ctx context.Context, companyID string, keywords []string) (map[string]interface{}, error) {
	s.log.Debug("[researchSocialMedia] Starting", "company_id", companyID, "keywords", keywords)

	// Get company info and social media handles for context
	var name, industry, city, socialMediaStr string
	err := s.pool.QueryRow(ctx, `
		SELECT name, COALESCE(industry, ''), COALESCE(city, ''), COALESCE(social_media_handles::text, '{}')
		FROM companies WHERE id = $1
	`, companyID).Scan(&name, &industry, &city, &socialMediaStr)
	if err != nil {
		return nil, fmt.Errorf("failed to get company: %w", err)
	}
	s.log.Debug("[researchSocialMedia] Company info", "name", name, "industry", industry, "city", city)

	// Parse existing social media platforms
	var socialMedia map[string]interface{}
	json.Unmarshal([]byte(socialMediaStr), &socialMedia)
	platforms := make([]string, 0)
	for k, v := range socialMedia {
		if v != nil && v != "" {
			platforms = append(platforms, k)
		}
	}
	s.log.Debug("[researchSocialMedia] Platforms", "platforms", platforms)

	// Step 1: Use Exa.ai to search for real social media trends data
	var exaData string
	var exaSources []string
	if s.exaClient != nil && s.exaClient.IsConfigured() {
		s.log.Info("[researchSocialMedia] >>> EXA.AI SEARCH STARTING", "industry", industry, "city", city)

		exaResp, err := s.exaClient.SearchSocialMediaTrends(ctx, industry, city, keywords)
		if err != nil {
			s.log.Warn("[researchSocialMedia] EXA.AI SEARCH FAILED", "error", err)
		} else if len(exaResp.Results) > 0 {
			exaData = exa.FormatResultsForAI(exaResp.Results)
			for _, r := range exaResp.Results {
				exaSources = append(exaSources, r.URL)
			}
			s.log.Info("[researchSocialMedia] <<< EXA.AI SEARCH SUCCESS", "results_count", len(exaResp.Results), "sources", exaSources)
		} else {
			s.log.Warn("[researchSocialMedia] EXA.AI returned 0 results")
		}
	} else {
		s.log.Debug("[researchSocialMedia] Exa.ai not configured, using AI-only mode")
	}

	// Step 2: Build prompt with real data (if available) + AI analysis
	var prompt string
	if exaData != "" {
		prompt = fmt.Sprintf(`Kamu adalah ahli social media marketing untuk UMKM Indonesia.

PROFIL BISNIS:
- Nama: %s
- Industri: %s
- Lokasi: %s
- Platform aktif: %v
- Keyword bisnis: %v

DATA TREN TERKINI (dari riset web):
%s

Berdasarkan profil bisnis dan DATA TREN TERKINI di atas, berikan analisis dan rekomendasi social media:

1. **Platform Prioritas**: Berdasarkan data tren, platform mana yang paling efektif untuk industri %s di %s? Jelaskan dengan mengacu pada data.

2. **Tren Konten Saat Ini**: Jenis konten apa yang sedang trending berdasarkan data? (Reels, carousel, story, live, dll)

3. **Hashtag Relevan**: Berikan 10-15 hashtag yang relevan untuk bisnis ini.

4. **Jadwal Posting Optimal**: Kapan waktu terbaik posting untuk target audience di %s?

5. **Ide Konten**: Berikan 5 ide konten spesifik yang terinspirasi dari tren terkini.

Format jawaban yang actionable dan praktis untuk UMKM. Sebutkan sumber data jika relevan.`,
			name, industry, city, platforms, keywords, exaData, industry, city, city)
	} else {
		// Fallback to AI-only if no Exa data
		prompt = fmt.Sprintf(`Kamu adalah ahli social media marketing untuk UMKM Indonesia.

PROFIL BISNIS:
- Nama: %s
- Industri: %s
- Lokasi: %s
- Platform aktif: %v
- Keyword bisnis: %v

Berdasarkan profil di atas, berikan analisis dan rekomendasi social media:

1. **Platform Prioritas**: Platform mana yang paling efektif untuk industri %s di %s? Jelaskan alasannya.

2. **Tren Konten Saat Ini**: Jenis konten apa yang sedang trending untuk industri ini? (Reels, carousel, story, live, dll)

3. **Hashtag Relevan**: Berikan 10-15 hashtag yang relevan untuk bisnis ini (campuran populer dan niche).

4. **Jadwal Posting Optimal**: Kapan waktu terbaik posting untuk target audience di %s?

5. **Ide Konten**: Berikan 5 ide konten spesifik yang bisa langsung dieksekusi.

Format jawaban yang actionable dan praktis untuk UMKM.`, name, industry, city, platforms, keywords, industry, city, city)
	}

	resp, err := s.chatProvider.CreateChatCompletion(ctx, chat.ChatCompletionRequest{
		Model: s.chatModel,
		Messages: []chat.ChatCompletionMessage{
			{Role: "user", Content: prompt},
		},
		MaxTokens:   1500,
		Temperature: 0.7,
	})
	if err != nil {
		return nil, err
	}

	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("no response from AI")
	}

	result := map[string]interface{}{
		"analysis":  resp.Choices[0].Message.Content,
		"keywords":  keywords,
		"platforms": platforms,
		"industry":  industry,
		"location":  city,
	}

	// Add sources if we used Exa data
	if len(exaSources) > 0 {
		result["sources"] = exaSources
		result["data_source"] = "exa.ai + AI analysis"
	} else {
		result["data_source"] = "AI analysis only"
	}

	return result, nil
}

// generateForecastSummary generates forecast for company products
func (s *Service) generateForecastSummary(ctx context.Context, companyID string) (string, error) {
	// Get products with enough sales data
	rows, err := s.pool.Query(ctx, `
		SELECT p.id, p.name
		FROM products p
		JOIN sales_history s ON s.product_id = p.id
		WHERE p.company_id = $1 AND p.is_active = true
		GROUP BY p.id, p.name
		HAVING COUNT(s.id) >= 7
		LIMIT 3
	`, companyID)
	if err != nil {
		return "", err
	}
	defer rows.Close()

	var summaries []string
	for rows.Next() {
		var productID, productName string
		rows.Scan(&productID, &productName)

		forecast, err := s.forecastService.GenerateMonthlyForecast(ctx, productID)
		if err != nil {
			s.log.Warn("Failed to generate forecast for product", "product_id", productID, "error", err)
			continue
		}

		// Create summary
		if len(forecast.Forecasts) > 0 {
			totalPredicted := 0
			for _, f := range forecast.Forecasts {
				totalPredicted += f.PredictedQuantity
			}
			summaries = append(summaries, fmt.Sprintf("%s: prediksi %d unit dalam 12 bulan", productName, totalPredicted))
		}
	}

	if len(summaries) == 0 {
		return "Tidak ada data penjualan yang cukup untuk membuat forecast. Tambahkan minimal 7 data penjualan untuk setiap produk.", nil
	}

	result := "Ringkasan Forecast:\n"
	for _, s := range summaries {
		result += "• " + s + "\n"
	}
	return result, nil
}

// generateMarketPrediction generates market prediction using Exa.ai + AI analysis
func (s *Service) generateMarketPrediction(ctx context.Context, companyID string, keywords []string) (string, error) {
	s.log.Debug("[generateMarketPrediction] Starting", "company_id", companyID, "keywords", keywords)

	// Get company info
	var industry, city string
	err := s.pool.QueryRow(ctx, `
		SELECT COALESCE(industry, ''), COALESCE(city, '')
		FROM companies WHERE id = $1
	`, companyID).Scan(&industry, &city)
	if err != nil {
		return "", err
	}
	s.log.Debug("[generateMarketPrediction] Company info", "industry", industry, "city", city)

	// Step 1: Use Exa.ai to search for real market trends data
	var exaData string
	var exaSources []string
	if s.exaClient != nil && s.exaClient.IsConfigured() {
		s.log.Info("[generateMarketPrediction] >>> EXA.AI MARKET SEARCH STARTING", "industry", industry, "city", city)

		exaResp, err := s.exaClient.SearchMarketTrends(ctx, industry, city, keywords)
		if err != nil {
			s.log.Warn("[generateMarketPrediction] EXA.AI MARKET SEARCH FAILED", "error", err)
		} else if len(exaResp.Results) > 0 {
			exaData = exa.FormatResultsForAI(exaResp.Results)
			for _, r := range exaResp.Results {
				exaSources = append(exaSources, r.URL)
			}
			s.log.Info("[generateMarketPrediction] <<< EXA.AI MARKET SEARCH SUCCESS", "results_count", len(exaResp.Results), "sources", exaSources)
		} else {
			s.log.Warn("[generateMarketPrediction] EXA.AI returned 0 market results")
		}
	} else {
		s.log.Debug("[generateMarketPrediction] Exa.ai not configured, using AI-only mode")
	}

	// Step 2: Build prompt with real data (if available) + AI analysis
	var prompt string
	if exaData != "" {
		prompt = fmt.Sprintf(`Kamu adalah analis pasar untuk UMKM Indonesia.

KONTEKS BISNIS:
- Industri: %s
- Lokasi: %s
- Keyword: %v

DATA PASAR TERKINI (dari riset web):
%s

Berdasarkan DATA PASAR TERKINI di atas, berikan prediksi dan analisis:

1. **Tren Pasar 6 Bulan Ke Depan**: Berdasarkan data, bagaimana proyeksi pasar untuk industri %s?

2. **Peluang Pertumbuhan**: Peluang spesifik apa yang terlihat dari data tren?

3. **Tantangan & Risiko**: Apa tantangan yang mungkin dihadapi berdasarkan kondisi pasar saat ini?

4. **Rekomendasi Strategi**: Langkah konkret apa yang harus diambil UMKM untuk memanfaatkan tren ini?

5. **Kompetitor & Posisi Pasar**: Insight tentang lanskap kompetitif berdasarkan data.

Format: analisis mendalam dalam Bahasa Indonesia. Sebutkan sumber data jika relevan.`,
			industry, city, keywords, exaData, industry)
	} else {
		// Fallback to AI-only
		prompt = fmt.Sprintf(`Berikan prediksi pasar untuk bisnis di industri "%s" di wilayah "%s".

Keyword terkait: %v

Berikan analisis:
1. Tren pasar 6 bulan ke depan
2. Peluang pertumbuhan
3. Tantangan yang mungkin dihadapi
4. Rekomendasi strategi

Format: ringkasan dalam Bahasa Indonesia, maksimal 300 kata.`, industry, city, keywords)
	}

	resp, err := s.chatProvider.CreateChatCompletion(ctx, chat.ChatCompletionRequest{
		Model: s.chatModel,
		Messages: []chat.ChatCompletionMessage{
			{Role: "user", Content: prompt},
		},
		MaxTokens:   1200,
		Temperature: 0.7,
	})
	if err != nil {
		return "", err
	}

	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("no response from AI")
	}

	result := resp.Choices[0].Message.Content

	// Add sources if we used Exa data
	if len(exaSources) > 0 {
		result += "\n\n---\n📊 Sumber Data:\n"
		for _, src := range exaSources {
			result += "• " + src + "\n"
		}
	}

	return result, nil
}

// generateMarketingRecs generates marketing recommendations
func (s *Service) generateMarketingRecs(ctx context.Context, companyID string, keywords []string) (string, error) {
	// Get company social media
	var socialMediaStr string
	err := s.pool.QueryRow(ctx, `
		SELECT COALESCE(social_media_handles::text, '{}')
		FROM companies WHERE id = $1
	`, companyID).Scan(&socialMediaStr)
	if err != nil {
		return "", err
	}

	var socialMedia map[string]interface{}
	json.Unmarshal([]byte(socialMediaStr), &socialMedia)

	platforms := make([]string, 0)
	for k := range socialMedia {
		platforms = append(platforms, k)
	}

	prompt := fmt.Sprintf(`Berikan rekomendasi marketing untuk bisnis dengan:
- Platform aktif: %v
- Keyword bisnis: %v

Berikan rekomendasi:
1. Strategi konten untuk setiap platform
2. Ide campaign yang bisa dijalankan
3. Target audience yang disarankan
4. Budget marketing yang realistis untuk UMKM

Format: ringkasan dalam Bahasa Indonesia, actionable dan praktis.`, platforms, keywords)

	resp, err := s.chatProvider.CreateChatCompletion(ctx, chat.ChatCompletionRequest{
		Model: s.chatModel,
		Messages: []chat.ChatCompletionMessage{
			{Role: "user", Content: prompt},
		},
		MaxTokens:   800,
		Temperature: 0.7,
	})
	if err != nil {
		return "", err
	}

	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("no response from AI")
	}

	return resp.Choices[0].Message.Content, nil
}

// fetchRegulations fetches relevant government regulations
func (s *Service) fetchRegulations(ctx context.Context, companyID string, keywords []string) (string, error) {
	s.log.Debug("[fetchRegulations] Starting", "company_id", companyID, "keywords", keywords)

	// Get company industry
	var industry string
	err := s.pool.QueryRow(ctx, `
		SELECT COALESCE(industry, '') FROM companies WHERE id = $1
	`, companyID).Scan(&industry)
	if err != nil {
		return "", err
	}
	s.log.Debug("[fetchRegulations] Industry", "industry", industry)

	// Try RAG with vector search first (if embedder is configured)
	var ragContext string
	var ragSources []string
	if s.embedder != nil {
		s.log.Info("[fetchRegulations] >>> RAG VECTOR SEARCH STARTING", "industry", industry, "has_embedder", true)
		ragContext, ragSources = s.searchRegulationsRAG(ctx, industry, keywords)
	} else {
		s.log.Debug("[fetchRegulations] Embedder not configured, skipping RAG")
	}

	// If RAG found results, use AI to analyze them
	if ragContext != "" {
		s.log.Info("[fetchRegulations] <<< RAG FOUND RESULTS", "sources_count", len(ragSources), "context_length", len(ragContext))
		return s.generateRegulationsWithRAG(ctx, industry, keywords, ragContext, ragSources)
	}
	s.log.Debug("[fetchRegulations] RAG returned no results, falling back to text search")

	// Fallback to text search
	var regulations []string
	rows, err := s.pool.Query(ctx, `
		SELECT title, summary
		FROM regulations
		WHERE to_tsvector('indonesian', title || ' ' || COALESCE(summary, '')) @@ plainto_tsquery('indonesian', $1)
		LIMIT 5
	`, industry)
	if err != nil {
		s.log.Debug("[fetchRegulations] Text search failed, using AI-only", "error", err)
		// If regulations table doesn't exist or query fails, use AI
		return s.generateRegulationsAdvice(ctx, industry, keywords)
	}
	defer rows.Close()

	for rows.Next() {
		var title, summary string
		rows.Scan(&title, &summary)
		regulations = append(regulations, fmt.Sprintf("- %s: %s", title, summary))
	}
	s.log.Debug("[fetchRegulations] Text search results", "count", len(regulations))

	if len(regulations) == 0 {
		s.log.Debug("[fetchRegulations] No text search results, using AI-only")
		return s.generateRegulationsAdvice(ctx, industry, keywords)
	}

	result := "Regulasi terkait bisnis Anda:\n"
	for _, r := range regulations {
		result += r + "\n"
	}
	return result, nil
}

// searchRegulationsRAG uses vector search to find relevant regulations
func (s *Service) searchRegulationsRAG(ctx context.Context, industry string, keywords []string) (string, []string) {
	s.log.Debug("[searchRegulationsRAG] Starting RAG search", "industry", industry, "keywords", keywords)

	// Build query for vector search
	query := fmt.Sprintf("regulasi perizinan UMKM %s Indonesia", industry)
	if len(keywords) > 0 {
		query += " " + keywords[0]
	}
	s.log.Debug("[searchRegulationsRAG] Query", "query", query)

	// Generate query embedding
	s.log.Debug("[searchRegulationsRAG] Generating embedding...")
	queryEmbed, err := s.embedder.Embed(ctx, query)
	if err != nil {
		s.log.Warn("[searchRegulationsRAG] EMBEDDING GENERATION FAILED", "error", err)
		return "", nil
	}
	s.log.Debug("[searchRegulationsRAG] Embedding generated", "dimension", len(queryEmbed))

	// Convert to pgvector format string for query
	vectorStr := "["
	for i, v := range queryEmbed {
		if i > 0 {
			vectorStr += ","
		}
		vectorStr += fmt.Sprintf("%f", v)
	}
	vectorStr += "]"

	// Perform vector search on regulation_chunks + embeddings
	s.log.Debug("[searchRegulationsRAG] Executing vector search query...")
	rows, err := s.pool.Query(ctx, `
		SELECT rc.content, r.title, r.source_url,
		       1 - (e.embedding <=> $1::vector) AS similarity
		FROM regulation_chunks rc
		JOIN regulation_embeddings re ON re.chunk_id = rc.id
		JOIN embeddings e ON e.id = re.embedding_id
		JOIN regulations r ON r.id = rc.regulation_id
		WHERE 1 - (e.embedding <=> $1::vector) > 0.5
		ORDER BY e.embedding <=> $1::vector
		LIMIT 5
	`, vectorStr)
	if err != nil {
		s.log.Warn("[searchRegulationsRAG] VECTOR SEARCH QUERY FAILED", "error", err)
		return "", nil
	}
	defer rows.Close()

	var context string
	var sources []string
	seenSources := make(map[string]bool)
	resultCount := 0

	for rows.Next() {
		var content, title, sourceURL string
		var similarity float64
		if err := rows.Scan(&content, &title, &sourceURL, &similarity); err != nil {
			s.log.Warn("[searchRegulationsRAG] Failed to scan row", "error", err)
			continue
		}
		resultCount++
		s.log.Debug("[searchRegulationsRAG] Found result", "title", title, "similarity", similarity, "content_length", len(content))

		context += fmt.Sprintf("\n--- %s (similarity: %.2f) ---\n%s\n", title, similarity, content)

		if sourceURL != "" && !seenSources[sourceURL] {
			sources = append(sources, sourceURL)
			seenSources[sourceURL] = true
		}
	}

	s.log.Debug("[searchRegulationsRAG] Vector search complete", "results_found", resultCount, "sources_count", len(sources))
	return context, sources
}

// generateRegulationsWithRAG generates regulations advice using RAG context
func (s *Service) generateRegulationsWithRAG(ctx context.Context, industry string, keywords []string, ragContext string, sources []string) (string, error) {
	prompt := fmt.Sprintf(`Kamu adalah konsultan regulasi untuk UMKM Indonesia.

KONTEKS BISNIS:
- Industri: %s
- Keyword: %v

DATA REGULASI DARI DATABASE (gunakan ini sebagai sumber utama):
%s

Berdasarkan DATA REGULASI di atas, berikan panduan yang spesifik dan actionable:

1. **Perizinan Wajib**: Izin apa saja yang wajib dimiliki berdasarkan regulasi di atas?

2. **Regulasi yang Harus Dipatuhi**: Sebutkan regulasi spesifik dan poin pentingnya.

3. **Langkah-langkah Kepatuhan**: Berikan checklist praktis untuk UMKM.

4. **Peringatan**: Hal-hal yang perlu dihindari atau diperhatikan.

Format: praktis dan mudah dipahami untuk pemilik UMKM. Sebutkan nomor/nama regulasi jika ada.`, industry, keywords, ragContext)

	resp, err := s.chatProvider.CreateChatCompletion(ctx, chat.ChatCompletionRequest{
		Model: s.chatModel,
		Messages: []chat.ChatCompletionMessage{
			{Role: "user", Content: prompt},
		},
		MaxTokens:   1200,
		Temperature: 0.7,
	})
	if err != nil {
		return "", err
	}

	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("no response from AI")
	}

	result := resp.Choices[0].Message.Content

	// Add sources
	if len(sources) > 0 {
		result += "\n\n---\n📋 Sumber Regulasi:\n"
		for _, src := range sources {
			result += "• " + src + "\n"
		}
	}

	return result, nil
}

// generateRegulationsAdvice uses AI to give regulations advice
func (s *Service) generateRegulationsAdvice(ctx context.Context, industry string, keywords []string) (string, error) {
	prompt := fmt.Sprintf(`Berikan informasi regulasi dan perizinan yang perlu diperhatikan untuk bisnis di industri "%s" di Indonesia.

Keyword: %v

Berikan:
1. Perizinan wajib yang diperlukan
2. Regulasi terkait yang perlu dipatuhi
3. Tips kepatuhan untuk UMKM

Format: ringkasan dalam Bahasa Indonesia, praktis dan mudah dipahami.`, industry, keywords)

	resp, err := s.chatProvider.CreateChatCompletion(ctx, chat.ChatCompletionRequest{
		Model: s.chatModel,
		Messages: []chat.ChatCompletionMessage{
			{Role: "user", Content: prompt},
		},
		MaxTokens:   800,
		Temperature: 0.7,
	})
	if err != nil {
		return "", err
	}

	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("no response from AI")
	}

	return resp.Choices[0].Message.Content, nil
}

// updateJobStatus updates the job status
func (s *Service) updateJobStatus(ctx context.Context, jobID string, status JobStatus, startedAt, completedAt *time.Time) {
	_, err := s.pool.Exec(ctx, `
		UPDATE prediction_jobs
		SET status = $2, started_at = COALESCE($3, started_at), completed_at = COALESCE($4, completed_at)
		WHERE id = $1
	`, jobID, status, startedAt, completedAt)
	if err != nil {
		s.log.Error("Failed to update job status", "job_id", jobID, "error", err)
	}
}

// updateJobProgress updates job progress and results
func (s *Service) updateJobProgress(ctx context.Context, jobID string, progress Progress, results Results) {
	progressJSON, _ := json.Marshal(progress)
	resultsJSON, _ := json.Marshal(results)

	_, err := s.pool.Exec(ctx, `
		UPDATE prediction_jobs
		SET progress = $2, results = $3
		WHERE id = $1
	`, jobID, progressJSON, resultsJSON)
	if err != nil {
		s.log.Error("Failed to update job progress", "job_id", jobID, "error", err)
	}
}

// failJob marks a job as failed
func (s *Service) failJob(ctx context.Context, jobID, errorMessage string) {
	completedAt := time.Now()
	_, err := s.pool.Exec(ctx, `
		UPDATE prediction_jobs
		SET status = $2, error_message = $3, completed_at = $4
		WHERE id = $1
	`, jobID, StatusFailed, errorMessage, completedAt)
	if err != nil {
		s.log.Error("Failed to fail job", "job_id", jobID, "error", err)
	}
}

// Pool returns the database pool for direct queries
func (s *Service) Pool() *pgxpool.Pool {
	return s.pool
}

// sendCompletionNotification sends a notification when job completes
func (s *Service) sendCompletionNotification(ctx context.Context, companyID, jobID string) {
	notificationID := uuid.New().String()
	_, err := s.pool.Exec(ctx, `
		INSERT INTO notifications (id, company_id, title, body, type, status, created_at)
		VALUES ($1, $2, $3, $4, $5, 'unread', NOW())
	`, notificationID, companyID,
		"Analisis Bisnis Selesai!",
		"Prediksi dan rekomendasi bisnis Anda sudah siap. Lihat hasilnya di Dashboard.",
		"prediction_complete",
	)
	if err != nil {
		s.log.Error("Failed to send notification", "company_id", companyID, "error", err)
	}
}
