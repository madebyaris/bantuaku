import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import {
  TrendingUp,
  TrendingDown,
  ArrowRight,
  Loader2,
  MessageSquareText,
  FileText,
  Sparkles,
  BarChart3,
  Globe,
  Megaphone,
  Scale,
  Upload,
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
import { api, DashboardSummary, Sale } from '@/lib/api'
import { formatCurrency, formatPercentage, cn } from '@/lib/utils'

export function DashboardPage() {
  const [loading, setLoading] = useState(true)
  const [summary, setSummary] = useState<DashboardSummary | null>(null)
  const [chartData, setChartData] = useState<Array<{ name: string; sales: number }>>([])

  useEffect(() => {
    async function loadData() {
      try {
        const [summaryData, salesData] = await Promise.all([
          api.dashboard.summary(),
          api.sales.list(),
        ])
        setSummary(summaryData)
        
        // Process sales data for chart (last 7 days)
        const last7Days = processSalesForChart(salesData)
        setChartData(last7Days)
      } catch (err) {
        console.error('Failed to load dashboard:', err)
      } finally {
        setLoading(false)
      }
    }
    loadData()
  }, [])

  function processSalesForChart(sales: Sale[]): Array<{ name: string; sales: number }> {
    // Get last 7 days
    const days = ['Min', 'Sen', 'Sel', 'Rab', 'Kam', 'Jum', 'Sab']
    const today = new Date()
    today.setHours(0, 0, 0, 0) // Start of today
    const sevenDaysAgo = new Date(today)
    sevenDaysAgo.setDate(sevenDaysAgo.getDate() - 6) // Include today, so 6 days back
    
    const last7Days: Date[] = []
    for (let i = 0; i < 7; i++) {
      const date = new Date(sevenDaysAgo)
      date.setDate(date.getDate() + i)
      last7Days.push(date)
    }

    // Aggregate sales by day (only include sales from last 7 days)
    const dailySales: { [key: string]: number } = {}
    last7Days.forEach(date => {
      const dateStr = date.toISOString().split('T')[0]
      dailySales[dateStr] = 0
    })

    sales.forEach(sale => {
      const saleDate = new Date(sale.sale_date)
      saleDate.setHours(0, 0, 0, 0)
      const dateStr = saleDate.toISOString().split('T')[0]
      
      // Only include sales from the last 7 days
      if (saleDate >= sevenDaysAgo && saleDate <= today && dailySales.hasOwnProperty(dateStr)) {
        dailySales[dateStr] += sale.quantity * sale.price
      }
    })

    // Convert to chart format
    return last7Days.map((date) => {
      const dateStr = date.toISOString().split('T')[0]
      const dayName = days[date.getDay()]
      return {
        name: dayName,
        sales: dailySales[dateStr] || 0,
      }
    })
  }

  if (loading) {
    return (
      <div className="flex items-center justify-center h-96">
        <Loader2 className="w-8 h-8 animate-spin text-emerald-400" />
      </div>
    )
  }

  return (
    <div className="space-y-6 animate-fade-in-up">
      {/* Company Profile Card */}
      {summary?.company_name && (
        <Card className="hover-card-effect border-white/10 bg-white/5">
          <CardContent className="pt-6">
            <div className="flex items-start justify-between">
              <div>
                <h2 className="text-2xl font-bold text-slate-100">{summary.company_name}</h2>
                {summary.company_industry && (
                  <p className="text-sm text-slate-400 mt-1">{summary.company_industry}</p>
                )}
                {summary.company_location && (
                  <p className="text-sm text-slate-400">{summary.company_location}</p>
                )}
              </div>
              <div className="p-3 bg-emerald-500/10 rounded-xl border border-emerald-500/20 shadow-[0_0_15px_rgba(16,185,129,0.2)]">
                <Sparkles className="w-6 h-6 text-emerald-400" />
              </div>
            </div>
          </CardContent>
        </Card>
      )}

      {/* KPI Cards */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
        <Card className="hover-card-effect border-white/10 bg-white/5">
          <CardContent className="pt-6">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm font-medium text-slate-400">Revenue Bulan Ini</p>
                <p className="text-2xl font-bold text-slate-100 mt-1">
                  {formatCurrency(summary?.revenue_this_month || 0)}
                </p>
                <div className={cn(
                  'flex items-center gap-1 mt-1 text-xs font-medium',
                  (summary?.revenue_trend || 0) >= 0 ? 'text-emerald-400' : 'text-red-400'
                )}>
                  {(summary?.revenue_trend || 0) >= 0 ? (
                    <TrendingUp className="w-3 h-3" />
                  ) : (
                    <TrendingDown className="w-3 h-3" />
                  )}
                  {formatPercentage(summary?.revenue_trend || 0)} dari bulan lalu
                </div>
              </div>
              <div className="p-3 bg-emerald-500/10 rounded-xl border border-emerald-500/20">
                <TrendingUp className="w-6 h-6 text-emerald-400" />
              </div>
            </div>
          </CardContent>
        </Card>

        <Card className="hover-card-effect border-white/10 bg-white/5">
          <CardContent className="pt-6">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm font-medium text-slate-400">Total Percakapan</p>
                <p className="text-3xl font-bold text-slate-100 mt-1">
                  {summary?.total_conversations || 0}
                </p>
                <p className="text-xs text-slate-500 mt-1">Dengan AI Assistant</p>
              </div>
              <div className="p-3 bg-blue-500/10 rounded-xl border border-blue-500/20">
                <MessageSquareText className="w-6 h-6 text-blue-400" />
              </div>
            </div>
          </CardContent>
        </Card>

        <Card className="hover-card-effect border-white/10 bg-white/5">
          <CardContent className="pt-6">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm font-medium text-slate-400">Total Insights</p>
                <p className="text-3xl font-bold text-slate-100 mt-1">
                  {summary?.total_insights || 0}
                </p>
                <p className="text-xs text-slate-500 mt-1">Forecast, Market, etc.</p>
              </div>
              <div className="p-3 bg-purple-500/10 rounded-xl border border-purple-500/20">
                <BarChart3 className="w-6 h-6 text-purple-400" />
              </div>
            </div>
          </CardContent>
        </Card>

        <Card className="hover-card-effect border-white/10 bg-white/5">
          <CardContent className="pt-6">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm font-medium text-slate-400">File Uploads</p>
                <p className="text-3xl font-bold text-slate-100 mt-1">
                  {summary?.total_file_uploads || 0}
                </p>
                <p className="text-xs text-slate-500 mt-1">CSV, XLSX, PDF</p>
              </div>
              <div className="p-3 bg-orange-500/10 rounded-xl border border-orange-500/20">
                <Upload className="w-6 h-6 text-orange-400" />
              </div>
            </div>
          </CardContent>
        </Card>
      </div>

      {/* Main content grid */}
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        {/* Sales Chart */}
        <Card className="lg:col-span-2 hover-card-effect border-white/10 bg-white/5">
          <CardHeader className="flex flex-row items-center justify-between border-b border-white/5 pb-4">
            <CardTitle className="text-lg text-slate-100">Penjualan Minggu Ini</CardTitle>
            <Link to="/forecast">
              <Button variant="ghost" size="sm" className="text-slate-400 hover:text-emerald-400 hover:bg-emerald-500/10">
                Lihat Forecast <ArrowRight className="w-4 h-4 ml-1" />
              </Button>
            </Link>
          </CardHeader>
          <CardContent className="pt-6">
            {chartData.length === 0 ? (
              <div className="h-80 flex items-center justify-center">
                <p className="text-sm text-slate-500">
                  Belum ada data penjualan. Mulai chat dengan AI Assistant untuk menginput data.
                </p>
              </div>
            ) : (
              <div className="h-80">
                <ResponsiveContainer width="100%" height="100%">
                  <LineChart data={chartData}>
                    <CartesianGrid strokeDasharray="3 3" stroke="#334155" opacity={0.3} />
                    <XAxis dataKey="name" stroke="#94a3b8" fontSize={12} tickLine={false} axisLine={false} />
                    <YAxis
                      stroke="#94a3b8"
                      fontSize={12}
                      tickFormatter={(value) => `${(value / 1000000).toFixed(1)}jt`}
                      tickLine={false}
                      axisLine={false}
                    />
                    <Tooltip
                      formatter={(value: number) => [formatCurrency(value), 'Penjualan']}
                      contentStyle={{
                        backgroundColor: '#0f172a',
                        border: '1px solid rgba(255,255,255,0.1)',
                        borderRadius: '8px',
                        color: '#f8fafc',
                      }}
                      itemStyle={{ color: '#34d399' }}
                    />
                    <Line
                      type="monotone"
                      dataKey="sales"
                      stroke="#10b981"
                      strokeWidth={3}
                      dot={{ fill: '#059669', strokeWidth: 2, r: 4 }}
                      activeDot={{ r: 6, fill: '#34d399', stroke: '#059669' }}
                    />
                  </LineChart>
                </ResponsiveContainer>
              </div>
            )}
          </CardContent>
        </Card>

        {/* Quick Actions */}
        <Card className="hover-card-effect border-white/10 bg-white/5">
          <CardHeader className="border-b border-white/5 pb-4">
            <CardTitle className="text-lg text-slate-100">Aksi Cepat</CardTitle>
          </CardHeader>
          <CardContent className="space-y-6 pt-6">
            <Link to="/ai-chat" className="block">
              <Button className="w-full h-12 justify-start bg-gradient-to-r from-emerald-600 to-emerald-400 hover:from-emerald-500 hover:to-emerald-300 text-black font-semibold shadow-[0_0_20px_rgba(16,185,129,0.3)] border-0">
                <MessageSquareText className="w-5 h-5 mr-3" />
                Mulai Chat dengan AI
              </Button>
            </Link>
            <Link to="/forecast" className="block">
              <Button variant="outline" className="w-full h-12 justify-start border-white/10 bg-white/5 text-slate-300 hover:bg-white/10 hover:text-emerald-400 hover:border-emerald-500/30">
                <BarChart3 className="w-5 h-5 mr-3" />
                Lihat Forecast
              </Button>
            </Link>
            <Link to="/market-prediction" className="block">
              <Button variant="outline" className="w-full h-12 justify-start border-white/10 bg-white/5 text-slate-300 hover:bg-white/10 hover:text-blue-400 hover:border-blue-500/30">
                <Globe className="w-5 h-5 mr-3" />
                Prediksi Pasar
              </Button>
            </Link>
            <Link to="/marketing" className="block">
              <Button variant="outline" className="w-full h-12 justify-start border-white/10 bg-white/5 text-slate-300 hover:bg-white/10 hover:text-purple-400 hover:border-purple-500/30">
                <Megaphone className="w-5 h-5 mr-3" />
                Rekomendasi Marketing
              </Button>
            </Link>
            <Link to="/regulation" className="block">
              <Button variant="outline" className="w-full h-12 justify-start border-white/10 bg-white/5 text-slate-300 hover:bg-white/10 hover:text-orange-400 hover:border-orange-500/30">
                <Scale className="w-5 h-5 mr-3" />
                Peraturan Pemerintah
              </Button>
            </Link>
          </CardContent>
        </Card>
      </div>

      {/* Recent Activity */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* Recent Conversations */}
        <Card className="hover-card-effect border-white/10 bg-white/5">
          <CardHeader className="flex flex-row items-center justify-between border-b border-white/5 pb-4">
            <CardTitle className="text-lg text-slate-100">Percakapan Terakhir</CardTitle>
            <Link to="/ai-chat">
              <Button variant="ghost" size="sm" className="text-slate-400 hover:text-emerald-400 hover:bg-emerald-500/10">
                Semua <ArrowRight className="w-4 h-4 ml-1" />
              </Button>
            </Link>
          </CardHeader>
          <CardContent className="pt-6">
            {!summary?.recent_conversations || summary.recent_conversations.length === 0 ? (
              <div className="text-center py-8">
                <MessageSquareText className="w-12 h-12 text-slate-700 mx-auto mb-3" />
                <p className="text-sm text-slate-500 mb-4">
                  Belum ada percakapan. Mulai chat dengan AI Assistant.
                </p>
                <Link to="/ai-chat">
                  <Button size="sm" variant="outline" className="border-emerald-500/30 text-emerald-400 hover:bg-emerald-500/10">Mulai Chat</Button>
                </Link>
              </div>
            ) : (
              <div className="space-y-3">
                {summary.recent_conversations.map((conv) => (
                  <Link
                    key={conv.id}
                    to={`/ai-chat?conversation=${conv.id}`}
                    className="block p-3 rounded-lg bg-white/5 hover:bg-white/10 border border-transparent hover:border-emerald-500/30 transition-all group"
                  >
                    <p className="font-medium text-slate-200 group-hover:text-emerald-400 transition-colors truncate">{conv.title}</p>
                    {conv.last_message && (
                      <p className="text-sm text-slate-500 truncate mt-1">{conv.last_message}</p>
                    )}
                    <p className="text-xs text-slate-600 mt-1">
                      {new Date(conv.updated_at).toLocaleDateString('id-ID')}
                    </p>
                  </Link>
                ))}
              </div>
            )}
          </CardContent>
        </Card>

        {/* Insights Summary */}
        <Card className="hover-card-effect border-white/10 bg-white/5">
          <CardHeader className="flex flex-row items-center justify-between border-b border-white/5 pb-4">
            <CardTitle className="text-lg text-slate-100">Ringkasan Insights</CardTitle>
            <Link to="/forecast">
              <Button variant="ghost" size="sm" className="text-slate-400 hover:text-emerald-400 hover:bg-emerald-500/10">
                Lihat Semua <ArrowRight className="w-4 h-4 ml-1" />
              </Button>
            </Link>
          </CardHeader>
          <CardContent className="pt-6">
            {!summary?.insights_summary || summary.total_insights === 0 ? (
              <div className="text-center py-8">
                <BarChart3 className="w-12 h-12 text-slate-700 mx-auto mb-3" />
                <p className="text-sm text-slate-500 mb-4">
                  Belum ada insights. Mulai chat dengan AI Assistant.
                </p>
                <Link to="/ai-chat">
                  <Button size="sm" variant="outline" className="border-emerald-500/30 text-emerald-400 hover:bg-emerald-500/10">Mulai Chat</Button>
                </Link>
              </div>
            ) : (
              <div className="space-y-3">
                <Link
                  to="/forecast"
                  className="flex items-center justify-between p-3 rounded-lg bg-white/5 hover:bg-white/10 border border-transparent hover:border-emerald-500/30 transition-all"
                >
                  <div className="flex items-center gap-3">
                    <BarChart3 className="w-5 h-5 text-emerald-400" />
                    <span className="font-medium text-slate-200">Forecast</span>
                  </div>
                  <span className="text-sm font-semibold text-emerald-400">
                    {summary.insights_summary.forecast}
                  </span>
                </Link>
                <Link
                  to="/market-prediction"
                  className="flex items-center justify-between p-3 rounded-lg bg-white/5 hover:bg-white/10 border border-transparent hover:border-blue-500/30 transition-all"
                >
                  <div className="flex items-center gap-3">
                    <Globe className="w-5 h-5 text-blue-400" />
                    <span className="font-medium text-slate-200">Market Prediction</span>
                  </div>
                  <span className="text-sm font-semibold text-blue-400">
                    {summary.insights_summary.market}
                  </span>
                </Link>
                <Link
                  to="/marketing"
                  className="flex items-center justify-between p-3 rounded-lg bg-white/5 hover:bg-white/10 border border-transparent hover:border-purple-500/30 transition-all"
                >
                  <div className="flex items-center gap-3">
                    <Megaphone className="w-5 h-5 text-purple-400" />
                    <span className="font-medium text-slate-200">Marketing</span>
                  </div>
                  <span className="text-sm font-semibold text-purple-400">
                    {summary.insights_summary.marketing}
                  </span>
                </Link>
                <Link
                  to="/regulation"
                  className="flex items-center justify-between p-3 rounded-lg bg-white/5 hover:bg-white/10 border border-transparent hover:border-orange-500/30 transition-all"
                >
                  <div className="flex items-center gap-3">
                    <Scale className="w-5 h-5 text-orange-400" />
                    <span className="font-medium text-slate-200">Regulation</span>
                  </div>
                  <span className="text-sm font-semibold text-orange-400">
                    {summary.insights_summary.regulation}
                  </span>
                </Link>
              </div>
            )}
          </CardContent>
        </Card>
      </div>

      {/* Recent File Uploads */}
      {summary?.recent_file_uploads && summary.recent_file_uploads.length > 0 && (
        <Card className="hover-card-effect border-white/10 bg-white/5">
          <CardHeader className="flex flex-row items-center justify-between border-b border-white/5 pb-4">
            <CardTitle className="text-lg text-slate-100">File Upload Terakhir</CardTitle>
            <Link to="/ai-chat">
              <Button variant="ghost" size="sm" className="text-slate-400 hover:text-emerald-400 hover:bg-emerald-500/10">
                Upload File <ArrowRight className="w-4 h-4 ml-1" />
              </Button>
            </Link>
          </CardHeader>
          <CardContent className="pt-6">
            <div className="space-y-3">
              {summary.recent_file_uploads.map((file) => (
                <div
                  key={file.id}
                  className="flex items-center justify-between p-3 rounded-lg bg-white/5 hover:bg-white/10 transition-colors"
                >
                  <div className="flex items-center gap-3">
                    <FileText className="w-5 h-5 text-slate-500" />
                    <div>
                      <p className="font-medium text-slate-200 truncate">{file.original_filename}</p>
                      <p className="text-xs text-slate-500">
                        {file.source_type.toUpperCase()} • {file.status}
                      </p>
                    </div>
                  </div>
                  <span className="text-xs text-slate-500">
                    {new Date(file.created_at).toLocaleDateString('id-ID')}
                  </span>
                </div>
              ))}
            </div>
          </CardContent>
        </Card>
      )}
    </div>
  )
}

