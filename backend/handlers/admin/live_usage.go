package admin

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/bantuaku/backend/errors"
	"github.com/bantuaku/backend/logger"
	"github.com/golang-jwt/jwt/v5"
)

type liveUsagePayload struct {
	ChatMessages   int       `json:"chat_messages"`
	FileUploads    int       `json:"file_uploads"`
	RAGQueries     int       `json:"rag_queries"`
	TotalTokens    int       `json:"total_tokens"`
	RefreshedAtUTC time.Time `json:"refreshed_at_utc"`
}

// LiveUsageSSE streams live counters using Server-Sent Events.
// Auth: accepts Bearer token or ?token= query param; requires admin/super_admin.
func (h *AdminHandler) LiveUsageSSE(w http.ResponseWriter, r *http.Request) {
	tokenStr := extractToken(r)
	if tokenStr == "" {
		h.respondError(w, errors.NewUnauthorizedError("missing token"), r)
		return
	}

	claims := jwt.MapClaims{}
	_, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(h.jwtSecret), nil
	})
	if err != nil {
		h.respondError(w, errors.NewUnauthorizedError("invalid token"), r)
		return
	}

	role, _ := claims["role"].(string)
	if role != "admin" && role != "super_admin" {
		h.respondError(w, errors.NewUnauthorizedError("forbidden"), r)
		return
	}

	// SSE headers
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	flusher, ok := w.(http.Flusher)
	if !ok {
		h.respondError(w, errors.NewInternalError(fmt.Errorf("streaming unsupported"), "streaming unsupported"), r)
		return
	}

	hub := ensureLiveHub(h.db)
	sub, cancel := hub.subscribe()
	defer cancel()

	// send initial heartbeat
	fmt.Fprintf(w, "event: ping\ndata: {}\n\n")
	flusher.Flush()

	for {
		select {
		case <-r.Context().Done():
			return
		case msg, ok := <-sub:
			if !ok {
				return
			}
			fmt.Fprintf(w, "data: %s\n\n", msg)
			flusher.Flush()
		}
	}
}

func (h *AdminHandler) fetchLiveUsage(ctx context.Context) (*liveUsagePayload, error) {
	today := time.Now().UTC().Format("2006-01-02")
	var chatCount, uploadCount, ragCount, tokens int
	log := logger.Default()

	// Activity aggregates (daily)
	if err := h.db.Pool().QueryRow(ctx, `
		SELECT COALESCE(SUM(count),0) FROM activity_aggregates 
		WHERE date = $1 AND action_type = 'activity.chat.message'
	`, today).Scan(&chatCount); err != nil {
		if payload, fallbackErr := gracefulPgFallback(err, log); fallbackErr == nil {
			return payload, nil
		}
		return nil, err
	}
	if err := h.db.Pool().QueryRow(ctx, `
		SELECT COALESCE(SUM(count),0) FROM activity_aggregates 
		WHERE date = $1 AND action_type = 'activity.file.uploaded'
	`, today).Scan(&uploadCount); err != nil {
		if payload, fallbackErr := gracefulPgFallback(err, log); fallbackErr == nil {
			return payload, nil
		}
		return nil, err
	}
	if err := h.db.Pool().QueryRow(ctx, `
		SELECT COALESCE(SUM(count),0) FROM activity_aggregates 
		WHERE date = $1 AND action_type = 'activity.rag.query'
	`, today).Scan(&ragCount); err != nil {
		if payload, fallbackErr := gracefulPgFallback(err, log); fallbackErr == nil {
			return payload, nil
		}
		return nil, err
	}

	// Token aggregates (daily)
	if err := h.db.Pool().QueryRow(ctx, `
		SELECT COALESCE(SUM(total_tokens),0) FROM token_usage_aggregates 
		WHERE date = $1
	`, today).Scan(&tokens); err != nil {
		if payload, fallbackErr := gracefulPgFallback(err, log); fallbackErr == nil {
			return payload, nil
		}
		return nil, err
	}

	return &liveUsagePayload{
		ChatMessages:   chatCount,
		FileUploads:    uploadCount,
		RAGQueries:     ragCount,
		TotalTokens:    tokens,
		RefreshedAtUTC: time.Now().UTC(),
	}, nil
}

func extractToken(r *http.Request) string {
	// Header Bearer
	auth := r.Header.Get("Authorization")
	if strings.HasPrefix(auth, "Bearer ") {
		return strings.TrimSpace(strings.TrimPrefix(auth, "Bearer "))
	}
	// Query param
	if t := r.URL.Query().Get("token"); t != "" {
		return t
	}
	return ""
}

// gracefulPgFallback converts missing-table errors into zeroed payloads for live usage.
func gracefulPgFallback(err error, log *logger.Logger) (*liveUsagePayload, error) {
	errStr := err.Error()
	// Check for undefined_table (42P01) or undefined_column errors
	if strings.Contains(errStr, "does not exist") || strings.Contains(errStr, "42P01") {
		log.Warn("Live usage aggregates not found (table/column missing); returning zeros", "error", errStr)
		return &liveUsagePayload{
			ChatMessages:   0,
			FileUploads:    0,
			RAGQueries:     0,
			TotalTokens:    0,
			RefreshedAtUTC: time.Now().UTC(),
		}, nil
	}
	return nil, err
}
