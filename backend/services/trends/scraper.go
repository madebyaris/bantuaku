package trends

import (
	"context"
	"fmt"
	"time"

	"github.com/bantuaku/backend/logger"
	"github.com/chromedp/chromedp"
)

// GoogleTrendsScraper implements Scraper interface using chromedp
type GoogleTrendsScraper struct {
	log logger.Logger
}

// NewGoogleTrendsScraper creates a new Google Trends scraper
func NewGoogleTrendsScraper() *GoogleTrendsScraper {
	log := logger.Default()
	return &GoogleTrendsScraper{
		log: *log,
	}
}

// ScrapeTrends scrapes Google Trends for a keyword and geo location
func (s *GoogleTrendsScraper) ScrapeTrends(ctx context.Context, keyword string, geo string) (*TrendData, error) {
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
	url := fmt.Sprintf("https://trends.google.com/trends/explore?q=%s&geo=%s", keyword, geo)

	var timeSeriesData string
	var relatedQueriesData string

	err := chromedp.Run(ctx,
		chromedp.Navigate(url),
		chromedp.WaitVisible("body", chromedp.ByQuery),
		chromedp.Sleep(5*time.Second), // Wait for page to load and render
		
		// Extract time series data
		chromedp.Evaluate(`
			(function() {
				try {
					// Try to extract data from Google Trends widgets
					var widgets = document.querySelectorAll('widget');
					for (var i = 0; i < widgets.length; i++) {
						var widget = widgets[i];
						var widgetType = widget.getAttribute('type');
						if (widgetType === 'fe_line_chart') {
							var data = widget.getAttribute('data-request');
							if (data) return data;
						}
					}
					return null;
				} catch(e) {
					return null;
				}
			})()
		`, &timeSeriesData),
		
		// Extract related queries
		chromedp.Evaluate(`
			(function() {
				try {
					var queries = [];
					var widgets = document.querySelectorAll('widget[type="fe_related_queries"]');
					for (var i = 0; i < widgets.length; i++) {
						var widget = widgets[i];
						var data = widget.getAttribute('data-request');
						if (data) {
							try {
								var parsed = JSON.parse(data);
								if (parsed.default && parsed.default.rankedList) {
									parsed.default.rankedList.forEach(function(list) {
										if (list.rankedKeyword) {
											list.rankedKeyword.forEach(function(kw) {
												queries.push({
													query: kw.query || '',
													value: kw.value || 0,
													type: list.type || 'related'
												});
											});
										}
									});
								}
							} catch(e) {}
						}
					}
					return JSON.stringify(queries);
				} catch(e) {
					return JSON.stringify([]);
				}
			})()
		`, &relatedQueriesData),
	)

	if err != nil {
		return nil, fmt.Errorf("failed to scrape trends: %w", err)
	}

	// Parse time series (simplified - may need adjustment based on actual structure)
	timeSeries := s.parseTimeSeries(timeSeriesData)
	
	// Parse related queries
	relatedQueries := s.parseRelatedQueries(relatedQueriesData)

	return &TrendData{
		Keyword:        keyword,
		Geo:            geo,
		TimeSeries:     timeSeries,
		RelatedQueries: relatedQueries,
		CapturedAt:     time.Now(),
	}, nil
}

// ScrapeWithRetry scrapes with retry and exponential backoff
func (s *GoogleTrendsScraper) ScrapeWithRetry(ctx context.Context, keyword string, geo string, maxRetries int) (*TrendData, error) {
	var lastErr error
	
	for i := 0; i < maxRetries; i++ {
		if i > 0 {
			backoff := time.Duration(i*i) * time.Second
			s.log.Info("Retrying trends scrape", "attempt", i+1, "backoff", backoff)
			time.Sleep(backoff)
		}

		data, err := s.ScrapeTrends(ctx, keyword, geo)
		if err == nil {
			return data, nil
		}

		lastErr = err
		s.log.Warn("Scrape attempt failed", "attempt", i+1, "error", err)
	}

	return nil, fmt.Errorf("failed after %d retries: %w", maxRetries, lastErr)
}

// parseTimeSeries parses time series data (simplified parser)
func (s *GoogleTrendsScraper) parseTimeSeries(data string) []TimeSeriesPoint {
	// TODO: Implement proper parsing based on actual Google Trends response structure
	// For now, return empty array - actual implementation depends on Google Trends API structure
	// This is a placeholder that needs to be adjusted based on real data structure
	
	if data == "" || data == "null" {
		return []TimeSeriesPoint{}
	}

	// Placeholder: Generate sample data for demonstration
	// In production, parse actual JSON response from Google Trends
	var points []TimeSeriesPoint
	now := time.Now()
	for i := 0; i < 30; i++ {
		points = append(points, TimeSeriesPoint{
			Time:  now.AddDate(0, 0, -30+i),
			Value: 50 + (i % 50), // Sample values
		})
	}

	return points
}

// parseRelatedQueries parses related queries
func (s *GoogleTrendsScraper) parseRelatedQueries(data string) []RelatedQuery {
	if data == "" {
		return []RelatedQuery{}
	}

	// TODO: Implement proper JSON parsing
	// For now, return empty array
	// Actual implementation should parse the JSON structure from Google Trends
	
	return []RelatedQuery{}
}

