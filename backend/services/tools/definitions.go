package tools

import "github.com/bantuaku/backend/models"

// GetToolDefinitions returns all available tool definitions for AI function calling
func GetToolDefinitions() []models.Tool {
	return []models.Tool{
		{
			Type: "function",
			Function: models.Function{
				Name:        "check_company_profile",
				Description: "Check current company profile and identify missing required fields (industry, location, social media, etc.)",
				Parameters: map[string]interface{}{
					"type":       "object",
					"properties": map[string]interface{}{},
				},
			},
		},
		{
			Type: "function",
			Function: models.Function{
				Name:        "update_company_info",
				Description: "Update company information fields (industry, location, business model, description). Only updates provided fields, does not delete existing data.",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"industry": map[string]interface{}{
							"type":        "string",
							"description": "Business industry/type (e.g., 'Kuliner', 'Retail', 'Jasa', 'Manufacturing')",
						},
						"business_model": map[string]interface{}{
							"type":        "string",
							"description": "Business model (e.g., 'retail', 'service', 'manufacturing', 'online', 'offline', 'hybrid')",
						},
						"city": map[string]interface{}{
							"type":        "string",
							"description": "City where business operates",
						},
						"location_region": map[string]interface{}{
							"type":        "string",
							"description": "Region/province (e.g., 'DKI Jakarta', 'Jawa Barat')",
						},
						"description": map[string]interface{}{
							"type":        "string",
							"description": "Company description or about text",
						},
					},
				},
			},
		},
		{
			Type: "function",
			Function: models.Function{
				Name:        "update_company_social_media",
				Description: "Add or update social media handles for the company. Merges with existing handles, does not replace all.",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"platform": map[string]interface{}{
							"type":        "string",
							"enum":        []string{"instagram", "tiktok", "facebook", "twitter", "tokopedia", "shopee", "lazada", "bukalapak"},
							"description": "Social media platform name",
						},
						"handle": map[string]interface{}{
							"type":        "string",
							"description": "Social media handle/username (without @ symbol)",
						},
					},
					"required": []string{"platform", "handle"},
				},
			},
		},
		{
			Type: "function",
			Function: models.Function{
				Name:        "create_product",
				Description: "Create a new product or service for the company. Product can be modified or removed later by the owner.",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"name": map[string]interface{}{
							"type":        "string",
							"description": "Product/service name",
						},
						"category": map[string]interface{}{
							"type":        "string",
							"description": "Product category (e.g., 'Makanan', 'Minuman', 'Jasa', 'Elektronik')",
						},
						"unit_price": map[string]interface{}{
							"type":        "number",
							"description": "Price per unit in IDR",
						},
						"cost": map[string]interface{}{
							"type":        "number",
							"description": "Cost per unit in IDR",
						},
						"sku": map[string]interface{}{
							"type":        "string",
							"description": "SKU code (optional)",
						},
					},
					"required": []string{"name"},
				},
			},
		},
		{
			Type: "function",
			Function: models.Function{
				Name:        "list_products",
				Description: "List all products/services for the company. Can filter by category.",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"category": map[string]interface{}{
							"type":        "string",
							"description": "Filter by category (optional)",
						},
					},
				},
			},
		},
	}
}
