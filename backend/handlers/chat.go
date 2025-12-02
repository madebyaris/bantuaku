package handlers

import (
	"net/http"
	"time"

	"github.com/bantuaku/backend/errors"
	"github.com/bantuaku/backend/models"
	"github.com/bantuaku/backend/services/kolosal"
	"github.com/bantuaku/backend/validation"

	"github.com/google/uuid"
)

// StartConversationRequest represents a request to start a new conversation
type StartConversationRequest struct {
	Purpose string `json:"purpose" validate:"required,oneof=onboarding forecasting market_research analysis"`
}

// StartConversationResponse represents the response when starting a conversation
type StartConversationResponse struct {
	ConversationID string    `json:"conversation_id"`
	Title          string    `json:"title"`
	CreatedAt      time.Time `json:"created_at"`
}

// SendMessageRequest represents a request to send a message
type SendMessageRequest struct {
	ConversationID string   `json:"conversation_id" validate:"required"`
	Message        string   `json:"message" validate:"required"`
	FileUploadIDs  []string `json:"file_upload_ids,omitempty"`
}

// SendMessageResponse represents the response when sending a message
type SendMessageResponse struct {
	MessageID             string                 `json:"message_id"`
	AssistantReply        string                 `json:"assistant_reply"`
	StructuredPayload     map[string]interface{} `json:"structured_payload,omitempty"`
	UpdatedProfileSummary map[string]interface{} `json:"updated_profile_summary,omitempty"`
}

// GetConversationsResponse represents a list of conversations
type GetConversationsResponse struct {
	Conversations []ConversationSummary `json:"conversations"`
}

// ConversationSummary represents a summary of a conversation
type ConversationSummary struct {
	ID            string    `json:"id"`
	Title         string    `json:"title"`
	Purpose       string    `json:"purpose"`
	CreatedAt     time.Time `json:"created_at"`
	LastMessageAt time.Time `json:"last_message_at"`
}

// GetMessagesResponse represents a list of messages
type GetMessagesResponse struct {
	Messages []models.Message `json:"messages"`
}

// StartConversation creates a new conversation
func (h *Handler) StartConversation(w http.ResponseWriter, r *http.Request) {
	_ = r.Context().Value("user_id")  // TODO: Use userID when implementing DB storage
	_ = r.Context().Value("store_id") // TODO: Update to company_id

	var req StartConversationRequest
	if err := h.parseJSON(r, &req); err != nil {
		h.respondError(w, err, r)
		return
	}

	if err := validation.Validate(&req); err != nil {
		h.respondError(w, err, r)
		return
	}

	// TODO: Implement conversation creation in database
	// For now, return a mock response
	conversationID := uuid.New().String()
	title := "New Conversation"
	if req.Purpose == "onboarding" {
		title = "Onboarding"
	}

	h.respondJSON(w, http.StatusOK, StartConversationResponse{
		ConversationID: conversationID,
		Title:          title,
		CreatedAt:      time.Now(),
	})
}

// SendMessage handles sending a message in a conversation
func (h *Handler) SendMessage(w http.ResponseWriter, r *http.Request) {
	_ = r.Context().Value("user_id")  // TODO: Use userID when implementing DB storage
	_ = r.Context().Value("store_id") // TODO: Update to company_id

	var req SendMessageRequest
	if err := h.parseJSON(r, &req); err != nil {
		h.respondError(w, err, r)
		return
	}

	if err := validation.Validate(&req); err != nil {
		h.respondError(w, err, r)
		return
	}

	// TODO: Implement message handling with AI assistant
	// For now, return a mock response or use Kolosal.ai if API key is available
	messageID := uuid.New().String()
	var assistantReply string
	var structuredPayload map[string]interface{}

	if h.config.KolosalAPIKey != "" {
		// Use Kolosal.ai for chat completion
		client := kolosal.NewClient(h.config.KolosalAPIKey)
		ctx := r.Context()

		systemPrompt := "Kamu adalah Asisten Bantuaku, AI assistant untuk membantu UMKM Indonesia. Jawab dalam Bahasa Indonesia yang natural dan ramah."
		userPrompt := req.Message

		resp, err := client.CreateChatCompletion(ctx, kolosal.ChatCompletionRequest{
			Model: "default",
			Messages: []kolosal.ChatCompletionMessage{
				{Role: "system", Content: systemPrompt},
				{Role: "user", Content: userPrompt},
			},
			MaxTokens:   1000,
			Temperature: 0.7,
		})

		if err == nil && len(resp.Choices) > 0 {
			assistantReply = resp.Choices[0].Message.Content
		} else {
			assistantReply = "Terima kasih atas pesan Anda. Saya sedang memproses permintaan Anda."
		}
	} else {
		assistantReply = "Terima kasih atas pesan Anda. Fitur AI chat sedang dalam pengembangan. Silakan coba lagi nanti."
	}

	h.respondJSON(w, http.StatusOK, SendMessageResponse{
		MessageID:         messageID,
		AssistantReply:    assistantReply,
		StructuredPayload: structuredPayload,
	})
}

// GetConversations retrieves all conversations for a company
func (h *Handler) GetConversations(w http.ResponseWriter, r *http.Request) {
	_ = r.Context().Value("user_id")  // TODO: Use userID when implementing DB storage
	_ = r.Context().Value("store_id") // TODO: Update to company_id

	// TODO: Implement conversation retrieval from database
	// For now, return empty list
	h.respondJSON(w, http.StatusOK, GetConversationsResponse{
		Conversations: []ConversationSummary{},
	})
}

// GetMessages retrieves messages for a conversation
func (h *Handler) GetMessages(w http.ResponseWriter, r *http.Request) {
	_ = r.Context().Value("user_id") // TODO: Use userID when implementing DB storage
	conversationID := r.URL.Query().Get("conversation_id")

	if conversationID == "" {
		h.respondError(w, errors.NewValidationError("conversation_id is required", ""), r)
		return
	}

	// TODO: Implement message retrieval from database
	// For now, return empty list
	h.respondJSON(w, http.StatusOK, GetMessagesResponse{
		Messages: []models.Message{},
	})
}
