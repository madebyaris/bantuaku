package embedding

import (
	"fmt"

	"github.com/bantuaku/backend/config"
)

// NewEmbedder creates a new embedder based on configuration
func NewEmbedder(cfg *config.Config) (Embedder, error) {
	provider := Provider(cfg.EmbeddingProvider)
	
	switch provider {
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
		return nil, fmt.Errorf("OpenAI provider not yet implemented")
		
	case ProviderCohere:
		return nil, fmt.Errorf("Cohere provider not yet implemented")
		
	default:
		return nil, fmt.Errorf("unknown embedding provider: %s", provider)
	}
}

