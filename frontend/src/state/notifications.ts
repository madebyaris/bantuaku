import { create } from 'zustand'
import { api, Notification } from '@/lib/api'

interface NotificationState {
  items: Notification[]
  loading: boolean
  load: (status?: string) => Promise<void>
  markRead: (id: string) => Promise<void>
  remove: (id: string) => Promise<void>
}

export const useNotificationStore = create<NotificationState>((set, get) => ({
  items: [],
  loading: false,
  load: async (status) => {
    set({ loading: true })
    try {
      const res = await api.notifications.list(status)
      set({ items: res.notifications || [] })
    } catch (err) {
      // Keep silent for UI; optionally add toast in caller
    } finally {
      set({ loading: false })
    }
  },
  markRead: async (id: string) => {
    try {
      await api.notifications.markRead(id)
      const items = get().items.map((n) =>
        n.id === id ? { ...n, status: 'read', read_at: new Date().toISOString() } : n
      )
      set({ items })
    } catch (err) {
      // Ignore
    }
  },
  remove: async (id: string) => {
    try {
      await api.notifications.delete(id)
      set({ items: get().items.filter((n) => n.id !== id) })
    } catch (err) {
      // Ignore
    }
  },
}))
