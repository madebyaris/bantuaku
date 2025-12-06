package aggregates

import (
	"context"
	"time"

	"github.com/bantuaku/backend/services/storage"
)

// TokenUsageAggregator aggregates token usage into token_usage_aggregates for fast dashboard queries.
type TokenUsageAggregator struct {
	db *storage.Postgres
}

// NewTokenUsageAggregator creates a new aggregator.
func NewTokenUsageAggregator(db *storage.Postgres) *TokenUsageAggregator {
	return &TokenUsageAggregator{db: db}
}

// AggregateDaily aggregates token_usage for a single day (UTC) into token_usage_aggregates.
func (a *TokenUsageAggregator) AggregateDaily(ctx context.Context, day time.Time) error {
	start := time.Date(day.Year(), day.Month(), day.Day(), 0, 0, 0, 0, time.UTC)
	end := start.Add(24 * time.Hour)

	_, err := a.db.Pool().Exec(ctx, `
		INSERT INTO token_usage_aggregates (
			date, user_id, company_id, model, provider,
			prompt_tokens, completion_tokens, total_tokens,
			created_at, updated_at
		)
		SELECT
			DATE(tu.created_at) AS date,
			tu.user_id,
			tu.company_id,
			tu.model,
			tu.provider,
			COALESCE(SUM(tu.prompt_tokens), 0) AS prompt_tokens,
			COALESCE(SUM(tu.completion_tokens), 0) AS completion_tokens,
			COALESCE(SUM(tu.total_tokens), 0) AS total_tokens,
			NOW(),
			NOW()
		FROM token_usage tu
		WHERE tu.created_at >= $1 AND tu.created_at < $2
		GROUP BY DATE(tu.created_at), tu.user_id, tu.company_id, tu.model, tu.provider
		ON CONFLICT (date, user_id, company_id, model, provider) DO UPDATE
		SET 
			prompt_tokens = EXCLUDED.prompt_tokens,
			completion_tokens = EXCLUDED.completion_tokens,
			total_tokens = EXCLUDED.total_tokens,
			updated_at = NOW();
	`, start, end)

	return err
}

// AggregateRange aggregates token usage for a date range (inclusive start, exclusive end).
func (a *TokenUsageAggregator) AggregateRange(ctx context.Context, start, end time.Time) error {
	day := time.Date(start.Year(), start.Month(), start.Day(), 0, 0, 0, 0, time.UTC)
	for day.Before(end) {
		if err := a.AggregateDaily(ctx, day); err != nil {
			return err
		}
		day = day.Add(24 * time.Hour)
	}
	return nil
}

