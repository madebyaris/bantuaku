package strategy

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/bantuaku/backend/logger"
	"github.com/bantuaku/backend/services/forecast"
	"github.com/bantuaku/backend/services/kolosal"
)

// Generator generates monthly strategies from forecasts
type Generator struct {
	kolosalClient *kolosal.Client
	log           logger.Logger
}

// NewGenerator creates a new strategy generator
func NewGenerator(kolosalClient *kolosal.Client) *Generator {
	log := logger.Default()
	return &Generator{
		kolosalClient: kolosalClient,
		log:           *log,
	}
}

// MonthlyStrategy represents a strategy for a month
type MonthlyStrategy struct {
	ProductID      string
	Month          int
	ForecastID     *string
	StrategyText   string
	Actions        json.RawMessage
	Priority       string
	EstimatedImpact json.RawMessage
}

// GenerateStrategies generates strategies for all 12 months
func (g *Generator) GenerateStrategies(
	ctx context.Context,
	productID string,
	forecast *forecast.ForecastResponse,
) ([]MonthlyStrategy, error) {
	var strategies []MonthlyStrategy

	for _, monthlyForecast := range forecast.Forecasts {
		strategy := g.generateMonthlyStrategy(ctx, productID, monthlyForecast, forecast)
		strategies = append(strategies, strategy)
	}

	return strategies, nil
}

// generateMonthlyStrategy generates strategy for a single month
func (g *Generator) generateMonthlyStrategy(
	ctx context.Context,
	productID string,
	monthlyForecast forecast.MonthlyForecast,
	forecastResp *forecast.ForecastResponse,
) MonthlyStrategy {
	predicted := monthlyForecast.PredictedQuantity

	// Determine strategy type
	var strategyText string
	var actions map[string]interface{}
	var priority string
	var estimatedImpact map[string]interface{}

	if predicted == 0 {
		strategyText = fmt.Sprintf("Bulan %d: Proyeksi permintaan sangat rendah. Pertimbangkan untuk mengurangi stok atau melakukan promosi untuk meningkatkan penjualan.", monthlyForecast.Month)
		actions = map[string]interface{}{
			"pricing": map[string]interface{}{
				"action":     "reduce",
				"percentage": 10,
				"reason":     "Meningkatkan daya tarik produk dengan harga lebih kompetitif",
			},
			"inventory": map[string]interface{}{
				"action":   "reduce",
				"quantity": 0,
				"reason":   "Menghindari overstock",
			},
			"marketing": map[string]interface{}{
				"channels": []string{"social", "email"},
				"budget":   200000,
				"reason":   "Meningkatkan awareness produk",
			},
		}
		priority = "low"
		estimatedImpact = map[string]interface{}{
			"sales_increase":  "5-10%",
			"cost_reduction":  "10-15%",
		}
	} else if predicted < 50 {
		strategyText = fmt.Sprintf("Bulan %d: Proyeksi permintaan rendah (%d unit). Fokus pada optimasi stok dan pemasaran bertarget.", monthlyForecast.Month, predicted)
		actions = map[string]interface{}{
			"pricing": map[string]interface{}{
				"action":     "maintain",
				"percentage": 0,
				"reason":     "Mempertahankan harga saat ini",
			},
			"inventory": map[string]interface{}{
				"action":   "optimize",
				"quantity": predicted,
				"reason":   fmt.Sprintf("Menjaga stok sesuai proyeksi (%d unit)", predicted),
			},
			"marketing": map[string]interface{}{
				"channels": []string{"social"},
				"budget":   300000,
				"reason":   "Meningkatkan visibilitas produk",
			},
		}
		priority = "medium"
		estimatedImpact = map[string]interface{}{
			"sales_increase":        "10-15%",
			"inventory_optimization": "20-30%",
		}
	} else if predicted < 200 {
		strategyText = fmt.Sprintf("Bulan %d: Proyeksi permintaan sedang (%d unit). Pertahankan operasi normal dengan monitoring ketat.", monthlyForecast.Month, predicted)
		actions = map[string]interface{}{
			"pricing": map[string]interface{}{
				"action":     "maintain",
				"percentage": 0,
				"reason":     "Harga sudah optimal",
			},
			"inventory": map[string]interface{}{
				"action":   "restock",
				"quantity": predicted + 20,
				"reason":   "Memastikan ketersediaan dengan safety stock",
			},
			"marketing": map[string]interface{}{
				"channels": []string{"social", "email"},
				"budget":   500000,
				"reason":   "Mempertahankan momentum penjualan",
			},
		}
		priority = "medium"
		estimatedImpact = map[string]interface{}{
			"sales_stability":      "±5%",
			"customer_satisfaction": "High",
		}
	} else {
		strategyText = fmt.Sprintf("Bulan %d: Proyeksi permintaan tinggi (%d unit). Siapkan stok yang cukup dan pertimbangkan peningkatan kapasitas.", monthlyForecast.Month, predicted)
		actions = map[string]interface{}{
			"pricing": map[string]interface{}{
				"action":     "increase",
				"percentage": 5,
				"reason":     "Mengoptimalkan margin pada permintaan tinggi",
			},
			"inventory": map[string]interface{}{
				"action":   "restock",
				"quantity": predicted + 50,
				"reason":   "Mencegah stockout dengan stok yang memadai",
			},
			"marketing": map[string]interface{}{
				"channels": []string{"social", "email", "marketplace"},
				"budget":   1000000,
				"reason":   "Maksimalkan penjualan pada periode permintaan tinggi",
			},
		}
		priority = "high"
		estimatedImpact = map[string]interface{}{
			"sales_increase":  "20-30%",
			"revenue_increase": "25-35%",
		}
	}

	// Convert to JSON
	actionsJSON, _ := json.Marshal(actions)
	impactJSON, _ := json.Marshal(estimatedImpact)

	return MonthlyStrategy{
		ProductID:       productID,
		Month:          monthlyForecast.Month,
		ForecastID:     nil, // Will be set when stored
		StrategyText:   strategyText,
		Actions:        actionsJSON,
		Priority:       priority,
		EstimatedImpact: impactJSON,
	}
}

