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
    const error = await response.json().catch(() => ({ error: 'Request failed' }))
    throw new Error(error.error || 'Request failed')
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
        token: string
        user_id: string
        store_id: string
        store_name: string
        plan: string
      }>('/auth/register', {
        method: 'POST',
        body: { email, password, store_name: storeName, industry },
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
  stock: number
  created_at: string
  updated_at: string
}

export interface CreateProductRequest {
  product_name: string
  sku?: string
  category?: string
  unit_price: number
  cost?: number
  stock: number
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
  current_stock: number
  historical_sales: { date: string; quantity: number }[]
}

export interface Recommendation {
  product_id: string
  product_name: string
  current_stock: number
  recommended_qty: number
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
  total_products: number
  low_stock_count: number
  forecast_accuracy: number
  revenue_this_month: number
  revenue_trend: number
  top_selling_product: string
}
