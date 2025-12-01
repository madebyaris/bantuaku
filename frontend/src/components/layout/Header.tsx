import { useLocation } from 'react-router-dom'
import { Bell, Search } from 'lucide-react'
import { Input } from '@/components/ui/input'
import { Button } from '@/components/ui/button'

const pageTitles: Record<string, { title: string; description: string }> = {
  '/dashboard': {
    title: 'Dashboard',
    description: 'Pantau performa bisnis Anda secara real-time',
  },
  '/products': {
    title: 'Produk',
    description: 'Kelola inventaris dan lihat forecast penjualan',
  },
  '/data-input': {
    title: 'Input Data',
    description: 'Tambahkan data penjualan manual atau import CSV',
  },
  '/integrations': {
    title: 'Integrasi',
    description: 'Hubungkan toko online Anda dengan Bantuaku',
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
    <header className="sticky top-0 z-40 bg-white/80 backdrop-blur-lg border-b border-slate-200">
      <div className="flex items-center justify-between px-6 py-4">
        <div>
          <h2 className="text-2xl font-display font-bold text-slate-900">
            {pageInfo.title}
          </h2>
          {pageInfo.description && (
            <p className="text-sm text-slate-500 mt-0.5">
              {pageInfo.description}
            </p>
          )}
        </div>

        <div className="flex items-center gap-4">
          {/* Search */}
          <div className="relative hidden md:block">
            <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-slate-400" />
            <Input
              type="search"
              placeholder="Cari produk..."
              className="w-64 pl-10 bg-slate-50 border-slate-200 focus:bg-white"
            />
          </div>

          {/* Notifications */}
          <Button variant="ghost" size="icon" className="relative">
            <Bell className="w-5 h-5 text-slate-600" />
            <span className="absolute top-1 right-1 w-2 h-2 bg-red-500 rounded-full" />
          </Button>
        </div>
      </div>
    </header>
  )
}
