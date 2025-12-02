import { Link, useLocation } from 'react-router-dom'
import {
  LayoutDashboard,
  TrendingUp,
  Globe,
  Megaphone,
  Scale,
  MessageSquareText,
  LogOut,
  Sparkles,
} from 'lucide-react'
import { cn } from '@/lib/utils'
import { useAuthStore } from '@/state/auth'

const navigation = [
  { name: 'Dashboard', href: '/dashboard', icon: LayoutDashboard },
  { name: 'Forecast', href: '/forecast', icon: TrendingUp },
  { name: 'Market Prediction', href: '/market-prediction', icon: Globe },
  { name: 'Marketing Recommendation', href: '/marketing', icon: Megaphone },
  { name: 'Government Regulation', href: '/regulation', icon: Scale },
  { name: 'AI Assistant', href: '/ai-chat', icon: MessageSquareText },
]

export function Sidebar() {
  const location = useLocation()
  const { storeName, plan, logout } = useAuthStore()

  return (
    <div className="fixed inset-y-0 left-0 z-50 w-64 bg-white border-r border-slate-200">
      {/* Logo */}
      <div className="flex items-center gap-2 px-6 py-5 border-b border-slate-100">
        <div className="flex items-center justify-center w-10 h-10 rounded-xl bg-gradient-to-br from-purple-600 to-indigo-600 shadow-lg shadow-purple-500/20">
          <Sparkles className="w-5 h-5 text-white" />
        </div>
        <div>
          <h1 className="font-display text-xl font-bold text-slate-900">
            Bantuaku
          </h1>
          <p className="text-xs text-slate-500">Forecasting UMKM</p>
        </div>
      </div>

      {/* Store info */}
      <div className="px-4 py-4 border-b border-slate-100">
        <div className="px-3 py-2 rounded-lg bg-slate-50">
          <p className="font-medium text-sm text-slate-900 truncate">
            {storeName || 'Toko Anda'}
          </p>
          <span
            className={cn(
              'inline-flex items-center px-2 py-0.5 mt-1 rounded-full text-xs font-medium',
              plan === 'pro'
                ? 'bg-purple-100 text-purple-700'
                : plan === 'enterprise'
                ? 'bg-amber-100 text-amber-700'
                : 'bg-slate-200 text-slate-600'
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
                'flex items-center gap-3 px-3 py-2.5 rounded-lg text-sm font-medium transition-all duration-200',
                isActive
                  ? 'bg-purple-50 text-purple-700 shadow-sm'
                  : 'text-slate-600 hover:bg-slate-50 hover:text-slate-900'
              )}
            >
              <item.icon
                className={cn(
                  'w-5 h-5 transition-colors',
                  isActive ? 'text-purple-600' : 'text-slate-400'
                )}
              />
              {item.name}
              {item.name === 'AI Assistant' && (
                <span className="ml-auto px-1.5 py-0.5 text-[10px] font-bold bg-gradient-to-r from-purple-600 to-indigo-600 text-white rounded">
                  AI
                </span>
              )}
            </Link>
          )
        })}
      </nav>

      {/* Logout */}
      <div className="px-4 py-4 border-t border-slate-100">
        <button
          onClick={logout}
          className="flex items-center gap-3 w-full px-3 py-2.5 rounded-lg text-sm font-medium text-slate-600 hover:bg-red-50 hover:text-red-600 transition-colors"
        >
          <LogOut className="w-5 h-5" />
          Keluar
        </button>
      </div>
    </div>
  )
}
