package chat

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/bantuaku/backend/config"
	"github.com/bantuaku/backend/services/kolosal"
	"github.com/bantuaku/backend/services/openrouter"
)

// Provider represents chat provider type
type Provider string

const (
	ProviderOpenRouter Provider = "openrouter"
	ProviderKolosal    Provider = "kolosal"
)

// SettingsReader interface for reading settings (to avoid circular dependency)
type SettingsReader interface {
	GetSetting(ctx context.Context, key string) (string, error)
}

// NewChatProvider creates a new chat provider based on configuration and settings
func NewChatProvider(ctx context.Context, cfg *config.Config, settingsService SettingsReader) (ChatProvider, error) {
	// Try to get provider from settings first
	var providerStr string
	if settingsService != nil {
		if setting, err := settingsService.GetSetting(ctx, "ai_provider"); err == nil && setting != "" {
			// Parse JSON value: {"provider": "openrouter"}
			var settingData map[string]interface{}
			if err := json.Unmarshal([]byte(setting), &settingData); err == nil {
				if p, ok := settingData["provider"].(string); ok && p != "" {
					providerStr = p
				}
			}
		}
	}

	// Fallback to default if not set
	if providerStr == "" {
		providerStr = "openrouter" // Default to OpenRouter
	}

	provider := Provider(providerStr)

	switch provider {
	case ProviderOpenRouter, "": // Default to OpenRouter if empty
		if cfg.OpenRouterAPIKey == "" {
			return nil, fmt.Errorf("OpenRouter API key not configured (set OPENROUTER_API_KEY)")
		}
		return &OpenRouterAdapter{client: openrouter.NewClient(cfg.OpenRouterAPIKey)}, nil

	case ProviderKolosal:
		if cfg.KolosalAPIKey == "" {
			return nil, fmt.Errorf("Kolosal API key not configured (set KOLOSAL_API_KEY)")
		}
		return &KolosalAdapter{client: kolosal.NewClient(cfg.KolosalAPIKey)}, nil

	default:
		return nil, fmt.Errorf("unknown chat provider: %s", provider)
	}
}

// Adapter types to make Kolosal and OpenRouter clients implement ChatProvider interface

// KolosalAdapter adapts kolosal.Client to ChatProvider
type KolosalAdapter struct {
	client *kolosal.Client
}

func (a *KolosalAdapter) CreateChatCompletion(ctx context.Context, req ChatCompletionRequest) (*ChatCompletionResponse, error) {
	kolosalReq := kolosal.ChatCompletionRequest{
		Model:       req.Model,
		Messages:    convertMessagesToKolosal(req.Messages),
		MaxTokens:   req.MaxTokens,
		Temperature: req.Temperature,
	}
	resp, err := a.client.CreateChatCompletion(ctx, kolosalReq)
	if err != nil {
		return nil, err
	}
	return convertKolosalResponse(resp), nil
}

// OpenRouterAdapter adapts openrouter.Client to ChatProvider
type OpenRouterAdapter struct {
	client *openrouter.Client
}

func (a *OpenRouterAdapter) CreateChatCompletion(ctx context.Context, req ChatCompletionRequest) (*ChatCompletionResponse, error) {
	openrouterReq := openrouter.ChatCompletionRequest{
		Model:       req.Model,
		Messages:    convertMessagesToOpenRouter(req.Messages),
		MaxTokens:   req.MaxTokens,
		Temperature: req.Temperature,
	}
	resp, err := a.client.CreateChatCompletion(ctx, openrouterReq)
	if err != nil {
		return nil, err
	}
	return convertOpenRouterResponse(resp), nil
}

// Helper functions to convert between types

func convertMessagesToKolosal(msgs []ChatCompletionMessage) []kolosal.ChatCompletionMessage {
	result := make([]kolosal.ChatCompletionMessage, len(msgs))
	for i, msg := range msgs {
		result[i] = kolosal.ChatCompletionMessage{
			Role:    msg.Role,
			Content: msg.Content,
		}
	}
	return result
}

func convertMessagesToOpenRouter(msgs []ChatCompletionMessage) []openrouter.ChatCompletionMessage {
	result := make([]openrouter.ChatCompletionMessage, len(msgs))
	for i, msg := range msgs {
		result[i] = openrouter.ChatCompletionMessage{
			Role:    msg.Role,
			Content: msg.Content,
		}
	}
	return result
}

func convertKolosalResponse(resp *kolosal.ChatCompletionResponse) *ChatCompletionResponse {
	choices := make([]ChatChoice, len(resp.Choices))
	for i, choice := range resp.Choices {
		choices[i] = ChatChoice{
			Index:        choice.Index,
			Message:      ChatCompletionMessage{Role: choice.Message.Role, Content: choice.Message.Content},
			FinishReason: choice.FinishReason,
		}
	}
	var usage *Usage
	if resp.Usage != nil {
		usage = &Usage{
			PromptTokens:     resp.Usage.PromptTokens,
			CompletionTokens: resp.Usage.CompletionTokens,
			TotalTokens:      resp.Usage.TotalTokens,
		}
	}
	return &ChatCompletionResponse{
		ID:      resp.ID,
		Object:  resp.Object,
		Created: resp.Created,
		Model:   resp.Model,
		Choices: choices,
		Usage:   usage,
	}
}

func convertOpenRouterResponse(resp *openrouter.ChatCompletionResponse) *ChatCompletionResponse {
	choices := make([]ChatChoice, len(resp.Choices))
	for i, choice := range resp.Choices {
		choices[i] = ChatChoice{
			Index:        choice.Index,
			Message:      ChatCompletionMessage{Role: choice.Message.Role, Content: choice.Message.Content},
			FinishReason: choice.FinishReason,
		}
	}
	var usage *Usage
	if resp.Usage != nil {
		usage = &Usage{
			PromptTokens:     resp.Usage.PromptTokens,
			CompletionTokens: resp.Usage.CompletionTokens,
			TotalTokens:      resp.Usage.TotalTokens,
		}
	}
	return &ChatCompletionResponse{
		ID:      resp.ID,
		Object:  resp.Object,
		Created: resp.Created,
		Model:   resp.Model,
		Choices: choices,
		Usage:   usage,
	}
}
