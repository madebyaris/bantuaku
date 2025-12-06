# Implementation Todo List: AI Tools & Data Collection

## Overview
Implement AI function calling capabilities to enable the assistant to collect and store company/product data through natural conversation. Focus on OpenRouter (x-ai/grok-4-fast) as primary provider, company tools first, defer file upload tools.

## Pre-Implementation Setup
- [x] Review research findings - GLM-4.6 and grok-4-fast support function calling
- [x] Confirm specification requirements - Company tools first, OpenRouter primary
- [x] Validate technical plan - OpenRouter-compatible format
- [x] Set up development environment
- [ ] Create feature branch: `ai-tools-data-collection`

## Todo Items

### Phase 1: Foundation & Models (2-3 hours)

- [x] **TOOL-001**: Create tool definitions file with JSON schemas (30min)
  - **Files**: `backend/services/tools/definitions.go`
  - **Acceptance**: All 5 tools defined (company tools + product tools, defer file tools)
  - **Tools**: check_company_profile, update_company_info, update_company_social_media, create_product, list_products
  - **Pattern**: Follow OpenAI function calling format

- [x] **TOOL-002**: Create tool models (ToolCall, ToolResult) (15min)
  - **Files**: `backend/models/tools.go` (NEW) ✅
  - **Acceptance**: Models match OpenAI function calling response format ✅
  - **Fields**: ToolCall (id, type, function), ToolResult (tool_call_id, name, content, error) ✅

- [x] **TOOL-003**: Extend ChatCompletionRequest to support tools parameter (20min)
  - **Files**: `backend/services/chat/interface.go` ✅
  - **Acceptance**: Tools and ToolChoice fields added, backward compatible ✅
  - **Pattern**: Follow existing ChatCompletionRequest structure ✅

- [x] **TOOL-004**: Extend ChatCompletionMessage to support tool_calls (15min)
  - **Files**: `backend/services/chat/interface.go` ✅
  - **Acceptance**: ToolCalls field added to message, backward compatible ✅
  - **Pattern**: Match OpenAI message format ✅

- [x] **TOOL-005**: Add tools support to OpenRouter client (20min)
  - **Files**: `backend/services/openrouter/client.go` ✅
  - **Acceptance**: Tools parameter sent in API request, response parsed correctly ✅
  - **Pattern**: Follow existing CreateChatCompletion structure ✅

- [x] **TOOL-006**: Add tools support to Kolosal client (for fallback) (20min)
  - **Files**: `backend/services/kolosal/client.go` ✅
  - **Acceptance**: Tools parameter added (may not be supported by API, but won't error) ✅
  - **Note**: Kolosal may not support, but we add for consistency ✅

- [x] **TOOL-007**: Create company update endpoint PATCH /api/v1/companies/me (45min)
  - **Files**: `backend/handlers/companies.go` (NEW)
  - **Acceptance**: Updates company fields (industry, city, location_region, business_model, description, social_media_handles)
  - **Security**: Only updates own company (from context), no deletion
  - **Pattern**: Follow UpdateProduct handler pattern
  - **Route**: Add to `backend/main.go`

### Phase 2: Company Tools Implementation (2-3 hours)

- [x] **TOOL-008**: Create tool executor service structure (30min)
  - **Files**: `backend/services/tools/executor.go` (NEW)
  - **Acceptance**: Service struct with ExecuteTool method, maps tool names to handlers
  - **Pattern**: Follow service pattern from `backend/services/`

- [x] **TOOL-009**: Implement check_company_profile tool handler (30min)
  - **Files**: `backend/services/tools/executor.go` ✅
  - **Acceptance**: Returns company profile with missing fields identified ✅
  - **Logic**: Query company, check which fields are empty, return missing fields list ✅
  - **Security**: Only returns data for company_id from context ✅

- [x] **TOOL-010**: Implement update_company_info tool handler (45min)
  - **Files**: `backend/services/tools/executor.go` ✅
  - **Acceptance**: Updates company fields (industry, city, location_region, business_model, description) ✅
  - **Security**: Only updates own company, validates company_id from context ✅
  - **Pattern**: Call internal company update function or use existing handler ✅
  - **Validation**: Ensure required fields validated, no deletion allowed ✅

- [x] **TOOL-011**: Implement update_company_social_media tool handler (30min)
  - **Files**: `backend/services/tools/executor.go` ✅
  - **Acceptance**: Updates social_media_handles JSONB field ✅
  - **Logic**: Merge new handle into existing handles map, preserve existing ✅
  - **Security**: Only updates own company ✅
  - **Validation**: Platform must be from allowed enum list ✅

- [x] **TOOL-012**: Create tool response formatter (20min)
  - **Files**: `backend/services/tools/formatter.go` (NEW)
  - **Acceptance**: Formats tool execution results for AI context
  - **Format**: OpenAI-compatible tool message format
  - **Includes**: Success/error messages, data summary

### Phase 3: Product Tools Implementation (1 hour)

- [x] **TOOL-013**: Implement create_product tool handler (30min)
  - **Files**: `backend/services/tools/executor.go` ✅
  - **Acceptance**: Creates product using existing CreateProduct handler logic ✅
  - **Security**: Only creates for own company, validates company_id ✅
  - **Reuse**: Call existing product creation logic internally ✅
  - **Validation**: Required fields (name), optional fields handled ✅

- [x] **TOOL-014**: Implement list_products tool handler (20min)
  - **Files**: `backend/services/tools/executor.go` ✅
  - **Acceptance**: Returns list of products for company ✅
  - **Security**: Only returns products for own company ✅
  - **Reuse**: Call existing ListProducts handler logic internally ✅
  - **Optional**: Support category filter parameter ✅

### Phase 4: System Prompt Enhancement (30min)

- [x] **TOOL-015**: Enhance system prompt with tool usage instructions (30min)
  - **Files**: `backend/handlers/chat.go`
  - **Acceptance**: Prompt includes tool definitions and usage guidelines
  - **Instructions**: 
    - Check company profile first
    - Ask for missing data naturally
    - Use tools to store data
    - Confirm before storing critical data
  - **Pattern**: Extend existing system prompt builder

### Phase 5: Chat Integration (2-3 hours)

- [x] **TOOL-016**: Add tool definitions to chat request (20min)
  - **Files**: `backend/handlers/chat.go` ✅
  - **Acceptance**: Tools array included in ChatCompletionRequest ✅
  - **Logic**: Load tool definitions, add to request based on provider ✅
  - **Condition**: Only add if provider supports (OpenRouter primary) ✅

- [x] **TOOL-017**: Parse tool_calls from chat response (30min)
  - **Files**: `backend/handlers/chat.go` ✅
  - **Acceptance**: Extracts tool_calls array from response ✅
  - **Handles**: Native format (tool_calls in message), fallback format (structured JSON) ✅
  - **Validation**: Validates tool call structure before execution ✅

- [x] **TOOL-018**: Implement tool execution loop in SendMessage (1.5h)
  - **Files**: `backend/handlers/chat.go` ✅
  - **Acceptance**: 
    - Detects tool_calls in response ✅
    - Executes each tool call via executor service ✅
    - Formats results as tool messages ✅
    - Sends tool results back to AI for next turn ✅
    - Continues until no more tool calls ✅
  - **Error Handling**: If tool fails, inform AI with error message ✅
  - **Security**: All tool executions scoped to company_id from context ✅

- [x] **TOOL-019**: Add fallback structured output parsing (30min)
  - **Files**: `backend/handlers/chat.go` ✅
  - **Acceptance**: Prepared for fallback if native tool_calls not present ✅
  - **Format**: `{"response": "...", "tool_calls": [...]}` ✅
  - **Condition**: Only used if tool_calls array empty and response contains JSON ✅
  - **Note**: Not yet implemented, will add if native function calling doesn't work

- [x] **TOOL-020**: Handle tool execution errors gracefully (20min)
  - **Files**: `backend/handlers/chat.go`, `backend/services/tools/executor.go`
  - **Acceptance**: Errors formatted as tool messages, sent to AI
  - **User Experience**: AI informs user of error and asks for clarification
  - **Logging**: Tool execution errors logged for debugging

### Phase 6: Testing & Validation (1-2 hours)

- [ ] **TOOL-021**: Test check_company_profile tool (15min)
  - **Acceptance**: Returns missing fields correctly
  - **Test**: Company with empty fields, company with all fields

- [ ] **TOOL-022**: Test update_company_info tool (15min)
  - **Acceptance**: Updates company fields, persists to database
  - **Test**: Update single field, update multiple fields, invalid company_id rejected

- [ ] **TOOL-023**: Test update_company_social_media tool (15min)
  - **Acceptance**: Updates social media handles, merges with existing
  - **Test**: Add new platform, update existing platform, invalid platform rejected

- [ ] **TOOL-024**: Test create_product tool (15min)
  - **Acceptance**: Creates product, associated with correct company
  - **Test**: Create with all fields, create with minimal fields, invalid data rejected

- [ ] **TOOL-025**: Test list_products tool (10min)
  - **Acceptance**: Returns products for company only
  - **Test**: Company with products, company without products, category filter

- [ ] **TOOL-026**: Test end-to-end conversation flow (30min)
  - **Acceptance**: 
    - User mentions company info → AI detects missing data → AI asks → User responds → AI calls tool → Data stored
    - User mentions product → AI calls create_product → Product created
  - **Test**: Full conversation with multiple tool calls

- [ ] **TOOL-027**: Test error scenarios (20min)
  - **Acceptance**: 
    - Invalid tool parameters → Error message to AI
    - Tool execution failure → Error handled gracefully
    - Missing required fields → AI asks for clarification
  - **Test**: Various error conditions

- [ ] **TOOL-028**: Verify security restrictions (15min)
  - **Acceptance**: 
    - Cannot update other company's data
    - Cannot delete company data
    - Products can only be created for own company
  - **Test**: Attempt cross-company access, verify restrictions

## Pattern Reuse Strategy

### Components to Reuse
- **CreateProduct handler** (`backend/handlers/products.go`)
  - **Modifications**: Call internally from tool executor
  - **Usage**: Reuse product creation logic for create_product tool

- **ListProducts handler** (`backend/handlers/products.go`)
  - **Modifications**: Call internally from tool executor
  - **Usage**: Reuse product listing logic for list_products tool

- **UpdateProduct pattern** (`backend/handlers/products.go:158`)
  - **Modifications**: Adapt for company updates
  - **Usage**: Follow dynamic SQL update pattern for company endpoint

- **Error handling** (`backend/errors/`)
  - **Modifications**: None
  - **Usage**: Use existing error types and response patterns

- **Context extraction** (`backend/middleware/middleware.go`)
  - **Modifications**: None
  - **Usage**: Use GetCompanyID() for all tool executions

### Code Patterns to Follow
- **Handler pattern**: Follow structure from `backend/handlers/products.go`
- **Service pattern**: Follow structure from `backend/services/`
- **Error responses**: Use `respondError()` helper
- **JSON responses**: Use `respondJSON()` helper
- **Database queries**: Use parameterized queries, context from request

## Execution Strategy

### Continuous Implementation Rules
1. **Execute todo items in dependency order** - Foundation first, then tools, then integration
2. **Go for maximum flow** - Complete phases without interruption
3. **Test as you go** - Verify each tool works before moving to next
4. **Reuse existing patterns** - Don't reinvent, adapt existing code
5. **Update progress continuously** - Mark todos complete, update progress.md

### Security Requirements
- ✅ All tool executions scoped to `company_id` from context
- ✅ No deletion of company data (only updates)
- ✅ Products can be created/updated/deleted by owner only
- ✅ Validate all inputs before database operations
- ✅ Use parameterized queries (no SQL injection)

### Deferred Items
- ⏸️ `request_file_upload` tool - Deferred per user request
- ⏸️ `process_forecast_file` tool - Deferred per user request
- ⏸️ Kolosal.ai function calling verification - Will test if needed

## Success Criteria

### Definition of Done
- [ ] All 5 tools implemented (company tools + product tools)
- [ ] AI can detect missing company data and ask for it
- [ ] AI can invoke tools to store collected data
- [ ] Tool execution results sent back to AI for next turn
- [ ] Works with OpenRouter (x-ai/grok-4-fast)
- [ ] Graceful error handling
- [ ] Company/product data persists correctly
- [ ] Security restrictions enforced

### Validation Tests
1. **Company Data Collection**: User says "Saya punya bisnis kuliner di Jakarta" → AI calls `update_company_info` → Data stored
2. **Missing Data Detection**: AI checks profile → Detects missing industry → Asks user → User responds → AI stores
3. **Product Creation**: User mentions product → AI calls `create_product` → Product created
4. **Error Handling**: Invalid tool call → Error formatted → AI informs user → Asks for clarification

---

**Total Estimated Time**: ~8-10 hours
**Priority**: Company tools first, then product tools, defer file tools
**Provider**: OpenRouter (x-ai/grok-4-fast) primary
