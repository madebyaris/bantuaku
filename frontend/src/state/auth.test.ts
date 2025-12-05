import { renderHook, act } from '@testing-library/react'
import { useAuthStore } from './auth'

describe('Auth Store', () => {
  beforeEach(() => {
    useAuthStore.setState({
      token: null,
      userId: null,
      companyId: null,
      companyName: null,
      plan: null,
      isAuthenticated: false,
      login: useAuthStore.getState().login,
      logout: useAuthStore.getState().logout,
    })
  })

  it('initializes with empty state', () => {
    const { result } = renderHook(() => useAuthStore())
    expect(result.current.token).toBeNull()
    expect(result.current.userId).toBeNull()
    expect(result.current.companyId).toBeNull()
    expect(result.current.isAuthenticated).toBe(false)
  })

  it('sets auth data on login', () => {
    const { result } = renderHook(() => useAuthStore())
    act(() => {
      result.current.login({
        token: 't',
        user_id: 'u',
        company_id: 'c',
        company_name: 'Company',
        plan: 'free',
      })
    })

    expect(result.current.isAuthenticated).toBe(true)
    expect(result.current.token).toBe('t')
    expect(result.current.userId).toBe('u')
    expect(result.current.companyId).toBe('c')
    expect(result.current.companyName).toBe('Company')
    expect(result.current.plan).toBe('free')
  })

  it('clears auth data on logout', () => {
    const { result } = renderHook(() => useAuthStore())
    act(() => {
      result.current.login({
        token: 't',
        user_id: 'u',
        company_id: 'c',
        company_name: 'Company',
        plan: 'free',
      })
    })

    act(() => {
      result.current.logout()
    })

    expect(result.current.isAuthenticated).toBe(false)
    expect(result.current.token).toBeNull()
    expect(result.current.userId).toBeNull()
    expect(result.current.companyId).toBeNull()
    expect(result.current.companyName).toBeNull()
    expect(result.current.plan).toBeNull()
  })
})