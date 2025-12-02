import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'

export function MarketingPage() {
  return (
    <div className="container mx-auto px-4 py-8">
      <div className="mb-6">
        <h1 className="text-3xl font-bold">Marketing Recommendation</h1>
        <p className="text-muted-foreground mt-2">
          Rekomendasi kampanye marketing dan strategi promosi untuk bisnis Anda
        </p>
      </div>

      <Card>
        <CardHeader>
          <CardTitle>Rekomendasi Marketing</CardTitle>
        </CardHeader>
        <CardContent>
          <p className="text-muted-foreground">
            Rekomendasi marketing akan ditampilkan di sini setelah AI Assistant mengumpulkan
            informasi tentang bisnis Anda.
          </p>
        </CardContent>
      </Card>
    </div>
  )
}
