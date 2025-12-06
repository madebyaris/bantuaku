# Implementation Todo List: OpenRouter Integration with Admin Configuration

## Overview
Add OpenRouter as primary AI provider for chat while keeping Kolosal as option. Make provider selection configurable via admin panel. Default to OpenRouter.

## Pre-Implementation Setup
- [x] Review feature brief
- [x] Understand existing patterns (Kolosal client, embedding factory, admin handlers)
- [x] Set up development environment

## Todo Items

### Phase 1: Backend - OpenRouter Client Service

- [x] **BE-001**: Create OpenRouter client service (2h)
  - **Acceptance criteria:** 
    - `backend/services/openrouter/client.go` exists
    - Implements `CreateChatCompletion()` method matching Kolosal interface
    - Handles authentication, timeouts (120s), error responses
    - Returns compatible `ChatCompletionResponse` format
  - **Files:** `backend/services/openrouter/client.go`
  - **Dependencies:** None
  - **Pattern:** Mirror `backend/services/kolosal/client.go` structure
  - **Details:** 
    - Base URL: `https://openrouter.ai/api/v1`
    - Endpoint: `/chat/completions`
    - Auth: `Authorization: Bearer {apiKey}`
    - Same request/response structs as Kolosal

### Phase 2: Backend - Chat Provider Factory

- [x] **BE-002**: Create chat provider interface (30min)
  - **Acceptance criteria:**
    - `backend/services/chat/interface.go` defines `ChatProvider` interface
    - Interface has `CreateChatCompletion()` method
  - **Files:** `backend/services/chat/interface.go`
  - **Dependencies:** None
  - **Pattern:** Similar to `backend/services/embedding/interface.go`

- [x] **BE-003**: Create chat provider factory (1h)
  - **Acceptance criteria:**
    - `backend/services/chat/factory.go` exists
    - `NewChatProvider(cfg, settings)` returns ChatProvider
    - Supports "openrouter" and "kolosal" providers
    - Defaults to "openrouter" if not configured
  - **Files:** `backend/services/chat/factory.go`
  - **Dependencies:** BE-001, BE-002
  - **Pattern:** Similar to `backend/services/embedding/factory.go`
  - **Details:** 
    - Check settings service for provider preference
    - Fallback to config/env if settings not available
    - Return appropriate client (OpenRouter or Kolosal)

### Phase 3: Backend - Settings Service & Database

- [x] **BE-004**: Create settings table migration (30min)
  - **Acceptance criteria:**
    - Migration file `009_add_settings_table.sql` exists
    - Creates `settings` table with id, key (unique), value (jsonb), timestamps
    - Inserts default `ai_provider` setting with value "openrouter"
  - **Files:** `database/migrations/009_add_settings_table.sql`
  - **Dependencies:** None
  - **Pattern:** Follow existing migration file numbering

- [x] **BE-005**: Create settings service (1h)
  - **Acceptance criteria:**
    - `backend/services/settings/service.go` exists
    - `GetSetting(ctx, key)` returns setting value
    - `SetSetting(ctx, key, value)` updates/creates setting
    - Handles JSONB value encoding/decoding
  - **Files:** `backend/services/settings/service.go`
  - **Dependencies:** BE-004
  - **Pattern:** Similar to other service patterns in codebase
  - **Details:** 
    - Use pgxpool for database access
    - Return string values (JSONB decoded)
    - Handle missing settings gracefully

### Phase 4: Backend - Admin Settings API

- [x] **BE-006**: Create admin settings handler (2h)
  - **Acceptance criteria:**
    - `backend/handlers/admin/settings.go` exists
    - `GetAIProvider()` returns current provider
    - `UpdateAIProvider()` updates provider (validates "openrouter" or "kolosal")
    - Requires admin role (RBAC)
    - Proper error handling and JSON responses
  - **Files:** `backend/handlers/admin/settings.go`
  - **Dependencies:** BE-005
  - **Pattern:** Follow `backend/handlers/admin/users.go` pattern
  - **Details:**
    - Use AdminHandler struct pattern
    - Validate provider value
    - Return JSON: `{"provider": "openrouter"}`

- [x] **BE-007**: Register admin settings routes (15min)
  - **Acceptance criteria:**
    - Routes registered in `main.go`
    - `GET /api/v1/admin/settings/ai-provider` mapped
    - `PUT /api/v1/admin/settings/ai-provider` mapped
    - Protected with admin auth middleware
  - **Files:** `backend/main.go`
  - **Dependencies:** BE-006
  - **Pattern:** Follow existing admin route registration

### Phase 5: Backend - Chat Handler Integration

- [x] **BE-008**: Update chat handler to use provider factory (1h)
  - **Acceptance criteria:**
    - Chat handler uses `chat.NewChatProvider()` instead of direct Kolosal client
    - Works with both OpenRouter and Kolosal
    - No breaking changes to API response format
  - **Files:** `backend/handlers/chat.go`
  - **Dependencies:** BE-003, BE-005
  - **Pattern:** Replace direct client instantiation with factory call
  - **Details:**
    - Get provider from factory using config and settings
    - Use provider interface methods
    - Maintain existing error handling and logging

### Phase 6: Backend - Configuration

- [x] **BE-009**: Add OpenRouter API key to config (30min)
  - **Acceptance criteria:**
    - `OpenRouterAPIKey` field added to Config struct
    - Loads from `OPENROUTER_API_KEY` env var
    - Added to docker-compose.yml
  - **Files:** `backend/config/config.go`, `docker-compose.yml`
  - **Dependencies:** None
  - **Pattern:** Follow existing KolosalAPIKey pattern

### Phase 7: Frontend - Admin Settings UI

- [x] **FE-001**: Add settings API client methods (30min)
  - **Acceptance criteria:**
    - `api.admin.settings.getAIProvider()` method exists
    - `api.admin.settings.updateAIProvider(provider)` method exists
    - Proper TypeScript types
  - **Files:** `admin/src/lib/api.ts`
  - **Dependencies:** None
  - **Pattern:** Follow existing admin API methods

- [x] **FE-002**: Create Settings page component (3h)
  - **Acceptance criteria:**
    - `admin/src/pages/SettingsPage.tsx` exists
    - Shows current AI provider
    - Dropdown to select provider (OpenRouter, Kolosal)
    - Save button updates setting
    - Loading states and error handling
    - Success/error toast notifications
  - **Files:** `admin/src/pages/SettingsPage.tsx`
  - **Dependencies:** FE-001
  - **Pattern:** Follow `admin/src/pages/UsersPage.tsx` pattern
  - **Details:**
    - Use shadcn/ui components (Card, Select, Button)
    - useState for form state
    - useEffect to load current setting on mount

- [x] **FE-003**: Add Settings route to admin app (15min)
  - **Acceptance criteria:**
    - Route `/settings` added to App.tsx
    - Protected with admin auth
  - **Files:** `admin/src/App.tsx`
  - **Dependencies:** FE-002
  - **Pattern:** Follow existing route patterns

- [x] **FE-004**: Add Settings menu item to sidebar (15min)
  - **Acceptance criteria:**
    - Settings link added to sidebar navigation
    - Icon and label appropriate
  - **Files:** `admin/src/components/layout/Sidebar.tsx`
  - **Dependencies:** FE-003
  - **Pattern:** Follow existing sidebar menu items

### Phase 8: Testing & Documentation

- [ ] **TEST-001**: Test OpenRouter provider (30min)
  - **Acceptance criteria:**
    - Chat works with OpenRouter provider
    - Responses are correct format
    - Error handling works
  - **Dependencies:** All backend tasks complete

- [ ] **TEST-002**: Test Kolosal provider still works (30min)
  - **Acceptance criteria:**
    - Can switch to Kolosal via admin UI
    - Chat works with Kolosal provider
    - No regressions
  - **Dependencies:** All tasks complete

- [ ] **TEST-003**: Test admin settings UI (30min)
  - **Acceptance criteria:**
    - Can view current provider
    - Can switch providers
    - Changes persist after restart
  - **Dependencies:** All tasks complete

- [ ] **DOC-001**: Update README with OpenRouter setup (30min)
  - **Acceptance criteria:**
    - OPENROUTER_API_KEY documented
    - Admin settings usage documented
  - **Files:** `README.md`
  - **Dependencies:** All tasks complete

## Progress Tracking

### Completed
- [x] Brief created
- [x] Todo-list created
- [x] All backend implementation tasks (BE-001 through BE-009)
- [x] All frontend implementation tasks (FE-001 through FE-004)
- [x] Docker configuration updated

### In Progress
- [ ] Testing phase

### Blockers
- None

## Notes
- OpenRouter API is OpenAI-compatible, so request/response format matches Kolosal
- Provider selection is runtime-configurable via admin UI
- Settings are stored in database, persist across restarts
- Default provider is OpenRouter if not configured
