# Implementation Plan: Regulations Scraper URL Fix & KB Integration

## What Will Be Created

**Files:**
- `specs/active/regulations-scraper-fix/todo-list.md` - Comprehensive implementation checklist
- `specs/active/regulations-scraper-fix/progress.md` - Progress tracking document

## Implementation Strategy

### Execution Order

**Phase 1: Research & Analysis (Foundation)**
- Analyze search page HTML structure
- Understand pagination mechanism
- Document HTML selectors needed
- Identify business categorization requirements

**Phase 2: Crawler Updates (Core Functionality)**
- Update URL pattern to use search endpoint
- Implement search result parsing
- Update pagination logic
- Test crawler with search endpoint

**Phase 3: Business Categorization (Enhancement)**
- Define UMKM-relevant categories
- Implement categorization logic
- Apply categories during storage
- Filter regulations by relevance

**Phase 4: Integration & Testing (Polish)**
- Integrate with existing scheduler
- Test end-to-end flow
- Verify KB storage
- Document changes

### Pattern Reuse Approach

**Existing Patterns to Follow:**
- Current crawler structure (`Crawler` struct pattern)
- HTML parsing with goquery (same approach)
- Deduplication logic (keep as-is)
- Storage pattern (`Store.UpsertRegulation`)

**Components to Reuse:**
- `Regulation` struct (no changes needed)
- `Store` struct (enhance with categorization)
- `Scheduler` (no changes needed)
- Existing error handling patterns

**Conventions to Follow:**
- Go naming conventions
- Logging patterns (use logger.Default())
- Error wrapping with fmt.Errorf
- Context propagation

## Todo-List Structure Preview

### Phase 1: Research & Analysis
- [ ] Analyze search page HTML structure (30min)
- [ ] Document search result selectors (15min)
- [ ] Understand pagination mechanism (15min)
- [ ] Define business categorization rules (30min)

### Phase 2: Crawler Updates
- [ ] Update CrawlRegulations() to use search endpoint (1h)
- [ ] Implement parseSearchResult() helper (1h)
- [ ] Update crawlListingPage() for search results (1h)
- [ ] Fix pagination logic for search endpoint (30min)
- [ ] Test crawler with sample search (30min)

### Phase 3: Business Categorization
- [ ] Create categorization rules (1h)
- [ ] Implement CategorizeRegulation() helper (1h)
- [ ] Update Store.UpsertRegulation() to apply categories (30min)
- [ ] Add UMKM relevance filtering (30min)

### Phase 4: Integration & Testing
- [ ] Test end-to-end scraping flow (1h)
- [ ] Verify regulations stored correctly (30min)
- [ ] Check categorization applied (30min)
- [ ] Update documentation (30min)

## Implementation Approach

### File Organization

**Files to Modify:**
- `backend/services/scraper/regulations/crawler.go`
  - Update `CrawlRegulations()` method
  - Update `crawlListingPage()` method
  - Add `parseSearchResult()` helper
  - Update URL building logic

- `backend/services/scraper/regulations/store.go`
  - Enhance `UpsertRegulation()` with categorization
  - Add categorization helper (or new file)

**Files to Create (Optional):**
- `backend/services/scraper/regulations/categorizer.go`
  - Business categorization logic
  - UMKM relevance rules

### Testing Strategy

**Unit Tests:**
- Test search result parsing
- Test categorization logic
- Test URL building

**Integration Tests:**
- Test crawler with search endpoint
- Test storage with categorization
- Test end-to-end flow

**Manual Testing:**
- Run scraper with search endpoint
- Verify regulations in database
- Check categorization applied

## Success Criteria

**Definition of Done:**
- [ ] Crawler uses correct search endpoint
- [ ] Regulations extracted from search results
- [ ] Business categorization applied correctly
- [ ] Regulations stored in KB with categories
- [ ] No duplicate regulations
- [ ] All tests passing
- [ ] Documentation updated

**Validation:**
- Scraper successfully discovers regulations from search
- Regulations have proper business categories
- KB queries can filter by category
- No regressions in existing functionality

## Why This Approach

- **Research-first**: Need to understand HTML structure before coding
- **Incremental**: Build on existing infrastructure
- **Testable**: Each phase can be tested independently
- **Maintainable**: Follows existing patterns and conventions
