import { useState } from 'react'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Tabs, TabsList, TabsTrigger, TabsContent } from '@/components/ui/tabs'

export function MarketPredictionPage() {
  const [scope, setScope] = useState<'local' | 'global'>('local')

  return (
    <div className="container mx-auto px-4 py-8">
      <div className="mb-6">
        <h1 className="text-3xl font-bold">Market Prediction</h1>
        <p className="text-muted-foreground mt-2">
          Prediksi tren pasar lokal dan global untuk produk Anda
        </p>
      </div>

      <Tabs value={scope} onValueChange={(v) => setScope(v as 'local' | 'global')}>
        <TabsList>
          <TabsTrigger value="local">Lokal (Indonesia)</TabsTrigger>
          <TabsTrigger value="global">Global</TabsTrigger>
        </TabsList>

        <TabsContent value="local" className="mt-6">
          <Card>
            <CardHeader>
              <CardTitle>Tren Pasar Lokal</CardTitle>
            </CardHeader>
            <CardContent>
              <p className="text-muted-foreground">
                Analisis tren pasar Indonesia akan ditampilkan di sini setelah data tersedia.
              </p>
            </CardContent>
          </Card>
        </TabsContent>

        <TabsContent value="global" className="mt-6">
          <Card>
            <CardHeader>
              <CardTitle>Tren Pasar Global</CardTitle>
            </CardHeader>
            <CardContent>
              <p className="text-muted-foreground">
                Analisis tren pasar global akan ditampilkan di sini setelah data tersedia.
              </p>
            </CardContent>
          </Card>
        </TabsContent>
      </Tabs>
    </div>
  )
}
