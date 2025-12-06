# ADR-001: Use pgvector Extension for Vector Storage

## Status
Accepted

## Context

We need to store vector embeddings for regulations text to enable semantic search (RAG). Options considered:

1. **PostgreSQL with pgvector extension** - Native extension, same database
2. **External vector DB** (Pinecone, Weaviate, Qdrant) - Separate service
3. **Hybrid approach** - PostgreSQL for metadata, external for vectors

## Decision

Use **PostgreSQL with pgvector extension** for vector storage.

## Rationale

### Advantages

1. **Simplicity**: Single database instance, no additional infrastructure
2. **Consistency**: ACID transactions for metadata + vectors together
3. **Cost**: No additional service costs (PostgreSQL already required)
4. **Performance**: pgvector with ivfflat index provides efficient KNN search
5. **Maturity**: pgvector is production-ready, widely adopted
6. **Integration**: Seamless with existing Go codebase (pgx driver supports pgvector)

### Trade-offs

1. **PostgreSQL version requirement**: Need PostgreSQL 11+ (we use 18, satisfied)
2. **Index tuning**: ivfflat requires tuning (lists parameter) for optimal performance
3. **Scalability**: For very large datasets (>100M vectors), external DBs may scale better
   - **Mitigation**: Our use case (regulations) is bounded (~10K-100K chunks), well within pgvector limits

## Implementation

### Database Migration

```sql
-- Enable pgvector extension
CREATE EXTENSION IF NOT EXISTS vector;

-- Create embeddings table
CREATE TABLE regulation_embeddings (
    id VARCHAR(36) PRIMARY KEY,
    chunk_id VARCHAR(36) NOT NULL REFERENCES regulation_chunks(id) ON DELETE CASCADE,
    embedding vector(1536), -- Adjust dimension based on provider
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Create ivfflat index for KNN search
CREATE INDEX idx_regulation_embeddings_vector 
ON regulation_embeddings 
USING ivfflat (embedding vector_cosine_ops)
WITH (lists = 100); -- Tune based on data size
```

### Go Code Integration

```go
import "github.com/pgvector/pgvector-go"

// Store embedding
embedding := pgvector.NewVector(embeddingSlice)
_, err := db.Exec(ctx, 
    "INSERT INTO regulation_embeddings (id, chunk_id, embedding) VALUES ($1, $2, $3)",
    id, chunkID, embedding)

// KNN search
rows, err := db.Query(ctx, `
    SELECT chunk_id, embedding <=> $1::vector AS distance
    FROM regulation_embeddings
    ORDER BY distance
    LIMIT $2
`, queryEmbedding, k)
```

## Alternatives Considered

### External Vector DB (Pinecone/Weaviate)

**Pros:**
- Optimized for vector search at scale
- Managed service (less ops)

**Cons:**
- Additional infrastructure cost
- Network latency for queries
- Data consistency challenges (separate from PostgreSQL)
- Vendor lock-in

**Decision**: Not chosen due to complexity and cost for our scale.

### Hybrid Approach

**Pros:**
- Best of both worlds

**Cons:**
- Complex data synchronization
- Higher operational overhead

**Decision**: Not chosen - unnecessary complexity for our use case.

## Consequences

### Positive

- Simplified architecture (one database)
- Lower operational overhead
- Better data consistency guarantees
- Easier local development setup

### Negative

- PostgreSQL must support pgvector (version requirement)
- Index tuning required for optimal performance
- Vector dimension fixed at table creation (requires migration to change)

### Mitigations

- Use PostgreSQL 18 (pgvector fully supported)
- Benchmark and tune ivfflat lists parameter during Phase 3
- Choose embedding dimension carefully (1536 for most providers)

## References

- [pgvector GitHub](https://github.com/pgvector/pgvector)
- [pgvector Go driver](https://github.com/pgvector/pgvector-go)
- [ivfflat index documentation](https://github.com/pgvector/pgvector#ivfflat)

