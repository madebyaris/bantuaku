import { useEffect, useState } from 'react'
import { Edit, CheckCircle, XCircle, Clock } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Card } from '@/components/ui/card'
import { api } from '@/lib/api'
import { toast } from '@/components/ui/toaster'
import { formatDate } from '@/lib/utils'
import { cn } from '@/lib/utils'

interface Subscription {
  id: string
  company_id: string
  company_name: string
  plan_id: string
  plan_name: string
  status: string
  current_period_start: string
  current_period_end: string
  created_at: string
}

export function SubscriptionsPage() {
  const [subscriptions, setSubscriptions] = useState<Subscription[]>([])
  const [loading, setLoading] = useState(true)
  const [page, setPage] = useState(1)
  const [total, setTotal] = useState(0)
  const [editingSub, setEditingSub] = useState<Subscription | null>(null)
  const [newStatus, setNewStatus] = useState('active')

  useEffect(() => {
    loadSubscriptions()
  }, [page])

  async function loadSubscriptions() {
    try {
      setLoading(true)
      const response = await api.admin.subscriptions.list(page, 20)
      setSubscriptions(response.subscriptions ?? [])
      setTotal(response.pagination?.total ?? (response.subscriptions?.length ?? 0))
    } catch (error) {
      toast({
        title: 'Error',
        description: error instanceof Error ? error.message : 'Failed to load subscriptions',
        variant: 'destructive',
      })
    } finally {
      setLoading(false)
    }
  }

  async function handleUpdateStatus() {
    if (!editingSub) return
    try {
      await api.admin.subscriptions.updateStatus(editingSub.id, newStatus)
      toast({ title: 'Success', description: 'Subscription status updated', variant: 'success' })
      setEditingSub(null)
      loadSubscriptions()
    } catch (error) {
      toast({
        title: 'Error',
        description: error instanceof Error ? error.message : 'Failed to update status',
        variant: 'destructive',
      })
    }
  }

  const getStatusBadge = (status: string) => {
    const statusConfig = {
      active: { icon: CheckCircle, color: 'bg-emerald-500/20 text-emerald-400 border-emerald-500/20' },
      canceled: { icon: XCircle, color: 'bg-red-500/20 text-red-400 border-red-500/20' },
      past_due: { icon: Clock, color: 'bg-amber-500/20 text-amber-400 border-amber-500/20' },
      trialing: { icon: Clock, color: 'bg-blue-500/20 text-blue-400 border-blue-500/20' },
    }
    return statusConfig[status as keyof typeof statusConfig] || statusConfig.active
  }

  return (
    <div className="space-y-6 animate-fade-in-up">
      <div>
        <h1 className="text-3xl font-display font-bold text-slate-100 mb-2">Subscriptions</h1>
        <p className="text-slate-400">Kelola langganan dan pembayaran</p>
      </div>

      <Card className="p-6 hover-card-effect">
        {loading ? (
          <div className="text-center py-12 text-slate-400">Loading...</div>
        ) : subscriptions.length === 0 ? (
          <div className="text-center py-12 text-slate-500">Belum ada subscription.</div>
        ) : (
          <div className="overflow-x-auto">
            <table className="w-full">
              <thead>
                <tr className="border-b border-white/10">
                  <th className="text-left py-3 px-4 text-sm font-semibold text-slate-300">Company</th>
                  <th className="text-left py-3 px-4 text-sm font-semibold text-slate-300">Plan</th>
                  <th className="text-left py-3 px-4 text-sm font-semibold text-slate-300">Status</th>
                  <th className="text-left py-3 px-4 text-sm font-semibold text-slate-300">Period</th>
                  <th className="text-right py-3 px-4 text-sm font-semibold text-slate-300">Actions</th>
                </tr>
              </thead>
              <tbody>
                {subscriptions.map((sub) => {
                  const statusConfig = getStatusBadge(sub.status)
                  const StatusIcon = statusConfig.icon
                  return (
                    <tr key={sub.id} className="border-b border-white/5 hover:bg-white/5 transition-colors">
                      <td className="py-3 px-4">
                        <span className="text-slate-200">{sub.company_name}</span>
                      </td>
                      <td className="py-3 px-4">
                        <span className="text-slate-300 font-medium">{sub.plan_name}</span>
                      </td>
                      <td className="py-3 px-4">
                        <span className={cn('inline-flex items-center gap-1 px-2 py-1 rounded-full text-xs font-medium border', statusConfig.color)}>
                          <StatusIcon className="w-3 h-3" />
                          {sub.status}
                        </span>
                      </td>
                      <td className="py-3 px-4 text-sm text-slate-400">
                        {formatDate(sub.current_period_start)} - {formatDate(sub.current_period_end)}
                      </td>
                      <td className="py-3 px-4">
                        <div className="flex items-center justify-end">
                          <Button
                            variant="ghost"
                            size="icon"
                            onClick={() => {
                              setEditingSub(sub)
                              setNewStatus(sub.status)
                            }}
                            className="h-8 w-8 text-slate-400 hover:text-emerald-400"
                          >
                            <Edit className="w-4 h-4" />
                          </Button>
                        </div>
                      </td>
                    </tr>
                  )
                })}
              </tbody>
            </table>
          </div>
        )}

        {total > 20 && (
          <div className="flex items-center justify-between mt-4 pt-4 border-t border-white/10">
            <p className="text-sm text-slate-400">
              Showing {(page - 1) * 20 + 1} to {Math.min(page * 20, total)} of {total} subscriptions
            </p>
            <div className="flex gap-2">
              <Button
                variant="outline"
                size="sm"
                onClick={() => setPage(p => Math.max(1, p - 1))}
                disabled={page === 1}
              >
                Previous
              </Button>
              <Button
                variant="outline"
                size="sm"
                onClick={() => setPage(p => p + 1)}
                disabled={page * 20 >= total}
              >
                Next
              </Button>
            </div>
          </div>
        )}
      </Card>

      {/* Edit Modal */}
      {editingSub && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/80 backdrop-blur-sm">
          <Card className="w-full max-w-md p-6 bg-black/90 border-white/20">
            <h2 className="text-xl font-bold text-slate-100 mb-4">Update Subscription Status</h2>
            <div className="space-y-4">
              <div>
                <label className="text-sm text-slate-300 mb-2 block">Company</label>
                <input value={editingSub.company_name} disabled className="w-full px-3 py-2 bg-white/5 border border-white/10 rounded-lg text-slate-400" />
              </div>
              <div>
                <label className="text-sm text-slate-300 mb-2 block">Status</label>
                <select
                  value={newStatus}
                  onChange={(e) => setNewStatus(e.target.value)}
                  className="w-full px-3 py-2 bg-white/5 border border-white/10 rounded-lg text-slate-200"
                >
                  <option value="active">Active</option>
                  <option value="canceled">Canceled</option>
                  <option value="past_due">Past Due</option>
                  <option value="trialing">Trialing</option>
                </select>
              </div>
              <div className="flex gap-2 pt-4">
                <Button onClick={handleUpdateStatus} className="flex-1">Update</Button>
                <Button variant="outline" onClick={() => setEditingSub(null)}>Cancel</Button>
              </div>
            </div>
          </Card>
        </div>
      )}
    </div>
  )
}

