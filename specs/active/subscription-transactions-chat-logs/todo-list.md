# Implementation Todo List: Subscription Transactions, Chat Logs & Token Tracking

## Overview
Build comprehensive tracking system for subscription transactions, chat usage logs (aggregate), and AI token usage. This enables admin visibility into billing history, usage patterns, and cost management without exposing sensitive message content.

## Pre-Implementation Setup
- [x] Review research findings
- [x] Confirm specification requirements
- [x] Validate technical plan
- [x] Set up development environment
- [ ] Create feature branch: `subscription-transactions-chat-logs`

## Todo Items

### Phase 1: Database Schema (Foundation)
- [x] **DB-001**: Create migration `016_subscription_transactions_chat_logs_token_tracking.sql`
  - **Estimated Time**: 1h
  - **Dependencies**: None
  - **Existing Pattern**: Follow `008_admin_billing_rbac.sql` structure
  - **Files to Create**: `database/migrations/016_subscription_transactions_chat_logs_token_tracking.sql`
  - **Acceptance Criteria**: 
    - `subscription_transactions` table with: id, subscription_id, company_id, event_type, old_plan_id, new_plan_id, old_status, new_status, changed_by_user_id, metadata (JSONB), created_at
    - `chat_usage_logs` table with: id, company_id, date, total_messages, total_conversations, unique_users, created_at
    - `token_usage` table with: id, company_id, conversation_id, message_id, model, prompt_tokens, completion_tokens, total_tokens, created_at
    - Proper indexes on company_id, subscription_id, created_at, date
    - Foreign key constraints

### Phase 2: Backend Services (Core Implementation)
- [x] **SVC-001**: Create `services/transactions/service.go` for subscription transaction logging
  - **Estimated Time**: 1h
  - **Dependencies**: DB-001
  - **Existing Pattern**: Follow `services/audit/logger.go` pattern
  - **Files to Create**: `backend/services/transactions/service.go`
  - **Acceptance Criteria**:
    - `LogTransaction()` method to record subscription events
    - `GetTransactionHistory()` method to retrieve history for a subscription
    - Support for event types: create, upgrade, downgrade, cancel, renew, status_change

- [x] **SVC-002**: Create `services/chatlogs/service.go` for aggregate chat usage logging
  - **Estimated Time**: 1h
  - **Dependencies**: DB-001
  - **Existing Pattern**: Follow `services/usage/service.go` pattern
  - **Files to Create**: `backend/services/chatlogs/service.go`
  - **Acceptance Criteria**:
    - `LogDailyUsage()` method to aggregate daily chat stats
    - `GetUsageStats()` method to retrieve usage with filters (company, date range)
    - Aggregate: total messages, conversations, unique users per day

- [x] **SVC-003**: Create `services/tokenusage/service.go` for token usage tracking
  - **Estimated Time**: 1h
  - **Dependencies**: DB-001
  - **Existing Pattern**: Follow `services/usage/service.go` pattern
  - **Files to Create**: `backend/services/tokenusage/service.go`
  - **Acceptance Criteria**:
    - `LogTokenUsage()` method to store token data from chat completion
    - `GetTokenUsage()` method to retrieve usage with filters (company, model, date range)
    - Calculate estimated costs based on model pricing

### Phase 3: Backend Handlers (Integration)
- [x] **HND-001**: Add transaction logging to subscription handlers
  - **Estimated Time**: 1h
  - **Dependencies**: SVC-001
  - **Existing Pattern**: Follow `handlers/admin/subscriptions.go` pattern
  - **Files to Modify**: `backend/handlers/admin/subscriptions.go`
  - **Acceptance Criteria**:
    - Log transaction in `CreateSubscription()` when subscription created
    - Log transaction in `UpdateSubscriptionStatus()` when status changes
    - Log transaction when plan changes (upgrade/downgrade)
    - Include user_id from context in transaction log

- [x] **HND-002**: Add token tracking to chat completion handler
  - **Estimated Time**: 1h
  - **Dependencies**: SVC-003
  - **Existing Pattern**: Follow `handlers/chat.go` pattern
  - **Files to Modify**: `backend/handlers/chat.go`
  - **Acceptance Criteria**:
    - Extract `Usage` data from `ChatCompletionResponse`
    - Call `LogTokenUsage()` after successful chat completion
    - Store: company_id, conversation_id, message_id, model, prompt_tokens, completion_tokens, total_tokens
    - Handle errors gracefully (don't block chat if logging fails)

- [x] **HND-003**: Create admin API handler for subscription transactions
  - **Estimated Time**: 1h
  - **Dependencies**: SVC-001
  - **Existing Pattern**: Follow `handlers/admin/subscriptions.go` pattern
  - **Files to Create**: `backend/handlers/admin/transactions.go`
  - **Acceptance Criteria**:
    - `GET /api/v1/admin/subscriptions/{id}/transactions` endpoint
    - Returns chronological transaction history
    - Supports pagination (page, limit)
    - Admin authentication required

- [x] **HND-004**: Create admin API handler for chat usage logs
  - **Estimated Time**: 1h
  - **Dependencies**: SVC-002
  - **Existing Pattern**: Follow `handlers/admin/users.go` pattern
  - **Files to Create**: `backend/handlers/admin/chatlogs.go`
  - **Acceptance Criteria**:
    - `GET /api/v1/admin/chat-usage` endpoint
    - Query params: company_id (optional), start_date, end_date
    - Returns aggregate usage stats (total messages, conversations, unique users)
    - Admin authentication required

- [x] **HND-005**: Create admin API handler for token usage
  - **Estimated Time**: 1h
  - **Dependencies**: SVC-003
  - **Existing Pattern**: Follow `handlers/admin/users.go` pattern
  - **Files to Create**: `backend/handlers/admin/tokenusage.go`
  - **Acceptance Criteria**:
    - `GET /api/v1/admin/token-usage` endpoint
    - Query params: company_id (optional), model (optional), start_date, end_date
    - Returns token usage stats with estimated costs
    - Admin authentication required

- [x] **HND-006**: Register new admin routes in main.go
  - **Estimated Time**: 30min
  - **Dependencies**: HND-003, HND-004, HND-005
  - **Existing Pattern**: Follow `backend/main.go` route registration pattern
  - **Files to Modify**: `backend/main.go`
  - **Acceptance Criteria**:
    - Routes registered with admin authentication middleware
    - Proper path patterns: `/api/v1/admin/subscriptions/{id}/transactions`, `/api/v1/admin/chat-usage`, `/api/v1/admin/token-usage`

### Phase 4: Frontend API Client (Integration)
- [x] **FE-001**: Add admin API client methods for new endpoints
  - **Estimated Time**: 30min
  - **Dependencies**: HND-003, HND-004, HND-005
  - **Existing Pattern**: Follow `admin/src/lib/api.ts` pattern
  - **Files to Modify**: `admin/src/lib/api.ts`
  - **Acceptance Criteria**:
    - `api.admin.transactions.get(subscriptionId, page, limit)` method
    - `api.admin.chatUsage.get(filters)` method
    - `api.admin.tokenUsage.get(filters)` method
    - Proper TypeScript types for all responses

### Phase 5: Testing & Validation
- [ ] **TEST-001**: Test subscription transaction logging
  - **Estimated Time**: 1h
  - **Dependencies**: HND-001, HND-003
  - **Test Type**: Integration test
  - **Acceptance Criteria**:
    - Create subscription → transaction logged
    - Update status → transaction logged
    - API endpoint returns correct transaction history
    - Pagination works correctly

- [ ] **TEST-002**: Test token usage tracking
  - **Estimated Time**: 1h
  - **Dependencies**: HND-002, HND-005
  - **Test Type**: Integration test
  - **Acceptance Criteria**:
    - Send chat message → token usage logged
    - API endpoint returns correct token stats
    - Filters (company, model, date) work correctly
    - Estimated costs calculated correctly

- [ ] **TEST-003**: Test chat usage logging
  - **Estimated Time**: 1h
  - **Dependencies**: HND-004
  - **Test Type**: Integration test
  - **Acceptance Criteria**:
    - API endpoint returns aggregate chat stats
    - Filters (company, date range) work correctly
    - No message content exposed

## Pattern Reuse Strategy

### Components to Reuse
- **`services/audit/logger.go`** → Transaction logging pattern
- **`services/usage/service.go`** → Usage tracking pattern
- **`handlers/admin/subscriptions.go`** → Admin handler pattern
- **`handlers/chat.go`** → Chat handler pattern for token extraction
- **`database/migrations/008_admin_billing_rbac.sql`** → Table structure patterns

### Code Patterns to Follow
- **Admin handlers**: Follow `handlers/admin/subscriptions.go` with admin auth middleware
- **Service layer**: Follow `services/usage/service.go` with database operations
- **Error handling**: Use `errors.NewDatabaseError()`, `errors.NewNotFoundError()`
- **Pagination**: Follow existing pagination pattern (page, limit, offset)
- **JSON responses**: Follow existing response format with pagination metadata

## Execution Strategy

### Continuous Implementation Rules
1. **Execute todo items in dependency order**
2. **Go for maximum flow - complete as much as possible without interruption**  
3. **Group all ambiguous questions for batch resolution at the end**
4. **Reuse existing patterns and components wherever possible**
5. **Update progress continuously**
6. **Document any deviations from plan**

### Checkpoint Schedule
- **Phase 1 Complete**: Database schema created and migrated
- **Phase 2 Complete**: All services implemented and tested
- **Phase 3 Complete**: All handlers implemented and routes registered
- **Phase 4 Complete**: Frontend API client updated
- **Phase 5 Complete**: All tests passing, ready for review

## Questions for Batch Resolution
- None currently - proceed with implementation

## Progress Tracking

### Completed Items
- [ ] Update this section as items are completed

### Blockers & Issues
- [ ] Document any blockers encountered

### Discoveries & Deviations
- [ ] Document any plan changes needed

## Definition of Done
- [ ] All todo items completed
- [ ] Database migration applied successfully
- [ ] All API endpoints tested and working
- [ ] Token usage tracked correctly from chat completions
- [ ] Transaction history logged for all subscription changes
- [ ] Chat usage stats available without exposing message content
- [ ] Code follows existing patterns
- [ ] No security vulnerabilities introduced

---
**Created:** 2025-12-05  
**Estimated Duration:** 8-10 hours  
**Implementation Start:** 2025-12-05  
**Target Completion:** 2025-12-05
