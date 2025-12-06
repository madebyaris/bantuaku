package regulations

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/bantuaku/backend/logger"
	"github.com/bantuaku/backend/services/exa"
)

// DiscoveredRegulation represents a regulation found via search
type DiscoveredRegulation struct {
	Title         string
	Summary       string
	Content       string
	SourceURL     string
	PDFURL        string // May be empty if no PDF
	PublishedDate string
	Category      string
	Source        string // "exa" or "official"
}

// DiscoveryService handles multi-source regulation discovery
type DiscoveryService struct {
	exaClient *exa.Client
	log       logger.Logger
}

// NewDiscoveryService creates a new discovery service
func NewDiscoveryService(exaClient *exa.Client) *DiscoveryService {
	return &DiscoveryService{
		exaClient: exaClient,
		log:       *logger.Default(),
	}
}

// DiscoverRegulations searches for regulations using keywords via Exa.ai
func (ds *DiscoveryService) DiscoverRegulations(ctx context.Context, keywords []string, maxPerKeyword int) ([]DiscoveredRegulation, error) {
	if ds.exaClient == nil || !ds.exaClient.IsConfigured() {
		return nil, fmt.Errorf("Exa.ai client not configured")
	}

	ds.log.Info("Starting regulation discovery", "keywords_count", len(keywords), "max_per_keyword", maxPerKeyword)

	var allResults []DiscoveredRegulation
	seen := make(map[string]bool) // Dedupe by URL

	for i, keyword := range keywords {
		ds.log.Debug("Searching keyword", "index", i+1, "keyword", keyword)

		results, err := ds.searchKeyword(ctx, keyword, maxPerKeyword)
		if err != nil {
			ds.log.Warn("Search failed for keyword", "keyword", keyword, "error", err)
			continue
		}

		// Dedupe and add results
		for _, r := range results {
			if !seen[r.SourceURL] {
				seen[r.SourceURL] = true
				allResults = append(allResults, r)
			}
		}

		// Rate limiting between keywords
		if i < len(keywords)-1 {
			time.Sleep(500 * time.Millisecond)
		}
	}

	ds.log.Info("Discovery completed", "total_found", len(allResults))
	return allResults, nil
}

// searchKeyword searches for a single keyword via Exa.ai
func (ds *DiscoveryService) searchKeyword(ctx context.Context, keyword string, numResults int) ([]DiscoveredRegulation, error) {
	// Build search query for Indonesian regulations
	query := fmt.Sprintf("regulasi peraturan pemerintah Indonesia %s", keyword)

	// Search for recent content (last 2 years)
	startDate := time.Now().AddDate(-2, 0, 0).Format("2006-01-02")

	req := exa.ContentsRequest{
		SearchRequest: exa.SearchRequest{
			Query:              query,
			NumResults:         numResults,
			Type:               "neural",
			StartPublishedDate: startDate,
			// Focus on Indonesian sources - prioritize news/articles that have full text
			IncludeDomains: []string{
				"hukumonline.com",
				"kompas.com",
				"detik.com",
				"bisnis.com",
				"kontan.co.id",
				"cnbcindonesia.com",
				"ekonomi.republika.co.id",
				"kemenkopukm.go.id",
				"pajak.go.id",
				"bpom.go.id",
				"oss.go.id",
			},
		},
		Text: exa.TextOptions{
			MaxCharacters:   5000, // Get more content for summarization
			IncludeHtmlTags: false,
		},
		Highlights: exa.HighlightsOptions{
			Query:            keyword,
			NumSentences:     8,
			HighlightsPerUrl: 5,
		},
	}

	resp, err := ds.exaClient.SearchAndContents(ctx, req)
	if err != nil {
		return nil, err
	}

	var results []DiscoveredRegulation
	for _, r := range resp.Results {
		reg := DiscoveredRegulation{
			Title:         r.Title,
			SourceURL:     r.URL,
			PublishedDate: r.PublishedDate,
			Source:        "exa",
		}

		// Use text content if available
		if r.Text != "" {
			reg.Content = r.Text
		}

		// Build summary from highlights
		if len(r.Highlights) > 0 {
			reg.Summary = strings.Join(r.Highlights, " ")
		}

		// Check if URL points to a PDF
		if strings.HasSuffix(strings.ToLower(r.URL), ".pdf") {
			reg.PDFURL = r.URL
		}

		// Categorize based on URL and content
		reg.Category = ds.categorizeRegulation(r.URL, r.Title, reg.Content)

		results = append(results, reg)
	}

	return results, nil
}

// categorizeRegulation determines the UMKM category of a regulation
func (ds *DiscoveryService) categorizeRegulation(url, title, content string) string {
	combined := strings.ToLower(url + " " + title + " " + content)

	// Check for category indicators
	categories := map[string][]string{
		"pajak":           {"pajak", "pph", "ppn", "npwp", "perpajakan", "tax", "fiskal"},
		"perizinan":       {"izin", "perizinan", "nib", "siup", "oss", "berusaha", "license"},
		"ketenagakerjaan": {"tenaga kerja", "ketenagakerjaan", "umr", "bpjs", "upah", "karyawan", "pekerja"},
		"pangan":          {"pangan", "bpom", "halal", "makanan", "minuman", "pirt", "food"},
		"haki":            {"haki", "merek", "paten", "hak cipta", "kekayaan intelektual"},
		"ekspor_impor":    {"ekspor", "impor", "bea cukai", "customs", "perdagangan luar"},
		"lingkungan":      {"lingkungan", "amdal", "limbah", "environment"},
		"standar":         {"sni", "standar", "sertifikasi", "mutu"},
	}

	for category, keywords := range categories {
		for _, kw := range keywords {
			if strings.Contains(combined, kw) {
				return category
			}
		}
	}

	return "umum"
}

// SearchOfficialSite searches peraturan.go.id directly (existing crawler logic)
// This is a wrapper to maintain compatibility with existing code
func (ds *DiscoveryService) SearchOfficialSite(ctx context.Context, crawler *Crawler, maxPages int) ([]DiscoveredRegulation, error) {
	ds.log.Info("Searching official site", "max_pages", maxPages)

	// Use existing crawler
	regulations, err := crawler.CrawlRegulations(ctx, maxPages)
	if err != nil {
		return nil, err
	}

	// Convert to DiscoveredRegulation format
	var results []DiscoveredRegulation
	for _, reg := range regulations {
		discovered := DiscoveredRegulation{
			Title:     reg.Title,
			SourceURL: reg.SourceURL,
			PDFURL:    reg.PDFURL,
			Category:  reg.Category,
			Source:    "official",
		}

		if reg.PublishedDate != nil {
			discovered.PublishedDate = reg.PublishedDate.Format("2006-01-02")
		}

		results = append(results, discovered)
	}

	return results, nil
}
