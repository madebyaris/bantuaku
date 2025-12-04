package regulations

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/bantuaku/backend/logger"
	"github.com/bantuaku/backend/services/kolosal"
)

// Extractor handles PDF text extraction
type Extractor struct {
	httpClient *http.Client
	kolosal    *kolosal.Client
	log        logger.Logger
	tempDir    string
}

// ExtractedText represents extracted text from PDF
type ExtractedText struct {
	Text      string
	PageCount int
	IsScanned bool // True if PDF required OCR
}

// NewExtractor creates a new extractor instance
func NewExtractor(kolosalClient *kolosal.Client) *Extractor {
	tempDir := os.TempDir()
	return &Extractor{
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
		kolosal: kolosalClient,
		log:     *logger.Default(),
		tempDir: tempDir,
	}
}

// ExtractPDF downloads PDF, extracts text, and discards the file
func (e *Extractor) ExtractPDF(ctx context.Context, pdfURL string) (*ExtractedText, error) {
	e.log.Info("Extracting PDF", "url", pdfURL)

	// Download PDF to temp file
	tempFile, err := e.downloadPDF(ctx, pdfURL)
	if err != nil {
		return nil, fmt.Errorf("failed to download PDF: %w", err)
	}
	defer os.Remove(tempFile) // Always clean up

	// Try text extraction first
	text, pageCount, err := e.extractTextFromPDF(tempFile)
	if err != nil {
		e.log.Warn("Text extraction failed, trying OCR", "error", err)

		// Fallback to OCR if text extraction fails (scanned PDF)
		text, err = e.extractWithOCR(ctx, tempFile)
		if err != nil {
			return nil, fmt.Errorf("failed to extract text (both methods failed): %w", err)
		}

		return &ExtractedText{
			Text:      text,
			PageCount: pageCount,
			IsScanned: true,
		}, nil
	}

	return &ExtractedText{
		Text:      text,
		PageCount: pageCount,
		IsScanned: false,
	}, nil
}

// downloadPDF downloads PDF to temporary file
func (e *Extractor) downloadPDF(ctx context.Context, url string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")

	resp, err := e.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to download PDF: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	// Create temp file
	tempFile := filepath.Join(e.tempDir, fmt.Sprintf("regulation_%d.pdf", time.Now().UnixNano()))
	out, err := os.Create(tempFile)
	if err != nil {
		return "", fmt.Errorf("failed to create temp file: %w", err)
	}
	defer out.Close()

	// Copy PDF to temp file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		os.Remove(tempFile)
		return "", fmt.Errorf("failed to write PDF: %w", err)
	}

	return tempFile, nil
}

// extractTextFromPDF extracts text from PDF using Go library
// Note: This is a placeholder - you'll need to add a PDF library like unidoc
func (e *Extractor) extractTextFromPDF(filePath string) (string, int, error) {
	// TODO: Implement PDF text extraction using unidoc or similar library
	// For now, return error to trigger OCR fallback
	
	// Example implementation would be:
	// import "github.com/unidoc/unipdf/v3/extractor"
	// pdfReader, err := model.NewPdfReaderFromFile(filePath)
	// if err != nil {
	//     return "", 0, err
	// }
	// numPages, _ := pdfReader.GetNumPages()
	// var text strings.Builder
	// for i := 1; i <= numPages; i++ {
	//     page, _ := pdfReader.GetPage(i)
	//     extractor, _ := extractor.New(page)
	//     pageText, _ := extractor.ExtractText()
	//     text.WriteString(pageText)
	// }
	// return text.String(), numPages, nil

	return "", 0, fmt.Errorf("text extraction not implemented - using OCR fallback")
}

// extractWithOCR extracts text using OCR (Kolosal.ai or Tesseract)
func (e *Extractor) extractWithOCR(ctx context.Context, pdfPath string) (string, error) {
	// Read PDF file (for future use)
	_, err := os.ReadFile(pdfPath)
	if err != nil {
		return "", fmt.Errorf("failed to read PDF: %w", err)
	}

	// Convert PDF pages to images and OCR each page
	// For now, use Kolosal.ai OCR if available
	// TODO: Implement PDF to image conversion, then OCR each image

	// Placeholder: If Kolosal.ai supports PDF OCR directly
	// For now, return error indicating OCR needs implementation
	return "", fmt.Errorf("OCR extraction not yet implemented - requires PDF to image conversion")
}

// CleanText cleans and normalizes extracted text for Indonesian locale
func (e *Extractor) CleanText(text string) string {
	// Remove excessive whitespace
	text = strings.TrimSpace(text)
	
	// Normalize line breaks
	text = strings.ReplaceAll(text, "\r\n", "\n")
	text = strings.ReplaceAll(text, "\r", "\n")
	
	// Remove multiple consecutive newlines
	for strings.Contains(text, "\n\n\n") {
		text = strings.ReplaceAll(text, "\n\n\n", "\n\n")
	}
	
	// Remove page numbers and headers/footers (simple heuristics)
	lines := strings.Split(text, "\n")
	var cleanedLines []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		// Skip lines that look like page numbers
		if len(line) <= 3 && strings.ContainsAny(line, "0123456789") {
			continue
		}
		if line != "" {
			cleanedLines = append(cleanedLines, line)
		}
	}
	
	return strings.Join(cleanedLines, "\n")
}

