import { useEffect, useState } from 'react'
import {
  Scale,
  AlertTriangle,
  CheckCircle2,
  FileText,
  ExternalLink,
  Calendar,
  Search,
  Loader2,
} from 'lucide-react'
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Input } from '@/components/ui/input'
import { api, Regulation } from '@/lib/api'

export function RegulationPage() {
  const [regulations, setRegulations] = useState<Regulation[]>([])
  const [loading, setLoading] = useState(false)
  const [query, setQuery] = useState('')
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    loadRegulations()
  }, [])

  async function loadRegulations() {
    setLoading(true)
    setError(null)
    try {
      const resp = await api.regulations.list(undefined, 20, 0)
      setRegulations(resp.regulations || [])
    } catch (err) {
      setError('Gagal memuat regulasi')
    } finally {
      setLoading(false)
    }
  }

  async function handleSearch(e: React.FormEvent) {
    e.preventDefault()
    if (!query.trim()) {
      loadRegulations()
      return
    }
    setLoading(true)
    setError(null)
    try {
      const resp = await api.regulations.search(query.trim(), 10)
      const results = resp.results?.map((r) => r.regulation) || []
      setRegulations(results)
    } catch (err) {
      setError('Gagal mencari regulasi')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="max-w-7xl mx-auto px-4 py-8 animate-fade-in-up space-y-8 pb-20">
      <div className="flex flex-col md:flex-row md:items-end justify-between gap-4 border-b border-white/10 pb-6">
        <div>
          <div className="flex items-center gap-2 mb-2">
            <div className="p-2 bg-orange-500/10 rounded-lg border border-orange-500/20">
              <Scale className="w-5 h-5 text-orange-400" />
            </div>
            <span className="text-sm font-medium text-orange-400 uppercase tracking-wider">
              Legal & Compliance
            </span>
          </div>
          <h1 className="text-3xl md:text-4xl font-display font-bold text-slate-100">
            Peraturan Pemerintah
          </h1>
          <p className="text-slate-400 mt-2 max-w-2xl text-lg">
            Monitor kepatuhan bisnis Anda terhadap regulasi terbaru.
          </p>
        </div>
        <form onSubmit={handleSearch} className="flex items-center gap-2 w-full md:w-80">
          <Input
            placeholder="Cari regulasi..."
            value={query}
            onChange={(e) => setQuery(e.target.value)}
            className="bg-white/5 border-white/10"
          />
          <Button type="submit" variant="outline" className="border-white/10 text-slate-200">
            <Search className="w-4 h-4 mr-2" />
            Cari
          </Button>
        </form>
      </div>

      {error && (
        <Card className="border-red-500/20 bg-red-500/5">
          <CardContent className="py-3 text-sm text-red-200">{error}</CardContent>
        </Card>
      )}

      {loading ? (
        <div className="flex justify-center py-12">
          <Loader2 className="w-8 h-8 animate-spin text-orange-400" />
        </div>
      ) : regulations.length === 0 ? (
        <Card>
          <CardContent className="py-12 text-center text-slate-400">
            Tidak ada regulasi ditemukan.
          </CardContent>
        </Card>
      ) : (
        <div className="grid grid-cols-1 gap-4">
          {regulations.map((item) => (
            <Card key={item.id} className="hover-card-effect border-white/10 bg-white/5 group">
              <CardContent className="p-6 flex flex-col md:flex-row gap-6 items-start md:items-center">
                <div className="p-3 rounded-xl bg-white/5 border border-white/5 group-hover:border-white/10 transition-colors">
                  {item.status === 'urgent' && <AlertTriangle className="w-6 h-6 text-red-400" />}
                  {item.status === 'compliant' && <CheckCircle2 className="w-6 h-6 text-emerald-400" />}
                  {!item.status && <FileText className="w-6 h-6 text-blue-400" />}
                </div>

                <div className="flex-1">
                  <div className="flex items-center gap-3 mb-1">
                    <h3 className="font-bold text-lg text-slate-100">{item.title}</h3>
                    {item.category && (
                      <Badge
                        variant="outline"
                        className="bg-white/5 text-slate-400 border-white/10 font-normal"
                      >
                        {item.category}
                      </Badge>
                    )}
                    {item.status === 'urgent' && (
                      <Badge className="bg-red-500/20 text-red-400 hover:bg-red-500/30 border-red-500/20">
                        Urgent
                      </Badge>
                    )}
                  </div>
                  <p className="text-slate-400 mb-2">
                    {item.published_date || item.effective_date
                      ? `Diterbitkan: ${item.published_date || '-'}`
                      : 'Tanggal tidak tersedia'}
                  </p>
                  <p className="text-slate-400 mb-2">{item.title}</p>
                  <div className="flex items-center gap-4 text-xs text-slate-500">
                    <span className="flex items-center gap-1">
                      <Calendar className="w-3 h-3" /> Status: {item.status || 'unknown'}
                    </span>
                  </div>
                </div>

                <Button
                  asChild
                  variant="outline"
                  className="border-white/10 bg-transparent hover:bg-white/5 text-slate-300 group-hover:text-white shrink-0 w-full md:w-auto"
                >
                  <a href={item.pdf_url || item.source_url} target="_blank" rel="noreferrer">
                    Lihat Dokumen <ExternalLink className="w-3 h-3 ml-2 opacity-50" />
                  </a>
                </Button>
              </CardContent>
            </Card>
          ))}
        </div>
      )}
    </div>
  )
}
