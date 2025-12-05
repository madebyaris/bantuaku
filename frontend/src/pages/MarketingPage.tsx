import { useEffect, useState } from 'react'
import { Megaphone, Sparkles, Loader2 } from 'lucide-react'
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { api, Insight } from '@/lib/api'

export function MarketingPage() {
  const [insights, setInsights] = useState<Insight[]>([])
  const [latest, setLatest] = useState<Insight | null>(null)
  const [loading, setLoading] = useState(false)
  const [generating, setGenerating] = useState(false)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    loadInsights()
  }, [])

  async function loadInsights() {
    setLoading(true)
    setError(null)
    try {
      const list = await api.insights.list(undefined, 'marketing_recommendation')
      setInsights(list)
      setLatest(list[0] || null)
    } catch (err) {
      setError('Gagal memuat insights marketing')
    } finally {
      setLoading(false)
    }
  }

  async function handleGenerate() {
    setGenerating(true)
    setError(null)
    try {
      const created = await api.insights.generateMarketing()
      const mapped: Insight = {
        id: created.insight_id,
        company_id: '',
        type: created.type,
        input_context: {},
        result: created.result,
        created_at: created.created_at,
      }
      setInsights((prev) => [mapped, ...prev])
      setLatest(mapped)
    } catch (err) {
      setError('Gagal menghasilkan rekomendasi')
    } finally {
      setGenerating(false)
    }
  }

  return (
    <div className="max-w-6xl mx-auto px-4 py-8 animate-fade-in-up space-y-6 pb-20">
      <div className="flex flex-col md:flex-row md:items-end justify-between gap-4 border-b border-white/10 pb-6">
        <div>
          <div className="flex items-center gap-2 mb-2">
            <div className="p-2 bg-purple-500/10 rounded-lg border border-purple-500/20">
              <Megaphone className="w-5 h-5 text-purple-400" />
            </div>
            <span className="text-sm font-medium text-purple-400 uppercase tracking-wider">
              Marketing Strategy
            </span>
          </div>
          <h1 className="text-3xl md:text-4xl font-display font-bold text-slate-100">
            Rekomendasi Marketing
          </h1>
          <p className="text-slate-400 mt-2 max-w-2xl text-lg">
            Generate rekomendasi otomatis untuk kampanye Anda.
          </p>
        </div>
        <Button
          onClick={handleGenerate}
          disabled={generating}
          className="bg-purple-600 hover:bg-purple-500 text-white border-0"
        >
          {generating ? <Loader2 className="w-4 h-4 animate-spin mr-2" /> : <Sparkles className="w-4 h-4 mr-2" />}
          Generate Rekomendasi
        </Button>
      </div>

      {error && (
        <Card className="border-red-500/20 bg-red-500/5">
          <CardContent className="py-3 text-sm text-red-200">{error}</CardContent>
        </Card>
      )}

      {loading ? (
        <div className="flex justify-center py-12">
          <Loader2 className="w-8 h-8 animate-spin text-purple-400" />
        </div>
      ) : latest ? (
        <Card className="bg-gradient-to-br from-purple-900/20 to-black border-purple-500/20">
          <CardHeader>
            <CardTitle className="text-lg text-slate-100">Rekomendasi Terbaru</CardTitle>
            <CardDescription className="text-slate-400">
              Dibuat pada {new Date(latest.created_at).toLocaleString('id-ID')}
            </CardDescription>
          </CardHeader>
          <CardContent className="space-y-3">
            {Array.isArray(latest.result?.recommendations) && latest.result.recommendations.length > 0 ? (
              (latest.result.recommendations as Array<Record<string, unknown>>).map((rec, idx) => (
                <div key={idx} className="p-3 rounded-lg border border-white/10 bg-white/5">
                  <p className="text-sm text-slate-200 font-semibold">
                    {(rec.title as string) || `Rekomendasi ${idx + 1}`}
                  </p>
                  {rec.description && (
                    <p className="text-sm text-slate-400 mt-1">{rec.description as string}</p>
                  )}
                </div>
              ))
            ) : (
              <p className="text-sm text-slate-300">
                {latest.result?.message ||
                  'Belum ada rekomendasi. Gunakan tombol Generate untuk membuat rekomendasi baru.'}
              </p>
            )}
          </CardContent>
        </Card>
      ) : (
        <Card>
          <CardContent className="py-12 text-center text-slate-400">
            Belum ada rekomendasi. Klik “Generate Rekomendasi” untuk memulai.
          </CardContent>
        </Card>
      )}

      {insights.length > 1 && (
        <Card className="border-white/10 bg-white/5">
          <CardHeader>
            <CardTitle className="text-lg text-slate-100">Riwayat</CardTitle>
            <CardDescription className="text-slate-400">
              Rekomendasi sebelumnya
            </CardDescription>
          </CardHeader>
          <CardContent className="space-y-2">
            {insights.slice(1).map((insight) => (
              <div
                key={insight.id}
                className="p-3 rounded-lg border border-white/10 bg-white/5 flex items-center justify-between"
              >
                <div>
                  <p className="text-sm text-slate-200">Rekomendasi</p>
                  <p className="text-xs text-slate-500">
                    {new Date(insight.created_at).toLocaleString('id-ID')}
                  </p>
                </div>
                <Button
                  size="sm"
                  variant="outline"
                  className="border-white/10 text-slate-200"
                  onClick={() => setLatest(insight)}
                >
                  Lihat
                </Button>
              </div>
            ))}
          </CardContent>
        </Card>
      )}
    </div>
  )
}
