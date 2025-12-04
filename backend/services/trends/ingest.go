package trends

import (
	"context"
	"fmt"
	"time"

	"github.com/bantuaku/backend/logger"
	"github.com/jackc/pgx/v5/pgxpool"
)

// IngestService handles ingestion of trends data
type IngestService struct {
	store *Store
	log   logger.Logger
}

// NewIngestService creates a new ingestion service
func NewIngestService(pool *pgxpool.Pool) *IngestService {
	return &IngestService{
		store: NewStore(pool),
		log:   logger.Default(),
	}
}

// Scraper interface for trends scraping
type Scraper interface {
	ScrapeTrends(ctx context.Context, keyword string, geo string) (*TrendData, error)
	ScrapeWithRetry(ctx context.Context, keyword string, geo string, maxRetries int) (*TrendData, error)
}

// TrendData represents scraped trend data
type TrendData struct {
	Keyword        string
	Geo            string
	TimeSeries     []TimeSeriesPoint
	RelatedQueries []RelatedQuery
	CapturedAt     time.Time
}

// IngestKeyword ingests trends data for a keyword
func (i *IngestService) IngestKeyword(ctx context.Context, keywordID string, keyword string, geo string, scraper Scraper) error {
	i.log.Info("Ingesting trends data", "keyword", keyword, "geo", geo)

	// Scrape trends data
	trendData, err := scraper.ScrapeWithRetry(ctx, keyword, geo, 3)
	if err != nil {
		return fmt.Errorf("failed to scrape trends: %w", err)
	}

	// Store time series
	for _, point := range trendData.TimeSeries {
		if err := i.store.StoreTimeSeries(ctx, keywordID, point.Time, point.Value, geo); err != nil {
			i.log.Warn("Failed to store time series point", "error", err, "time", point.Time)
			continue
		}
	}

	// Store related queries
	if len(trendData.RelatedQueries) > 0 {
		if err := i.store.StoreRelatedQueries(ctx, keywordID, trendData.RelatedQueries, geo); err != nil {
			i.log.Warn("Failed to store related queries", "error", err)
		}
	}

	i.log.Info("Trends ingestion completed",
		"keyword", keyword,
		"time_series_points", len(trendData.TimeSeries),
		"related_queries", len(trendData.RelatedQueries),
	)

	return nil
}

// IngestCompanyKeywords ingests trends for all active keywords of a company
func (i *IngestService) IngestCompanyKeywords(ctx context.Context, companyID string, scraper Scraper) error {
	// Get active keywords
	keywords, err := i.store.GetKeywords(ctx, companyID)
	if err != nil {
		return fmt.Errorf("failed to get keywords: %w", err)
	}

	if len(keywords) == 0 {
		i.log.Info("No active keywords found", "company_id", companyID)
		return nil
	}

	i.log.Info("Ingesting trends for company", "company_id", companyID, "keywords", len(keywords))

	// Ingest each keyword
	for _, kw := range keywords {
		if err := i.IngestKeyword(ctx, kw.ID, kw.Keyword, kw.Geo, scraper); err != nil {
			i.log.Warn("Failed to ingest keyword", "keyword", kw.Keyword, "error", err)
			continue
		}

		// Rate limiting - delay between keywords
		time.Sleep(2 * time.Second)
	}

	return nil
}

// IngestAllActiveKeywords ingests trends for all active keywords across all companies
func (i *IngestService) IngestAllActiveKeywords(ctx context.Context, scraper Scraper) error {
	// Get all active keywords
	pool := i.store.pool
	rows, err := pool.Query(ctx,
		`SELECT DISTINCT company_id FROM trends_keywords WHERE is_active = true`,
	)
	if err != nil {
		return fmt.Errorf("failed to get companies: %w", err)
	}
	defer rows.Close()

	var companyIDs []string
	for rows.Next() {
		var companyID string
		if err := rows.Scan(&companyID); err != nil {
			continue
		}
		companyIDs = append(companyIDs, companyID)
	}

	i.log.Info("Ingesting trends for all companies", "companies", len(companyIDs))

	for _, companyID := range companyIDs {
		if err := i.IngestCompanyKeywords(ctx, companyID, scraper); err != nil {
			i.log.Warn("Failed to ingest company keywords", "company_id", companyID, "error", err)
			continue
		}
	}

	return nil
}

