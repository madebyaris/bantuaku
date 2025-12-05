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
  
  if (response.status === 401 || response.status === 419) {
    // 401 = Unauthorized, 419 = Token Expired
    logout()
    window.location.href = '/login'
    throw new Error(response.status === 419 ? 'Token has expired' : 'Unauthorized')
  }
  
  if (!response.ok) {
    const error = await response.json().catch(() => ({ message: 'Request failed' }))
    // Backend returns { code, message, details } format
    throw new Error(error.message || error.error || 'Request failed')
  }
  
  return response.json()
}

// Auth
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
    register: (email: string, password: string, storeName: string, industry?: string) =>
      request<{
        message: string
        email: string
      }>('/auth/register', {
        method: 'POST',
        body: { email, password, store_name: storeName, industry },
      }),
    verifyEmail: (email: string, otp: string) =>
      request<{ message: string }>('/auth/verify-email', {
        method: 'POST',
        body: { email, otp },
      }),
    resendVerification: (email: string) =>
      request<{ message: string }>('/auth/resend-verification', {
        method: 'POST',
        body: { email },
      }),
    forgotPassword: (email: string) =>
      request<{ message: string }>('/auth/forgot-password', {
        method: 'POST',
        body: { email },
      }),
    resetPassword: (token: string, newPassword: string) =>
      request<{ message: string }>('/auth/reset-password', {
        method: 'POST',
        body: { token, new_password: newPassword },
      }),
  },
  
  products: {
    list: (category?: string) =>
      request<Product[]>(`/products${category ? `?category=${category}` : ''}`),
    get: (id: string) => request<Product>(`/products/${id}`),
    create: (data: CreateProductRequest) =>
      request<Product>('/products', { method: 'POST', body: data }),
    update: (id: string, data: Partial<CreateProductRequest>) =>
      request<Product>(`/products/${id}`, { method: 'PUT', body: data }),
    delete: (id: string) =>
      request<{ message: string }>(`/products/${id}`, { method: 'DELETE' }),
  },
  
  sales: {
    list: (productId?: string) =>
      request<Sale[]>(`/sales${productId ? `?product_id=${productId}` : ''}`),
    record: (data: RecordSaleRequest) =>
      request<Sale>('/sales/manual', { method: 'POST', body: data }),
    importCSV: async (file: File) => {
      const { token } = useAuthStore.getState()
      const formData = new FormData()
      formData.append('file', file)
      
      const response = await fetch(`${API_BASE}/sales/import-csv`, {
        method: 'POST',
        headers: {
          Authorization: `Bearer ${token}`,
        },
        body: formData,
      })
      
      if (!response.ok) {
        const error = await response.json().catch(() => ({ error: 'Import failed' }))
        throw new Error(error.error || 'Import failed')
      }
      
      return response.json() as Promise<ImportResult>
    },
  },
  
  integrations: {
    woocommerce: {
      connect: (storeUrl: string, consumerKey: string, consumerSecret: string) =>
        request<{ status: string; message: string }>('/integrations/woocommerce/connect', {
          method: 'POST',
          body: { store_url: storeUrl, consumer_key: consumerKey, consumer_secret: consumerSecret },
        }),
      status: () =>
        request<WooCommerceSyncStatus>('/integrations/woocommerce/sync-status'),
      syncNow: () =>
        request<{
          status: string
          products_synced: number
          orders_synced: number
          last_sync: string
        }>('/integrations/woocommerce/sync-now', { method: 'POST' }),
    },
  },
  
  forecasts: {
    get: (productId: string) => request<ForecastResponse>(`/forecasts/${productId}`),
  },
  
  recommendations: {
    list: () => request<Recommendation[]>('/recommendations'),
  },
  
  sentiment: {
    get: (productId: string) => request<SentimentData>(`/sentiment/${productId}`),
  },
  
  market: {
    trends: () => request<MarketTrend[]>('/market/trends'),
  },
  
  ai: {
    analyze: (question: string) =>
      request<AIAnalyzeResponse>('/ai/analyze', {
        method: 'POST',
        body: { question },
      }),
  },
  
  dashboard: {
    summary: () => request<DashboardSummary>('/dashboard/summary'),
  },
  
  chat: {
    startConversation: (purpose: string) =>
      request<{
        conversation_id: string
        title: string
        created_at: string
      }>('/chat/start', {
        method: 'POST',
        body: { purpose },
      }),
    sendMessage: (conversationId: string, message: string) =>
      request<{
        message_id: string
        assistant_reply: string
        structured_payload?: Record<string, unknown>
        citations?: Array<{ text: string; source: string }>
        rag_used: boolean
      }>('/chat/message', {
        method: 'POST',
        body: { conversation_id: conversationId, message },
      }),
    conversations: {
      list: (limit?: number, offset?: number) => {
        const params = new URLSearchParams()
        if (limit !== undefined) params.append('limit', limit.toString())
        if (offset !== undefined) params.append('offset', offset.toString())
        const query = params.toString()
        return request<{
          conversations: ConversationSummary[]
        }>(`/chat/conversations${query ? `?${query}` : ''}`).then(res => res.conversations || [])
      },
      get: (id: string) => request<Conversation>(`/chat/conversations/${id}`),
      messages: (conversationId: string, limit?: number, offset?: number) => {
        const params = new URLSearchParams()
        params.append('conversation_id', conversationId)
        if (limit !== undefined) params.append('limit', limit.toString())
        if (offset !== undefined) params.append('offset', offset.toString())
        return request<{ messages: Message[] }>(`/chat/messages?${params.toString()}`).then(res => res.messages || [])
      },
    },
  },
  
  insights: {
    list: (companyId?: string, type?: string) => {
      const params = new URLSearchParams()
      if (companyId) params.append('company_id', companyId)
      if (type) params.append('type', type)
      return request<Insight[]>(`/insights?${params.toString()}`)
    },
  },
  
  companies: {
    list: () => request<Company[]>('/companies'),
    get: (id: string) => request<CompanyProfile>(`/companies/${id}`),
  },
  
  files: {
    list: (companyId?: string) => {
      const params = new URLSearchParams()
      if (companyId) params.append('company_id', companyId)
      return request<FileUpload[]>(`/files?${params.toString()}`)
    },
    get: (id: string) => request<FileUpload>(`/files/${id}`),
  },

  regulations: {
    scrape: (maxPages?: number) => {
      const params = new URLSearchParams()
      if (maxPages) params.append('max_pages', maxPages.toString())
      return request<{
        message: string
        max_pages: number
        status: string
      }>(`/regulations/scrape?${params.toString()}`, { method: 'POST' })
    },
    status: () =>
      request<{
        total_regulations: number
        total_chunks: number
        last_scrape: string | null
      }>('/regulations/status'),
    list: (category?: string, limit?: number, offset?: number) => {
      const params = new URLSearchParams()
      if (category) params.append('category', category)
      if (limit) params.append('limit', limit.toString())
      if (offset) params.append('offset', offset.toString())
      return request<{
        regulations: Regulation[]
        count: number
        limit: number
        offset: number
      }>(`/regulations?${params.toString()}`)
    },
    search: (query: string, k?: number, filters?: {
      year?: number
      category?: string
      status?: string
    }) => {
      const params = new URLSearchParams()
      params.append('q', query)
      if (k) params.append('k', k.toString())
      if (filters?.year) params.append('year', filters.year.toString())
      if (filters?.category) params.append('category', filters.category)
      if (filters?.status) params.append('status', filters.status)
      return request<{
        query: string
        results: RegulationSearchResult[]
        count: number
      }>(`/regulations/search?${params.toString()}`)
    },
    indexChunks: (limit?: number) => {
      const params = new URLSearchParams()
      if (limit) params.append('limit', limit.toString())
      return request<{
        message: string
        limit: number
        status: string
      }>(`/embeddings/index?${params.toString()}`, { method: 'POST' })
    },
  },
}

// Types
export interface Product {
  id: string
  store_id: string
  product_name: string
  sku: string
  category: string
  unit_price: number
  cost: number
  created_at: string
  updated_at: string
}

export interface CreateProductRequest {
  product_name: string
  sku?: string
  category?: string
  unit_price: number
  cost?: number
}

export interface Sale {
  id: number
  store_id: string
  product_id: string
  quantity: number
  price: number
  sale_date: string
  source: string
  created_at: string
}

export interface RecordSaleRequest {
  product_id: string
  quantity: number
  price: number
  sale_date: string
}

export interface ImportResult {
  success_count: number
  errors: { row: number; error: string }[]
}

export interface WooCommerceSyncStatus {
  status: string
  last_sync?: string
  product_count: number
  order_count: number
  error_message?: string
}

export interface ForecastResponse {
  id: string
  product_id: string
  forecast_30d: number
  forecast_60d: number
  forecast_90d: number
  confidence: number
  algorithm: string
  generated_at: string
  expires_at: string
  product_name: string
  historical_sales: { date: string; quantity: number }[]
}

export interface Recommendation {
  product_id: string
  product_name: string
  projected_demand: number
  reason: string
  risk_level: string
}

export interface SentimentData {
  product_id: string
  sentiment_score: number
  positive_count: number
  negative_count: number
  neutral_count: number
  recent_mentions: {
    source: string
    text: string
    sentiment: number
    date: string
  }[]
}

export interface MarketTrend {
  name: string
  category: string
  trend_score: number
  growth_rate: number
  source: string
}

export interface AIAnalyzeResponse {
  answer: string
  confidence: number
  data_sources: string[]
}

export interface DashboardSummary {
  // Company Info
  company_name?: string
  company_industry?: string
  company_location?: string
  
  // Revenue Metrics
  revenue_this_month: number
  revenue_trend: number
  top_selling_product?: string
  
  // Activity Metrics
  total_conversations: number
  total_insights: number
  total_file_uploads: number
  
  // Insights Summary
  insights_summary: InsightsCounts
  
  // Recent Activity
  recent_conversations?: ConversationSummary[]
  recent_file_uploads?: FileUploadSummary[]
}

export interface InsightsCounts {
  forecast: number
  market: number
  marketing: number
  regulation: number
}

export interface ConversationSummary {
  id: string
  title: string
  purpose?: string
  created_at: string
  last_message_at: string
  last_message?: string
  updated_at: string
}

export interface FileUploadSummary {
  id: string
  original_filename: string
  source_type: string
  status: string
  created_at: string
}

export interface Conversation {
  id: string
  company_id: string
  user_id: string
  title?: string
  purpose: string
  created_at: string
  updated_at: string
}

export interface Message {
  id: string
  conversation_id: string
  sender: string
  content: string
  structured_payload?: Record<string, unknown>
  file_upload_id?: string
  created_at: string
}

export interface Insight {
  id: string
  company_id: string
  type: string
  input_context?: Record<string, unknown>
  result: Record<string, unknown>
  created_at: string
}

export interface Company {
  id: string
  owner_user_id: string
  name: string
  description?: string
  industry?: string
  business_model?: string
  founded_year?: number
  location_region?: string
  city?: string
  country: string
  website?: string
  social_media_handles?: Record<string, string>
  marketplaces?: Record<string, string>
  created_at: string
  updated_at: string
}

export interface CompanyProfile {
  company: Company
  products: Product[]
  data_sources: DataSource[]
  sales_data?: Sale[]
  last_updated: string
}

export interface DataSource {
  id: string
  company_id: string
  type: string
  provider?: string
  meta?: Record<string, unknown>
  status: string
  created_at: string
  updated_at: string
}

export interface FileUpload {
  id: string
  company_id: string
  user_id: string
  source_type: string
  original_filename: string
  storage_path: string
  mime_type?: string
  size_bytes: number
  status: string
  error_message?: string
  created_at: string
  processed_at?: string
}

export interface Regulation {
  id: string
  title: string
  regulation_number: string | null
  year: number | null
  category: string | null
  status: string
  source_url: string
  pdf_url: string | null
  published_date: string | null
  effective_date: string | null
  created_at: string
}

export interface RegulationSearchResult {
  chunk_id: string
  regulation_id: string
  chunk_text: string
  similarity: number
  regulation: {
    id: string
    title: string
    regulation_number: string | null
    year: number | null
    category: string | null
    pdf_url: string | null
  }
}
