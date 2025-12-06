package handlers

import (
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/bantuaku/backend/errors"
	"github.com/bantuaku/backend/logger"
	"github.com/bantuaku/backend/middleware"
	"github.com/bantuaku/backend/models"
	"github.com/bantuaku/backend/services/kolosal"
	"github.com/bantuaku/backend/services/usage"

	"github.com/google/uuid"
)

const (
	maxFileSizeDefault = 10 * 1024 * 1024 // 10MB default
	uploadDir          = "./uploads"
)

// UploadFileResponse represents the response when uploading a file
type UploadFileResponse struct {
	FileUploadID     string                `json:"file_upload_id"`
	OriginalFilename string                `json:"original_filename"`
	MimeType         string                `json:"mime_type"`
	SizeBytes        int64                 `json:"size_bytes"`
	Status           string                `json:"status"`
	ExtractedData    *models.ExtractedData `json:"extracted_data,omitempty"`
}

// UploadFile handles file uploads (CSV/XLSX/PDF)
func (h *Handler) UploadFile(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	companyID := middleware.GetCompanyID(ctx)

	// Check upload usage limit
	usageService := usage.NewService(h.db)
	canUpload, limitMsg, err := usageService.CheckUploadLimit(ctx, companyID)
	if err != nil {
		logger.Warn("Failed to check upload limit", "error", err.Error())
		// Continue on error - don't block user
	} else if !canUpload {
		h.respondError(w, errors.NewAppError(errors.ErrCodeForbidden, limitMsg, "upload_limit_exceeded"), r)
		return
	}

	// Parse multipart form with a large limit (will check actual file size later)
	err = r.ParseMultipartForm(100 * 1024 * 1024) // 100MB parse limit
	if err != nil {
		h.respondError(w, fmt.Errorf("failed to parse multipart form: %w", err), r)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		h.respondError(w, fmt.Errorf("failed to get file: %w", err), r)
		return
	}
	defer file.Close()

	// Check file size against plan limit
	canUploadSize, sizeMsg, err := usageService.CheckFileSizeLimit(ctx, companyID, header.Size)
	if err != nil {
		logger.Warn("Failed to check file size limit", "error", err.Error())
		// Continue on error - don't block user, use default limit
		if header.Size > maxFileSizeDefault {
			h.respondError(w, fmt.Errorf("file size exceeds maximum of %d bytes", maxFileSizeDefault), r)
			return
		}
	} else if !canUploadSize {
		h.respondError(w, errors.NewAppError(errors.ErrCodeForbidden, sizeMsg, "file_size_exceeded"), r)
		return
	}

	// Determine source type from file extension
	ext := filepath.Ext(header.Filename)
	sourceType := ""
	switch ext {
	case ".csv":
		sourceType = "csv"
	case ".xlsx", ".xls":
		sourceType = "xlsx"
	case ".pdf":
		sourceType = "pdf"
	default:
		h.respondError(w, fmt.Errorf("unsupported file type: %s", ext), r)
		return
	}

	// Create upload directory if it doesn't exist
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		logger.Error("Failed to create upload directory", "error", err.Error())
		h.respondError(w, fmt.Errorf("failed to create upload directory"), r)
		return
	}

	// Generate unique filename
	fileID := uuid.New().String()
	filename := fmt.Sprintf("%s%s", fileID, ext)
	storagePath := filepath.Join(uploadDir, filename)

	// Save file
	dst, err := os.Create(storagePath)
	if err != nil {
		logger.Error("Failed to create file", "error", err.Error())
		h.respondError(w, fmt.Errorf("failed to save file"), r)
		return
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		logger.Error("Failed to copy file", "error", err.Error())
		h.respondError(w, fmt.Errorf("failed to save file"), r)
		return
	}

	fileUploadID := uuid.New().String()

	response := UploadFileResponse{
		FileUploadID:     fileUploadID,
		OriginalFilename: header.Filename,
		MimeType:         header.Header.Get("Content-Type"),
		SizeBytes:        header.Size,
		Status:           "uploaded",
	}

	// Process file based on type
	if sourceType == "pdf" && h.config.KolosalAPIKey != "" {
		// Use Kolosal.ai OCR for PDF processing
		ctx := r.Context()
		client := kolosal.NewClient(h.config.KolosalAPIKey)

		// Read file from disk and encode to base64
		savedFile, err := os.Open(storagePath)
		if err == nil {
			defer savedFile.Close()
			fileBytes, err := io.ReadAll(savedFile)
			if err == nil {
				imageBase64 := base64.StdEncoding.EncodeToString(fileBytes)
				imageDataURL := "data:application/pdf;base64," + imageBase64

				_, err := client.OCR(ctx, kolosal.OCRRequest{
					ImageData: imageDataURL,
				})

				if err == nil {
					response.Status = "processed"
					// TODO: Parse OCR text to extract structured data (products, sales)
					// For now, just mark as processed
					logger.Info("PDF processed with OCR", "file_id", fileUploadID)
				} else {
					response.Status = "failed"
					logger.Error("OCR processing failed", "error", err.Error())
				}
			}
		}
	} else if sourceType == "csv" || sourceType == "xlsx" {
		// TODO: Implement CSV/XLSX parsing
		// For now, mark as uploaded
		response.Status = "uploaded"
	}

	h.respondJSON(w, http.StatusOK, response)
}

// GetFile retrieves file upload information
func (h *Handler) GetFile(w http.ResponseWriter, r *http.Request) {
	fileID := r.URL.Query().Get("id")
	if fileID == "" {
		h.respondError(w, errors.NewValidationError("file id is required", ""), r)
		return
	}

	// TODO: Implement file retrieval from database
	// For now, return error
	h.respondError(w, fmt.Errorf("file not found"), r)
}
