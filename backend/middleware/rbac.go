package middleware

import (
	"context"
	"net/http"

	apperrors "github.com/bantuaku/backend/errors"
	"github.com/bantuaku/backend/logger"
)

// RequireRole enforces role-based access control
func RequireRole(allowedRoles ...string) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			requestID, _ := r.Context().Value(RequestIDKey).(string)
			log := logger.With("request_id", requestID)

			role, ok := r.Context().Value(RoleKey).(string)
			if !ok || role == "" {
				role = "user" // Default role
			}

			// Check if user's role is in allowed roles
			allowed := false
			for _, allowedRole := range allowedRoles {
				if role == allowedRole {
					allowed = true
					break
				}
			}

			if !allowed {
				err := apperrors.NewForbiddenError("Insufficient permissions")
				log.Error("RBAC check failed", "user_role", role, "required_roles", allowedRoles)
				log.LogError(err, "RBAC check failed", r.Context())
				apperrors.WriteJSONError(w, err, err.Code)
				return
			}

			log.Debug("RBAC check passed", "user_role", role, "required_roles", allowedRoles)
			next.ServeHTTP(w, r)
		}
	}
}

// RequireAdmin is a convenience function for admin-only endpoints
func RequireAdmin(next http.HandlerFunc) http.HandlerFunc {
	return RequireRole("admin", "super_admin")(next)
}

// RequireSuperAdmin is a convenience function for super_admin-only endpoints
func RequireSuperAdmin(next http.HandlerFunc) http.HandlerFunc {
	return RequireRole("super_admin")(next)
}

// GetRole extracts role from context
func GetRole(ctx context.Context) string {
	role, _ := ctx.Value(RoleKey).(string)
	if role == "" {
		return "user"
	}
	return role
}

