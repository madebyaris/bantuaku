import { useAuthStore } from '@/state/auth'

const API_BASE = '/api/v1'

interface RequestOptions {
  method?: string
  body?: unknown
  headers?: Record<string, string>
}

async function request<T>(endpoint: string, options: RequestOptions = {}): Promise<T> {
  const { token, logout } = useAuthStore.getState()
  
  const headers: Record<string, string> = {
    'Content-Type': 'application/json',
    ...options.headers,
  }
  
  if (token) {
    headers['Authorization'] = `Bearer ${token}`
  }
  
  const response = await fetch(`${API_BASE}${endpoint}`, {
    method: options.method || 'GET',
    headers,
    body: options.body ? JSON.stringify(options.body) : undefined,
  })
  
  if (response.status === 401) {
    logout()
    window.location.href = '/login'
    throw new Error('Unauthorized')
  }
  
  if (!response.ok) {
    const error = await response.json().catch(() => ({ message: 'Request failed' }))
    throw new Error(error.message || error.error || 'Request failed')
  }
  
  return response.json()
}

export const api = {
  auth: {
    login: (email: string, password: string) =>
      request<{
        token: string
        user_id: string
        store_id: string
        store_name: string
        plan: string
      }>('/auth/login', {
        method: 'POST',
        body: { email, password },
      }),
  },
  admin: {
    users: {
      list: (page = 1, limit = 20) =>
        request<{
          users: Array<{
            id: string
            company_id?: string
            email: string
            role: string
            status: string
            store_name?: string
            industry?: string
            subscription_plan?: string
            subscription_status?: string
            created_at: string
            updated_at?: string
          }>
          pagination: {
            page: number
            limit: number
            total: number
          }
        }>(`/admin/users?page=${page}&limit=${limit}`),
      get: (id: string) =>
        request<{
          id: string
          company_id?: string
          email: string
          role: string
          status: string
          store_name?: string
          industry?: string
          subscription_plan?: string
          subscription_status?: string
          created_at: string
          updated_at?: string
        }>(`/admin/users/${id}`),
      create: (data: { email: string; password: string; role: string; store_name: string; industry: string; status?: 'active' | 'suspended' }) =>
        request<{ id: string; email: string; role: string; status?: string }>('/admin/users', {
          method: 'POST',
          body: data,
        }),
      update: (id: string, data: { email: string; store_name: string; industry?: string; role?: string; status?: 'active' | 'suspended' }) =>
        request<{
          id: string
          email: string
          role: string
          status: string
          store_name?: string
          industry?: string
          created_at?: string
          updated_at?: string
        }>(`/admin/users/${id}`, {
          method: 'PUT',
          body: data,
        }),
      updateRole: (id: string, role: string) =>
        request<{ message: string }>(`/admin/users/${id}/role`, {
          method: 'PUT',
          body: { role },
        }),
      updateStatus: (id: string, status: 'active' | 'suspended') =>
        request<{ message: string }>(`/admin/users/${id}/status`, {
          method: 'PUT',
          body: { status },
        }),
      upgradeSubscription: (id: string) =>
        request<{ message: string }>(`/admin/users/${id}/upgrade-subscription`, {
          method: 'PUT',
        }),
      delete: (id: string) =>
        request<{ message: string }>(`/admin/users/${id}`, {
          method: 'DELETE',
        }),
    },
    subscriptions: {
      list: (page = 1, limit = 20) =>
        request<{
          subscriptions: Array<{
            id: string
            company_id: string
            company_name: string
            plan_id: string
            plan_name: string
            status: string
            current_period_start: string
            current_period_end: string
            created_at: string
          }>
          pagination: {
            page: number
            limit: number
            total: number
          }
        }>(`/admin/subscriptions?page=${page}&limit=${limit}`),
      get: (id: string) =>
        request<{
          id: string
          company_id: string
          company_name: string
          plan_id: string
          plan_name: string
          status: string
          current_period_start: string
          current_period_end: string
        }>(`/admin/subscriptions/${id}`),
      create: (data: {
        company_id: string
        plan_id: string
        current_period_start?: string
        current_period_end?: string
      }) =>
        request<{ id: string; status: string }>('/admin/subscriptions', {
          method: 'POST',
          body: data,
        }),
      updateStatus: (id: string, status: string) =>
        request<{ message: string }>(`/admin/subscriptions/${id}/status`, {
          method: 'PUT',
          body: { status },
        }),
      updatePlan: (id: string, plan_id: string, status?: string) =>
        request<{ message: string }>(`/admin/subscriptions/${id}/plan`, {
          method: 'PUT',
          body: { plan_id, status },
        }),
    },
    auditLogs: {
      list: (page = 1, limit = 50, filters?: { action?: string; resource_type?: string; user_id?: string }) => {
        const params = new URLSearchParams({
          page: page.toString(),
          limit: limit.toString(),
          ...(filters?.action && { action: filters.action }),
          ...(filters?.resource_type && { resource_type: filters.resource_type }),
          ...(filters?.user_id && { user_id: filters.user_id }),
        })
        return request<{
          logs: Array<{
            id: number
            user_id?: string
            company_id?: string
            action: string
            resource_type?: string
            resource_id?: string
            ip_address?: string
            user_agent?: string
            metadata?: Record<string, unknown>
            created_at: string
          }>
          pagination: {
            page: number
            limit: number
            total: number
          }
        }>(`/admin/audit-logs?${params}`)
      },
    },
    stats: {
      get: () =>
        request<{
          total_users: number
          total_subscriptions: number
          active_subscriptions: number
          total_audit_logs: number
        }>('/admin/stats'),
    },
    activityAggregates: {
      list: (params: {
        page?: number
        limit?: number
        company_id?: string
        user_id?: string
        action_type?: string
        start_date?: string
        end_date?: string
      }) => {
        const search = new URLSearchParams()
        if (params.page) search.set('page', params.page.toString())
        if (params.limit) search.set('limit', params.limit.toString())
        if (params.company_id) search.set('company_id', params.company_id)
        if (params.user_id) search.set('user_id', params.user_id)
        if (params.action_type) search.set('action_type', params.action_type)
        if (params.start_date) search.set('start_date', params.start_date)
        if (params.end_date) search.set('end_date', params.end_date)
        return request<{
          data: Array<{
            date: string
            user_id?: string
            company_id?: string
            action_type: string
            count: number
          }>
          pagination: { page: number; limit: number; total: number }
        }>(`/admin/activity-aggregates?${search.toString()}`)
      },
    },
    tokenUsageAggregates: {
      list: (params: {
        page?: number
        limit?: number
        company_id?: string
        user_id?: string
        model?: string
        provider?: string
        start_date?: string
        end_date?: string
      }) => {
        const search = new URLSearchParams()
        if (params.page) search.set('page', params.page.toString())
        if (params.limit) search.set('limit', params.limit.toString())
        if (params.company_id) search.set('company_id', params.company_id)
        if (params.user_id) search.set('user_id', params.user_id)
        if (params.model) search.set('model', params.model)
        if (params.provider) search.set('provider', params.provider)
        if (params.start_date) search.set('start_date', params.start_date)
        if (params.end_date) search.set('end_date', params.end_date)
        return request<{
          data: Array<{
            date: string
            user_id?: string
            company_id?: string
            model: string
            provider: string
            prompt_tokens: number
            completion_tokens: number
            total_tokens: number
          }>
          pagination: { page: number; limit: number; total: number }
        }>(`/admin/token-usage-aggregates?${search.toString()}`)
      },
    },
  },
}

