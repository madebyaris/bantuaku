import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'

export function RegulationPage() {
  return (
    <div className="container mx-auto px-4 py-8 animate-fade-in-up">
      <div className="mb-6">
        <h1 className="text-3xl font-bold font-display text-slate-100">Government Regulation</h1>
        <p className="text-slate-400 mt-2">
          Peraturan pemerintah Indonesia yang relevan dengan bisnis Anda
        </p>
      </div>

      <Card className="hover-card-effect border-white/10 bg-white/5 backdrop-blur-xl">
        <CardHeader>
          <CardTitle className="text-slate-100">Peraturan Pemerintah</CardTitle>
        </CardHeader>
        <CardContent>
          <p className="text-slate-400">
            Informasi peraturan pemerintah akan ditampilkan di sini berdasarkan industri dan
            lokasi bisnis Anda. Data akan diupdate setelah AI Assistant mengetahui detail bisnis Anda.
          </p>
        </CardContent>
      </Card>
    </div>
  )
}
