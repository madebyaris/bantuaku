-- Bantuaku - Add pgvector Extension and Embeddings Schema
-- Migration 004: Vector Database Foundation
-- PostgreSQL 18

-- Enable pgvector extension
CREATE EXTENSION IF NOT EXISTS vector;

-- Create embeddings table (generic, can be used for any text embeddings)
CREATE TABLE IF NOT EXISTS embeddings (
    id VARCHAR(36) PRIMARY KEY,
    entity_type VARCHAR(50) NOT NULL,  -- 'regulation_chunk', 'product', etc.
    entity_id VARCHAR(36) NOT NULL,
    embedding vector(1536) NOT NULL,  -- Dimension based on provider (Kolosal.ai = 1536)
    provider VARCHAR(50) NOT NULL DEFAULT 'kolosal',  -- 'kolosal', 'openai', 'cohere'
    model_version VARCHAR(100),  -- Track model version for future migrations
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Index for KNN search (ivfflat with cosine similarity)
-- Note: lists parameter should be tuned based on data size (~rows/1000)
CREATE INDEX IF NOT EXISTS idx_embeddings_vector 
ON embeddings 
USING ivfflat (embedding vector_cosine_ops)
WITH (lists = 100);

-- Index for entity lookups
CREATE INDEX IF NOT EXISTS idx_embeddings_entity ON embeddings(entity_type, entity_id);

-- Comments
COMMENT ON TABLE embeddings IS 'Vector embeddings for semantic search';
COMMENT ON COLUMN embeddings.embedding IS 'Vector embedding (dimension varies by provider, default 1536 for Kolosal.ai)';
-- Note: ivfflat index lists parameter should be tuned based on data size (recommended: rows/1000)

