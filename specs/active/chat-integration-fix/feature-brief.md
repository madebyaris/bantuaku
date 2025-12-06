# Chat Integration Fix Feature Brief

## 🎯 Context (2min)
**Problem**: Chat interface exists but doesn't work because:
1. Frontend calls wrong API endpoint (`/ai/analyze` instead of `/chat/message`)
2. Backend chat handlers have TODOs - conversations and messages aren't saved to database
3. No conversation history loading - users can't see past chats
4. No infinite scroll for loading more conversations

**Users**: End users trying to use the AI chat feature in the application

**Success**: 
- Users can send messages and receive responses from Kolosal.ai API
- Conversations are saved and persist across sessions
- Users see past 5 conversations on load
- Users can scroll up to load more conversations (infinite scroll)
- Messages load for selected conversation

## 🔍 Quick Research (15min)

### Existing Patterns

#### Backend Chat Implementation ⚠️
- **Location**: `backend/handlers/chat.go`
- **Endpoints**: 
  - `POST /api/v1/chat/start` - Create conversation (has TODO, not saving to DB)
  - `POST /api/v1/chat/message` - Send message (has TODO, not saving to DB)
  - `GET /api/v1/chat/conversations` - List conversations (has TODO, returns empty)
  - `GET /api/v1/chat/messages?conversation_id=...` - Get messages (has TODO, returns empty)
- **Database Tables**: ✅ Exist (`conversations`, `messages` tables in migration 003)
- **Models**: ✅ `models.Conversation` and `models.Message` exist
- **Kolosal Integration**: ✅ Already configured and working in SendMessage
- **RAG Support**: ✅ Includes RAG context retrieval
- **Issue**: Handlers need DB implementation (currently return empty/mock data)
- **Pattern**: Use `h.db.Pool().Query()` and `h.db.Pool().Exec()` like in `insights.go`

#### Frontend Chat Component ✅
- **Location**: `frontend/src/components/chat/ChatInterface.tsx`
- **Current Issue**: Calls `api.ai.analyze()` which goes to `/ai/analyze` endpoint
- **State Management**: Uses Zustand store (`useChatStore`) ✅
- **UI**: Complete chat interface with messages, loading states ✅
- **Reuse**: Component is good, just needs correct API call

#### API Client ❌
- **Location**: `frontend/src/lib/api.ts`
- **Missing**: `api.chat.sendMessage()` method
- **Has**: `api.chat.conversations.*` methods but no send message
- **Reuse**: Add new method following existing pattern

#### Kolosal API Client ✅
- **Location**: `backend/services/kolosal/client.go`
- **Status**: Fully implemented with `CreateChatCompletion()` method
- **Config**: API key loaded from `KOLOSAL_API_KEY` env var ✅
- **Model**: Uses "default" model from Kolosal.ai
- **Reuse**: Already working, no changes needed

### Tech Decision

**Approach**: 
1. **Backend**: Implement database persistence for conversations and messages
2. **Backend**: Add pagination support for conversations (limit 5 initially, then load more)
3. **Backend**: Add pagination support for messages (for infinite scroll)
4. **Frontend**: Wire up to chat endpoints, add conversation list UI, implement infinite scroll

**Why**: 
- Database tables already exist, just need to implement handlers
- Kolosal integration is working, just need to save responses
- Frontend UI is complete, needs conversation list and loading logic
- Standard chat app pattern (load recent, scroll for more)

**Avoid**: 
- Creating new database tables (already exist)
- Rewriting chat component (UI is fine, just needs data)
- Mock data (users need real persistence)

## ✅ Requirements (10min)

### Must-Have (MVP)

#### Backend
1. **Save Conversations** → `StartConversation` saves to `conversations` table with company_id, user_id, purpose
2. **Save Messages** → `SendMessage` saves both user message and assistant reply to `messages` table
3. **Load Conversations** → `GetConversations` queries DB, returns last 5 conversations ordered by `updated_at DESC`
4. **Load Messages** → `GetMessages` queries DB for conversation, returns messages ordered by `created_at ASC`
5. **Pagination Support** → Add `limit` and `offset` query params for loading more conversations/messages

#### Frontend
6. **API Client Methods** → Add:
   - `api.chat.sendMessage(conversationId, message)`
   - `api.chat.startConversation(purpose)` 
   - `api.chat.conversations.list(limit?, offset?)` (update existing)
   - `api.chat.conversations.messages(conversationId, limit?, offset?)` (update existing)
7. **Conversation List UI** → Show sidebar/list of past 5 conversations with titles
8. **Load More Conversations** → Infinite scroll (scroll up) to load more conversations
9. **Load Messages** → When conversation selected, load its messages from API
10. **Update ChatInterface** → Replace `api.ai.analyze()` with `api.chat.sendMessage()`, save messages to state
11. **Conversation Selection** → Click conversation → load messages → show in chat interface

### Nice-to-Have (Post-MVP)
- Conversation search/filter
- Conversation deletion
- File upload integration
- Citations display in UI
- RAG indicator in UI
- Real-time message updates (WebSocket)

## 🏗️ Implementation (5min)

### Backend Components to Modify

1. **`backend/handlers/chat.go`**
   - **StartConversation**: 
     - Generate UUID for conversation ID
     - Insert into `conversations` table with company_id, user_id, purpose, title
     - Return conversation ID and title
   - **SendMessage**:
     - Save user message to `messages` table (sender='user')
     - After getting Kolosal response, save assistant message (sender='assistant')
     - Update conversation `updated_at` timestamp
     - Return message_id and assistant_reply
   - **GetConversations**:
     - Query `conversations` table filtered by company_id
     - Order by `updated_at DESC`
     - Support `limit` (default 5) and `offset` query params
     - Return list of ConversationSummary
   - **GetMessages**:
     - Query `messages` table filtered by conversation_id
     - Order by `created_at ASC` (oldest first)
     - Support `limit` and `offset` for pagination
     - Return list of Message models

2. **Database Pattern** (follow `insights.go` pattern):
   ```go
   rows, err := h.db.Pool().Query(ctx, `
       SELECT id, company_id, user_id, title, purpose, created_at, updated_at
       FROM conversations
       WHERE company_id = $1
       ORDER BY updated_at DESC
       LIMIT $2 OFFSET $3
   `, companyID, limit, offset)
   ```

### Frontend Components to Modify

1. **`frontend/src/lib/api.ts`**
   - Add `sendMessage(conversationId, message)` to `api.chat`
   - Add `startConversation(purpose)` to `api.chat`
   - Update `conversations.list(limit?, offset?)` to support pagination
   - Update `conversations.messages(conversationId, limit?, offset?)` to support pagination

2. **`frontend/src/components/chat/ChatInterface.tsx`**
   - Add conversation selection state
   - On mount: Load past 5 conversations
   - Update `sendMessage()` to:
     - Create conversation if none selected (call `startConversation`)
     - Call `api.chat.sendMessage()` instead of `api.ai.analyze()`
     - Save both user and assistant messages to state
   - Add infinite scroll handler for loading more conversations
   - Add conversation list sidebar/component

3. **`frontend/src/state/chat.ts`**
   - Add `conversations: Conversation[]` array
   - Add `currentConversationId: string | null`
   - Add `hasMoreConversations: boolean` for pagination
   - Add `loadConversations()`, `loadMoreConversations()`, `selectConversation()`
   - Add `loadMessages(conversationId)` for loading message history

4. **New Component: `ConversationList.tsx`** (optional, can be in ChatInterface)
   - Display list of conversations
   - Show conversation title, last message preview, timestamp
   - Handle click to select conversation
   - Infinite scroll trigger (scroll up to load more)

### API Endpoints Used
- `POST /api/v1/chat/start` - Create conversation (now saves to DB)
- `POST /api/v1/chat/message` - Send message (now saves to DB)
- `GET /api/v1/chat/conversations?limit=5&offset=0` - List conversations (now queries DB)
- `GET /api/v1/chat/messages?conversation_id=...&limit=50&offset=0` - Get messages (now queries DB)

### Data Changes
- **Backend**: Implement DB queries for conversations and messages tables
- **Frontend**: Add conversation list state, pagination state, message history loading

### Backend Status
- ✅ Kolosal API key configured
- ✅ Database tables exist
- ✅ Models exist
- ⚠️ **Handlers need DB implementation** (currently TODOs)
- ✅ RAG integration ready
- ✅ Error handling pattern exists

## 📋 Next Actions (2min)

### Backend Tasks
- [ ] Implement `StartConversation` DB save (insert into conversations table) (20min)
- [ ] Implement `SendMessage` DB save (insert user + assistant messages) (25min)
- [ ] Implement `GetConversations` DB query with pagination (limit/offset) (20min)
- [ ] Implement `GetMessages` DB query with pagination (limit/offset) (20min)
- [ ] Test backend: Create conversation → Send message → Verify DB saves (15min)

### Frontend Tasks
- [ ] Add `sendMessage()` and `startConversation()` to `api.chat` in `api.ts` (15min)
- [ ] Update `api.chat.conversations.*` methods to support pagination params (10min)
- [ ] Update `chat.ts` state: Add conversations array, currentConversationId, pagination state (20min)
- [ ] Add `loadConversations()`, `loadMoreConversations()`, `selectConversation()` to chat store (25min)
- [ ] Update `ChatInterface.tsx`: Load conversations on mount, handle conversation selection (30min)
- [ ] Add infinite scroll for loading more conversations (scroll up detection) (25min)
- [ ] Load messages when conversation selected (call GetMessages API) (20min)
- [ ] Update `sendMessage()` to use chat endpoint and save to state (20min)
- [ ] Test end-to-end: Create conversation → Send message → Load history → Scroll for more (20min)

**Start Coding In**: ~4-5 hours total (Backend: ~2h, Frontend: ~2.5h)

---

**Total Planning Time**: ~30min | **Owner**: Development Team | 2025-01-27

<!-- Living Document - Update as you code -->

## 🔄 Implementation Tracking

**CRITICAL**: Follow the todo-list systematically. Mark items as complete, document blockers, update progress.

### Progress
- [ ] Backend: StartConversation DB save implemented
- [ ] Backend: SendMessage DB save implemented  
- [ ] Backend: GetConversations DB query implemented
- [ ] Backend: GetMessages DB query implemented
- [ ] Frontend: API client methods added
- [ ] Frontend: Chat state updated with conversations
- [ ] Frontend: Conversation list UI added
- [ ] Frontend: Infinite scroll implemented
- [ ] Frontend: Message loading implemented
- [ ] Frontend: ChatInterface updated to use chat endpoint
- [ ] End-to-end testing completed
- [ ] Error handling added

### Blockers
- None identified yet

### Notes
- Backend database tables exist - just need to implement handlers
- Kolosal API key already configured and working
- Follow `insights.go` pattern for DB queries (`h.db.Pool().Query()`)
- Frontend needs conversation list sidebar/component
- Infinite scroll: Detect scroll to top, load more conversations with offset
- Messages load oldest first (ASC) for proper chat display
- Conversations load newest first (DESC) for recent chats at top

**See**: [.sdd/IMPLEMENTATION_GUIDE.md](mdc:.sdd/IMPLEMENTATION_GUIDE.md) for detailed execution rules.
