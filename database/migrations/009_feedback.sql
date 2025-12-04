-- Bantuaku - Chat Feedback and Retrieval Diagnostics
-- Migration 009: User Feedback and RAG Logging
-- PostgreSQL 18
-- Dependencies: 003_add_chat_tables.sql (messages table)

-- Chat feedback (thumbs up/down)
CREATE TABLE IF NOT EXISTS chat_feedback (
    id VARCHAR(36) PRIMARY KEY,
    message_id VARCHAR(36) NOT NULL REFERENCES messages(id) ON DELETE CASCADE,
    user_id VARCHAR(36) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    feedback_type VARCHAR(20) NOT NULL,  -- 'positive', 'negative', 'neutral'
    comment TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Retrieval diagnostics (for RAG queries)
CREATE TABLE IF NOT EXISTS retrieval_diagnostics (
    id VARCHAR(36) PRIMARY KEY,
    message_id VARCHAR(36) REFERENCES messages(id) ON DELETE CASCADE,
    query_text TEXT NOT NULL,
    query_embedding_id VARCHAR(36) REFERENCES embeddings(id),
    chunks_retrieved INT NOT NULL,
    top_k INT NOT NULL,
    filters JSONB,  -- Applied filters (year, category, etc.)
    retrieval_time_ms INT,  -- Time taken for retrieval
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Retrieval results (links to chunks retrieved)
CREATE TABLE IF NOT EXISTS retrieval_results (
    id VARCHAR(36) PRIMARY KEY,
    diagnostic_id VARCHAR(36) NOT NULL REFERENCES retrieval_diagnostics(id) ON DELETE CASCADE,
    chunk_id VARCHAR(36) NOT NULL REFERENCES regulation_chunks(id) ON DELETE CASCADE,
    rank INT NOT NULL,  -- Rank in results (1 = most relevant)
    distance REAL NOT NULL,  -- Cosine distance
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Indexes
CREATE INDEX IF NOT EXISTS idx_chat_feedback_message_id ON chat_feedback(message_id);
CREATE INDEX IF NOT EXISTS idx_chat_feedback_user_id ON chat_feedback(user_id);
CREATE INDEX IF NOT EXISTS idx_chat_feedback_type ON chat_feedback(feedback_type);
CREATE INDEX IF NOT EXISTS idx_chat_feedback_created ON chat_feedback(created_at DESC);

CREATE INDEX IF NOT EXISTS idx_retrieval_diagnostics_message_id ON retrieval_diagnostics(message_id);
CREATE INDEX IF NOT EXISTS idx_retrieval_diagnostics_created ON retrieval_diagnostics(created_at DESC);

CREATE INDEX IF NOT EXISTS idx_retrieval_results_diagnostic_id ON retrieval_results(diagnostic_id);
CREATE INDEX IF NOT EXISTS idx_retrieval_results_chunk_id ON retrieval_results(chunk_id);
CREATE INDEX IF NOT EXISTS idx_retrieval_results_rank ON retrieval_results(diagnostic_id, rank);

-- Comments
COMMENT ON TABLE chat_feedback IS 'User feedback (thumbs up/down) for chat messages';
COMMENT ON TABLE retrieval_diagnostics IS 'Logging for RAG retrieval operations';
COMMENT ON TABLE retrieval_results IS 'Individual chunk results from RAG retrieval';

