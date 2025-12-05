import { useEffect, useState } from 'react'
import { Loader2, Crown } from 'lucide-react'
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { api, BillingPlan } from '@/lib/api'
import { formatCurrency } from '@/lib/utils'

export function PricingPage() {
  const [plans, setPlans] = useState<BillingPlan[]>([])
  const [loading, setLoading] = useState(false)
  const [checkoutLoading, setCheckoutLoading] = useState<string | null>(null)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    loadPlans()
  }, [])

  async function loadPlans() {
    setLoading(true)
    setError(null)
    try {
      const res = await api.billing.plans()
      setPlans(res.plans || [])
    } catch (err) {
      setError('Gagal memuat paket')
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
    <div className="max-w-6xl mx-auto px-4 py-10 space-y-6 animate-fade-in-up">
      <div className="text-center space-y-2">
        <div className="flex items-center justify-center gap-2 text-emerald-400">
          <Crown className="w-5 h-5" />
          <span className="text-sm uppercase tracking-wide font-semibold">Pilih Paket</span>
        </div>
        <h1 className="text-3xl font-display font-bold text-slate-100">Harga & Paket</h1>
        <p className="text-slate-400 text-sm">
          Mulai dari paket Free hingga Pro, lengkap dengan limit yang jelas.
        </p>
      </div>

      {error && (
        <Card className="border-red-500/20 bg-red-500/5">
          <CardContent className="py-3 text-sm text-red-200 text-center">{error}</CardContent>
        </Card>
      )}

      {loading ? (
        <div className="flex justify-center py-12">
          <Loader2 className="w-8 h-8 animate-spin text-emerald-400" />
        </div>
      ) : (
        <div className="grid grid-cols-1 md:grid-cols-3 gap-5">
          {plans.map((plan) => (
            <Card key={plan.id} className="border-white/10 bg-white/5 flex flex-col">
              <CardHeader>
                <div className="flex items-center justify-between">
                  <CardTitle className="text-xl text-slate-100">{plan.display_name}</CardTitle>
                  <Badge variant="outline" className="text-slate-300 border-white/10">
                    {plan.currency.toUpperCase()}
                  </Badge>
                </div>
                <CardDescription className="text-slate-400 text-lg">
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
                    'Pilih Paket'
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
