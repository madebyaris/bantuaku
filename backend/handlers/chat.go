package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/bantuaku/backend/errors"
	"github.com/bantuaku/backend/logger"
	"github.com/bantuaku/backend/middleware"
	"github.com/bantuaku/backend/models"
	"github.com/bantuaku/backend/services/embedding"
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
	log := logger.With("request_id", r.Context().Value("request_id"))
	userID := middleware.GetUserID(r.Context())
	companyID := middleware.GetCompanyID(r.Context())

	if userID == "" || companyID == "" {
		h.respondError(w, errors.NewValidationError("user_id and company_id are required", ""), r)
		return
	}

	var req StartConversationRequest
	if err := h.parseJSON(r, &req); err != nil {
		h.respondError(w, err, r)
		return
	}

	if err := validation.Validate(&req); err != nil {
		h.respondError(w, err, r)
		return
	}

	ctx := r.Context()
	conversationID := uuid.New().String()
	title := "New Conversation"
	if req.Purpose == "onboarding" {
		title = "Onboarding"
	} else if req.Purpose == "forecasting" {
		title = "Forecasting"
	} else if req.Purpose == "market_research" {
		title = "Market Research"
	} else if req.Purpose == "analysis" {
		title = "Analysis"
	}

	now := time.Now()
	_, err := h.db.Pool().Exec(ctx, `
		INSERT INTO conversations (id, company_id, user_id, title, purpose, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, conversationID, companyID, userID, title, req.Purpose, now, now)

	if err != nil {
		log.Error("Failed to create conversation", "error", err)
		h.respondError(w, errors.NewAppError(errors.ErrCodeInternal, "Failed to create conversation", err.Error()), r)
		return
	}

	h.respondJSON(w, http.StatusOK, StartConversationResponse{
		ConversationID: conversationID,
		Title:          title,
		CreatedAt:      now,
	})
}

// SendMessage handles sending a message in a conversation with RAG integration
func (h *Handler) SendMessage(w http.ResponseWriter, r *http.Request) {
	log := logger.With("request_id", r.Context().Value("request_id"))
	userID := middleware.GetUserID(r.Context())
	companyID := middleware.GetCompanyID(r.Context())
	_ = userID    // TODO: Use userID when implementing DB storage
	_ = companyID // TODO: Use companyID when implementing DB storage

	var req SendMessageRequest
	if err := h.parseJSON(r, &req); err != nil {
		h.respondError(w, err, r)
		return
	}

	if err := validation.Validate(&req); err != nil {
		h.respondError(w, err, r)
		return
	}

	var assistantReply string
	var structuredPayload map[string]interface{}
	var citations []Citation
	ragUsed := false

	ctx := r.Context()

	// Save user message to database
	userMessageID := uuid.New().String()
	_, err := h.db.Pool().Exec(ctx, `
		INSERT INTO messages (id, conversation_id, sender, content, created_at)
		VALUES ($1, $2, $3, $4, $5)
	`, userMessageID, req.ConversationID, "user", req.Message, time.Now())

	if err != nil {
		log.Error("Failed to save user message", "error", err)
		h.respondError(w, errors.NewAppError(errors.ErrCodeInternal, "Failed to save message", err.Error()), r)
		return
	}

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
			log.Info("Using RAG-enhanced prompt", "rag_context_length", len(ragContext))
		} else {
			// Fallback to basic prompt
			systemPrompt = "Kamu adalah Asisten Bantuaku, AI assistant untuk membantu UMKM Indonesia. Jawab dalam Bahasa Indonesia yang natural dan ramah."
			userPrompt = req.Message
			log.Debug("Using basic prompt (no RAG)")
		}

		log.Info("Calling Kolosal.ai chat completion",
			"model", "default",
			"message_length", len(req.Message),
			"conversation_id", req.ConversationID,
			"user_message", req.Message)

		resp, err := client.CreateChatCompletion(ctx, kolosal.ChatCompletionRequest{
			Model: "default",
			Messages: []kolosal.ChatCompletionMessage{
				{Role: "system", Content: systemPrompt},
				{Role: "user", Content: userPrompt},
			},
			MaxTokens:   2000, // Increased for RAG responses
			Temperature: 0.7,
		})

		if err != nil {
			log.Error("Kolosal.ai chat completion failed",
				"error", err,
				"conversation_id", req.ConversationID,
				"error_details", err.Error(),
				"user_message", req.Message)
			// Return error to user instead of generic template
			assistantReply = fmt.Sprintf("Maaf, terjadi kesalahan saat memproses pesan Anda. Error: %v. Silakan coba lagi atau hubungi support.", err)
		} else if resp == nil {
			log.Error("Kolosal.ai returned nil response",
				"conversation_id", req.ConversationID,
				"user_message", req.Message)
			assistantReply = "Maaf, terjadi kesalahan saat memproses pesan Anda. Response kosong dari server. Silakan coba lagi."
		} else if len(resp.Choices) == 0 {
			log.Error("Kolosal.ai returned empty choices",
				"conversation_id", req.ConversationID,
				"response_id", resp.ID,
				"response_model", resp.Model,
				"user_message", req.Message)
			assistantReply = "Maaf, terjadi kesalahan saat memproses pesan Anda. Tidak ada response dari AI. Silakan coba lagi."
		} else {
			assistantReply = resp.Choices[0].Message.Content
			if assistantReply == "" {
				log.Error("Kolosal.ai returned empty content",
					"conversation_id", req.ConversationID,
					"choice_index", resp.Choices[0].Index,
					"finish_reason", resp.Choices[0].FinishReason,
					"user_message", req.Message)
				assistantReply = "Maaf, terjadi kesalahan saat memproses pesan Anda. Response kosong. Silakan coba lagi."
			} else {
				log.Info("Kolosal.ai chat completion successful",
					"conversation_id", req.ConversationID,
					"response_length", len(assistantReply),
					"response_preview", func() string {
						if len(assistantReply) > 100 {
							return assistantReply[:100] + "..."
						}
						return assistantReply
					}(),
					"rag_used", ragUsed,
					"model", resp.Model)
			}
		}

		// Extract citations from retrieved chunks
		if len(retrievedChunks) > 0 {
			citations = ExtractCitations(retrievedChunks)
		}
	} else {
		log.Warn("Kolosal API key not configured")
		assistantReply = "Terima kasih atas pesan Anda. Fitur AI chat sedang dalam pengembangan. Silakan coba lagi nanti."
	}

	// Save assistant reply to database
	assistantMessageID := uuid.New().String()
	var structuredPayloadJSON []byte
	// structuredPayload is reserved for future use (extracted fields, tool calls)
	// For now, it's always nil, so we don't marshal it

	_, err = h.db.Pool().Exec(ctx, `
		INSERT INTO messages (id, conversation_id, sender, content, structured_payload, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, assistantMessageID, req.ConversationID, "assistant", assistantReply, structuredPayloadJSON, time.Now())

	if err != nil {
		log.Error("Failed to save assistant message", "error", err)
		// Continue anyway - message was sent, just logging failed
	}

	// Update conversation updated_at timestamp
	_, err = h.db.Pool().Exec(ctx, `
		UPDATE conversations
		SET updated_at = $1
		WHERE id = $2
	`, time.Now(), req.ConversationID)

	if err != nil {
		log.Warn("Failed to update conversation timestamp", "error", err)
		// Continue anyway - not critical
	}

	// Log retrieval diagnostics
	if ragUsed {
		log.Info("RAG retrieval completed",
			"query", req.Message,
			"chunks_retrieved", len(retrievedChunks),
			"citations", len(citations),
		)
	}

	h.respondJSON(w, http.StatusOK, SendMessageResponse{
		MessageID:         assistantMessageID,
		AssistantReply:    assistantReply,
		StructuredPayload: structuredPayload,
		Citations:         citations,
		RAGUsed:           ragUsed,
	})
}

// GetConversations retrieves all conversations for a company
func (h *Handler) GetConversations(w http.ResponseWriter, r *http.Request) {
	log := logger.With("request_id", r.Context().Value("request_id"))
	companyID := middleware.GetCompanyID(r.Context())

	if companyID == "" {
		h.respondError(w, errors.NewValidationError("company_id is required", ""), r)
		return
	}

	// Parse pagination parameters
	limit := 5 // Default limit
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	offset := 0
	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		if parsedOffset, err := strconv.Atoi(offsetStr); err == nil && parsedOffset >= 0 {
			offset = parsedOffset
		}
	}

	ctx := r.Context()
	rows, err := h.db.Pool().Query(ctx, `
		SELECT id, title, purpose, created_at, updated_at
		FROM conversations
		WHERE company_id = $1
		ORDER BY updated_at DESC
		LIMIT $2 OFFSET $3
	`, companyID, limit, offset)

	if err != nil {
		log.Error("Failed to fetch conversations", "error", err)
		h.respondError(w, errors.NewAppError(errors.ErrCodeInternal, "Failed to fetch conversations", err.Error()), r)
		return
	}
	defer rows.Close()

	conversations := []ConversationSummary{}
	for rows.Next() {
		var conv ConversationSummary
		if err := rows.Scan(&conv.ID, &conv.Title, &conv.Purpose, &conv.CreatedAt, &conv.LastMessageAt); err != nil {
			log.Warn("Failed to scan conversation row", "error", err)
			continue
		}
		conversations = append(conversations, conv)
	}

	h.respondJSON(w, http.StatusOK, GetConversationsResponse{
		Conversations: conversations,
	})
}

// GetMessages retrieves messages for a conversation
func (h *Handler) GetMessages(w http.ResponseWriter, r *http.Request) {
	log := logger.With("request_id", r.Context().Value("request_id"))
	conversationID := r.URL.Query().Get("conversation_id")

	if conversationID == "" {
		h.respondError(w, errors.NewValidationError("conversation_id is required", ""), r)
		return
	}

	// Parse pagination parameters
	limit := 50 // Default limit
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	offset := 0
	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		if parsedOffset, err := strconv.Atoi(offsetStr); err == nil && parsedOffset >= 0 {
			offset = parsedOffset
		}
	}

	ctx := r.Context()
	rows, err := h.db.Pool().Query(ctx, `
		SELECT id, conversation_id, sender, content, structured_payload, file_upload_id, created_at
		FROM messages
		WHERE conversation_id = $1
		ORDER BY created_at ASC
		LIMIT $2 OFFSET $3
	`, conversationID, limit, offset)

	if err != nil {
		log.Error("Failed to fetch messages", "error", err)
		h.respondError(w, errors.NewAppError(errors.ErrCodeInternal, "Failed to fetch messages", err.Error()), r)
		return
	}
	defer rows.Close()

	messages := []models.Message{}
	for rows.Next() {
		var msg models.Message
		var structuredPayloadJSON []byte
		var fileUploadID *string

		if err := rows.Scan(&msg.ID, &msg.ConversationID, &msg.Sender, &msg.Content, &structuredPayloadJSON, &fileUploadID, &msg.CreatedAt); err != nil {
			log.Warn("Failed to scan message row", "error", err)
			continue
		}

		// Parse structured_payload JSONB
		if len(structuredPayloadJSON) > 0 {
			if err := json.Unmarshal(structuredPayloadJSON, &msg.StructuredPayload); err != nil {
				log.Warn("Failed to unmarshal structured_payload", "error", err)
			}
		}

		if fileUploadID != nil {
			msg.FileUploadID = fileUploadID
		}

		messages = append(messages, msg)
	}

	h.respondJSON(w, http.StatusOK, GetMessagesResponse{
		Messages: messages,
	})
}
