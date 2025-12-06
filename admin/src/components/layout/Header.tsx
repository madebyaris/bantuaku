import { useLocation } from 'react-router-dom'
import { Menu } from 'lucide-react'
import { Button } from '@/components/ui/button'

const pageTitles: Record<string, { title: string; description: string }> = {
  '/dashboard': {
    title: 'Dashboard',
    description: 'Overview sistem dan statistik',
  },
  '/users': {
    title: 'User Management',
    description: 'Kelola pengguna dan peran',
  },
  '/subscriptions': {
    title: 'Subscriptions',
    description: 'Kelola langganan dan pembayaran',
  },
  '/audit-logs': {
    title: 'Audit Logs',
    description: 'Riwayat aktivitas sistem',
  },
}

interface HeaderProps {
  onOpenSidebar: () => void
}

export function Header({ onOpenSidebar }: HeaderProps) {
  const location = useLocation()
  
  const pageInfo = pageTitles[location.pathname] || {
    title: 'Admin Panel',
    description: '',
  }

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
      </div>
    </header>
  )
}

