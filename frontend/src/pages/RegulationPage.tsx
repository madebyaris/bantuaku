import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'

export function RegulationPage() {
  return (
    <div className="container mx-auto px-4 py-8">
      <div className="mb-6">
        <h1 className="text-3xl font-bold">Government Regulation</h1>
        <p className="text-muted-foreground mt-2">
          Peraturan pemerintah Indonesia yang relevan dengan bisnis Anda
        </p>
      </div>

      <Card>
        <CardHeader>
          <CardTitle>Peraturan Pemerintah</CardTitle>
        </CardHeader>
        <CardContent>
          <p className="text-muted-foreground">
            Informasi peraturan pemerintah akan ditampilkan di sini berdasarkan industri dan
            lokasi bisnis Anda. Data akan diupdate setelah AI Assistant mengetahui detail bisnis Anda.
          </p>
        </CardContent>
      </Card>
    </div>
  )
}
