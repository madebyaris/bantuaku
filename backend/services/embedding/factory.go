package embedding

import (
	"fmt"

	"github.com/bantuaku/backend/config"
)

// NewEmbedder creates a new embedder based on configuration
func NewEmbedder(cfg *config.Config) (Embedder, error) {
	provider := Provider(cfg.EmbeddingProvider)

	switch provider {
	case ProviderOpenRouter:
		// OpenRouter embeddings (preferred if configured)
		if cfg.OpenRouterAPIKey == "" {
			return nil, fmt.Errorf("OpenRouter API key not configured (set OPENROUTER_API_KEY)")
		}
		model := cfg.OpenRouterModelEmbed
		if model == "" {
			model = "openai/text-embedding-3-small" // Default model
		}
		return NewOpenRouterEmbedder(cfg.OpenRouterAPIKey, model), nil

	case ProviderKolosal, "": // Default to Kolosal if empty
		apiKey := cfg.EmbeddingAPIKey
		if apiKey == "" {
			apiKey = cfg.KolosalAPIKey // Fallback to Kolosal API key
		}
		if apiKey == "" {
			return nil, fmt.Errorf("embedding API key not configured (set EMBEDDING_API_KEY or KOLOSAL_API_KEY)")
		}
		return NewKolosalEmbedder(apiKey), nil

	case ProviderOpenAI:
		return nil, fmt.Errorf("OpenAI provider not yet implemented - use 'openrouter' provider instead")

	case ProviderCohere:
		return nil, fmt.Errorf("Cohere provider not yet implemented")

	default:
		return nil, fmt.Errorf("unknown embedding provider: %s", provider)
	}
}
