package forecast

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Store handles database operations for forecasts
type Store struct {
	pool *pgxpool.Pool
}

// NewStore creates a new forecast store
func NewStore(pool *pgxpool.Pool) *Store {
	return &Store{pool: pool}
}

// StoreMonthlyStrategy stores a monthly strategy
func (s *Store) StoreMonthlyStrategy(
	ctx context.Context,
	productID string,
	forecastID string,
	month int,
	strategyText string,
	actions []byte,
	priority string,
	estimatedImpact []byte,
) (string, error) {
	id := uuid.New().String()
	_, err := s.pool.Exec(ctx,
		`INSERT INTO monthly_strategies 
		 (id, product_id, forecast_id, month, strategy_text, actions, priority, estimated_impact, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW(), NOW())
		 ON CONFLICT DO NOTHING`,
		id, productID, forecastID, month, strategyText, actions, priority, estimatedImpact,
	)
	if err != nil {
		return "", fmt.Errorf("failed to store strategy: %w", err)
	}
	return id, nil
}

// GetMonthlyForecasts retrieves monthly forecasts for a product
func (s *Store) GetMonthlyForecasts(ctx context.Context, productID string, forecastDate *time.Time) ([]MonthlyForecastDB, error) {
	query := `SELECT id, product_id, forecast_input_id, month, forecast_date, predicted_quantity,
	          confidence_lower, confidence_upper, confidence_score, algorithm, model_version, created_at
	          FROM forecasts_monthly WHERE product_id = $1`
	args := []interface{}{productID}
	
	if forecastDate != nil {
		query += " AND forecast_date = $2"
		args = append(args, *forecastDate)
	} else {
		// Get latest forecast
		query += " AND forecast_date = (SELECT MAX(forecast_date) FROM forecasts_monthly WHERE product_id = $1)"
	}
	
	query += " ORDER BY month ASC"

	rows, err := s.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query forecasts: %w", err)
	}
	defer rows.Close()

	var forecasts []MonthlyForecastDB
	for rows.Next() {
		var f MonthlyForecastDB
		err := rows.Scan(
			&f.ID, &f.ProductID, &f.ForecastInputID, &f.Month, &f.ForecastDate,
			&f.PredictedQuantity, &f.ConfidenceLower, &f.ConfidenceUpper,
			&f.ConfidenceScore, &f.Algorithm, &f.ModelVersion, &f.CreatedAt,
		)
		if err != nil {
			continue
		}
		forecasts = append(forecasts, f)
	}

	return forecasts, nil
}

// MonthlyForecastDB represents a monthly forecast from database
type MonthlyForecastDB struct {
	ID               string
	ProductID        string
	ForecastInputID  *string
	Month            int
	ForecastDate     time.Time
	PredictedQuantity int
	ConfidenceLower   *int
	ConfidenceUpper   *int
	ConfidenceScore   float64
	Algorithm         string
	ModelVersion      string
	CreatedAt         time.Time
}

