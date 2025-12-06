package embedding

import (
	"context"
	"fmt"
	"strings"

	"github.com/bantuaku/backend/logger"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pgvector/pgvector-go"
)

// RetrievalService handles KNN search for regulations
type RetrievalService struct {
	pool     *pgxpool.Pool
	embedder Embedder
	log      logger.Logger
}

// NewRetrievalService creates a new retrieval service
func NewRetrievalService(pool *pgxpool.Pool, embedder Embedder) *RetrievalService {
	return &RetrievalService{
		pool:     pool,
		embedder: embedder,
		log:      *logger.Default(),
	}
}

// Filters for search queries
type Filters struct {
	Year     *int
	Category *string
	Status   *string
}

// RetrievedChunk represents a retrieved chunk with metadata
type RetrievedChunk struct {
	ChunkID       string
	RegulationID  string
	RegulationTitle string
	RegulationNumber string
	Year          int
	Category      string
	ChunkText     string
	SectionNumber *string
	SectionTitle  *string
	SourceURL     string
	PDFURL        string
	Distance      float64 // Cosine distance
}

// Search performs KNN search for similar regulation chunks
func (r *RetrievalService) Search(ctx context.Context, query string, k int, filters Filters) ([]RetrievedChunk, error) {
	// Generate query embedding
	queryEmbedding, err := r.embedder.Embed(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to generate query embedding: %w", err)
	}

	// Convert to pgvector format
	queryVector := pgvector.NewVector(queryEmbedding)

	// Build query with filters
	querySQL := `
		SELECT 
			rc.id,
			rc.regulation_id,
			r.title,
			r.regulation_number,
			r.year,
			r.category,
			rc.chunk_text,
			rs.section_number,
			rs.section_title,
			r.source_url,
			r.pdf_url,
			1 - (e.embedding <=> $1::vector) AS distance
		FROM regulation_chunks rc
		JOIN regulation_embeddings re ON re.chunk_id = rc.id
		JOIN embeddings e ON e.id = re.embedding_id
		JOIN regulations r ON r.id = rc.regulation_id
		LEFT JOIN regulation_sections rs ON rs.id = rc.section_id
		WHERE 1=1
	`

	args := []interface{}{queryVector}
	argIndex := 2

	// Apply filters
	if filters.Year != nil {
		querySQL += fmt.Sprintf(" AND r.year = $%d", argIndex)
		args = append(args, *filters.Year)
		argIndex++
	}

	if filters.Category != nil {
		querySQL += fmt.Sprintf(" AND r.category = $%d", argIndex)
		args = append(args, *filters.Category)
		argIndex++
	}

	if filters.Status != nil {
		querySQL += fmt.Sprintf(" AND r.status = $%d", argIndex)
		args = append(args, *filters.Status)
		argIndex++
	}

	// Order by distance and limit
	querySQL += fmt.Sprintf(" ORDER BY e.embedding <=> $1::vector LIMIT $%d", argIndex)
	args = append(args, k)

	rows, err := r.pool.Query(ctx, querySQL, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to execute search query: %w", err)
	}
	defer rows.Close()

	var chunks []RetrievedChunk
	for rows.Next() {
		var chunk RetrievedChunk
		err := rows.Scan(
			&chunk.ChunkID,
			&chunk.RegulationID,
			&chunk.RegulationTitle,
			&chunk.RegulationNumber,
			&chunk.Year,
			&chunk.Category,
			&chunk.ChunkText,
			&chunk.SectionNumber,
			&chunk.SectionTitle,
			&chunk.SourceURL,
			&chunk.PDFURL,
			&chunk.Distance,
		)
		if err != nil {
			r.log.Warn("Failed to scan retrieved chunk", "error", err)
			continue
		}
		chunks = append(chunks, chunk)
	}

	return chunks, nil
}

// SearchByEmbedding performs KNN search using a pre-computed embedding
func (r *RetrievalService) SearchByEmbedding(ctx context.Context, queryEmbedding []float32, k int, filters Filters) ([]RetrievedChunk, error) {
	// Convert to pgvector format
	queryVector := pgvector.NewVector(queryEmbedding)

	// Build query with filters
	querySQL := `
		SELECT 
			rc.id,
			rc.regulation_id,
			r.title,
			r.regulation_number,
			r.year,
			r.category,
			rc.chunk_text,
			rs.section_number,
			rs.section_title,
			r.source_url,
			r.pdf_url,
			1 - (e.embedding <=> $1::vector) AS distance
		FROM regulation_chunks rc
		JOIN regulation_embeddings re ON re.chunk_id = rc.id
		JOIN embeddings e ON e.id = re.embedding_id
		JOIN regulations r ON r.id = rc.regulation_id
		LEFT JOIN regulation_sections rs ON rs.id = rc.section_id
		WHERE 1=1
	`

	args := []interface{}{queryVector}
	argIndex := 2

	// Apply filters
	if filters.Year != nil {
		querySQL += fmt.Sprintf(" AND r.year = $%d", argIndex)
		args = append(args, *filters.Year)
		argIndex++
	}

	if filters.Category != nil {
		querySQL += fmt.Sprintf(" AND r.category = $%d", argIndex)
		args = append(args, *filters.Category)
		argIndex++
	}

	if filters.Status != nil {
		querySQL += fmt.Sprintf(" AND r.status = $%d", argIndex)
		args = append(args, *filters.Status)
		argIndex++
	}

	// Order by distance and limit
	querySQL += fmt.Sprintf(" ORDER BY e.embedding <=> $1::vector LIMIT $%d", argIndex)
	args = append(args, k)

	rows, err := r.pool.Query(ctx, querySQL, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to execute search query: %w", err)
	}
	defer rows.Close()

	var chunks []RetrievedChunk
	for rows.Next() {
		var chunk RetrievedChunk
		err := rows.Scan(
			&chunk.ChunkID,
			&chunk.RegulationID,
			&chunk.RegulationTitle,
			&chunk.RegulationNumber,
			&chunk.Year,
			&chunk.Category,
			&chunk.ChunkText,
			&chunk.SectionNumber,
			&chunk.SectionTitle,
			&chunk.SourceURL,
			&chunk.PDFURL,
			&chunk.Distance,
		)
		if err != nil {
			r.log.Warn("Failed to scan retrieved chunk", "error", err)
			continue
		}
		chunks = append(chunks, chunk)
	}

	return chunks, nil
}

// FormatChunksForContext formats retrieved chunks for LLM context
func FormatChunksForContext(chunks []RetrievedChunk) string {
	var builder strings.Builder
	
	for i, chunk := range chunks {
		builder.WriteString(fmt.Sprintf("[%d] %s\n", i+1, chunk.RegulationTitle))
		if chunk.RegulationNumber != "" {
			builder.WriteString(fmt.Sprintf("   Regulation: %s\n", chunk.RegulationNumber))
		}
		if chunk.SectionNumber != nil {
			builder.WriteString(fmt.Sprintf("   Section: %s\n", *chunk.SectionNumber))
		}
		builder.WriteString(fmt.Sprintf("   Text: %s\n", chunk.ChunkText))
		builder.WriteString(fmt.Sprintf("   Source: %s\n\n", chunk.SourceURL))
	}
	
	return builder.String()
}

