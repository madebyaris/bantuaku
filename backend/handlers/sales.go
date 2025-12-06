package handlers

import (
	"encoding/csv"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/bantuaku/backend/middleware"
	"github.com/bantuaku/backend/models"
)

// RecordSaleRequest represents a manual sale entry
type RecordSaleRequest struct {
	ProductID string    `json:"product_id"`
	Quantity  int       `json:"quantity"`
	Price     float64   `json:"price"`
	SaleDate  time.Time `json:"sale_date"`
}

// ImportResult represents the result of a CSV import
type ImportResult struct {
	SuccessCount int           `json:"success_count"`
	Errors       []ImportError `json:"errors"`
}

// ImportError represents an error during import
type ImportError struct {
	Row   int    `json:"row"`
	Error string `json:"error"`
}

// RecordSale records a single manual sale
func (h *Handler) RecordSale(w http.ResponseWriter, r *http.Request) {
	companyID := middleware.GetCompanyID(r.Context())
	if companyID == "" {
		respondError(w, http.StatusUnauthorized, "Company not found in context")
		return
	}

	var req RecordSaleRequest
	if err := parseJSON(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate input
	if req.ProductID == "" {
		respondError(w, http.StatusBadRequest, "Product ID is required")
		return
	}
	if req.Quantity <= 0 {
		respondError(w, http.StatusBadRequest, "Quantity must be greater than 0")
		return
	}
	if req.Price < 0 {
		respondError(w, http.StatusBadRequest, "Price cannot be negative")
		return
	}
	if req.SaleDate.IsZero() {
		req.SaleDate = time.Now()
	}

	// Verify product belongs to store
	var productExists bool
	err := h.db.Pool().QueryRow(r.Context(), `
		SELECT EXISTS(SELECT 1 FROM products WHERE id = $1 AND company_id = $2)
	`, req.ProductID, companyID).Scan(&productExists)
	if err != nil || !productExists {
		respondError(w, http.StatusBadRequest, "Product not found")
		return
	}

	// Insert sale record
	var saleID int64
	err = h.db.Pool().QueryRow(r.Context(), `
		INSERT INTO sales_history (company_id, product_id, quantity, price, sale_date, source, created_at)
		VALUES ($1, $2, $3, $4, $5, 'manual', $6)
		RETURNING id
	`, companyID, req.ProductID, req.Quantity, req.Price, req.SaleDate, time.Now()).Scan(&saleID)

	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to record sale")
		return
	}

	// Invalidate forecast cache for this product
	cacheKey := fmt.Sprintf("forecast:%s", req.ProductID)
	h.redis.Delete(r.Context(), cacheKey)

	respondJSON(w, http.StatusCreated, models.Sale{
		ID:        saleID,
		CompanyID: companyID,
		ProductID: req.ProductID,
		Quantity:  req.Quantity,
		Price:     req.Price,
		SaleDate:  req.SaleDate,
		Source:    "manual",
		CreatedAt: time.Now(),
	})
}

// ImportCSV handles bulk CSV sales import
func (h *Handler) ImportCSV(w http.ResponseWriter, r *http.Request) {
	companyID := middleware.GetCompanyID(r.Context())
	if companyID == "" {
		respondError(w, http.StatusUnauthorized, "Company not found in context")
		return
	}

	// Parse multipart form (max 10MB)
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		respondError(w, http.StatusBadRequest, "Failed to parse form data")
		return
	}

	file, _, err := r.FormFile("file")
	if err != nil {
		respondError(w, http.StatusBadRequest, "File is required")
		return
	}
	defer file.Close()

	// Parse CSV
	reader := csv.NewReader(file)

	// Read header
	header, err := reader.Read()
	if err != nil {
		respondError(w, http.StatusBadRequest, "Failed to read CSV header")
		return
	}

	// Map column indices
	colMap := make(map[string]int)
	for i, col := range header {
		colMap[col] = i
	}

	// Verify required columns
	requiredCols := []string{"product_name", "quantity", "sale_date"}
	for _, col := range requiredCols {
		if _, exists := colMap[col]; !exists {
			respondError(w, http.StatusBadRequest, fmt.Sprintf("Missing required column: %s", col))
			return
		}
	}

	// Build product name to ID mapping
	productMap := make(map[string]string)
	rows, err := h.db.Pool().Query(r.Context(), `
		SELECT id, name FROM products WHERE company_id = $1
	`, companyID)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var id, name string
			if rows.Scan(&id, &name) == nil {
				productMap[name] = id
			}
		}
	}

	result := ImportResult{}
	rowNum := 1 // Header is row 0

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		rowNum++

		if err != nil {
			result.Errors = append(result.Errors, ImportError{Row: rowNum, Error: "Failed to parse row"})
			continue
		}

		// Extract values
		productName := record[colMap["product_name"]]
		quantityStr := record[colMap["quantity"]]
		saleDateStr := record[colMap["sale_date"]]

		priceStr := "0"
		if idx, exists := colMap["price"]; exists && idx < len(record) {
			priceStr = record[idx]
		}

		// Find product ID
		productID, exists := productMap[productName]
		if !exists {
			result.Errors = append(result.Errors, ImportError{Row: rowNum, Error: fmt.Sprintf("Product not found: %s", productName)})
			continue
		}

		// Parse quantity
		quantity, err := strconv.Atoi(quantityStr)
		if err != nil || quantity <= 0 {
			result.Errors = append(result.Errors, ImportError{Row: rowNum, Error: "Invalid quantity"})
			continue
		}

		// Parse price
		price, _ := strconv.ParseFloat(priceStr, 64)

		// Parse date (try multiple formats)
		var saleDate time.Time
		dateFormats := []string{"2006-01-02", "02/01/2006", "01/02/2006", "2006/01/02"}
		for _, format := range dateFormats {
			if parsed, err := time.Parse(format, saleDateStr); err == nil {
				saleDate = parsed
				break
			}
		}
		if saleDate.IsZero() {
			result.Errors = append(result.Errors, ImportError{Row: rowNum, Error: "Invalid date format"})
			continue
		}

		// Insert sale
		_, err = h.db.Pool().Exec(r.Context(), `
			INSERT INTO sales_history (company_id, product_id, quantity, price, sale_date, source, created_at)
			VALUES ($1, $2, $3, $4, $5, 'csv', $6)
		`, companyID, productID, quantity, price, saleDate, time.Now())

		if err != nil {
			result.Errors = append(result.Errors, ImportError{Row: rowNum, Error: "Database error"})
			continue
		}

		result.SuccessCount++

		// Invalidate forecast cache for this product
		cacheKey := fmt.Sprintf("forecast:%s", productID)
		h.redis.Delete(r.Context(), cacheKey)
	}

	respondJSON(w, http.StatusOK, result)
}

// ListSales returns sales history for the store
func (h *Handler) ListSales(w http.ResponseWriter, r *http.Request) {
	companyID := middleware.GetCompanyID(r.Context())
	if companyID == "" {
		respondError(w, http.StatusUnauthorized, "Company not found in context")
		return
	}

	productID := r.URL.Query().Get("product_id")
	limit := 100

	var query string
	var args []interface{}

	if productID != "" {
		query = `
			SELECT s.id, s.company_id, s.product_id, s.quantity, s.price, s.sale_date, s.source, s.created_at
			FROM sales_history s
			WHERE s.company_id = $1 AND s.product_id = $2
			ORDER BY s.sale_date DESC
			LIMIT $3
		`
		args = []interface{}{companyID, productID, limit}
	} else {
		query = `
			SELECT s.id, s.company_id, s.product_id, s.quantity, s.price, s.sale_date, s.source, s.created_at
			FROM sales_history s
			WHERE s.company_id = $1
			ORDER BY s.sale_date DESC
			LIMIT $2
		`
		args = []interface{}{companyID, limit}
	}

	rows, err := h.db.Pool().Query(r.Context(), query, args...)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to fetch sales")
		return
	}
	defer rows.Close()

	sales := []models.Sale{}
	for rows.Next() {
		var s models.Sale
		err := rows.Scan(&s.ID, &s.CompanyID, &s.ProductID, &s.Quantity, &s.Price, &s.SaleDate, &s.Source, &s.CreatedAt)
		if err != nil {
			continue
		}
		sales = append(sales, s)
	}

	respondJSON(w, http.StatusOK, sales)
}
