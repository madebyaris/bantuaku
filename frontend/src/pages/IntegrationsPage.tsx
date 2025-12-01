import { useEffect, useState } from 'react'
import { Link2, Check, AlertCircle, Loader2, RefreshCw } from 'lucide-react'
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { api, WooCommerceSyncStatus } from '@/lib/api'
import { toast } from '@/components/ui/toaster'
import { cn, getRelativeTime } from '@/lib/utils'

export function IntegrationsPage() {
  const [wooStatus, setWooStatus] = useState<WooCommerceSyncStatus | null>(null)
  const [loading, setLoading] = useState(true)
  const [showConnectForm, setShowConnectForm] = useState(false)

  useEffect(() => {
    loadStatus()
  }, [])

  async function loadStatus() {
    try {
      const status = await api.integrations.woocommerce.status()
      setWooStatus(status)
    } catch {
      // Ignore
    } finally {
      setLoading(false)
    }
  }

  const isConnected = wooStatus?.status === 'connected'

  return (
    <div className="max-w-3xl mx-auto space-y-6 animate-fade-in">
      {/* WooCommerce */}
      <Card>
        <CardHeader>
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-3">
              <div className="p-2 bg-purple-100 rounded-lg">
                <svg
                  className="w-8 h-8 text-purple-600"
                  viewBox="0 0 24 24"
                  fill="currentColor"
                >
                  <path d="M2.514 4.6C1.614 4.6 1 5.214 1 6.114v8.4c0 .9.614 1.514 1.514 1.514h18.972c.9 0 1.514-.614 1.514-1.514v-8.4c0-.9-.614-1.514-1.514-1.514H2.514zM2.4 6.4h19.2v7.714H2.4V6.4zm3.886 1.286c-.686 0-1.029.343-1.029 1.028v4.114c0 .686.343 1.029 1.029 1.029s1.028-.343 1.028-1.029V8.714c0-.685-.343-1.028-1.028-1.028zm4.114 0c-.686 0-1.029.343-1.029 1.028v4.114c0 .686.343 1.029 1.029 1.029s1.029-.343 1.029-1.029V8.714c0-.685-.343-1.028-1.029-1.028zm4.114 0c-.685 0-1.028.343-1.028 1.028v4.114c0 .686.343 1.029 1.028 1.029.686 0 1.029-.343 1.029-1.029V8.714c0-.685-.343-1.028-1.029-1.028zM1.714 17.743v1.543c0 .9.614 1.514 1.514 1.514h17.544c.9 0 1.514-.614 1.514-1.514v-1.543H1.714z" />
                </svg>
              </div>
              <div>
                <CardTitle className="text-lg">WooCommerce</CardTitle>
                <CardDescription>
                  Sinkronkan produk dan pesanan dari toko WooCommerce
                </CardDescription>
              </div>
            </div>
            <div
              className={cn(
                'px-3 py-1 rounded-full text-sm font-medium',
                isConnected
                  ? 'bg-green-100 text-green-700'
                  : 'bg-slate-100 text-slate-600'
              )}
            >
              {isConnected ? 'Terhubung' : 'Tidak Terhubung'}
            </div>
          </div>
        </CardHeader>
        <CardContent>
          {loading ? (
            <div className="flex justify-center py-8">
              <Loader2 className="w-6 h-6 animate-spin text-purple-600" />
            </div>
          ) : isConnected ? (
            <ConnectedStatus status={wooStatus!} onRefresh={loadStatus} />
          ) : showConnectForm ? (
            <ConnectForm
              onCancel={() => setShowConnectForm(false)}
              onSuccess={() => {
                setShowConnectForm(false)
                loadStatus()
              }}
            />
          ) : (
            <div className="text-center py-6">
              <p className="text-slate-500 mb-4">
                Hubungkan toko WooCommerce Anda untuk sinkronisasi otomatis
              </p>
              <Button onClick={() => setShowConnectForm(true)}>
                <Link2 className="w-4 h-4 mr-2" />
                Hubungkan WooCommerce
              </Button>
            </div>
          )}
        </CardContent>
      </Card>

      {/* Coming Soon */}
      <Card className="opacity-60">
        <CardHeader>
          <div className="flex items-center gap-3">
            <div className="p-2 bg-orange-100 rounded-lg">
              <span className="text-2xl">🛒</span>
            </div>
            <div>
              <CardTitle className="text-lg">Shopee & Tokopedia</CardTitle>
              <CardDescription>Segera hadir di versi berikutnya</CardDescription>
            </div>
          </div>
        </CardHeader>
      </Card>
    </div>
  )
}

function ConnectedStatus({
  status,
  onRefresh,
}: {
  status: WooCommerceSyncStatus
  onRefresh: () => void
}) {
  const [syncing, setSyncing] = useState(false)

  async function handleSync() {
    setSyncing(true)
    try {
      const result = await api.integrations.woocommerce.syncNow()
      toast({
        title: 'Sinkronisasi berhasil!',
        description: `${result.products_synced} produk, ${result.orders_synced} pesanan`,
        variant: 'success',
      })
      onRefresh()
    } catch (err) {
      toast({
        title: 'Gagal sinkronisasi',
        description: err instanceof Error ? err.message : 'Terjadi kesalahan',
        variant: 'destructive',
      })
    } finally {
      setSyncing(false)
    }
  }

  return (
    <div className="space-y-4">
      {status.error_message && (
        <div className="flex items-center gap-2 p-3 bg-red-50 text-red-700 rounded-lg">
          <AlertCircle className="w-5 h-5" />
          <span>{status.error_message}</span>
        </div>
      )}

      <div className="grid grid-cols-3 gap-4">
        <div className="p-4 bg-slate-50 rounded-lg text-center">
          <p className="text-2xl font-bold text-slate-900">{status.product_count}</p>
          <p className="text-sm text-slate-500">Produk</p>
        </div>
        <div className="p-4 bg-slate-50 rounded-lg text-center">
          <p className="text-2xl font-bold text-slate-900">{status.order_count}</p>
          <p className="text-sm text-slate-500">Pesanan</p>
        </div>
        <div className="p-4 bg-slate-50 rounded-lg text-center">
          <p className="text-sm font-medium text-slate-900">
            {status.last_sync ? getRelativeTime(status.last_sync) : 'Belum pernah'}
          </p>
          <p className="text-sm text-slate-500">Sinkron Terakhir</p>
        </div>
      </div>

      <Button onClick={handleSync} disabled={syncing} className="w-full">
        {syncing ? (
          <>
            <Loader2 className="w-4 h-4 mr-2 animate-spin" />
            Menyinkronkan...
          </>
        ) : (
          <>
            <RefreshCw className="w-4 h-4 mr-2" />
            Sinkronkan Sekarang
          </>
        )}
      </Button>
    </div>
  )
}

function ConnectForm({
  onCancel,
  onSuccess,
}: {
  onCancel: () => void
  onSuccess: () => void
}) {
  const [loading, setLoading] = useState(false)
  const [form, setForm] = useState({
    store_url: '',
    consumer_key: '',
    consumer_secret: '',
  })

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    setLoading(true)

    try {
      await api.integrations.woocommerce.connect(
        form.store_url,
        form.consumer_key,
        form.consumer_secret
      )
      toast({ title: 'WooCommerce terhubung!', variant: 'success' })
      onSuccess()
    } catch (err) {
      toast({
        title: 'Gagal menghubungkan',
        description: err instanceof Error ? err.message : 'Periksa kredensial Anda',
        variant: 'destructive',
      })
    } finally {
      setLoading(false)
    }
  }

  return (
    <form onSubmit={handleSubmit} className="space-y-4">
      <div className="space-y-2">
        <Label>URL Toko</Label>
        <Input
          value={form.store_url}
          onChange={(e) => setForm({ ...form, store_url: e.target.value })}
          placeholder="https://tokoanda.com"
          required
        />
      </div>
      <div className="space-y-2">
        <Label>Consumer Key</Label>
        <Input
          value={form.consumer_key}
          onChange={(e) => setForm({ ...form, consumer_key: e.target.value })}
          placeholder="ck_xxxxxxxxxxxxxxxx"
          required
        />
      </div>
      <div className="space-y-2">
        <Label>Consumer Secret</Label>
        <Input
          type="password"
          value={form.consumer_secret}
          onChange={(e) => setForm({ ...form, consumer_secret: e.target.value })}
          placeholder="cs_xxxxxxxxxxxxxxxx"
          required
        />
      </div>
      <div className="p-3 bg-blue-50 rounded-lg text-sm text-blue-700">
        💡 Dapatkan API keys di WooCommerce → Settings → Advanced → REST API
      </div>
      <div className="flex gap-3">
        <Button type="button" variant="outline" className="flex-1" onClick={onCancel}>
          Batal
        </Button>
        <Button type="submit" className="flex-1" disabled={loading}>
          {loading ? <Loader2 className="w-4 h-4 animate-spin" /> : 'Hubungkan'}
        </Button>
      </div>
    </form>
  )
}
