package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/bantuaku/backend/middleware"
	"github.com/bantuaku/backend/models"
)

// UpdateCompanyRequest represents a request to update company information
type UpdateCompanyRequest struct {
	Industry       string `json:"industry,omitempty"`
	BusinessModel  string `json:"business_model,omitempty"`
	City           string `json:"city,omitempty"`
	LocationRegion string `json:"location_region,omitempty"`
	Description    string `json:"description,omitempty"`
}

// UpdateCompanyResponse represents the response when updating company
type UpdateCompanyResponse struct {
	Company *models.Company `json:"company"`
	Message string          `json:"message"`
}

// UpdateCompany updates the authenticated user's company information
// PATCH /api/v1/companies/me
func (h *Handler) UpdateCompany(w http.ResponseWriter, r *http.Request) {
	companyID := middleware.GetCompanyID(r.Context())
	if companyID == "" {
		respondError(w, http.StatusUnauthorized, "Company not found in context")
		return
	}

	var req UpdateCompanyRequest
	if err := parseJSON(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	ctx := r.Context()
	now := time.Now()

	// Build dynamic update query - only update provided fields
	// Use COALESCE to preserve existing values if new value is empty
	_, err := h.db.Pool().Exec(ctx, `
		UPDATE companies 
		SET industry = COALESCE(NULLIF($2, ''), industry),
			business_model = COALESCE(NULLIF($3, ''), business_model),
			city = COALESCE(NULLIF($4, ''), city),
			location_region = COALESCE(NULLIF($5, ''), location_region),
			description = COALESCE(NULLIF($6, ''), description),
			updated_at = $7
		WHERE id = $1
	`, companyID, req.Industry, req.BusinessModel, req.City, req.LocationRegion, req.Description, now)

	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to update company")
		return
	}

	// Fetch and return updated company
	var company models.Company
	var socialMediaHandlesJSON, marketplacesJSON []byte
	err = h.db.Pool().QueryRow(ctx, `
		SELECT id, owner_user_id, name, description, industry, business_model, 
		       founded_year, location_region, city, country, website, 
		       social_media_handles, marketplaces, created_at, updated_at
		FROM companies WHERE id = $1
	`, companyID).Scan(
		&company.ID, &company.OwnerUserID, &company.Name, &company.Description,
		&company.Industry, &company.BusinessModel, &company.FoundedYear,
		&company.LocationRegion, &company.City, &company.Country, &company.Website,
		&socialMediaHandlesJSON, &marketplacesJSON, &company.CreatedAt, &company.UpdatedAt,
	)

	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to fetch updated company")
		return
	}

	// Parse JSONB fields
	if len(socialMediaHandlesJSON) > 0 {
		json.Unmarshal(socialMediaHandlesJSON, &company.SocialMediaHandles)
	}
	if len(marketplacesJSON) > 0 {
		json.Unmarshal(marketplacesJSON, &company.Marketplaces)
	}

	respondJSON(w, http.StatusOK, UpdateCompanyResponse{
		Company: &company,
		Message: "Company updated successfully",
	})
}

// UpdateCompanySocialMediaRequest represents a request to update social media handles
type UpdateCompanySocialMediaRequest struct {
	Platform string `json:"platform" validate:"required,oneof=instagram tiktok facebook twitter tokopedia shopee lazada bukalapak"`
	Handle   string `json:"handle" validate:"required"`
}

// UpdateCompanySocialMedia updates social media handles for the company
// PATCH /api/v1/companies/me/social-media
func (h *Handler) UpdateCompanySocialMedia(w http.ResponseWriter, r *http.Request) {
	companyID := middleware.GetCompanyID(r.Context())
	if companyID == "" {
		respondError(w, http.StatusUnauthorized, "Company not found in context")
		return
	}

	var req UpdateCompanySocialMediaRequest
	if err := parseJSON(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	ctx := r.Context()
	now := time.Now()

	// Get existing social media handles
	var existingHandlesJSON []byte
	err := h.db.Pool().QueryRow(ctx, `
		SELECT COALESCE(social_media_handles, '{}'::jsonb)
		FROM companies WHERE id = $1
	`, companyID).Scan(&existingHandlesJSON)

	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to fetch company")
		return
	}

	// Parse existing handles
	var handles map[string]string
	if len(existingHandlesJSON) > 0 {
		json.Unmarshal(existingHandlesJSON, &handles)
	}
	if handles == nil {
		handles = make(map[string]string)
	}

	// Update/add the new handle
	handles[req.Platform] = req.Handle

	// Convert back to JSON
	handlesJSON, err := json.Marshal(handles)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to serialize handles")
		return
	}

	// Update database
	_, err = h.db.Pool().Exec(ctx, `
		UPDATE companies 
		SET social_media_handles = $2,
		    updated_at = $3
		WHERE id = $1
	`, companyID, string(handlesJSON), now)

	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to update social media handles")
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"message":  "Social media handle updated successfully",
		"platform": req.Platform,
		"handle":   req.Handle,
	})
}

// GetCompanyProfile returns the company profile with missing fields identified
// GET /api/v1/companies/me/profile
func (h *Handler) GetCompanyProfile(w http.ResponseWriter, r *http.Request) {
	companyID := middleware.GetCompanyID(r.Context())
	if companyID == "" {
		respondError(w, http.StatusUnauthorized, "Company not found in context")
		return
	}

	ctx := r.Context()

	var company models.Company
	var socialMediaHandlesJSON, marketplacesJSON []byte
	err := h.db.Pool().QueryRow(ctx, `
		SELECT id, owner_user_id, name, description, industry, business_model, 
		       founded_year, location_region, city, country, website, 
		       social_media_handles, marketplaces, created_at, updated_at
		FROM companies WHERE id = $1
	`, companyID).Scan(
		&company.ID, &company.OwnerUserID, &company.Name, &company.Description,
		&company.Industry, &company.BusinessModel, &company.FoundedYear,
		&company.LocationRegion, &company.City, &company.Country, &company.Website,
		&socialMediaHandlesJSON, &marketplacesJSON, &company.CreatedAt, &company.UpdatedAt,
	)

	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to fetch company")
		return
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

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"company":        company,
		"missing_fields": missingFields,
	})
}
