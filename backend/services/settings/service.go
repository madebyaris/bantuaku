package settings

import (
	"context"
	"encoding/json"
	"fmt"

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
	if s.db == nil || s.db.Pool() == nil {
		return "", fmt.Errorf("database connection not available")
	}

	var value json.RawMessage
	err := s.db.Pool().QueryRow(ctx, `
		SELECT value
		FROM settings
		WHERE key = $1
	`, key).Scan(&value)

	if err != nil {
		// Return empty string only if setting not found (no rows)
		// Propagate other errors (like table doesn't exist)
		if err == pgx.ErrNoRows {
			return "", nil
		}
		// Real error (table doesn't exist, connection issue, etc.)
		return "", fmt.Errorf("failed to query setting: %w", err)
	}

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
			return nil, fmt.Errorf("failed to scan setting row: %w", err)
		}

		var valueObj interface{}
		if err := json.Unmarshal(value, &valueObj); err != nil {
			return nil, fmt.Errorf("failed to unmarshal setting %q: %w", key, err)
		}

		settings[key] = valueObj
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating settings: %w", err)
	}

	return settings, nil
}
