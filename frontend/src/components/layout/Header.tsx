import { useLocation } from 'react-router-dom'
import { Bell, Search } from 'lucide-react'
import { Input } from '@/components/ui/input'
import { Button } from '@/components/ui/button'

const pageTitles: Record<string, { title: string; description: string }> = {
  '/dashboard': {
    title: 'Dashboard',
    description: 'Pantau performa bisnis Anda secara real-time',
  },
  '/forecast': {
    title: 'Forecast Penjualan',
    description: 'Proyeksi penjualan untuk perencanaan stok yang lebih baik',
  },
  '/market-prediction': {
    title: 'Prediksi Pasar',
    description: 'Analisis tren pasar dan peluang bisnis',
  },
  '/marketing': {
    title: 'Rekomendasi Marketing',
    description: 'Strategi pemasaran berbasis data',
  },
  '/regulation': {
    title: 'Peraturan Pemerintah',
    description: 'Update regulasi UMKM terkini',
  },
  '/ai-chat': {
    title: 'AI Assistant',
    description: 'Tanya AI untuk insight dan rekomendasi bisnis',
  },
}

export function Header() {
  const location = useLocation()
  const pageInfo = pageTitles[location.pathname] || {
    title: 'Bantuaku',
    description: '',
  }

  return (
    <header className="sticky top-0 z-40 bg-black/50 backdrop-blur-xl border-b border-white/10">
      <div className="flex items-center justify-between px-6 py-4">
        <div>
          <h2 className="text-2xl font-display font-bold text-slate-100">
            {pageInfo.title}
          </h2>
          {pageInfo.description && (
            <p className="text-sm text-slate-400 mt-0.5">
              {pageInfo.description}
            </p>
          )}
        </div>

        <div className="flex items-center gap-4">
          {/* Search */}
          <div className="relative hidden md:block">
            <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-slate-500" />
            <Input
              type="search"
              placeholder="Cari insight..."
              className="w-64 pl-10 bg-white/5 border-white/10 text-slate-200 placeholder:text-slate-600 focus:bg-white/10 focus:border-emerald-500/50"
            />
          </div>

          {/* Notifications */}
          <Button variant="ghost" size="icon" className="relative hover:bg-white/5 text-slate-400 hover:text-emerald-400">
            <Bell className="w-5 h-5" />
            <span className="absolute top-2.5 right-2.5 w-2 h-2 bg-emerald-500 rounded-full shadow-[0_0_8px_rgba(16,185,129,0.5)]" />
          </Button>
        </div>
      </div>
    </header>
  )
}
