# Implementation Todo List: Dashboard Improvement

## Overview
Refactor dashboard page to align with AI-chat-first architecture. Remove outdated features (products, data-input, restocking), update terminology (store→company), and add new widgets (conversations, insights, company profile).

## Pre-Implementation Setup
- [x] Review research findings
- [x] Confirm specification requirements
- [x] Validate technical plan
- [ ] Set up development environment
- [ ] Create feature branch: `dashboard-improvement`

## Todo Items

### Phase 1: Backend Updates (Foundation)

- [ ] **BACKEND-001**: Update dashboard handler to use company_id instead of store_id
  - **Estimated Time**: 1h
  - **Dependencies**: None
  - **Existing Pattern**: Follow `handlers/dashboard.go` structure
  - **Files to Modify**: 
    - `backend/handlers/dashboard.go`
    - `backend/middleware/auth.go` (check GetCompanyID function)
  - **Acceptance Criteria**: 
    - Dashboard endpoint uses `company_id` from context
    - All queries use `companies` table instead of `stores`
    - Response includes company profile data

- [ ] **BACKEND-002**: Add queries for conversations, insights, and file uploads to dashboard summary
  - **Estimated Time**: 1.5h
  - **Dependencies**: BACKEND-001
  - **Existing Pattern**: Follow existing query patterns in dashboard.go
  - **Files to Modify**: 
    - `backend/handlers/dashboard.go`
    - `backend/models/dashboard.go` (update DashboardSummary struct)
  - **Acceptance Criteria**:
    - Dashboard summary includes recent conversations count
    - Dashboard summary includes insights summary (count by type)
    - Dashboard summary includes file uploads status
    - All queries use company_id

- [ ] **BACKEND-003**: Update DashboardSummary model to include new fields
  - **Estimated Time**: 30min
  - **Dependencies**: BACKEND-002
  - **Existing Pattern**: Follow existing model structure
  - **Files to Modify**:
    - `backend/models/dashboard.go`
  - **Acceptance Criteria**:
    - Model includes CompanyProfile fields
    - Model includes RecentConversations array
    - Model includes InsightsSummary object
    - Model includes FileUploadsStatus array

### Phase 2: Frontend API Client Updates

- [ ] **FRONTEND-001**: Update API client to add missing endpoints
  - **Estimated Time**: 1h
  - **Dependencies**: None
  - **Existing Pattern**: Follow `frontend/src/lib/api.ts` structure
  - **Files to Modify**:
    - `frontend/src/lib/api.ts`
  - **Acceptance Criteria**:
    - `api.chat.conversations.list()` exists
    - `api.insights.list()` exists
    - `api.companies.get(id)` exists
    - `api.files.list()` exists
    - All return proper TypeScript types

- [ ] **FRONTEND-002**: Update TypeScript interfaces for dashboard data
  - **Estimated Time**: 30min
  - **Dependencies**: FRONTEND-001
  - **Existing Pattern**: Follow existing interface definitions
  - **Files to Modify**:
    - `frontend/src/lib/api.ts`
  - **Acceptance Criteria**:
    - DashboardSummary interface updated
    - New interfaces for Conversation, Insight, FileUpload
    - All types match backend models

### Phase 3: Frontend Component Refactoring

- [ ] **FRONTEND-003**: Remove outdated components from DashboardPage
  - **Estimated Time**: 1h
  - **Dependencies**: None
  - **Existing Pattern**: Follow existing component structure
  - **Files to Modify**:
    - `frontend/src/pages/DashboardPage.tsx`
  - **Acceptance Criteria**:
    - "Total Produk" KPI card removed
    - "Rekomendasi Restok" card removed
    - "Trend Pasar" card removed
    - Links to `/products` removed
    - Links to `/data-input` removed
    - "Hubungkan Toko Online" quick action removed
    - Old API calls removed (`api.recommendations.list()`, `api.market.trends()`)

- [ ] **FRONTEND-004**: Update terminology from store to company
  - **Estimated Time**: 30min
  - **Dependencies**: FRONTEND-003
  - **Existing Pattern**: Follow existing text patterns
  - **Files to Modify**:
    - `frontend/src/pages/DashboardPage.tsx`
    - `frontend/src/components/layout/Sidebar.tsx` (if needed)
  - **Acceptance Criteria**:
    - All "store" references changed to "company"
    - All "toko" references changed to "perusahaan" or "bisnis"
    - UI text updated throughout

- [ ] **FRONTEND-005**: Create CompanyProfileCard component
  - **Estimated Time**: 1.5h
  - **Dependencies**: FRONTEND-001, FRONTEND-002
  - **Existing Pattern**: Follow Card component pattern from DashboardPage
  - **Files to Create**:
    - `frontend/src/components/dashboard/CompanyProfileCard.tsx`
  - **Files to Modify**:
    - `frontend/src/pages/DashboardPage.tsx`
  - **Acceptance Criteria**:
    - Shows company name, industry, location
    - Shows company description (if available)
    - Displays company creation date
    - Links to company profile detail page (if exists)

- [ ] **FRONTEND-006**: Create RecentConversationsWidget component
  - **Estimated Time**: 2h
  - **Dependencies**: FRONTEND-001, FRONTEND-002
  - **Existing Pattern**: Follow card widget pattern
  - **Files to Create**:
    - `frontend/src/components/dashboard/RecentConversationsWidget.tsx`
  - **Files to Modify**:
    - `frontend/src/pages/DashboardPage.tsx`
  - **Acceptance Criteria**:
    - Shows last 3-5 conversations
    - Displays conversation title and last message preview
    - Shows conversation date
    - Links to `/ai-chat` with conversation ID
    - Empty state when no conversations

- [ ] **FRONTEND-007**: Create InsightsSummaryWidget component
  - **Estimated Time**: 1.5h
  - **Dependencies**: FRONTEND-001, FRONTEND-002
  - **Existing Pattern**: Follow card widget pattern
  - **Files to Create**:
    - `frontend/src/components/dashboard/InsightsSummaryWidget.tsx`
  - **Files to Modify**:
    - `frontend/src/pages/DashboardPage.tsx`
  - **Acceptance Criteria**:
    - Shows count for each insight type (forecast, market, marketing, regulation)
    - Displays last generated insight date
    - Links to respective insight pages
    - Empty state when no insights

- [ ] **FRONTEND-008**: Create FileUploadsWidget component
  - **Estimated Time**: 1.5h
  - **Dependencies**: FRONTEND-001, FRONTEND-002
  - **Existing Pattern**: Follow card widget pattern
  - **Files to Create**:
    - `frontend/src/components/dashboard/FileUploadsWidget.tsx`
  - **Files to Modify**:
    - `frontend/src/pages/DashboardPage.tsx`
  - **Acceptance Criteria**:
    - Shows recent file uploads (last 5)
    - Displays file name, type, status
    - Shows upload date
    - Status indicators (uploaded, processing, processed, failed)
    - Empty state when no uploads

- [ ] **FRONTEND-009**: Update Quick Actions section
  - **Estimated Time**: 1h
  - **Dependencies**: FRONTEND-003
  - **Existing Pattern**: Follow existing quick actions pattern
  - **Files to Modify**:
    - `frontend/src/pages/DashboardPage.tsx`
  - **Acceptance Criteria**:
    - "Start AI Chat" button (primary, prominent)
    - "Upload File" button
    - "View Forecast" button (links to `/forecast`)
    - "View Market Prediction" button (links to `/market-prediction`)
    - All buttons styled consistently

- [ ] **FRONTEND-010**: Update sales chart to use real data
  - **Estimated Time**: 1h
  - **Dependencies**: BACKEND-001, BACKEND-002
  - **Existing Pattern**: Follow existing chart pattern
  - **Files to Modify**:
    - `frontend/src/pages/DashboardPage.tsx`
  - **Acceptance Criteria**:
    - Chart uses real sales data from API
    - Shows last 7 days or last month (configurable)
    - Empty state when no sales data
    - Loading state while fetching

- [ ] **FRONTEND-011**: Update KPIs to reflect new architecture
  - **Estimated Time**: 1h
  - **Dependencies**: BACKEND-001, BACKEND-002
  - **Existing Pattern**: Follow existing KPI card pattern
  - **Files to Modify**:
    - `frontend/src/pages/DashboardPage.tsx`
  - **Acceptance Criteria**:
    - Revenue KPI kept (updated to use company_id)
    - Forecast Accuracy KPI updated or removed (if no data)
    - Add "Total Conversations" KPI
    - Add "Total Insights" KPI
    - All KPIs use real data

### Phase 4: Integration & Polish

- [ ] **INTEGRATION-001**: Connect all widgets to backend APIs
  - **Estimated Time**: 1h
  - **Dependencies**: All Phase 2 and Phase 3 items
  - **Integration Points**: 
    - CompanyProfileCard → `api.companies.get()`
    - RecentConversationsWidget → `api.chat.conversations.list()`
    - InsightsSummaryWidget → `api.insights.list()`
    - FileUploadsWidget → `api.files.list()`
  - **Acceptance Criteria**:
    - All widgets fetch data successfully
    - Error handling implemented
    - Loading states work correctly
    - Empty states display when no data

- [ ] **INTEGRATION-002**: Add loading and error states
  - **Estimated Time**: 1h
  - **Dependencies**: INTEGRATION-001
  - **Existing Pattern**: Follow existing loading patterns
  - **Files to Modify**:
    - `frontend/src/pages/DashboardPage.tsx`
    - All widget components
  - **Acceptance Criteria**:
    - Loading spinners during data fetch
    - Error messages displayed on failure
    - Retry functionality for failed requests
    - Graceful degradation

- [ ] **INTEGRATION-003**: Update dashboard layout and styling
  - **Estimated Time**: 1h
  - **Dependencies**: All widget components created
  - **Existing Pattern**: Follow existing Tailwind CSS patterns
  - **Files to Modify**:
    - `frontend/src/pages/DashboardPage.tsx`
    - Widget component files
  - **Acceptance Criteria**:
    - Responsive grid layout
    - Consistent spacing and styling
    - Mobile-friendly layout
    - Proper card shadows and hover effects

### Phase 5: Testing & Documentation

- [ ] **TEST-001**: Test dashboard with real data
  - **Estimated Time**: 1h
  - **Test Type**: Manual testing
  - **Coverage Target**: All widgets and features
  - **Test Files**: N/A (manual)
  - **Acceptance Criteria**:
    - Dashboard loads correctly
    - All widgets display data
    - All links work
    - Empty states work
    - Error handling works

- [ ] **TEST-002**: Test dashboard with no data (empty states)
  - **Estimated Time**: 30min
  - **Test Type**: Manual testing
  - **Coverage Target**: Empty state scenarios
  - **Test Files**: N/A (manual)
  - **Acceptance Criteria**:
    - Empty states display correctly
    - No errors in console
    - User can still navigate

- [ ] **DOC-001**: Update README if needed
  - **Estimated Time**: 30min
  - **Documentation Type**: User-facing
  - **Target Audience**: Developers
  - **Acceptance Criteria**:
    - Dashboard features documented
    - API endpoints documented

## Pattern Reuse Strategy

### Components to Reuse
- **Card, CardHeader, CardContent, CardTitle** (`@/components/ui/card`)
  - **Modifications needed**: None
  - **Usage**: All widget components

- **Button** (`@/components/ui/button`)
  - **Modifications needed**: None
  - **Usage**: Quick actions, links

- **LineChart, ResponsiveContainer** (recharts)
  - **Modifications needed**: Update data source
  - **Usage**: Sales chart

### Code Patterns to Follow
- **API Client Pattern**: Follow `api.ts` structure for new endpoints
- **Component Structure**: Follow existing DashboardPage component organization
- **State Management**: Use useState and useEffect hooks
- **Error Handling**: Follow existing error handling patterns

## Execution Strategy

### Continuous Implementation Rules
1. **Execute todo items in dependency order**
2. **Go for maximum flow - complete as much as possible without interruption**  
3. **Group all ambiguous questions for batch resolution at the end**
4. **Reuse existing patterns and components wherever possible**
5. **Update progress continuously**
6. **Document any deviations from plan**

### Checkpoint Schedule
- **Phase 1 Complete**: Backend updates done
  - **Expected Completion**: 3h
  - **Deliverables**: Updated dashboard API endpoint
  - **Review Criteria**: API returns correct data structure

- **Phase 2-3 Complete**: Frontend components done
  - **Expected Completion**: 8h
  - **Deliverables**: All widgets created and integrated
  - **Review Criteria**: Dashboard displays all new widgets

- **Phase 4 Complete**: Integration done
  - **Expected Completion**: 3h
  - **Deliverables**: Fully functional dashboard
  - **Review Criteria**: All features working end-to-end

## Questions for Batch Resolution
- **Data Availability**: What happens if user has no sales data? Show empty chart or hide it?
- **Company Profile**: Should we show full company profile or just summary?
- **Conversations**: How many conversations to show? 3, 5, or configurable?
- **Insights**: Should we show last generated insights or just counts?

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
- [ ] Dashboard shows new widgets correctly
- [ ] All outdated features removed
- [ ] Terminology updated (store→company)
- [ ] All API endpoints working
- [ ] Loading and error states implemented
- [ ] Empty states work correctly
- [ ] Responsive design verified
- [ ] Manual testing completed
- [ ] No console errors
- [ ] Code review ready

---
**Created:** 2025-12-02  
**Estimated Duration:** ~14 hours  
**Implementation Start:** 2025-12-02  
**Target Completion:** 2025-12-03
