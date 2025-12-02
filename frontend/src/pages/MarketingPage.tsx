import { 
  Megaphone, 
  Mail, 
  Instagram, 
  Target, 
  TrendingUp, 
  Users, 
  ArrowRight,
  Sparkles
} from 'lucide-react'
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Progress } from '@/components/ui/progress'

export function MarketingPage() {
  return (
    <div className="max-w-7xl mx-auto px-4 py-8 animate-fade-in-up space-y-8 pb-20">
      {/* Header */}
      <div className="flex flex-col md:flex-row md:items-end justify-between gap-4 border-b border-white/10 pb-6">
        <div>
          <div className="flex items-center gap-2 mb-2">
            <div className="p-2 bg-purple-500/10 rounded-lg border border-purple-500/20">
              <Megaphone className="w-5 h-5 text-purple-400" />
            </div>
            <span className="text-sm font-medium text-purple-400 uppercase tracking-wider">Marketing Strategy</span>
          </div>
          <h1 className="text-3xl md:text-4xl font-display font-bold text-slate-100">
            Rekomendasi Marketing
          </h1>
          <p className="text-slate-400 mt-2 max-w-2xl text-lg">
            Strategi promosi yang dipersonalisasi untuk meningkatkan engagement dan penjualan bisnis Anda.
          </p>
        </div>
      </div>

      {/* Highlighted Strategy */}
      <Card className="bg-gradient-to-br from-purple-900/20 to-black border-purple-500/20 relative overflow-hidden">
        <div className="absolute top-0 right-0 p-32 bg-purple-500/10 blur-3xl rounded-full -mr-16 -mt-16 pointer-events-none" />
        <CardContent className="p-6 md:p-8 flex flex-col md:flex-row gap-8 items-center relative z-10">
          <div className="flex-1">
            <Badge className="bg-purple-500/20 text-purple-300 hover:bg-purple-500/30 border-purple-500/20 mb-4">Top Recommendation</Badge>
            <h2 className="text-2xl font-bold text-slate-100 mb-2">Bundle "Ramadhan Special"</h2>
            <p className="text-slate-300 mb-6 leading-relaxed">
              Berdasarkan tren musiman, membuat paket bundling produk Best Seller dengan item pelengkap dapat meningkatkan Average Order Value (AOV) sebesar 15-20%.
            </p>
            <div className="flex gap-4">
              <Button className="bg-purple-600 hover:bg-purple-500 text-white border-0">
                <Sparkles className="w-4 h-4 mr-2" />
                Generate Konten Iklan
              </Button>
              <Button variant="outline" className="border-white/10 hover:bg-white/5 text-slate-300">
                Lihat Detail Strategi
              </Button>
            </div>
          </div>
          <div className="w-full md:w-1/3 bg-black/40 p-6 rounded-xl border border-white/5 backdrop-blur-sm">
            <h4 className="text-sm font-medium text-slate-400 mb-4">Estimasi Dampak</h4>
            <div className="space-y-4">
              <div>
                <div className="flex justify-between text-sm mb-1">
                  <span className="text-slate-200">Potential Reach</span>
                  <span className="text-purple-400 font-bold">15k+</span>
                </div>
                <Progress value={75} className="h-2 bg-white/5 [&>div]:bg-purple-500" />
              </div>
              <div>
                <div className="flex justify-between text-sm mb-1">
                  <span className="text-slate-200">Conversion Rate</span>
                  <span className="text-purple-400 font-bold">4.2%</span>
                </div>
                <Progress value={60} className="h-2 bg-white/5 [&>div]:bg-purple-500" />
              </div>
            </div>
          </div>
        </CardContent>
      </Card>

      {/* Channel Strategies */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
        <Card className="hover-card-effect border-white/10 bg-white/5">
          <CardHeader>
            <div className="w-10 h-10 rounded-full bg-pink-500/10 flex items-center justify-center mb-2">
              <Instagram className="w-5 h-5 text-pink-400" />
            </div>
            <CardTitle className="text-slate-100">Social Media</CardTitle>
            <CardDescription className="text-slate-400">Optimasi konten Instagram & TikTok</CardDescription>
          </CardHeader>
          <CardContent>
            <ul className="space-y-3 mb-4">
              <li className="text-sm text-slate-300 flex gap-2">
                <CheckCircleIcon className="w-4 h-4 text-emerald-400 shrink-0" />
                Posting Reels "Behind the Scene"
              </li>
              <li className="text-sm text-slate-300 flex gap-2">
                <CheckCircleIcon className="w-4 h-4 text-emerald-400 shrink-0" />
                Kolaborasi dengan Micro-Influencer lokal
              </li>
            </ul>
            <Button variant="link" className="p-0 h-auto text-pink-400 hover:text-pink-300">
              Lihat Kalender Konten <ArrowRight className="w-4 h-4 ml-1" />
            </Button>
          </CardContent>
        </Card>

        <Card className="hover-card-effect border-white/10 bg-white/5">
          <CardHeader>
            <div className="w-10 h-10 rounded-full bg-blue-500/10 flex items-center justify-center mb-2">
              <Mail className="w-5 h-5 text-blue-400" />
            </div>
            <CardTitle className="text-slate-100">Email & Chat</CardTitle>
            <CardDescription className="text-slate-400">Retensi pelanggan via WhatsApp/Email</CardDescription>
          </CardHeader>
          <CardContent>
            <ul className="space-y-3 mb-4">
              <li className="text-sm text-slate-300 flex gap-2">
                <CheckCircleIcon className="w-4 h-4 text-emerald-400 shrink-0" />
                Blast promo "Payday Sale"
              </li>
              <li className="text-sm text-slate-300 flex gap-2">
                <CheckCircleIcon className="w-4 h-4 text-emerald-400 shrink-0" />
                Reminder keranjang belanja (Abandoned Cart)
              </li>
            </ul>
            <Button variant="link" className="p-0 h-auto text-blue-400 hover:text-blue-300">
              Buat Template Pesan <ArrowRight className="w-4 h-4 ml-1" />
            </Button>
          </CardContent>
        </Card>

        <Card className="hover-card-effect border-white/10 bg-white/5">
          <CardHeader>
            <div className="w-10 h-10 rounded-full bg-amber-500/10 flex items-center justify-center mb-2">
              <Target className="w-5 h-5 text-amber-400" />
            </div>
            <CardTitle className="text-slate-100">Paid Ads</CardTitle>
            <CardDescription className="text-slate-400">Iklan tertarget Meta & Google</CardDescription>
          </CardHeader>
          <CardContent>
            <ul className="space-y-3 mb-4">
              <li className="text-sm text-slate-300 flex gap-2">
                <CheckCircleIcon className="w-4 h-4 text-emerald-400 shrink-0" />
                Target audiens umur 18-25 di radius 5km
              </li>
              <li className="text-sm text-slate-300 flex gap-2">
                <CheckCircleIcon className="w-4 h-4 text-emerald-400 shrink-0" />
                Retargeting pengunjung website
              </li>
            </ul>
            <Button variant="link" className="p-0 h-auto text-amber-400 hover:text-amber-300">
              Setup Iklan <ArrowRight className="w-4 h-4 ml-1" />
            </Button>
          </CardContent>
        </Card>
      </div>
    </div>
  )
}

function CheckCircleIcon({ className }: { className?: string }) {
    return (
        <svg 
            xmlns="http://www.w3.org/2000/svg" 
            viewBox="0 0 24 24" 
            fill="none" 
            stroke="currentColor" 
            strokeWidth="2" 
            strokeLinecap="round" 
            strokeLinejoin="round" 
            className={className}
        >
            <path d="M22 11.08V12a10 10 0 1 1-5.93-9.14" />
            <polyline points="22 4 12 14.01 9 11.01" />
        </svg>
    )
}
