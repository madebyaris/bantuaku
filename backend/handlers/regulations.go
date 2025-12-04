package handlers

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/bantuaku/backend/logger"
	"github.com/bantuaku/backend/services/kolosal"
	"github.com/bantuaku/backend/services/scraper/regulations"
	"github.com/jackc/pgx/v5"
)

// TriggerScraping triggers a manual scraping job
func (h *Handler) TriggerScraping(w http.ResponseWriter, r *http.Request) {
	log := logger.With("request_id", r.Context().Value("request_id"))

	// Get max_pages query parameter (default: 5)
	maxPages := 5
	if maxPagesStr := r.URL.Query().Get("max_pages"); maxPagesStr != "" {
		if parsed, err := strconv.Atoi(maxPagesStr); err == nil && parsed > 0 {
			maxPages = parsed
		}
	}

	// Get database pool
	pool := h.db.Pool()

	// Create Kolosal client
	kolosalClient := kolosal.NewClient(h.config.KolosalAPIKey)

	// Create scheduler
	scheduler := regulations.NewScheduler(
		pool,
		h.config.RegulationsBaseURL,
		kolosalClient,
	)

	// Check if job is already running
	if scheduler.IsRunning() {
		h.respondJSON(w, http.StatusConflict, map[string]string{
			"error": "Scraping job is already running",
		})
		return
	}

	// Run job in goroutine
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
		defer cancel()

		if err := scheduler.RunJob(ctx, maxPages); err != nil {
			log.Error("Scraping job failed", "error", err)
		}
	}()

	log.Info("Scraping job triggered", "max_pages", maxPages)

	h.respondJSON(w, http.StatusAccepted, map[string]interface{}{
		"message":  "Scraping job started",
		"max_pages": maxPages,
		"status":   "running",
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

