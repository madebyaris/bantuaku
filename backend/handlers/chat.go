package handlers

import (
	"net/http"
	"time"

	"github.com/bantuaku/backend/errors"
	"github.com/bantuaku/backend/logger"
	"github.com/bantuaku/backend/models"
	"github.com/bantuaku/backend/middleware"
	"github.com/bantuaku/backend/services/embedding"
	"github.com/bantuaku/backend/services/kolosal"
	"github.com/bantuaku/backend/services/tokenusage"
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
	Citations             []Citation             `json:"citations,omitempty"`
	RAGUsed               bool                   `json:"rag_used"`
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
	ctx := r.Context()
	userID := middleware.GetUserID(ctx)
	companyID := middleware.GetCompanyID(ctx)

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

	resp := StartConversationResponse{
		ConversationID: conversationID,
		Title:          title,
		CreatedAt:      time.Now(),
	}

	// Log activity (do not store message content)
	h.auditLogger.LogResourceAction(ctx, r, "activity.conversation.started", "conversation", conversationID, map[string]interface{}{
		"user_id":    userID,
		"company_id": companyID,
		"purpose":    req.Purpose,
	})

	h.respondJSON(w, http.StatusOK, resp)
}

// SendMessage handles sending a message in a conversation with RAG integration
func (h *Handler) SendMessage(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logger.With("request_id", ctx.Value("request_id"))
	userID := middleware.GetUserID(ctx)
	companyID := middleware.GetCompanyID(ctx)

	var req SendMessageRequest
	if err := h.parseJSON(r, &req); err != nil {
		h.respondError(w, err, r)
		return
	}

	if err := validation.Validate(&req); err != nil {
		h.respondError(w, err, r)
		return
	}

	messageID := uuid.New().String()
	var assistantReply string
	var structuredPayload map[string]interface{}
	var citations []Citation
	ragUsed := false

	// Initialize RAG service if embedding is configured
	var ragService *RAGService
	if h.config.KolosalAPIKey != "" {
		embedder, err := embedding.NewEmbedder(h.config)
		if err == nil {
			retrieval := embedding.NewRetrievalService(h.db.Pool(), embedder)
			ragService = NewRAGService(retrieval)
		}
	}

	// Perform RAG retrieval if service is available
	var ragContext string
	var retrievedChunks []embedding.RetrievedChunk
	if ragService != nil {
		context, chunks, err := ragService.BuildRAGContext(ctx, req.Message, 5) // Top 5 chunks
		if err == nil && context != "" {
			ragContext = context
			retrievedChunks = chunks
			ragUsed = true
			log.Info("RAG context retrieved", "chunks", len(chunks))
		}
	}

	// Use Kolosal.ai for chat completion
	if h.config.KolosalAPIKey != "" {
		client := kolosal.NewClient(h.config.KolosalAPIKey)

		var systemPrompt, userPrompt string
		if ragService != nil && ragContext != "" {
			// Build prompt with RAG context
			systemPrompt, userPrompt = ragService.BuildRAGPrompt(req.Message, ragContext)
		} else {
			// Fallback to basic prompt
			systemPrompt = "Kamu adalah Asisten Bantuaku, AI assistant untuk membantu UMKM Indonesia. Jawab dalam Bahasa Indonesia yang natural dan ramah."
			userPrompt = req.Message
		}

		resp, err := client.CreateChatCompletion(ctx, kolosal.ChatCompletionRequest{
			Model: "default",
			Messages: []kolosal.ChatCompletionMessage{
				{Role: "system", Content: systemPrompt},
				{Role: "user", Content: userPrompt},
			},
			MaxTokens:   2000, // Increased for RAG responses
			Temperature: 0.7,
		})

		if err == nil && len(resp.Choices) > 0 {
			assistantReply = resp.Choices[0].Message.Content
		} else {
			log.Warn("Chat completion failed", "error", err)
			assistantReply = "Terima kasih atas pesan Anda. Saya sedang memproses permintaan Anda."
		}

		// Token usage logging (best-effort, non-blocking)
		if err == nil && resp.Usage != nil {
			tokens := resp.Usage
			tokenSvc := tokenusage.NewService(h.db)
			if logErr := tokenSvc.CreateTokenUsage(
				ctx,
				&userID,
				companyID,
				&req.ConversationID,
				&messageID,
				resp.Model,
				"kolosal",
				tokens.PromptTokens,
				tokens.CompletionTokens,
				tokens.TotalTokens,
			); logErr != nil {
				log.Warn("Failed to log token usage", "error", logErr)
			}
		}

		// Extract citations from retrieved chunks
		if len(retrievedChunks) > 0 {
			citations = ExtractCitations(retrievedChunks)
		}
	} else {
		assistantReply = "Terima kasih atas pesan Anda. Fitur AI chat sedang dalam pengembangan. Silakan coba lagi nanti."
	}

	// Log retrieval diagnostics
	if ragUsed {
		log.Info("RAG retrieval completed",
			"query", req.Message,
			"chunks_retrieved", len(retrievedChunks),
			"citations", len(citations),
		)

		// Log RAG query activity (omit raw content)
		h.auditLogger.LogResourceAction(ctx, r, "activity.rag.query", "conversation", req.ConversationID, map[string]interface{}{
			"user_id":      userID,
			"company_id":   companyID,
			"message_len":  len(req.Message),
			"rag_used":     true,
			"chunks_count": len(retrievedChunks),
		})
	}

	resp := SendMessageResponse{
		MessageID:         messageID,
		AssistantReply:    assistantReply,
		StructuredPayload: structuredPayload,
		Citations:         citations,
		RAGUsed:           ragUsed,
	}

	// Log chat message activity (do not store content)
	h.auditLogger.LogResourceAction(ctx, r, "activity.chat.message", "conversation", req.ConversationID, map[string]interface{}{
		"user_id":         userID,
		"company_id":      companyID,
		"message_id":      messageID,
		"message_len":     len(req.Message),
		"file_upload_ids": req.FileUploadIDs,
		"rag_used":        ragUsed,
	})

	h.respondJSON(w, http.StatusOK, resp)
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
