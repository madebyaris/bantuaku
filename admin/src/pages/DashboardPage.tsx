import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import { Users, CreditCard, FileText, Shield } from 'lucide-react'
import { Card } from '@/components/ui/card'
import { api } from '@/lib/api'

interface Stats {
  totalUsers: number
  totalSubscriptions: number
  activeSubscriptions: number
  totalAuditLogs: number
}

export function DashboardPage() {
  const [stats, setStats] = useState<Stats>({
    totalUsers: 0,
    totalSubscriptions: 0,
    activeSubscriptions: 0,
    totalAuditLogs: 0,
  })
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    async function loadStats() {
      try {
        // #region agent log
        fetch('http://127.0.0.1:7242/ingest/caa1e494-1c2c-46ae-ab69-48afbc48a0f9',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify({location:'DashboardPage.tsx:loadStats:entry',message:'loadStats called',data:{},timestamp:Date.now(),sessionId:'debug-session',runId:'run1',hypothesisId:'C'})}).catch(()=>{});
        // #endregion
        const statsRes = await api.admin.stats.get()
        // #region agent log
        fetch('http://127.0.0.1:7242/ingest/caa1e494-1c2c-46ae-ab69-48afbc48a0f9',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify({location:'DashboardPage.tsx:loadStats:success',message:'stats.get success',data:{totalUsers:statsRes.total_users},timestamp:Date.now(),sessionId:'debug-session',runId:'run1',hypothesisId:'C'})}).catch(()=>{});
        // #endregion
        setStats({
          totalUsers: statsRes.total_users,
          totalSubscriptions: statsRes.total_subscriptions,
          activeSubscriptions: statsRes.active_subscriptions,
          totalAuditLogs: statsRes.total_audit_logs,
        })
      } catch (error) {
        // #region agent log
        fetch('http://127.0.0.1:7242/ingest/caa1e494-1c2c-46ae-ab69-48afbc48a0f9',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify({location:'DashboardPage.tsx:loadStats:error',message:'stats.get failed',data:{error:error instanceof Error?error.message:String(error)},timestamp:Date.now(),sessionId:'debug-session',runId:'run1',hypothesisId:'C'})}).catch(()=>{});
        // #endregion
        console.error('Failed to load stats:', error)
      } finally {
        setLoading(false)
      }
    }
    loadStats()
  }, [])

  const statCards = [
    {
      title: 'Total Users',
      value: stats.totalUsers,
      icon: Users,
      color: 'from-blue-500 to-blue-600',
      bgColor: 'bg-blue-500/10',
      borderColor: 'border-blue-500/20',
    },
    {
      title: 'Total Subscriptions',
      value: stats.totalSubscriptions,
      icon: CreditCard,
      color: 'from-emerald-500 to-emerald-600',
      bgColor: 'bg-emerald-500/10',
      borderColor: 'border-emerald-500/20',
    },
    {
      title: 'Active Subscriptions',
      value: stats.activeSubscriptions,
      icon: Shield,
      color: 'from-purple-500 to-purple-600',
      bgColor: 'bg-purple-500/10',
      borderColor: 'border-purple-500/20',
    },
    {
      title: 'Audit Logs',
      value: stats.totalAuditLogs,
      icon: FileText,
      color: 'from-amber-500 to-amber-600',
      bgColor: 'bg-amber-500/10',
      borderColor: 'border-amber-500/20',
    },
  ]

  const quickActions = [
    {
      title: 'User Management',
      description: 'Kelola pengguna dan peran',
      href: '/users',
    },
    {
      title: 'Subscriptions',
      description: 'Kelola langganan dan pembayaran',
      href: '/subscriptions',
    },
    {
      title: 'Audit Logs',
      description: 'Lihat riwayat aktivitas sistem',
      href: '/audit-logs',
    },
  ]

  return (
    <div className="space-y-6 animate-fade-in-up">
      <div>
        <h1 className="text-3xl font-display font-bold text-slate-100 mb-2">Admin Dashboard</h1>
        <p className="text-slate-400">Overview sistem dan statistik platform</p>
      </div>

      {loading ? (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
          {[1, 2, 3, 4].map((i) => (
            <Card key={i} className="p-6 hover-card-effect animate-pulse">
              <div className="h-20 bg-white/5 rounded-lg" />
            </Card>
          ))}
        </div>
      ) : (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
          {statCards.map((stat) => {
            const Icon = stat.icon
            return (
              <Card
                key={stat.title}
                className={`p-6 hover-card-effect ${stat.bgColor} ${stat.borderColor} border`}
              >
                <div className="flex items-center justify-between mb-4">
                  <div className={`p-3 rounded-xl bg-gradient-to-br ${stat.color} shadow-lg`}>
                    <Icon className="w-6 h-6 text-white" />
                  </div>
                </div>
                <h3 className="text-sm font-medium text-slate-400 mb-1">{stat.title}</h3>
                <p className="text-3xl font-bold text-slate-100">{stat.value.toLocaleString('id-ID')}</p>
              </Card>
            )
          })}
        </div>
      )}

      <Card className="p-6 hover-card-effect">
        <h2 className="text-xl font-display font-bold text-slate-100 mb-4">Quick Actions</h2>
        <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
          {quickActions.map((action) => (
            <Link
              key={action.title}
              to={action.href}
              className="p-4 rounded-lg bg-white/5 border border-white/10 hover:bg-white/10 transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-emerald-500/70 focus-visible:ring-offset-2 focus-visible:ring-offset-black"
            >
              <h3 className="font-semibold text-slate-200 mb-2">{action.title}</h3>
              <p className="text-sm text-slate-400">{action.description}</p>
            </Link>
          ))}
        </div>
      </Card>
    </div>
  )
}

