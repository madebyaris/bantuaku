# Implementation Todo List: Chat Integration Fix

## Overview
Implement full chat functionality with database persistence, conversation history loading, and infinite scroll. Backend needs DB implementation, frontend needs API wiring and UI updates.

## Pre-Implementation Setup
- [x] Review research findings
- [x] Confirm specification requirements
- [x] Validate technical plan
- [x] Set up development environment

## Todo Items

### Phase 1: Backend Database Implementation

- [x] **BE-001**: Implement `StartConversation` DB save (20min)
  - **Acceptance criteria:** Conversation saved to DB with company_id, user_id, purpose, title
  - **Files:** `backend/handlers/chat.go`
  - **Dependencies:** None
  - **Pattern:** Follow `insights.go` INSERT pattern using `h.db.Pool().Exec()`
  - **Details:** Generate UUID, insert into conversations table, return conversation ID

- [x] **BE-002**: Implement `SendMessage` DB save for user message (15min)
  - **Acceptance criteria:** User message saved to messages table before Kolosal call
  - **Files:** `backend/handlers/chat.go`
  - **Dependencies:** BE-001 (conversation must exist)
  - **Pattern:** INSERT into messages table with sender='user'
  - **Details:** Save message with conversation_id, sender='user', content

- [x] **BE-003**: Implement `SendMessage` DB save for assistant reply (15min)
  - **Acceptance criteria:** Assistant message saved to messages table after Kolosal response
  - **Files:** `backend/handlers/chat.go`
  - **Dependencies:** BE-002
  - **Pattern:** INSERT into messages table with sender='assistant'
  - **Details:** Save assistant_reply with conversation_id, sender='assistant', update conversation updated_at

- [x] **BE-004**: Implement `GetConversations` DB query with pagination (20min)
  - **Acceptance criteria:** Returns conversations filtered by company_id, ordered by updated_at DESC, supports limit/offset
  - **Files:** `backend/handlers/chat.go`
  - **Dependencies:** None
  - **Pattern:** Follow `insights.go` Query pattern, scan rows into ConversationSummary
  - **Details:** Parse limit (default 5) and offset from query params, return list

- [x] **BE-005**: Implement `GetMessages` DB query with pagination (20min)
  - **Acceptance criteria:** Returns messages filtered by conversation_id, ordered by created_at ASC, supports limit/offset
  - **Files:** `backend/handlers/chat.go`
  - **Dependencies:** None
  - **Pattern:** Follow `insights.go` Query pattern, scan rows into models.Message
  - **Details:** Parse limit/offset from query params, handle JSONB structured_payload

### Phase 2: Frontend API Client

- [x] **FE-001**: Add `api.chat.sendMessage()` method (10min)
  - **Acceptance criteria:** POST to `/api/v1/chat/message` with conversationId and message
  - **Files:** `frontend/src/lib/api.ts`
  - **Dependencies:** None
  - **Pattern:** Follow existing `api.chat.conversations.*` pattern
  - **Details:** Returns SendMessageResponse type

- [x] **FE-002**: Add `api.chat.startConversation()` method (10min)
  - **Acceptance criteria:** POST to `/api/v1/chat/start` with purpose
  - **Files:** `frontend/src/lib/api.ts`
  - **Dependencies:** None
  - **Pattern:** Follow existing API method pattern
  - **Details:** Returns StartConversationResponse type

- [x] **FE-003**: Update `api.chat.conversations.list()` with pagination (5min)
  - **Acceptance criteria:** Accepts optional limit and offset params, appends to query string
  - **Files:** `frontend/src/lib/api.ts`
  - **Dependencies:** None
  - **Pattern:** Use URLSearchParams for query string building
  - **Details:** Default limit=5 if not provided

- [x] **FE-004**: Update `api.chat.conversations.messages()` with pagination (5min)
  - **Acceptance criteria:** Accepts optional limit and offset params, appends to query string
  - **Files:** `frontend/src/lib/api.ts`
  - **Dependencies:** None
  - **Pattern:** Use URLSearchParams for query string building
  - **Details:** Default limit=50 if not provided

### Phase 3: Frontend State Management

- [x] **FE-005**: Extend chat store with conversations array and currentConversationId (15min)
  - **Acceptance criteria:** State has conversations: Conversation[], currentConversationId: string | null
  - **Files:** `frontend/src/state/chat.ts`
  - **Dependencies:** FE-003
  - **Pattern:** Follow existing Zustand store pattern
  - **Details:** Add to ChatState interface, initialize in store

- [x] **FE-006**: Add pagination state to chat store (10min)
  - **Acceptance criteria:** State has hasMoreConversations: boolean, conversationsOffset: number
  - **Files:** `frontend/src/state/chat.ts`
  - **Dependencies:** FE-005
  - **Pattern:** Add to ChatState interface
  - **Details:** Track pagination state for infinite scroll

- [x] **FE-007**: Add `loadConversations()` function to chat store (15min)
  - **Acceptance criteria:** Calls API, loads first 5 conversations, updates state
  - **Files:** `frontend/src/state/chat.ts`
  - **Dependencies:** FE-003, FE-005
  - **Pattern:** Async function that calls api.chat.conversations.list()
  - **Details:** Reset conversations array, set hasMoreConversations

- [x] **FE-008**: Add `loadMoreConversations()` function to chat store (15min)
  - **Acceptance criteria:** Calls API with offset, appends to conversations array
  - **Files:** `frontend/src/state/chat.ts`
  - **Dependencies:** FE-007
  - **Pattern:** Increment offset, append results
  - **Details:** Update hasMoreConversations based on results length

- [x] **FE-009**: Add `selectConversation()` function to chat store (10min)
  - **Acceptance criteria:** Sets currentConversationId, loads messages for conversation
  - **Files:** `frontend/src/state/chat.ts`
  - **Dependencies:** FE-004, FE-005
  - **Pattern:** Set conversation ID, call loadMessages()
  - **Details:** Clear current messages, load new ones

- [x] **FE-010**: Add `loadMessages()` function to chat store (20min)
  - **Acceptance criteria:** Calls API, loads messages, converts to ChatMessage format, updates messages array
  - **Files:** `frontend/src/state/chat.ts`
  - **Dependencies:** FE-004
  - **Pattern:** Call api.chat.conversations.messages(), map to ChatMessage
  - **Details:** Convert sender to role, map content to text, handle timestamps

### Phase 4: Frontend UI Integration

- [x] **FE-011**: Update ChatInterface to load conversations on mount (15min)
  - **Acceptance criteria:** useEffect calls loadConversations() when component mounts
  - **Files:** `frontend/src/components/chat/ChatInterface.tsx`
  - **Dependencies:** FE-007
  - **Pattern:** useEffect hook with empty dependency array
  - **Details:** Call loadConversations from store

- [x] **FE-012**: Update ChatInterface sendMessage() to use chat endpoint (20min)
  - **Acceptance criteria:** Creates conversation if needed, calls api.chat.sendMessage(), saves both messages
  - **Files:** `frontend/src/components/chat/ChatInterface.tsx`
  - **Dependencies:** FE-001, FE-002, FE-009
  - **Pattern:** Check currentConversationId, create if null, then send message
  - **Details:** Replace api.ai.analyze() call, handle response.assistant_reply

- [x] **FE-013**: Add conversation list sidebar to ChatInterface (30min)
  - **Acceptance criteria:** Shows list of conversations with title, last message preview, clickable
  - **Files:** `frontend/src/components/chat/ChatInterface.tsx`
  - **Dependencies:** FE-005
  - **Pattern:** Map over conversations array, render list items
  - **Details:** Show conversation title or "New Conversation", highlight active conversation

- [x] **FE-014**: Implement infinite scroll for conversations (25min)
  - **Acceptance criteria:** Scroll to top triggers loadMoreConversations(), shows loading state
  - **Files:** `frontend/src/components/chat/ChatInterface.tsx`
  - **Dependencies:** FE-008
  - **Pattern:** Use scroll event listener or IntersectionObserver
  - **Details:** Detect scroll to top, check hasMoreConversations, call loadMoreConversations

- [x] **FE-015**: Handle conversation selection click (15min)
  - **Acceptance criteria:** Clicking conversation calls selectConversation(), messages load
  - **Files:** `frontend/src/components/chat/ChatInterface.tsx`
  - **Dependencies:** FE-009
  - **Pattern:** onClick handler on conversation list items
  - **Details:** Call selectConversation(conversationId) from store

### Phase 5: Testing & Polish

- [ ] **TEST-001**: Test backend - Create conversation and verify DB (10min)
  - **Acceptance criteria:** POST /chat/start creates DB entry, returns conversation_id
  - **Files:** Manual testing or test file
  - **Dependencies:** BE-001
  - **Details:** Verify conversation in database

- [ ] **TEST-002**: Test backend - Send message and verify DB saves (10min)
  - **Acceptance criteria:** POST /chat/message saves both user and assistant messages
  - **Files:** Manual testing
  - **Dependencies:** BE-002, BE-003
  - **Details:** Check messages table has both entries

- [ ] **TEST-003**: Test backend - Get conversations with pagination (10min)
  - **Acceptance criteria:** GET /chat/conversations returns correct list, pagination works
  - **Files:** Manual testing
  - **Dependencies:** BE-004
  - **Details:** Test limit=5, offset=0, verify ordering

- [ ] **TEST-004**: Test backend - Get messages with pagination (10min)
  - **Acceptance criteria:** GET /chat/messages returns messages for conversation, ordered correctly
  - **Files:** Manual testing
  - **Dependencies:** BE-005
  - **Details:** Verify messages ordered by created_at ASC

- [ ] **TEST-005**: End-to-end test - Full chat flow (20min)
  - **Acceptance criteria:** Create conversation → Send message → See response → Load history → Scroll for more
  - **Files:** Manual testing in browser
  - **Dependencies:** All FE tasks
  - **Details:** Test complete user flow, verify persistence

- [ ] **TEST-006**: Error handling verification (10min)
  - **Acceptance criteria:** API failures show user-friendly errors, no crashes
  - **Files:** Manual testing
  - **Dependencies:** All tasks
  - **Details:** Test network errors, invalid responses

## Pattern Reuse Strategy

### Components to Reuse
- **insights.go** (backend/handlers/insights.go) - DB query pattern for GetConversations/GetMessages
- **ChatInterface.tsx** (frontend/src/components/chat/ChatInterface.tsx) - UI structure, just update data flow
- **api.ts** (frontend/src/lib/api.ts) - Request pattern for new API methods
- **chat.ts** (frontend/src/state/chat.ts) - Zustand store pattern

### Code Patterns to Follow
- **DB Queries**: Use `h.db.Pool().Query()` with parameterized queries ($1, $2), scan rows
- **DB Inserts**: Use `h.db.Pool().Exec()` with INSERT statements
- **JSONB Handling**: Marshal/Unmarshal JSON for structured_payload
- **API Requests**: Use `request<T>()` helper with proper types
- **State Management**: Extend Zustand store with new state and actions

## Execution Strategy

### Continuous Implementation Rules
1. Execute todo items in dependency order
2. Go for maximum flow - complete as much as possible without interruption
3. Group all ambiguous questions for batch resolution at the end
4. Reuse existing patterns and components wherever possible
5. Update progress continuously
6. Document any deviations from plan

### Checkpoint Schedule
- **Backend Complete**: After Phase 1 todos done - Verify DB operations work
- **API Client Complete**: After Phase 2 todos done - Verify API methods work
- **State Complete**: After Phase 3 todos done - Verify store functions work
- **UI Complete**: After Phase 4 todos done - Verify UI displays correctly
- **Ready to Test**: After Phase 5 todos done - Full system ready

## Progress Tracking

### Completed Items
- [x] Plan created and approved

### Blockers & Issues
- None yet

### Discoveries & Deviations
- Will document as implementation progresses

## Definition of Done
- [ ] All backend todos completed (BE-001 through BE-005)
- [ ] All frontend API todos completed (FE-001 through FE-004)
- [ ] All frontend state todos completed (FE-005 through FE-010)
- [ ] All frontend UI todos completed (FE-011 through FE-015)
- [ ] All testing todos completed (TEST-001 through TEST-006)
- [ ] Conversations persist in database
- [ ] Messages persist in database
- [ ] Past 5 conversations load on mount
- [ ] Infinite scroll loads more conversations
- [ ] Messages load when conversation selected
- [ ] Chat sends and receives messages successfully
- [ ] Error handling works correctly

---
**Created:** 2025-01-27  
**Estimated Duration:** ~4-5 hours  
**Implementation Start:** Now  
**Target Completion:** Today
