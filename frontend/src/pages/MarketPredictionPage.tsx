import { useEffect, useMemo, useState } from 'react'
import { Globe, MapPin, TrendingUp, Zap, Loader2 } from 'lucide-react'
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { api, TrendKeyword, TrendPoint } from '@/lib/api'

export function MarketPredictionPage() {
  const [keywords, setKeywords] = useState<TrendKeyword[]>([])
  const [selectedKeywordId, setSelectedKeywordId] = useState<string | null>(null)
  const [series, setSeries] = useState<TrendPoint[]>([])
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    loadKeywords()
  }, [])

  useEffect(() => {
    if (selectedKeywordId) {
      loadSeries(selectedKeywordId)
    }
  }, [selectedKeywordId])

  async function loadKeywords() {
    setLoading(true)
    setError(null)
    try {
      const resp = await api.trends.keywords()
      setKeywords(resp.keywords || [])
      if (resp.keywords?.length) {
        setSelectedKeywordId(resp.keywords[0].id)
      }
    } catch (err) {
      setError('Gagal memuat keyword tren')
    } finally {
      setLoading(false)
    }
  }

  async function loadSeries(keywordId: string) {
    setLoading(true)
    setError(null)
    try {
      const resp = await api.trends.series(keywordId)
      setSeries(resp.time_series || [])
    } catch (err) {
      setError('Gagal memuat data tren')
      setSeries([])
    } finally {
      setLoading(false)
    }
  }

  const selectedKeyword = useMemo(
    () => keywords.find((k) => k.id === selectedKeywordId) || null,
    [keywords, selectedKeywordId]
  )

  return (
    <div className="max-w-6xl mx-auto px-4 py-8 space-y-6 animate-fade-in-up">
      <div className="flex items-center justify-between gap-4 border-b border-white/10 pb-4">
        <div>
          <div className="flex items-center gap-2 mb-2">
            <div className="p-2 bg-blue-500/10 rounded-lg border border-blue-500/20">
              <Globe className="w-5 h-5 text-blue-400" />
            </div>
            <span className="text-sm font-medium text-blue-400 uppercase tracking-wider">
              Market Intelligence
            </span>
          </div>
          <h1 className="text-3xl md:text-4xl font-display font-bold text-slate-100">
            Prediksi Pasar
          </h1>
          <p className="text-slate-400 mt-2 max-w-2xl text-sm">
            Pantau keyword yang sedang di-track dan lihat time-series trennya.
          </p>
        </div>
        <div className="hidden md:flex items-center gap-3 text-slate-400">
          <MapPin className="w-4 h-4" />
          <span className="text-sm">Fokus pada keyword per company</span>
        </div>
      </div>

      {error && (
        <Card className="border-red-500/20 bg-red-500/5">
          <CardContent className="py-3 text-sm text-red-200">{error}</CardContent>
        </Card>
      )}

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        <Card className="lg:col-span-2 border-white/10 bg-white/5">
          <CardHeader className="border-b border-white/5">
            <CardTitle className="text-lg text-slate-100">Time Series Tren</CardTitle>
            <CardDescription className="text-slate-400">
              {selectedKeyword ? selectedKeyword.keyword : 'Pilih keyword'}
            </CardDescription>
          </CardHeader>
          <CardContent className="pt-4">
            {loading ? (
              <div className="flex justify-center py-12">
                <Loader2 className="w-6 h-6 animate-spin text-emerald-500" />
              </div>
            ) : series.length === 0 ? (
              <p className="text-sm text-slate-400">
                Belum ada data tren. Tambah keyword atau jalankan ingestion.
              </p>
            ) : (
              <div className="space-y-3">
                {series.map((point) => (
                  <div
                    key={`${point.timestamp}-${point.score}`}
                    className="flex items-center justify-between p-3 bg-white/5 border border-white/10 rounded-lg"
                  >
                    <div className="text-sm text-slate-200">
                      {new Date(point.timestamp).toLocaleDateString('id-ID')}
                    </div>
                    <div className="text-sm font-semibold text-emerald-400 flex items-center gap-2">
                      <TrendingUp className="w-4 h-4" />
                      {point.score}
                    </div>
                  </div>
                ))}
              </div>
            )}
          </CardContent>
        </Card>

        <Card className="border-white/10 bg-white/5">
          <CardHeader className="border-b border-white/5">
            <CardTitle className="text-lg text-slate-100 flex items-center gap-2">
              <Zap className="w-5 h-5 text-yellow-400" />
              Keyword Aktif
            </CardTitle>
            <CardDescription className="text-slate-400">
              Pilih keyword untuk melihat trennya
            </CardDescription>
          </CardHeader>
          <CardContent className="pt-4 space-y-4">
            {keywords.length === 0 ? (
              <p className="text-sm text-slate-400">Belum ada keyword yang di-track.</p>
            ) : (
              <>
                <Select
                  value={selectedKeywordId ?? undefined}
                  onValueChange={(v) => setSelectedKeywordId(v)}
                >
                  <SelectTrigger>
                    <SelectValue placeholder="Pilih keyword" />
                  </SelectTrigger>
                  <SelectContent>
                    {keywords.map((k) => (
                      <SelectItem key={k.id} value={k.id}>
                        {k.keyword} ({k.geo})
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>

                <div className="space-y-2">
                  {keywords.map((k) => (
                    <div
                      key={k.id}
                      className={`p-3 rounded-lg border ${
                        k.id === selectedKeywordId
                          ? 'border-emerald-500/40 bg-emerald-500/5'
                          : 'border-white/10 bg-white/5'
                      }`}
                      onClick={() => setSelectedKeywordId(k.id)}
                    >
                      <div className="flex items-center justify-between">
                        <p className="font-semibold text-slate-100">{k.keyword}</p>
                        <span className="text-xs text-slate-500">{k.geo}</span>
                      </div>
                      {k.category && (
                        <p className="text-xs text-slate-500 mt-1">Kategori: {k.category}</p>
                      )}
                    </div>
                  ))}
                </div>
              </>
            )}
          </CardContent>
        </Card>
      </div>
    </div>
  )
}
