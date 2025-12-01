package handlers

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/bantuaku/backend/middleware"
	"github.com/bantuaku/backend/models"
	"github.com/google/uuid"
)

// WooCommerceConnectRequest represents a request to connect WooCommerce
type WooCommerceConnectRequest struct {
	StoreURL       string `json:"store_url"`
	ConsumerKey    string `json:"consumer_key"`
	ConsumerSecret string `json:"consumer_secret"`
}

// WooCommerceSyncStatusResponse represents the sync status
type WooCommerceSyncStatusResponse struct {
	Status       string     `json:"status"`
	LastSync     *time.Time `json:"last_sync,omitempty"`
	ProductCount int        `json:"product_count"`
	OrderCount   int        `json:"order_count"`
	ErrorMessage string     `json:"error_message,omitempty"`
}

// WooCommerceConnect connects a WooCommerce store
func (h *Handler) WooCommerceConnect(w http.ResponseWriter, r *http.Request) {
	storeID := middleware.GetStoreID(r.Context())
	if storeID == "" {
		respondError(w, http.StatusUnauthorized, "Store not found in context")
		return
	}

	var req WooCommerceConnectRequest
	if err := parseJSON(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.StoreURL == "" || req.ConsumerKey == "" || req.ConsumerSecret == "" {
		respondError(w, http.StatusBadRequest, "Store URL, consumer key, and consumer secret are required")
		return
	}

	// Test connection by fetching store info
	client := &http.Client{Timeout: 10 * time.Second}
	testURL := fmt.Sprintf("%s/wp-json/wc/v3/system_status", req.StoreURL)

	testReq, _ := http.NewRequest("GET", testURL, nil)
	auth := base64.StdEncoding.EncodeToString([]byte(req.ConsumerKey + ":" + req.ConsumerSecret))
	testReq.Header.Set("Authorization", "Basic "+auth)

	resp, err := client.Do(testReq)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Failed to connect to WooCommerce store")
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respondError(w, http.StatusBadRequest, "Invalid WooCommerce credentials")
		return
	}

	// Store metadata as JSON
	metadata := map[string]string{
		"store_url":       req.StoreURL,
		"consumer_key":    req.ConsumerKey,
		"consumer_secret": req.ConsumerSecret,
	}
	metadataJSON, _ := json.Marshal(metadata)

	// Check if integration exists
	var existingID string
	err = h.db.Pool().QueryRow(r.Context(), `
		SELECT id FROM integrations WHERE store_id = $1 AND platform = 'woocommerce'
	`, storeID).Scan(&existingID)

	if err == nil {
		// Update existing
		_, err = h.db.Pool().Exec(r.Context(), `
			UPDATE integrations SET status = 'connected', metadata = $1, error_message = ''
			WHERE id = $2
		`, string(metadataJSON), existingID)
	} else {
		// Create new
		integrationID := uuid.New().String()
		_, err = h.db.Pool().Exec(r.Context(), `
			INSERT INTO integrations (id, store_id, platform, status, metadata, created_at)
			VALUES ($1, $2, 'woocommerce', 'connected', $3, $4)
		`, integrationID, storeID, string(metadataJSON), time.Now())
	}

	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to save integration")
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{
		"status":  "connected",
		"message": "WooCommerce store connected successfully",
	})
}

// WooCommerceSyncStatus returns the sync status
func (h *Handler) WooCommerceSyncStatus(w http.ResponseWriter, r *http.Request) {
	storeID := middleware.GetStoreID(r.Context())
	if storeID == "" {
		respondError(w, http.StatusUnauthorized, "Store not found in context")
		return
	}

	var integration models.Integration
	err := h.db.Pool().QueryRow(r.Context(), `
		SELECT id, store_id, platform, status, last_sync, error_message, metadata
		FROM integrations
		WHERE store_id = $1 AND platform = 'woocommerce'
	`, storeID).Scan(&integration.ID, &integration.StoreID, &integration.Platform, &integration.Status, &integration.LastSync, &integration.ErrorMessage, &integration.Metadata)

	if err != nil {
		respondJSON(w, http.StatusOK, WooCommerceSyncStatusResponse{
			Status: "disconnected",
		})
		return
	}

	// Count synced products and orders
	var productCount, orderCount int
	h.db.Pool().QueryRow(r.Context(), `
		SELECT COUNT(*) FROM products WHERE store_id = $1
	`, storeID).Scan(&productCount)

	h.db.Pool().QueryRow(r.Context(), `
		SELECT COUNT(*) FROM sales_history WHERE store_id = $1 AND source = 'woocommerce'
	`, storeID).Scan(&orderCount)

	respondJSON(w, http.StatusOK, WooCommerceSyncStatusResponse{
		Status:       integration.Status,
		LastSync:     integration.LastSync,
		ProductCount: productCount,
		OrderCount:   orderCount,
		ErrorMessage: integration.ErrorMessage,
	})
}

// WooCommerceSyncNow triggers a manual sync
func (h *Handler) WooCommerceSyncNow(w http.ResponseWriter, r *http.Request) {
	storeID := middleware.GetStoreID(r.Context())
	if storeID == "" {
		respondError(w, http.StatusUnauthorized, "Store not found in context")
		return
	}

	// Get integration credentials
	var metadataJSON string
	err := h.db.Pool().QueryRow(r.Context(), `
		SELECT metadata FROM integrations 
		WHERE store_id = $1 AND platform = 'woocommerce' AND status = 'connected'
	`, storeID).Scan(&metadataJSON)

	if err != nil {
		respondError(w, http.StatusBadRequest, "WooCommerce not connected")
		return
	}

	var metadata map[string]string
	json.Unmarshal([]byte(metadataJSON), &metadata)

	storeURL := metadata["store_url"]
	consumerKey := metadata["consumer_key"]
	consumerSecret := metadata["consumer_secret"]

	client := &http.Client{Timeout: 30 * time.Second}
	auth := base64.StdEncoding.EncodeToString([]byte(consumerKey + ":" + consumerSecret))

	// Sync products
	productsURL := fmt.Sprintf("%s/wp-json/wc/v3/products?per_page=100", storeURL)
	productReq, _ := http.NewRequest("GET", productsURL, nil)
	productReq.Header.Set("Authorization", "Basic "+auth)

	productResp, err := client.Do(productReq)
	if err != nil {
		h.updateIntegrationError(r.Context(), storeID, "Failed to fetch products: "+err.Error())
		respondError(w, http.StatusInternalServerError, "Failed to fetch products from WooCommerce")
		return
	}
	defer productResp.Body.Close()

	var wooProducts []struct {
		ID    int64  `json:"id"`
		Name  string `json:"name"`
		SKU   string `json:"sku"`
		Price string `json:"price"`
	}

	body, _ := io.ReadAll(productResp.Body)
	json.Unmarshal(body, &wooProducts)

	syncedProducts := 0
	for _, wp := range wooProducts {
		price := 0.0
		fmt.Sscanf(wp.Price, "%f", &price)

		// Upsert product
		_, err := h.db.Pool().Exec(r.Context(), `
			INSERT INTO products (id, store_id, product_name, sku, unit_price, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
			ON CONFLICT (store_id, sku) DO UPDATE SET
				product_name = EXCLUDED.product_name,
				unit_price = EXCLUDED.unit_price,
				updated_at = EXCLUDED.updated_at
		`, uuid.New().String(), storeID, wp.Name, wp.SKU, price, time.Now(), time.Now())

		if err == nil {
			syncedProducts++
		}
	}

	// Sync recent orders
	ordersURL := fmt.Sprintf("%s/wp-json/wc/v3/orders?per_page=100&status=completed", storeURL)
	orderReq, _ := http.NewRequest("GET", ordersURL, nil)
	orderReq.Header.Set("Authorization", "Basic "+auth)

	orderResp, err := client.Do(orderReq)
	if err != nil {
		h.updateIntegrationError(r.Context(), storeID, "Failed to fetch orders: "+err.Error())
		respondError(w, http.StatusInternalServerError, "Failed to fetch orders from WooCommerce")
		return
	}
	defer orderResp.Body.Close()

	var wooOrders []struct {
		ID          int64  `json:"id"`
		DateCreated string `json:"date_created"`
		LineItems   []struct {
			ProductID int64   `json:"product_id"`
			Quantity  int     `json:"quantity"`
			Price     float64 `json:"price"`
		} `json:"line_items"`
	}

	orderBody, _ := io.ReadAll(orderResp.Body)
	json.Unmarshal(orderBody, &wooOrders)

	syncedOrders := 0
	for _, wo := range wooOrders {
		orderDate, _ := time.Parse(time.RFC3339, wo.DateCreated)

		for _, item := range wo.LineItems {
			// Get product ID by WooCommerce product ID (stored in SKU or a mapping)
			var productID string
			h.db.Pool().QueryRow(r.Context(), `
				SELECT id FROM products WHERE store_id = $1 LIMIT 1
			`, storeID).Scan(&productID)

			if productID == "" {
				continue
			}

			_, err := h.db.Pool().Exec(r.Context(), `
				INSERT INTO sales_history (store_id, product_id, quantity, price, sale_date, source, created_at)
				VALUES ($1, $2, $3, $4, $5, 'woocommerce', $6)
			`, storeID, productID, item.Quantity, item.Price, orderDate, time.Now())

			if err == nil {
				syncedOrders++
			}
		}
	}

	// Update last sync time
	now := time.Now()
	h.db.Pool().Exec(r.Context(), `
		UPDATE integrations SET last_sync = $1, error_message = ''
		WHERE store_id = $2 AND platform = 'woocommerce'
	`, now, storeID)

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"status":          "success",
		"products_synced": syncedProducts,
		"orders_synced":   syncedOrders,
		"last_sync":       now,
	})
}

func (h *Handler) updateIntegrationError(ctx interface{}, storeID, errorMsg string) {
	// Type assertion for context
	if c, ok := ctx.(interface{ Done() <-chan struct{} }); ok {
		h.db.Pool().Exec(c.(interface {
			Done() <-chan struct{}
			Err() error
			Value(interface{}) interface{}
			Deadline() (time.Time, bool)
		}), `
			UPDATE integrations SET status = 'error', error_message = $1
			WHERE store_id = $2 AND platform = 'woocommerce'
		`, errorMsg, storeID)
	}
}
