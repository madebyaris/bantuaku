## Analysis
- **Goal**: Create a unified chat experience that exists as both a persistent floating widget (FAB) and a dedicated full page.
- **Key Interaction**:
  - **Floating Widget**: Visible on all pages (bottom-right). Clicking opens a chat overlay.
  - **Full Page**: Accessible via "AI Assistant" sidebar menu.
  - **Transition**: When navigating to the full page, the widget should conceptually "expand" or move to the center to become the page.
- **Technical Strategy**:
  - Extract core chat logic and UI from `AIChatPage.tsx` into a reusable `ChatInterface` component.
  - Implement a global `ChatWidget` in `Layout.tsx`.
  - Use conditional rendering/animations to handle the transition between "Widget Mode" and "Page Mode".
  - Ensure mobile responsiveness for both modes.

## Plan: Floating Chat Widget Brief

1.  **Create Brief File**: `specs/active/floating-chat-widget/feature-brief.md`
2.  **Brief Structure**:
    -   **Problem**: Need quick AI access without leaving current context.
    -   **Solution**: Hybrid Chat System (Widget + Page).
    -   **Core Requirements**:
        -   Floating Action Button (FAB) on bottom-right.
        -   Chat Overlay/Popover when FAB is clicked.
        -   Seamless state persistence (conversation continues from widget to page).
        -   "Morphing" transition effect (FAB hides when on `/ai-chat`, Page animates in).
        -   Mobile optimization (full-screen overlay on mobile).
    -   **Implementation Approach**:
        -   Refactor `AIChatPage` -> `ChatContainer` component.
        -   Update `Layout` to include `ChatWidget`.
        -   Add route-aware visibility logic (hide widget on `/ai-chat`).
    -   **Next Actions**: Refactor component, Create Widget, Integrate in Layout.

3.  **Research Focus (Quick Check)**:
    -   Check `AIChatPage.tsx` for ease of extraction.
    -   Check `Layout.tsx` structure for FAB placement.

**Why this approach?**
This plan separates the UI logic (Chat) from the Presentation logic (Page vs. Widget), allowing us to reuse the complex chat functionality (matrix effect, typing, history) in both contexts without code duplication.
