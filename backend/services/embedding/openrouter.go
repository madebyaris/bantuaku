package embedding

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/bantuaku/backend/logger"
)

const (
	OpenRouterEmbeddingAPIURL = "https://openrouter.ai/api/v1/embeddings"
)

// OpenRouterEmbedder implements Embedder using OpenRouter API
type OpenRouterEmbedder struct {
	apiKey     string
	model      string
	dimension  int
	httpClient *http.Client
	log        logger.Logger
}

// NewOpenRouterEmbedder creates a new OpenRouter embedder
func NewOpenRouterEmbedder(apiKey, model string) *OpenRouterEmbedder {
	// Determine dimension based on model
	dimension := 1536 // Default for text-embedding-ada-002
	switch model {
	case "openai/text-embedding-3-small":
		dimension = 1536
	case "openai/text-embedding-3-large":
		dimension = 3072
	case "openai/text-embedding-ada-002":
		dimension = 1536
	}

	return &OpenRouterEmbedder{
		apiKey:    apiKey,
		model:     model,
		dimension: dimension,
		httpClient: &http.Client{
			Timeout: DefaultTimeout,
		},
		log: *logger.Default(),
	}
}

// OpenRouterEmbeddingRequest represents an OpenRouter embedding request
type OpenRouterEmbeddingRequest struct {
	Input []string `json:"input"`
	Model string   `json:"model"`
}

// OpenRouterEmbeddingResponse represents an OpenRouter embedding response
type OpenRouterEmbeddingResponse struct {
	Data  []OpenRouterEmbeddingData `json:"data"`
	Model string                    `json:"model"`
	Usage struct {
		PromptTokens int `json:"prompt_tokens"`
		TotalTokens  int `json:"total_tokens"`
	} `json:"usage"`
}

// OpenRouterEmbeddingData represents a single embedding result
type OpenRouterEmbeddingData struct {
	Embedding []float32 `json:"embedding"`
	Index     int       `json:"index"`
	Object    string    `json:"object"`
}

// Embed generates embedding for a single text
func (o *OpenRouterEmbedder) Embed(ctx context.Context, text string) ([]float32, error) {
	embeddings, err := o.EmbedBatch(ctx, []string{text})
	if err != nil {
		return nil, err
	}
	if len(embeddings) == 0 {
		return nil, fmt.Errorf("no embedding returned")
	}
	return embeddings[0], nil
}

// EmbedBatch generates embeddings for multiple texts
func (o *OpenRouterEmbedder) EmbedBatch(ctx context.Context, texts []string) ([][]float32, error) {
	if len(texts) == 0 {
		return nil, fmt.Errorf("no texts provided")
	}

	reqBody := OpenRouterEmbeddingRequest{
		Input: texts,
		Model: o.model,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	o.log.Debug("OpenRouter embedding request", "model", o.model, "texts_count", len(texts))

	req, err := http.NewRequestWithContext(ctx, "POST", OpenRouterEmbeddingAPIURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", o.apiKey))
	req.Header.Set("HTTP-Referer", "https://bantuaku.com")
	req.Header.Set("X-Title", "Bantuaku")

	resp, err := o.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("OpenRouter API error: %d - %s", resp.StatusCode, string(body))
	}

	var embeddingResp OpenRouterEmbeddingResponse
	if err := json.NewDecoder(resp.Body).Decode(&embeddingResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	o.log.Debug("OpenRouter embedding response", "data_count", len(embeddingResp.Data), "tokens", embeddingResp.Usage.TotalTokens)

	if len(embeddingResp.Data) != len(texts) {
		return nil, fmt.Errorf("mismatched embedding count: expected %d, got %d", len(texts), len(embeddingResp.Data))
	}

	// Sort embeddings by index to ensure correct order
	embeddings := make([][]float32, len(texts))
	for _, data := range embeddingResp.Data {
		if data.Index >= 0 && data.Index < len(texts) {
			embeddings[data.Index] = data.Embedding
		}
	}

	// Verify all embeddings are present
	for i, emb := range embeddings {
		if emb == nil {
			return nil, fmt.Errorf("missing embedding at index %d", i)
		}
	}

	return embeddings, nil
}

// Dimension returns the embedding dimension
func (o *OpenRouterEmbedder) Dimension() int {
	return o.dimension
}
