package regulations

import (
	"context"
	"fmt"
	"strings"

	"github.com/bantuaku/backend/logger"
	"github.com/bantuaku/backend/services/chat"
)

// ProcessedRegulation represents a regulation after content processing
type ProcessedRegulation struct {
	Title         string
	Summary       string // AI-generated summary
	FullText      string // Full extracted text
	SourceURL     string
	PDFURL        string
	PublishedDate string
	Category      string
	Source        string
	WasPDF        bool // True if content was extracted from PDF
}

// ContentProcessor handles content extraction and summarization
type ContentProcessor struct {
	extractor    *Extractor
	chatProvider chat.ChatProvider
	chatModel    string
	log          logger.Logger
}

// NewContentProcessor creates a new content processor
func NewContentProcessor(extractor *Extractor, chatProvider chat.ChatProvider, chatModel string) *ContentProcessor {
	return &ContentProcessor{
		extractor:    extractor,
		chatProvider: chatProvider,
		chatModel:    chatModel,
		log:          *logger.Default(),
	}
}

// ProcessRegulation processes a discovered regulation to extract and summarize content
func (cp *ContentProcessor) ProcessRegulation(ctx context.Context, discovered DiscoveredRegulation) (*ProcessedRegulation, error) {
	cp.log.Info("Processing regulation", "title", discovered.Title, "has_pdf", discovered.PDFURL != "", "has_content", discovered.Content != "")

	result := &ProcessedRegulation{
		Title:         discovered.Title,
		SourceURL:     discovered.SourceURL,
		PDFURL:        discovered.PDFURL,
		PublishedDate: discovered.PublishedDate,
		Category:      discovered.Category,
		Source:        discovered.Source,
	}

	// Skip if no title (bad data)
	if discovered.Title == "" {
		cp.log.Warn("Skipping regulation with no title", "url", discovered.SourceURL)
		return nil, fmt.Errorf("regulation has no title")
	}

	// Try multiple content sources in order of preference
	contentSources := []struct {
		name    string
		content string
	}{
		{"web_content", discovered.Content},
		{"highlights", discovered.Summary},
	}

	// Try web content first (most reliable from Exa.ai)
	for _, source := range contentSources {
		if source.content != "" && len(source.content) > 100 {
			cp.log.Debug("Using content source", "source", source.name, "length", len(source.content))
			result.FullText = source.content
			break
		}
	}

	// Only try PDF if no web content and PDF URL exists
	if result.FullText == "" && discovered.PDFURL != "" {
		cp.log.Debug("Attempting PDF extraction", "url", discovered.PDFURL)
		text, err := cp.processPDF(ctx, discovered.PDFURL)
		if err != nil {
			cp.log.Warn("PDF processing failed", "error", err)
		} else if len(text) > 100 {
			result.FullText = text
			result.WasPDF = true
		}
	}

	// If still no content, create minimal content from title
	if result.FullText == "" {
		cp.log.Warn("No content available, creating from title", "title", discovered.Title)
		result.FullText = fmt.Sprintf("Regulasi: %s\n\nSumber: %s\nKategori: %s",
			discovered.Title, discovered.SourceURL, discovered.Category)
	}

	// Generate AI summary if we have meaningful content
	if len(result.FullText) > 100 {
		summary, err := cp.summarizeWithAI(ctx, result.Title, result.FullText, result.Category)
		if err != nil {
			cp.log.Warn("AI summarization failed", "error", err)
			// Use first 500 chars as fallback summary
			if len(result.FullText) > 500 {
				result.Summary = result.FullText[:500] + "..."
			} else {
				result.Summary = result.FullText
			}
		} else {
			result.Summary = summary
		}
	} else {
		// For minimal content, use title as summary
		result.Summary = discovered.Title
	}

	return result, nil
}

// processPDF downloads, extracts text, and cleans up PDF
func (cp *ContentProcessor) processPDF(ctx context.Context, pdfURL string) (string, error) {
	if cp.extractor == nil {
		return "", fmt.Errorf("PDF extractor not available")
	}

	// Extract PDF (handles download, OCR if needed, cleanup)
	extracted, err := cp.extractor.ExtractPDF(ctx, pdfURL)
	if err != nil {
		return "", fmt.Errorf("PDF extraction failed: %w", err)
	}

	// Clean the text
	cleanedText := cp.extractor.CleanText(extracted.Text)

	cp.log.Info("PDF processed", "pages", extracted.PageCount, "was_scanned", extracted.IsScanned, "text_length", len(cleanedText))

	return cleanedText, nil
}

// summarizeWithAI generates an UMKM-focused summary of the regulation
func (cp *ContentProcessor) summarizeWithAI(ctx context.Context, title, content, category string) (string, error) {
	if cp.chatProvider == nil {
		return "", fmt.Errorf("chat provider not available")
	}

	// Truncate content if too long
	maxContentLength := 6000
	if len(content) > maxContentLength {
		content = content[:maxContentLength] + "..."
	}

	prompt := fmt.Sprintf(`Kamu adalah ahli regulasi untuk UMKM Indonesia. Buatkan ringkasan regulasi berikut yang fokus pada dampak dan kewajiban untuk pelaku UMKM.

JUDUL: %s
KATEGORI: %s

ISI REGULASI:
%s

INSTRUKSI:
1. Buat ringkasan dalam 3-5 paragraf dalam Bahasa Indonesia
2. Fokus pada:
   - Apa yang diatur (ringkas)
   - Siapa yang terkena dampak (khususnya UMKM)
   - Kewajiban/persyaratan yang harus dipenuhi
   - Sanksi jika ada
   - Kapan berlaku
3. Gunakan bahasa yang mudah dipahami pelaku UMKM
4. Jika ada nominal/angka penting, sebutkan dengan jelas

RINGKASAN:`, title, category, content)

	resp, err := cp.chatProvider.CreateChatCompletion(ctx, chat.ChatCompletionRequest{
		Model: cp.chatModel,
		Messages: []chat.ChatCompletionMessage{
			{Role: "user", Content: prompt},
		},
		MaxTokens:   1000,
		Temperature: 0.5,
	})
	if err != nil {
		return "", err
	}

	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("no response from AI")
	}

	return strings.TrimSpace(resp.Choices[0].Message.Content), nil
}

// ProcessBatch processes multiple regulations
func (cp *ContentProcessor) ProcessBatch(ctx context.Context, regulations []DiscoveredRegulation) []*ProcessedRegulation {
	var results []*ProcessedRegulation

	for i, reg := range regulations {
		cp.log.Info("Processing batch item", "index", i+1, "total", len(regulations), "title", reg.Title)

		processed, err := cp.ProcessRegulation(ctx, reg)
		if err != nil {
			cp.log.Warn("Failed to process regulation", "title", reg.Title, "error", err)
			continue
		}

		results = append(results, processed)
	}

	return results
}
