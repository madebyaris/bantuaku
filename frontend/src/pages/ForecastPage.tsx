import { useEffect, useState } from 'react'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { MessageSquareText } from 'lucide-react'

export function ForecastPage() {
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    // TODO: Fetch forecast data from API
    setLoading(false)
  }, [])

  return (
    <div className="container mx-auto px-4 py-8 animate-fade-in-up">
      <div className="mb-6 flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold font-display text-slate-100">Forecast</h1>
          <p className="text-slate-400 mt-2">
            Proyeksi penjualan untuk produk Anda berdasarkan data yang Anda berikan
          </p>
        </div>
        <Button variant="outline" className="gap-2 border-white/10 bg-white/5 text-slate-300 hover:bg-white/10 hover:text-emerald-400 hover:border-emerald-500/30">
          <MessageSquareText className="w-4 h-4" />
          Tanya AI tentang forecast ini
        </Button>
      </div>

      {loading ? (
        <div className="flex items-center justify-center py-12">
          <p className="text-slate-500">Memuat data forecast...</p>
        </div>
      ) : (
        <Card className="hover-card-effect border-white/10 bg-white/5 backdrop-blur-xl">
          <CardHeader>
            <CardTitle className="text-slate-100">Forecast Penjualan</CardTitle>
          </CardHeader>
          <CardContent>
            <p className="text-slate-400">
              Fitur forecast sedang dalam pengembangan. Data akan ditampilkan di sini setelah
              Anda menginput data melalui AI Assistant.
            </p>
          </CardContent>
        </Card>
      )}
    </div>
  )
}
