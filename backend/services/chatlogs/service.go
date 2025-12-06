package chatlogs

import (
	"context"
	"fmt"
	"time"

	"github.com/bantuaku/backend/services/storage"
	"github.com/google/uuid"
)

// UsageLog represents aggregate chat usage statistics
type UsageLog struct {
	ID                 string    `json:"id"`
	CompanyID          string    `json:"company_id"`
	Date               time.Time `json:"date"`
	TotalMessages      int       `json:"total_messages"`
	TotalConversations int       `json:"total_conversations"`
	UniqueUsers        int       `json:"unique_users"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}

// UsageStats represents aggregated usage statistics
type UsageStats struct {
	TotalMessages      int       `json:"total_messages"`
	TotalConversations int       `json:"total_conversations"`
	UniqueUsers        int       `json:"unique_users"`
	Period             string    `json:"period"` // "daily", "monthly"
	StartDate          time.Time `json:"start_date"`
	EndDate            time.Time `json:"end_date"`
}

// Service handles chat usage logging and aggregation
type Service struct {
	db *storage.Postgres
}

// NewService creates a new chat logs service
func NewService(db *storage.Postgres) *Service {
	return &Service{db: db}
}

// LogDailyUsage creates or updates daily usage log for a company
func (s *Service) LogDailyUsage(ctx context.Context, companyID string, date time.Time) error {
	// Aggregate stats from messages and conversations tables
	var totalMessages, totalConversations, uniqueUsers int

	// Count messages for this company on this date
	err := s.db.Pool().QueryRow(ctx, `
		SELECT COUNT(*)
		FROM messages m
		JOIN conversations c ON c.id = m.conversation_id
		WHERE c.company_id = $1
		AND DATE(m.created_at) = $2
	`, companyID, date.Format("2006-01-02")).Scan(&totalMessages)
	if err != nil {
		totalMessages = 0
	}

	// Count conversations created on this date
	err = s.db.Pool().QueryRow(ctx, `
		SELECT COUNT(*)
		FROM conversations
		WHERE company_id = $1
		AND DATE(created_at) = $2
	`, companyID, date.Format("2006-01-02")).Scan(&totalConversations)
	if err != nil {
		totalConversations = 0
	}

	// Count unique users who chatted on this date
	err = s.db.Pool().QueryRow(ctx, `
		SELECT COUNT(DISTINCT user_id)
		FROM conversations
		WHERE company_id = $1
		AND DATE(created_at) = $2
		AND user_id IS NOT NULL
	`, companyID, date.Format("2006-01-02")).Scan(&uniqueUsers)
	if err != nil {
		uniqueUsers = 0
	}

	// Insert or update daily log
	_, err = s.db.Pool().Exec(ctx, `
		INSERT INTO chat_usage_logs (id, company_id, date, total_messages, total_conversations, unique_users, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW())
		ON CONFLICT (company_id, date) DO UPDATE SET
			total_messages = $4,
			total_conversations = $5,
			unique_users = $6,
			updated_at = NOW()
	`, uuid.New().String(), companyID, date.Format("2006-01-02"), totalMessages, totalConversations, uniqueUsers)

	return err
}

// GetUsageStats retrieves usage statistics with filters
func (s *Service) GetUsageStats(ctx context.Context, companyID *string, startDate, endDate *time.Time) (*UsageStats, error) {
	query := `
		SELECT 
			COALESCE(SUM(total_messages), 0) as total_messages,
			COALESCE(SUM(total_conversations), 0) as total_conversations,
			COALESCE(MAX(unique_users), 0) as unique_users
		FROM chat_usage_logs
		WHERE 1=1
	`
	args := []interface{}{}
	argIndex := 1

	if companyID != nil {
		query += ` AND company_id = $` + fmt.Sprintf("%d", argIndex)
		args = append(args, *companyID)
		argIndex++
	}

	if startDate != nil {
		query += ` AND date >= $` + fmt.Sprintf("%d", argIndex)
		args = append(args, startDate.Format("2006-01-02"))
		argIndex++
	}

	if endDate != nil {
		query += ` AND date <= $` + fmt.Sprintf("%d", argIndex)
		args = append(args, endDate.Format("2006-01-02"))
		argIndex++
	}

	stats := &UsageStats{}
	err := s.db.Pool().QueryRow(ctx, query, args...).Scan(
		&stats.TotalMessages,
		&stats.TotalConversations,
		&stats.UniqueUsers,
	)
	if err != nil {
		return nil, err
	}

	if startDate != nil {
		stats.StartDate = *startDate
	}
	if endDate != nil {
		stats.EndDate = *endDate
	}

	// Determine period type
	if startDate != nil && endDate != nil {
		days := endDate.Sub(*startDate).Hours() / 24
		if days <= 1 {
			stats.Period = "daily"
		} else if days <= 31 {
			stats.Period = "monthly"
		} else {
			stats.Period = "custom"
		}
	}

	return stats, nil
}

// GetDailyLogs retrieves daily usage logs with filters
func (s *Service) GetDailyLogs(ctx context.Context, companyID *string, startDate, endDate *time.Time, page, limit int) ([]UsageLog, int, error) {
	offset := (page - 1) * limit

	query := `
		SELECT id, company_id, date, total_messages, total_conversations, unique_users, created_at, updated_at
		FROM chat_usage_logs
		WHERE 1=1
	`
	args := []interface{}{}
	argIndex := 1

	if companyID != nil {
		query += ` AND company_id = $` + fmt.Sprintf("%d", argIndex)
		args = append(args, *companyID)
		argIndex++
	}

	if startDate != nil {
		query += ` AND date >= $` + fmt.Sprintf("%d", argIndex)
		args = append(args, startDate.Format("2006-01-02"))
		argIndex++
	}

	if endDate != nil {
		query += ` AND date <= $` + fmt.Sprintf("%d", argIndex)
		args = append(args, endDate.Format("2006-01-02"))
		argIndex++
	}

	query += ` ORDER BY date DESC LIMIT $` + fmt.Sprintf("%d", argIndex) + ` OFFSET $` + fmt.Sprintf("%d", argIndex+1)
	args = append(args, limit, offset)

	rows, err := s.db.Pool().Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var logs []UsageLog
	for rows.Next() {
		var log UsageLog
		var dateStr string
		if err := rows.Scan(
			&log.ID, &log.CompanyID, &dateStr,
			&log.TotalMessages, &log.TotalConversations, &log.UniqueUsers,
			&log.CreatedAt, &log.UpdatedAt,
		); err != nil {
			continue
		}
		log.Date, _ = time.Parse("2006-01-02", dateStr)
		logs = append(logs, log)
	}

	// Get total count
	countQuery := `SELECT COUNT(*) FROM chat_usage_logs WHERE 1=1`
	countArgs := []interface{}{}
	countArgIndex := 1

	if companyID != nil {
		countQuery += ` AND company_id = $` + fmt.Sprintf("%d", countArgIndex)
		countArgs = append(countArgs, *companyID)
		countArgIndex++
	}

	if startDate != nil {
		countQuery += ` AND date >= $` + fmt.Sprintf("%d", countArgIndex)
		countArgs = append(countArgs, startDate.Format("2006-01-02"))
		countArgIndex++
	}

	if endDate != nil {
		countQuery += ` AND date <= $` + fmt.Sprintf("%d", countArgIndex)
		countArgs = append(countArgs, endDate.Format("2006-01-02"))
		countArgIndex++
	}

	var total int
	s.db.Pool().QueryRow(ctx, countQuery, countArgs...).Scan(&total)

	return logs, total, nil
}
