# Plan: Regulations Scraper URL Fix & KB Integration

## What Will Be Created

**File**: `specs/active/regulations-scraper-fix/feature-brief.md`

## Brief Structure Outline

### 1. Context (2min)
- **Problem**: Current crawler uses wrong URL pattern (`/peraturan?page=N`) instead of search endpoint
- **Users**: System administrators, AI assistant (RAG system)
- **Success**: Regulations successfully scraped from search endpoint and stored in KB with proper business categorization

### 2. Quick Research (15min)
**Research Areas:**
- Current crawler implementation (`backend/services/scraper/regulations/crawler.go`)
- Search page structure at `https://peraturan.go.id/cariglobal?PeraturanSearch%5Bidglobal%5D=`
- Existing KB storage patterns (`backend/services/scraper/regulations/store.go`)
- Business categorization logic (how regulations are categorized for UMKM)
- Existing regulation models and data structures

**Patterns to Identify:**
- How search results are paginated
- HTML structure of search results page
- How to extract regulation metadata from search results
- How regulations are currently categorized and stored
- Business logic for filtering relevant regulations for UMKM

### 3. Requirements (10min)
**Must-Have:**
- Crawl from search endpoint `https://peraturan.go.id/cariglobal?PeraturanSearch%5Bidglobal%5D=`
- Extract regulation metadata (title, number, year, category, PDF URL)
- Handle pagination correctly
- Store regulations in KB with proper business categorization
- Maintain deduplication logic
- Support filtering by business relevance (UMKM-focused regulations)

**Nice-to-Have:**
- Configurable search parameters
- Better error handling for search page changes
- Rate limiting for search requests

### 4. Implementation Approach (5min)
**Components:**
- Update `Crawler.CrawlRegulations()` to use search endpoint
- Update `crawlListingPage()` to parse search results HTML
- Enhance `Store` to apply business categorization logic
- Add business relevance filtering

**APIs:**
- No new API endpoints (uses existing `/api/v1/regulations/scrape`)

**Data:**
- Regulations table (existing)
- May need to enhance categorization fields

### 5. Next Actions (2min)
- [ ] Analyze search page HTML structure (30min)
- [ ] Update crawler to use search endpoint (2h)
- [ ] Implement business categorization logic (2h)
- [ ] Test with sample search results (1h)

## Research Approach

1. **Examine current crawler** (5min)
   - Review `crawler.go` implementation
   - Understand current URL pattern and parsing logic

2. **Analyze search endpoint** (5min)
   - Check search page structure
   - Identify pagination mechanism
   - Find regulation link patterns

3. **Review KB storage** (3min)
   - Check `store.go` for storage patterns
   - Understand deduplication logic
   - Review regulation model structure

4. **Business logic research** (2min)
   - Identify UMKM-relevant regulation categories
   - Understand how to filter/prioritize regulations

## Why This Approach

- **Brief is appropriate**: This is a focused fix/enhancement, not a new feature
- **Research-first**: Need to understand search page structure before coding
- **Business-focused**: Must ensure regulations are relevant to UMKM use case
- **Incremental**: Builds on existing scraper infrastructure

## Success Criteria

- Crawler successfully uses search endpoint
- Regulations extracted with correct metadata
- Data stored in KB with business categorization
- No duplicate regulations
- Regulations filtered for UMKM relevance
