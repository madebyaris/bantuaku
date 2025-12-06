package exa

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
	baseURL = "https://api.exa.ai"
)

// Client is the Exa.ai API client
type Client struct {
	apiKey     string
	httpClient *http.Client
	log        logger.Logger
}

// NewClient creates a new Exa.ai client
func NewClient(apiKey string) *Client {
	log := logger.Default()
	return &Client{
		apiKey: apiKey,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		log: *log,
	}
}

// IsConfigured returns true if the client has an API key
func (c *Client) IsConfigured() bool {
	return c.apiKey != ""
}

// SearchRequest represents a search request to Exa.ai
type SearchRequest struct {
	Query              string   `json:"query"`
	NumResults         int      `json:"numResults,omitempty"`
	Type               string   `json:"type,omitempty"`     // "neural", "keyword", "auto", "fast"
	Category           string   `json:"category,omitempty"` // "company", "news", "tweet", etc.
	IncludeDomains     []string `json:"includeDomains,omitempty"`
	ExcludeDomains     []string `json:"excludeDomains,omitempty"`
	StartPublishedDate string   `json:"startPublishedDate,omitempty"` // ISO 8601 format
	EndPublishedDate   string   `json:"endPublishedDate,omitempty"`
	IncludeText        []string `json:"includeText,omitempty"`
	ExcludeText        []string `json:"excludeText,omitempty"`
	UseAutoprompt      bool     `json:"useAutoprompt,omitempty"`
}

// ContentsRequest represents a request to get contents with search
type ContentsRequest struct {
	SearchRequest
	Text       TextOptions       `json:"text,omitempty"`
	Highlights HighlightsOptions `json:"highlights,omitempty"`
}

// TextOptions for content retrieval
type TextOptions struct {
	MaxCharacters   int  `json:"maxCharacters,omitempty"`
	IncludeHtmlTags bool `json:"includeHtmlTags,omitempty"`
}

// HighlightsOptions for highlights retrieval
type HighlightsOptions struct {
	Query            string `json:"query,omitempty"`
	NumSentences     int    `json:"numSentences,omitempty"`
	HighlightsPerUrl int    `json:"highlightsPerUrl,omitempty"`
}

// SearchResult represents a single search result
type SearchResult struct {
	Title           string    `json:"title"`
	URL             string    `json:"url"`
	ID              string    `json:"id"`
	PublishedDate   string    `json:"publishedDate,omitempty"`
	Author          string    `json:"author,omitempty"`
	Score           float64   `json:"score,omitempty"`
	Text            string    `json:"text,omitempty"`
	Highlights      []string  `json:"highlights,omitempty"`
	HighlightScores []float64 `json:"highlightScores,omitempty"`
}

// SearchResponse represents the response from a search
type SearchResponse struct {
	Results   []SearchResult `json:"results"`
	RequestID string         `json:"requestId,omitempty"`
}

// Search performs a basic search
func (c *Client) Search(ctx context.Context, req SearchRequest) (*SearchResponse, error) {
	if !c.IsConfigured() {
		return nil, fmt.Errorf("Exa API key not configured")
	}

	if req.NumResults == 0 {
		req.NumResults = 5
	}
	if req.Type == "" {
		req.Type = "auto"
	}

	return c.doRequest(ctx, "/search", req)
}

// SearchAndContents performs a search and retrieves content/highlights
func (c *Client) SearchAndContents(ctx context.Context, req ContentsRequest) (*SearchResponse, error) {
	if !c.IsConfigured() {
		return nil, fmt.Errorf("Exa API key not configured")
	}

	if req.NumResults == 0 {
		req.NumResults = 5
	}
	if req.Type == "" {
		req.Type = "auto"
	}

	// Set default text options if not specified
	if req.Text.MaxCharacters == 0 {
		req.Text.MaxCharacters = 2000
	}

	// Set default highlights options if not specified
	if req.Highlights.NumSentences == 0 {
		req.Highlights.NumSentences = 3
	}
	if req.Highlights.HighlightsPerUrl == 0 {
		req.Highlights.HighlightsPerUrl = 2
	}

	return c.doRequest(ctx, "/search", req)
}

// SearchNews searches for news articles
func (c *Client) SearchNews(ctx context.Context, query string, numResults int, daysBack int) (*SearchResponse, error) {
	startDate := time.Now().AddDate(0, 0, -daysBack).Format("2006-01-02")

	req := ContentsRequest{
		SearchRequest: SearchRequest{
			Query:              query,
			NumResults:         numResults,
			Type:               "neural",
			Category:           "news",
			StartPublishedDate: startDate,
		},
		Text: TextOptions{
			MaxCharacters: 1500,
		},
		Highlights: HighlightsOptions{
			NumSentences:     3,
			HighlightsPerUrl: 2,
		},
	}

	return c.SearchAndContents(ctx, req)
}

// SearchSocialMediaTrends searches for social media trends and marketing articles
func (c *Client) SearchSocialMediaTrends(ctx context.Context, industry, location string, keywords []string) (*SearchResponse, error) {
	// Build search query with current year
	currentYear := time.Now().Year()
	query := fmt.Sprintf("social media marketing trends %s %s Indonesia %d", industry, location, currentYear)
	if len(keywords) > 0 {
		query += " " + keywords[0] // Add first keyword for relevance
	}

	startDate := time.Now().AddDate(0, -3, 0).Format("2006-01-02") // Last 3 months

	req := ContentsRequest{
		SearchRequest: SearchRequest{
			Query:              query,
			NumResults:         5,
			Type:               "neural",
			StartPublishedDate: startDate,
		},
		Text: TextOptions{
			MaxCharacters: 2000,
		},
		Highlights: HighlightsOptions{
			Query:            fmt.Sprintf("social media %s", industry),
			NumSentences:     3,
			HighlightsPerUrl: 3,
		},
	}

	return c.SearchAndContents(ctx, req)
}

// SearchMarketTrends searches for market and industry trends
func (c *Client) SearchMarketTrends(ctx context.Context, industry, location string, keywords []string) (*SearchResponse, error) {
	// Build search query with current year
	currentYear := time.Now().Year()
	query := fmt.Sprintf("market trends %s %s Indonesia %d business outlook", industry, location, currentYear)

	startDate := time.Now().AddDate(0, -6, 0).Format("2006-01-02") // Last 6 months

	req := ContentsRequest{
		SearchRequest: SearchRequest{
			Query:              query,
			NumResults:         5,
			Type:               "neural",
			Category:           "news",
			StartPublishedDate: startDate,
		},
		Text: TextOptions{
			MaxCharacters: 2000,
		},
		Highlights: HighlightsOptions{
			Query:            fmt.Sprintf("market %s growth", industry),
			NumSentences:     3,
			HighlightsPerUrl: 3,
		},
	}

	return c.SearchAndContents(ctx, req)
}

// doRequest performs the HTTP request to Exa.ai
func (c *Client) doRequest(ctx context.Context, endpoint string, body interface{}) (*SearchResponse, error) {
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	c.log.Debug("Exa API request", "endpoint", endpoint, "body_length", len(jsonBody))

	req, err := http.NewRequestWithContext(ctx, "POST", baseURL+endpoint, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	c.log.Debug("Exa API response", "status", resp.StatusCode, "body_length", len(respBody))

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, string(respBody))
	}

	var searchResp SearchResponse
	if err := json.Unmarshal(respBody, &searchResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &searchResp, nil
}

// FormatResultsForAI formats search results into a string for AI consumption
func FormatResultsForAI(results []SearchResult) string {
	if len(results) == 0 {
		return "Tidak ada hasil pencarian yang ditemukan."
	}

	var output string
	for i, r := range results {
		output += fmt.Sprintf("\n--- Sumber %d: %s ---\n", i+1, r.Title)
		if r.PublishedDate != "" {
			output += fmt.Sprintf("Tanggal: %s\n", r.PublishedDate)
		}
		if r.Author != "" {
			output += fmt.Sprintf("Penulis: %s\n", r.Author)
		}

		// Add highlights first (more relevant)
		if len(r.Highlights) > 0 {
			output += "Poin penting:\n"
			for _, h := range r.Highlights {
				output += fmt.Sprintf("• %s\n", h)
			}
		}

		// Add some text content if available
		if r.Text != "" {
			// Truncate to first 500 chars if too long
			text := r.Text
			if len(text) > 500 {
				text = text[:500] + "..."
			}
			output += fmt.Sprintf("Konten: %s\n", text)
		}

		output += fmt.Sprintf("URL: %s\n", r.URL)
	}

	return output
}
