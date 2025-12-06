# Roadmap: User Activity Logging & Token Usage Observability

## Summary
Implement end-to-end user activity and token usage visibility for admins: log chats/uploads/RAG queries with timestamps and user IDs, meter tokens per request, aggregate usage (daily/weekly/monthly), and surface filterable dashboards plus real-time monitoring.

## Feature Summary (implemented)
- Activity logging: chat, upload, RAG actions recorded with user/company and timestamps (no content stored).
- Token metering: per-request prompt/completion/total tokens; per-user/company aggregates.
- Aggregation: daily tables for activity and token usage; services for rollups.
- Admin APIs: activity aggregates and token usage aggregates with filters/pagination.
- Admin UI: Activity & Tokens page with filters, aggregates, system stats widgets, and live counters.
- Live feed: SSE with in-memory fan-out/backpressure (shared hub).

## Goals
- Log all key actions (chat, upload, RAG, token usage) with user association and timestamps.
- Provide admin dashboards with filters (date range, activity type, user) and system-wide stats.
- Deliver token usage metering with per-request capture and aggregate summaries.
- Enable real-time monitoring for live usage counters.

## Non-Goals
- Persist full message content in logs (only metadata/ids).
- Build billing/charging logic (metering is for visibility, not billing).

## Epics & Tasks

### Epic 1: Data Model & Instrumentation Plan
- Define activity event schema (chat/upload/RAG/token) + indexes.
- Define aggregates (daily/weekly/monthly) and retention/TTL rules.
- PII redaction policy + log payload shape.

### Epic 2: Backend Logging & Token Metering
- Middleware/hooks to emit activity events (chat, upload, RAG) with user_id + timestamps.
- Token metering per request; persist raw token_usage rows.
- Aggregation jobs for activity and token usage (daily/weekly/monthly).
- Admin APIs: activity list with filters; token usage summaries.

### Epic 3: Admin Dashboard (Activity & Token)
- Admin-only pages (separate admin app) to view logs.
- Filters: date range, activity type, user; pagination.
- Token usage views: per-request and aggregated (daily/weekly/monthly).
- System-wide stats widgets (counts, charts).

### Epic 4: Real-Time Monitoring
- Websocket/SSE feed for live counters (chats/uploads/RAG/token).
- In-memory cache/Redis channel for fan-out; backpressure/rate limits.
- Admin UI live widgets.

### Epic 5: QA, Ops, and Docs
- Load test ingestion path; ensure logging non-blocking.
- Alerting for ingestion failures/queue backlogs.
- Docs: API, env vars, dashboard usage, retention/TTL.

## Risks & Mitigations
- **High write volume** → batching/async queue; DB indexes.
- **PII leakage** → strict payload schema, redaction.
- **Dashboard perf** → pagination + time-bounded queries + aggregates.
- **Real-time pressure** → rate limit SSE/WS; drop/merge updates when hot.

## Success Metrics
- 100% of chat/upload/RAG requests emit activity events with user_id + timestamp.
- Token usage recorded per request; aggregates available per user daily/weekly/monthly.
- Admin dashboards respond <300ms for recent filtered queries; real-time widgets stay updated within 2s.

## Data Model & Instrumentation (Epic 1)
- Activity events (raw, append-only):
  - `activity_events`: `id (uuid)`, `user_id`, `company_id`, `action_type` (`chat`, `upload`, `rag_query`, `token_usage`), `resource_id` (conversation/message/file id), `metadata` (JSONB, PII-scrubbed), `created_at`, `trace_id` (optional).
  - Indexes: `(user_id, created_at desc)`, `(company_id, created_at desc)`, `(action_type, created_at)`, GIN on `metadata` if needed.
  - Retention: hot 90 days, optional archive/TTL policy after that (configurable).
- Token usage (per request):
  - `token_usage`: `id (uuid)`, `user_id`, `company_id`, `conversation_id`, `message_id`, `model`, `prompt_tokens`, `completion_tokens`, `total_tokens`, `created_at`.
  - Indexes: `(user_id, created_at)`, `(company_id, created_at)`, `(model, created_at)`.
- Aggregates:
  - `activity_aggregates`: grain daily; dims: date, user_id, company_id, action_type; metrics: counts.
  - `token_usage_aggregates`: grain daily/weekly/monthly; dims: date, user_id, company_id, model; metrics: prompt_tokens, completion_tokens, total_tokens.
  - Materialization: scheduled batch jobs; keep aggregated tables small and queryable for dashboards.
- PII/Content handling:
  - Do not store chat content; only ids and counts/tokens.
  - Scrub filenames to hashed identifiers in metadata; keep mime/size if needed.
  - Log only minimal user-identifying fields (ids); no emails in event rows.

