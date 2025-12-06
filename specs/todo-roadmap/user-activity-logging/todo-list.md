# Todo List — User Activity Logging & Token Usage Observability

## Phase 0: Planning
- [x] Create roadmap files (roadmap.md, roadmap.json)
- [x] Create execution log scaffold
- [x] Create todo list for progress tracking

## Phase 1: Data Model & Instrumentation Plan
- [x] Define event schema (chat/upload/RAG/token) with indexes
- [x] Define aggregates (daily/weekly/monthly) and retention/TTL
- [x] Document PII redaction/payload rules

## Phase 2: Backend Logging & Token Metering
- [x] Emit activity events (chat, upload, RAG) with user_id + timestamps
- [x] Token metering per request; persist token_usage rows
- [x] Aggregation jobs (daily/weekly/monthly)
- [x] Admin APIs: activity list with filters; token usage summaries

## Phase 3: Admin Dashboard
- [x] Activity log views with filters (date range, type, user) + pagination
- [x] Token usage views (per-request and aggregates)
- [x] System-wide stats widgets/charts

## Phase 4: Real-Time Monitoring
- [x] WS/SSE feed for live counters (chats/uploads/RAG/token)
- [x] Redis/in-memory fan-out with rate limiting/backpressure
- [x] Admin UI live widgets

## Phase 5: QA, Ops, Docs
- [ ] Load/soak test ingestion path; ensure non-blocking logging
- [ ] Alerting for ingestion/queue failures
- [ ] Docs for APIs, env vars, dashboards, retention

## Definition of Done
- [ ] All phases completed; dashboards performant (<300ms recent queries)
- [ ] 100% activity events carry user_id + timestamp
- [ ] Token usage recorded per request with aggregates per user (daily/weekly/monthly)
- [ ] Real-time widgets update within 2 seconds

