package trends

import (
	"context"
	"fmt"
	"time"

	"github.com/bantuaku/backend/logger"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Store handles database persistence for trends data
type Store struct {
	pool *pgxpool.Pool
	log  logger.Logger
}

// NewStore creates a new trends store
func NewStore(pool *pgxpool.Pool) *Store {
	return &Store{
		pool: pool,
		log:  logger.Default(),
	}
}

// TrendKeyword represents a tracked keyword
type TrendKeyword struct {
	ID        string
	CompanyID string
	Keyword   string
	Geo       string
	Category  *string
	IsActive  bool
}

// UpsertKeyword upserts a keyword for a company
func (s *Store) UpsertKeyword(ctx context.Context, companyID string, keyword string, geo string, category *string) (string, error) {
	// Check if keyword exists
	var existingID string
	err := s.pool.QueryRow(ctx,
		"SELECT id FROM trends_keywords WHERE company_id = $1 AND keyword = $2 AND geo = $3",
		companyID, keyword, geo,
	).Scan(&existingID)

	if err == nil {
		// Update existing keyword
		_, err = s.pool.Exec(ctx,
			`UPDATE trends_keywords 
			 SET category = $1, is_active = true, updated_at = NOW()
			 WHERE id = $2`,
			category, existingID,
		)
		if err != nil {
			return "", fmt.Errorf("failed to update keyword: %w", err)
		}
		return existingID, nil
	}

	// Check if error is "no rows"
	if err.Error() != "no rows in result set" {
		return "", fmt.Errorf("failed to check existing keyword: %w", err)
	}

	// Create new keyword
	id := uuid.New().String()
	_, err = s.pool.Exec(ctx,
		`INSERT INTO trends_keywords (id, company_id, keyword, geo, category, is_active, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW())`,
		id, companyID, keyword, geo, category, true,
	)
	if err != nil {
		return "", fmt.Errorf("failed to insert keyword: %w", err)
	}

	return id, nil
}

// StoreTimeSeries stores time series data with deduplication
func (s *Store) StoreTimeSeries(ctx context.Context, keywordID string, time time.Time, value int, geo string) error {
	_, err := s.pool.Exec(ctx,
		`INSERT INTO trends_series (id, keyword_id, time, value, geo, created_at)
		 VALUES ($1, $2, $3, $4, $5, NOW())
		 ON CONFLICT (keyword_id, time, geo) DO UPDATE 
		 SET value = EXCLUDED.value, created_at = NOW()`,
		uuid.New().String(), keywordID, time, value, geo,
	)
	if err != nil {
		return fmt.Errorf("failed to store time series: %w", err)
	}
	return nil
}

// StoreRelatedQueries stores related queries
func (s *Store) StoreRelatedQueries(ctx context.Context, keywordID string, queries []RelatedQuery, geo string) error {
	capturedAt := time.Now()
	
	for _, query := range queries {
		_, err := s.pool.Exec(ctx,
			`INSERT INTO trends_related_queries 
			 (id, keyword_id, related_keyword, relationship_type, value, geo, captured_at, created_at)
			 VALUES ($1, $2, $3, $4, $5, $6, $7, NOW())
			 ON CONFLICT DO NOTHING`,
			uuid.New().String(), keywordID, query.Query, query.RelationshipType, query.Value, geo, capturedAt,
		)
		if err != nil {
			s.log.Warn("Failed to store related query", "error", err, "query", query.Query)
			continue
		}
	}
	
	return nil
}

// GetKeywords retrieves active keywords for a company
func (s *Store) GetKeywords(ctx context.Context, companyID string) ([]TrendKeyword, error) {
	rows, err := s.pool.Query(ctx,
		`SELECT id, company_id, keyword, geo, category, is_active
		 FROM trends_keywords
		 WHERE company_id = $1 AND is_active = true
		 ORDER BY keyword`,
		companyID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query keywords: %w", err)
	}
	defer rows.Close()

	var keywords []TrendKeyword
	for rows.Next() {
		var kw TrendKeyword
		err := rows.Scan(&kw.ID, &kw.CompanyID, &kw.Keyword, &kw.Geo, &kw.Category, &kw.IsActive)
		if err != nil {
			s.log.Warn("Failed to scan keyword", "error", err)
			continue
		}
		keywords = append(keywords, kw)
	}

	return keywords, nil
}

// GetTimeSeries retrieves time series for a keyword
func (s *Store) GetTimeSeries(ctx context.Context, keywordID string, startTime, endTime *time.Time) ([]TimeSeriesPoint, error) {
	query := `SELECT time, value FROM trends_series WHERE keyword_id = $1`
	args := []interface{}{keywordID}
	argIndex := 2

	if startTime != nil {
		query += fmt.Sprintf(" AND time >= $%d", argIndex)
		args = append(args, *startTime)
		argIndex++
	}

	if endTime != nil {
		query += fmt.Sprintf(" AND time <= $%d", argIndex)
		args = append(args, *endTime)
		argIndex++
	}

	query += " ORDER BY time ASC"

	rows, err := s.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query time series: %w", err)
	}
	defer rows.Close()

	var points []TimeSeriesPoint
	for rows.Next() {
		var point TimeSeriesPoint
		err := rows.Scan(&point.Time, &point.Value)
		if err != nil {
			s.log.Warn("Failed to scan time series point", "error", err)
			continue
		}
		points = append(points, point)
	}

	return points, nil
}

// RelatedQuery represents a related query
type RelatedQuery struct {
	Query          string
	RelationshipType string
	Value          int
}

// TimeSeriesPoint represents a time series data point
type TimeSeriesPoint struct {
	Time  time.Time
	Value int
}

