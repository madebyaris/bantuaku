package handlers

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/bantuaku/backend/logger"
	"github.com/bantuaku/backend/services/audit"
	"github.com/bantuaku/backend/services/chat"
	"github.com/bantuaku/backend/services/embedding"
	"github.com/bantuaku/backend/services/exa"
	"github.com/bantuaku/backend/services/kolosal"
	"github.com/bantuaku/backend/services/scraper/regulations"
	"github.com/bantuaku/backend/services/settings"
)

// TriggerScraping triggers a manual scraping job (AI-powered v2)
func (h *Handler) TriggerScraping(w http.ResponseWriter, r *http.Request) {
	log := logger.With("request_id", r.Context().Value("request_id"))

	// Get max_results query parameter (default: 3 results per keyword)
	maxResults := 3
	if maxResultsStr := r.URL.Query().Get("max_results"); maxResultsStr != "" {
		if parsed, err := strconv.Atoi(maxResultsStr); err == nil && parsed > 0 {
			maxResults = parsed
		}
	}

	// Get database pool
	pool := h.db.Pool()

	// Create Kolosal client
	kolosalClient := kolosal.NewClient(h.config.KolosalAPIKey)

	// Create Exa client (for AI-powered discovery)
	var exaClient *exa.Client
	if h.config.ExaAPIKey != "" {
		exaClient = exa.NewClient(h.config.ExaAPIKey)
		log.Info("Exa.ai client available for regulation discovery")
	}

	// Create chat provider (for AI keyword generation and summarization)
	var chatProvider chat.ChatProvider
	var chatModel string
	settingsService := settings.NewService(h.db)
	chatProvider, err := chat.NewChatProvider(r.Context(), h.config, settingsService)
	if err != nil {
		log.Warn("Chat provider not available", "error", err)
	} else {
		chatModel = "x-ai/grok-4-fast"
		if modelSetting, err := settingsService.GetSetting(r.Context(), "ai_model"); err == nil && modelSetting != "" {
			chatModel = modelSetting
		}
	}

	// Create embedder (for RAG)
	var embedder embedding.Embedder
	embedder, err = embedding.NewEmbedder(h.config)
	if err != nil {
		log.Warn("Embedder not available", "error", err)
	}

	// Create AI-powered scheduler (v2)
	scheduler := regulations.NewSchedulerV2(
		pool,
		h.config.RegulationsBaseURL,
		kolosalClient,
		exaClient,
		chatProvider,
		embedder,
		chatModel,
	)

	// Check if job is already running
	if scheduler.IsRunning() {
		h.respondJSON(w, http.StatusConflict, map[string]string{
			"error": "Scraping job is already running",
		})
		return
	}

	// Determine mode
	mode := "v2_ai_powered"
	if exaClient == nil || chatProvider == nil {
		mode = "legacy"
	}

	// Run job in goroutine
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
		defer cancel()

		if err := scheduler.RunJob(ctx, maxResults); err != nil {
			log.Error("Scraping job failed", "error", err)
		}
	}()

	log.Info("Scraping job triggered", "max_results", maxResults, "mode", mode)

	// Log audit event for scraping action
	if h.auditLogger != nil {
		h.auditLogger.LogAction(r.Context(), r, audit.ActionRegulationScraped, map[string]interface{}{
			"max_results": maxResults,
			"mode":        mode,
		})
	}

	h.respondJSON(w, http.StatusAccepted, map[string]interface{}{
		"message":     "Scraping job started",
		"max_results": maxResults,
		"mode":        mode,
		"status":      "running",
	})
}

// GetScrapingStatus returns the status of scraping jobs
func (h *Handler) GetScrapingStatus(w http.ResponseWriter, r *http.Request) {
	log := logger.With("request_id", r.Context().Value("request_id"))

	pool := h.db.Pool()
	ctx := r.Context()

	// Get statistics
	var totalRegulations, totalChunks, lastScrapeTime int64
	var lastScrapeTimestamp *time.Time

	err := pool.QueryRow(ctx,
		"SELECT COUNT(*) FROM regulations",
	).Scan(&totalRegulations)
	if err != nil {
		log.Warn("Failed to get regulation count", "error", err)
	}

	err = pool.QueryRow(ctx,
		"SELECT COUNT(*) FROM regulation_chunks",
	).Scan(&totalChunks)
	if err != nil {
		log.Warn("Failed to get chunk count", "error", err)
	}

	err = pool.QueryRow(ctx,
		`SELECT EXTRACT(EPOCH FROM MAX(created_at))::bigint 
		 FROM regulation_sources`,
	).Scan(&lastScrapeTime)
	if err == nil && lastScrapeTime > 0 {
		t := time.Unix(lastScrapeTime, 0)
		lastScrapeTimestamp = &t
	}

	h.respondJSON(w, http.StatusOK, map[string]interface{}{
		"total_regulations": totalRegulations,
		"total_chunks":      totalChunks,
		"last_scrape":       lastScrapeTimestamp,
	})
}

// ListRegulations lists scraped regulations
func (h *Handler) ListRegulations(w http.ResponseWriter, r *http.Request) {
	log := logger.With("request_id", r.Context().Value("request_id"))

	pool := h.db.Pool()
	ctx := r.Context()

	// Get pagination parameters
	limit := 50
	offset := 0
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if parsed, err := strconv.Atoi(limitStr); err == nil && parsed > 0 && parsed <= 100 {
			limit = parsed
		}
	}
	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		if parsed, err := strconv.Atoi(offsetStr); err == nil && parsed >= 0 {
			offset = parsed
		}
	}

	// Get category filter
	category := r.URL.Query().Get("category")
	query := `
		SELECT id, title, regulation_number, year, category, status, source_url, pdf_url,
		       published_date, effective_date, created_at
		FROM regulations
		WHERE ($1 = '' OR category = $1)
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := pool.Query(ctx, query, category, limit, offset)
	if err != nil {
		log.Error("Failed to query regulations", "error", err)
		h.respondJSON(w, http.StatusInternalServerError, map[string]string{
			"error": "Failed to fetch regulations",
		})
		return
	}
	defer rows.Close()

	var regulations []map[string]interface{}
	for rows.Next() {
		var id, title, regulationNumber, category, status, sourceURL, pdfURL string
		var year *int
		var publishedDate, effectiveDate *time.Time
		var createdAt time.Time

		err := rows.Scan(&id, &title, &regulationNumber, &year, &category, &status,
			&sourceURL, &pdfURL, &publishedDate, &effectiveDate, &createdAt)
		if err != nil {
			log.Warn("Failed to scan regulation", "error", err)
			continue
		}

		reg := map[string]interface{}{
			"id":                id,
			"title":             title,
			"regulation_number": regulationNumber,
			"year":              year,
			"category":          category,
			"status":            status,
			"source_url":        sourceURL,
			"pdf_url":           pdfURL,
			"published_date":    publishedDate,
			"effective_date":    effectiveDate,
			"created_at":        createdAt,
		}
		regulations = append(regulations, reg)
	}

	h.respondJSON(w, http.StatusOK, map[string]interface{}{
		"regulations": regulations,
		"count":       len(regulations),
		"limit":       limit,
		"offset":      offset,
	})
}
