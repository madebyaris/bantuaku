import { useState } from 'react'
import { 
  TrendingUp, 
  Globe, 
  MapPin, 
  Search,
  ArrowRight,
  BarChart3,
  Zap
} from 'lucide-react'
import {
  LineChart,
  Line,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  ResponsiveContainer,
  Legend
} from 'recharts'
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Tabs, TabsList, TabsTrigger, TabsContent } from '@/components/ui/tabs'
import { Badge } from '@/components/ui/badge'

// Mock Data
const trendData = [
  { month: 'Sep', local: 45, global: 30 },
  { month: 'Okt', local: 52, global: 35 },
  { month: 'Nov', local: 48, global: 42 },
  { month: 'Des', local: 61, global: 48 },
  { month: 'Jan', local: 55, global: 52 },
  { month: 'Feb', local: 67, global: 58 },
]

export function MarketPredictionPage() {
  const [scope, setScope] = useState<'local' | 'global'>('local')

  return (
    <div className="max-w-7xl mx-auto px-4 py-8 animate-fade-in-up space-y-8 pb-20">
      {/* Header */}
      <div className="flex flex-col md:flex-row md:items-end justify-between gap-4 border-b border-white/10 pb-6">
        <div>
          <div className="flex items-center gap-2 mb-2">
            <div className="p-2 bg-blue-500/10 rounded-lg border border-blue-500/20">
              <Globe className="w-5 h-5 text-blue-400" />
            </div>
            <span className="text-sm font-medium text-blue-400 uppercase tracking-wider">Market Intelligence</span>
          </div>
          <h1 className="text-3xl md:text-4xl font-display font-bold text-slate-100">
            Prediksi Pasar
          </h1>
          <p className="text-slate-400 mt-2 max-w-2xl text-lg">
            Analisis tren pasar real-time dan identifikasi peluang bisnis baru sebelum kompetitor Anda.
          </p>
        </div>
      </div>

      <Tabs value={scope} onValueChange={(v) => setScope(v as 'local' | 'global')} className="space-y-6">
        <TabsList className="bg-white/5 border border-white/10 p-1 h-auto rounded-xl">
          <TabsTrigger 
            value="local" 
            className="data-[state=active]:bg-emerald-500/20 data-[state=active]:text-emerald-400 text-slate-400 px-6 py-2.5 rounded-lg transition-all"
          >
            <MapPin className="w-4 h-4 mr-2" />
            Pasar Lokal (Indonesia)
          </TabsTrigger>
          <TabsTrigger 
            value="global" 
            className="data-[state=active]:bg-blue-500/20 data-[state=active]:text-blue-400 text-slate-400 px-6 py-2.5 rounded-lg transition-all"
          >
            <Globe className="w-4 h-4 mr-2" />
            Tren Global
          </TabsTrigger>
        </TabsList>

        {/* Content Area */}
        <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
          {/* Main Chart */}
          <Card className="lg:col-span-2 hover-card-effect border-white/10 bg-white/5">
            <CardHeader>
              <CardTitle className="text-lg text-slate-100 flex items-center gap-2">
                <BarChart3 className="w-5 h-5 text-slate-400" />
                Indeks Minat Konsumen
              </CardTitle>
              <CardDescription className="text-slate-400">
                Perbandingan minat pasar Lokal vs Global 6 bulan terakhir
              </CardDescription>
            </CardHeader>
            <CardContent>
              <div className="h-[300px] w-full">
                <ResponsiveContainer width="100%" height="100%">
                  <LineChart data={trendData}>
                    <CartesianGrid strokeDasharray="3 3" stroke="#334155" opacity={0.2} vertical={false} />
                    <XAxis dataKey="month" stroke="#94a3b8" fontSize={12} tickLine={false} axisLine={false} />
                    <YAxis stroke="#94a3b8" fontSize={12} tickLine={false} axisLine={false} />
                    <Tooltip 
                      contentStyle={{
                        backgroundColor: '#0f172a',
                        border: '1px solid rgba(255,255,255,0.1)',
                        borderRadius: '8px',
                        color: '#f8fafc',
                      }}
                    />
                    <Legend />
                    <Line type="monotone" dataKey="local" name="Lokal" stroke="#10b981" strokeWidth={3} dot={{r:4}} activeDot={{r:6}} />
                    <Line type="monotone" dataKey="global" name="Global" stroke="#3b82f6" strokeWidth={3} dot={{r:4}} activeDot={{r:6}} />
                  </LineChart>
                </ResponsiveContainer>
              </div>
            </CardContent>
          </Card>

          {/* Trending Keywords / Products */}
          <Card className="hover-card-effect border-white/10 bg-white/5">
            <CardHeader>
              <CardTitle className="text-lg text-slate-100 flex items-center gap-2">
                <Zap className="w-5 h-5 text-yellow-400" />
                Produk Trending Saat Ini
              </CardTitle>
            </CardHeader>
            <CardContent>
              <div className="space-y-4">
                {[
                  { name: 'Kopi Susu Gula Aren', growth: '+24%', volume: 'High' },
                  { name: 'Keripik Kaca Pedas', growth: '+18%', volume: 'Med' },
                  { name: 'Baso Aci Instan', growth: '+15%', volume: 'Med' },
                  { name: 'Minuman Collagen', growth: '+12%', volume: 'Low' },
                ].map((item, i) => (
                  <div key={i} className="flex items-center justify-between p-3 rounded-lg bg-white/5 border border-white/5 hover:border-white/10 transition-colors">
                    <div>
                      <p className="font-medium text-slate-200">{item.name}</p>
                      <div className="flex items-center gap-2 mt-1">
                        <Badge variant="outline" className="text-[10px] border-white/10 text-slate-400">{item.volume} Vol</Badge>
                      </div>
                    </div>
                    <div className="text-right">
                      <span className="text-emerald-400 font-bold text-sm">{item.growth}</span>
                      <p className="text-[10px] text-slate-500">vs bulan lalu</p>
                    </div>
                  </div>
                ))}
              </div>
              <Button variant="ghost" className="w-full mt-4 text-sm text-emerald-400 hover:text-emerald-300 hover:bg-emerald-500/10">
                Lihat Semua Tren <ArrowRight className="w-4 h-4 ml-2" />
              </Button>
            </CardContent>
          </Card>
        </div>

        {/* Detailed Insight Tabs Content */}
        <TabsContent value="local" className="mt-0 animate-in fade-in slide-in-from-bottom-4">
          <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
            <Card className="hover-card-effect border-white/10 bg-white/5">
              <CardHeader>
                <CardTitle className="text-slate-100">Peluang Pasar Lokal</CardTitle>
              </CardHeader>
              <CardContent>
                <ul className="space-y-3">
                  <li className="flex gap-3 text-sm text-slate-300">
                    <div className="w-1.5 h-1.5 rounded-full bg-emerald-400 mt-2 shrink-0" />
                    Permintaan produk ramah lingkungan meningkat 30% di kota besar.
                  </li>
                  <li className="flex gap-3 text-sm text-slate-300">
                    <div className="w-1.5 h-1.5 rounded-full bg-emerald-400 mt-2 shrink-0" />
                    Tren "Local Pride" mendorong penjualan brand fashion lokal.
                  </li>
                </ul>
              </CardContent>
            </Card>
            <Card className="hover-card-effect border-white/10 bg-white/5">
              <CardHeader>
                <CardTitle className="text-slate-100">Kompetisi Area Anda</CardTitle>
              </CardHeader>
              <CardContent>
                <p className="text-slate-400 text-sm leading-relaxed">
                  Kompetisi di sektor F&B di area Anda tergolong <strong>Tinggi</strong>. Disarankan untuk fokus pada diferensiasi produk atau bundling paket hemat.
                </p>
              </CardContent>
            </Card>
          </div>
        </TabsContent>

        <TabsContent value="global" className="mt-0 animate-in fade-in slide-in-from-bottom-4">
          <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
            <Card className="hover-card-effect border-white/10 bg-white/5">
              <CardHeader>
                <CardTitle className="text-slate-100">Insight Pasar Global</CardTitle>
              </CardHeader>
              <CardContent>
                <ul className="space-y-3">
                  <li className="flex gap-3 text-sm text-slate-300">
                    <div className="w-1.5 h-1.5 rounded-full bg-blue-400 mt-2 shrink-0" />
                    Adopsi AI dalam retail meningkat drastis di Asia Pasifik.
                  </li>
                  <li className="flex gap-3 text-sm text-slate-300">
                    <div className="w-1.5 h-1.5 rounded-full bg-blue-400 mt-2 shrink-0" />
                    Supply chain global mulai pulih, menurunkan biaya logistik.
                  </li>
                </ul>
              </CardContent>
            </Card>
          </div>
        </TabsContent>
      </Tabs>
    </div>
  )
}
