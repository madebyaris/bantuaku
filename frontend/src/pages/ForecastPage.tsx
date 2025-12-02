import { useEffect, useState } from 'react'
import { 
  TrendingUp, 
  TrendingDown, 
  Calendar, 
  ArrowRight, 
  MessageSquareText, 
  Download, 
  Info,
  Sparkles
} from 'lucide-react'
import {
  AreaChart,
  Area,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  ResponsiveContainer,
  ReferenceLine
} from 'recharts'
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { formatCurrency, cn } from '@/lib/utils'
import { useChatStore } from '@/state/chat'
import { useNavigate } from 'react-router-dom'

export function ForecastPage() {
  const [loading, setLoading] = useState(true)
  const navigate = useNavigate()
  const { addMessage } = useChatStore()

  useEffect(() => {
    // Simulate data loading
    const timer = setTimeout(() => setLoading(false), 1000)
    return () => clearTimeout(timer)
  }, [])

  // Mock Forecast Data (History + Future)
  const forecastData = [
    { month: 'Sep', actual: 12000000, forecast: null, lower: null, upper: null },
    { month: 'Okt', actual: 13500000, forecast: null, lower: null, upper: null },
    { month: 'Nov', actual: 11800000, forecast: null, lower: null, upper: null },
    { month: 'Des', actual: 15200000, forecast: null, lower: null, upper: null },
    { month: 'Jan', actual: 14100000, forecast: 14100000, lower: 14100000, upper: 14100000 }, // Connect point
    { month: 'Feb', actual: null, forecast: 15800000, lower: 14800000, upper: 16800000 },
    { month: 'Mar', actual: null, forecast: 16500000, lower: 15200000, upper: 17800000 },
    { month: 'Apr', actual: null, forecast: 17200000, lower: 15800000, upper: 18600000 },
    { month: 'Mei', actual: null, forecast: 18100000, lower: 16500000, upper: 19700000 },
    { month: 'Jun', actual: null, forecast: 17500000, lower: 15900000, upper: 19100000 },
  ]

  const handleAskAI = () => {
    addMessage({
        id: Date.now().toString(),
        role: 'user',
        text: 'Jelaskan detail forecast penjualan untuk 3 bulan ke depan dan faktor apa yang mempengaruhinya.',
        timestamp: new Date()
    })
    navigate('/ai-chat')
  }

  if (loading) {
    return (
      <div className="flex items-center justify-center h-[calc(100vh-10rem)]">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-emerald-500" />
      </div>
    )
  }

  return (
    <div className="space-y-6 animate-fade-in-up max-w-7xl mx-auto pb-20">
      {/* Header Section */}
      <div className="flex flex-col md:flex-row md:items-end justify-between gap-4 border-b border-white/10 pb-6">
        <div>
          <div className="flex items-center gap-2 mb-2">
            <div className="p-2 bg-emerald-500/10 rounded-lg border border-emerald-500/20">
              <TrendingUp className="w-5 h-5 text-emerald-400" />
            </div>
            <span className="text-sm font-medium text-emerald-400 uppercase tracking-wider">Sales Intelligence</span>
          </div>
          <h1 className="text-3xl md:text-4xl font-display font-bold text-slate-100">
            Sales Forecast
          </h1>
          <p className="text-slate-400 mt-2 max-w-2xl text-lg">
            Proyeksi pertumbuhan penjualan berbasis AI untuk 6 bulan ke depan dengan tingkat akurasi 85%.
          </p>
        </div>
        
        <div className="flex items-center gap-3">
          <Button variant="outline" className="border-white/10 bg-white/5 text-slate-300 hover:bg-white/10 hover:text-white">
            <Download className="w-4 h-4 mr-2" />
            Export
          </Button>
          <Button 
            onClick={handleAskAI}
            className="bg-emerald-500 hover:bg-emerald-400 text-black font-semibold shadow-[0_0_15px_rgba(16,185,129,0.3)]"
          >
            <Sparkles className="w-4 h-4 mr-2" />
            Analisis dengan AI
          </Button>
        </div>
      </div>

      {/* KPI Grid */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
        <Card className="hover-card-effect border-white/10 bg-white/5">
          <CardContent className="pt-6">
            <div className="flex justify-between items-start mb-4">
              <div>
                <p className="text-sm font-medium text-slate-400">Proyeksi Revenue (Feb)</p>
                <h3 className="text-3xl font-bold text-slate-100 mt-2">{formatCurrency(15800000)}</h3>
              </div>
              <div className="p-2 bg-emerald-500/10 rounded-lg">
                <TrendingUp className="w-5 h-5 text-emerald-400" />
              </div>
            </div>
            <div className="flex items-center gap-2 text-sm">
              <span className="text-emerald-400 font-medium flex items-center">
                +12.5% <TrendingUp className="w-3 h-3 ml-1" />
              </span>
              <span className="text-slate-500">vs bulan lalu</span>
            </div>
          </CardContent>
        </Card>

        <Card className="hover-card-effect border-white/10 bg-white/5">
          <CardContent className="pt-6">
            <div className="flex justify-between items-start mb-4">
              <div>
                <p className="text-sm font-medium text-slate-400">Growth Rate (Q1)</p>
                <h3 className="text-3xl font-bold text-slate-100 mt-2">+18.2%</h3>
              </div>
              <div className="p-2 bg-blue-500/10 rounded-lg">
                <Calendar className="w-5 h-5 text-blue-400" />
              </div>
            </div>
            <div className="flex items-center gap-2 text-sm">
              <span className="text-emerald-400 font-medium">On Track</span>
              <span className="text-slate-500">mencapai target Q1</span>
            </div>
          </CardContent>
        </Card>

        <Card className="hover-card-effect border-white/10 bg-white/5">
          <CardContent className="pt-6">
            <div className="flex justify-between items-start mb-4">
              <div>
                <p className="text-sm font-medium text-slate-400">Confidence Score</p>
                <h3 className="text-3xl font-bold text-slate-100 mt-2">85%</h3>
              </div>
              <div className="p-2 bg-purple-500/10 rounded-lg">
                <Info className="w-5 h-5 text-purple-400" />
              </div>
            </div>
            <div className="flex items-center gap-2 text-sm">
              <span className="text-slate-400">Berdasarkan 6 bulan data historis</span>
            </div>
          </CardContent>
        </Card>
      </div>

      {/* Main Chart Section */}
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        <Card className="lg:col-span-2 hover-card-effect border-white/10 bg-white/5">
          <CardHeader className="border-b border-white/5">
            <div className="flex items-center justify-between">
              <div>
                <CardTitle className="text-lg text-slate-100">Trend & Forecast</CardTitle>
                <CardDescription className="text-slate-400">Data historis vs prediksi AI</CardDescription>
              </div>
              <div className="flex items-center gap-2 text-xs text-slate-400 bg-black/20 p-1 rounded-lg border border-white/5">
                <span className="px-2 py-1 rounded bg-emerald-500/10 text-emerald-400 border border-emerald-500/20">History</span>
                <span className="px-2 py-1 rounded bg-blue-500/10 text-blue-400 border border-blue-500/20">Forecast</span>
              </div>
            </div>
          </CardHeader>
          <CardContent className="pt-6">
            <div className="h-[350px] w-full">
              <ResponsiveContainer width="100%" height="100%">
                <AreaChart data={forecastData} margin={{ top: 10, right: 10, left: 0, bottom: 0 }}>
                  <defs>
                    <linearGradient id="colorActual" x1="0" y1="0" x2="0" y2="1">
                      <stop offset="5%" stopColor="#10b981" stopOpacity={0.3}/>
                      <stop offset="95%" stopColor="#10b981" stopOpacity={0}/>
                    </linearGradient>
                    <linearGradient id="colorForecast" x1="0" y1="0" x2="0" y2="1">
                      <stop offset="5%" stopColor="#3b82f6" stopOpacity={0.3}/>
                      <stop offset="95%" stopColor="#3b82f6" stopOpacity={0}/>
                    </linearGradient>
                  </defs>
                  <CartesianGrid strokeDasharray="3 3" stroke="#334155" opacity={0.2} vertical={false} />
                  <XAxis 
                    dataKey="month" 
                    stroke="#94a3b8" 
                    fontSize={12} 
                    tickLine={false} 
                    axisLine={false}
                    dy={10}
                  />
                  <YAxis 
                    stroke="#94a3b8" 
                    fontSize={12} 
                    tickFormatter={(value) => `${(value / 1000000).toFixed(0)}jt`}
                    tickLine={false} 
                    axisLine={false}
                  />
                  <Tooltip
                    contentStyle={{
                      backgroundColor: '#0f172a',
                      border: '1px solid rgba(255,255,255,0.1)',
                      borderRadius: '12px',
                      boxShadow: '0 10px 25px -5px rgba(0, 0, 0, 0.5)',
                      color: '#f8fafc',
                    }}
                    formatter={(value: number, name: string) => [
                      formatCurrency(value),
                      name === 'actual' ? 'Actual Revenue' : 'AI Forecast'
                    ]}
                  />
                  
                  {/* Historical Area */}
                  <Area 
                    type="monotone" 
                    dataKey="actual" 
                    stroke="#10b981" 
                    strokeWidth={3}
                    fillOpacity={1} 
                    fill="url(#colorActual)" 
                    connectNulls
                  />

                  {/* Forecast Area */}
                  <Area 
                    type="monotone" 
                    dataKey="forecast" 
                    stroke="#3b82f6" 
                    strokeWidth={3}
                    strokeDasharray="5 5"
                    fillOpacity={1} 
                    fill="url(#colorForecast)" 
                    connectNulls
                  />
                  
                  <ReferenceLine x="Jan" stroke="#94a3b8" strokeDasharray="3 3" label={{ value: 'Today', position: 'insideTopRight', fill: '#94a3b8', fontSize: 12 }} />
                </AreaChart>
              </ResponsiveContainer>
            </div>
          </CardContent>
        </Card>

        {/* Insights Panel */}
        <div className="space-y-4">
          <Card className="hover-card-effect border-white/10 bg-white/5 h-full">
            <CardHeader>
              <CardTitle className="text-lg text-slate-100">AI Insights</CardTitle>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="p-4 rounded-xl bg-emerald-500/5 border border-emerald-500/10">
                <div className="flex items-start gap-3">
                  <Sparkles className="w-5 h-5 text-emerald-400 mt-0.5 shrink-0" />
                  <div>
                    <h4 className="font-medium text-emerald-400 text-sm mb-1">Trend Positif</h4>
                    <p className="text-sm text-slate-400 leading-relaxed">
                      Penjualan diprediksi meningkat 15% bulan depan karena faktor musiman dan tren pasar positif.
                    </p>
                  </div>
                </div>
              </div>

              <div className="p-4 rounded-xl bg-blue-500/5 border border-blue-500/10">
                <div className="flex items-start gap-3">
                  <Info className="w-5 h-5 text-blue-400 mt-0.5 shrink-0" />
                  <div>
                    <h4 className="font-medium text-blue-400 text-sm mb-1">Rekomendasi Stok</h4>
                    <p className="text-sm text-slate-400 leading-relaxed">
                      Tingkatkan stok produk kategori A sebesar 20% untuk mengantisipasi lonjakan permintaan di bulan Maret.
                    </p>
                  </div>
                </div>
              </div>

              <div className="pt-4">
                <Button 
                    onClick={() => navigate('/ai-chat')}
                    variant="outline" 
                    className="w-full border-white/10 hover:bg-white/5 text-slate-400 hover:text-emerald-400"
                >
                  Lihat Analisis Lengkap <ArrowRight className="w-4 h-4 ml-2" />
                </Button>
              </div>
            </CardContent>
          </Card>
        </div>
      </div>
    </div>
  )
}
