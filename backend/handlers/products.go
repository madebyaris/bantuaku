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
	Stock       int     `json:"stock"`
}

// UpdateProductRequest represents a request to update a product
type UpdateProductRequest struct {
	ProductName string  `json:"product_name,omitempty"`
	SKU         string  `json:"sku,omitempty"`
	Category    string  `json:"category,omitempty"`
	UnitPrice   float64 `json:"unit_price,omitempty"`
	Cost        float64 `json:"cost,omitempty"`
	Stock       int     `json:"stock,omitempty"`
}

// ListProducts returns all products for the authenticated store
func (h *Handler) ListProducts(w http.ResponseWriter, r *http.Request) {
	storeID := middleware.GetStoreID(r.Context())
	if storeID == "" {
		respondError(w, http.StatusUnauthorized, "Store not found in context")
		return
	}

	category := r.URL.Query().Get("category")

	var query string
	var args []interface{}

	if category != "" {
		query = `
			SELECT id, store_id, product_name, sku, category, unit_price, cost, stock, created_at, updated_at
			FROM products 
			WHERE store_id = $1 AND category = $2
			ORDER BY product_name
		`
		args = []interface{}{storeID, category}
	} else {
		query = `
			SELECT id, store_id, product_name, sku, category, unit_price, cost, stock, created_at, updated_at
			FROM products 
			WHERE store_id = $1
			ORDER BY product_name
		`
		args = []interface{}{storeID}
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
		err := rows.Scan(&p.ID, &p.StoreID, &p.ProductName, &p.SKU, &p.Category, &p.UnitPrice, &p.Cost, &p.Stock, &p.CreatedAt, &p.UpdatedAt)
		if err != nil {
			continue
		}
		products = append(products, p)
	}

	respondJSON(w, http.StatusOK, products)
}

// CreateProduct creates a new product
func (h *Handler) CreateProduct(w http.ResponseWriter, r *http.Request) {
	storeID := middleware.GetStoreID(r.Context())
	if storeID == "" {
		respondError(w, http.StatusUnauthorized, "Store not found in context")
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
		INSERT INTO products (id, store_id, product_name, sku, category, unit_price, cost, stock, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`, productID, storeID, req.ProductName, req.SKU, req.Category, req.UnitPrice, req.Cost, req.Stock, now, now)

	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to create product")
		return
	}

	product := models.Product{
		ID:          productID,
		StoreID:     storeID,
		ProductName: req.ProductName,
		SKU:         req.SKU,
		Category:    req.Category,
		UnitPrice:   req.UnitPrice,
		Cost:        req.Cost,
		Stock:       req.Stock,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	respondJSON(w, http.StatusCreated, product)
}

// GetProduct returns a single product by ID
func (h *Handler) GetProduct(w http.ResponseWriter, r *http.Request) {
	storeID := middleware.GetStoreID(r.Context())
	productID := r.PathValue("id")

	if productID == "" {
		respondError(w, http.StatusBadRequest, "Product ID is required")
		return
	}

	var p models.Product
	err := h.db.Pool().QueryRow(r.Context(), `
		SELECT id, store_id, product_name, sku, category, unit_price, cost, stock, created_at, updated_at
		FROM products
		WHERE id = $1 AND store_id = $2
	`, productID, storeID).Scan(&p.ID, &p.StoreID, &p.ProductName, &p.SKU, &p.Category, &p.UnitPrice, &p.Cost, &p.Stock, &p.CreatedAt, &p.UpdatedAt)

	if err != nil {
		respondError(w, http.StatusNotFound, "Product not found")
		return
	}

	respondJSON(w, http.StatusOK, p)
}

// UpdateProduct updates an existing product
func (h *Handler) UpdateProduct(w http.ResponseWriter, r *http.Request) {
	storeID := middleware.GetStoreID(r.Context())
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
		SET product_name = COALESCE(NULLIF($3, ''), product_name),
			sku = COALESCE(NULLIF($4, ''), sku),
			category = COALESCE(NULLIF($5, ''), category),
			unit_price = CASE WHEN $6 > 0 THEN $6 ELSE unit_price END,
			cost = CASE WHEN $7 > 0 THEN $7 ELSE cost END,
			stock = CASE WHEN $8 >= 0 THEN $8 ELSE stock END,
			updated_at = $9
		WHERE id = $1 AND store_id = $2
	`, productID, storeID, req.ProductName, req.SKU, req.Category, req.UnitPrice, req.Cost, req.Stock, time.Now())

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
		SELECT id, store_id, product_name, sku, category, unit_price, cost, stock, created_at, updated_at
		FROM products WHERE id = $1
	`, productID).Scan(&p.ID, &p.StoreID, &p.ProductName, &p.SKU, &p.Category, &p.UnitPrice, &p.Cost, &p.Stock, &p.CreatedAt, &p.UpdatedAt)

	respondJSON(w, http.StatusOK, p)
}

// DeleteProduct deletes a product
func (h *Handler) DeleteProduct(w http.ResponseWriter, r *http.Request) {
	storeID := middleware.GetStoreID(r.Context())
	productID := r.PathValue("id")

	if productID == "" {
		respondError(w, http.StatusBadRequest, "Product ID is required")
		return
	}

	result, err := h.db.Pool().Exec(r.Context(), `
		DELETE FROM products WHERE id = $1 AND store_id = $2
	`, productID, storeID)

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
