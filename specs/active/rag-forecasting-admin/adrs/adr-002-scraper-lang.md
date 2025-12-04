# ADR-002: Use Go for Regulations Scraper

## Status
Accepted

## Context

We need to build a scraper for peraturan.go.id to extract regulation PDFs and text. Language options:

1. **Go** - Same language as backend
2. **Node.js/TypeScript** - JavaScript ecosystem, Playwright
3. **Python** - Rich scraping libraries (Scrapy, BeautifulSoup)

## Decision

Use **Go** for the regulations scraper, integrated into the main backend.

## Rationale

### Advantages

1. **Codebase Consistency**: Same language as backend, shared utilities
2. **Performance**: Go's concurrency model (goroutines) excellent for parallel crawling
3. **Memory Efficiency**: Lower memory footprint than Node.js/Python
4. **Deployment Simplicity**: Single binary, no separate service needed
5. **Type Safety**: Strong typing catches errors at compile time
6. **Library Support**: chromedp provides headless Chrome automation

### Trade-offs

1. **Scraping Libraries**: Fewer high-level scraping libraries than Python
   - **Mitigation**: chromedp is mature and sufficient for our needs
2. **PDF Processing**: Go PDF libraries less mature than Python
   - **Mitigation**: Use external tool (pdftotext) or Go bindings (unidoc)
3. **OCR Support**: Go OCR libraries less common
   - **Mitigation**: Use Tesseract CLI or Python microservice for OCR fallback

## Implementation

### Structure

```
backend/services/scraper/regulations/
├── crawler.go      # Enumerate PDFs from peraturan.go.id
├── extract.go      # PDF text extraction
├── chunker.go      # Semantic chunking
├── store.go        # Database persistence + dedup
└── scheduler.go    # Cron job + manual trigger
```

### Key Components

**Crawler (chromedp):**
```go
import "github.com/chromedp/chromedp"

func CrawlRegulations(ctx context.Context) ([]Regulation, error) {
    // Navigate to peraturan.go.id
    // Extract PDF links + metadata
    // Return list of regulations
}
```

**PDF Extraction:**
```go
// Option 1: Go library (unidoc)
import "github.com/unidoc/unipdf/v3/extractor"

// Option 2: External tool (pdftotext)
// Fallback to Python OCR service if needed
```

**Chunking:**
```go
func ChunkText(text string, chunkSize int, overlap int) []Chunk {
    // Split into semantic chunks with overlap
    // Preserve section boundaries
}
```

## Alternatives Considered

### Node.js/TypeScript with Playwright

**Pros:**
- Excellent browser automation (Playwright)
- Rich ecosystem
- Easy async/await patterns

**Cons:**
- Separate service deployment
- Higher memory usage
- Different language from backend

**Decision**: Not chosen - prefer codebase consistency.

### Python with Scrapy/BeautifulSoup

**Pros:**
- Best scraping libraries (Scrapy, BeautifulSoup)
- Excellent PDF processing (PyPDF2, pdfplumber)
- Best OCR support (Tesseract, pytesseract)

**Cons:**
- Separate service deployment
- Different language from backend
- Slower runtime performance

**Decision**: Not chosen - prefer Go for consistency, acceptable trade-offs.

## Consequences

### Positive

- Unified codebase (easier maintenance)
- Better performance (goroutines for parallel crawling)
- Single deployment unit
- Shared database connection pool

### Negative

- May need external tools for PDF/OCR (acceptable)
- Less mature scraping libraries than Python (chromedp sufficient)

### Mitigations

- Use chromedp for browser automation (mature Go library)
- Use unidoc or pdftotext for PDF extraction
- Create Python microservice for OCR fallback if needed (rare case)

## References

- [chromedp GitHub](https://github.com/chromedp/chromedp)
- [unidoc PDF library](https://github.com/unidoc/unipdf)
- [pdftotext tool](https://www.xpdfreader.com/pdftotext-man.html)

