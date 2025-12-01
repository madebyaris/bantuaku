import { create } from 'zustand'
import { persist } from 'zustand/middleware'

interface AuthState {
  token: string | null
  userId: string | null
  storeId: string | null
  storeName: string | null
  plan: string | null
  isAuthenticated: boolean
  login: (data: {
    token: string
    user_id: string
    store_id: string
    store_name: string
    plan: string
  }) => void
  logout: () => void
}

export const useAuthStore = create<AuthState>()(
  persist(
    (set) => ({
      token: null,
      userId: null,
      storeId: null,
      storeName: null,
      plan: null,
      isAuthenticated: false,
      login: (data) =>
        set({
          token: data.token,
          userId: data.user_id,
          storeId: data.store_id,
          storeName: data.store_name,
          plan: data.plan,
          isAuthenticated: true,
        }),
      logout: () =>
        set({
          token: null,
          userId: null,
          storeId: null,
          storeName: null,
          plan: null,
          isAuthenticated: false,
        }),
    }),
    {
      name: 'bantuaku-auth',
    }
  )
)
