import { create } from 'zustand'
import { persist } from 'zustand/middleware'

interface AuthState {
  token: string | null
  user: {
    id: string
    email: string
    role: string
  } | null
  isAuthenticated: boolean
  login: (token: string, user: { id: string; email: string; role: string }) => void
  logout: () => void
  decodeToken: (token: string) => { id: string; email: string; role: string } | null
}

// Simple JWT decode (without verification - backend handles that)
function decodeJWT(token: string): { user_id?: string; role?: string; [key: string]: unknown } | null {
  try {
    const base64Url = token.split('.')[1]
    const base64 = base64Url.replace(/-/g, '+').replace(/_/g, '/')
    const jsonPayload = decodeURIComponent(
      atob(base64)
        .split('')
        .map((c) => '%' + ('00' + c.charCodeAt(0).toString(16)).slice(-2))
        .join('')
    )
    return JSON.parse(jsonPayload)
  } catch {
    return null
  }
}

export const useAuthStore = create<AuthState>()(
  persist(
    (set, get) => ({
      token: null,
      user: null,
      isAuthenticated: false,
      login: (token, user) => {
        // Decode role from JWT if not provided
        const decoded = get().decodeToken(token)
        const role = decoded?.role || user.role || 'user'
        set({ 
          token, 
          user: { ...user, role }, 
          isAuthenticated: true 
        })
      },
      logout: () => {
        set({ token: null, user: null, isAuthenticated: false })
      },
      decodeToken: (token) => {
        const decoded = decodeJWT(token)
        if (!decoded || !decoded.user_id) return null
        return {
          id: decoded.user_id as string,
          email: '', // Email not in JWT, will be set from login response
          role: (decoded.role as string) || 'user',
        }
      },
    }),
    {
      name: 'admin-auth-storage',
    }
  )
)
