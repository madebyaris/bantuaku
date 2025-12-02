package models

import (
	"time"
)

// Conversation represents a chat thread
type Conversation struct {
	ID        string    `json:"id"`
	CompanyID string    `json:"company_id"`
	UserID    string    `json:"user_id"`
	Title     string    `json:"title,omitempty"`
	Purpose   string    `json:"purpose"` // "onboarding", "forecasting", "market_research", "analysis"
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Message represents a single message in a conversation
type Message struct {
	ID                string                 `json:"id"`
	ConversationID    string                 `json:"conversation_id"`
	Sender            string                 `json:"sender"` // "user", "assistant", "system"
	Content           string                 `json:"content"`
	StructuredPayload map[string]interface{} `json:"structured_payload,omitempty"` // JSONB - extracted fields, tool calls
	FileUploadID      *string                `json:"file_upload_id,omitempty"`
	CreatedAt         time.Time              `json:"created_at"`
}
