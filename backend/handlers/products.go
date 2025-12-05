package handlers

import (
	"net/http"
	"time"

	"github.com/bantuaku/backend/middleware"
	"github.com/bantuaku/backend/models"
	"github.com/google/uuid"
)

// CreateProductRequest represents a request to create a product
type CreateProductRequest struct {
	ProductName string  `json:"product_name"`
	SKU         string  `json:"sku,omitempty"`
	Category    string  `json:"category,omitempty"`
	UnitPrice   float64 `json:"unit_price"`
	Cost        float64 `json:"cost,omitempty"`
}

// UpdateProductRequest represents a request to update a product
type UpdateProductRequest struct {
	ProductName string  `json:"product_name,omitempty"`
	SKU         string  `json:"sku,omitempty"`
	Category    string  `json:"category,omitempty"`
	UnitPrice   float64 `json:"unit_price,omitempty"`
	Cost        float64 `json:"cost,omitempty"`
}

// ListProducts returns all products for the authenticated company
func (h *Handler) ListProducts(w http.ResponseWriter, r *http.Request) {
	companyID := middleware.GetCompanyID(r.Context())
	if companyID == "" {
		respondError(w, http.StatusUnauthorized, "Company not found in context")
		return
	}

	category := r.URL.Query().Get("category")

	var query string
	var args []interface{}

	if category != "" {
		query = `
			SELECT id, company_id, name, sku, category, unit_price, cost, created_at, updated_at
			FROM products 
			WHERE company_id = $1 AND category = $2
			ORDER BY name
		`
		args = []interface{}{companyID, category}
	} else {
		query = `
			SELECT id, company_id, name, sku, category, unit_price, cost, created_at, updated_at
			FROM products 
			WHERE company_id = $1
			ORDER BY name
		`
		args = []interface{}{companyID}
	}

	rows, err := h.db.Pool().Query(r.Context(), query, args...)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to fetch products")
		return
	}
	defer rows.Close()

	products := []models.Product{}
	for rows.Next() {
		var p models.Product
		err := rows.Scan(&p.ID, &p.CompanyID, &p.Name, &p.SKU, &p.Category, &p.UnitPrice, &p.Cost, &p.CreatedAt, &p.UpdatedAt)
		if err != nil {
			continue
		}
		products = append(products, p)
	}

	respondJSON(w, http.StatusOK, products)
}

// CreateProduct creates a new product
func (h *Handler) CreateProduct(w http.ResponseWriter, r *http.Request) {
	companyID := middleware.GetCompanyID(r.Context())
	if companyID == "" {
		respondError(w, http.StatusUnauthorized, "Company not found in context")
		return
	}

	var req CreateProductRequest
	if err := parseJSON(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.ProductName == "" {
		respondError(w, http.StatusBadRequest, "Product name is required")
		return
	}
	if req.UnitPrice < 0 {
		respondError(w, http.StatusBadRequest, "Unit price cannot be negative")
		return
	}

	productID := uuid.New().String()
	now := time.Now()

	_, err := h.db.Pool().Exec(r.Context(), `
		INSERT INTO products (id, company_id, name, sku, category, unit_price, cost, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`, productID, companyID, req.ProductName, req.SKU, req.Category, req.UnitPrice, req.Cost, now, now)

	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to create product")
		return
	}

	product := models.Product{
		ID:        productID,
		CompanyID: companyID,
		Name:      req.ProductName,
		SKU:       req.SKU,
		Category:  req.Category,
		UnitPrice: req.UnitPrice,
		Cost:      req.Cost,
		CreatedAt: now,
		UpdatedAt: now,
	}

	respondJSON(w, http.StatusCreated, product)
}

// GetProduct returns a single product by ID
func (h *Handler) GetProduct(w http.ResponseWriter, r *http.Request) {
	companyID := middleware.GetCompanyID(r.Context())
	productID := r.PathValue("id")

	if productID == "" {
		respondError(w, http.StatusBadRequest, "Product ID is required")
		return
	}

	var p models.Product
	err := h.db.Pool().QueryRow(r.Context(), `
		SELECT id, company_id, name, sku, category, unit_price, cost, created_at, updated_at
		FROM products
		WHERE id = $1 AND company_id = $2
	`, productID, companyID).Scan(&p.ID, &p.CompanyID, &p.Name, &p.SKU, &p.Category, &p.UnitPrice, &p.Cost, &p.CreatedAt, &p.UpdatedAt)

	if err != nil {
		respondError(w, http.StatusNotFound, "Product not found")
		return
	}

	respondJSON(w, http.StatusOK, p)
}

// UpdateProduct updates an existing product
func (h *Handler) UpdateProduct(w http.ResponseWriter, r *http.Request) {
	companyID := middleware.GetCompanyID(r.Context())
	productID := r.PathValue("id")

	if productID == "" {
		respondError(w, http.StatusBadRequest, "Product ID is required")
		return
	}

	var req UpdateProductRequest
	if err := parseJSON(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Build dynamic update query based on provided fields
	result, err := h.db.Pool().Exec(r.Context(), `
		UPDATE products 
		SET name = COALESCE(NULLIF($3, ''), name),
			sku = COALESCE(NULLIF($4, ''), sku),
			category = COALESCE(NULLIF($5, ''), category),
			unit_price = CASE WHEN $6 > 0 THEN $6 ELSE unit_price END,
			cost = CASE WHEN $7 > 0 THEN $7 ELSE cost END,
			updated_at = $8
		WHERE id = $1 AND company_id = $2
	`, productID, companyID, req.ProductName, req.SKU, req.Category, req.UnitPrice, req.Cost, time.Now())

	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to update product")
		return
	}

	if result.RowsAffected() == 0 {
		respondError(w, http.StatusNotFound, "Product not found")
		return
	}

	// Fetch and return updated product
	var p models.Product
	h.db.Pool().QueryRow(r.Context(), `
		SELECT id, company_id, name, sku, category, unit_price, cost, created_at, updated_at
		FROM products WHERE id = $1
	`, productID).Scan(&p.ID, &p.CompanyID, &p.Name, &p.SKU, &p.Category, &p.UnitPrice, &p.Cost, &p.CreatedAt, &p.UpdatedAt)

	respondJSON(w, http.StatusOK, p)
}

// DeleteProduct deletes a product
func (h *Handler) DeleteProduct(w http.ResponseWriter, r *http.Request) {
	companyID := middleware.GetCompanyID(r.Context())
	productID := r.PathValue("id")

	if productID == "" {
		respondError(w, http.StatusBadRequest, "Product ID is required")
		return
	}

	result, err := h.db.Pool().Exec(r.Context(), `
		DELETE FROM products WHERE id = $1 AND company_id = $2
	`, productID, companyID)

	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to delete product")
		return
	}

	if result.RowsAffected() == 0 {
		respondError(w, http.StatusNotFound, "Product not found")
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"message": "Product deleted"})
}
