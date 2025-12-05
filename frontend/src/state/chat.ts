import { create } from 'zustand'
import { api, ConversationSummary } from '@/lib/api'

export interface ChatMessage {
  id: string
  role: 'user' | 'assistant'
  text: string
  confidence?: number
  dataSources?: string[]
  timestamp: Date
}

interface ChatState {
  // Messages for current conversation
  messages: ChatMessage[]
  loading: boolean
  
  // Conversations list
  conversations: ConversationSummary[]
  currentConversationId: string | null
  hasMoreConversations: boolean
  conversationsOffset: number
  loadingConversations: boolean
  
  // Actions
  addMessage: (message: ChatMessage) => void
  setLoading: (loading: boolean) => void
  clearMessages: () => void
  
  // Conversation management
  loadConversations: () => Promise<void>
  loadMoreConversations: () => Promise<void>
  selectConversation: (conversationId: string) => Promise<void>
  loadMessages: (conversationId: string) => Promise<void>
  
  // Create new conversation
  createConversation: (purpose: string) => Promise<string>
  
  // Initialize single continuous conversation
  initializeConversation: () => Promise<void>
}

export const useChatStore = create<ChatState>((set, get) => ({
  messages: [],
  loading: false,
  conversations: [],
  currentConversationId: null,
  hasMoreConversations: false,
  conversationsOffset: 0,
  loadingConversations: false,
  
  addMessage: (message) => set((state) => ({ messages: [...state.messages, message] })),
  setLoading: (loading) => set({ loading }),
  clearMessages: () => set({ messages: [] }),
  
  loadConversations: async () => {
    set({ loadingConversations: true })
    try {
      const conversations = await api.chat.conversations.list(5, 0)
      set({
        conversations,
        conversationsOffset: conversations.length,
        hasMoreConversations: conversations.length === 5,
        loadingConversations: false,
      })
    } catch (error) {
      console.error('Failed to load conversations:', error)
      set({ loadingConversations: false })
    }
  },
  
  loadMoreConversations: async () => {
    const { conversationsOffset, hasMoreConversations } = get()
    if (!hasMoreConversations) return
    
    set({ loadingConversations: true })
    try {
      const moreConversations = await api.chat.conversations.list(5, conversationsOffset)
      set((state) => ({
        conversations: [...state.conversations, ...moreConversations],
        conversationsOffset: state.conversationsOffset + moreConversations.length,
        hasMoreConversations: moreConversations.length === 5,
        loadingConversations: false,
      }))
    } catch (error) {
      console.error('Failed to load more conversations:', error)
      set({ loadingConversations: false })
    }
  },
  
  selectConversation: async (conversationId: string) => {
    set({ currentConversationId: conversationId })
    await get().loadMessages(conversationId)
  },
  
  loadMessages: async (conversationId: string) => {
    try {
      const messages = await api.chat.conversations.messages(conversationId, 50, 0)
      // Ensure messages is an array
      if (!Array.isArray(messages)) {
        console.error('Messages is not an array:', messages)
        set({ messages: [] })
        return
      }
      // Convert Message[] to ChatMessage[]
      const chatMessages: ChatMessage[] = messages.map((msg) => ({
        id: msg.id,
        role: msg.sender === 'user' ? 'user' : 'assistant',
        text: msg.content,
        timestamp: new Date(msg.created_at),
      }))
      set({ messages: chatMessages })
    } catch (error) {
      console.error('Failed to load messages:', error)
      set({ messages: [] })
    }
  },
  
  createConversation: async (purpose: string) => {
    try {
      const response = await api.chat.startConversation(purpose)
      // Reload conversations to include the new one
      await get().loadConversations()
      // Select the new conversation
      await get().selectConversation(response.conversation_id)
      return response.conversation_id
    } catch (error) {
      console.error('Failed to create conversation:', error)
      throw error
    }
  },
  
  initializeConversation: async () => {
    try {
      // Try to get the most recent conversation
      const conversations = await api.chat.conversations.list(1, 0)
      
      if (conversations.length > 0) {
        // Use existing conversation
        const conversationId = conversations[0].id
        set({ currentConversationId: conversationId })
        await get().loadMessages(conversationId)
      } else {
        // Create new conversation
        const response = await api.chat.startConversation('analysis')
        set({ currentConversationId: response.conversation_id })
        // Messages will be empty for new conversation
        set({ messages: [] })
      }
    } catch (error) {
      console.error('Failed to initialize conversation:', error)
      // Try to create a new one as fallback
      try {
        const response = await api.chat.startConversation('analysis')
        set({ currentConversationId: response.conversation_id, messages: [] })
      } catch (createError) {
        console.error('Failed to create fallback conversation:', createError)
        throw createError
      }
    }
  },
}))
