# Subscription Transactions, Chat Logs & Token Tracking Feature Brief

## 🎯 Context (2min)
**Problem**: Admins need visibility into subscription transaction history, chat usage patterns, and AI token costs for billing, analytics, and cost management. Currently, subscription changes aren't tracked as transactions, chat usage is only visible through raw message counts, and token usage isn't stored despite being available in API responses.

**Users**: 
- **Admins** - Need transaction history and usage analytics for customer support
- **Finance Team** - Need payment history and billing reconciliation
- **Product Team** - Need usage patterns and cost analysis

**Success**: Admins can view complete subscription transaction history, aggregate chat usage stats (without exposing message content), and token usage/costs per company/model for accurate billing and analytics.

## 🔍 Quick Research (15min)

### Existing Patterns

- **`payments` table** (`008_admin_billing_rbac.sql`) → Transaction history structure with `subscription_id`, `company_id`, `amount`, `status`, `stripe_payment_intent_id` | **Reuse**: Extend for subscription events or create `subscription_transactions` table
- **`subscriptions` table** → Tracks current state but not history of changes | **Reuse**: Add transaction logging when status/plan changes
- **`messages` table** (`003_add_chat_tables.sql`) → Stores all chat messages | **Reuse**: Query for aggregate counts, create `chat_usage_logs` for performance
- **`audit_logs` table** (`008_admin_billing_rbac.sql`) → Event tracking pattern with `action`, `resource_type`, `resource_id`, `metadata` JSONB | **Reuse**: Similar pattern for transaction events
- **`ChatCompletionResponse.Usage` struct** (`backend/services/chat/interface.go`) → Contains `PromptTokens`, `CompletionTokens`, `TotalTokens` | **Reuse**: Store this data in new `token_usage` table
- **Subscription handlers** (`backend/handlers/admin/subscriptions.go`) → `CreateSubscription`, `UpdateSubscriptionStatus` | **Reuse**: Add transaction logging here
- **Chat handler** (`backend/handlers/chat.go`) → `SendMessage` receives `ChatCompletionResponse` with `Usage` field | **Reuse**: Extract and store token data here

### Tech Decision

**Approach**: 
1. **Subscription Transactions**: Create `subscription_transactions` table to track all subscription events (create, upgrade, downgrade, cancel, renew, status change) - separate from `payments` which tracks actual payments
2. **Chat Usage Logs**: Create `chat_usage_logs` aggregate table (daily/monthly summaries) for performance, query `messages` table for real-time counts
3. **Token Usage**: Create `token_usage` table to store per-completion token data from `ChatCompletionResponse.Usage`

**Why**: 
- Separate transaction events from payments (subscription changes vs actual charges)
- Aggregate tables prevent expensive queries on large `messages` table
- Token data already available in responses, just needs storage for analytics
- Follows existing `audit_logs` pattern for event tracking

**Avoid**: 
- Storing full message content in logs (privacy concern)
- Querying `messages` table directly for admin dashboards (performance issue)
- Mixing subscription events with payment records (different concerns)

## ✅ Requirements (10min)

### Subscription Transaction History
- **Story**: As an admin, I want to see complete transaction history for a subscription including upgrades, downgrades, cancellations, and renewals
- **Acceptance**: 
  - `subscription_transactions` table tracks all subscription events
  - Admin API endpoint `GET /admin/subscriptions/{id}/transactions` returns chronological history
  - Each transaction includes: event type, old plan, new plan, old status, new status, timestamp, user who made change
  - Frontend shows transaction timeline in subscription detail view

### Chat Usage Logs (Aggregate)
- **Story**: As an admin, I want to see chat usage statistics per company without accessing message content
- **Acceptance**:
  - `chat_usage_logs` table stores daily/monthly aggregates (total messages, conversations, unique users)
  - Admin API endpoint `GET /admin/chat-usage` returns usage stats with filters (company, date range)
  - No message content stored in logs (privacy)
  - Frontend shows usage charts/graphs in admin dashboard

### Token Usage Tracking
- **Story**: As an admin, I want to track AI token usage and costs per company/model for billing and analytics
- **Acceptance**:
  - `token_usage` table stores input/output tokens per chat completion
  - Token data extracted from `ChatCompletionResponse.Usage` in chat handler
  - Admin API endpoint `GET /admin/token-usage` returns token stats with filters (company, model, date range)
  - Frontend shows token usage and estimated costs in admin dashboard

## 🏗️ Implementation (5min)

**Components**:
- Database migrations: `016_subscription_transactions_chat_logs_token_tracking.sql`
- Backend services: `services/transactions/`, `services/chatlogs/`, `services/tokenusage/`
- Backend handlers: `handlers/admin/transactions.go`, `handlers/admin/chatlogs.go`, `handlers/admin/tokenusage.go`
- Frontend API client: Extend `admin/src/lib/api.ts` with new endpoints
- Frontend pages: Optional admin dashboard widgets (can be added later)

**APIs**:
- `GET /api/v1/admin/subscriptions/{id}/transactions` - Subscription transaction history
- `GET /api/v1/admin/chat-usage` - Chat usage statistics (query params: company_id, start_date, end_date)
- `GET /api/v1/admin/token-usage` - Token usage statistics (query params: company_id, model, start_date, end_date)

**Data**:
- New tables: `subscription_transactions`, `chat_usage_logs`, `token_usage`
- Modified handlers: `subscriptions.go` (add transaction logging), `chat.go` (add token tracking)
- Indexes: Add indexes for query performance on `company_id`, `created_at`, `subscription_id`

## 📋 Next Actions (2min)

- [ ] Create database migration `016_subscription_transactions_chat_logs_token_tracking.sql` (1h)
- [ ] Add transaction logging to subscription handlers (1h)
- [ ] Add token tracking to chat completion handler (1h)
- [ ] Create admin API endpoints for transactions, chat usage, token usage (2h)
- [ ] Update frontend API client with new endpoints (30min)
- [ ] Test all endpoints and verify data storage (1h)

**Start Coding In**: Ready now

---
**Total Planning Time**: ~30min | **Owner**: Development Team | 2025-12-05

<!-- Living Document - Update as you code -->

## 🔄 Implementation Tracking

**CRITICAL**: Follow the todo-list systematically. Mark items as complete, document blockers, update progress.

### Progress
- [ ] Track completed items here
- [ ] Update daily

### Blockers
- [ ] Document any blockers

**See**: [.sdd/IMPLEMENTATION_GUIDE.md](mdc:.sdd/IMPLEMENTATION_GUIDE.md) for detailed execution rules.
