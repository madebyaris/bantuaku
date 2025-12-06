package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/bantuaku/backend/errors"
	"github.com/bantuaku/backend/middleware"
	"github.com/bantuaku/backend/models"
	"github.com/google/uuid"
)

// ListNotifications returns notifications for the authenticated company/user
func (h *Handler) ListNotifications(w http.ResponseWriter, r *http.Request) {
	companyID := middleware.GetCompanyID(r.Context())
	if companyID == "" {
		respondError(w, http.StatusUnauthorized, "Company not found in context")
		return
	}

	status := r.URL.Query().Get("status")

	rows, err := h.db.Pool().Query(r.Context(), `
		SELECT id, company_id, COALESCE(user_id, ''), title, COALESCE(body, ''), COALESCE(type, ''), status, created_at, read_at
		FROM notifications
		WHERE company_id = $1
		AND ($2 = '' OR status = $2)
		ORDER BY created_at DESC
		LIMIT 100
	`, companyID, status)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to fetch notifications")
		return
	}
	defer rows.Close()

	var items []models.Notification
	for rows.Next() {
		var n models.Notification
		if err := rows.Scan(&n.ID, &n.CompanyID, &n.UserID, &n.Title, &n.Body, &n.Type, &n.Status, &n.CreatedAt, &n.ReadAt); err == nil {
			items = append(items, n)
		}
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"notifications": items,
		"count":         len(items),
	})
}

// MarkNotificationRead marks a notification as read
func (h *Handler) MarkNotificationRead(w http.ResponseWriter, r *http.Request) {
	companyID := middleware.GetCompanyID(r.Context())
	id := r.PathValue("id")

	if companyID == "" {
		respondError(w, http.StatusUnauthorized, "Company not found in context")
		return
	}
	if id == "" {
		respondError(w, http.StatusBadRequest, "Notification ID is required")
		return
	}

	_, err := h.db.Pool().Exec(r.Context(), `
		UPDATE notifications
		SET status = 'read', read_at = $3
		WHERE id = $1 AND company_id = $2
	`, id, companyID, time.Now())
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to mark notification as read")
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"message": "Notification marked as read"})
}

// DeleteNotification deletes a notification
func (h *Handler) DeleteNotification(w http.ResponseWriter, r *http.Request) {
	companyID := middleware.GetCompanyID(r.Context())
	id := r.PathValue("id")

	if companyID == "" {
		respondError(w, http.StatusUnauthorized, "Company not found in context")
		return
	}
	if id == "" {
		respondError(w, http.StatusBadRequest, "Notification ID is required")
		return
	}

	_, err := h.db.Pool().Exec(r.Context(), `
		DELETE FROM notifications WHERE id = $1 AND company_id = $2
	`, id, companyID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to delete notification")
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"message": "Notification deleted"})
}

// CreateNotification is a helper to create notification (internal use)
func (h *Handler) CreateNotification(ctx context.Context, companyID, userID, title, body, notifType string) error {
	if companyID == "" || title == "" {
		return errors.NewValidationError("company_id and title required", "")
	}

	_, err := h.db.Pool().Exec(ctx, `
		INSERT INTO notifications (id, company_id, user_id, title, body, type, status, created_at)
		VALUES ($1, $2, NULLIF($3, ''), $4, $5, $6, 'unread', NOW())
	`, uuid.New().String(), companyID, userID, title, body, notifType)
	return err
}
