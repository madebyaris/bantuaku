import { useState, useRef, useEffect } from 'react'
import { useLocation } from 'react-router-dom'
import { Bell, Search, Menu, X, Info, CheckCircle, AlertTriangle } from 'lucide-react'
import { Input } from '@/components/ui/input'
import { Button } from '@/components/ui/button'
import { cn } from '@/lib/utils'

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

interface Notification {
  id: string
  title: string
  message: string
  type: 'info' | 'success' | 'warning'
  time: string
  read: boolean
}

const mockNotifications: Notification[] = [
  {
    id: '1',
    title: 'Forecast Selesai',
    message: 'Analisis forecast penjualan bulan depan telah siap.',
    type: 'success',
    time: 'Baru saja',
    read: false,
  },
  {
    id: '2',
    title: 'Update Regulasi',
    message: 'Ada perubahan peraturan pajak UMKM yang perlu diperhatikan.',
    type: 'warning',
    time: '1 jam yang lalu',
    read: false,
  },
  {
    id: '3',
    title: 'Tren Pasar Baru',
    message: 'Produk kategori "Minuman Sehat" sedang naik daun.',
    type: 'info',
    time: '2 jam yang lalu',
    read: true,
  },
]

interface HeaderProps {
  onOpenSidebar: () => void
}

export function Header({ onOpenSidebar }: HeaderProps) {
  const location = useLocation()
  const [showNotifications, setShowNotifications] = useState(false)
  const notificationRef = useRef<HTMLDivElement>(null)
  
  const pageInfo = pageTitles[location.pathname] || {
    title: 'Bantuaku',
    description: '',
  }

  // Close notifications when clicking outside
  useEffect(() => {
    function handleClickOutside(event: MouseEvent) {
      if (notificationRef.current && !notificationRef.current.contains(event.target as Node)) {
        setShowNotifications(false)
      }
    }
    document.addEventListener('mousedown', handleClickOutside)
    return () => {
      document.removeEventListener('mousedown', handleClickOutside)
    }
  }, [])

  return (
    <header className="sticky top-0 z-40 bg-black/50 backdrop-blur-xl border-b border-white/10">
      <div className="flex items-center justify-between px-4 lg:px-6 py-4">
        <div className="flex items-center gap-4">
          <Button
            variant="ghost"
            size="icon"
            className="lg:hidden text-slate-400 hover:text-white -ml-2"
            onClick={onOpenSidebar}
          >
            <Menu className="w-6 h-6" />
          </Button>
          
          <div>
            <h2 className="text-xl lg:text-2xl font-display font-bold text-slate-100">
              {pageInfo.title}
            </h2>
            {pageInfo.description && (
              <p className="hidden sm:block text-sm text-slate-400 mt-0.5">
                {pageInfo.description}
              </p>
            )}
          </div>
        </div>

        <div className="flex items-center gap-2 lg:gap-4">
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
          <div className="relative" ref={notificationRef}>
            <Button 
              variant="ghost" 
              size="icon" 
              className={cn(
                "relative hover:bg-white/5 text-slate-400 hover:text-emerald-400 transition-colors",
                showNotifications && "text-emerald-400 bg-white/5"
              )}
              onClick={() => setShowNotifications(!showNotifications)}
            >
              <Bell className="w-5 h-5" />
              <span className="absolute top-2.5 right-2.5 w-2 h-2 bg-emerald-500 rounded-full shadow-[0_0_8px_rgba(16,185,129,0.5)]" />
            </Button>

            {/* Notification Dropdown */}
            {showNotifications && (
              <div className="absolute right-0 mt-2 w-80 md:w-96 bg-[#0a0a0a] border border-white/10 rounded-xl shadow-2xl overflow-hidden animate-in fade-in slide-in-from-top-2 z-50 ring-1 ring-white/5">
                <div className="flex items-center justify-between px-4 py-3 border-b border-white/10 bg-white/5">
                  <h3 className="font-semibold text-slate-100">Notifikasi</h3>
                  <Button variant="ghost" size="icon" className="h-6 w-6 text-slate-400 hover:text-white" onClick={() => setShowNotifications(false)}>
                    <X className="w-4 h-4" />
                  </Button>
                </div>
                <div className="max-h-[400px] overflow-y-auto bg-[#0a0a0a]">
                  {mockNotifications.map((notification) => (
                    <div 
                      key={notification.id}
                      className={cn(
                        "p-4 border-b border-white/5 hover:bg-white/5 transition-colors cursor-pointer",
                        !notification.read && "bg-emerald-500/5"
                      )}
                    >
                      <div className="flex gap-3">
                        <div className={cn(
                          "mt-1 w-2 h-2 rounded-full shrink-0",
                          notification.type === 'success' ? "bg-emerald-500 shadow-[0_0_8px_rgba(16,185,129,0.5)]" :
                          notification.type === 'warning' ? "bg-amber-500 shadow-[0_0_8px_rgba(245,158,11,0.5)]" :
                          "bg-blue-500 shadow-[0_0_8px_rgba(59,130,246,0.5)]"
                        )} />
                        <div className="flex-1 space-y-1">
                          <div className="flex items-center justify-between">
                            <p className={cn("text-sm font-medium", !notification.read ? "text-slate-100" : "text-slate-400")}>
                              {notification.title}
                            </p>
                            <span className="text-[10px] text-slate-500">{notification.time}</span>
                          </div>
                          <p className="text-xs text-slate-400 leading-relaxed">
                            {notification.message}
                          </p>
                        </div>
                      </div>
                    </div>
                  ))}
                </div>
                <div className="p-2 border-t border-white/10 bg-white/5 text-center">
                  <Button variant="ghost" size="sm" className="text-xs text-emerald-400 hover:text-emerald-300 hover:bg-transparent w-full h-8">
                    Tandai semua sudah dibaca
                  </Button>
                </div>
              </div>
            )}
          </div>
        </div>
      </div>
    </header>
  )
}
