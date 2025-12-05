package admin

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/bantuaku/backend/errors"
	"github.com/bantuaku/backend/logger"
	"github.com/bantuaku/backend/services/settings"
)

// GetAIProvider returns the current AI provider setting
func (h *AdminHandler) GetAIProvider(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logger.With("request_id", r.Context().Value("request_id"))

	// #region agent log
	if f, err := os.OpenFile("/Volumes/app/hackathon/imphxkolosal/bantuaku/.cursor/debug.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644); err == nil {
		fmt.Fprintf(f, `{"sessionId":"debug-session","runId":"run1","hypothesisId":"A","location":"admin/settings.go:GetAIProvider:entry","message":"GetAIProvider handler called","data":{"dbNil":%t},"timestamp":%d}`+"\n", h.db == nil, 0)
		f.Close()
	}
	// #endregion
	settingsService := settings.NewService(h.db)
	settingValue, err := settingsService.GetSetting(ctx, "ai_provider")
	// #region agent log
	if f, err2 := os.OpenFile("/Volumes/app/hackathon/imphxkolosal/bantuaku/.cursor/debug.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644); err2 == nil {
		errorMsg := ""
		if err != nil {
			errorMsg = err.Error()
		}
		fmt.Fprintf(f, `{"sessionId":"debug-session","runId":"run1","hypothesisId":"A","location":"admin/settings.go:GetAIProvider:afterGetSetting","message":"After GetSetting call","data":{"settingValue":"%s","error":%t,"errorMsg":"%s"},"timestamp":%d}`+"\n", settingValue, err != nil, errorMsg, 0)
		f.Close()
	}
	// #endregion
	if err != nil {
		// #region agent log
		if f, err2 := os.OpenFile("/Volumes/app/hackathon/imphxkolosal/bantuaku/.cursor/debug.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644); err2 == nil {
			fmt.Fprintf(f, `{"sessionId":"debug-session","runId":"run1","hypothesisId":"A","location":"admin/settings.go:GetAIProvider:error","message":"GetSetting returned error","data":{"error":"%s"},"timestamp":%d}`+"\n", err.Error(), 0)
			f.Close()
		}
		// #endregion
		log.Error("Failed to get AI provider setting", "error", err)
		// Check if error is about missing table - provide helpful message
		if err != nil && (err.Error() != "" && (err.Error() == "relation \"settings\" does not exist" || err.Error() == "failed to query setting: relation \"settings\" does not exist")) {
			h.respondError(w, errors.NewAppError(errors.ErrCodeInternal, "Settings table not found. Please run migration 014_add_settings_table.sql", err.Error()), r)
		} else {
			h.respondError(w, errors.NewAppError(errors.ErrCodeInternal, "Failed to get AI provider setting", err.Error()), r)
		}
		return
	}

	// Parse JSON value
	var settingData map[string]interface{}
	provider := "openrouter" // default
	if settingValue != "" {
		if err := json.Unmarshal([]byte(settingValue), &settingData); err == nil {
			if p, ok := settingData["provider"].(string); ok && p != "" {
				provider = p
			}
		}
	}

	h.respondJSON(w, http.StatusOK, map[string]string{
		"provider": provider,
	})
}

// UpdateAIProviderRequest represents the request to update AI provider
type UpdateAIProviderRequest struct {
	Provider string `json:"provider" validate:"required,oneof=openrouter kolosal"`
}

// UpdateAIProvider updates the AI provider setting
func (h *AdminHandler) UpdateAIProvider(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logger.With("request_id", r.Context().Value("request_id"))

	var req UpdateAIProviderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, errors.NewValidationError("Invalid JSON format", err.Error()), r)
		return
	}

	// Validate provider value
	if req.Provider != "openrouter" && req.Provider != "kolosal" {
		h.respondError(w, errors.NewValidationError("Invalid provider. Must be 'openrouter' or 'kolosal'", ""), r)
		return
	}

	settingsService := settings.NewService(h.db)
	settingValue := map[string]string{
		"provider": req.Provider,
	}

	if err := settingsService.SetSetting(ctx, "ai_provider", settingValue); err != nil {
		log.Error("Failed to update AI provider setting", "error", err, "provider", req.Provider)
		h.respondError(w, errors.NewAppError(errors.ErrCodeInternal, "Failed to update AI provider setting", err.Error()), r)
		return
	}

	log.Info("AI provider updated", "provider", req.Provider)

	h.respondJSON(w, http.StatusOK, map[string]string{
		"provider": req.Provider,
		"message":  "AI provider updated successfully",
	})
}
