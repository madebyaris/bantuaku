# Feature Brief: Floating Chat Widget

## Problem
Users currently have to navigate away from their current task (Dashboard, Forecast, etc.) to use the AI Assistant. This breaks workflow and context.

## Solution
Implement a persistent **Floating Chat Widget** that allows users to interact with the AI Assistant from any page.

## Core Requirements
1.  **Floating Action Button (FAB)**:
    -   Located at bottom-right of the screen.
    -   Visible on all pages *except* the dedicated AI Chat page.
2.  **Chat Overlay**:
    -   Clicking the FAB opens a chat interface in a popover/overlay.
    -   Must preserve conversation state (messages) between widget and full page.
3.  **Seamless Transition**:
    -   If the user navigates to `/ai-chat` (via sidebar or "expand" action), the widget should hide, and the full page should take over seamlessly.
4.  **Mobile Responsiveness**:
    -   Widget should work well on mobile, potentially taking up full screen when expanded.

## Implementation Approach
-   **Refactoring**: Move core chat logic from `AIChatPage` to a reusable `ChatInterface` component.
-   **Global Component**: Add `ChatWidget` to `Layout.tsx`.
-   **Conditional Rendering**: Use `useLocation` to hide widget on `/ai-chat`.
