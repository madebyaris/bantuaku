package scheduler

import (
	"context"
	"fmt"
	"time"

	"github.com/bantuaku/backend/logger"
	"github.com/bantuaku/backend/services/forecast"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ForecastScheduler handles scheduled forecast generation
type ForecastScheduler struct {
	forecastService *forecast.Service
	pool            *pgxpool.Pool
	log             logger.Logger
}

// NewForecastScheduler creates a new forecast scheduler
func NewForecastScheduler(forecastService *forecast.Service, pool *pgxpool.Pool) *ForecastScheduler {
	log := logger.Default()
	return &ForecastScheduler{
		forecastService: forecastService,
		pool:            pool,
		log:             *log,
	}
}

// RunMonthlyForecastJob generates forecasts for all products
func (s *ForecastScheduler) RunMonthlyForecastJob(ctx context.Context) error {
	s.log.Info("Starting monthly forecast job")

	// Get all active products
	rows, err := s.pool.Query(ctx,
		`SELECT DISTINCT p.id 
		 FROM products p
		 JOIN sales_history s ON s.product_id = p.id
		 WHERE p.is_active = true
		 GROUP BY p.id
		 HAVING COUNT(s.id) >= 7`,
	)
	if err != nil {
		return fmt.Errorf("failed to get products: %w", err)
	}
	defer rows.Close()

	var productIDs []string
	for rows.Next() {
		var productID string
		if err := rows.Scan(&productID); err != nil {
			continue
		}
		productIDs = append(productIDs, productID)
	}

	s.log.Info("Generating forecasts", "products", len(productIDs))

	// Generate forecast for each product
	successCount := 0
	errorCount := 0

	for _, productID := range productIDs {
		forecastResp, err := s.forecastService.GenerateMonthlyForecast(ctx, productID)
		if err != nil {
			s.log.Warn("Failed to generate forecast", "product_id", productID, "error", err)
			errorCount++
			continue
		}

		// Store forecast
		_, err = s.forecastService.StoreForecast(ctx, forecastResp)
		if err != nil {
			s.log.Warn("Failed to store forecast", "product_id", productID, "error", err)
			errorCount++
			continue
		}

		successCount++
		
		// Rate limiting
		time.Sleep(500 * time.Millisecond)
	}

	s.log.Info("Monthly forecast job completed",
		"success", successCount,
		"errors", errorCount,
		"total", len(productIDs),
	)

	return nil
}

// StartMonthlySchedule starts monthly forecast generation schedule
func (s *ForecastScheduler) StartMonthlySchedule(ctx context.Context, scheduleTime time.Time) {
	ticker := time.NewTicker(30 * 24 * time.Hour) // Monthly
	defer ticker.Stop()

	// Calculate time until first run
	now := time.Now()
	firstRun := scheduleTime
	if scheduleTime.Before(now) {
		firstRun = scheduleTime.Add(30 * 24 * time.Hour)
	}
	duration := firstRun.Sub(now)

	s.log.Info("Scheduling monthly forecast job", "first_run", firstRun, "duration", duration)

	// Wait for first run
	time.Sleep(duration)

	// Run immediately
	if err := s.RunMonthlyForecastJob(ctx); err != nil {
		s.log.Error("Monthly forecast job failed", "error", err)
	}

	// Then run on schedule
	for {
		select {
		case <-ctx.Done():
			s.log.Info("Stopping monthly forecast scheduler")
			return
		case <-ticker.C:
			if err := s.RunMonthlyForecastJob(ctx); err != nil {
				s.log.Error("Monthly forecast job failed", "error", err)
			}
		}
	}
}

