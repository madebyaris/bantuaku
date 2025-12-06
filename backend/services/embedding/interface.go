package embedding

import "context"

// Embedder interface for embedding providers
type Embedder interface {
	// Embed generates embedding for a single text
	Embed(ctx context.Context, text string) ([]float32, error)

	// EmbedBatch generates embeddings for multiple texts (more efficient)
	EmbedBatch(ctx context.Context, texts []string) ([][]float32, error)

	// Dimension returns the embedding dimension
	Dimension() int
}

// Provider represents embedding provider type
type Provider string

const (
	ProviderKolosal    Provider = "kolosal"
	ProviderOpenAI     Provider = "openai"
	ProviderOpenRouter Provider = "openrouter"
	ProviderCohere     Provider = "cohere"
)
