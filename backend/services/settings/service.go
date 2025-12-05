package settings

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/bantuaku/backend/services/storage"
	"github.com/jackc/pgx/v5"
)

// Service handles application settings
type Service struct {
	db *storage.Postgres
}

// NewService creates a new settings service
func NewService(db *storage.Postgres) *Service {
	return &Service{
		db: db,
	}
}

// GetSetting retrieves a setting value by key
// Returns the JSON string value, or empty string if not found
func (s *Service) GetSetting(ctx context.Context, key string) (string, error) {
	// #region agent log
	if f, err := os.OpenFile("/Volumes/app/hackathon/imphxkolosal/bantuaku/.cursor/debug.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644); err == nil {
		fmt.Fprintf(f, `{"sessionId":"debug-session","runId":"run1","hypothesisId":"A","location":"settings/service.go:GetSetting:entry","message":"GetSetting called","data":{"key":"%s","dbNil":%t,"poolNil":%t},"timestamp":%d}`+"\n", key, s.db == nil, s.db != nil && s.db.Pool() == nil, 0)
		f.Close()
	}
	// #endregion
	if s.db == nil || s.db.Pool() == nil {
		// #region agent log
		if f, err := os.OpenFile("/Volumes/app/hackathon/imphxkolosal/bantuaku/.cursor/debug.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644); err == nil {
			fmt.Fprintf(f, `{"sessionId":"debug-session","runId":"run1","hypothesisId":"D","location":"settings/service.go:GetSetting:dbCheck","message":"Database connection unavailable","data":{"dbNil":%t},"timestamp":%d}`+"\n", s.db == nil, 0)
			f.Close()
		}
		// #endregion
		return "", fmt.Errorf("database connection not available")
	}

	var value json.RawMessage
	err := s.db.Pool().QueryRow(ctx, `
		SELECT value
		FROM settings
		WHERE key = $1
	`, key).Scan(&value)

	if err != nil {
		// #region agent log
		if f, err2 := os.OpenFile("/Volumes/app/hackathon/imphxkolosal/bantuaku/.cursor/debug.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644); err2 == nil {
			fmt.Fprintf(f, `{"sessionId":"debug-session","runId":"run1","hypothesisId":"A","location":"settings/service.go:GetSetting:queryError","message":"Query error occurred","data":{"key":"%s","error":"%s","isNoRows":%t},"timestamp":%d}`+"\n", key, err.Error(), err == pgx.ErrNoRows, 0)
			f.Close()
		}
		// #endregion
		// Return empty string only if setting not found (no rows)
		// Propagate other errors (like table doesn't exist)
		if err == pgx.ErrNoRows {
			return "", nil
		}
		// Real error (table doesn't exist, connection issue, etc.)
		return "", fmt.Errorf("failed to query setting: %w", err)
	}

	// #region agent log
	if f, err := os.OpenFile("/Volumes/app/hackathon/imphxkolosal/bantuaku/.cursor/debug.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644); err == nil {
		fmt.Fprintf(f, `{"sessionId":"debug-session","runId":"run1","hypothesisId":"A","location":"settings/service.go:GetSetting:success","message":"GetSetting success","data":{"key":"%s","valueLength":%d},"timestamp":%d}`+"\n", key, len(string(value)), 0)
		f.Close()
	}
	// #endregion
	return string(value), nil
}

// SetSetting creates or updates a setting
func (s *Service) SetSetting(ctx context.Context, key string, value interface{}) error {
	if s.db == nil || s.db.Pool() == nil {
		return fmt.Errorf("database connection not available")
	}

	valueJSON, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal value: %w", err)
	}

	_, err = s.db.Pool().Exec(ctx, `
		INSERT INTO settings (id, key, value, updated_at)
		VALUES (gen_random_uuid(), $1, $2::jsonb, NOW())
		ON CONFLICT (key) 
		DO UPDATE SET value = $2::jsonb, updated_at = NOW()
	`, key, string(valueJSON))

	if err != nil {
		return fmt.Errorf("failed to save setting: %w", err)
	}

	return nil
}

// GetAllSettings retrieves all settings (for admin UI)
func (s *Service) GetAllSettings(ctx context.Context) (map[string]interface{}, error) {
	if s.db == nil || s.db.Pool() == nil {
		return nil, fmt.Errorf("database connection not available")
	}

	rows, err := s.db.Pool().Query(ctx, `
		SELECT key, value
		FROM settings
		ORDER BY key
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to query settings: %w", err)
	}
	defer rows.Close()

	settings := make(map[string]interface{})
	for rows.Next() {
		var key string
		var value json.RawMessage

		if err := rows.Scan(&key, &value); err != nil {
			continue
		}

		var valueObj interface{}
		if err := json.Unmarshal(value, &valueObj); err != nil {
			continue
		}

		settings[key] = valueObj
	}

	return settings, nil
}
