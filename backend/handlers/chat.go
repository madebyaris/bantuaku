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
	"github.com/bantuaku/backend/services/chat"
	"github.com/bantuaku/backend/services/embedding"
	"github.com/bantuaku/backend/services/settings"
	"github.com/bantuaku/backend/services/tokenusage"
	"github.com/bantuaku/backend/services/tools"
	"github.com/bantuaku/backend/services/usage"
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

	// Check if database connection is available
	if h.db == nil {
		log.Error("Database connection not available")
		h.respondError(w, errors.NewAppError(errors.ErrCodeInternal, "Database connection not available", ""), r)
		return
	}

	pool := h.db.Pool()
	if pool == nil {
		log.Error("Database pool not available")
		h.respondError(w, errors.NewAppError(errors.ErrCodeInternal, "Database pool not available", ""), r)
		return
	}

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
	_, err := pool.Exec(ctx, `
		INSERT INTO conversations (id, company_id, user_id, title, purpose, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, conversationID, companyID, userID, title, req.Purpose, now, now)

	if err != nil {
		log.Error("Failed to create conversation", "error", err, "company_id", companyID, "user_id", userID)
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
	_ = userID // TODO: Use userID when implementing DB storage

	ctx := r.Context()

	// Check chat usage limit
	usageService := usage.NewService(h.db)
	canChat, limitMsg, err := usageService.CheckChatLimit(ctx, companyID)
	if err != nil {
		log.Warn("Failed to check chat limit", "error", err)
		// Continue on error - don't block user
	} else if !canChat {
		h.respondError(w, errors.NewAppError(errors.ErrCodeForbidden, limitMsg, "chat_limit_exceeded"), r)
		return
	}

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
	var model, provider string            // Track model and provider for token logging
	var resp *chat.ChatCompletionResponse // Track response for token logging

	// Save user message to database
	userMessageID := uuid.New().String()
	_, dbErr := h.db.Pool().Exec(ctx, `
		INSERT INTO messages (id, conversation_id, sender, content, created_at)
		VALUES ($1, $2, $3, $4, $5)
	`, userMessageID, req.ConversationID, "user", req.Message, time.Now())

	if dbErr != nil {
		log.Error("Failed to save user message", "error", dbErr)
		h.respondError(w, errors.NewAppError(errors.ErrCodeInternal, "Failed to save message", dbErr.Error()), r)
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

	// Use chat provider factory to get the configured provider (OpenRouter or Kolosal)
	settingsService := settings.NewService(h.db)
	chatProvider, err := chat.NewChatProvider(ctx, h.config, settingsService)
	if err != nil {
		log.Error("Failed to initialize chat provider",
			"error", err,
			"conversation_id", req.ConversationID)
		assistantReply = fmt.Sprintf("Maaf, terjadi kesalahan saat memproses pesan Anda. Error: %v. Silakan coba lagi atau hubungi support.", err)
	} else {
		var systemPrompt, userPrompt string
		if ragService != nil && ragContext != "" {
			// Build prompt with RAG context
			systemPrompt, userPrompt = ragService.BuildRAGPrompt(req.Message, ragContext)
			log.Info("Using RAG-enhanced prompt", "rag_context_length", len(ragContext))
		} else {
			// Fallback to basic prompt with tool instructions
			systemPrompt = buildSystemPromptWithTools()
			userPrompt = req.Message
			log.Debug("Using basic prompt (no RAG)")
		}

		// Determine model and provider based on settings
		// Get provider from settings to determine which model to use
		model = h.config.OpenRouterModelChat // Default for OpenRouter (from config)
		provider = "openrouter"              // Default provider
		if model == "" {
			model = "openai/gpt-4o-mini" // Fallback default
		}

		if providerSetting, _ := settingsService.GetSetting(ctx, "ai_provider"); providerSetting != "" {
			var settingData map[string]interface{}
			if json.Unmarshal([]byte(providerSetting), &settingData) == nil {
				if p, ok := settingData["provider"].(string); ok && p == "kolosal" {
					model = "GLM 4.6" // Kolosal model
					provider = "kolosal"
				}
			}
		} else {
			// If no setting, check API keys to infer provider
			if h.config.KolosalAPIKey != "" && h.config.OpenRouterAPIKey == "" {
				model = "GLM 4.6" // Only Kolosal key set
				provider = "kolosal"
			}
		}

		// Get tool definitions
		modelTools := tools.GetToolDefinitions()
		log.Info("Retrieved tool definitions",
			"model_tools_count", len(modelTools),
			"conversation_id", req.ConversationID)

		// Convert models.Tool to chat.Tool
		chatTools := make([]chat.Tool, len(modelTools))
		for i, mt := range modelTools {
			chatTools[i] = chat.Tool{
				Type: mt.Type,
				Function: chat.Function{
					Name:        mt.Function.Name,
					Description: mt.Function.Description,
					Parameters:  mt.Function.Parameters,
				},
			}
		}
		log.Info("Converted tools to chat format",
			"chat_tools_count", len(chatTools),
			"conversation_id", req.ConversationID)

		// Build conversation messages (load history if needed)
		messages := []chat.ChatCompletionMessage{
			{Role: "system", Content: systemPrompt},
		}

		// Load conversation history (last 20 messages to keep context manageable)
		historyRows, err := h.db.Pool().Query(ctx, `
			SELECT sender, content, structured_payload, created_at
			FROM messages
			WHERE conversation_id = $1
			ORDER BY created_at ASC
			LIMIT 20
		`, req.ConversationID)
		if err == nil {
			defer historyRows.Close()
			for historyRows.Next() {
				var sender, content string
				var structuredPayloadJSON []byte
				var createdAt time.Time
				if err := historyRows.Scan(&sender, &content, &structuredPayloadJSON, &createdAt); err == nil {
					// Skip the current user message (we'll add it at the end)
					// Convert sender to role format
					role := sender
					if sender == "assistant" {
						role = "assistant"
					} else if sender == "user" {
						role = "user"
					} else {
						role = "user" // Default
					}

					// Parse structured_payload for tool calls if present
					var toolCalls []chat.ToolCall
					if len(structuredPayloadJSON) > 0 {
						var payload map[string]interface{}
						if json.Unmarshal(structuredPayloadJSON, &payload) == nil {
							if toolCallsRaw, ok := payload["tool_calls"].([]interface{}); ok {
								toolCalls = make([]chat.ToolCall, 0, len(toolCallsRaw))
								for _, tc := range toolCallsRaw {
									if tcMap, ok := tc.(map[string]interface{}); ok {
										if funcMap, ok := tcMap["function"].(map[string]interface{}); ok {
											toolCalls = append(toolCalls, chat.ToolCall{
												ID:   getString(tcMap, "id"),
												Type: getString(tcMap, "type"),
												Function: chat.FuncCall{
													Name:      getString(funcMap, "name"),
													Arguments: getString(funcMap, "arguments"),
												},
											})
										}
									}
								}
							}
						}
					}

					msg := chat.ChatCompletionMessage{
						Role:      role,
						Content:   content,
						ToolCalls: toolCalls,
					}
					messages = append(messages, msg)
				}
			}
			log.Info("Loaded conversation history",
				"conversation_id", req.ConversationID,
				"history_messages", len(messages)-1) // -1 for system message
		} else {
			log.Warn("Failed to load conversation history", "error", err)
		}

		// Add current user message
		messages = append(messages, chat.ChatCompletionMessage{
			Role:    "user",
			Content: userPrompt,
		})

		// Debug: Log tool names
		if len(chatTools) > 0 {
			toolNames := make([]string, len(chatTools))
			for i, t := range chatTools {
				toolNames[i] = t.Function.Name
			}
			log.Info("Tools being sent to AI",
				"tool_names", toolNames,
				"tools_count", len(chatTools),
				"conversation_id", req.ConversationID)
		} else {
			log.Warn("No tools configured - tools array is empty",
				"conversation_id", req.ConversationID)
		}

		log.Info("Calling chat completion",
			"model", model,
			"message_length", len(req.Message),
			"conversation_id", req.ConversationID,
			"user_message", req.Message,
			"tools_count", len(chatTools),
			"total_messages", len(messages))

		// Initial chat completion request with tools
		chatReq := chat.ChatCompletionRequest{
			Model:       model,
			Messages:    messages,
			Tools:       chatTools,
			ToolChoice:  "auto", // Let model decide when to use tools
			MaxTokens:   2000,   // Increased for RAG responses
			Temperature: 0.7,
		}

		// Debug: Log the actual request being sent
		if len(chatTools) > 0 {
			log.Info("Chat request includes tools",
				"tools_count", len(chatReq.Tools),
				"tool_choice", chatReq.ToolChoice,
				"conversation_id", req.ConversationID)
		}

		// Tool execution loop - continue until no more tool calls
		maxToolIterations := 5 // Prevent infinite loops
		toolExecutor := tools.NewExecutor(h.db)

		for iteration := 0; iteration < maxToolIterations; iteration++ {
			resp, err = chatProvider.CreateChatCompletion(ctx, chatReq)

			if err != nil {
				log.Error("Chat completion failed",
					"error", err,
					"conversation_id", req.ConversationID,
					"error_details", err.Error(),
					"user_message", req.Message,
					"iteration", iteration)
				assistantReply = fmt.Sprintf("Maaf, terjadi kesalahan saat memproses pesan Anda. Error: %v. Silakan coba lagi atau hubungi support.", err)
				break
			}

			if resp == nil || len(resp.Choices) == 0 {
				log.Error("Chat provider returned invalid response",
					"conversation_id", req.ConversationID,
					"iteration", iteration)
				assistantReply = "Maaf, terjadi kesalahan saat memproses pesan Anda. Silakan coba lagi."
				break
			}

			choice := resp.Choices[0]
			message := choice.Message

			// Check if there are tool calls
			if len(message.ToolCalls) > 0 {
				log.Info("Tool calls detected",
					"conversation_id", req.ConversationID,
					"tool_calls_count", len(message.ToolCalls),
					"iteration", iteration)

				// Execute all tool calls
				toolResults := []*models.ToolResult{}
				for _, toolCall := range message.ToolCalls {
					// Convert chat.ToolCall to models.ToolCall
					modelToolCall := models.ToolCall{
						ID:   toolCall.ID,
						Type: toolCall.Type,
						Function: models.FuncCall{
							Name:      toolCall.Function.Name,
							Arguments: toolCall.Function.Arguments,
						},
					}
					result, execErr := toolExecutor.ExecuteTool(ctx, companyID, modelToolCall)
					if execErr != nil {
						log.Error("Tool execution error",
							"tool", toolCall.Function.Name,
							"error", execErr,
							"conversation_id", req.ConversationID)
						result = &models.ToolResult{
							ToolCallID: toolCall.ID,
							Name:       toolCall.Function.Name,
							Error:      fmt.Sprintf("Execution error: %v", execErr),
						}
					}
					toolResults = append(toolResults, result)
				}

				// Format tool results as tool messages
				toolMessages := tools.FormatToolResults(toolResults)

				// Add assistant message with tool calls and tool results to conversation
				messages = append(messages, message)         // Assistant message with tool_calls
				messages = append(messages, toolMessages...) // Tool result messages

				// Update chatReq.Messages for the next iteration
				chatReq.Messages = messages

				// Continue loop - send tool results back to AI
				log.Info("Sending tool results back to AI",
					"conversation_id", req.ConversationID,
					"results_count", len(toolResults),
					"iteration", iteration,
					"total_messages", len(messages))
				continue
			}

			// No tool calls - get final response
			assistantReply = message.Content
			if assistantReply == "" && choice.FinishReason != "tool_calls" {
				log.Error("Chat provider returned empty content",
					"conversation_id", req.ConversationID,
					"choice_index", choice.Index,
					"finish_reason", choice.FinishReason,
					"user_message", req.Message)
				assistantReply = "Maaf, terjadi kesalahan saat memproses pesan Anda. Response kosong. Silakan coba lagi."
			}

			// Success - break out of loop
			log.Info("Chat completion successful",
				"conversation_id", req.ConversationID,
				"response_length", len(assistantReply),
				"rag_used", ragUsed,
				"model", model,
				"iterations", iteration+1)
			break
		}

		// Fallback if we hit max iterations
		if assistantReply == "" {
			assistantReply = "Maaf, terjadi kesalahan saat memproses pesan Anda. Terlalu banyak iterasi tool. Silakan coba lagi."
			log.Error("Max tool iterations reached",
				"conversation_id", req.ConversationID,
				"max_iterations", maxToolIterations)
		}

		// Extract citations from retrieved chunks
		if len(retrievedChunks) > 0 {
			citations = ExtractCitations(retrievedChunks)
		}
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

	// Log token usage if available
	if resp != nil && resp.Usage != nil {
		tokenService := tokenusage.NewService(h.db)
		promptTokens := resp.Usage.PromptTokens
		completionTokens := resp.Usage.CompletionTokens
		totalTokens := resp.Usage.TotalTokens
		if totalTokens == 0 {
			totalTokens = promptTokens + completionTokens
		}
		if err := tokenService.CreateTokenUsage(ctx, companyID, &req.ConversationID, &assistantMessageID, model, provider, promptTokens, completionTokens, totalTokens); err != nil {
			log.Warn("Failed to log token usage", "error", err)
			// Don't fail the request if token logging fails
		}
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

	// Check if database connection is available
	if h.db == nil {
		log.Error("Database connection not available")
		h.respondError(w, errors.NewAppError(errors.ErrCodeInternal, "Database connection not available", ""), r)
		return
	}

	pool := h.db.Pool()
	if pool == nil {
		log.Error("Database pool not available")
		h.respondError(w, errors.NewAppError(errors.ErrCodeInternal, "Database pool not available", ""), r)
		return
	}

	rows, err := pool.Query(ctx, `
		SELECT 
			c.id, 
			c.title, 
			c.purpose, 
			c.created_at, 
			COALESCE(MAX(m.created_at), c.updated_at) as last_message_at
		FROM conversations c
		LEFT JOIN messages m ON m.conversation_id = c.id
		WHERE c.company_id = $1
		GROUP BY c.id, c.title, c.purpose, c.created_at, c.updated_at
		ORDER BY last_message_at DESC
		LIMIT $2 OFFSET $3
	`, companyID, limit, offset)

	if err != nil {
		log.Error("Failed to fetch conversations", "error", err, "company_id", companyID)
		h.respondError(w, errors.NewAppError(errors.ErrCodeInternal, "Failed to fetch conversations", err.Error()), r)
		return
	}
	defer rows.Close()

	conversations := []ConversationSummary{}
	for rows.Next() {
		var conv ConversationSummary
		if err := rows.Scan(&conv.ID, &conv.Title, &conv.Purpose, &conv.CreatedAt, &conv.LastMessageAt); err != nil {
			log.Error("Failed to scan conversation row", "error", err)
			continue
		}
		conversations = append(conversations, conv)
	}

	if err := rows.Err(); err != nil {
		log.Error("Error iterating conversation rows", "error", err)
		h.respondError(w, errors.NewAppError(errors.ErrCodeInternal, "Failed to process conversations", err.Error()), r)
		return
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

// Helper function to safely get string from interface{}
func getString(m map[string]interface{}, key string) string {
	if val, ok := m[key]; ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return ""
}

// buildSystemPromptWithTools creates an enhanced system prompt with tool usage instructions
func buildSystemPromptWithTools() string {
	return `Kamu adalah Asisten Bantuaku, AI assistant untuk membantu UMKM Indonesia mengumpulkan dan mengelola data bisnis mereka.

PANDUAN PENGUMPULAN DATA:
1. Sebelum menjawab pertanyaan, periksa profil perusahaan menggunakan tool check_company_profile
2. Jika ada field yang kosong (industry, city, location_region, business_model, social_media), tanyakan kepada user dengan ramah
3. Gunakan tool update_company_info untuk menyimpan informasi perusahaan (industry, city, location, business_model, description)
4. Gunakan tool update_company_social_media untuk menyimpan akun social media (instagram, tiktok, tokopedia, shopee, dll)
5. Gunakan tool create_product untuk menambahkan produk/layanan yang disebutkan user
6. Gunakan tool list_products untuk melihat produk yang sudah ada

PANDUAN INTERAKSI:
- JANGAN gunakan sapaan "Halo" atau "Hai" berulang kali dalam percakapan yang sama
- Sapaan hanya digunakan di awal percakapan baru, setelah itu langsung ke inti pembahasan
- Tanyakan dengan natural dan ramah, seperti sedang berbicara dengan teman
- Gunakan bahasa yang singkat dan to-the-point, tidak bertele-tele
- Sebelum menyimpan data penting, konfirmasi dengan user
- Hanya panggil tool setelah user mengkonfirmasi atau memberikan informasi yang jelas
- Jika user tidak yakin, berikan contoh atau pilihan
- Selalu jawab dalam Bahasa Indonesia yang natural

CONTOH ALUR:
User: "Saya punya bisnis kuliner"
AI: [Cek profil] → "Di kota mana bisnis kuliner Anda beroperasi?"
User: "Jakarta"
AI: "Baik, di Jakarta ya. Saya simpan informasinya." [Panggil update_company_info]
AI: "Sudah tersimpan! Ada produk unggulan yang mau ditambahkan?"

Jawab dalam Bahasa Indonesia yang natural, singkat, dan to-the-point.`
}
