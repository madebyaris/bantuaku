import { useEffect, useState } from 'react'
import { Edit, CheckCircle, XCircle, Clock, Plus, TrendingUp, Users, CreditCard, AlertCircle, Package } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Card } from '@/components/ui/card'
import { api } from '@/lib/api'
import { toast } from '@/components/ui/toaster'
import { formatDateShort } from '@/lib/utils'
import { cn } from '@/lib/utils'

interface Subscription {
  id: string
  company_id: string
  company_name: string
  owner_email: string
  plan_id: string
  plan_name: string
  status: string
  current_period_start: string
  current_period_end: string
  created_at: string
}

interface SubscriptionStats {
  total_subscriptions: number
  active_subscriptions: number
  trialing_count: number
  canceled_count: number
  past_due_count: number
  mrr: number
  plan_breakdown: Array<{
    plan_id: string
    plan_name: string
    count: number
    price_monthly: number
  }>
}

interface Plan {
  id: string
  name: string
  display_name: string
  price_monthly: number
}

interface Company {
  id: string
  name: string
}

export function SubscriptionsPage() {
  const [subscriptions, setSubscriptions] = useState<Subscription[]>([])
  const [stats, setStats] = useState<SubscriptionStats | null>(null)
  const [loading, setLoading] = useState(true)
  const [page, setPage] = useState(1)
  const [total, setTotal] = useState(0)
  const [editingSub, setEditingSub] = useState<Subscription | null>(null)
  const [newStatus, setNewStatus] = useState('active')
  const [showCreateModal, setShowCreateModal] = useState(false)
  const [plans, setPlans] = useState<Plan[]>([])
  const [companies, setCompanies] = useState<Company[]>([])
  const [newSubscription, setNewSubscription] = useState({ company_id: '', plan_id: '' })
  const [creating, setCreating] = useState(false)

  useEffect(() => {
    loadData()
  }, [page])

  async function loadData() {
    try {
      setLoading(true)
      const [subResponse, statsResponse] = await Promise.all([
        api.admin.subscriptions.list(page, 20),
        api.admin.subscriptions.stats(),
      ])
      setSubscriptions(subResponse.subscriptions ?? [])
      setTotal(subResponse.pagination?.total ?? (subResponse.subscriptions?.length ?? 0))
      setStats(statsResponse)
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

  async function loadFormData() {
    try {
      const [plansResponse, usersResponse] = await Promise.all([
        api.admin.plans.list(1, 100),
        api.admin.users.list(1, 100),
      ])
      setPlans(plansResponse.plans?.filter(p => p.is_active) ?? [])
      // Extract companies from users (those with company_id and store_name are companies)
      const companiesFromUsers = usersResponse.users
        ?.filter(u => u.company_id && u.store_name)
        .map(u => ({ id: u.company_id!, name: u.store_name! })) ?? []
      setCompanies(companiesFromUsers)
    } catch (error) {
      console.error('Failed to load form data:', error)
    }
  }

  async function handleUpdateStatus() {
    if (!editingSub) return
    try {
      await api.admin.subscriptions.updateStatus(editingSub.id, newStatus)
      toast({ title: 'Success', description: 'Subscription status updated', variant: 'success' })
      setEditingSub(null)
      loadData()
    } catch (error) {
      toast({
        title: 'Error',
        description: error instanceof Error ? error.message : 'Failed to update status',
        variant: 'destructive',
      })
    }
  }

  async function handleCreateSubscription() {
    if (!newSubscription.company_id || !newSubscription.plan_id) {
      toast({ title: 'Error', description: 'Please select a company and plan', variant: 'destructive' })
      return
    }
    try {
      setCreating(true)
      await api.admin.subscriptions.create({
        company_id: newSubscription.company_id,
        plan_id: newSubscription.plan_id,
      })
      toast({ title: 'Success', description: 'Subscription created successfully', variant: 'success' })
      setShowCreateModal(false)
      setNewSubscription({ company_id: '', plan_id: '' })
      loadData()
    } catch (error) {
      toast({
        title: 'Error',
        description: error instanceof Error ? error.message : 'Failed to create subscription',
        variant: 'destructive',
      })
    } finally {
      setCreating(false)
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

  const formatCurrency = (amount: number) => {
    return new Intl.NumberFormat('id-ID', {
      style: 'currency',
      currency: 'IDR',
      minimumFractionDigits: 0,
      maximumFractionDigits: 0,
    }).format(amount)
  }

  return (
    <div className="space-y-6 animate-fade-in-up">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-display font-bold text-slate-100 mb-2">Subscriptions</h1>
          <p className="text-slate-400">Kelola langganan dan pembayaran</p>
        </div>
        <Button
          onClick={() => {
            loadFormData()
            setShowCreateModal(true)
          }}
          className="bg-gradient-to-r from-emerald-600 to-emerald-500 hover:from-emerald-500 hover:to-emerald-400"
        >
          <Plus className="w-4 h-4 mr-2" />
          New Subscription
        </Button>
      </div>

      {/* Stats Cards */}
      {stats && (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
          <Card className="p-5 hover-card-effect">
            <div className="flex items-center gap-4">
              <div className="p-3 rounded-xl bg-emerald-500/20">
                <TrendingUp className="w-6 h-6 text-emerald-400" />
              </div>
              <div>
                <p className="text-sm text-slate-400">Monthly Revenue</p>
                <p className="text-2xl font-bold text-slate-100">{formatCurrency(stats.mrr)}</p>
              </div>
            </div>
          </Card>

          <Card className="p-5 hover-card-effect">
            <div className="flex items-center gap-4">
              <div className="p-3 rounded-xl bg-blue-500/20">
                <Users className="w-6 h-6 text-blue-400" />
              </div>
              <div>
                <p className="text-sm text-slate-400">Active Subscriptions</p>
                <p className="text-2xl font-bold text-slate-100">{stats.active_subscriptions}</p>
              </div>
            </div>
          </Card>

          <Card className="p-5 hover-card-effect">
            <div className="flex items-center gap-4">
              <div className="p-3 rounded-xl bg-purple-500/20">
                <CreditCard className="w-6 h-6 text-purple-400" />
              </div>
              <div>
                <p className="text-sm text-slate-400">Trialing</p>
                <p className="text-2xl font-bold text-slate-100">{stats.trialing_count}</p>
              </div>
            </div>
          </Card>

          <Card className="p-5 hover-card-effect">
            <div className="flex items-center gap-4">
              <div className="p-3 rounded-xl bg-amber-500/20">
                <AlertCircle className="w-6 h-6 text-amber-400" />
              </div>
              <div>
                <p className="text-sm text-slate-400">Past Due</p>
                <p className="text-2xl font-bold text-slate-100">{stats.past_due_count}</p>
              </div>
            </div>
          </Card>
        </div>
      )}

      {/* Plan Breakdown */}
      {stats && stats.plan_breakdown && stats.plan_breakdown.length > 0 && (
        <Card className="p-6 hover-card-effect">
          <h2 className="text-lg font-semibold text-slate-100 mb-4 flex items-center gap-2">
            <Package className="w-5 h-5 text-emerald-400" />
            Plan Distribution
          </h2>
          <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
            {stats.plan_breakdown.map((plan) => (
              <div
                key={plan.plan_id}
                className="p-4 rounded-xl bg-white/5 border border-white/10"
              >
                <div className="flex items-center justify-between mb-2">
                  <span className="font-medium text-slate-200">{plan.plan_name}</span>
                  <span className="text-2xl font-bold text-emerald-400">{plan.count}</span>
                </div>
                <p className="text-sm text-slate-400">
                  {formatCurrency(plan.price_monthly)}/month
                </p>
                {plan.count > 0 && (
                  <div className="mt-2 w-full bg-white/10 rounded-full h-2">
                    <div
                      className="bg-gradient-to-r from-emerald-600 to-emerald-400 h-2 rounded-full"
                      style={{ width: `${(plan.count / stats.total_subscriptions) * 100}%` }}
                    />
                  </div>
                )}
              </div>
            ))}
          </div>
        </Card>
      )}

      {/* Subscriptions Table */}
      <Card className="p-6 hover-card-effect">
        <h2 className="text-lg font-semibold text-slate-100 mb-4">All Subscriptions</h2>
        {loading ? (
          <div className="text-center py-12 text-slate-400">Loading...</div>
        ) : subscriptions.length === 0 ? (
          <div className="text-center py-12">
            <div className="w-16 h-16 mx-auto mb-4 rounded-full bg-slate-800/50 flex items-center justify-center">
              <CreditCard className="w-8 h-8 text-slate-500" />
            </div>
            <h3 className="text-lg font-medium text-slate-300 mb-2">No subscriptions yet</h3>
            <p className="text-slate-500 mb-4 max-w-md mx-auto">
              Create a subscription to start tracking revenue and manage customer billing.
            </p>
            <Button
              onClick={() => {
                loadFormData()
                setShowCreateModal(true)
              }}
              className="bg-gradient-to-r from-emerald-600 to-emerald-500"
            >
              <Plus className="w-4 h-4 mr-2" />
              Create First Subscription
            </Button>
          </div>
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
                        <div>
                          <span className="text-slate-200">{sub.company_name}</span>
                          {sub.owner_email && (
                            <p className="text-xs text-slate-500">{sub.owner_email}</p>
                          )}
                        </div>
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
                        {formatDateShort(sub.current_period_start)} - {formatDateShort(sub.current_period_end)}
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

      {/* Create Modal */}
      {showCreateModal && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/80 backdrop-blur-sm">
          <Card className="w-full max-w-md p-6 bg-black/90 border-white/20">
            <h2 className="text-xl font-bold text-slate-100 mb-4">Create Subscription</h2>
            <div className="space-y-4">
              <div>
                <label className="text-sm text-slate-300 mb-2 block">Company</label>
                <select
                  value={newSubscription.company_id}
                  onChange={(e) => setNewSubscription(prev => ({ ...prev, company_id: e.target.value }))}
                  className="w-full px-3 py-2 bg-white/5 border border-white/10 rounded-lg text-slate-200"
                >
                  <option value="">Select a company...</option>
                  {companies.map(c => (
                    <option key={c.id} value={c.id}>{c.name}</option>
                  ))}
                </select>
              </div>
              <div>
                <label className="text-sm text-slate-300 mb-2 block">Plan</label>
                <select
                  value={newSubscription.plan_id}
                  onChange={(e) => setNewSubscription(prev => ({ ...prev, plan_id: e.target.value }))}
                  className="w-full px-3 py-2 bg-white/5 border border-white/10 rounded-lg text-slate-200"
                >
                  <option value="">Select a plan...</option>
                  {plans.map(p => (
                    <option key={p.id} value={p.id}>
                      {p.display_name} - {formatCurrency(p.price_monthly)}/month
                    </option>
                  ))}
                </select>
              </div>
              <div className="flex gap-2 pt-4">
                <Button
                  onClick={handleCreateSubscription}
                  disabled={creating}
                  className="flex-1 bg-gradient-to-r from-emerald-600 to-emerald-500"
                >
                  {creating ? 'Creating...' : 'Create Subscription'}
                </Button>
                <Button variant="outline" onClick={() => setShowCreateModal(false)}>Cancel</Button>
              </div>
            </div>
          </Card>
        </div>
      )}
    </div>
  )
}
