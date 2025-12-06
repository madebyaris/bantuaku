package forecast

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/bantuaku/backend/logger"
)

// Adapter connects Go backend to Python forecasting service
type Adapter struct {
	baseURL    string
	httpClient *http.Client
	log        logger.Logger
}

// NewAdapter creates a new forecast adapter
func NewAdapter(baseURL string) *Adapter {
	log := logger.Default()
	return &Adapter{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
		log: *log,
	}
}

// ForecastInputs represents inputs for forecasting
type ForecastInputs struct {
	ProductID       string                 `json:"product_id"`
	SalesHistory    []SalesHistoryPoint    `json:"sales_history"`
	TrendsData      []TrendSignal          `json:"trends_data,omitempty"`
	RegulationFlags []RegulationFlag      `json:"regulation_flags,omitempty"`
	ExogenousFactors map[string]interface{} `json:"exogenous_factors,omitempty"`
}

// SalesHistoryPoint represents a sales data point
type SalesHistoryPoint struct {
	Date     string `json:"date"`
	Quantity int    `json:"quantity"`
}

// TrendSignal represents a trend signal
type TrendSignal struct {
	Keyword string `json:"keyword"`
	Time    string `json:"time"`
	Value   int    `json:"value"`
}

// RegulationFlag represents a regulation flag
type RegulationFlag struct {
	RegulationID   string  `json:"regulation_id"`
	RelevanceScore float64 `json:"relevance_score"`
	Impact         string  `json:"impact"`
}

// ForecastResponse represents forecast response from Python service
type ForecastResponse struct {
	ProductID    string           `json:"product_id"`
	ForecastDate string           `json:"forecast_date"`
	Algorithm    string           `json:"algorithm"`
	ModelVersion string           `json:"model_version"`
	Forecasts    []MonthlyForecast `json:"forecasts"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

// MonthlyForecast represents a monthly forecast
type MonthlyForecast struct {
	Month             int     `json:"month"`
	PredictedQuantity int     `json:"predicted_quantity"`
	ConfidenceLower   *int    `json:"confidence_lower,omitempty"`
	ConfidenceUpper   *int    `json:"confidence_upper,omitempty"`
	ConfidenceScore   float64 `json:"confidence_score"`
}

// GenerateForecast calls Python forecasting service
func (a *Adapter) GenerateForecast(ctx context.Context, inputs ForecastInputs) (*ForecastResponse, error) {
	url := fmt.Sprintf("%s/forecast", a.baseURL)

	reqBody, err := json.Marshal(inputs)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("forecast service error: %d - %s", resp.StatusCode, string(body))
	}

	var forecastResp ForecastResponse
	if err := json.NewDecoder(resp.Body).Decode(&forecastResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &forecastResp, nil
}

// HealthCheck checks if forecasting service is healthy
func (a *Adapter) HealthCheck(ctx context.Context) error {
	url := fmt.Sprintf("%s/health", a.baseURL)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return err
	}

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("health check failed: status %d", resp.StatusCode)
	}

	return nil
}

