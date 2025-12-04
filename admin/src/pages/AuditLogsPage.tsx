import { useEffect, useState } from 'react'
import { Button } from '@/components/ui/button'
import { Card } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { api } from '@/lib/api'
import { toast } from '@/components/ui/toaster'
import { getRelativeTime } from '@/lib/utils'
import { cn } from '@/lib/utils'

interface AuditLog {
  id: number
  user_id?: string
  company_id?: string
  action: string
  resource_type?: string
  resource_id?: string
  ip_address?: string
  user_agent?: string
  metadata?: Record<string, unknown>
  created_at: string
}

export function AuditLogsPage() {
  const [logs, setLogs] = useState<AuditLog[]>([])
  const [loading, setLoading] = useState(true)
  const [page, setPage] = useState(1)
  const [total, setTotal] = useState(0)
  const [filters, setFilters] = useState({ action: '', resource_type: '', user_id: '' })

  useEffect(() => {
    loadLogs()
  }, [page, filters])

  async function loadLogs() {
    try {
      setLoading(true)
      const response = await api.admin.auditLogs.list(page, 50, filters)
      setLogs(response.logs ?? [])
      setTotal(response.pagination?.total ?? (response.logs?.length ?? 0))
    } catch (error) {
      toast({
        title: 'Error',
        description: error instanceof Error ? error.message : 'Failed to load audit logs',
        variant: 'destructive',
      })
    } finally {
      setLoading(false)
    }
  }

  const getActionColor = (action: string) => {
    if (action.includes('created')) return 'text-emerald-400'
    if (action.includes('updated')) return 'text-blue-400'
    if (action.includes('deleted')) return 'text-red-400'
    return 'text-slate-400'
  }

  return (
    <div className="space-y-6 animate-fade-in-up">
      <div>
        <h1 className="text-3xl font-display font-bold text-slate-100 mb-2">Audit Logs</h1>
        <p className="text-slate-400">Riwayat aktivitas sistem dan perubahan data</p>
      </div>

      {/* Filters */}
      <Card className="p-4 hover-card-effect">
        <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
          <div>
            <label className="text-sm text-slate-300 mb-2 block">Action</label>
            <Input
              placeholder="Filter by action..."
              value={filters.action}
              onChange={(e) => setFilters({ ...filters, action: e.target.value })}
              className="bg-white/5 border-white/10"
            />
          </div>
          <div>
            <label className="text-sm text-slate-300 mb-2 block">Resource Type</label>
            <Input
              placeholder="Filter by resource..."
              value={filters.resource_type}
              onChange={(e) => setFilters({ ...filters, resource_type: e.target.value })}
              className="bg-white/5 border-white/10"
            />
          </div>
          <div>
            <label className="text-sm text-slate-300 mb-2 block">User ID</label>
            <Input
              placeholder="Filter by user..."
              value={filters.user_id}
              onChange={(e) => setFilters({ ...filters, user_id: e.target.value })}
              className="bg-white/5 border-white/10"
            />
          </div>
        </div>
      </Card>

      <Card className="p-6 hover-card-effect">
        {loading ? (
          <div className="text-center py-12 text-slate-400">Loading...</div>
        ) : logs.length === 0 ? (
          <div className="text-center py-12 text-slate-500">Belum ada audit log.</div>
        ) : (
          <div className="space-y-4">
            {logs.map((log) => (
              <div
                key={log.id}
                className="p-4 rounded-lg bg-white/5 border border-white/10 hover:bg-white/10 transition-colors"
              >
                <div className="flex items-start justify-between mb-2">
                  <div className="flex-1">
                    <div className="flex items-center gap-2 mb-1">
                      <span className={cn('text-sm font-semibold', getActionColor(log.action))}>
                        {log.action}
                      </span>
                      {log.resource_type && (
                        <span className="text-xs px-2 py-0.5 rounded bg-white/5 text-slate-400">
                          {log.resource_type}
                        </span>
                      )}
                    </div>
                    <div className="text-xs text-slate-500 space-y-1">
                      {log.user_id && <div>User: {log.user_id}</div>}
                      {log.company_id && <div>Company: {log.company_id}</div>}
                      {log.ip_address && <div>IP: {log.ip_address}</div>}
                    </div>
                    {log.metadata && Object.keys(log.metadata).length > 0 && (
                      <div className="mt-2 text-xs text-slate-400 bg-black/20 p-2 rounded">
                        <pre className="whitespace-pre-wrap">{JSON.stringify(log.metadata, null, 2)}</pre>
                      </div>
                    )}
                  </div>
                  <div className="text-xs text-slate-500 ml-4">
                    {getRelativeTime(log.created_at)}
                  </div>
                </div>
              </div>
            ))}
          </div>
        )}

        {total > 50 && (
          <div className="flex items-center justify-between mt-4 pt-4 border-t border-white/10">
            <p className="text-sm text-slate-400">
              Showing {(page - 1) * 50 + 1} to {Math.min(page * 50, total)} of {total} logs
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
                disabled={page * 50 >= total}
              >
                Next
              </Button>
            </div>
          </div>
        )}
      </Card>
    </div>
  )
}

