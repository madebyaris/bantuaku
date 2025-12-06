# OpenRouter Integration Feature Brief

## 🎯 Context (2min)
**Problem**: Need to add OpenRouter as the primary AI provider for chat while keeping Kolosal as an option. Currently, Kolosal is hardcoded and requires code changes to switch providers. Admins need a way to configure AI providers without code deployments.

**Users**: 
- **Admins**: Need to configure and switch AI providers via admin panel
- **End Users**: Using chat feature (unaffected by provider choice)

**Success**: 
- OpenRouter is the default provider for chat
- Kolosal remains available as an option
- Admins can switch providers via admin panel UI
- No code changes needed to switch providers
- Both providers work seamlessly with existing chat interface

## 🔍 Quick Research (15min)

### Existing Patterns

**1. Kolosal Client Service** (`backend/services/kolosal/client.go`)
- **Usage**: HTTP client for Kolosal.ai API with chat completion endpoint
- **Reusable**: Pattern can be replicated for OpenRouter client
- **Key Methods**: `CreateChatCompletion(ctx, req) -> (*ChatCompletionResponse, error)`
- **Timeout**: 120 seconds for long AI responses

**2. Embedding Provider Factory** (`backend/services/embedding/factory.go`)
- **Usage**: Factory pattern for selecting embedding providers (Kolosal, OpenAI, Cohere)
- **Reusable**: Same pattern can be used for chat provider selection
- **Pattern**: `NewEmbedder(cfg) -> (Embedder, error)` with switch statement
- **Config**: Uses `cfg.EmbeddingProvider` from config

**3. Admin Handler Pattern** (`backend/handlers/admin/users.go`)
- **Usage**: Admin API handlers with database queries, pagination, RBAC
- **Reusable**: Pattern for creating settings management endpoints
- **Structure**: Handler struct with db, logger, jwtSecret, auditLogger
- **Response**: JSON responses with proper error handling

**4. Admin UI Pattern** (`admin/src/pages/UsersPage.tsx`)
- **Usage**: React page with CRUD operations, modals, forms
- **Reusable**: Pattern for settings configuration page
- **Components**: Uses shadcn/ui (Button, Card, Select, Input)
- **State**: useState for data, loading, modals

**5. Database Migrations** (`database/migrations/`)
- **Usage**: SQL migrations for schema changes
- **Reusable**: Create new migration for settings table
- **Pattern**: Numbered files (001_, 002_, etc.)

### Tech Decision

**Approach**: 
1. Create OpenRouter client service (mirror Kolosal client pattern)
2. Create chat provider factory (similar to embedding factory)
3. Add database settings table for provider preference
4. Create admin API endpoints for settings management
5. Build admin UI for provider selection
6. Update chat handler to use factory-selected provider

**Why**: 
- **Reuses existing patterns**: Kolosal client, embedding factory, admin handlers
- **Maintains backward compatibility**: Kolosal still works, no breaking changes
- **Admin-configurable**: No code deployments needed to switch providers
- **Database-backed**: Settings persist across restarts
- **Follows project conventions**: Matches existing code structure

**Avoid**: 
- Environment variable-only approach (not admin-configurable)
- Removing Kolosal support (breaks existing setup)
- Hardcoding provider selection (defeats the purpose)
- Creating separate chat handlers per provider (too much duplication)

### OpenRouter API Research

**Base URL**: `https://openrouter.ai/api/v1`
**Endpoint**: `/chat/completions` (OpenAI-compatible)
**Authentication**: `Authorization: Bearer sk-or-v1-...`
**Request Format**: Same as OpenAI/Kolosal (messages array, model, temperature, etc.)
**Models**: Wide selection (gpt-4, claude, llama, etc.)
**Timeout**: Similar to Kolosal (30-120 seconds for responses)

## ✅ Requirements (10min)

**Core Requirements:**

1. **OpenRouter Client Service**
   - Story: As a developer, I need an OpenRouter client service so chat can use OpenRouter API
   - Acceptance: 
     - `backend/services/openrouter/client.go` exists
     - Implements `CreateChatCompletion()` method
     - Handles authentication, timeouts, error responses
     - Returns compatible response format

2. **Chat Provider Factory**
   - Story: As a developer, I need a factory to select chat providers dynamically
   - Acceptance:
     - `backend/services/chat/factory.go` exists
     - `NewChatProvider(cfg, settings)` returns provider client
     - Supports "openrouter" and "kolosal" providers
     - Defaults to OpenRouter if not configured

3. **Settings Database Table**
   - Story: As a system, I need to store AI provider preference persistently
   - Acceptance:
     - Migration creates `settings` table
     - Table has `key` (string), `value` (jsonb), `updated_at` (timestamp)
     - Can store `ai_provider` setting
     - Supports future settings expansion

4. **Admin Settings API**
   - Story: As an admin, I need API endpoints to get/update AI provider settings
   - Acceptance:
     - `GET /api/v1/admin/settings/ai-provider` returns current provider
     - `PUT /api/v1/admin/settings/ai-provider` updates provider
     - Requires admin role (RBAC)
     - Returns proper error responses

5. **Admin Settings UI**
   - Story: As an admin, I need a UI to configure AI provider
   - Acceptance:
     - Settings page/section in admin panel
     - Dropdown to select provider (OpenRouter, Kolosal)
     - Save button updates setting
     - Shows current provider
     - Success/error toast notifications

6. **Chat Handler Integration**
   - Story: As a user, I need chat to use the configured provider
   - Acceptance:
     - Chat handler uses provider factory
     - Works with both OpenRouter and Kolosal
     - No code changes needed when provider switches
     - Maintains existing chat functionality

7. **Configuration Updates**
   - Story: As a developer, I need environment variables for OpenRouter API key
   - Acceptance:
     - `OPENROUTER_API_KEY` env var supported
     - Added to `config.go` Config struct
     - Added to `docker-compose.yml`
     - Documented in README

## 🏗️ Implementation (5min)

**Components**:
- Backend: OpenRouter client service (`backend/services/openrouter/client.go`)
- Backend: Chat provider factory (`backend/services/chat/factory.go`)
- Backend: Settings service (`backend/services/settings/service.go`)
- Backend: Admin settings handler (`backend/handlers/admin/settings.go`)
- Database: Settings table migration (`database/migrations/009_add_settings_table.sql`)
- Admin UI: Settings page (`admin/src/pages/SettingsPage.tsx`)
- Config: OpenRouter API key support (`backend/config/config.go`)

**APIs**:
- `GET /api/v1/admin/settings/ai-provider` - Get current AI provider
- `PUT /api/v1/admin/settings/ai-provider` - Update AI provider
- (Internal) Settings service methods for reading/writing settings

**Data**:
- New `settings` table:
  - `id` (uuid, primary key)
  - `key` (varchar, unique)
  - `value` (jsonb)
  - `created_at` (timestamp)
  - `updated_at` (timestamp)
- Initial row: `key='ai_provider'`, `value='{"provider": "openrouter"}'`

**Files Modified**:
- `backend/handlers/chat.go` - Use provider factory instead of direct Kolosal client
- `backend/config/config.go` - Add `OpenRouterAPIKey` field
- `backend/main.go` - Register admin settings routes
- `docker-compose.yml` - Add `OPENROUTER_API_KEY` env var
- `admin/src/App.tsx` - Add Settings route
- `admin/src/components/layout/Sidebar.tsx` - Add Settings menu item

## 📋 Next Actions (2min)

- [ ] Create OpenRouter client service (2h)
- [ ] Create chat provider factory (1h)
- [ ] Create settings table migration (30min)
- [ ] Create settings service (1h)
- [ ] Create admin settings API handler (2h)
- [ ] Update chat handler to use factory (1h)
- [ ] Build admin settings UI page (3h)
- [ ] Add OpenRouter config support (30min)
- [ ] Test both providers (1h)
- [ ] Update documentation (30min)

**Start Coding In**: ~30 minutes

---
**Total Planning Time**: ~30min | **Owner**: Development Team | 2025-12-05

<!-- Living Document - Update as you code -->

## 🔄 Implementation Tracking

**CRITICAL**: Follow the todo-list systematically. Mark items as complete, document blockers, update progress.

### Progress
- [x] Brief created
- [x] Implementation complete
- [x] Backend: OpenRouter client, factory, settings service, admin API
- [x] Frontend: Admin settings UI page
- [ ] Testing pending
- [ ] Documentation update pending

### Blockers
- None identified

### Notes
- OpenRouter API is OpenAI-compatible, so request/response format should be similar
- Need to verify OpenRouter API key format and authentication
- Consider adding provider health check/validation in admin UI
- May want to add provider-specific model selection in future

**See**: [.sdd/IMPLEMENTATION_GUIDE.md](mdc:.sdd/IMPLEMENTATION_GUIDE.md) for detailed execution rules.
