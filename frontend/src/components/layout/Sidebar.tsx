import { Link, useLocation } from 'react-router-dom'
import {
  LayoutDashboard,
  TrendingUp,
  Globe,
  Megaphone,
  Scale,
  MessageSquareText,
  LogOut,
  Zap,
} from 'lucide-react'
import { cn } from '@/lib/utils'
import { useAuthStore } from '@/state/auth'

const navigation = [
  { name: 'Dashboard', href: '/dashboard', icon: LayoutDashboard },
  { name: 'AI Assistant', href: '/ai-chat', icon: MessageSquareText },
  { name: 'Forecast', href: '/forecast', icon: TrendingUp },
  { name: 'Market Prediction', href: '/market-prediction', icon: Globe },
  { name: 'Marketing Recommendation', href: '/marketing', icon: Megaphone },
  { name: 'Government Regulation', href: '/regulation', icon: Scale },
]

export function Sidebar() {
  const location = useLocation()
  const { storeName, plan, logout } = useAuthStore()

  return (
    <div className="fixed inset-y-0 left-0 z-50 w-64 bg-black/50 backdrop-blur-xl border-r border-white/10">
      {/* Logo */}
      <div className="flex items-center gap-3 px-6 py-5 border-b border-white/5">
        <div className="flex items-center justify-center w-10 h-10 rounded-full bg-gradient-to-br from-emerald-400 to-emerald-600 shadow-[0_0_15px_rgba(16,185,129,0.4)]">
          <Zap className="w-5 h-5 text-black fill-black" />
        </div>
        <div>
          <h1 className="font-display text-xl font-bold text-slate-100">
            Bantuaku
          </h1>
          <p className="text-xs text-slate-500">Forecasting UMKM</p>
        </div>
      </div>

      {/* Store info */}
      <div className="px-4 py-4 border-b border-white/5">
        <div className="px-3 py-2 rounded-lg bg-white/5 border border-white/5">
          <p className="font-medium text-sm text-slate-200 truncate">
            {storeName || 'Toko Anda'}
          </p>
          <span
            className={cn(
              'inline-flex items-center px-2 py-0.5 mt-1 rounded-full text-xs font-medium',
              plan === 'pro'
                ? 'bg-emerald-500/20 text-emerald-400 border border-emerald-500/20'
                : plan === 'enterprise'
                ? 'bg-amber-500/20 text-amber-400 border border-amber-500/20'
                : 'bg-slate-500/20 text-slate-400 border border-slate-500/20'
            )}
          >
            {plan === 'pro' ? 'Pro' : plan === 'enterprise' ? 'Enterprise' : 'Free'}
          </span>
        </div>
      </div>

      {/* Navigation */}
      <nav className="flex-1 px-4 py-4 space-y-1">
        {navigation.map((item) => {
          const isActive = location.pathname === item.href
          return (
            <Link
              key={item.name}
              to={item.href}
              className={cn(
                'flex items-center gap-3 px-3 py-2.5 rounded-lg text-sm font-medium transition-all duration-200 group',
                isActive
                  ? 'bg-white/10 text-emerald-400 shadow-sm border border-white/5'
                  : 'text-slate-400 hover:bg-white/5 hover:text-slate-200'
              )}
            >
              <item.icon
                className={cn(
                  'w-5 h-5 transition-colors',
                  isActive ? 'text-emerald-400' : 'text-slate-500 group-hover:text-slate-300'
                )}
              />
              {item.name}
              {item.name === 'AI Assistant' && (
                <span className="ml-auto px-1.5 py-0.5 text-[10px] font-bold bg-gradient-to-r from-emerald-500 to-emerald-400 text-black rounded shadow-[0_0_10px_rgba(16,185,129,0.3)]">
                  AI
                </span>
              )}
            </Link>
          )
        })}
      </nav>

      {/* Logout */}
      <div className="px-4 py-4 border-t border-white/5">
        <button
          onClick={logout}
          className="flex items-center gap-3 w-full px-3 py-2.5 rounded-lg text-sm font-medium text-slate-400 hover:bg-red-500/10 hover:text-red-400 transition-colors border border-transparent hover:border-red-500/10"
        >
          <LogOut className="w-5 h-5" />
          Keluar
        </button>
      </div>
    </div>
  )
}
