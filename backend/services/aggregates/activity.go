package aggregates

import (
	"context"
	"time"

	"github.com/bantuaku/backend/services/storage"
)

// ActivityAggregator aggregates activity events (from audit_logs) into activity_aggregates.
type ActivityAggregator struct {
	db *storage.Postgres
}

// NewActivityAggregator creates a new aggregator.
func NewActivityAggregator(db *storage.Postgres) *ActivityAggregator {
	return &ActivityAggregator{db: db}
}

// AggregateDaily aggregates activity.* audit logs for a single day (UTC).
// It upserts into activity_aggregates for fast dashboard queries.
func (a *ActivityAggregator) AggregateDaily(ctx context.Context, day time.Time) error {
	start := time.Date(day.Year(), day.Month(), day.Day(), 0, 0, 0, 0, time.UTC)
	end := start.Add(24 * time.Hour)

	_, err := a.db.Pool().Exec(ctx, `
		INSERT INTO activity_aggregates (date, user_id, company_id, action_type, count, created_at, updated_at)
		SELECT 
			DATE(al.created_at) AS date,
			al.user_id,
			al.company_id,
			al.action AS action_type,
			COUNT(*) AS count,
			NOW(),
			NOW()
		FROM audit_logs al
		WHERE al.action LIKE 'activity.%'
		  AND al.created_at >= $1
		  AND al.created_at < $2
		GROUP BY DATE(al.created_at), al.user_id, al.company_id, al.action
		ON CONFLICT (date, user_id, company_id, action_type) DO UPDATE
		SET count = EXCLUDED.count, updated_at = NOW();
	`, start, end)

	return err
}

// AggregateRange aggregates for a date range (inclusive start, exclusive end).
func (a *ActivityAggregator) AggregateRange(ctx context.Context, start, end time.Time) error {
	day := time.Date(start.Year(), start.Month(), start.Day(), 0, 0, 0, 0, time.UTC)
	for day.Before(end) {
		if err := a.AggregateDaily(ctx, day); err != nil {
			return err
		}
		day = day.Add(24 * time.Hour)
	}
	return nil
}

