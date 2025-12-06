package tokenusage

import (
	"context"
	"fmt"
	"time"

	"github.com/bantuaku/backend/services/storage"
	"github.com/google/uuid"
)

// TokenUsage represents token consumption for a single chat completion
type TokenUsage struct {
	ID               string    `json:"id"`
	UserID           *string   `json:"user_id,omitempty"`
	CompanyID        string    `json:"company_id"`
	ConversationID   *string   `json:"conversation_id,omitempty"`
	MessageID        *string   `json:"message_id,omitempty"`
	Model            string    `json:"model"`
	Provider         string    `json:"provider"`
	PromptTokens     int       `json:"prompt_tokens"`
	CompletionTokens int       `json:"completion_tokens"`
	TotalTokens      int       `json:"total_tokens"`
	CreatedAt        time.Time `json:"created_at"`
}

// UsageStats represents aggregated token usage statistics
type UsageStats struct {
	TotalPromptTokens     int          `json:"total_prompt_tokens"`
	TotalCompletionTokens int          `json:"total_completion_tokens"`
	TotalTokens           int          `json:"total_tokens"`
	EstimatedCost         float64      `json:"estimated_cost"` // Estimated cost in IDR
	ModelBreakdown        []ModelUsage `json:"model_breakdown"`
	StartDate             *time.Time   `json:"start_date,omitempty"`
	EndDate               *time.Time   `json:"end_date,omitempty"`
}

// ModelUsage represents token usage per model
type ModelUsage struct {
	Model            string  `json:"model"`
	Provider         string  `json:"provider"`
	PromptTokens     int     `json:"prompt_tokens"`
	CompletionTokens int     `json:"completion_tokens"`
	TotalTokens      int     `json:"total_tokens"`
	EstimatedCost    float64 `json:"estimated_cost"`
	RequestCount     int     `json:"request_count"`
}

// Model pricing (IDR per token) - approximate rates
var modelPricing = map[string]float64{
	"openai/gpt-4o-mini":              0.000001,  // ~0.000001 IDR per token
	"openai/gpt-4o":                   0.00001,   // ~0.00001 IDR per token
	"GLM 4.6":                         0.0000005, // ~0.0000005 IDR per token
	"qwen/qwen-3-vl-30b-a3b-instruct": 0.000002,  // ~0.000002 IDR per token
}

// Service handles token usage tracking
type Service struct {
	db *storage.Postgres
}

// NewService creates a new token usage service
func NewService(db *storage.Postgres) *Service {
	return &Service{db: db}
}

// LogTokenUsage logs token usage for a chat completion
func (s *Service) LogTokenUsage(ctx context.Context, usage *TokenUsage) error {
	_, err := s.db.Pool().Exec(ctx, `
		INSERT INTO token_usage (
			id, user_id, company_id, conversation_id, message_id,
			model, provider, prompt_tokens, completion_tokens, total_tokens, created_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`, usage.ID, usage.UserID, usage.CompanyID, usage.ConversationID, usage.MessageID,
		usage.Model, usage.Provider, usage.PromptTokens, usage.CompletionTokens,
		usage.TotalTokens, usage.CreatedAt)

	return err
}

// CreateTokenUsage creates a new token usage record with auto-generated ID
func (s *Service) CreateTokenUsage(ctx context.Context, userID *string, companyID string, conversationID, messageID *string, model, provider string, promptTokens, completionTokens, totalTokens int) error {
	usage := &TokenUsage{
		ID:               uuid.New().String(),
		UserID:           userID,
		CompanyID:        companyID,
		ConversationID:   conversationID,
		MessageID:        messageID,
		Model:            model,
		Provider:         provider,
		PromptTokens:     promptTokens,
		CompletionTokens: completionTokens,
		TotalTokens:      totalTokens,
		CreatedAt:        time.Now(),
	}
	return s.LogTokenUsage(ctx, usage)
}

// GetUsageStats retrieves aggregated token usage statistics with filters
func (s *Service) GetUsageStats(ctx context.Context, companyID *string, model *string, startDate, endDate *time.Time) (*UsageStats, error) {
	query := `
		SELECT 
			COALESCE(SUM(prompt_tokens), 0) as total_prompt_tokens,
			COALESCE(SUM(completion_tokens), 0) as total_completion_tokens,
			COALESCE(SUM(total_tokens), 0) as total_tokens
		FROM token_usage
		WHERE 1=1
	`
	args := []interface{}{}
	argIndex := 1

	if companyID != nil {
		query += ` AND company_id = $` + fmt.Sprintf("%d", argIndex)
		args = append(args, *companyID)
		argIndex++
	}

	if model != nil {
		query += ` AND model = $` + fmt.Sprintf("%d", argIndex)
		args = append(args, *model)
		argIndex++
	}

	if startDate != nil {
		query += ` AND created_at >= $` + fmt.Sprintf("%d", argIndex)
		args = append(args, *startDate)
		argIndex++
	}

	if endDate != nil {
		query += ` AND created_at <= $` + fmt.Sprintf("%d", argIndex)
		args = append(args, *endDate)
	}

	stats := &UsageStats{}
	err := s.db.Pool().QueryRow(ctx, query, args...).Scan(
		&stats.TotalPromptTokens,
		&stats.TotalCompletionTokens,
		&stats.TotalTokens,
	)
	if err != nil {
		return nil, err
	}

	// Get model breakdown
	breakdownQuery := `
		SELECT 
			model, provider,
			SUM(prompt_tokens) as prompt_tokens,
			SUM(completion_tokens) as completion_tokens,
			SUM(total_tokens) as total_tokens,
			COUNT(*) as request_count
		FROM token_usage
		WHERE 1=1
	`
	breakdownArgs := []interface{}{}
	breakdownArgIndex := 1

	if companyID != nil {
		breakdownQuery += ` AND company_id = $` + fmt.Sprintf("%d", breakdownArgIndex)
		breakdownArgs = append(breakdownArgs, *companyID)
		breakdownArgIndex++
	}

	if model != nil {
		breakdownQuery += ` AND model = $` + fmt.Sprintf("%d", breakdownArgIndex)
		breakdownArgs = append(breakdownArgs, *model)
		breakdownArgIndex++
	}

	if startDate != nil {
		breakdownQuery += ` AND created_at >= $` + fmt.Sprintf("%d", breakdownArgIndex)
		breakdownArgs = append(breakdownArgs, *startDate)
		breakdownArgIndex++
	}

	if endDate != nil {
		breakdownQuery += ` AND created_at <= $` + fmt.Sprintf("%d", breakdownArgIndex)
		breakdownArgs = append(breakdownArgs, *endDate)
	}

	breakdownQuery += ` GROUP BY model, provider ORDER BY total_tokens DESC`

	rows, err := s.db.Pool().Query(ctx, breakdownQuery, breakdownArgs...)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var mu ModelUsage
			if err := rows.Scan(
				&mu.Model, &mu.Provider,
				&mu.PromptTokens, &mu.CompletionTokens, &mu.TotalTokens,
				&mu.RequestCount,
			); err == nil {
				// Calculate estimated cost
				pricePerToken := modelPricing[mu.Model]
				if pricePerToken == 0 {
					pricePerToken = 0.000001 // Default fallback
				}
				mu.EstimatedCost = float64(mu.TotalTokens) * pricePerToken
				stats.ModelBreakdown = append(stats.ModelBreakdown, mu)
			}
		}
	}

	// Calculate total estimated cost
	totalCost := 0.0
	for _, mu := range stats.ModelBreakdown {
		totalCost += mu.EstimatedCost
	}
	stats.EstimatedCost = totalCost

	if startDate != nil {
		stats.StartDate = startDate
	}
	if endDate != nil {
		stats.EndDate = endDate
	}

	return stats, nil
}

// GetTokenUsage retrieves token usage records with filters
func (s *Service) GetTokenUsage(ctx context.Context, companyID *string, model *string, startDate, endDate *time.Time, page, limit int) ([]TokenUsage, int, error) {
	offset := (page - 1) * limit

	query := `
		SELECT id, user_id, company_id, conversation_id, message_id,
			model, provider, prompt_tokens, completion_tokens, total_tokens, created_at
		FROM token_usage
		WHERE 1=1
	`
	args := []interface{}{}
	argIndex := 1

	if companyID != nil {
		query += ` AND company_id = $` + fmt.Sprintf("%d", argIndex)
		args = append(args, *companyID)
		argIndex++
	}

	if model != nil {
		query += ` AND model = $` + fmt.Sprintf("%d", argIndex)
		args = append(args, *model)
		argIndex++
	}

	if startDate != nil {
		query += ` AND created_at >= $` + fmt.Sprintf("%d", argIndex)
		args = append(args, *startDate)
		argIndex++
	}

	if endDate != nil {
		query += ` AND created_at <= $` + fmt.Sprintf("%d", argIndex)
		args = append(args, *endDate)
		argIndex++
	}

	query += ` ORDER BY created_at DESC LIMIT $` + fmt.Sprintf("%d", argIndex) + ` OFFSET $` + fmt.Sprintf("%d", argIndex+1)
	args = append(args, limit, offset)

	rows, err := s.db.Pool().Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var usages []TokenUsage
	for rows.Next() {
		var usage TokenUsage
		var conversationID, messageID interface{}
		if err := rows.Scan(
			&usage.ID, &usage.UserID, &usage.CompanyID, &conversationID, &messageID,
			&usage.Model, &usage.Provider,
			&usage.PromptTokens, &usage.CompletionTokens, &usage.TotalTokens,
			&usage.CreatedAt,
		); err != nil {
			continue
		}
		if convID, ok := conversationID.(string); ok && convID != "" {
			usage.ConversationID = &convID
		}
		if msgID, ok := messageID.(string); ok && msgID != "" {
			usage.MessageID = &msgID
		}
		usages = append(usages, usage)
	}

	// Get total count
	countQuery := `SELECT COUNT(*) FROM token_usage WHERE 1=1`
	countArgs := []interface{}{}
	countArgIndex := 1

	if companyID != nil {
		countQuery += ` AND company_id = $` + fmt.Sprintf("%d", countArgIndex)
		countArgs = append(countArgs, *companyID)
		countArgIndex++
	}

	if model != nil {
		countQuery += ` AND model = $` + fmt.Sprintf("%d", countArgIndex)
		countArgs = append(countArgs, *model)
		countArgIndex++
	}

	if startDate != nil {
		countQuery += ` AND created_at >= $` + fmt.Sprintf("%d", countArgIndex)
		countArgs = append(countArgs, *startDate)
		countArgIndex++
	}

	if endDate != nil {
		countQuery += ` AND created_at <= $` + fmt.Sprintf("%d", countArgIndex)
		countArgs = append(countArgs, *endDate)
	}

	var total int
	if err := s.db.Pool().QueryRow(ctx, countQuery, countArgs...).Scan(&total); err != nil {
		// Log error but continue with total = 0 for pagination
		total = 0
	}

	return usages, total, nil
}
