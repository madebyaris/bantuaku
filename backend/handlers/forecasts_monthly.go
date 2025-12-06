package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/bantuaku/backend/errors"
	"github.com/bantuaku/backend/logger"
	"github.com/bantuaku/backend/services/forecast"
	"github.com/bantuaku/backend/services/kolosal"
	"github.com/bantuaku/backend/services/strategy"
)

// GetMonthlyForecasts retrieves 12-month forecasts for a product
func (h *Handler) GetMonthlyForecasts(w http.ResponseWriter, r *http.Request) {
	log := logger.With("request_id", r.Context().Value("request_id"))
	productID := r.URL.Query().Get("product_id")
	if productID == "" {
		h.respondError(w, errors.NewValidationError("product_id is required", ""), r)
		return
	}

	// Parse optional forecast_date
	var forecastDate *time.Time
	if dateStr := r.URL.Query().Get("forecast_date"); dateStr != "" {
		if t, err := time.Parse("2006-01-02", dateStr); err == nil {
			forecastDate = &t
		}
	}

	// Get forecasts from database
	store := forecast.NewStore(h.db.Pool())
	forecasts, err := store.GetMonthlyForecasts(r.Context(), productID, forecastDate)
	if err != nil {
		log.Error("Failed to get monthly forecasts", "error", err)
		h.respondError(w, errors.NewInternalError(err, "Failed to get forecasts"), r)
		return
	}

	// Convert to response format
	responseForecasts := make([]map[string]interface{}, len(forecasts))
	for i, f := range forecasts {
		responseForecasts[i] = map[string]interface{}{
			"id":                f.ID,
			"month":            f.Month,
			"predicted_quantity": f.PredictedQuantity,
			"confidence_lower":  f.ConfidenceLower,
			"confidence_upper":  f.ConfidenceUpper,
			"confidence_score":  f.ConfidenceScore,
			"algorithm":        f.Algorithm,
			"forecast_date":    f.ForecastDate.Format("2006-01-02"),
		}
	}

	h.respondJSON(w, http.StatusOK, map[string]interface{}{
		"product_id": productID,
		"forecasts": responseForecasts,
		"count":     len(responseForecasts),
	})
}

// GenerateMonthlyForecast generates a new 12-month forecast
func (h *Handler) GenerateMonthlyForecast(w http.ResponseWriter, r *http.Request) {
	log := logger.With("request_id", r.Context().Value("request_id"))
	productID := r.URL.Query().Get("product_id")
	if productID == "" {
		h.respondError(w, errors.NewValidationError("product_id is required", ""), r)
		return
	}

	// Create forecast adapter
	adapter := forecast.NewAdapter(h.config.ForecastingServiceURL)

	// Create forecast service
	forecastService := forecast.NewService(adapter, h.db.Pool())

	// Generate forecast
	forecastResp, err := forecastService.GenerateMonthlyForecast(r.Context(), productID)
	if err != nil {
		log.Error("Failed to generate forecast", "error", err)
		h.respondError(w, errors.NewInternalError(err, "Failed to generate forecast"), r)
		return
	}

	// Store forecast
	forecastID, err := forecastService.StoreForecast(r.Context(), forecastResp)
	if err != nil {
		log.Warn("Failed to store forecast", "error", err)
		// Continue - forecast is still returned
	}

	log.Info("Monthly forecast generated", "product_id", productID, "forecast_id", forecastID)

	h.respondJSON(w, http.StatusOK, map[string]interface{}{
		"forecast_id": forecastID,
		"product_id":  productID,
		"forecasts":   forecastResp.Forecasts,
		"algorithm":   forecastResp.Algorithm,
		"forecast_date": forecastResp.ForecastDate,
	})
}

// GetMonthlyStrategies retrieves monthly strategies for a product
func (h *Handler) GetMonthlyStrategies(w http.ResponseWriter, r *http.Request) {
	log := logger.With("request_id", r.Context().Value("request_id"))
	productID := r.URL.Query().Get("product_id")
	if productID == "" {
		h.respondError(w, errors.NewValidationError("product_id is required", ""), r)
		return
	}

	// Get strategies from database
	pool := h.db.Pool()
	rows, err := pool.Query(r.Context(),
		`SELECT id, product_id, forecast_id, month, strategy_text, actions, priority, estimated_impact, created_at
		 FROM monthly_strategies
		 WHERE product_id = $1
		 ORDER BY month ASC`,
		productID,
	)
	if err != nil {
		log.Error("Failed to get strategies", "error", err)
		h.respondError(w, errors.NewInternalError(err, "Failed to get strategies"), r)
		return
	}
	defer rows.Close()

	var strategies []map[string]interface{}
	for rows.Next() {
		var id, productID, strategyText, priority string
		var forecastID *string
		var month int
		var actions, estimatedImpact []byte
		var createdAt time.Time

		if err := rows.Scan(&id, &productID, &forecastID, &month, &strategyText, &actions, &priority, &estimatedImpact, &createdAt); err != nil {
			continue
		}

		var actionsMap map[string]interface{}
		json.Unmarshal(actions, &actionsMap)

		var impactMap map[string]interface{}
		json.Unmarshal(estimatedImpact, &impactMap)

		strategies = append(strategies, map[string]interface{}{
			"id":               id,
			"product_id":      productID,
			"forecast_id":     forecastID,
			"month":           month,
			"strategy_text":   strategyText,
			"actions":         actionsMap,
			"priority":        priority,
			"estimated_impact": impactMap,
			"created_at":      createdAt,
		})
	}

	h.respondJSON(w, http.StatusOK, map[string]interface{}{
		"product_id": productID,
		"strategies": strategies,
		"count":     len(strategies),
	})
}

// GenerateStrategies generates strategies from a forecast
func (h *Handler) GenerateStrategies(w http.ResponseWriter, r *http.Request) {
	log := logger.With("request_id", r.Context().Value("request_id"))
	productID := r.URL.Query().Get("product_id")
	if productID == "" {
		h.respondError(w, errors.NewValidationError("product_id is required", ""), r)
		return
	}

	forecastDateStr := r.URL.Query().Get("forecast_date")
	var forecastDate *time.Time
	if forecastDateStr != "" {
		if t, err := time.Parse("2006-01-02", forecastDateStr); err == nil {
			forecastDate = &t
		}
	}

	// Get latest forecast
	store := forecast.NewStore(h.db.Pool())
	forecasts, err := store.GetMonthlyForecasts(r.Context(), productID, forecastDate)
	if err != nil || len(forecasts) == 0 {
		h.respondError(w, errors.NewValidationError("No forecast found. Generate forecast first.", ""), r)
		return
	}

	// Convert to ForecastResponse format
	forecastResp := &forecast.ForecastResponse{
		ProductID: productID,
		Forecasts: make([]forecast.MonthlyForecast, len(forecasts)),
	}
	for i, f := range forecasts {
		forecastResp.Forecasts[i] = forecast.MonthlyForecast{
			Month:            f.Month,
			PredictedQuantity: f.PredictedQuantity,
			ConfidenceLower:   f.ConfidenceLower,
			ConfidenceUpper:   f.ConfidenceUpper,
			ConfidenceScore:   f.ConfidenceScore,
		}
	}

	// Generate strategies
	kolosalClient := kolosal.NewClient(h.config.KolosalAPIKey)
	generator := strategy.NewGenerator(kolosalClient)
	strategies, err := generator.GenerateStrategies(r.Context(), productID, forecastResp)
	if err != nil {
		log.Error("Failed to generate strategies", "error", err)
		h.respondError(w, errors.NewInternalError(err, "Failed to generate strategies"), r)
		return
	}

	// Store strategies
	forecastID := forecasts[0].ForecastInputID
	if forecastID == nil {
		forecastIDStr := ""
		forecastID = &forecastIDStr
	}

	for _, strat := range strategies {
		_, err := store.StoreMonthlyStrategy(
			r.Context(),
			strat.ProductID,
			*forecastID,
			strat.Month,
			strat.StrategyText,
			strat.Actions,
			strat.Priority,
			strat.EstimatedImpact,
		)
		if err != nil {
			log.Warn("Failed to store strategy", "month", strat.Month, "error", err)
		}
	}

	log.Info("Strategies generated", "product_id", productID, "count", len(strategies))

	h.respondJSON(w, http.StatusOK, map[string]interface{}{
		"product_id": productID,
		"strategies": strategies,
		"count":     len(strategies),
	})
}

