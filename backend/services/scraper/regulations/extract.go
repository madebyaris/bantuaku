package regulations

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
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
		text, pageCount, err = e.extractWithOCR(ctx, tempFile)
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

// extractWithOCR extracts text using OCR via Kolosal.ai
// Converts PDF pages to images, then OCRs each image
// Returns extracted text and page count
func (e *Extractor) extractWithOCR(ctx context.Context, pdfPath string) (string, int, error) {
	if e.kolosal == nil {
		return "", 0, fmt.Errorf("Kolosal client not available")
	}

	// Convert PDF pages to images
	imageFiles, err := e.convertPDFToImages(pdfPath)
	if err != nil {
		return "", 0, fmt.Errorf("failed to convert PDF to images: %w", err)
	}
	defer func() {
		// Clean up image files and directory
		for _, imgFile := range imageFiles {
			os.Remove(imgFile)
		}
		// Also remove the temp directory if it exists
		if len(imageFiles) > 0 {
			imageDir := filepath.Dir(imageFiles[0])
			os.RemoveAll(imageDir)
		}
	}()

	pageCount := len(imageFiles)
	if pageCount == 0 {
		return "", 0, fmt.Errorf("no images generated from PDF")
	}

	// OCR each image using Kolosal
	var allText strings.Builder
	for i, imgFile := range imageFiles {
		e.log.Info("OCR processing page", "page", i+1, "total", pageCount)

		// Read image file
		imageBytes, err := os.ReadFile(imgFile)
		if err != nil {
			e.log.Warn("Failed to read image file", "file", imgFile, "error", err)
			continue
		}

		// Encode to base64
		imageBase64 := base64.StdEncoding.EncodeToString(imageBytes)

		// Call Kolosal OCR
		ocrResp, err := e.kolosal.OCR(ctx, kolosal.OCRRequest{
			Image:    imageBase64,
			Language: "id", // Indonesian
		})
		if err != nil {
			e.log.Warn("OCR failed for page", "page", i+1, "error", err)
			continue
		}

		if ocrResp.Text != "" {
			allText.WriteString(ocrResp.Text)
			if i < len(imageFiles)-1 {
				allText.WriteString("\n\n") // Page separator
			}
		}
	}

	if allText.Len() == 0 {
		return "", pageCount, fmt.Errorf("no text extracted from PDF via OCR")
	}

	return allText.String(), pageCount, nil
}

// convertPDFToImages converts PDF pages to PNG images using pdftoppm (Poppler)
// Returns list of image file paths
func (e *Extractor) convertPDFToImages(pdfPath string) ([]string, error) {
	// Check if pdftoppm is available
	if _, err := exec.LookPath("pdftoppm"); err != nil {
		return nil, fmt.Errorf("pdftoppm not found - Poppler utils must be installed")
	}

	// Create temp directory for images
	imageDir := filepath.Join(e.tempDir, fmt.Sprintf("pdf_images_%d", time.Now().UnixNano()))
	if err := os.MkdirAll(imageDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create temp directory: %w", err)
	}

	// Output pattern for pdftoppm
	outputPattern := filepath.Join(imageDir, "page")

	// Run pdftoppm to convert PDF to PNG images
	cmd := exec.Command("pdftoppm", "-png", "-r", "300", pdfPath, outputPattern)
	cmd.Dir = imageDir
	if err := cmd.Run(); err != nil {
		os.RemoveAll(imageDir)
		return nil, fmt.Errorf("pdftoppm failed: %w", err)
	}

	// Find all generated PNG files
	pattern := outputPattern + "-*.png"
	matches, err := filepath.Glob(pattern)
	if err != nil {
		os.RemoveAll(imageDir)
		return nil, fmt.Errorf("failed to find generated images: %w", err)
	}

	if len(matches) == 0 {
		os.RemoveAll(imageDir)
		return nil, fmt.Errorf("no images generated from PDF")
	}

	// Sort files by page number (pdftoppm generates page-01.png, page-02.png, etc.)
	// Simple sort: just return matches as-is (they should be in order)
	return matches, nil
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
