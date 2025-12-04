package embedding

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/bantuaku/backend/logger"
)

const (
	KolosalEmbeddingAPIURL = "https://api.kolosal.ai/v1/embeddings"
	DefaultTimeout         = 60 * time.Second
	KolosalDimension       = 1536 // Kolosal.ai embedding dimension
)

// KolosalEmbedder implements Embedder using Kolosal.ai API
type KolosalEmbedder struct {
	apiKey     string
	httpClient *http.Client
	log        logger.Logger
}

// NewKolosalEmbedder creates a new Kolosal.ai embedder
func NewKolosalEmbedder(apiKey string) *KolosalEmbedder {
	return &KolosalEmbedder{
		apiKey: apiKey,
		httpClient: &http.Client{
			Timeout: DefaultTimeout,
		},
		log: *logger.Default(),
	}
}

// EmbeddingRequest represents a Kolosal.ai embedding request
type EmbeddingRequest struct {
	Input []string `json:"input"`
	Model string   `json:"model,omitempty"`
}

// EmbeddingResponse represents a Kolosal.ai embedding response
type EmbeddingResponse struct {
	Data []EmbeddingData `json:"data"`
}

// EmbeddingData represents a single embedding result
type EmbeddingData struct {
	Embedding []float32 `json:"embedding"`
	Index     int       `json:"index"`
}

// Embed generates embedding for a single text
func (k *KolosalEmbedder) Embed(ctx context.Context, text string) ([]float32, error) {
	embeddings, err := k.EmbedBatch(ctx, []string{text})
	if err != nil {
		return nil, err
	}
	if len(embeddings) == 0 {
		return nil, fmt.Errorf("no embedding returned")
	}
	return embeddings[0], nil
}

// EmbedBatch generates embeddings for multiple texts
func (k *KolosalEmbedder) EmbedBatch(ctx context.Context, texts []string) ([][]float32, error) {
	if len(texts) == 0 {
		return nil, fmt.Errorf("no texts provided")
	}

	reqBody := EmbeddingRequest{
		Input: texts,
		Model: "text-embedding-ada-002", // Default model, adjust if needed
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", KolosalEmbeddingAPIURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", k.apiKey))

	resp, err := k.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error: %d - %s", resp.StatusCode, string(body))
	}

	var embeddingResp EmbeddingResponse
	if err := json.NewDecoder(resp.Body).Decode(&embeddingResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

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
func (k *KolosalEmbedder) Dimension() int {
	return KolosalDimension
}

