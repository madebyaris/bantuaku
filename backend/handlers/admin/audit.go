package admin

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/bantuaku/backend/errors"
)

// AuditLog represents an audit log entry
type AuditLog struct {
	ID           int64                  `json:"id"`
	UserID       *string                `json:"user_id,omitempty"`
	CompanyID    *string                `json:"company_id,omitempty"`
	Action       string                 `json:"action"`
	ResourceType *string                `json:"resource_type,omitempty"`
	ResourceID   *string                `json:"resource_id,omitempty"`
	IPAddress    *string                `json:"ip_address,omitempty"`
	UserAgent    *string                `json:"user_agent,omitempty"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt    time.Time              `json:"created_at"`
}

// ListAuditLogs lists audit logs with pagination and filtering
func (h *AdminHandler) ListAuditLogs(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit < 1 || limit > 100 {
		limit = 50
	}
	offset := (page - 1) * limit

	// Filter parameters
	action := r.URL.Query().Get("action")
	resourceType := r.URL.Query().Get("resource_type")
	userID := r.URL.Query().Get("user_id")

	query := `
		SELECT id, user_id, company_id, action, resource_type, resource_id,
		       ip_address, user_agent, metadata, created_at
		FROM audit_logs
		WHERE 1=1
	`
	args := []interface{}{}
	argPos := 1

	if action != "" {
		query += ` AND action = $` + strconv.Itoa(argPos)
		args = append(args, action)
		argPos++
	}
	if resourceType != "" {
		query += ` AND resource_type = $` + strconv.Itoa(argPos)
		args = append(args, resourceType)
		argPos++
	}
	if userID != "" {
		query += ` AND user_id = $` + strconv.Itoa(argPos)
		args = append(args, userID)
		argPos++
	}

	query += ` ORDER BY created_at DESC LIMIT $` + strconv.Itoa(argPos) + ` OFFSET $` + strconv.Itoa(argPos+1)
	args = append(args, limit, offset)

	rows, err := h.db.Pool().Query(ctx, query, args...)
	if err != nil {
		appErr := errors.NewDatabaseError(err, "list audit logs")
		h.respondError(w, appErr, r)
		return
	}
	defer rows.Close()

	var logs []AuditLog
	for rows.Next() {
		var log AuditLog
		var userID, companyID, resourceType, resourceID, ipAddress, userAgent *string
		var metadataJSON []byte

		if err := rows.Scan(
			&log.ID, &userID, &companyID, &log.Action,
			&resourceType, &resourceID, &ipAddress, &userAgent,
			&metadataJSON, &log.CreatedAt,
		); err != nil {
			h.log.Error("Failed to scan audit log", "error", err)
			continue
		}

		log.UserID = userID
		log.CompanyID = companyID
		log.ResourceType = resourceType
		log.ResourceID = resourceID
		if ipAddress != nil {
			log.IPAddress = ipAddress
		}
		if userAgent != nil {
			log.UserAgent = userAgent
		}

		if len(metadataJSON) > 0 {
			if err := json.Unmarshal(metadataJSON, &log.Metadata); err != nil {
				h.log.Error("Failed to unmarshal metadata", "error", err)
			}
		}

		logs = append(logs, log)
	}

	// Get total count
	countQuery := `SELECT COUNT(*) FROM audit_logs WHERE 1=1`
	countArgs := []interface{}{}
	countPos := 1

	if action != "" {
		countQuery += ` AND action = $` + strconv.Itoa(countPos)
		countArgs = append(countArgs, action)
		countPos++
	}
	if resourceType != "" {
		countQuery += ` AND resource_type = $` + strconv.Itoa(countPos)
		countArgs = append(countArgs, resourceType)
		countPos++
	}
	if userID != "" {
		countQuery += ` AND user_id = $` + strconv.Itoa(countPos)
		countArgs = append(countArgs, userID)
	}

	var total int
	err = h.db.Pool().QueryRow(ctx, countQuery, countArgs...).Scan(&total)
	if err != nil {
		appErr := errors.NewDatabaseError(err, "count audit logs")
		h.respondError(w, appErr, r)
		return
	}

	h.respondJSON(w, http.StatusOK, map[string]interface{}{
		"logs": logs,
		"pagination": map[string]interface{}{
			"page":  page,
			"limit": limit,
			"total": total,
		},
	})
}

