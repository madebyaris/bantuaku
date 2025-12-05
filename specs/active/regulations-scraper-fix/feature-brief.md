# Regulations Scraper URL Fix & KB Integration Feature Brief

## 🎯 Context (2min)
**Problem**: Current regulations scraper crawls from wrong URL pattern (`/peraturan?page=N`) instead of the actual search endpoint (`/cariglobal?PeraturanSearch%5Bidglobal%5D=`). This causes the scraper to miss regulations or fail to discover them correctly. Additionally, regulations need to be stored in the Knowledge Base (KB) with proper business categorization for UMKM relevance.

**Users**: 
- System administrators (trigger scraping jobs)
- AI Assistant / RAG system (retrieve regulations for insights)
- UMKM users (benefit from relevant regulation insights)

**Success**: Regulations successfully scraped from correct search endpoint, stored in KB with proper categorization, and filtered for UMKM business relevance. Scraper can discover and process regulations reliably.

## 🔍 Quick Research (15min)

### Existing Patterns

**Current Crawler Implementation** (`backend/services/scraper/regulations/crawler.go`):
- Uses `{baseURL}/peraturan?page={n}` pattern
- Parses HTML with goquery
- Extracts regulation links with selector `a[href*='/peraturan/']`
- Crawls detail pages to get PDF URLs and metadata
- **Issue**: Wrong URL pattern, may not match actual site structure

**KB Storage Pattern** (`backend/services/scraper/regulations/store.go`):
- Uses deduplication via hash (regulation_number + year + PDF URL)
- Stores in `regulations` table
- Creates sections and chunks for RAG
- **Reuse**: Keep deduplication logic, enhance with business categorization

**Regulation Data Model** (`database/migrations/005_regulations_kb.sql`):
- Tables: `regulations`, `regulation_sections`, `regulation_chunks`, `regulation_sources`
- Fields: title, regulation_number, year, category, status, source_url, pdf_url
- **Enhancement needed**: Business relevance categorization

**Scheduler Pattern** (`backend/services/scraper/regulations/scheduler.go`):
- Runs jobs with maxPages parameter
- Processes regulations sequentially
- **Reuse**: Keep job structure, update crawler call

### Tech Decision

**Approach**: 
1. Update crawler to use search endpoint `https://peraturan.go.id/cariglobal?PeraturanSearch%5Bidglobal%5D=`
2. Parse search results HTML structure (need to analyze actual page)
3. Extract regulation metadata from search result items
4. Apply business categorization logic for UMKM relevance
5. Store in existing KB structure with enhanced categorization

**Why**: 
- Search endpoint is the correct way to discover regulations
- Reuse existing storage and chunking infrastructure
- Business categorization ensures UMKM-relevant regulations are prioritized

**Avoid**: 
- Don't create new storage tables (reuse existing)
- Don't change chunking logic (works fine)
- Don't remove deduplication (critical for data quality)

## ✅ Requirements (10min)

### Core Requirements

1. **Update Crawler URL Pattern**
   - **Story**: As a system admin, I need the scraper to use the correct search endpoint
   - **Acceptance**: Crawler uses `https://peraturan.go.id/cariglobal?PeraturanSearch%5Bidglobal%5D=` with proper pagination
   - **Files**: `backend/services/scraper/regulations/crawler.go`

2. **Parse Search Results**
   - **Story**: As a scraper, I need to extract regulation metadata from search results page
   - **Acceptance**: Successfully extracts title, regulation number, year, category, PDF URL from search result items
   - **Files**: `backend/services/scraper/regulations/crawler.go`

3. **Business Categorization**
   - **Story**: As UMKM users, I need regulations filtered for business relevance
   - **Acceptance**: Regulations categorized by UMKM relevance (tax, employment, food safety, import/export, etc.)
   - **Files**: `backend/services/scraper/regulations/store.go`, possibly new categorization logic

4. **KB Storage with Categorization**
   - **Story**: As the RAG system, I need regulations stored with business categories
   - **Acceptance**: Regulations stored in KB with proper category tags for filtering
   - **Files**: `backend/services/scraper/regulations/store.go`

5. **Maintain Deduplication**
   - **Story**: As a system, I need to avoid duplicate regulations
   - **Acceptance**: Existing hash-based deduplication continues to work
   - **Files**: `backend/services/scraper/regulations/store.go` (no changes needed)

### Nice-to-Have

- Configurable search parameters (keywords, date ranges)
- Better error handling for search page structure changes
- Rate limiting specific to search endpoint

## 🏗️ Implementation (5min)

### Components

**Modified Files:**
- `backend/services/scraper/regulations/crawler.go`
  - Update `CrawlRegulations()` to use search endpoint
  - Update `crawlListingPage()` to parse search results HTML
  - May need new `parseSearchResult()` helper function

- `backend/services/scraper/regulations/store.go`
  - Add business categorization logic
  - Enhance `UpsertRegulation()` to apply categories
  - Possibly add `CategorizeRegulation()` helper

**New Files (if needed):**
- `backend/services/scraper/regulations/categorizer.go` (optional)
  - Business categorization rules
  - UMKM relevance filtering

### APIs
- No new API endpoints
- Uses existing `/api/v1/regulations/scrape` endpoint

### Data
- **Regulations table**: May need to enhance `category` field usage
- **No schema changes**: Use existing structure, enhance categorization logic

## 📋 Next Actions (2min)

- [ ] Analyze search page HTML structure at `https://peraturan.go.id/cariglobal?PeraturanSearch%5Bidglobal%5D=` (30min)
  - Identify search result item selectors
  - Understand pagination mechanism
  - Document HTML structure
  
- [ ] Update crawler to use search endpoint (2h)
  - Modify `CrawlRegulations()` method
  - Update `crawlListingPage()` to parse search results
  - Test with sample search queries

- [ ] Implement business categorization logic (2h)
  - Define UMKM-relevant categories
  - Create categorization rules
  - Apply during storage

- [ ] Test end-to-end scraping flow (1h)
  - Run scraper with search endpoint
  - Verify regulations stored correctly
  - Check categorization applied

**Start Coding In**: After brief approval and HTML structure analysis

---
**Total Planning Time**: ~30min | **Owner**: Development Team | 2025-12-05

<!-- Living Document - Update as you code -->

## 🔄 Implementation Tracking

**CRITICAL**: Follow the todo-list systematically. Mark items as complete, document blockers, update progress.

### Progress
- [ ] Track completed items here
- [ ] Update daily

### Blockers
- [ ] Need to analyze actual search page HTML structure
- [ ] May need to understand search pagination mechanism

**See**: [.sdd/IMPLEMENTATION_GUIDE.md](mdc:.sdd/IMPLEMENTATION_GUIDE.md) for detailed execution rules.
