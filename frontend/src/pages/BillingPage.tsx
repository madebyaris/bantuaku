import { useEffect, useState } from 'react'
import { CreditCard, CheckCircle, Loader2 } from 'lucide-react'
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { api, BillingPlan, BillingSubscription } from '@/lib/api'
import { formatCurrency } from '@/lib/utils'

export function BillingPage() {
  const [plans, setPlans] = useState<BillingPlan[]>([])
  const [subscription, setSubscription] = useState<BillingSubscription | null>(null)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [checkoutLoading, setCheckoutLoading] = useState<string | null>(null)

  useEffect(() => {
    loadData()
  }, [])

  async function loadData() {
    setLoading(true)
    setError(null)
    try {
      const [plansResp, subResp] = await Promise.allSettled([
        api.billing.plans(),
        api.billing.subscription(),
      ])

      if (plansResp.status === 'fulfilled') {
        setPlans(plansResp.value.plans || [])
      }
      if (subResp.status === 'fulfilled') {
        setSubscription(subResp.value)
      }
    } catch (err) {
      setError('Gagal memuat billing')
    } finally {
      setLoading(false)
    }
  }

  async function handleCheckout(planId: string) {
    setCheckoutLoading(planId)
    setError(null)
    try {
      const resp = await api.billing.checkout(planId, window.location.href, window.location.href)
      if (resp.url) {
        window.location.href = resp.url
      }
    } catch (err) {
      setError('Gagal memulai checkout')
    } finally {
      setCheckoutLoading(null)
    }
  }

  return (
    <div className="max-w-5xl mx-auto px-4 py-8 space-y-6 animate-fade-in-up">
      <div className="flex items-center justify-between border-b border-white/10 pb-4">
        <div>
          <div className="flex items-center gap-2 mb-2">
            <div className="p-2 bg-emerald-500/10 rounded-lg border border-emerald-500/20">
              <CreditCard className="w-5 h-5 text-emerald-400" />
            </div>
            <span className="text-sm font-medium text-emerald-400 uppercase tracking-wider">
              Billing & Subscription
            </span>
          </div>
          <h1 className="text-3xl font-display font-bold text-slate-100">Kelola Langganan</h1>
          <p className="text-slate-400 mt-1 text-sm">
            Pilih paket dan kelola langganan Stripe.
          </p>
        </div>
      </div>

      {error && (
        <Card className="border-red-500/20 bg-red-500/5">
          <CardContent className="py-3 text-sm text-red-200">{error}</CardContent>
        </Card>
      )}

      {subscription && (
        <Card className="border-white/10 bg-white/5">
          <CardHeader>
            <CardTitle className="text-lg text-slate-100 flex items-center gap-2">
              <CheckCircle className="w-5 h-5 text-emerald-400" />
              Langganan Aktif
            </CardTitle>
            <CardDescription className="text-slate-400">
              Status: {subscription.status} • Plan: {subscription.plan_id}
            </CardDescription>
          </CardHeader>
        </Card>
      )}

      {loading ? (
        <div className="flex justify-center py-12">
          <Loader2 className="w-8 h-8 animate-spin text-emerald-400" />
        </div>
      ) : (
        <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
          {plans.map((plan) => (
            <Card key={plan.id} className="border-white/10 bg-white/5 flex flex-col">
              <CardHeader>
                <div className="flex items-center justify-between">
                  <CardTitle className="text-lg text-slate-100">{plan.display_name}</CardTitle>
                  <Badge variant="outline" className="text-slate-300 border-white/10">
                    {plan.currency.toUpperCase()}
                  </Badge>
                </div>
                <CardDescription className="text-slate-400">
                  {formatCurrency(plan.price_monthly)} / bulan
                </CardDescription>
              </CardHeader>
              <CardContent className="flex-1 space-y-2">
                {plan.features && (
                  <ul className="text-sm text-slate-300 space-y-1">
                    {Object.keys(plan.features).map((key) => (
                      <li key={key}>• {key}</li>
                    ))}
                  </ul>
                )}
                <Button
                  className="w-full mt-4"
                  onClick={() => handleCheckout(plan.id)}
                  disabled={checkoutLoading === plan.id}
                >
                  {checkoutLoading === plan.id ? (
                    <Loader2 className="w-4 h-4 animate-spin" />
                  ) : (
                    'Pilih Plan'
                  )}
                </Button>
              </CardContent>
            </Card>
          ))}
        </div>
      )}
    </div>
  )
}
