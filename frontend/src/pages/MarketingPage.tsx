import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'

export function MarketingPage() {
  return (
    <div className="container mx-auto px-4 py-8 animate-fade-in-up">
      <div className="mb-6">
        <h1 className="text-3xl font-bold font-display text-slate-100">Marketing Recommendation</h1>
        <p className="text-slate-400 mt-2">
          Rekomendasi kampanye marketing dan strategi promosi untuk bisnis Anda
        </p>
      </div>

      <Card className="hover-card-effect border-white/10 bg-white/5 backdrop-blur-xl">
        <CardHeader>
          <CardTitle className="text-slate-100">Rekomendasi Marketing</CardTitle>
        </CardHeader>
        <CardContent>
          <p className="text-slate-400">
            Rekomendasi marketing akan ditampilkan di sini setelah AI Assistant mengumpulkan
            informasi tentang bisnis Anda.
          </p>
        </CardContent>
      </Card>
    </div>
  )
}
