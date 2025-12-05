import { create } from 'zustand'
import { persist } from 'zustand/middleware'

interface AuthState {
  token: string | null
  userId: string | null
  companyId: string | null
  companyName: string | null
  plan: string | null
  isAuthenticated: boolean
  login: (data: {
    token: string
    user_id: string
    company_id: string
    company_name: string
    plan: string
  }) => void
  logout: () => void
}

export const useAuthStore = create<AuthState>()(
  persist(
    (set) => ({
      token: null,
      userId: null,
      companyId: null,
      companyName: null,
      plan: null,
      isAuthenticated: false,
      login: (data) =>
        set({
          token: data.token,
          userId: data.user_id,
          companyId: data.company_id,
          companyName: data.company_name,
          plan: data.plan,
          isAuthenticated: true,
        }),
      logout: () =>
        set({
          token: null,
          userId: null,
          companyId: null,
          companyName: null,
          plan: null,
          isAuthenticated: false,
        }),
    }),
    {
      name: 'bantuaku-auth',
    }
  )
)
