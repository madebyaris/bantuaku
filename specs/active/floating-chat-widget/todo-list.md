# Implementation Todo List: Floating Chat Widget

## Documentation
- [ ] Create feature brief and todo list (Done)
- [ ] Create progress tracking file

## Phase 1: Refactoring Core Chat Logic
- [ ] Extract `ChatInterface` component from `AIChatPage.tsx`
  - **Goal**: Make chat UI reusable in both widget and page.
  - **Files**: `frontend/src/components/chat/ChatInterface.tsx`, `frontend/src/pages/AIChatPage.tsx`
- [ ] Verify `AIChatPage` still works correctly after refactoring

## Phase 2: Build Chat Widget
- [ ] Create `ChatWidget` component
  - **Goal**: Implement FAB and Popover logic.
  - **Details**: Use `lucide-react` icons, `shadcn` styling.
  - **Files**: `frontend/src/components/chat/ChatWidget.tsx`

## Phase 3: Integration & Logic
- [ ] Integrate `ChatWidget` into `Layout.tsx`
  - **Goal**: Make it global.
  - **Files**: `frontend/src/components/layout/Layout.tsx`
- [ ] Implement "Hide on Chat Page" logic
  - **Goal**: Check `useLocation()` and hide FAB if path is `/ai-chat`.
- [ ] Implement "Expand" behavior (Optional enhancement)
  - **Goal**: Button to navigate from widget to full page.

## Phase 4: Polish & Mobile
- [ ] Optimize mobile view
  - **Goal**: Ensure widget doesn't overlap with mobile navigation or content.
- [ ] Verify animations and transitions
