package embedding

import (
	"context"
	"fmt"
	"time"

	"github.com/bantuaku/backend/logger"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pgvector/pgvector-go"
)

// Indexer handles vectorization and storage of embeddings
type Indexer struct {
	pool     *pgxpool.Pool
	embedder Embedder
	log      logger.Logger
	batchSize int
}

// NewIndexer creates a new indexer instance
func NewIndexer(pool *pgxpool.Pool, embedder Embedder) *Indexer {
	return &Indexer{
		pool:      pool,
		embedder: embedder,
		log:      logger.Default(),
		batchSize: 10, // Process 10 chunks at a time
	}
}

// ChunkInfo represents a chunk that needs to be embedded
type ChunkInfo struct {
	ID           string
	RegulationID string
	SectionID    *string
	Text         string
}

// IndexChunks indexes regulation chunks by generating embeddings
func (i *Indexer) IndexChunks(ctx context.Context, limit int) (int, error) {
	i.log.Info("Starting chunk indexing", "limit", limit)

	// Get chunks without embeddings
	query := `
		SELECT rc.id, rc.regulation_id, rc.section_id, rc.chunk_text
		FROM regulation_chunks rc
		LEFT JOIN regulation_embeddings re ON re.chunk_id = rc.id
		WHERE re.id IS NULL
		LIMIT $1
	`

	rows, err := i.pool.Query(ctx, query, limit)
	if err != nil {
		return 0, fmt.Errorf("failed to query chunks: %w", err)
	}
	defer rows.Close()

	var chunks []ChunkInfo
	for rows.Next() {
		var chunk ChunkInfo
		err := rows.Scan(&chunk.ID, &chunk.RegulationID, &chunk.SectionID, &chunk.Text)
		if err != nil {
			i.log.Warn("Failed to scan chunk", "error", err)
			continue
		}
		chunks = append(chunks, chunk)
	}

	if len(chunks) == 0 {
		i.log.Info("No chunks to index")
		return 0, nil
	}

	i.log.Info("Found chunks to index", "count", len(chunks))

	// Process in batches
	indexed := 0
	for start := 0; start < len(chunks); start += i.batchSize {
		end := start + i.batchSize
		if end > len(chunks) {
			end = len(chunks)
		}

		batch := chunks[start:end]
		count, err := i.indexBatch(ctx, batch)
		if err != nil {
			i.log.Warn("Failed to index batch", "error", err, "start", start, "end", end)
			continue
		}

		indexed += count
		i.log.Info("Indexed batch", "count", count, "total", indexed)

		// Rate limiting
		time.Sleep(100 * time.Millisecond)
	}

	i.log.Info("Chunk indexing completed", "indexed", indexed, "total", len(chunks))
	return indexed, nil
}

// indexBatch indexes a batch of chunks
func (i *Indexer) indexBatch(ctx context.Context, chunks []ChunkInfo) (int, error) {
	// Extract texts
	texts := make([]string, len(chunks))
	for j, chunk := range chunks {
		texts[j] = chunk.Text
	}

	// Generate embeddings
	embeddings, err := i.embedder.EmbedBatch(ctx, texts)
	if err != nil {
		return 0, fmt.Errorf("failed to generate embeddings: %w", err)
	}

	if len(embeddings) != len(chunks) {
		return 0, fmt.Errorf("mismatched embedding count: expected %d, got %d", len(chunks), len(embeddings))
	}

	// Store embeddings in database
	indexed := 0
	for j, chunk := range chunks {
		embedding := embeddings[j]
		
		// Convert to pgvector format
		vector := pgvector.NewVector(embedding)

		// Insert into embeddings table
		embeddingID := uuid.New().String()
		_, err := i.pool.Exec(ctx,
			`INSERT INTO embeddings (id, entity_type, entity_id, embedding, provider, model_version, created_at)
			 VALUES ($1, $2, $3, $4, $5, $6, NOW())`,
			embeddingID, "regulation_chunk", chunk.ID, vector, "kolosal", "v1",
		)
		if err != nil {
			i.log.Warn("Failed to insert embedding", "chunk_id", chunk.ID, "error", err)
			continue
		}

		// Link to regulation_embeddings table
		_, err = i.pool.Exec(ctx,
			`INSERT INTO regulation_embeddings (id, chunk_id, embedding_id, created_at)
			 VALUES ($1, $2, $3, NOW())`,
			uuid.New().String(), chunk.ID, embeddingID,
		)
		if err != nil {
			i.log.Warn("Failed to link regulation embedding", "chunk_id", chunk.ID, "error", err)
			// Continue - embedding is still stored
		}

		indexed++
	}

	return indexed, nil
}

// IndexChunk indexes a single chunk (for real-time indexing)
func (i *Indexer) IndexChunk(ctx context.Context, chunkID string, text string) error {
	// Generate embedding
	embedding, err := i.embedder.Embed(ctx, text)
	if err != nil {
		return fmt.Errorf("failed to generate embedding: %w", err)
	}

	// Convert to pgvector format
	vector := pgvector.NewVector(embedding)

	// Insert into embeddings table
	embeddingID := uuid.New().String()
	_, err = i.pool.Exec(ctx,
		`INSERT INTO embeddings (id, entity_type, entity_id, embedding, provider, model_version, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6, NOW())`,
		embeddingID, "regulation_chunk", chunkID, vector, "kolosal", "v1",
	)
	if err != nil {
		return fmt.Errorf("failed to insert embedding: %w", err)
	}

	// Link to regulation_embeddings table
	_, err = i.pool.Exec(ctx,
		`INSERT INTO regulation_embeddings (id, chunk_id, embedding_id, created_at)
		 VALUES ($1, $2, $3, NOW())`,
		uuid.New().String(), chunkID, embeddingID,
	)
	if err != nil {
		return fmt.Errorf("failed to link regulation embedding: %w", err)
	}

	return nil
}

