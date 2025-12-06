package regulations

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/bantuaku/backend/logger"
	"github.com/bantuaku/backend/services/chat"
	"github.com/bantuaku/backend/services/embedding"
	"github.com/bantuaku/backend/services/exa"
	"github.com/bantuaku/backend/services/kolosal"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Scheduler handles scheduled scraping jobs with AI-powered discovery
type Scheduler struct {
	// Legacy components
	crawler   *Crawler
	extractor *Extractor
	chunker   *Chunker

	// New AI-powered components
	keywordGen *KeywordGenerator
	discovery  *DiscoveryService
	processor  *ContentProcessor
	store      *Store

	log     logger.Logger
	running bool
	mu      sync.Mutex
}

// NewScheduler creates a new scheduler (legacy mode without AI)
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

// NewSchedulerV2 creates a new AI-powered scheduler
func NewSchedulerV2(
	pool *pgxpool.Pool,
	baseURL string,
	kolosalClient *kolosal.Client,
	exaClient *exa.Client,
	chatProvider chat.ChatProvider,
	embedder embedding.Embedder,
	chatModel string,
) *Scheduler {
	crawler := NewCrawler(baseURL)
	extractor := NewExtractor(kolosalClient)
	chunker := NewChunker()

	// Use store with embedder
	var store *Store
	if embedder != nil {
		store = NewStoreWithEmbedder(pool, embedder)
	} else {
		store = NewStore(pool)
	}

	// Create new AI-powered components
	var keywordGen *KeywordGenerator
	var discovery *DiscoveryService
	var processor *ContentProcessor

	if chatProvider != nil {
		keywordGen = NewKeywordGenerator(chatProvider, chatModel)
		processor = NewContentProcessor(extractor, chatProvider, chatModel)
	}

	if exaClient != nil {
		discovery = NewDiscoveryService(exaClient)
	}

	return &Scheduler{
		crawler:    crawler,
		extractor:  extractor,
		chunker:    chunker,
		keywordGen: keywordGen,
		discovery:  discovery,
		processor:  processor,
		store:      store,
		log:        *logger.Default(),
		running:    false,
	}
}

// RunJob runs a single scraping job (uses AI-powered pipeline if available)
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

	// Use AI-powered pipeline if components are available
	if s.keywordGen != nil && s.discovery != nil && s.processor != nil {
		return s.runJobV2(ctx, maxPages)
	}

	// Fall back to legacy pipeline
	return s.runLegacyJob(ctx, maxPages)
}

// runJobV2 runs the AI-powered scraping pipeline
func (s *Scheduler) runJobV2(ctx context.Context, maxResultsPerKeyword int) error {
	s.log.Info("========== REGULATION SCRAPER V2 STARTED ==========")

	startTime := time.Now()
	var totalStored, totalErrors int

	// Step 1: Generate UMKM keywords using AI
	s.log.Info(">>> STEP 1/4: Generating UMKM keywords...")
	keywords, err := s.keywordGen.GenerateKeywords(ctx)
	if err != nil {
		s.log.Warn("AI keyword generation failed, using fallback", "error", err)
		keywords = s.keywordGen.getFallbackKeywords()
	}
	s.log.Info("<<< STEP 1 COMPLETE", "keywords_count", len(keywords))

	// Step 2: Discover regulations via Exa.ai
	s.log.Info(">>> STEP 2/4: Discovering regulations via Exa.ai...")
	discovered, err := s.discovery.DiscoverRegulations(ctx, keywords, maxResultsPerKeyword)
	if err != nil {
		return fmt.Errorf("discovery failed: %w", err)
	}
	s.log.Info("<<< STEP 2 COMPLETE", "discovered_count", len(discovered))

	if len(discovered) == 0 {
		s.log.Warn("No regulations discovered, job ending early")
		return nil
	}

	// Step 3: Process content (extract, summarize)
	s.log.Info(">>> STEP 3/4: Processing regulation content...")
	processed := s.processor.ProcessBatch(ctx, discovered)
	s.log.Info("<<< STEP 3 COMPLETE", "processed_count", len(processed))

	// Step 4: Store with embeddings
	s.log.Info(">>> STEP 4/4: Storing regulations with embeddings...")
	for i, reg := range processed {
		s.log.Debug("Storing regulation", "index", i+1, "title", reg.Title)

		_, err := s.store.StoreProcessedRegulation(ctx, reg)
		if err != nil {
			s.log.Warn("Failed to store regulation", "title", reg.Title, "error", err)
			totalErrors++
			continue
		}
		totalStored++
	}
	s.log.Info("<<< STEP 4 COMPLETE", "stored_count", totalStored, "errors", totalErrors)

	totalDuration := time.Since(startTime)
	s.log.Info("========== REGULATION SCRAPER V2 COMPLETED ==========",
		"total_duration_sec", totalDuration.Seconds(),
		"keywords_used", len(keywords),
		"discovered", len(discovered),
		"processed", len(processed),
		"stored", totalStored,
		"errors", totalErrors,
	)

	return nil
}

// runLegacyJob runs the original scraping job (for backwards compatibility)
func (s *Scheduler) runLegacyJob(ctx context.Context, maxPages int) error {
	s.log.Info("Starting LEGACY regulation scraping job", "max_pages", maxPages)

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

	s.log.Info("Legacy scraping job completed",
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
