package regulations

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/bantuaku/backend/logger"
	"github.com/bantuaku/backend/services/kolosal"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Scheduler handles scheduled scraping jobs
type Scheduler struct {
	crawler   *Crawler
	extractor *Extractor
	chunker   *Chunker
	store     *Store
	log       logger.Logger
	running   bool
	mu        sync.Mutex
}

// NewScheduler creates a new scheduler
func NewScheduler(pool *pgxpool.Pool, baseURL string, kolosalClient *kolosal.Client) *Scheduler {
	crawler := NewCrawler(baseURL)
	extractor := NewExtractor(kolosalClient)
	chunker := NewChunker()
	store := NewStore(pool)

	return &Scheduler{
		crawler:   crawler,
		extractor: extractor,
		chunker:   chunker,
		store:     store,
		log:       *logger.Default(),
		running:   false,
	}
}

// RunJob runs a single scraping job
func (s *Scheduler) RunJob(ctx context.Context, maxPages int) error {
	s.mu.Lock()
	if s.running {
		s.mu.Unlock()
		return fmt.Errorf("scraping job already running")
	}
	s.running = true
	s.mu.Unlock()

	defer func() {
		s.mu.Lock()
		s.running = false
		s.mu.Unlock()
	}()

	s.log.Info("Starting regulation scraping job", "max_pages", maxPages)

	// Step 1: Crawl regulations
	regulations, err := s.crawler.CrawlRegulations(ctx, maxPages)
	if err != nil {
		return fmt.Errorf("failed to crawl regulations: %w", err)
	}

	s.log.Info("Crawled regulations", "count", len(regulations))

	// Step 2: Process each regulation
	processed := 0
	skipped := 0
	errors := 0

	for _, reg := range regulations {
		// Check if already processed
		regulationID, err := s.store.UpsertRegulation(ctx, &reg)
		if err != nil {
			s.log.Warn("Failed to upsert regulation", "title", reg.Title, "error", err)
			errors++
			continue
		}

		// Check if already processed
		alreadyProcessed, err := s.store.IsRegulationProcessed(ctx, regulationID)
		if err == nil && alreadyProcessed {
			s.log.Info("Regulation already processed, skipping", "id", regulationID)
			skipped++
			continue
		}

		// Extract PDF text
		extracted, err := s.extractor.ExtractPDF(ctx, reg.PDFURL)
		if err != nil {
			s.log.Warn("Failed to extract PDF", "url", reg.PDFURL, "error", err)
			errors++
			continue
		}

		// Clean text
		cleanedText := s.extractor.CleanText(extracted.Text)

		// Store section (entire document as one section for now)
		sectionID, err := s.store.StoreSection(ctx, regulationID, "", reg.Title, cleanedText, 0, 0)
		if err != nil {
			s.log.Warn("Failed to store section", "error", err)
			errors++
			continue
		}

		// Chunk text
		chunks := s.chunker.ChunkText(cleanedText)

		// Store chunks
		for _, chunk := range chunks {
			_, err := s.store.StoreChunk(ctx, regulationID, &sectionID, chunk)
			if err != nil {
				s.log.Warn("Failed to store chunk", "error", err)
				errors++
				continue
			}
		}

		processed++
		s.log.Info("Processed regulation", "id", regulationID, "chunks", len(chunks))

		// Rate limiting
		time.Sleep(1 * time.Second)
	}

	s.log.Info("Scraping job completed",
		"processed", processed,
		"skipped", skipped,
		"errors", errors,
		"total", len(regulations),
	)

	return nil
}

// StartDailyJob starts a daily scheduled job
func (s *Scheduler) StartDailyJob(ctx context.Context, scheduleTime time.Time) {
	ticker := time.NewTicker(24 * time.Hour)
	defer ticker.Stop()

	// Calculate time until first run
	now := time.Now()
	firstRun := scheduleTime
	if scheduleTime.Before(now) {
		firstRun = scheduleTime.Add(24 * time.Hour)
	}
	duration := firstRun.Sub(now)

	s.log.Info("Scheduling daily scraping job", "first_run", firstRun, "duration", duration)

	// Wait for first run
	time.Sleep(duration)

	// Run immediately
	if err := s.RunJob(ctx, 10); err != nil {
		s.log.Error("Daily scraping job failed", "error", err)
	}

	// Then run on schedule
	for {
		select {
		case <-ctx.Done():
			s.log.Info("Stopping daily scraping job")
			return
		case <-ticker.C:
			if err := s.RunJob(ctx, 10); err != nil {
				s.log.Error("Daily scraping job failed", "error", err)
			}
		}
	}
}

// IsRunning returns whether a job is currently running
func (s *Scheduler) IsRunning() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.running
}

