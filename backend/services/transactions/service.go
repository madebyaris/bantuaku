package transactions

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/bantuaku/backend/services/storage"
	"github.com/google/uuid"
)

// EventType represents the type of subscription transaction event
type EventType string

const (
	EventTypeCreate       EventType = "create"
	EventTypeUpgrade      EventType = "upgrade"
	EventTypeDowngrade    EventType = "downgrade"
	EventTypeCancel       EventType = "cancel"
	EventTypeRenew        EventType = "renew"
	EventTypeStatusChange EventType = "status_change"
)

// Transaction represents a subscription transaction event
type Transaction struct {
	ID              string                 `json:"id"`
	SubscriptionID  string                 `json:"subscription_id"`
	CompanyID       string                 `json:"company_id"`
	EventType       string                 `json:"event_type"`
	OldPlanID       *string                `json:"old_plan_id,omitempty"`
	NewPlanID       *string                `json:"new_plan_id,omitempty"`
	OldStatus       *string                `json:"old_status,omitempty"`
	NewStatus       *string                `json:"new_status,omitempty"`
	ChangedByUserID *string                `json:"changed_by_user_id,omitempty"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt       time.Time              `json:"created_at"`
}

// Service handles subscription transaction logging
type Service struct {
	db *storage.Postgres
}

// NewService creates a new transactions service
func NewService(db *storage.Postgres) *Service {
	return &Service{db: db}
}

// LogTransaction logs a subscription transaction event
func (s *Service) LogTransaction(ctx context.Context, tx *Transaction) error {
	// Marshal metadata to JSON
	metadataJSON := "{}"
	if tx.Metadata != nil {
		if b, err := json.Marshal(tx.Metadata); err == nil {
			metadataJSON = string(b)
		}
	}

	_, err := s.db.Pool().Exec(ctx, `
		INSERT INTO subscription_transactions (
			id, subscription_id, company_id, event_type,
			old_plan_id, new_plan_id, old_status, new_status,
			changed_by_user_id, metadata, created_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`, tx.ID, tx.SubscriptionID, tx.CompanyID, tx.EventType,
		tx.OldPlanID, tx.NewPlanID, tx.OldStatus, tx.NewStatus,
		tx.ChangedByUserID, metadataJSON, tx.CreatedAt)

	return err
}

// GetTransactionHistory retrieves transaction history for a subscription
func (s *Service) GetTransactionHistory(ctx context.Context, subscriptionID string, page, limit int) ([]Transaction, int, error) {
	offset := (page - 1) * limit

	rows, err := s.db.Pool().Query(ctx, `
		SELECT 
			id, subscription_id, company_id, event_type,
			old_plan_id, new_plan_id, old_status, new_status,
			changed_by_user_id, metadata, created_at
		FROM subscription_transactions
		WHERE subscription_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`, subscriptionID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var transactions []Transaction
	for rows.Next() {
		var tx Transaction
		var oldPlanID, newPlanID, oldStatus, newStatus, changedByUserID sql.NullString
		var metadataJSON sql.NullString

		if err := rows.Scan(
			&tx.ID, &tx.SubscriptionID, &tx.CompanyID, &tx.EventType,
			&oldPlanID, &newPlanID, &oldStatus, &newStatus,
			&changedByUserID, &metadataJSON, &tx.CreatedAt,
		); err != nil {
			continue
		}

		if oldPlanID.Valid {
			tx.OldPlanID = &oldPlanID.String
		}
		if newPlanID.Valid {
			tx.NewPlanID = &newPlanID.String
		}
		if oldStatus.Valid {
			tx.OldStatus = &oldStatus.String
		}
		if newStatus.Valid {
			tx.NewStatus = &newStatus.String
		}
		if changedByUserID.Valid {
			tx.ChangedByUserID = &changedByUserID.String
		}
		if metadataJSON.Valid && metadataJSON.String != "" {
			json.Unmarshal([]byte(metadataJSON.String), &tx.Metadata)
		}

		transactions = append(transactions, tx)
	}

	// Get total count
	var total int
	err = s.db.Pool().QueryRow(ctx, `
		SELECT COUNT(*) FROM subscription_transactions WHERE subscription_id = $1
	`, subscriptionID).Scan(&total)
	if err != nil {
		total = len(transactions)
	}

	return transactions, total, nil
}

// CreateTransaction creates a new transaction with auto-generated ID
func (s *Service) CreateTransaction(ctx context.Context, subscriptionID, companyID string, eventType EventType, oldPlanID, newPlanID, oldStatus, newStatus *string, changedByUserID *string, metadata map[string]interface{}) error {
	tx := &Transaction{
		ID:              uuid.New().String(),
		SubscriptionID:  subscriptionID,
		CompanyID:       companyID,
		EventType:       string(eventType),
		OldPlanID:       oldPlanID,
		NewPlanID:       newPlanID,
		OldStatus:       oldStatus,
		NewStatus:       newStatus,
		ChangedByUserID: changedByUserID,
		Metadata:        metadata,
		CreatedAt:       time.Now(),
	}
	return s.LogTransaction(ctx, tx)
}
