import { useEffect, useState } from 'react'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { MessageSquareText } from 'lucide-react'

export function ForecastPage() {
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    // TODO: Fetch forecast data from API
    setLoading(false)
  }, [])

  return (
    <div className="container mx-auto px-4 py-8">
      <div className="mb-6 flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold">Forecast</h1>
          <p className="text-muted-foreground mt-2">
            Proyeksi penjualan dan stok untuk produk Anda
          </p>
        </div>
        <Button variant="outline" className="gap-2">
          <MessageSquareText className="w-4 h-4" />
          Tanya AI tentang forecast ini
        </Button>
      </div>

      {loading ? (
        <div className="flex items-center justify-center py-12">
          <p className="text-muted-foreground">Memuat data forecast...</p>
        </div>
      ) : (
        <Card>
          <CardHeader>
            <CardTitle>Forecast Penjualan</CardTitle>
          </CardHeader>
          <CardContent>
            <p className="text-muted-foreground">
              Fitur forecast sedang dalam pengembangan. Data akan ditampilkan di sini setelah
              Anda menginput data melalui AI Assistant.
            </p>
          </CardContent>
        </Card>
      )}
    </div>
  )
}
