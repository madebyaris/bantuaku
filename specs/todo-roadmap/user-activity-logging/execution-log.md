# Execution Log — User Activity Logging & Token Usage Observability

- **2025-12-06** — Initialized roadmap (roadmap.md, roadmap.json) and created todo list scaffold.
- **2025-12-06** — Defined Epic 1 data model: raw events, token_usage, aggregates (daily/weekly/monthly), indexes, retention/PII rules.
- **2025-12-06** — Implemented activity hooks (chat, RAG, file upload) logging to audit logs with user_id/company_id, timestamps, and metadata (no content stored).
- **2025-12-06** — Added token metering via Kolosal responses; logs prompt/completion/total tokens to token_usage (best-effort, non-blocking).
- **2025-12-06** — Added aggregation tables (activity_aggregates, token_usage_aggregates) and user_id on token_usage for per-user metrics; added aggregation services for daily rollups.
- **2025-12-06** — Added admin endpoints for activity aggregates and token usage aggregates with filters (company/user/action/model/provider/date) and pagination.
- **2025-12-06** — Added admin UI (Activity & Tokens) with filters (company/user/action/model/provider/date) and aggregated views; charts/widgets pending.
- **2025-12-06** — Implemented SSE live counters (chat/uploads/RAG/token) endpoint and wired to admin UI live widget.
- **2025-12-06** — Added system stats widgets (totals + top actions) on admin activity dashboard.
- **2025-12-06** — Added in-memory fan-out for live feed with backpressure (hub broadcaster) to serve SSE without per-connection DB polling; Redis optional.

## Notes
- Awaiting task breakdown and scheduling per epic.

