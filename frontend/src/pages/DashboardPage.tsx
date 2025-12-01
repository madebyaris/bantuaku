import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import {
  Package,
  AlertTriangle,
  TrendingUp,
  TrendingDown,
  Target,
  ArrowRight,
  Loader2,
  MessageSquareText,
  FileInput,
} from 'lucide-react'
import {
  LineChart,
  Line,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  ResponsiveContainer,
} from 'recharts'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { api, DashboardSummary, Recommendation, MarketTrend } from '@/lib/api'
import { formatCurrency, formatPercentage, cn, getRiskColor } from '@/lib/utils'

export function DashboardPage() {
  const [loading, setLoading] = useState(true)
  const [summary, setSummary] = useState<DashboardSummary | null>(null)
  const [recommendations, setRecommendations] = useState<Recommendation[]>([])
  const [trends, setTrends] = useState<MarketTrend[]>([])

  useEffect(() => {
    async function loadData() {
      try {
        const [summaryData, recsData, trendsData] = await Promise.all([
          api.dashboard.summary(),
          api.recommendations.list(),
          api.market.trends(),
        ])
        setSummary(summaryData)
        setRecommendations(recsData.slice(0, 5))
        setTrends(trendsData.slice(0, 5))
      } catch (err) {
        console.error('Failed to load dashboard:', err)
      } finally {
        setLoading(false)
      }
    }
    loadData()
  }, [])

  if (loading) {
    return (
      <div className="flex items-center justify-center h-96">
        <Loader2 className="w-8 h-8 animate-spin text-purple-600" />
      </div>
    )
  }

  // Mock chart data
  const chartData = [
    { name: 'Sen', sales: 4500000 },
    { name: 'Sel', sales: 5200000 },
    { name: 'Rab', sales: 4800000 },
    { name: 'Kam', sales: 6100000 },
    { name: 'Jum', sales: 7200000 },
    { name: 'Sab', sales: 8500000 },
    { name: 'Min', sales: 6800000 },
  ]

  return (
    <div className="space-y-6 animate-fade-in">
      {/* KPI Cards */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
        <Card className="card-hover">
          <CardContent className="pt-6">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm font-medium text-slate-500">Total Produk</p>
                <p className="text-3xl font-bold text-slate-900 mt-1">
                  {summary?.total_products || 0}
                </p>
              </div>
              <div className="p-3 bg-purple-100 rounded-xl">
                <Package className="w-6 h-6 text-purple-600" />
              </div>
            </div>
          </CardContent>
        </Card>


        <Card className="card-hover">
          <CardContent className="pt-6">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm font-medium text-slate-500">Revenue Bulan Ini</p>
                <p className="text-2xl font-bold text-slate-900 mt-1">
                  {formatCurrency(summary?.revenue_this_month || 0)}
                </p>
                <div className={cn(
                  'flex items-center gap-1 mt-1 text-xs font-medium',
                  (summary?.revenue_trend || 0) >= 0 ? 'text-green-600' : 'text-red-600'
                )}>
                  {(summary?.revenue_trend || 0) >= 0 ? (
                    <TrendingUp className="w-3 h-3" />
                  ) : (
                    <TrendingDown className="w-3 h-3" />
                  )}
                  {formatPercentage(summary?.revenue_trend || 0)} dari bulan lalu
                </div>
              </div>
              <div className="p-3 bg-green-100 rounded-xl">
                <TrendingUp className="w-6 h-6 text-green-600" />
              </div>
            </div>
          </CardContent>
        </Card>

        <Card className="card-hover">
          <CardContent className="pt-6">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm font-medium text-slate-500">Akurasi Forecast</p>
                <p className="text-3xl font-bold text-slate-900 mt-1">
                  {(summary?.forecast_accuracy || 0).toFixed(1)}%
                </p>
                <p className="text-xs text-slate-500 mt-1">30 hari terakhir</p>
              </div>
              <div className="p-3 bg-blue-100 rounded-xl">
                <Target className="w-6 h-6 text-blue-600" />
              </div>
            </div>
          </CardContent>
        </Card>
      </div>

      {/* Main content grid */}
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        {/* Sales Chart */}
        <Card className="lg:col-span-2">
          <CardHeader className="flex flex-row items-center justify-between">
            <CardTitle className="text-lg">Penjualan Minggu Ini</CardTitle>
            <Link to="/products">
              <Button variant="ghost" size="sm">
                Lihat Detail <ArrowRight className="w-4 h-4 ml-1" />
              </Button>
            </Link>
          </CardHeader>
          <CardContent>
            <div className="h-80">
              <ResponsiveContainer width="100%" height="100%">
                <LineChart data={chartData}>
                  <CartesianGrid strokeDasharray="3 3" stroke="#e2e8f0" />
                  <XAxis dataKey="name" stroke="#64748b" fontSize={12} />
                  <YAxis
                    stroke="#64748b"
                    fontSize={12}
                    tickFormatter={(value) => `${(value / 1000000).toFixed(1)}jt`}
                  />
                  <Tooltip
                    formatter={(value: number) => [formatCurrency(value), 'Penjualan']}
                    contentStyle={{
                      backgroundColor: 'white',
                      border: '1px solid #e2e8f0',
                      borderRadius: '8px',
                    }}
                  />
                  <Line
                    type="monotone"
                    dataKey="sales"
                    stroke="#7c3aed"
                    strokeWidth={3}
                    dot={{ fill: '#7c3aed', strokeWidth: 2 }}
                    activeDot={{ r: 6, fill: '#7c3aed' }}
                  />
                </LineChart>
              </ResponsiveContainer>
            </div>
          </CardContent>
        </Card>

        {/* Quick Actions */}
        <Card>
          <CardHeader>
            <CardTitle className="text-lg">Aksi Cepat</CardTitle>
          </CardHeader>
          <CardContent className="space-y-3">
            <Link to="/data-input">
              <Button variant="outline" className="w-full justify-start">
                <FileInput className="w-4 h-4 mr-2" />
                Input Data Penjualan
              </Button>
            </Link>
            <Link to="/ai-chat">
              <Button className="w-full justify-start bg-gradient-to-r from-purple-600 to-indigo-600 hover:from-purple-700 hover:to-indigo-700">
                <MessageSquareText className="w-4 h-4 mr-2" />
                Tanya AI Assistant
              </Button>
            </Link>
            <Link to="/integrations">
              <Button variant="outline" className="w-full justify-start">
                <Package className="w-4 h-4 mr-2" />
                Hubungkan Toko Online
              </Button>
            </Link>
          </CardContent>
        </Card>
      </div>

      {/* Recommendations & Trends */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* Recommendations */}
        <Card>
          <CardHeader className="flex flex-row items-center justify-between">
            <CardTitle className="text-lg">Rekomendasi Restok</CardTitle>
            <Link to="/products">
              <Button variant="ghost" size="sm">
                Semua <ArrowRight className="w-4 h-4 ml-1" />
              </Button>
            </Link>
          </CardHeader>
          <CardContent>
            {recommendations.length === 0 ? (
              <p className="text-sm text-slate-500 text-center py-8">
                Belum ada rekomendasi. Tambahkan data penjualan terlebih dahulu.
              </p>
            ) : (
              <div className="space-y-3">
                {recommendations.map((rec) => (
                  <div
                    key={rec.product_id}
                    className="flex items-center justify-between p-3 rounded-lg bg-slate-50 hover:bg-slate-100 transition-colors"
                  >
                    <div className="flex-1 min-w-0">
                      <p className="font-medium text-slate-900 truncate">
                        {rec.product_name}
                      </p>
                      <p className="text-sm text-slate-500">
                        Proyeksi: {rec.projected_demand} unit
                      </p>
                    </div>
                    <div className="flex items-center gap-3">
                      <span
                        className={cn(
                          'px-2 py-1 rounded-full text-xs font-medium',
                          getRiskColor(rec.risk_level)
                        )}
                      >
                        {rec.risk_level === 'high'
                          ? 'Urgent'
                          : rec.risk_level === 'medium'
                          ? 'Segera'
                          : 'Aman'}
                      </span>
                      {rec.recommended_qty > 0 && (
                        <span className="text-sm font-semibold text-purple-600">
                          +{rec.recommended_qty}
                        </span>
                      )}
                    </div>
                  </div>
                ))}
              </div>
            )}
          </CardContent>
        </Card>

        {/* Market Trends */}
        <Card>
          <CardHeader>
            <CardTitle className="text-lg">Trend Pasar</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="space-y-3">
              {trends.map((trend, i) => (
                <div
                  key={i}
                  className="flex items-center justify-between p-3 rounded-lg bg-slate-50"
                >
                  <div className="flex-1">
                    <p className="font-medium text-slate-900">{trend.name}</p>
                    <p className="text-sm text-slate-500">{trend.category}</p>
                  </div>
                  <div className="text-right">
                    <div className="flex items-center gap-1 text-green-600 font-medium">
                      <TrendingUp className="w-4 h-4" />
                      +{trend.growth_rate.toFixed(1)}%
                    </div>
                    <p className="text-xs text-slate-500">{trend.source}</p>
                  </div>
                </div>
              ))}
            </div>
          </CardContent>
        </Card>
      </div>
    </div>
  )
}
