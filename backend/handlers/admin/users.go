package admin

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/bantuaku/backend/errors"
	"github.com/bantuaku/backend/logger"
	"github.com/bantuaku/backend/middleware"
	"github.com/bantuaku/backend/services/audit"
	"github.com/bantuaku/backend/services/storage"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type AdminHandler struct {
	db          *storage.Postgres
	log         logger.Logger
	jwtSecret   string
	auditLogger *audit.Logger
}

func NewAdminHandler(db *storage.Postgres, jwtSecret string, auditLogger *audit.Logger) *AdminHandler {
	return &AdminHandler{
		db:          db,
		log:         *logger.Default(),
		jwtSecret:   jwtSecret,
		auditLogger: auditLogger,
	}
}

// User represents a user in admin context
type User struct {
	ID                 string    `json:"id"`
	CompanyID          string    `json:"company_id,omitempty"`
	Email              string    `json:"email"`
	Role               string    `json:"role"`
	Status             string    `json:"status"`
	StoreName          string    `json:"store_name"`
	Industry           string    `json:"industry"`
	SubscriptionPlan   string    `json:"subscription_plan,omitempty"`
	SubscriptionStatus string    `json:"subscription_status,omitempty"`
	CreatedAt          time.Time `json:"created_at"`
}

type AdminStats struct {
	TotalUsers          int `json:"total_users"`
	TotalSubscriptions  int `json:"total_subscriptions"`
	ActiveSubscriptions int `json:"active_subscriptions"`
	TotalAuditLogs      int `json:"total_audit_logs"`
}

func (h *AdminHandler) GetStats(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var stats AdminStats

	if err := h.db.Pool().QueryRow(ctx, `SELECT COUNT(*) FROM users`).Scan(&stats.TotalUsers); err != nil {
		appErr := errors.NewDatabaseError(err, "count users")
		h.respondError(w, appErr, r)
		return
	}

	if err := h.db.Pool().QueryRow(ctx, `SELECT COUNT(*) FROM subscriptions`).Scan(&stats.TotalSubscriptions); err != nil {
		appErr := errors.NewDatabaseError(err, "count subscriptions")
		h.respondError(w, appErr, r)
		return
	}

	if err := h.db.Pool().QueryRow(ctx, `SELECT COUNT(*) FROM subscriptions WHERE status = 'active'`).Scan(&stats.ActiveSubscriptions); err != nil {
		appErr := errors.NewDatabaseError(err, "count active subscriptions")
		h.respondError(w, appErr, r)
		return
	}

	if err := h.db.Pool().QueryRow(ctx, `SELECT COUNT(*) FROM audit_logs`).Scan(&stats.TotalAuditLogs); err != nil {
		appErr := errors.NewDatabaseError(err, "count audit logs")
		h.respondError(w, appErr, r)
		return
	}

	h.respondJSON(w, http.StatusOK, stats)
}

// ListUsers lists all users with pagination
func (h *AdminHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit < 1 || limit > 100 {
		limit = 20
	}
	offset := (page - 1) * limit

	rows, err := h.db.Pool().Query(ctx, `
		SELECT 
			u.id, 
			COALESCE(c.id, '') as company_id,
			u.email, 
			COALESCE(u.role, 'user'), 
			COALESCE(u.status, 'active'), 
			COALESCE(c.name, '') as store_name,
			COALESCE(c.industry, '') as industry,
			COALESCE(c.plan_name, 'free') as subscription_plan,
			COALESCE(c.plan_status, '') as subscription_status,
			u.created_at
		FROM users u
		LEFT JOIN LATERAL (
			SELECT 
				c.id,
				c.name,
				c.industry,
				COALESCE(sub.plan_name, c.subscription_plan) as plan_name,
				COALESCE(sub.status, '') as plan_status,
				COALESCE(sub.sort_time, c.created_at) as sort_time
			FROM companies c
			LEFT JOIN LATERAL (
				SELECT sp.name AS plan_name, s.status, COALESCE(s.current_period_start, s.updated_at, s.created_at) as sort_time
				FROM subscriptions s
				JOIN subscription_plans sp ON sp.id = s.plan_id
				WHERE s.company_id = c.id
				ORDER BY 
					CASE 
						WHEN s.status = 'active' THEN 0
						WHEN s.status = 'trialing' THEN 1
						WHEN s.status = 'past_due' THEN 2
						WHEN s.status = 'canceled' THEN 3
						ELSE 4
					END,
					sort_time DESC
				LIMIT 1
			) sub ON TRUE
			WHERE c.owner_user_id = u.id
			ORDER BY 
				CASE 
					WHEN COALESCE(sub.plan_name, c.subscription_plan, '') <> '' THEN 0
					ELSE 1
				END,
				COALESCE(sub.sort_time, c.created_at) DESC
			LIMIT 1
		) c ON TRUE
		ORDER BY u.created_at DESC
		LIMIT $1 OFFSET $2
	`, limit, offset)
	if err != nil {
		appErr := errors.NewDatabaseError(err, "list users")
		h.respondError(w, appErr, r)
		return
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var u User
		if err := rows.Scan(&u.ID, &u.CompanyID, &u.Email, &u.Role, &u.Status, &u.StoreName, &u.Industry, &u.SubscriptionPlan, &u.SubscriptionStatus, &u.CreatedAt); err != nil {
			h.log.Error("Failed to scan user", "error", err)
			continue
		}
		users = append(users, u)
	}

	// Get total count
	var total int
	err = h.db.Pool().QueryRow(ctx, `SELECT COUNT(*) FROM users`).Scan(&total)
	if err != nil {
		appErr := errors.NewDatabaseError(err, "count users")
		h.respondError(w, appErr, r)
		return
	}

	h.respondJSON(w, http.StatusOK, map[string]interface{}{
		"users": users,
		"pagination": map[string]interface{}{
			"page":  page,
			"limit": limit,
			"total": total,
		},
	})
}

// GetUser retrieves a single user by ID
func (h *AdminHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID := r.PathValue("id")
	if userID == "" {
		appErr := errors.NewValidationError("User ID is required", "")
		h.respondError(w, appErr, r)
		return
	}

	var u User
	var storeName, industry, subscriptionPlan, subscriptionStatus sql.NullString
	err := h.db.Pool().QueryRow(ctx, `
		SELECT 
			u.id, 
			COALESCE(c.id, '') as company_id,
			u.email, 
			COALESCE(u.role, 'user'), 
			COALESCE(u.status, 'active'),
			c.name as store_name,
			c.industry as industry,
			COALESCE(c.plan_name, 'free') as subscription_plan,
			COALESCE(c.plan_status, '') as subscription_status,
			u.created_at
		FROM users u
		LEFT JOIN LATERAL (
			SELECT 
				c.id,
				c.name,
				c.industry,
				COALESCE(sub.plan_name, c.subscription_plan) as plan_name,
				COALESCE(sub.status, '') as plan_status,
				COALESCE(sub.sort_time, c.created_at) as sort_time
			FROM companies c
			LEFT JOIN LATERAL (
				SELECT sp.name AS plan_name, s.status, COALESCE(s.current_period_start, s.updated_at, s.created_at) as sort_time
				FROM subscriptions s
				JOIN subscription_plans sp ON sp.id = s.plan_id
				WHERE s.company_id = c.id
				ORDER BY 
					CASE 
						WHEN s.status = 'active' THEN 0
						WHEN s.status = 'trialing' THEN 1
						WHEN s.status = 'past_due' THEN 2
						WHEN s.status = 'canceled' THEN 3
						ELSE 4
					END,
					sort_time DESC
				LIMIT 1
			) sub ON TRUE
			WHERE c.owner_user_id = u.id
			ORDER BY 
				CASE 
					WHEN COALESCE(sub.plan_name, c.subscription_plan, '') <> '' THEN 0
					ELSE 1
				END,
				COALESCE(sub.sort_time, c.created_at) DESC
			LIMIT 1
		) c ON TRUE
		WHERE u.id = $1
		ORDER BY COALESCE(c.created_at, '1970-01-01'::timestamp) DESC NULLS LAST
		LIMIT 1
	`, userID).Scan(&u.ID, &u.CompanyID, &u.Email, &u.Role, &u.Status, &storeName, &industry, &subscriptionPlan, &subscriptionStatus, &u.CreatedAt)
	if err != nil {
		appErr := errors.NewNotFoundError("User not found")
		h.respondError(w, appErr, r)
		return
	}
	u.StoreName = storeName.String
	u.Industry = industry.String
	if subscriptionPlan.Valid && subscriptionPlan.String != "" {
		u.SubscriptionPlan = subscriptionPlan.String
	}
	if subscriptionStatus.Valid && subscriptionStatus.String != "" {
		u.SubscriptionStatus = subscriptionStatus.String
	}

	// Log for debugging - check if company exists
	if !storeName.Valid {
		h.log.Warn("GetUser: No company found for user", "user_id", userID, "email", u.Email)
		// Try to find any company (even inactive) as fallback
		var fallbackStoreName, fallbackIndustry sql.NullString
		fallbackErr := h.db.Pool().QueryRow(ctx, `
			SELECT name, industry 
			FROM companies 
			WHERE owner_user_id = $1 
			ORDER BY created_at DESC 
			LIMIT 1
		`, userID).Scan(&fallbackStoreName, &fallbackIndustry)
		if fallbackErr == nil && fallbackStoreName.Valid {
			h.log.Info("GetUser: Found fallback company", "user_id", userID, "store_name", fallbackStoreName.String)
			u.StoreName = fallbackStoreName.String
			u.Industry = fallbackIndustry.String
		}
	}

	h.log.Info("GetUser result", "user_id", userID, "store_name", u.StoreName, "industry", u.Industry)

	h.respondJSON(w, http.StatusOK, u)
}

// UpdateUserRole updates a user's role
type UpdateUserRoleRequest struct {
	Role string `json:"role" validate:"required,oneof=user admin super_admin"`
}

func (h *AdminHandler) UpdateUserRole(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	// Extract user ID from path parameter
	userID := r.PathValue("id")
	if userID == "" {
		appErr := errors.NewValidationError("User ID is required", "")
		h.respondError(w, appErr, r)
		return
	}
	currentUserID := middleware.GetUserID(ctx)

	var req UpdateUserRoleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		appErr := errors.NewValidationError("Invalid request body", err.Error())
		h.respondError(w, appErr, r)
		return
	}

	// Prevent self-demotion from super_admin
	if userID == currentUserID {
		var currentRole string
		err := h.db.Pool().QueryRow(ctx, `SELECT COALESCE(role, 'user') FROM users WHERE id = $1`, currentUserID).Scan(&currentRole)
		if err == nil && currentRole == "super_admin" && req.Role != "super_admin" {
			appErr := errors.NewValidationError("Cannot change your own role from super_admin", "")
			h.respondError(w, appErr, r)
			return
		}
	}

	_, err := h.db.Pool().Exec(ctx, `
		UPDATE users
		SET role = $1
		WHERE id = $2
	`, req.Role, userID)
	if err != nil {
		appErr := errors.NewDatabaseError(err, "update user role")
		h.respondError(w, appErr, r)
		return
	}

	// Clear company-level caches if any (future hook)
	// No-op for now.

	// Log audit event
	if h.auditLogger != nil {
		h.auditLogger.LogResourceAction(ctx, r, audit.ActionUserRoleUpdated, "user", userID, map[string]interface{}{
			"new_role": req.Role,
		})
	}

	// Return updated user payload for UI immediate refresh
	var updated User
	err = h.db.Pool().QueryRow(ctx, `
		SELECT 
			u.id, 
			u.email, 
			COALESCE(u.role, 'user'), 
			COALESCE(u.status, 'active'), 
			COALESCE(c.name, '') as store_name,
			COALESCE(c.industry, '') as industry,
			u.created_at
		FROM users u
		LEFT JOIN companies c ON c.owner_user_id = u.id
		WHERE u.id = $1
		ORDER BY u.created_at DESC
		LIMIT 1
	`, userID).Scan(&updated.ID, &updated.Email, &updated.Role, &updated.Status, &updated.StoreName, &updated.Industry, &updated.CreatedAt)
	if err != nil {
		// fallback to simple message if fetch fails
		h.respondJSON(w, http.StatusOK, map[string]string{"message": "User role updated successfully"})
		return
	}

	h.respondJSON(w, http.StatusOK, map[string]string{"message": "User role updated successfully"})
}

// CreateUser creates a new user (admin only)
type CreateUserRequest struct {
	Email     string `json:"email" validate:"required,email"`
	Password  string `json:"password" validate:"required,min:6"`
	Role      string `json:"role" validate:"required,oneof=user admin super_admin"`
	StoreName string `json:"store_name" validate:"required"`
	Industry  string `json:"industry,omitempty"`
}

func (h *AdminHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		appErr := errors.NewValidationError("Invalid request body", err.Error())
		h.respondError(w, appErr, r)
		return
	}

	req.Email = strings.TrimSpace(strings.ToLower(req.Email))
	if req.Email == "" {
		appErr := errors.NewValidationError("Email is required", "")
		h.respondError(w, appErr, r)
		return
	}

	req.StoreName = strings.TrimSpace(req.StoreName)
	if req.StoreName == "" {
		appErr := errors.NewValidationError("Store name (Nama Toko) is required", "")
		h.respondError(w, appErr, r)
		return
	}

	req.Industry = strings.TrimSpace(req.Industry)

	// Check if email already exists
	var existingID string
	err := h.db.Pool().QueryRow(ctx, `SELECT id FROM users WHERE email = $1`, req.Email).Scan(&existingID)
	if err == nil {
		appErr := errors.NewConflictError("Email already registered", "")
		h.respondError(w, appErr, r)
		return
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		appErr := errors.NewInternalError(err, "Failed to hash password")
		h.respondError(w, appErr, r)
		return
	}

	// Use transaction to ensure both user and company are created atomically
	tx, err := h.db.Pool().Begin(ctx)
	if err != nil {
		appErr := errors.NewDatabaseError(err, "begin transaction")
		h.respondError(w, appErr, r)
		return
	}
	defer tx.Rollback(ctx)

	userID := uuid.New().String()
	_, err = tx.Exec(ctx, `
		INSERT INTO users (id, email, password_hash, role, status, created_at)
		VALUES ($1, $2, $3, $4, 'active', NOW())
	`, userID, req.Email, string(hashedPassword), req.Role)
	if err != nil {
		appErr := errors.NewDatabaseError(err, "create user")
		h.respondError(w, appErr, r)
		return
	}

	// Create default company for login compatibility
	companyID := uuid.New().String()
	companyName := req.StoreName
	if companyName == "" {
		companyName = "Default Company" // Fallback name
	}
	_, err = tx.Exec(ctx, `
		INSERT INTO companies (id, owner_user_id, name, industry, subscription_plan, status, created_at)
		VALUES ($1, $2, $3, $4, 'free', 'active', NOW())
	`, companyID, userID, companyName, req.Industry)
	if err != nil {
		appErr := errors.NewDatabaseError(err, "create company for user")
		h.respondError(w, appErr, r)
		return
	}

	// Commit transaction
	if err := tx.Commit(ctx); err != nil {
		appErr := errors.NewDatabaseError(err, "commit transaction")
		h.respondError(w, appErr, r)
		return
	}

	// Verify company was created (for debugging)
	var verifyCompanyID string
	verifyErr := h.db.Pool().QueryRow(ctx, `SELECT id FROM companies WHERE owner_user_id = $1 LIMIT 1`, userID).Scan(&verifyCompanyID)
	if verifyErr != nil {
		h.log.Error("Company verification failed after creation - this should not happen!", "user_id", userID, "company_id", companyID, "error", verifyErr.Error())
		// Don't fail the request, but log the error - login handler will auto-create if needed
	} else {
		h.log.Info("User and company created successfully", "user_id", userID, "company_id", verifyCompanyID, "email", req.Email)
	}

	// Log audit event (after successful commit)
	if h.auditLogger != nil {
		h.auditLogger.LogResourceAction(ctx, r, audit.ActionUserCreated, "user", userID, map[string]interface{}{
			"email": req.Email,
			"role":  req.Role,
		})
	}

	h.respondJSON(w, http.StatusCreated, map[string]string{
		"id":    userID,
		"email": req.Email,
		"role":  req.Role,
	})
}

// DeleteUser deletes a user (soft delete by setting role to inactive or hard delete)
func (h *AdminHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	// Extract user ID from path parameter
	userID := r.PathValue("id")
	if userID == "" {
		appErr := errors.NewValidationError("User ID is required", "")
		h.respondError(w, appErr, r)
		return
	}
	currentUserID := middleware.GetUserID(ctx)

	// Prevent self-deletion
	if userID == currentUserID {
		appErr := errors.NewValidationError("Cannot delete your own account", "")
		h.respondError(w, appErr, r)
		return
	}

	_, err := h.db.Pool().Exec(ctx, `DELETE FROM users WHERE id = $1`, userID)
	if err != nil {
		appErr := errors.NewDatabaseError(err, "delete user")
		h.respondError(w, appErr, r)
		return
	}

	// Log audit event
	if h.auditLogger != nil {
		h.auditLogger.LogResourceAction(ctx, r, audit.ActionUserDeleted, "user", userID, nil)
	}

	h.respondJSON(w, http.StatusOK, map[string]string{"message": "User deleted successfully"})
}

// UpdateUserStatus updates a user's status (suspend/unsuspend)
type UpdateUserStatusRequest struct {
	Status string `json:"status" validate:"required,oneof=active suspended"`
}

func (h *AdminHandler) UpdateUserStatus(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	// Extract user ID from path parameter
	userID := r.PathValue("id")
	if userID == "" {
		appErr := errors.NewValidationError("User ID is required", "")
		h.respondError(w, appErr, r)
		return
	}
	currentUserID := middleware.GetUserID(ctx)

	// Prevent self-suspension
	if userID == currentUserID {
		appErr := errors.NewValidationError("Cannot suspend your own account", "")
		h.respondError(w, appErr, r)
		return
	}

	var req UpdateUserStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		appErr := errors.NewValidationError("Invalid request body", err.Error())
		h.respondError(w, appErr, r)
		return
	}

	_, err := h.db.Pool().Exec(ctx, `
		UPDATE users
		SET status = $1
		WHERE id = $2
	`, req.Status, userID)
	if err != nil {
		appErr := errors.NewDatabaseError(err, "update user status")
		h.respondError(w, appErr, r)
		return
	}

	// Log audit event
	if h.auditLogger != nil {
		h.auditLogger.LogResourceAction(ctx, r, "user.status_updated", "user", userID, map[string]interface{}{
			"new_status": req.Status,
		})
	}

	h.respondJSON(w, http.StatusOK, map[string]string{"message": "User status updated successfully"})
}

// UpdateUser updates user data (email, etc.)
type UpdateUserRequest struct {
	Email     string `json:"email" validate:"required,email"`
	StoreName string `json:"store_name" validate:"required,max:255"`
	Industry  string `json:"industry,omitempty" validate:"max:100"`
	Role      string `json:"role,omitempty" validate:"omitempty,oneof=user admin super_admin"`
	Status    string `json:"status,omitempty" validate:"omitempty,oneof=active suspended"`
}

func (h *AdminHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	// Extract user ID from path parameter
	userID := r.PathValue("id")
	if userID == "" {
		appErr := errors.NewValidationError("User ID is required", "")
		h.respondError(w, appErr, r)
		return
	}

	var req UpdateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		appErr := errors.NewValidationError("Invalid request body", err.Error())
		h.respondError(w, appErr, r)
		return
	}

	req.Email = strings.TrimSpace(strings.ToLower(req.Email))
	if req.Email == "" {
		appErr := errors.NewValidationError("Email is required", "")
		h.respondError(w, appErr, r)
		return
	}
	req.StoreName = strings.TrimSpace(req.StoreName)
	req.Industry = strings.TrimSpace(req.Industry)

	// Load current role/status to preserve when not provided
	var currentRole, currentStatus string
	err := h.db.Pool().QueryRow(ctx, `SELECT COALESCE(role, 'user'), COALESCE(status, 'active') FROM users WHERE id = $1`, userID).
		Scan(&currentRole, &currentStatus)
	if err != nil {
		appErr := errors.NewNotFoundError("User not found")
		h.respondError(w, appErr, r)
		return
	}
	role := currentRole
	status := currentStatus
	if strings.TrimSpace(req.Role) != "" {
		role = req.Role
	}
	if strings.TrimSpace(req.Status) != "" {
		status = req.Status
	}

	// Check if email already exists for another user
	var existingID string
	err = h.db.Pool().QueryRow(ctx, `SELECT id FROM users WHERE email = $1 AND id != $2`, req.Email, userID).Scan(&existingID)
	if err == nil {
		appErr := errors.NewConflictError("Email already registered", "")
		h.respondError(w, appErr, r)
		return
	}

	_, err = h.db.Pool().Exec(ctx, `
		UPDATE users
		SET email = $1,
		    role = $2,
		    status = $3
		WHERE id = $4
	`, req.Email, role, status, userID)
	if err != nil {
		appErr := errors.NewDatabaseError(err, "update user")
		h.respondError(w, appErr, r)
		return
	}

	// Upsert company (store) data for the user
	var companyID string
	err = h.db.Pool().QueryRow(ctx, `SELECT id FROM companies WHERE owner_user_id = $1 LIMIT 1`, userID).Scan(&companyID)
	if err != nil {
		// Create company if missing
		companyID = uuid.New().String()
		_, cErr := h.db.Pool().Exec(ctx, `
			INSERT INTO companies (id, owner_user_id, name, industry, subscription_plan, status, created_at)
			VALUES ($1, $2, $3, $4, 'free', 'active', NOW())
		`, companyID, userID, req.StoreName, req.Industry)
		if cErr != nil {
			appErr := errors.NewDatabaseError(cErr, "create company for user")
			h.respondError(w, appErr, r)
			return
		}
	} else {
		// Update existing company
		_, cErr := h.db.Pool().Exec(ctx, `
			UPDATE companies
			SET name = $1, industry = $2
			WHERE id = $3
		`, req.StoreName, req.Industry, companyID)
		if cErr != nil {
			appErr := errors.NewDatabaseError(cErr, "update company for user")
			h.respondError(w, appErr, r)
			return
		}
	}

	// Log audit event
	if h.auditLogger != nil {
		h.auditLogger.LogResourceAction(ctx, r, "user.updated", "user", userID, map[string]interface{}{
			"email":      req.Email,
			"store_name": req.StoreName,
			"industry":   req.Industry,
			"role":       role,
			"status":     status,
		})
	}

	// Return updated user payload for UI refresh
	var updated User
	err = h.db.Pool().QueryRow(ctx, `
		SELECT 
			u.id, 
			u.email, 
			COALESCE(u.role, 'user'), 
			COALESCE(u.status, 'active'), 
			COALESCE(c.name, '') as store_name,
			COALESCE(c.industry, '') as industry,
			u.created_at
		FROM users u
		LEFT JOIN companies c ON c.owner_user_id = u.id
		WHERE u.id = $1
		ORDER BY u.created_at DESC
		LIMIT 1
	`, userID).Scan(&updated.ID, &updated.Email, &updated.Role, &updated.Status, &updated.StoreName, &updated.Industry, &updated.CreatedAt)
	if err != nil {
		h.respondJSON(w, http.StatusOK, map[string]string{"message": "User updated successfully"})
		return
	}

	h.respondJSON(w, http.StatusOK, updated)
}

// UpgradeUserSubscription manually upgrades a user's company subscription to pro
func (h *AdminHandler) UpgradeUserSubscription(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	// Extract user ID from path parameter
	userID := r.PathValue("id")
	if userID == "" {
		appErr := errors.NewValidationError("User ID is required", "")
		h.respondError(w, appErr, r)
		return
	}

	// Get user's company
	var companyID string
	err := h.db.Pool().QueryRow(ctx, `
		SELECT id FROM companies WHERE owner_user_id = $1 LIMIT 1
	`, userID).Scan(&companyID)
	if err != nil {
		appErr := errors.NewNotFoundError("User has no associated company")
		h.respondError(w, appErr, r)
		return
	}

	// Get pro plan ID
	var proPlanID string
	err = h.db.Pool().QueryRow(ctx, `
		SELECT id FROM subscription_plans WHERE name = 'pro' LIMIT 1
	`).Scan(&proPlanID)
	if err != nil {
		appErr := errors.NewNotFoundError("Pro plan not found")
		h.respondError(w, appErr, r)
		return
	}

	// Update company subscription_plan
	_, err = h.db.Pool().Exec(ctx, `
		UPDATE companies
		SET subscription_plan = 'pro'
		WHERE id = $1
	`, companyID)
	if err != nil {
		appErr := errors.NewDatabaseError(err, "update company subscription")
		h.respondError(w, appErr, r)
		return
	}

	// Create or update subscription record
	now := time.Now()
	periodEnd := now.AddDate(0, 1, 0) // 1 month from now

	// Check if subscription already exists
	var existingSubID string
	checkErr := h.db.Pool().QueryRow(ctx, `SELECT id FROM subscriptions WHERE company_id = $1`, companyID).Scan(&existingSubID)
	if checkErr == nil {
		// Update existing subscription
		_, err = h.db.Pool().Exec(ctx, `
			UPDATE subscriptions
			SET plan_id = $1, status = 'active', current_period_start = $2, current_period_end = $3, updated_at = NOW()
			WHERE company_id = $4
		`, proPlanID, now, periodEnd, companyID)
		if err != nil {
			appErr := errors.NewDatabaseError(err, "update subscription")
			h.respondError(w, appErr, r)
			return
		}
	} else {
		// Create new subscription
		subscriptionID := uuid.New().String()
		_, err = h.db.Pool().Exec(ctx, `
			INSERT INTO subscriptions (id, company_id, plan_id, status, current_period_start, current_period_end, created_at)
			VALUES ($1, $2, $3, 'active', $4, $5, NOW())
		`, subscriptionID, companyID, proPlanID, now, periodEnd)
		if err != nil {
			appErr := errors.NewDatabaseError(err, "create subscription")
			h.respondError(w, appErr, r)
			return
		}
	}

	// Log audit event
	if h.auditLogger != nil {
		h.auditLogger.LogResourceAction(ctx, r, "subscription.manually_upgraded", "subscription", companyID, map[string]interface{}{
			"user_id":    userID,
			"company_id": companyID,
			"plan":       "pro",
		})
	}

	h.respondJSON(w, http.StatusOK, map[string]string{"message": "Subscription upgraded to pro successfully"})
}

// Helper methods
func (h *AdminHandler) respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func (h *AdminHandler) respondError(w http.ResponseWriter, err *errors.AppError, r *http.Request) {
	errors.WriteJSONError(w, err, err.Code)
}
