package forecast

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/bantuaku/backend/logger"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Service handles forecasting operations
type Service struct {
	adapter *Adapter
	pool    *pgxpool.Pool
	log     logger.Logger
}

// NewService creates a new forecast service
func NewService(adapter *Adapter, pool *pgxpool.Pool) *Service {
	log := logger.Default()
	return &Service{
		adapter: adapter,
		pool:    pool,
		log:     *log,
	}
}

// GenerateMonthlyForecast generates 12-month forecast for a product
func (s *Service) GenerateMonthlyForecast(ctx context.Context, productID string) (*ForecastResponse, error) {
	// Get sales history
	salesHistory, err := s.getSalesHistory(ctx, productID)
	if err != nil {
		return nil, fmt.Errorf("failed to get sales history: %w", err)
	}

	if len(salesHistory) < 7 {
		return nil, fmt.Errorf("insufficient sales data: need at least 7 data points, got %d", len(salesHistory))
	}

	// Get trends data (if available)
	trendsData, err := s.getTrendsData(ctx, productID)
	if err != nil {
		s.log.Warn("Failed to get trends data", "error", err)
		trendsData = []TrendSignal{}
	}

	// Get regulation flags (if available)
	regulationFlags, err := s.getRegulationFlags(ctx, productID)
	if err != nil {
		s.log.Warn("Failed to get regulation flags", "error", err)
		regulationFlags = []RegulationFlag{}
	}

	// Build forecast inputs
	inputs := ForecastInputs{
		ProductID:       productID,
		SalesHistory:    salesHistory,
		TrendsData:      trendsData,
		RegulationFlags: regulationFlags,
	}

	// Call forecasting service
	forecastResp, err := s.adapter.GenerateForecast(ctx, inputs)
	if err != nil {
		return nil, fmt.Errorf("failed to generate forecast: %w", err)
	}

	return forecastResp, nil
}

// StoreForecast stores forecast in database
func (s *Service) StoreForecast(ctx context.Context, forecastResp *ForecastResponse) (string, error) {
	forecastDate := time.Now()
	forecastID := uuid.New().String()

	// Store forecast inputs
	inputID := uuid.New().String()
	salesHistoryJSON, _ := json.Marshal(forecastResp.Forecasts)
	trendsJSON, _ := json.Marshal([]TrendSignal{})
	regulationJSON, _ := json.Marshal([]RegulationFlag{})

	_, err := s.pool.Exec(ctx,
		`INSERT INTO forecast_inputs 
		 (id, product_id, forecast_period_start, forecast_period_end, sales_history, trends_data, regulation_flags, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, NOW())`,
		inputID, forecastResp.ProductID,
		forecastDate, forecastDate.AddDate(1, 0, 0),
		salesHistoryJSON, trendsJSON, regulationJSON,
	)
	if err != nil {
		return "", fmt.Errorf("failed to store forecast inputs: %w", err)
	}

	// Store monthly forecasts
	for _, monthlyForecast := range forecastResp.Forecasts {
		_, err := s.pool.Exec(ctx,
			`INSERT INTO forecasts_monthly 
			 (id, product_id, forecast_input_id, month, forecast_date, predicted_quantity, 
			  confidence_lower, confidence_upper, confidence_score, algorithm, model_version, created_at)
			 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, NOW())
			 ON CONFLICT (product_id, forecast_date, month) DO UPDATE
			 SET predicted_quantity = EXCLUDED.predicted_quantity,
			     confidence_lower = EXCLUDED.confidence_lower,
			     confidence_upper = EXCLUDED.confidence_upper,
			     confidence_score = EXCLUDED.confidence_score`,
			uuid.New().String(), forecastResp.ProductID, inputID,
			monthlyForecast.Month, forecastDate,
			monthlyForecast.PredictedQuantity,
			monthlyForecast.ConfidenceLower,
			monthlyForecast.ConfidenceUpper,
			monthlyForecast.ConfidenceScore,
			forecastResp.Algorithm,
			forecastResp.ModelVersion,
		)
		if err != nil {
			s.log.Warn("Failed to store monthly forecast", "month", monthlyForecast.Month, "error", err)
			continue
		}
	}

	return forecastID, nil
}

// GetSalesHistory retrieves sales history for a product (public method)
func (s *Service) GetSalesHistory(ctx context.Context, productID string) ([]SalesHistoryPoint, error) {
	return s.getSalesHistory(ctx, productID)
}

// getSalesHistory retrieves sales history for a product
func (s *Service) getSalesHistory(ctx context.Context, productID string) ([]SalesHistoryPoint, error) {
	rows, err := s.pool.Query(ctx,
		`SELECT sale_date, SUM(quantity) as total_qty
		 FROM sales_history
		 WHERE product_id = $1
		 GROUP BY sale_date
		 ORDER BY sale_date ASC`,
		productID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var history []SalesHistoryPoint
	for rows.Next() {
		var date time.Time
		var qty int
		if err := rows.Scan(&date, &qty); err != nil {
			continue
		}
		history = append(history, SalesHistoryPoint{
			Date:     date.Format("2006-01-02"),
			Quantity: qty,
		})
	}

	return history, nil
}

// getTrendsData retrieves trends data for a product
func (s *Service) getTrendsData(ctx context.Context, productID string) ([]TrendSignal, error) {
	// Get product category or keywords from company
	// For now, return empty - can be enhanced to link products to trends keywords
	return []TrendSignal{}, nil
}

// getRegulationFlags retrieves regulation flags for a product
func (s *Service) getRegulationFlags(ctx context.Context, productID string) ([]RegulationFlag, error) {
	// For now, return empty - can be enhanced to check relevant regulations
	return []RegulationFlag{}, nil
}

