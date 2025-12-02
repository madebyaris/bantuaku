import { 
  Scale, 
  AlertTriangle, 
  CheckCircle2, 
  FileText, 
  ExternalLink, 
  Calendar,
  Shield
} from 'lucide-react'
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'

const regulations = [
  {
    id: 1,
    title: 'Sertifikasi Halal Wajib 2024',
    category: 'Perizinan',
    status: 'urgent',
    deadline: '17 Okt 2024',
    description: 'Seluruh produk makanan dan minuman wajib bersertifikat halal sebelum tenggat waktu.',
    action: 'Daftar Sertifikasi'
  },
  {
    id: 2,
    title: 'Pajak UMKM 0.5%',
    category: 'Perpajakan',
    status: 'compliant',
    deadline: 'Bulanan',
    description: 'Tarif PPh Final 0.5% berlaku untuk omzet di bawah 4.8M per tahun.',
    action: 'Lihat Detail'
  },
  {
    id: 3,
    title: 'Izin Edar BPOM',
    category: 'Keamanan Pangan',
    status: 'pending',
    deadline: '-',
    description: 'Diperlukan untuk produk pangan olahan kemasan eceran yang diproduksi dalam negeri.',
    action: 'Cek Syarat'
  }
]

export function RegulationPage() {
  return (
    <div className="max-w-7xl mx-auto px-4 py-8 animate-fade-in-up space-y-8 pb-20">
      {/* Header */}
      <div className="flex flex-col md:flex-row md:items-end justify-between gap-4 border-b border-white/10 pb-6">
        <div>
          <div className="flex items-center gap-2 mb-2">
            <div className="p-2 bg-orange-500/10 rounded-lg border border-orange-500/20">
              <Scale className="w-5 h-5 text-orange-400" />
            </div>
            <span className="text-sm font-medium text-orange-400 uppercase tracking-wider">Legal & Compliance</span>
          </div>
          <h1 className="text-3xl md:text-4xl font-display font-bold text-slate-100">
            Peraturan Pemerintah
          </h1>
          <p className="text-slate-400 mt-2 max-w-2xl text-lg">
            Monitor kepatuhan bisnis Anda terhadap regulasi terbaru pemerintah Indonesia.
          </p>
        </div>
      </div>

      {/* Compliance Score Alert */}
      <div className="bg-gradient-to-r from-orange-900/20 to-black border border-orange-500/20 rounded-xl p-6 flex flex-col md:flex-row items-center gap-6">
        <div className="relative">
            <div className="w-20 h-20 rounded-full border-4 border-orange-500/30 flex items-center justify-center">
                <span className="text-2xl font-bold text-orange-400">75%</span>
            </div>
        </div>
        <div className="flex-1">
            <h3 className="text-xl font-bold text-slate-100 mb-1">Status Kepatuhan: Perlu Perhatian</h3>
            <p className="text-slate-400">Bisnis Anda telah memenuhi sebagian besar regulasi dasar, namun ada 1 item prioritas yang memerlukan tindakan segera.</p>
        </div>
        <Button className="bg-orange-600 hover:bg-orange-500 text-white border-0 shrink-0 w-full md:w-auto">
            Lengkapi Dokumen
        </Button>
      </div>

      {/* Regulation Cards */}
      <div className="grid grid-cols-1 gap-4">
        {regulations.map((item) => (
            <Card key={item.id} className="hover-card-effect border-white/10 bg-white/5 group">
                <CardContent className="p-6 flex flex-col md:flex-row gap-6 items-start md:items-center">
                    <div className="p-3 rounded-xl bg-white/5 border border-white/5 group-hover:border-white/10 transition-colors">
                        {item.status === 'urgent' && <AlertTriangle className="w-6 h-6 text-red-400" />}
                        {item.status === 'compliant' && <CheckCircle2 className="w-6 h-6 text-emerald-400" />}
                        {item.status === 'pending' && <FileText className="w-6 h-6 text-blue-400" />}
                    </div>
                    
                    <div className="flex-1">
                        <div className="flex items-center gap-3 mb-1">
                            <h3 className="font-bold text-lg text-slate-100">{item.title}</h3>
                            <Badge variant="outline" className="bg-white/5 text-slate-400 border-white/10 font-normal">
                                {item.category}
                            </Badge>
                            {item.status === 'urgent' && (
                                <Badge className="bg-red-500/20 text-red-400 hover:bg-red-500/30 border-red-500/20">
                                    Urgent
                                </Badge>
                            )}
                        </div>
                        <p className="text-slate-400 mb-2">{item.description}</p>
                        <div className="flex items-center gap-4 text-xs text-slate-500">
                            <span className="flex items-center gap-1">
                                <Calendar className="w-3 h-3" /> Deadline: {item.deadline}
                            </span>
                            <span className="flex items-center gap-1">
                                <Shield className="w-3 h-3" /> Mandatory
                            </span>
                        </div>
                    </div>

                    <Button variant="outline" className="border-white/10 bg-transparent hover:bg-white/5 text-slate-300 group-hover:text-white shrink-0 w-full md:w-auto">
                        {item.action} <ExternalLink className="w-3 h-3 ml-2 opacity-50" />
                    </Button>
                </CardContent>
            </Card>
        ))}
      </div>
    </div>
  )
}
