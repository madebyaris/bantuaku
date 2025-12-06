package kolosal

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	KolosalAPIBaseURL = "https://api.kolosal.ai/v1/"
	DefaultTimeout    = 120 * time.Second // Increased timeout for AI chat completions (can take 30+ seconds)
)

// Client represents a Kolosal.ai API client
type Client struct {
	APIKey     string
	HTTPClient *http.Client
	BaseURL    string
}

// NewClient creates a new Kolosal.ai API client
func NewClient(apiKey string) *Client {
	return &Client{
		APIKey: apiKey,
		HTTPClient: &http.Client{
			Timeout: DefaultTimeout,
		},
		BaseURL: KolosalAPIBaseURL,
	}
}

// ChatCompletionRequest represents a chat completion request
type ChatCompletionRequest struct {
	Model       string                  `json:"model"`
	Messages    []ChatCompletionMessage `json:"messages"`
	Tools       []Tool                  `json:"tools,omitempty"`
	ToolChoice  interface{}             `json:"tool_choice,omitempty"` // "none", "auto", "required", or object
	MaxTokens   int                     `json:"max_tokens,omitempty"`
	Temperature float64                 `json:"temperature,omitempty"`
}

// ChatCompletionMessage represents a message in a chat completion
type ChatCompletionMessage struct {
	Role       string     `json:"role"` // "system", "user", "assistant", "tool"
	Content    string     `json:"content,omitempty"`
	ToolCalls  []ToolCall `json:"tool_calls,omitempty"`
	ToolCallID string     `json:"tool_call_id,omitempty"` // For tool messages
}

// ToolCall represents a function call from the AI model
type ToolCall struct {
	ID       string   `json:"id"`
	Type     string   `json:"type"` // "function"
	Function FuncCall `json:"function"`
}

// FuncCall represents the function details in a tool call
type FuncCall struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"` // JSON string
}

// Tool represents a function definition for the AI model
type Tool struct {
	Type     string   `json:"type"` // "function"
	Function Function `json:"function"`
}

// Function represents a function definition
type Function struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	Parameters  interface{} `json:"parameters"` // JSON schema object
}

// ChatCompletionResponse represents a chat completion response
type ChatCompletionResponse struct {
	ID      string       `json:"id,omitempty"`
	Object  string       `json:"object,omitempty"`
	Created int64        `json:"created,omitempty"`
	Model   string       `json:"model,omitempty"`
	Choices []ChatChoice `json:"choices"`
	Usage   *Usage       `json:"usage,omitempty"`
}

// ChatChoice represents a choice in the chat completion response
type ChatChoice struct {
	Index        int                   `json:"index,omitempty"`
	Message      ChatCompletionMessage `json:"message"`
	FinishReason string                `json:"finish_reason,omitempty"`
}

// Usage represents token usage information
type Usage struct {
	PromptTokens     int `json:"prompt_tokens,omitempty"`
	CompletionTokens int `json:"completion_tokens,omitempty"`
	TotalTokens      int `json:"total_tokens,omitempty"`
}

// CreateChatCompletion calls Kolosal.ai chat completions API
func (c *Client) CreateChatCompletion(ctx context.Context, req ChatCompletionRequest) (*ChatCompletionResponse, error) {
	url := fmt.Sprintf("%schat/completions", c.BaseURL)

	reqBody, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.APIKey))

	// Log request details for debugging
	fmt.Printf("[Kolosal Chat] Request URL: %s\n", url)
	fmt.Printf("[Kolosal Chat] Request Body Length: %d bytes\n", len(reqBody))
	fmt.Printf("[Kolosal Chat] Request Body: %s\n", string(reqBody))
	apiKeyPreview := c.APIKey
	if len(apiKeyPreview) > 10 {
		apiKeyPreview = apiKeyPreview[:10] + "..."
	}
	fmt.Printf("[Kolosal Chat] Authorization Header: Bearer %s\n", apiKeyPreview)

	resp, err := c.HTTPClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("[Kolosal Chat] Error Response Status: %d\n", resp.StatusCode)
		fmt.Printf("[Kolosal Chat] Error Response Body: %s\n", string(bodyBytes))
		return nil, fmt.Errorf("API error: %d - %s", resp.StatusCode, string(bodyBytes))
	}

	fmt.Printf("[Kolosal Chat] Success Response Status: %d\n", resp.StatusCode)
	fmt.Printf("[Kolosal Chat] Response Body Length: %d bytes\n", len(bodyBytes))

	var chatResp ChatCompletionResponse
	if err := json.Unmarshal(bodyBytes, &chatResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w, body: %s", err, string(bodyBytes))
	}

	// Validate response has choices
	if len(chatResp.Choices) == 0 {
		return nil, fmt.Errorf("empty choices in response: %s", string(bodyBytes))
	}

	// Validate first choice has message content OR tool_calls
	// Tool calls are valid even without content
	if chatResp.Choices[0].Message.Content == "" && len(chatResp.Choices[0].Message.ToolCalls) == 0 {
		return nil, fmt.Errorf("empty message content and no tool calls in response: %s", string(bodyBytes))
	}

	return &chatResp, nil
}

// OCRRequest represents an OCR request
// According to https://api.kolosal.ai/docs#tag/ocr/post/ocr
type OCRRequest struct {
	ImageData string `json:"image_data"` // Base64 encoded image with data URL prefix (e.g. "data:image/png;base64,...")
}

// OCRResponse represents an OCR response
type OCRResponse struct {
	ExtractedText   string  `json:"extracted_text,omitempty"`
	Text            string  `json:"text,omitempty"` // Alias for backwards compat
	ConfidenceScore float64 `json:"confidence_score,omitempty"`
	Notes           string  `json:"notes,omitempty"`
	ProcessingTime  float64 `json:"processing_time,omitempty"`
	Title           string  `json:"title,omitempty"`
	Date            string  `json:"date,omitempty"`
}

// GetText returns the extracted text from either field
func (r *OCRResponse) GetText() string {
	if r.ExtractedText != "" {
		return r.ExtractedText
	}
	return r.Text
}

// OCR performs OCR on an image
// According to https://api.kolosal.ai/docs#tag/ocr/post/ocr
func (c *Client) OCR(ctx context.Context, req OCRRequest) (*OCRResponse, error) {
	// Try direct endpoint first (without /v1/ prefix)
	url := "https://api.kolosal.ai/ocr"

	reqBody, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.APIKey))

	// Debug logging
	fmt.Printf("[Kolosal OCR] Request URL: %s\n", url)
	fmt.Printf("[Kolosal OCR] Image size (base64): %d bytes\n", len(req.Image))
	apiKeyPreview := c.APIKey
	if len(apiKeyPreview) > 10 {
		apiKeyPreview = apiKeyPreview[:10] + "..."
	}
	fmt.Printf("[Kolosal OCR] Authorization: Bearer %s\n", apiKeyPreview)

	resp, err := c.HTTPClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("[Kolosal OCR] Error: %d - %s\n", resp.StatusCode, string(body))
		return nil, fmt.Errorf("API error: %d - %s", resp.StatusCode, string(body))
	}

	fmt.Printf("[Kolosal OCR] Success: %d\n", resp.StatusCode)

	var ocrResp OCRResponse
	if err := json.Unmarshal(body, &ocrResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &ocrResp, nil
}

// OCRFormRequest represents an OCR form extraction request
type OCRFormRequest struct {
	ImageURL string `json:"image_url,omitempty"`
	Image    string `json:"image,omitempty"` // Base64 encoded image
	Language string `json:"language,omitempty"`
}

// OCRFormResponse represents an OCR form extraction response
type OCRFormResponse struct {
	Fields map[string]interface{} `json:"fields"`
}

// OCRForm performs OCR form extraction on an image
func (c *Client) OCRForm(ctx context.Context, req OCRFormRequest) (*OCRFormResponse, error) {
	url := fmt.Sprintf("%socrform", c.BaseURL)

	reqBody, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.APIKey))

	resp, err := c.HTTPClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error: %d - %s", resp.StatusCode, string(body))
	}

	var ocrFormResp OCRFormResponse
	if err := json.NewDecoder(resp.Body).Decode(&ocrFormResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &ocrFormResp, nil
}
