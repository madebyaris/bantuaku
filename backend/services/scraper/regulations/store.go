package regulations

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"github.com/bantuaku/backend/logger"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Store handles database persistence with deduplication
type Store struct {
	pool *pgxpool.Pool
	log  logger.Logger
}

// NewStore creates a new store instance
func NewStore(pool *pgxpool.Pool) *Store {
	return &Store{
		pool: pool,
		log:  *logger.Default(),
	}
}

// UpsertRegulation upserts a regulation with deduplication
func (s *Store) UpsertRegulation(ctx context.Context, reg *Regulation) (string, error) {
	// Generate hash for deduplication
	hash := s.generateHash(reg)

	// Check if regulation exists
	var existingID string
	err := s.pool.QueryRow(ctx,
		"SELECT id FROM regulations WHERE hash = $1",
		hash,
	).Scan(&existingID)

	if err == nil {
		// Regulation exists, update version
		_, err = s.pool.Exec(ctx,
			`UPDATE regulations 
			 SET title = $1, regulation_number = $2, year = $3, category = $4,
			     source_url = $5, pdf_url = $6, published_date = $7, effective_date = $8,
			     version = version + 1, updated_at = NOW()
			 WHERE id = $9`,
			reg.Title, reg.RegulationNumber, reg.Year, reg.Category,
			reg.SourceURL, reg.PDFURL, reg.PublishedDate, reg.EffectiveDate,
			existingID,
		)
		if err != nil {
			return "", fmt.Errorf("failed to update regulation: %w", err)
		}

		s.log.Info("Updated existing regulation", "id", existingID, "hash", hash)
		return existingID, nil
	}

	// Check if error is "no rows" (regulation doesn't exist)
	if err.Error() == "no rows in result set" {
		// Continue to create new regulation
	} else {
		return "", fmt.Errorf("failed to check existing regulation: %w", err)
	}

	// Create new regulation
	id := uuid.New().String()
	_, err = s.pool.Exec(ctx,
		`INSERT INTO regulations 
		 (id, title, regulation_number, year, category, status, source_url, pdf_url, 
		  published_date, effective_date, hash, version, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, NOW(), NOW())`,
		id, reg.Title, reg.RegulationNumber, reg.Year, reg.Category, "active",
		reg.SourceURL, reg.PDFURL, reg.PublishedDate, reg.EffectiveDate, hash, 1,
	)
	if err != nil {
		return "", fmt.Errorf("failed to insert regulation: %w", err)
	}

	// Insert source
	_, err = s.pool.Exec(ctx,
		`INSERT INTO regulation_sources (id, regulation_id, source_type, source_url, discovered_at)
		 VALUES ($1, $2, $3, $4, NOW())`,
		uuid.New().String(), id, "peraturan_go_id", reg.SourceURL,
	)
	if err != nil {
		s.log.Warn("Failed to insert regulation source", "error", err)
	}

	s.log.Info("Created new regulation", "id", id, "hash", hash)
	return id, nil
}

// StoreSection stores a regulation section
func (s *Store) StoreSection(ctx context.Context, regulationID string, sectionNumber string, sectionTitle string, content string, pageNumber int, orderIndex int) (string, error) {
	id := uuid.New().String()
	_, err := s.pool.Exec(ctx,
		`INSERT INTO regulation_sections 
		 (id, regulation_id, section_number, section_title, content, page_number, order_index, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, NOW())`,
		id, regulationID, sectionNumber, sectionTitle, content, pageNumber, orderIndex,
	)
	if err != nil {
		return "", fmt.Errorf("failed to insert section: %w", err)
	}
	return id, nil
}

// StoreChunk stores a regulation chunk
func (s *Store) StoreChunk(ctx context.Context, regulationID string, sectionID *string, chunk Chunk) (string, error) {
	id := uuid.New().String()
	_, err := s.pool.Exec(ctx,
		`INSERT INTO regulation_chunks 
		 (id, regulation_id, section_id, chunk_text, chunk_index, start_char_offset, end_char_offset, metadata, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW())`,
		id, regulationID, sectionID, chunk.Text, chunk.Index,
		chunk.StartCharOffset, chunk.EndCharOffset, nil,
	)
	if err != nil {
		return "", fmt.Errorf("failed to insert chunk: %w", err)
	}
	return id, nil
}

// generateHash generates SHA-256 hash for deduplication
func (s *Store) generateHash(reg *Regulation) string {
	// Hash based on regulation number and year (most stable identifiers)
	hashInput := fmt.Sprintf("%s|%d|%s", reg.RegulationNumber, reg.Year, reg.PDFURL)
	hash := sha256.Sum256([]byte(hashInput))
	return hex.EncodeToString(hash[:])
}

// GetRegulationByHash retrieves regulation by hash
func (s *Store) GetRegulationByHash(ctx context.Context, hash string) (string, error) {
	var id string
	err := s.pool.QueryRow(ctx,
		"SELECT id FROM regulations WHERE hash = $1",
		hash,
	).Scan(&id)
	return id, err
}

// IsRegulationProcessed checks if regulation has been fully processed (has chunks)
func (s *Store) IsRegulationProcessed(ctx context.Context, regulationID string) (bool, error) {
	var count int
	err := s.pool.QueryRow(ctx,
		"SELECT COUNT(*) FROM regulation_chunks WHERE regulation_id = $1",
		regulationID,
	).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

