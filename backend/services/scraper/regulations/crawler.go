package regulations

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/bantuaku/backend/logger"
)

// Regulation represents a regulation metadata
type Regulation struct {
	Title           string
	RegulationNumber string
	Year            int
	Category        string
	SourceURL       string
	PDFURL          string
	PublishedDate   *time.Time
	EffectiveDate   *time.Time
}

// Crawler handles crawling peraturan.go.id
type Crawler struct {
	baseURL    string
	httpClient *http.Client
	log        logger.Logger
}

// NewCrawler creates a new crawler instance
func NewCrawler(baseURL string) *Crawler {
	return &Crawler{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		log: *logger.Default(),
	}
}

// CrawlRegulations crawls peraturan.go.id and discovers regulation PDFs
func (c *Crawler) CrawlRegulations(ctx context.Context, maxPages int) ([]Regulation, error) {
	var regulations []Regulation
	visited := make(map[string]bool)

	// Start from main listing page
	startURL := fmt.Sprintf("%s/peraturan", c.baseURL)
	
	c.log.Info("Starting regulation crawl", "base_url", c.baseURL)

	// Crawl listing pages
	for page := 1; page <= maxPages; page++ {
		listURL := fmt.Sprintf("%s?page=%d", startURL, page)
		
		c.log.Info("Crawling page", "page", page, "url", listURL)

		regs, hasMore, err := c.crawlListingPage(ctx, listURL, visited)
		if err != nil {
			c.log.Warn("Failed to crawl page", "page", page, "error", err)
			continue
		}

		regulations = append(regulations, regs...)

		if !hasMore {
			break
		}

		// Rate limiting
		time.Sleep(2 * time.Second)
	}

	c.log.Info("Crawl completed", "total_regulations", len(regulations))
	return regulations, nil
}

// crawlListingPage crawls a single listing page
func (c *Crawler) crawlListingPage(ctx context.Context, url string, visited map[string]bool) ([]Regulation, bool, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, false, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, false, fmt.Errorf("failed to fetch page: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, false, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, false, fmt.Errorf("failed to parse HTML: %w", err)
	}

	var regulations []Regulation

	// Extract regulation links (adjust selector based on actual site structure)
	doc.Find("a[href*='/peraturan/']").Each(func(i int, s *goquery.Selection) {
		href, exists := s.Attr("href")
		if !exists {
			return
		}

		// Build full URL
		detailURL := c.buildFullURL(href)
		if visited[detailURL] {
			return
		}

		// Crawl detail page
		reg, err := c.crawlDetailPage(ctx, detailURL)
		if err != nil {
			c.log.Warn("Failed to crawl detail page", "url", detailURL, "error", err)
			return
		}

		if reg != nil {
			visited[detailURL] = true
			regulations = append(regulations, *reg)
		}
	})

	// Check if there's a next page
	hasMore := doc.Find("a[href*='page=']").Length() > 0

	return regulations, hasMore, nil
}

// crawlDetailPage crawls a regulation detail page to extract metadata and PDF URL
func (c *Crawler) crawlDetailPage(ctx context.Context, url string) (*Regulation, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch page: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML: %w", err)
	}

	reg := &Regulation{
		SourceURL: url,
	}

	// Extract title
	reg.Title = strings.TrimSpace(doc.Find("h1, .title, .regulation-title").First().Text())

	// Extract regulation number (adjust selector based on actual site)
	reg.RegulationNumber = strings.TrimSpace(doc.Find(".regulation-number, .number").First().Text())

	// Extract year from regulation number or separate field
	reg.Year = c.extractYear(reg.RegulationNumber)

	// Extract category
	reg.Category = strings.TrimSpace(doc.Find(".category, .type").First().Text())

	// Find PDF link
	doc.Find("a[href$='.pdf'], a[href*='download']").Each(func(i int, s *goquery.Selection) {
		href, exists := s.Attr("href")
		if exists && strings.Contains(strings.ToLower(href), "pdf") {
			reg.PDFURL = c.buildFullURL(href)
		}
	})

	// Extract dates (adjust selectors based on actual site)
	publishedText := strings.TrimSpace(doc.Find(".published-date, .date").First().Text())
	if publishedText != "" {
		if t := c.parseDate(publishedText); t != nil {
			reg.PublishedDate = t
		}
	}

	// Validate we have essential data
	if reg.Title == "" || reg.PDFURL == "" {
		return nil, fmt.Errorf("missing essential data: title=%s, pdf_url=%s", reg.Title, reg.PDFURL)
	}

	return reg, nil
}

// buildFullURL builds a full URL from a relative or absolute URL
func (c *Crawler) buildFullURL(href string) string {
	if strings.HasPrefix(href, "http://") || strings.HasPrefix(href, "https://") {
		return href
	}
	if strings.HasPrefix(href, "/") {
		return c.baseURL + href
	}
	return c.baseURL + "/" + href
}

// extractYear extracts year from regulation number
func (c *Crawler) extractYear(regulationNumber string) int {
	// Try to find year pattern like "Tahun 2023" or "2023"
	// This is a simple implementation - may need refinement
	parts := strings.Fields(regulationNumber)
	for _, part := range parts {
		if len(part) == 4 {
			var year int
			if _, err := fmt.Sscanf(part, "%d", &year); err == nil {
				if year >= 2000 && year <= 2100 {
					return year
				}
			}
		}
	}
	return 0
}

// parseDate parses Indonesian date format
func (c *Crawler) parseDate(dateStr string) *time.Time {
	// Common Indonesian date formats
	formats := []string{
		"02-01-2006",
		"02/01/2006",
		"2006-01-02",
		"02 Januari 2006",
		"02 Jan 2006",
	}

	for _, format := range formats {
		if t, err := time.Parse(format, dateStr); err == nil {
			return &t
		}
	}

	return nil
}

