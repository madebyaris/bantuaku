# Implementation Todo List: Regulations Scraper URL Fix & KB Integration

## Overview
Fix regulations scraper to use correct search endpoint (`https://peraturan.go.id/cariglobal?PeraturanSearch%5Bidglobal%5D=`) and implement business categorization for UMKM-relevant regulations stored in KB.

## Pre-Implementation Setup
- [ ] Review research findings from feature brief
- [ ] Confirm specification requirements
- [ ] Validate technical plan
- [ ] Set up development environment
- [ ] Create feature branch: `regulations-scraper-fix`

## Todo Items

### Phase 1: Research & Analysis (Foundation)

- [ ] **RESEARCH-001**: Analyze search page HTML structure (30min)
  - **Estimated Time**: 30min
  - **Dependencies**: None
  - **Existing Pattern**: HTML parsing with goquery (see `crawler.go`)
  - **Files to Modify**: None (research only)
  - **Acceptance Criteria**: 
    - Document search result item HTML structure
    - Identify CSS selectors for regulation metadata
    - Understand pagination HTML structure
    - Document any JavaScript-rendered content

- [ ] **RESEARCH-002**: Document search result selectors (15min)
  - **Estimated Time**: 15min
  - **Dependencies**: RESEARCH-001
  - **Existing Pattern**: goquery selectors (see `crawlDetailPage()`)
  - **Files to Modify**: None (documentation)
  - **Acceptance Criteria**:
    - Selectors documented for: title, regulation number, year, category, PDF URL
    - Pagination selector identified
    - Search result container selector identified

- [ ] **RESEARCH-003**: Understand pagination mechanism (15min)
  - **Estimated Time**: 15min
  - **Dependencies**: RESEARCH-001
  - **Existing Pattern**: Current pagination uses `?page={n}` (see `CrawlRegulations()`)
  - **Files to Modify**: None (research only)
  - **Acceptance Criteria**:
    - Pagination URL pattern documented
    - Page parameter name identified
    - Max pages or "next" link pattern understood

- [ ] **RESEARCH-004**: Define business categorization rules (30min)
  - **Estimated Time**: 30min
  - **Dependencies**: None
  - **Existing Pattern**: Category field exists in regulations table
  - **Files to Modify**: None (planning)
  - **Acceptance Criteria**:
    - UMKM-relevant categories defined (tax, employment, food_safety, import_export, etc.)
    - Categorization rules documented
    - Filtering criteria for relevance established

### Phase 2: Crawler Updates (Core Functionality)

- [ ] **CRAWLER-001**: Update CrawlRegulations() to use search endpoint (1h)
  - **Estimated Time**: 1h
  - **Dependencies**: RESEARCH-001, RESEARCH-003
  - **Existing Pattern**: Current `CrawlRegulations()` method structure
  - **Files to Modify**: `backend/services/scraper/regulations/crawler.go`
  - **Acceptance Criteria**:
    - Method uses search endpoint URL pattern
    - Proper pagination handling
    - Error handling for search failures
    - Logging updated for search context

- [ ] **CRAWLER-002**: Implement parseSearchResult() helper (1h)
  - **Estimated Time**: 1h
  - **Dependencies**: RESEARCH-002
  - **Existing Pattern**: Similar to `crawlDetailPage()` parsing logic
  - **Files to Modify**: `backend/services/scraper/regulations/crawler.go`
  - **Acceptance Criteria**:
    - Extracts title, regulation number, year, category from search result
    - Returns Regulation struct or error
    - Handles missing fields gracefully
    - Uses documented selectors

- [ ] **CRAWLER-003**: Update crawlListingPage() for search results (1h)
  - **Estimated Time**: 1h
  - **Dependencies**: CRAWLER-002
  - **Existing Pattern**: Current `crawlListingPage()` structure
  - **Files to Modify**: `backend/services/scraper/regulations/crawler.go`
  - **Acceptance Criteria**:
    - Parses search results HTML correctly
    - Calls parseSearchResult() for each result item
    - Handles search result pagination
    - Maintains visited map for deduplication

- [ ] **CRAWLER-004**: Fix pagination logic for search endpoint (30min)
  - **Estimated Time**: 30min
  - **Dependencies**: RESEARCH-003, CRAWLER-001
  - **Existing Pattern**: Current pagination in `CrawlRegulations()`
  - **Files to Modify**: `backend/services/scraper/regulations/crawler.go`
  - **Acceptance Criteria**:
    - Pagination works correctly with search endpoint
    - Handles "no more results" case
    - Respects maxPages parameter

- [ ] **CRAWLER-005**: Test crawler with sample search (30min)
  - **Estimated Time**: 30min
  - **Dependencies**: CRAWLER-001, CRAWLER-002, CRAWLER-003, CRAWLER-004
  - **Existing Pattern**: Manual testing approach
  - **Files to Modify**: None (testing)
  - **Acceptance Criteria**:
    - Crawler successfully fetches search results
    - Regulations extracted correctly
    - Pagination works as expected
    - No errors in logs

### Phase 3: Business Categorization (Enhancement)

- [ ] **CATEG-001**: Create categorization rules (1h)
  - **Estimated Time**: 1h
  - **Dependencies**: RESEARCH-004
  - **Existing Pattern**: Category field in regulations table
  - **Files to Create**: `backend/services/scraper/regulations/categorizer.go` (optional)
  - **Files to Modify**: May add to `store.go` or new file
  - **Acceptance Criteria**:
    - Categorization rules implemented
    - UMKM-relevant categories mapped
    - Rules based on title/keywords/category

- [ ] **CATEG-002**: Implement CategorizeRegulation() helper (1h)
  - **Estimated Time**: 1h
  - **Dependencies**: CATEG-001
  - **Existing Pattern**: Helper function pattern
  - **Files to Modify**: `backend/services/scraper/regulations/store.go` or `categorizer.go`
  - **Acceptance Criteria**:
    - Takes Regulation struct as input
    - Returns business category
    - Handles edge cases (unknown categories)
    - Logs categorization decisions

- [ ] **CATEG-003**: Update Store.UpsertRegulation() to apply categories (30min)
  - **Estimated Time**: 30min
  - **Dependencies**: CATEG-002
  - **Existing Pattern**: Current `UpsertRegulation()` method
  - **Files to Modify**: `backend/services/scraper/regulations/store.go`
  - **Acceptance Criteria**:
    - Calls categorization helper
    - Stores category in database
    - Updates existing regulations with new categories
    - Maintains backward compatibility

- [ ] **CATEG-004**: Add UMKM relevance filtering (30min)
  - **Estimated Time**: 30min
  - **Dependencies**: CATEG-001
  - **Existing Pattern**: Filtering logic
  - **Files to Modify**: `backend/services/scraper/regulations/store.go` or `scheduler.go`
  - **Acceptance Criteria**:
    - Filters regulations by UMKM relevance
    - Skips irrelevant regulations (optional)
    - Logs filtering decisions
    - Configurable filtering (strict/lenient)

### Phase 4: Integration & Testing (Polish)

- [ ] **TEST-001**: Test end-to-end scraping flow (1h)
  - **Estimated Time**: 1h
  - **Dependencies**: All Phase 2 and Phase 3 items
  - **Test Type**: Integration test
  - **Coverage Target**: Main flow paths
  - **Test Files**: Manual testing, may add integration test
  - **Acceptance Criteria**:
    - Scraper runs successfully with search endpoint
    - Regulations discovered and extracted
    - Categories applied correctly
    - Data stored in KB

- [ ] **TEST-002**: Verify regulations stored correctly (30min)
  - **Estimated Time**: 30min
  - **Dependencies**: TEST-001
  - **Test Type**: Database verification
  - **Coverage Target**: Storage correctness
  - **Test Files**: Database queries
  - **Acceptance Criteria**:
    - Regulations in `regulations` table
    - Metadata correct (title, number, year, category)
    - PDF URLs valid
    - Deduplication working

- [ ] **TEST-003**: Check categorization applied (30min)
  - **Estimated Time**: 30min
  - **Dependencies**: TEST-002
  - **Test Type**: Data validation
  - **Coverage Target**: Categorization accuracy
  - **Test Files**: Database queries
  - **Acceptance Criteria**:
    - Categories assigned correctly
    - UMKM-relevant regulations categorized
    - Category distribution makes sense
    - No null/empty categories

- [ ] **DOC-001**: Update documentation (30min)
  - **Estimated Time**: 30min
  - **Dependencies**: All implementation complete
  - **Documentation Type**: Code comments, README updates
  - **Target Audience**: Developers, maintainers
  - **Acceptance Criteria**:
    - Code comments updated
    - Search endpoint documented
    - Categorization rules documented
    - Usage examples provided

## Pattern Reuse Strategy

### Components to Reuse
- **Crawler struct** (`backend/services/scraper/regulations/crawler.go`)
  - **Modifications needed**: Update methods, add helper
  - **Usage**: Keep structure, enhance functionality

- **Store struct** (`backend/services/scraper/regulations/store.go`)
  - **Modifications needed**: Add categorization logic
  - **Usage**: Enhance UpsertRegulation, add categorization helper

- **Regulation struct** (`backend/services/scraper/regulations/crawler.go`)
  - **Modifications needed**: None
  - **Usage**: Keep as-is, works for search results

- **Scheduler** (`backend/services/scraper/regulations/scheduler.go`)
  - **Modifications needed**: None (uses crawler interface)
  - **Usage**: No changes needed, works with updated crawler

### Code Patterns to Follow
- **HTML Parsing**: Use goquery with documented selectors (see `crawlDetailPage()`)
- **Error Handling**: Wrap errors with context using `fmt.Errorf` (see existing code)
- **Logging**: Use `logger.Default()` with structured logging (see existing code)
- **URL Building**: Use `buildFullURL()` helper (see `crawler.go`)
- **Deduplication**: Keep hash-based approach (see `store.go`)

## Execution Strategy

### Continuous Implementation Rules
1. **Execute todo items in dependency order**
2. **Go for maximum flow - complete as much as possible without interruption**  
3. **Group all ambiguous questions for batch resolution at the end**
4. **Reuse existing patterns and components wherever possible**
5. **Update progress continuously**
6. **Document any deviations from plan**

### Checkpoint Schedule
- **Phase 1 Complete**: Research and analysis done
  - **Expected Completion**: Day 1
  - **Deliverables**: HTML structure documentation, categorization rules
  - **Review Criteria**: Selectors documented, pagination understood

- **Phase 2 Complete**: Crawler updated
  - **Expected Completion**: Day 2
  - **Deliverables**: Working crawler with search endpoint
  - **Review Criteria**: Crawler extracts regulations from search

- **Phase 3 Complete**: Categorization implemented
  - **Expected Completion**: Day 2-3
  - **Deliverables**: Business categorization working
  - **Review Criteria**: Regulations categorized correctly

- **Phase 4 Complete**: Testing and documentation
  - **Expected Completion**: Day 3
  - **Deliverables**: Tested implementation, updated docs
  - **Review Criteria**: All tests passing, docs updated

## Questions for Batch Resolution
- **Search Endpoint**: Does search require any parameters beyond pagination?
  - **Context**: Need to understand if search needs keywords or filters
  - **Impact if unresolved**: May miss regulations or get wrong results

- **Categorization**: Should we skip non-UMKM regulations or just categorize them?
  - **Context**: Filtering vs categorization strategy
  - **Impact if unresolved**: May store irrelevant regulations

- **Pagination**: Does search endpoint support page parameter or use different mechanism?
  - **Context**: Need to understand pagination structure
  - **Impact if unresolved**: May not crawl all pages

## Progress Tracking

### Completed Items
- [ ] Update this section as items are completed
- [ ] Note any deviations or discoveries
- [ ] Record actual time vs estimates

### Blockers & Issues
- [ ] Document any blockers encountered
- [ ] Include resolution steps taken
- [ ] Note impact on timeline

### Discoveries & Deviations
- [ ] Document any plan changes needed
- [ ] Record new patterns or approaches discovered
- [ ] Note improvements to existing code

## Definition of Done
- [ ] All todo items completed
- [ ] Crawler uses correct search endpoint
- [ ] Regulations extracted correctly from search results
- [ ] Business categorization applied
- [ ] Regulations stored in KB with categories
- [ ] No duplicate regulations
- [ ] Tests passing (manual or automated)
- [ ] Documentation updated
- [ ] Code review ready
- [ ] No regressions in existing functionality

---
**Created:** 2025-12-05  
**Estimated Duration:** ~8-10 hours  
**Implementation Start:** TBD  
**Target Completion:** TBD
