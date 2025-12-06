package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/chromedp/chromedp"
)

// TrendData represents Google Trends data
type TrendData struct {
	Keyword        string
	Geo            string
	TimeSeries     []TimeSeriesPoint
	RelatedQueries []RelatedQuery
	CapturedAt     time.Time
}

// TimeSeriesPoint represents a single data point in time series
type TimeSeriesPoint struct {
	Time  time.Time
	Value int // 0-100 interest value
}

// RelatedQuery represents a related search query
type RelatedQuery struct {
	Query          string
	RelationshipType string // "rising", "top", "related"
	Value          int     // Interest value or growth percentage
}

// Scraper handles Google Trends scraping
type Scraper struct {
	baseURL    string
	httpClient *http.Client
}

// NewScraper creates a new trends scraper
func NewScraper() *Scraper {
	return &Scraper{
		baseURL: "https://trends.google.com",
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// ScrapeTrends scrapes Google Trends for a keyword and geo location
func (s *Scraper) ScrapeTrends(ctx context.Context, keyword string, geo string) (*TrendData, error) {
	// Create headless browser context
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", true),
		chromedp.Flag("disable-gpu", true),
		chromedp.Flag("no-sandbox", true),
		chromedp.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36"),
	)

	allocCtx, cancel := chromedp.NewExecAllocator(ctx, opts...)
	defer cancel()

	ctx, cancel = chromedp.NewContext(allocCtx)
	defer cancel()

	// Set timeout
	ctx, cancel = context.WithTimeout(ctx, 60*time.Second)
	defer cancel()

	// Build Google Trends URL
	url := fmt.Sprintf("%s/trends/explore?q=%s&geo=%s", s.baseURL, keyword, geo)

	var timeSeriesJSON string
	var relatedQueriesJSON string

	err := chromedp.Run(ctx,
		chromedp.Navigate(url),
		chromedp.WaitVisible("body", chromedp.ByQuery),
		chromedp.Sleep(3*time.Second), // Wait for page to load
		
		// Extract time series data from widget
		chromedp.Evaluate(`
			(function() {
				try {
					var widget = document.querySelector('widget[type="fe_line_chart"]');
					if (!widget) return null;
					var data = widget.getAttribute('data-request');
					if (data) {
						var parsed = JSON.parse(data);
						return JSON.stringify(parsed);
					}
					return null;
				} catch(e) {
					return null;
				}
			})()
		`, &timeSeriesJSON),
		
		// Extract related queries
		chromedp.Evaluate(`
			(function() {
				try {
					var relatedQueries = [];
					var widgets = document.querySelectorAll('widget[type="fe_related_queries"]');
					widgets.forEach(function(widget) {
						var data = widget.getAttribute('data-request');
						if (data) {
							var parsed = JSON.parse(data);
							if (parsed.default && parsed.default.rankedList) {
								parsed.default.rankedList.forEach(function(list) {
									if (list.rankedKeyword) {
										list.rankedKeyword.forEach(function(kw) {
											relatedQueries.push({
												query: kw.query,
												value: kw.value || 0,
												type: list.type || 'related'
											});
										});
									}
								});
							}
						}
					});
					return JSON.stringify(relatedQueries);
				} catch(e) {
					return JSON.stringify([]);
				}
			})()
		`, &relatedQueriesJSON),
	)

	if err != nil {
		return nil, fmt.Errorf("failed to scrape trends: %w", err)
	}

	// Parse time series data
	timeSeries, err := s.parseTimeSeries(timeSeriesJSON)
	if err != nil {
		log.Printf("Warning: Failed to parse time series: %v", err)
		timeSeries = []TimeSeriesPoint{}
	}

	// Parse related queries
	relatedQueries, err := s.parseRelatedQueries(relatedQueriesJSON)
	if err != nil {
		log.Printf("Warning: Failed to parse related queries: %v", err)
		relatedQueries = []RelatedQuery{}
	}

	return &TrendData{
		Keyword:        keyword,
		Geo:            geo,
		TimeSeries:     timeSeries,
		RelatedQueries: relatedQueries,
		CapturedAt:     time.Now(),
	}, nil
}

// parseTimeSeries parses time series JSON from Google Trends
func (s *Scraper) parseTimeSeries(jsonStr string) ([]TimeSeriesPoint, error) {
	if jsonStr == "" || jsonStr == "null" {
		return []TimeSeriesPoint{}, nil
	}

	var data map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &data); err != nil {
		return nil, err
	}

	// Extract timeline data (adjust based on actual Google Trends structure)
	// This is a simplified parser - may need adjustment based on actual API response
	var points []TimeSeriesPoint
	
	// Try to extract timeline data
	if timeline, ok := data["timelineData"].([]interface{}); ok {
		for _, item := range timeline {
			if point, ok := item.(map[string]interface{}); ok {
				if timeStr, ok := point["time"].(string); ok {
					if value, ok := point["value"].([]interface{}); ok && len(value) > 0 {
						if val, ok := value[0].(float64); ok {
							t, err := time.Parse("2006-01-02", timeStr)
							if err == nil {
								points = append(points, TimeSeriesPoint{
									Time:  t,
									Value: int(val),
								})
							}
						}
					}
				}
			}
		}
	}

	return points, nil
}

// parseRelatedQueries parses related queries JSON
func (s *Scraper) parseRelatedQueries(jsonStr string) ([]RelatedQuery, error) {
	if jsonStr == "" {
		return []RelatedQuery{}, nil
	}

	var queries []map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &queries); err != nil {
		return nil, err
	}

	var relatedQueries []RelatedQuery
	for _, q := range queries {
		query := RelatedQuery{
			Query:          getString(q, "query"),
			RelationshipType: getString(q, "type"),
			Value:          getInt(q, "value"),
		}
		relatedQueries = append(relatedQueries, query)
	}

	return relatedQueries, nil
}

// Helper functions
func getString(m map[string]interface{}, key string) string {
	if val, ok := m[key]; ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return ""
}

func getInt(m map[string]interface{}, key string) int {
	if val, ok := m[key]; ok {
		switch v := val.(type) {
		case float64:
			return int(v)
		case int:
			return v
		case string:
			var i int
			fmt.Sscanf(v, "%d", &i)
			return i
		}
	}
	return 0
}

// ScrapeWithRetry scrapes with retry and backoff
func (s *Scraper) ScrapeWithRetry(ctx context.Context, keyword string, geo string, maxRetries int) (*TrendData, error) {
	var lastErr error
	
	for i := 0; i < maxRetries; i++ {
		if i > 0 {
			// Exponential backoff
			backoff := time.Duration(i*i) * time.Second
			log.Printf("Retrying after %v...", backoff)
			time.Sleep(backoff)
		}

		data, err := s.ScrapeTrends(ctx, keyword, geo)
		if err == nil {
			return data, nil
		}

		lastErr = err
		log.Printf("Scrape attempt %d failed: %v", i+1, err)
	}

	return nil, fmt.Errorf("failed after %d retries: %w", maxRetries, lastErr)
}

// ValidateGeo validates geo code
func ValidateGeo(geo string) bool {
	// Basic validation - should be 2-letter country code
	return len(geo) == 2 && strings.ToUpper(geo) == geo
}

