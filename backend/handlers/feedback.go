package handlers

import (
	"net/http"
	"time"

	"github.com/bantuaku/backend/errors"
	"github.com/bantuaku/backend/logger"
	"github.com/bantuaku/backend/validation"
	"github.com/google/uuid"
)

// SubmitFeedbackRequest represents a feedback submission
type SubmitFeedbackRequest struct {
	MessageID    string  `json:"message_id" validate:"required"`
	FeedbackType string  `json:"feedback_type" validate:"required,oneof=positive negative neutral"`
	Comment      *string `json:"comment,omitempty"`
}

// SubmitFeedback handles user feedback submission
func (h *Handler) SubmitFeedback(w http.ResponseWriter, r *http.Request) {
	log := logger.With("request_id", r.Context().Value("request_id"))
	userID := r.Context().Value("user_id").(string)
	if userID == "" {
		h.respondError(w, errors.NewValidationError("user_id required", ""), r)
		return
	}

	var req SubmitFeedbackRequest
	if err := h.parseJSON(r, &req); err != nil {
		h.respondError(w, err, r)
		return
	}

	if err := validation.Validate(&req); err != nil {
		h.respondError(w, err, r)
		return
	}

	// Store feedback in database
	feedbackID := uuid.New().String()
	pool := h.db.Pool()
	ctx := r.Context()

	_, err := pool.Exec(ctx,
		`INSERT INTO chat_feedback (id, message_id, user_id, feedback_type, comment, created_at)
		 VALUES ($1, $2, $3, $4, $5, NOW())`,
		feedbackID, req.MessageID, userID, req.FeedbackType, req.Comment,
	)
	if err != nil {
		log.Error("Failed to store feedback", "error", err)
		h.respondError(w, errors.NewInternalError(err, "Failed to store feedback"), r)
		return
	}

	log.Info("Feedback submitted", "message_id", req.MessageID, "type", req.FeedbackType)

	h.respondJSON(w, http.StatusOK, map[string]interface{}{
		"feedback_id": feedbackID,
		"message":     "Feedback submitted successfully",
		"created_at":  time.Now(),
	})
}

