package audit

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/bantuaku/backend/services/storage"
)

// Logger provides centralized audit logging functionality
type Logger struct {
	db *storage.Postgres
}

// NewLogger creates a new audit logger
func NewLogger(db *storage.Postgres) *Logger {
	return &Logger{
		db: db,
	}
}

// Log records an audit event
func (l *Logger) Log(ctx context.Context, req *http.Request, action, resourceType, resourceID string, metadata map[string]interface{}) {
	// Extract user ID from context (set by Auth middleware)
	userID := extractUserID(ctx)
	companyID := extractCompanyID(ctx)

	// Extract IP address and user agent from request
	ipAddress := extractIPAddress(req)
	userAgent := req.Header.Get("User-Agent")

	// Marshal metadata to JSON
	metadataJSON := "{}"
	if metadata != nil {
		if b, err := json.Marshal(metadata); err == nil {
			metadataJSON = string(b)
		}
	}

	// Insert audit log (non-blocking - don't fail the request if audit logging fails)
	_, err := l.db.Pool().Exec(ctx, `
		INSERT INTO audit_logs (user_id, company_id, action, resource_type, resource_id, ip_address, user_agent, metadata)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`, userID, companyID, action, resourceType, resourceID, ipAddress, userAgent, metadataJSON)
	if err != nil {
		// Log error but don't fail the request
		// In production, you might want to use a background worker for audit logs
		_ = err
	}
}

// LogAction is a convenience method for logging actions without resource details
func (l *Logger) LogAction(ctx context.Context, req *http.Request, action string, metadata map[string]interface{}) {
	l.Log(ctx, req, action, "", "", metadata)
}

// LogResourceAction logs an action on a specific resource
func (l *Logger) LogResourceAction(ctx context.Context, req *http.Request, action, resourceType, resourceID string, metadata map[string]interface{}) {
	l.Log(ctx, req, action, resourceType, resourceID, metadata)
}

// extractUserID extracts user ID from context
func extractUserID(ctx context.Context) *string {
	if userID, ok := ctx.Value("user_id").(string); ok && userID != "" {
		return &userID
	}
	return nil
}

// extractCompanyID extracts company ID from context
func extractCompanyID(ctx context.Context) *string {
	if companyID, ok := ctx.Value("company_id").(string); ok && companyID != "" {
		return &companyID
	}
	if companyID, ok := ctx.Value("store_id").(string); ok && companyID != "" {
		return &companyID
	}
	return nil
}

// extractIPAddress extracts the real client IP from request
func extractIPAddress(r *http.Request) *string {
	// Check X-Forwarded-For header (for reverse proxies)
	xForwardedFor := r.Header.Get("X-Forwarded-For")
	if xForwardedFor != "" {
		ips := splitIPs(xForwardedFor)
		if len(ips) > 0 {
			ip := ips[0]
			return &ip
		}
	}

	// Check X-Real-IP header
	xRealIP := r.Header.Get("X-Real-IP")
	if xRealIP != "" {
		return &xRealIP
	}

	// Fall back to RemoteAddr
	if r.RemoteAddr != "" {
		// Remove port if present
		ip := r.RemoteAddr
		if idx := indexOf(ip, ":"); idx > 0 {
			ip = ip[:idx]
		}
		return &ip
	}

	return nil
}

// splitIPs splits comma-separated IP addresses
func splitIPs(ips string) []string {
	var result []string
	for _, ip := range split(ips, ",") {
		ip = trimSpace(ip)
		if ip != "" {
			result = append(result, ip)
		}
	}
	return result
}

// Helper functions (avoid importing strings package for simple operations)
func split(s, sep string) []string {
	var result []string
	start := 0
	for i := 0; i <= len(s)-len(sep); i++ {
		if s[i:i+len(sep)] == sep {
			result = append(result, s[start:i])
			start = i + len(sep)
		}
	}
	result = append(result, s[start:])
	return result
}

func trimSpace(s string) string {
	start := 0
	end := len(s)
	for start < end && (s[start] == ' ' || s[start] == '\t' || s[start] == '\n' || s[start] == '\r') {
		start++
	}
	for end > start && (s[end-1] == ' ' || s[end-1] == '\t' || s[end-1] == '\n' || s[end-1] == '\r') {
		end--
	}
	return s[start:end]
}

func indexOf(s string, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

// Sensitive actions that should always be audited
const (
	ActionUserCreated         = "user.created"
	ActionUserUpdated         = "user.updated"
	ActionUserDeleted         = "user.deleted"
	ActionUserRoleUpdated     = "user.role_updated"
	ActionSubscriptionCreated = "subscription.created"
	ActionSubscriptionUpdated = "subscription.updated"
	ActionSubscriptionDeleted = "subscription.deleted"
	ActionRegulationScraped   = "regulation.scraped"
	ActionTrendsIngested      = "trends.ingested"
	ActionForecastGenerated   = "forecast.generated"
	ActionAdminLogin          = "admin.login"
	ActionAdminLogout         = "admin.logout"
)
