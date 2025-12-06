package usage

import (
	"context"
	"fmt"
	"time"

	"github.com/bantuaku/backend/services/storage"
)

// PlanLimits represents the limits for a subscription plan
type PlanLimits struct {
	MaxChatsPerMonth             *int
	MaxFileUploadsPerMonth       *int
	MaxFileSizeMB                *int
	MaxForecastRefreshesPerMonth *int
}

// UsageStats represents current usage for a company
type UsageStats struct {
	ChatsThisMonth     int
	UploadsThisMonth   int
	ForecastsThisMonth int
}

// Service handles usage tracking and limit checking
type Service struct {
	db *storage.Postgres
}

// NewService creates a new usage service
func NewService(db *storage.Postgres) *Service {
	return &Service{db: db}
}

// GetPlanLimits retrieves the limits for a company's subscription plan
func (s *Service) GetPlanLimits(ctx context.Context, companyID string) (*PlanLimits, error) {
	var limits PlanLimits

	err := s.db.Pool().QueryRow(ctx, `
		SELECT 
			sp.max_chats_per_month,
			sp.max_file_uploads_per_month,
			sp.max_file_size_mb,
			sp.max_forecast_refreshes_per_month
		FROM companies c
		LEFT JOIN subscriptions sub ON sub.company_id = c.id AND sub.status = 'active'
		LEFT JOIN subscription_plans sp ON sp.id = sub.plan_id
		WHERE c.id = $1
	`, companyID).Scan(
		&limits.MaxChatsPerMonth,
		&limits.MaxFileUploadsPerMonth,
		&limits.MaxFileSizeMB,
		&limits.MaxForecastRefreshesPerMonth,
	)

	if err != nil {
		// If no subscription, get free plan limits
		err = s.db.Pool().QueryRow(ctx, `
			SELECT 
				max_chats_per_month,
				max_file_uploads_per_month,
				max_file_size_mb,
				max_forecast_refreshes_per_month
			FROM subscription_plans
			WHERE name = 'free'
		`).Scan(
			&limits.MaxChatsPerMonth,
			&limits.MaxFileUploadsPerMonth,
			&limits.MaxFileSizeMB,
			&limits.MaxForecastRefreshesPerMonth,
		)
		if err != nil {
			// Default limits if free plan not found
			defaultChats := 50
			defaultUploads := 5
			defaultFileSize := 5
			defaultForecasts := 10
			limits.MaxChatsPerMonth = &defaultChats
			limits.MaxFileUploadsPerMonth = &defaultUploads
			limits.MaxFileSizeMB = &defaultFileSize
			limits.MaxForecastRefreshesPerMonth = &defaultForecasts
		}
	}

	return &limits, nil
}

// GetUsageStats retrieves current usage stats for a company
func (s *Service) GetUsageStats(ctx context.Context, companyID string) (*UsageStats, error) {
	stats := &UsageStats{}

	// Get start of current month
	now := time.Now()
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)

	// Count chat messages this month
	err := s.db.Pool().QueryRow(ctx, `
		SELECT COUNT(*) 
		FROM messages m
		JOIN conversations c ON c.id = m.conversation_id
		WHERE c.company_id = $1 
		AND m.sender = 'user'
		AND m.created_at >= $2
	`, companyID, monthStart).Scan(&stats.ChatsThisMonth)
	if err != nil {
		stats.ChatsThisMonth = 0
	}

	// Count file uploads this month
	err = s.db.Pool().QueryRow(ctx, `
		SELECT COUNT(*) 
		FROM file_uploads
		WHERE company_id = $1 
		AND created_at >= $2
	`, companyID, monthStart).Scan(&stats.UploadsThisMonth)
	if err != nil {
		stats.UploadsThisMonth = 0
	}

	// Count forecast refreshes this month
	err = s.db.Pool().QueryRow(ctx, `
		SELECT COUNT(*) 
		FROM forecasts
		WHERE company_id = $1 
		AND created_at >= $2
	`, companyID, monthStart).Scan(&stats.ForecastsThisMonth)
	if err != nil {
		stats.ForecastsThisMonth = 0
	}

	return stats, nil
}

// CheckChatLimit checks if the company can send more chat messages
func (s *Service) CheckChatLimit(ctx context.Context, companyID string) (bool, string, error) {
	limits, err := s.GetPlanLimits(ctx, companyID)
	if err != nil {
		return false, "", err
	}

	// Unlimited if nil
	if limits.MaxChatsPerMonth == nil {
		return true, "", nil
	}

	stats, err := s.GetUsageStats(ctx, companyID)
	if err != nil {
		return false, "", err
	}

	if stats.ChatsThisMonth >= *limits.MaxChatsPerMonth {
		return false, fmt.Sprintf("Chat limit reached (%d/%d messages this month). Upgrade your plan for more.", stats.ChatsThisMonth, *limits.MaxChatsPerMonth), nil
	}

	return true, "", nil
}

// CheckUploadLimit checks if the company can upload more files
func (s *Service) CheckUploadLimit(ctx context.Context, companyID string) (bool, string, error) {
	limits, err := s.GetPlanLimits(ctx, companyID)
	if err != nil {
		return false, "", err
	}

	// Unlimited if nil
	if limits.MaxFileUploadsPerMonth == nil {
		return true, "", nil
	}

	stats, err := s.GetUsageStats(ctx, companyID)
	if err != nil {
		return false, "", err
	}

	if stats.UploadsThisMonth >= *limits.MaxFileUploadsPerMonth {
		return false, fmt.Sprintf("Upload limit reached (%d/%d files this month). Upgrade your plan for more.", stats.UploadsThisMonth, *limits.MaxFileUploadsPerMonth), nil
	}

	return true, "", nil
}

// CheckFileSizeLimit checks if the file size is within the plan limit
func (s *Service) CheckFileSizeLimit(ctx context.Context, companyID string, fileSizeBytes int64) (bool, string, error) {
	limits, err := s.GetPlanLimits(ctx, companyID)
	if err != nil {
		return false, "", err
	}

	// Unlimited if nil
	if limits.MaxFileSizeMB == nil {
		return true, "", nil
	}

	maxSizeBytes := int64(*limits.MaxFileSizeMB) * 1024 * 1024
	if fileSizeBytes > maxSizeBytes {
		return false, fmt.Sprintf("File size exceeds limit (%d MB max). Upgrade your plan for larger files.", *limits.MaxFileSizeMB), nil
	}

	return true, "", nil
}

// CheckForecastLimit checks if the company can refresh more forecasts
func (s *Service) CheckForecastLimit(ctx context.Context, companyID string) (bool, string, error) {
	limits, err := s.GetPlanLimits(ctx, companyID)
	if err != nil {
		return false, "", err
	}

	// Unlimited if nil
	if limits.MaxForecastRefreshesPerMonth == nil {
		return true, "", nil
	}

	stats, err := s.GetUsageStats(ctx, companyID)
	if err != nil {
		return false, "", err
	}

	if stats.ForecastsThisMonth >= *limits.MaxForecastRefreshesPerMonth {
		return false, fmt.Sprintf("Forecast refresh limit reached (%d/%d this month). Upgrade your plan for more.", stats.ForecastsThisMonth, *limits.MaxForecastRefreshesPerMonth), nil
	}

	return true, "", nil
}
