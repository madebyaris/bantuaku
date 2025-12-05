-- Bantuaku - Regulations Knowledge Base Tables
-- Migration 005: RAG Foundation for Regulations
-- PostgreSQL 18
-- Dependencies: 004_pgvector_and_embeddings.sql

-- Regulations table (metadata)
CREATE TABLE IF NOT EXISTS regulations (
    id VARCHAR(36) PRIMARY KEY,
    title VARCHAR(500) NOT NULL,
    regulation_number VARCHAR(100),  -- e.g., "PP No. 12 Tahun 2023"
    year INT,
    category VARCHAR(100),  -- 'peraturan_pemerintah', 'undang_undang', etc.
    status VARCHAR(50) DEFAULT 'active',  -- 'active', 'revoked', 'amended'
    source_url TEXT NOT NULL,
    pdf_url TEXT,  -- URL to PDF (not stored, only referenced)
    published_date DATE,
    effective_date DATE,
    hash VARCHAR(64),  -- SHA-256 hash for deduplication
    version INT DEFAULT 1,  -- Track updates
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Regulation sources (track where we found it)
CREATE TABLE IF NOT EXISTS regulation_sources (
    id VARCHAR(36) PRIMARY KEY,
    regulation_id VARCHAR(36) NOT NULL REFERENCES regulations(id) ON DELETE CASCADE,
    source_type VARCHAR(50) NOT NULL DEFAULT 'peraturan_go_id',
    source_url TEXT NOT NULL,
    discovered_at TIMESTAMPTZ DEFAULT NOW()
);

-- Regulation sections (raw text from PDF)
CREATE TABLE IF NOT EXISTS regulation_sections (
    id VARCHAR(36) PRIMARY KEY,
    regulation_id VARCHAR(36) NOT NULL REFERENCES regulations(id) ON DELETE CASCADE,
    section_number VARCHAR(50),  -- e.g., "Pasal 1", "Bab II"
    section_title VARCHAR(255),
    content TEXT NOT NULL,  -- Raw extracted text
    page_number INT,
    order_index INT NOT NULL,  -- Order within regulation
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Regulation chunks (semantic chunks for RAG)
CREATE TABLE IF NOT EXISTS regulation_chunks (
    id VARCHAR(36) PRIMARY KEY,
    regulation_id VARCHAR(36) NOT NULL REFERENCES regulations(id) ON DELETE CASCADE,
    section_id VARCHAR(36) REFERENCES regulation_sections(id) ON DELETE CASCADE,
    chunk_text TEXT NOT NULL,
    chunk_index INT NOT NULL,  -- Order within section/regulation
    start_char_offset INT,  -- Character offset in original text
    end_char_offset INT,
    metadata JSONB,  -- Additional metadata (keywords, topics, etc.)
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Regulation embeddings (links to embeddings table)
CREATE TABLE IF NOT EXISTS regulation_embeddings (
    id VARCHAR(36) PRIMARY KEY,
    chunk_id VARCHAR(36) NOT NULL REFERENCES regulation_chunks(id) ON DELETE CASCADE,
    embedding_id VARCHAR(36) NOT NULL REFERENCES embeddings(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(chunk_id, embedding_id)
);

-- Indexes
CREATE INDEX IF NOT EXISTS idx_regulations_category ON regulations(category);
CREATE INDEX IF NOT EXISTS idx_regulations_year ON regulations(year);
CREATE INDEX IF NOT EXISTS idx_regulations_hash ON regulations(hash);
CREATE INDEX IF NOT EXISTS idx_regulations_status ON regulations(status);
CREATE INDEX IF NOT EXISTS idx_regulations_created ON regulations(created_at DESC);

CREATE INDEX IF NOT EXISTS idx_regulation_sources_regulation_id ON regulation_sources(regulation_id);

CREATE INDEX IF NOT EXISTS idx_regulation_sections_regulation_id ON regulation_sections(regulation_id);
CREATE INDEX IF NOT EXISTS idx_regulation_sections_order ON regulation_sections(regulation_id, order_index);

CREATE INDEX IF NOT EXISTS idx_regulation_chunks_regulation_id ON regulation_chunks(regulation_id);
CREATE INDEX IF NOT EXISTS idx_regulation_chunks_section_id ON regulation_chunks(section_id);

CREATE INDEX IF NOT EXISTS idx_regulation_embeddings_chunk_id ON regulation_embeddings(chunk_id);
CREATE INDEX IF NOT EXISTS idx_regulation_embeddings_embedding_id ON regulation_embeddings(embedding_id);

-- Comments
COMMENT ON TABLE regulations IS 'Regulation metadata from peraturan.go.id';
COMMENT ON TABLE regulation_sections IS 'Raw text sections extracted from PDFs (PDFs not stored)';
COMMENT ON TABLE regulation_chunks IS 'Semantic chunks for RAG retrieval';
COMMENT ON TABLE regulation_embeddings IS 'Links regulation chunks to vector embeddings';

