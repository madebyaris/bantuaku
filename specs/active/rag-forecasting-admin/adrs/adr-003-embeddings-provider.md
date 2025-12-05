# ADR-003: Use Configurable Embedding Provider Interface

## Status
Accepted

## Context

We need to generate embeddings for regulation text chunks. Options:

1. **Vendor API** (OpenAI, Kolosal.ai, Cohere) - Managed service
2. **Self-hosted** (sentence-transformers) - Open source, local
3. **Hybrid** - Start with vendor, migrate to self-hosted later

## Decision

Use **configurable vendor API** with abstraction interface, starting with Kolosal.ai (already in use).

## Rationale

### Advantages

1. **Simplicity**: No ML infrastructure to manage
2. **Quality**: Vendor APIs provide high-quality embeddings
3. **Scalability**: Vendor handles scaling
4. **Flexibility**: Easy to switch providers via config
5. **Cost**: Pay-per-use, no idle infrastructure costs
6. **Indonesian Language**: Kolosal.ai optimized for Indonesian text

### Trade-offs

1. **API Dependency**: Requires internet, API rate limits
   - **Mitigation**: Batch processing, retry logic, caching
2. **Cost at Scale**: Per-request pricing can add up
   - **Mitigation**: Cache embeddings, only embed new chunks
3. **Data Privacy**: Text sent to external API
   - **Mitigation**: Regulations are public data, acceptable

## Implementation

### Interface Design

```go
// backend/services/embedding/interface.go
package embedding

type Embedder interface {
    Embed(ctx context.Context, text string) ([]float32, error)
    EmbedBatch(ctx context.Context, texts []string) ([][]float32, error)
    Dimension() int
}

type Provider string

const (
    ProviderKolosal Provider = "kolosal"
    ProviderOpenAI  Provider = "openai"
    ProviderCohere  Provider = "cohere"
)
```

### Provider Implementation

```go
// backend/services/embedding/kolosal.go
type KolosalEmbedder struct {
    client *kolosal.Client
}

func (k *KolosalEmbedder) Embed(ctx context.Context, text string) ([]float32, error) {
    // Call Kolosal.ai embedding API
    // Return vector
}

func (k *KolosalEmbedder) Dimension() int {
    return 1536 // Kolosal.ai dimension
}
```

### Factory Pattern

```go
// backend/services/embedding/factory.go
func NewEmbedder(cfg *config.Config) (Embedder, error) {
    provider := Provider(cfg.EmbeddingProvider)
    switch provider {
    case ProviderKolosal:
        return NewKolosalEmbedder(cfg.KolosalAPIKey)
    case ProviderOpenAI:
        return NewOpenAIEmbedder(cfg.OpenAIAPIKey)
    default:
        return nil, fmt.Errorf("unknown provider: %s", provider)
    }
}
```

### Configuration

```go
// backend/config/config.go
type Config struct {
    // ... existing fields
    EmbeddingProvider string // "kolosal", "openai", "cohere"
    EmbeddingAPIKey  string // Provider-specific API key
}
```

### Environment Variables

```bash
EMBEDDING_PROVIDER=kolosal
EMBEDDING_API_KEY=${KOLOSAL_API_KEY}
```

## Alternatives Considered

### Self-Hosted (sentence-transformers)

**Pros:**
- No API costs
- Full data privacy
- No rate limits

**Cons:**
- Requires ML infrastructure (GPU recommended)
- Operational overhead (model updates, scaling)
- Lower quality for Indonesian (less training data)

**Decision**: Not chosen - too much operational complexity for MVP.

### Hybrid Approach

**Pros:**
- Start with vendor, migrate later
- Best of both worlds

**Cons:**
- More complex implementation
- Unclear migration path

**Decision**: Not chosen - YAGNI principle, vendor API sufficient.

## Consequences

### Positive

- Fast time to market (no ML setup)
- High-quality embeddings
- Easy provider switching
- No infrastructure costs

### Negative

- API dependency (mitigated with retries)
- Per-request costs (mitigated with caching)
- Data sent to external service (acceptable for public data)

### Future Considerations

- If costs become prohibitive, can migrate to self-hosted
- Interface abstraction makes migration straightforward
- Monitor embedding quality and costs

## Migration Path

If needed in future:

1. Implement self-hosted provider (sentence-transformers service)
2. Add to factory switch statement
3. Update config to use new provider
4. Re-embed existing chunks (optional, can run in parallel)

## References

- [Kolosal.ai Documentation](https://kolosal.ai/docs)
- [OpenAI Embeddings API](https://platform.openai.com/docs/guides/embeddings)
- [sentence-transformers](https://www.sbert.net/)

