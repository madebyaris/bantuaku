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
  
  // Only redirect on 401 if we have a token (authenticated request failed)
  // Don't redirect for login endpoint (it's expected to return 401 on failure)
  if (response.status === 401 && token && !endpoint.includes('/auth/login')) {
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
            email: string
            role: string
            status: string
            company_id?: string
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
      create: (data: { email: string; password: string; role: string; store_name: string; industry: string }) =>
        request<{ id: string; email: string; role: string }>('/admin/users', {
          method: 'POST',
          body: data,
        }),
      update: (id: string, data: { email: string; store_name: string; industry?: string }) =>
        request<{ message: string }>(`/admin/users/${id}`, {
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
      stats: () =>
        request<{
          total_subscriptions: number
          active_subscriptions: number
          trialing_count: number
          canceled_count: number
          past_due_count: number
          mrr: number
          plan_breakdown: Array<{
            plan_id: string
            plan_name: string
            count: number
            price_monthly: number
          }>
        }>('/admin/subscriptions/stats'),
      list: (page = 1, limit = 20) =>
        request<{
          subscriptions: Array<{
            id: string
            company_id: string
            company_name: string
            owner_email: string
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
      getTransactions: (id: string, page = 1, limit = 20) =>
        request<{
          transactions: Array<{
            id: string
            subscription_id: string
            company_id: string
            event_type: string
            old_plan_id?: string
            new_plan_id?: string
            old_status?: string
            new_status?: string
            changed_by_user_id?: string
            metadata?: Record<string, unknown>
            created_at: string
          }>
          pagination: {
            page: number
            limit: number
            total: number
          }
        }>(`/admin/subscriptions/${id}/transactions?page=${page}&limit=${limit}`),
    },
    chatUsage: {
      get: (filters?: {
        company_id?: string
        start_date?: string
        end_date?: string
        page?: number
        limit?: number
      }) => {
        const params = new URLSearchParams()
        if (filters?.company_id) params.append('company_id', filters.company_id)
        if (filters?.start_date) params.append('start_date', filters.start_date)
        if (filters?.end_date) params.append('end_date', filters.end_date)
        if (filters?.page) params.append('page', filters.page.toString())
        if (filters?.limit) params.append('limit', filters.limit.toString())
        return request<{
          stats: {
            total_messages: number
            total_conversations: number
            unique_users: number
            period: string
            start_date?: string
            end_date?: string
          }
          daily_logs: Array<{
            id: string
            company_id: string
            date: string
            total_messages: number
            total_conversations: number
            unique_users: number
            created_at: string
            updated_at: string
          }>
          pagination: {
            page: number
            limit: number
            total: number
          }
        }>(`/admin/chat-usage?${params.toString()}`)
      },
    },
    tokenUsage: {
      get: (filters?: {
        company_id?: string
        model?: string
        start_date?: string
        end_date?: string
        page?: number
        limit?: number
      }) => {
        const params = new URLSearchParams()
        if (filters?.company_id) params.append('company_id', filters.company_id)
        if (filters?.model) params.append('model', filters.model)
        if (filters?.start_date) params.append('start_date', filters.start_date)
        if (filters?.end_date) params.append('end_date', filters.end_date)
        if (filters?.page) params.append('page', filters.page.toString())
        if (filters?.limit) params.append('limit', filters.limit.toString())
        return request<{
          stats: {
            total_prompt_tokens: number
            total_completion_tokens: number
            total_tokens: number
            estimated_cost: number
            model_breakdown: Array<{
              model: string
              provider: string
              prompt_tokens: number
              completion_tokens: number
              total_tokens: number
              estimated_cost: number
              request_count: number
            }>
            start_date?: string
            end_date?: string
          }
          usage: Array<{
            id: string
            company_id: string
            conversation_id?: string
            message_id?: string
            model: string
            provider: string
            prompt_tokens: number
            completion_tokens: number
            total_tokens: number
            created_at: string
          }>
          pagination: {
            page: number
            limit: number
            total: number
          }
        }>(`/admin/token-usage?${params.toString()}`)
      },
    },
    plans: {
      list: (page = 1, limit = 20) =>
        request<{
          plans: Array<{
            id: string
            name: string
            display_name: string
            price_monthly: number
            price_yearly?: number
            currency: string
            max_stores?: number
            max_products?: number
            max_chats_per_month?: number
            max_file_uploads_per_month?: number
            max_file_size_mb?: number
            max_forecast_refreshes_per_month?: number
            features: Record<string, boolean>
            is_active: boolean
            created_at: string
            updated_at?: string
          }>
          pagination: {
            page: number
            limit: number
            total: number
          }
        }>(`/admin/plans?page=${page}&limit=${limit}`),
      get: (id: string) =>
        request<{
          id: string
          name: string
          display_name: string
          price_monthly: number
          price_yearly?: number
          currency: string
          max_stores?: number
          max_products?: number
          max_chats_per_month?: number
          max_file_uploads_per_month?: number
          max_file_size_mb?: number
          max_forecast_refreshes_per_month?: number
          features: Record<string, boolean>
          is_active: boolean
          created_at: string
          updated_at?: string
        }>(`/admin/plans/${id}`),
      create: (data: {
        name: string
        display_name: string
        price_monthly: number
        price_yearly?: number
        currency?: string
        max_stores?: number
        max_products?: number
        max_chats_per_month?: number
        max_file_uploads_per_month?: number
        max_file_size_mb?: number
        max_forecast_refreshes_per_month?: number
        features?: Record<string, boolean>
      }) =>
        request<{ id: string; name: string }>('/admin/plans', {
          method: 'POST',
          body: data,
        }),
      update: (id: string, data: {
        display_name: string
        price_monthly: number
        price_yearly?: number
        currency?: string
        max_stores?: number
        max_products?: number
        max_chats_per_month?: number
        max_file_uploads_per_month?: number
        max_file_size_mb?: number
        max_forecast_refreshes_per_month?: number
        features?: Record<string, boolean>
        is_active?: boolean
      }) =>
        request<{ message: string }>(`/admin/plans/${id}`, {
          method: 'PUT',
          body: data,
        }),
      delete: (id: string) =>
        request<{ message: string }>(`/admin/plans/${id}`, {
          method: 'DELETE',
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
    settings: {
      getAIProvider: () =>
        request<{ provider: string }>('/admin/settings/ai-provider'),
      updateAIProvider: (provider: 'openrouter' | 'kolosal') =>
        request<{ provider: string; message: string }>('/admin/settings/ai-provider', {
          method: 'PUT',
          body: { provider },
        }),
    },
  },
}

