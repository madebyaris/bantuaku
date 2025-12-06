package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/bantuaku/backend/errors"
	"github.com/bantuaku/backend/logger"
	"github.com/bantuaku/backend/middleware"
	"github.com/bantuaku/backend/services/prediction"
)

// PredictionHandler handles prediction-related endpoints
type PredictionHandler struct {
	service *prediction.Service
}

// NewPredictionHandler creates a new prediction handler
func NewPredictionHandler(service *prediction.Service) *PredictionHandler {
	return &PredictionHandler{service: service}
}

// CheckCompletenessResponse represents the completeness check response
type CheckCompletenessResponse struct {
	IsComplete  bool     `json:"is_complete"`
	HasIndustry bool     `json:"has_industry"`
	HasCity     bool     `json:"has_city"`
	HasProducts bool     `json:"has_products"`
	HasSocial   bool     `json:"has_social"`
	Missing     []string `json:"missing,omitempty"`
}

// predictionRespondError sends an error response for prediction handlers
func predictionRespondError(w http.ResponseWriter, err error, r *http.Request) {
	log := logger.With("request_id", r.Context().Value("request_id"))
	log.LogError(err, "Prediction handler error", r.Context())
	errors.WriteJSONError(w, err, errors.GetErrorCode(err))
}

// CheckCompleteness checks if the company profile is complete enough for predictions
func (h *PredictionHandler) CheckCompleteness(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logger.With("request_id", ctx.Value("request_id"))

	companyID := middleware.GetCompanyID(ctx)
	if companyID == "" {
		log.Error("Company ID not found in context")
		predictionRespondError(w, errors.NewUnauthorizedError("Company not found"), r)
		return
	}

	result, err := h.service.CheckCompleteness(ctx, companyID)
	if err != nil {
		log.Error("Failed to check completeness", "error", err)
		predictionRespondError(w, errors.NewInternalError(err, "Failed to check profile completeness"), r)
		return
	}

	respondJSON(w, http.StatusOK, CheckCompletenessResponse{
		IsComplete:  result.IsComplete,
		HasIndustry: result.HasIndustry,
		HasCity:     result.HasCity,
		HasProducts: result.HasProducts,
		HasSocial:   result.HasSocial,
		Missing:     result.Missing,
	})
}

// StartPredictionResponse represents the response when starting a prediction job
type StartPredictionResponse struct {
	JobID   string `json:"job_id"`
	Status  string `json:"status"`
	Message string `json:"message"`
}

// StartPrediction starts a new prediction job
func (h *PredictionHandler) StartPrediction(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logger.With("request_id", ctx.Value("request_id"))

	companyID := middleware.GetCompanyID(ctx)
	if companyID == "" {
		log.Error("Company ID not found in context")
		predictionRespondError(w, errors.NewUnauthorizedError("Company not found"), r)
		return
	}

	// Check completeness first
	completeness, err := h.service.CheckCompleteness(ctx, companyID)
	if err != nil {
		log.Error("Failed to check completeness", "error", err)
		predictionRespondError(w, errors.NewInternalError(err, "Failed to check profile completeness"), r)
		return
	}

	if !completeness.IsComplete {
		predictionRespondError(w, errors.NewValidationError("Profile not complete. Missing: "+formatMissing(completeness.Missing), ""), r)
		return
	}

	// Start the job
	job, err := h.service.StartJob(ctx, companyID)
	if err != nil {
		log.Error("Failed to start prediction job", "error", err)
		predictionRespondError(w, errors.NewValidationError(err.Error(), ""), r)
		return
	}

	log.Info("Prediction job started", "job_id", job.ID)

	respondJSON(w, http.StatusAccepted, StartPredictionResponse{
		JobID:   job.ID,
		Status:  string(job.Status),
		Message: "Prediction job started. You will be notified when complete.",
	})
}

// JobStatusResponse represents the job status response
type JobStatusResponse struct {
	JobID        string              `json:"job_id,omitempty"`
	Status       string              `json:"status"`
	Progress     prediction.Progress `json:"progress,omitempty"`
	Results      *prediction.Results `json:"results,omitempty"`
	ErrorMessage string              `json:"error_message,omitempty"`
	HasActiveJob bool                `json:"has_active_job"`
}

// GetStatus gets the current prediction job status
func (h *PredictionHandler) GetStatus(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logger.With("request_id", ctx.Value("request_id"))

	companyID := middleware.GetCompanyID(ctx)
	if companyID == "" {
		log.Error("Company ID not found in context")
		predictionRespondError(w, errors.NewUnauthorizedError("Company not found"), r)
		return
	}

	// Check for specific job ID in query
	jobID := r.URL.Query().Get("job_id")

	var job *prediction.Job
	var err error

	if jobID != "" {
		job, err = h.service.GetJob(ctx, jobID)
	} else {
		job, err = h.service.GetActiveJob(ctx, companyID)
	}

	if err != nil {
		log.Error("Failed to get job status", "error", err)
		predictionRespondError(w, errors.NewInternalError(err, "Failed to get job status"), r)
		return
	}

	if job == nil {
		respondJSON(w, http.StatusOK, JobStatusResponse{
			Status:       "none",
			HasActiveJob: false,
		})
		return
	}

	response := JobStatusResponse{
		JobID:        job.ID,
		Status:       string(job.Status),
		Progress:     job.Progress,
		HasActiveJob: job.Status == prediction.StatusPending || job.Status == prediction.StatusProcessing,
	}

	if job.Status == prediction.StatusCompleted {
		response.Results = &job.Results
	}

	if job.Status == prediction.StatusFailed {
		response.ErrorMessage = job.ErrorMessage
	}

	respondJSON(w, http.StatusOK, response)
}

// GetLatestResults gets the latest completed prediction results
func (h *PredictionHandler) GetLatestResults(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logger.With("request_id", ctx.Value("request_id"))

	companyID := middleware.GetCompanyID(ctx)
	if companyID == "" {
		log.Error("Company ID not found in context")
		predictionRespondError(w, errors.NewUnauthorizedError("Company not found"), r)
		return
	}

	// Get latest completed job
	var job prediction.Job
	var progressJSON, resultsJSON []byte

	err := h.service.Pool().QueryRow(ctx, `
		SELECT id, status, progress, results, COALESCE(error_message, ''), completed_at
		FROM prediction_jobs
		WHERE company_id = $1 AND status = 'completed'
		ORDER BY completed_at DESC
		LIMIT 1
	`, companyID).Scan(&job.ID, &job.Status, &progressJSON, &resultsJSON, &job.ErrorMessage, &job.CompletedAt)

	if err != nil {
		if err.Error() == "no rows in result set" {
			respondJSON(w, http.StatusOK, map[string]interface{}{
				"has_results": false,
				"message":     "No prediction results available yet. Click 'Predict It!' to generate insights.",
			})
			return
		}
		log.Error("Failed to get latest results", "error", err)
		predictionRespondError(w, errors.NewInternalError(err, "Failed to get results"), r)
		return
	}

	json.Unmarshal(progressJSON, &job.Progress)
	json.Unmarshal(resultsJSON, &job.Results)

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"has_results":  true,
		"job_id":       job.ID,
		"completed_at": job.CompletedAt,
		"results":      job.Results,
	})
}

// formatMissing formats the missing fields into a readable string
func formatMissing(missing []string) string {
	if len(missing) == 0 {
		return ""
	}
	result := ""
	for i, m := range missing {
		if i > 0 {
			result += ", "
		}
		switch m {
		case "industry":
			result += "industri bisnis"
		case "city":
			result += "lokasi/kota"
		case "products":
			result += "produk/layanan"
		case "social_media":
			result += "akun social media"
		default:
			result += m
		}
	}
	return result
}

