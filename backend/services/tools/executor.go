package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/bantuaku/backend/models"
	"github.com/bantuaku/backend/services/storage"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// Executor handles tool execution for AI function calling
type Executor struct {
	db *storage.Postgres
}

// NewExecutor creates a new tool executor
func NewExecutor(db *storage.Postgres) *Executor {
	return &Executor{db: db}
}

// ExecuteTool executes a tool call and returns the result
func (e *Executor) ExecuteTool(ctx context.Context, companyID string, toolCall models.ToolCall) (*models.ToolResult, error) {
	// Parse function arguments
	var args map[string]interface{}
	if err := json.Unmarshal([]byte(toolCall.Function.Arguments), &args); err != nil {
		return &models.ToolResult{
			ToolCallID: toolCall.ID,
			Name:       toolCall.Function.Name,
			Error:      fmt.Sprintf("Invalid arguments: %v", err),
		}, nil // Return error as result, don't fail
	}

	// Route to appropriate handler
	switch toolCall.Function.Name {
	case "check_company_profile":
		return e.handleCheckCompanyProfile(ctx, companyID, toolCall.ID)

	case "update_company_info":
		return e.handleUpdateCompanyInfo(ctx, companyID, toolCall.ID, args)

	case "update_company_social_media":
		return e.handleUpdateCompanySocialMedia(ctx, companyID, toolCall.ID, args)

	case "create_product":
		return e.handleCreateProduct(ctx, companyID, toolCall.ID, args)

	case "list_products":
		return e.handleListProducts(ctx, companyID, toolCall.ID, args)

	default:
		return &models.ToolResult{
			ToolCallID: toolCall.ID,
			Name:       toolCall.Function.Name,
			Error:      fmt.Sprintf("Unknown tool: %s", toolCall.Function.Name),
		}, nil
	}
}

// handleCheckCompanyProfile checks company profile and identifies missing fields
func (e *Executor) handleCheckCompanyProfile(ctx context.Context, companyID string, toolCallID string) (*models.ToolResult, error) {
	var company models.Company
	var socialMediaHandlesJSON, marketplacesJSON []byte
	err := e.db.Pool().QueryRow(ctx, `
		SELECT id, owner_user_id, name, COALESCE(description, ''), COALESCE(industry, ''), COALESCE(business_model, ''), 
		       founded_year, COALESCE(location_region, ''), COALESCE(city, ''), COALESCE(country, 'ID'), COALESCE(website, ''), 
		       social_media_handles, marketplaces, created_at
		FROM companies WHERE id = $1
	`, companyID).Scan(
		&company.ID, &company.OwnerUserID, &company.Name, &company.Description,
		&company.Industry, &company.BusinessModel, &company.FoundedYear,
		&company.LocationRegion, &company.City, &company.Country, &company.Website,
		&socialMediaHandlesJSON, &marketplacesJSON, &company.CreatedAt,
	)

	if err != nil {
		return &models.ToolResult{
			ToolCallID: toolCallID,
			Name:       "check_company_profile",
			Error:      fmt.Sprintf("Failed to fetch company: %v", err),
		}, nil
	}

	// Parse JSONB fields
	if len(socialMediaHandlesJSON) > 0 {
		json.Unmarshal(socialMediaHandlesJSON, &company.SocialMediaHandles)
	}
	if len(marketplacesJSON) > 0 {
		json.Unmarshal(marketplacesJSON, &company.Marketplaces)
	}

	// Identify missing fields
	missingFields := []string{}
	if company.Industry == "" {
		missingFields = append(missingFields, "industry")
	}
	if company.City == "" {
		missingFields = append(missingFields, "city")
	}
	if company.LocationRegion == "" {
		missingFields = append(missingFields, "location_region")
	}
	if company.BusinessModel == "" {
		missingFields = append(missingFields, "business_model")
	}
	if company.SocialMediaHandles == nil || len(company.SocialMediaHandles) == 0 {
		missingFields = append(missingFields, "social_media")
	}

	resultData := map[string]interface{}{
		"company":        company,
		"missing_fields": missingFields,
	}

	resultJSON, _ := json.Marshal(resultData)
	return &models.ToolResult{
		ToolCallID: toolCallID,
		Name:       "check_company_profile",
		Content:    string(resultJSON),
		Data:       resultData,
	}, nil
}

// handleUpdateCompanyInfo updates company information fields
func (e *Executor) handleUpdateCompanyInfo(ctx context.Context, companyID string, toolCallID string, args map[string]interface{}) (*models.ToolResult, error) {
	// Extract fields (all optional)
	industry, _ := args["industry"].(string)
	businessModel, _ := args["business_model"].(string)
	city, _ := args["city"].(string)
	locationRegion, _ := args["location_region"].(string)
	description, _ := args["description"].(string)

	// Build update query - only update non-empty fields
	_, err := e.db.Pool().Exec(ctx, `
		UPDATE companies 
		SET industry = COALESCE(NULLIF($2, ''), industry),
			business_model = COALESCE(NULLIF($3, ''), business_model),
			city = COALESCE(NULLIF($4, ''), city),
			location_region = COALESCE(NULLIF($5, ''), location_region),
			description = COALESCE(NULLIF($6, ''), description),
			updated_at = NOW()
		WHERE id = $1
	`, companyID, industry, businessModel, city, locationRegion, description)

	if err != nil {
		return &models.ToolResult{
			ToolCallID: toolCallID,
			Name:       "update_company_info",
			Error:      fmt.Sprintf("Failed to update company: %v", err),
		}, nil
	}

	resultData := map[string]interface{}{
		"message": "Company information updated successfully",
		"updated_fields": map[string]interface{}{
			"industry":        industry,
			"business_model":  businessModel,
			"city":            city,
			"location_region": locationRegion,
			"description":     description,
		},
	}

	resultJSON, _ := json.Marshal(resultData)
	return &models.ToolResult{
		ToolCallID: toolCallID,
		Name:       "update_company_info",
		Content:    string(resultJSON),
		Data:       resultData,
	}, nil
}

// handleUpdateCompanySocialMedia updates social media handles
func (e *Executor) handleUpdateCompanySocialMedia(ctx context.Context, companyID string, toolCallID string, args map[string]interface{}) (*models.ToolResult, error) {
	platform, ok := args["platform"].(string)
	if !ok || platform == "" {
		return &models.ToolResult{
			ToolCallID: toolCallID,
			Name:       "update_company_social_media",
			Error:      "platform is required",
		}, nil
	}

	handle, ok := args["handle"].(string)
	if !ok || handle == "" {
		return &models.ToolResult{
			ToolCallID: toolCallID,
			Name:       "update_company_social_media",
			Error:      "handle is required",
		}, nil
	}

	// Validate platform
	validPlatforms := map[string]bool{
		"instagram": true, "tiktok": true, "facebook": true, "twitter": true,
		"tokopedia": true, "shopee": true, "lazada": true, "bukalapak": true,
	}
	if !validPlatforms[platform] {
		return &models.ToolResult{
			ToolCallID: toolCallID,
			Name:       "update_company_social_media",
			Error:      fmt.Sprintf("Invalid platform: %s. Must be one of: instagram, tiktok, facebook, twitter, tokopedia, shopee, lazada, bukalapak", platform),
		}, nil
	}

	// Get existing handles
	var existingHandlesJSON []byte
	err := e.db.Pool().QueryRow(ctx, `
		SELECT COALESCE(social_media_handles, '{}'::jsonb)
		FROM companies WHERE id = $1
	`, companyID).Scan(&existingHandlesJSON)

	if err != nil {
		return &models.ToolResult{
			ToolCallID: toolCallID,
			Name:       "update_company_social_media",
			Error:      fmt.Sprintf("Failed to fetch company: %v", err),
		}, nil
	}

	// Parse and update handles
	var handles map[string]string
	if len(existingHandlesJSON) > 0 {
		json.Unmarshal(existingHandlesJSON, &handles)
	}
	if handles == nil {
		handles = make(map[string]string)
	}

	handles[platform] = handle
	handlesJSON, err := json.Marshal(handles)
	if err != nil {
		return &models.ToolResult{
			ToolCallID: toolCallID,
			Name:       "update_company_social_media",
			Error:      fmt.Sprintf("Failed to serialize handles: %v", err),
		}, nil
	}

	// Update database
	_, err = e.db.Pool().Exec(ctx, `
		UPDATE companies 
		SET social_media_handles = $2,
		    updated_at = NOW()
		WHERE id = $1
	`, companyID, string(handlesJSON))

	if err != nil {
		return &models.ToolResult{
			ToolCallID: toolCallID,
			Name:       "update_company_social_media",
			Error:      fmt.Sprintf("Failed to update social media: %v", err),
		}, nil
	}

	resultData := map[string]interface{}{
		"message":  "Social media handle updated successfully",
		"platform": platform,
		"handle":   handle,
	}

	resultJSON, _ := json.Marshal(resultData)
	return &models.ToolResult{
		ToolCallID: toolCallID,
		Name:       "update_company_social_media",
		Content:    string(resultJSON),
		Data:       resultData,
	}, nil
}

// handleCreateProduct creates a new product
func (e *Executor) handleCreateProduct(ctx context.Context, companyID string, toolCallID string, args map[string]interface{}) (*models.ToolResult, error) {
	name, ok := args["name"].(string)
	if !ok || name == "" {
		return &models.ToolResult{
			ToolCallID: toolCallID,
			Name:       "create_product",
			Error:      "name is required",
		}, nil
	}

	category, _ := args["category"].(string)
	unitPrice := 0.0
	if up, ok := args["unit_price"].(float64); ok {
		unitPrice = up
	}
	cost := 0.0
	if c, ok := args["cost"].(float64); ok {
		cost = c
	}
	sku, _ := args["sku"].(string)

	// Create product using same logic as CreateProduct handler
	productID := uuid.New().String()

	_, err := e.db.Pool().Exec(ctx, `
		INSERT INTO products (id, company_id, name, sku, category, unit_price, cost, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, NOW(), NOW())
	`, productID, companyID, name, sku, category, unitPrice, cost)

	if err != nil {
		return &models.ToolResult{
			ToolCallID: toolCallID,
			Name:       "create_product",
			Error:      fmt.Sprintf("Failed to create product: %v", err),
		}, nil
	}

	// Fetch created product
	var p models.Product
	err = e.db.Pool().QueryRow(ctx, `
		SELECT id, company_id, name, sku, category, unit_price, cost, created_at, updated_at
		FROM products 
		WHERE id = $1
	`, productID).Scan(
		&p.ID, &p.CompanyID, &p.Name, &p.SKU, &p.Category, &p.UnitPrice, &p.Cost, &p.CreatedAt, &p.UpdatedAt,
	)

	if err != nil {
		return &models.ToolResult{
			ToolCallID: toolCallID,
			Name:       "create_product",
			Error:      fmt.Sprintf("Product created but failed to fetch: %v", err),
		}, nil
	}

	resultData := map[string]interface{}{
		"message": "Product created successfully",
		"product": p,
	}

	resultJSON, _ := json.Marshal(resultData)
	return &models.ToolResult{
		ToolCallID: toolCallID,
		Name:       "create_product",
		Content:    string(resultJSON),
		Data:       resultData,
	}, nil
}

// handleListProducts lists products for the company
func (e *Executor) handleListProducts(ctx context.Context, companyID string, toolCallID string, args map[string]interface{}) (*models.ToolResult, error) {
	category, _ := args["category"].(string)

	var rows pgx.Rows
	var err error

	if category != "" {
		rows, err = e.db.Pool().Query(ctx, `
			SELECT id, company_id, name, sku, category, unit_price, cost, created_at, updated_at
			FROM products 
			WHERE company_id = $1 AND category = $2
			ORDER BY name
		`, companyID, category)
	} else {
		rows, err = e.db.Pool().Query(ctx, `
			SELECT id, company_id, name, sku, category, unit_price, cost, created_at, updated_at
			FROM products 
			WHERE company_id = $1
			ORDER BY name
		`, companyID)
	}

	if err != nil {
		return &models.ToolResult{
			ToolCallID: toolCallID,
			Name:       "list_products",
			Error:      fmt.Sprintf("Failed to list products: %v", err),
		}, nil
	}
	defer rows.Close()

	products := []models.Product{}
	for rows.Next() {
		var p models.Product
		if err := rows.Scan(&p.ID, &p.CompanyID, &p.Name, &p.SKU, &p.Category, &p.UnitPrice, &p.Cost, &p.CreatedAt, &p.UpdatedAt); err == nil {
			products = append(products, p)
		}
	}

	resultData := map[string]interface{}{
		"products": products,
		"count":    len(products),
	}

	resultJSON, _ := json.Marshal(resultData)
	return &models.ToolResult{
		ToolCallID: toolCallID,
		Name:       "list_products",
		Content:    string(resultJSON),
		Data:       resultData,
	}, nil
}
