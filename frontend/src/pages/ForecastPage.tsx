import { useEffect, useMemo, useState } from 'react'
import { TrendingUp, Loader2 } from 'lucide-react'
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { api, Product, MonthlyForecast, MonthlyStrategy } from '@/lib/api'

export function ForecastPage() {
  const [products, setProducts] = useState<Product[]>([])
  const [selectedProductId, setSelectedProductId] = useState<string | null>(null)
  const [forecasts, setForecasts] = useState<MonthlyForecast[]>([])
  const [strategies, setStrategies] = useState<MonthlyStrategy[]>([])
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    loadProducts()
  }, [])

  useEffect(() => {
    if (selectedProductId) {
      loadForecasts(selectedProductId)
    }
  }, [selectedProductId])

  async function loadProducts() {
    setLoading(true)
    setError(null)
    try {
      const list = await api.products.list()
      setProducts(list)
      if (list.length > 0) {
        setSelectedProductId(list[0].id)
      }
    } catch (err) {
      setError('Gagal memuat produk')
    } finally {
      setLoading(false)
    }
  }

  async function loadForecasts(productId: string) {
    setLoading(true)
    setError(null)
    try {
      const [forecastResp, strategyResp] = await Promise.all([
        api.forecasts.monthly(productId),
        api.strategies.monthly(productId),
      ])
      setForecasts(forecastResp.forecasts || [])
      setStrategies(strategyResp.strategies || [])
    } catch (err) {
      setError('Gagal memuat forecast')
      setForecasts([])
      setStrategies([])
    } finally {
      setLoading(false)
    }
  }

  const selectedProduct = useMemo(
    () => products.find((p) => p.id === selectedProductId) || null,
    [products, selectedProductId]
  )

  if (loading && products.length === 0) {
    return (
      <div className="flex items-center justify-center h-[calc(100vh-10rem)]">
        <Loader2 className="w-8 h-8 animate-spin text-emerald-500" />
      </div>
    )
  }

  return (
    <div className="space-y-6 animate-fade-in-up max-w-6xl mx-auto pb-12">
      <div className="flex items-center justify-between gap-4">
        <div>
          <div className="flex items-center gap-2 mb-2">
            <div className="p-2 bg-emerald-500/10 rounded-lg border border-emerald-500/20">
              <TrendingUp className="w-5 h-5 text-emerald-400" />
            </div>
            <span className="text-sm font-medium text-emerald-400 uppercase tracking-wider">
              Sales Intelligence
            </span>
          </div>
          <h1 className="text-3xl font-display font-bold text-slate-100">Sales Forecast</h1>
          <p className="text-slate-400 mt-2 max-w-2xl text-sm">
            Gunakan data historis untuk melihat proyeksi bulanan dan strategi otomatis.
          </p>
        </div>

        {products.length > 0 && (
          <div className="w-64">
            <Select
              value={selectedProductId ?? undefined}
              onValueChange={(value) => setSelectedProductId(value)}
            >
              <SelectTrigger>
                <SelectValue placeholder="Pilih produk" />
              </SelectTrigger>
              <SelectContent>
                {products.map((p) => (
                  <SelectItem key={p.id} value={p.id}>
                    {p.name}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>
        )}
      </div>

      {error && (
        <Card className="border-red-500/20 bg-red-500/5">
          <CardContent className="py-4 text-sm text-red-200">{error}</CardContent>
        </Card>
      )}

      {!selectedProduct && (
        <Card>
          <CardContent className="py-12 text-center text-slate-400">
            Tidak ada produk. Tambahkan produk terlebih dahulu.
          </CardContent>
        </Card>
      )}

      {selectedProduct && (
        <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
          <Card className="lg:col-span-2 border-white/10 bg-white/5">
            <CardHeader className="border-b border-white/5">
              <CardTitle className="text-lg text-slate-100">Forecast Bulanan</CardTitle>
              <CardDescription className="text-slate-400">
                Produk: {selectedProduct.name}
              </CardDescription>
            </CardHeader>
            <CardContent className="pt-4 space-y-3">
              {loading ? (
                <div className="flex justify-center py-8">
                  <Loader2 className="w-6 h-6 animate-spin text-emerald-500" />
                </div>
              ) : forecasts.length === 0 ? (
                <p className="text-sm text-slate-400">Belum ada forecast. Coba generate dari backend.</p>
              ) : (
                forecasts.map((f) => (
                  <div
                    key={f.id}
                    className="flex items-center justify-between p-3 rounded-lg bg-white/5 border border-white/10"
                  >
                    <div>
                      <p className="text-sm text-slate-400">Bulan {f.month}</p>
                      <p className="text-lg font-semibold text-slate-100">
                        {f.predicted_quantity} unit
                      </p>
                      <p className="text-xs text-slate-500">
                        Confidence {Math.round(f.confidence_score * 100)}% • Algoritma {f.algorithm}
                      </p>
                    </div>
                    <div className="text-right text-xs text-slate-500">
                      Rentang: {f.confidence_lower} - {f.confidence_upper}
                      <br />
                      {new Date(f.forecast_date).toLocaleDateString('id-ID')}
                    </div>
                  </div>
                ))
              )}
            </CardContent>
          </Card>

          <Card className="border-white/10 bg-white/5">
            <CardHeader className="border-b border-white/5">
              <CardTitle className="text-lg text-slate-100">Strategi</CardTitle>
              <CardDescription className="text-slate-400">
                Rekomendasi otomatis dari forecast
              </CardDescription>
            </CardHeader>
            <CardContent className="pt-4 space-y-3">
              {loading ? (
                <div className="flex justify-center py-8">
                  <Loader2 className="w-6 h-6 animate-spin text-emerald-500" />
                </div>
              ) : strategies.length === 0 ? (
                <p className="text-sm text-slate-400">
                  Belum ada strategi. Generate forecast untuk melihat rekomendasi.
                </p>
              ) : (
                strategies.map((s) => (
                  <div
                    key={s.id}
                    className="p-3 rounded-lg bg-white/5 border border-white/10 space-y-1"
                  >
                    <p className="text-xs text-slate-500">Bulan {s.month}</p>
                    <p className="text-sm font-semibold text-slate-100">{s.strategy_text}</p>
                    {s.estimated_impact && (
                      <p className="text-xs text-emerald-400">
                        {JSON.stringify(s.estimated_impact)}
                      </p>
                    )}
                  </div>
                ))
              )}
            </CardContent>
          </Card>
        </div>
      )}
    </div>
  )
}
