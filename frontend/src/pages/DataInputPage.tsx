import { useEffect, useState } from 'react'
import { FileUp, Loader2, Check, AlertCircle } from 'lucide-react'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Tabs, TabsList, TabsTrigger, TabsContent } from '@/components/ui/tabs'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { api, Product, ImportResult } from '@/lib/api'
import { toast } from '@/components/ui/toaster'
import { formatCurrency } from '@/lib/utils'

export function DataInputPage() {
  const [products, setProducts] = useState<Product[]>([])
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    async function load() {
      try {
        const data = await api.products.list()
        setProducts(data)
      } catch {
        // Ignore
      } finally {
        setLoading(false)
      }
    }
    load()
  }, [])

  if (loading) {
    return (
      <div className="flex items-center justify-center h-96">
        <Loader2 className="w-8 h-8 animate-spin text-purple-600" />
      </div>
    )
  }

  return (
    <div className="max-w-3xl mx-auto animate-fade-in">
      <Tabs defaultValue="manual" className="space-y-6">
        <TabsList className="grid w-full grid-cols-2">
          <TabsTrigger value="manual">Input Manual</TabsTrigger>
          <TabsTrigger value="csv">Upload CSV</TabsTrigger>
        </TabsList>

        <TabsContent value="manual">
          <ManualEntryForm products={products} />
        </TabsContent>

        <TabsContent value="csv">
          <CSVUploadForm />
        </TabsContent>
      </Tabs>
    </div>
  )
}

function ManualEntryForm({ products }: { products: Product[] }) {
  const [loading, setLoading] = useState(false)
  const [form, setForm] = useState({
    product_id: '',
    quantity: '',
    price: '',
    sale_date: new Date().toISOString().split('T')[0],
  })

  const selectedProduct = products.find((p) => p.id === form.product_id)

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    setLoading(true)

    try {
      await api.sales.record({
        product_id: form.product_id,
        quantity: parseInt(form.quantity),
        price: parseFloat(form.price),
        sale_date: form.sale_date,
      })
      toast({ title: 'Penjualan berhasil dicatat!', variant: 'success' })
      setForm({
        product_id: '',
        quantity: '',
        price: '',
        sale_date: new Date().toISOString().split('T')[0],
      })
    } catch (err) {
      toast({
        title: 'Gagal mencatat penjualan',
        description: err instanceof Error ? err.message : 'Terjadi kesalahan',
        variant: 'destructive',
      })
    } finally {
      setLoading(false)
    }
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle className="text-lg">Catat Penjualan Manual</CardTitle>
      </CardHeader>
      <CardContent>
        {products.length === 0 ? (
          <div className="text-center py-8">
            <p className="text-slate-500 mb-4">
              Belum ada produk. Tambahkan produk terlebih dahulu.
            </p>
            <Button variant="outline" asChild>
              <a href="/products">Ke Halaman Produk</a>
            </Button>
          </div>
        ) : (
          <form onSubmit={handleSubmit} className="space-y-4">
            <div className="space-y-2">
              <Label>Produk *</Label>
              <Select
                value={form.product_id}
                onValueChange={(value) => {
                  const product = products.find((p) => p.id === value)
                  setForm({
                    ...form,
                    product_id: value,
                    price: product?.unit_price.toString() || '',
                  })
                }}
              >
                <SelectTrigger>
                  <SelectValue placeholder="Pilih produk..." />
                </SelectTrigger>
                <SelectContent>
                  {products.map((product) => (
                  <SelectItem key={product.id} value={product.id}>
                      {product.name} - {formatCurrency(product.unit_price)}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>

            <div className="grid grid-cols-2 gap-4">
              <div className="space-y-2">
                <Label>Jumlah Terjual *</Label>
                <Input
                  type="number"
                  value={form.quantity}
                  onChange={(e) => setForm({ ...form, quantity: e.target.value })}
                  placeholder="10"
                  min="1"
                  required
                />
              </div>
              <div className="space-y-2">
                <Label>Harga per Unit</Label>
                <Input
                  type="number"
                  value={form.price}
                  onChange={(e) => setForm({ ...form, price: e.target.value })}
                  placeholder={selectedProduct?.unit_price.toString() || '0'}
                />
              </div>
            </div>

            <div className="space-y-2">
              <Label>Tanggal Penjualan</Label>
              <Input
                type="date"
                value={form.sale_date}
                onChange={(e) => setForm({ ...form, sale_date: e.target.value })}
                max={new Date().toISOString().split('T')[0]}
              />
            </div>

            <Button type="submit" className="w-full" disabled={loading}>
              {loading ? (
                <>
                  <Loader2 className="w-4 h-4 mr-2 animate-spin" />
                  Menyimpan...
                </>
              ) : (
                'Catat Penjualan'
              )}
            </Button>
          </form>
        )}
      </CardContent>
    </Card>
  )
}

function CSVUploadForm() {
  const [uploading, setUploading] = useState(false)
  const [result, setResult] = useState<ImportResult | null>(null)
  const [dragActive, setDragActive] = useState(false)

  async function handleUpload(file: File) {
    if (!file) return

    setUploading(true)
    setResult(null)

    try {
      const importResult = await api.sales.importCSV(file)
      setResult(importResult)
      
      if (importResult.success_count > 0) {
        toast({
          title: 'Import berhasil!',
          description: `${importResult.success_count} baris berhasil diimport`,
          variant: 'success',
        })
      }
    } catch (err) {
      toast({
        title: 'Gagal import CSV',
        description: err instanceof Error ? err.message : 'Terjadi kesalahan',
        variant: 'destructive',
      })
    } finally {
      setUploading(false)
    }
  }

  function handleDrop(e: React.DragEvent) {
    e.preventDefault()
    setDragActive(false)
    const file = e.dataTransfer.files[0]
    if (file && (file.name.endsWith('.csv') || file.name.endsWith('.xlsx'))) {
      handleUpload(file)
    }
  }

  function handleFileChange(e: React.ChangeEvent<HTMLInputElement>) {
    const file = e.target.files?.[0]
    if (file) handleUpload(file)
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle className="text-lg">Import Data dari CSV</CardTitle>
      </CardHeader>
      <CardContent className="space-y-4">
        {/* Upload Area */}
        <div
          className={`border-2 border-dashed rounded-lg p-8 text-center transition-colors ${
            dragActive
              ? 'border-purple-500 bg-purple-50'
              : 'border-slate-200 hover:border-purple-300'
          }`}
          onDragOver={(e) => {
            e.preventDefault()
            setDragActive(true)
          }}
          onDragLeave={() => setDragActive(false)}
          onDrop={handleDrop}
        >
          <FileUp className="w-12 h-12 mx-auto text-slate-400 mb-4" />
          <p className="text-slate-600 mb-2">
            Drag & drop file CSV atau Excel di sini
          </p>
          <p className="text-sm text-slate-400 mb-4">atau</p>
          <label>
            <input
              type="file"
              accept=".csv,.xlsx"
              onChange={handleFileChange}
              className="hidden"
              disabled={uploading}
            />
            <Button variant="outline" disabled={uploading} asChild>
              <span>
                {uploading ? (
                  <>
                    <Loader2 className="w-4 h-4 mr-2 animate-spin" />
                    Mengupload...
                  </>
                ) : (
                  'Pilih File'
                )}
              </span>
            </Button>
          </label>
        </div>

        {/* Template Download */}
        <div className="p-4 bg-slate-50 rounded-lg">
          <p className="text-sm font-medium text-slate-700 mb-2">Format CSV:</p>
          <p className="text-xs text-slate-500 font-mono bg-white p-2 rounded border">
            product_name,quantity,sale_date,price
            <br />
            Kopi Arabica 250g,5,2025-11-25,85000
            <br />
            Teh Hijau 100g,3,2025-11-25,35000
          </p>
        </div>

        {/* Result */}
        {result && (
          <div className="space-y-3">
            {result.success_count > 0 && (
              <div className="flex items-center gap-2 p-3 bg-green-50 text-green-700 rounded-lg">
                <Check className="w-5 h-5" />
                <span>{result.success_count} baris berhasil diimport</span>
              </div>
            )}
            {result.errors.length > 0 && (
              <div className="p-3 bg-red-50 rounded-lg">
                <div className="flex items-center gap-2 text-red-700 mb-2">
                  <AlertCircle className="w-5 h-5" />
                  <span>{result.errors.length} error ditemukan</span>
                </div>
                <ul className="text-sm text-red-600 space-y-1">
                  {result.errors.slice(0, 5).map((err, i) => (
                    <li key={i}>
                      Baris {err.row}: {err.error}
                    </li>
                  ))}
                  {result.errors.length > 5 && (
                    <li>... dan {result.errors.length - 5} error lainnya</li>
                  )}
                </ul>
              </div>
            )}
          </div>
        )}
      </CardContent>
    </Card>
  )
}
