import { useEffect, useState } from "react";
import { Plus, Search, Loader2, TrendingUp, Package } from "lucide-react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { api, Product, ForecastResponse } from "@/lib/api";
import { formatCurrency, cn, getRiskColor } from "@/lib/utils";
import { toast } from "@/components/ui/toaster";

export function ProductsPage() {
  const [products, setProducts] = useState<Product[]>([]);
  const [loading, setLoading] = useState(true);
  const [searchTerm, setSearchTerm] = useState("");
  const [showAddModal, setShowAddModal] = useState(false);
  const [selectedProduct, setSelectedProduct] = useState<Product | null>(null);
  const [forecast, setForecast] = useState<ForecastResponse | null>(null);
  const [loadingForecast, setLoadingForecast] = useState(false);

  useEffect(() => {
    loadProducts();
  }, []);

  async function loadProducts() {
    try {
      const data = await api.products.list();
      setProducts(data);
    } catch (err) {
      toast({ title: "Gagal memuat produk", variant: "destructive" });
    } finally {
      setLoading(false);
    }
  }

  async function handleSelectProduct(product: Product) {
    setSelectedProduct(product);
    setLoadingForecast(true);
    try {
      const forecastData = await api.forecasts.get(product.id);
      setForecast(forecastData);
    } catch {
      setForecast(null);
    } finally {
      setLoadingForecast(false);
    }
  }

  const filteredProducts = products.filter((p) =>
    p.name.toLowerCase().includes(searchTerm.toLowerCase())
  );

  if (loading) {
    return (
      <div className="flex items-center justify-center h-96">
        <Loader2 className="w-8 h-8 animate-spin text-purple-600" />
      </div>
    );
  }

  return (
    <div className="space-y-6 animate-fade-in">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div className="relative flex-1 max-w-md">
          <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-slate-400" />
          <Input
            placeholder="Cari produk..."
            value={searchTerm}
            onChange={(e) => setSearchTerm(e.target.value)}
            className="pl-10"
          />
        </div>
        <Button onClick={() => setShowAddModal(true)}>
          <Plus className="w-4 h-4 mr-2" />
          Tambah Produk
        </Button>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        {/* Product List */}
        <div className="lg:col-span-2 space-y-3">
          {filteredProducts.length === 0 ? (
            <Card>
              <CardContent className="py-12 text-center">
                <Package className="w-12 h-12 mx-auto text-slate-300 mb-4" />
                <p className="text-slate-500">
                  {searchTerm
                    ? "Tidak ada produk ditemukan"
                    : "Belum ada produk"}
                </p>
                <Button className="mt-4" onClick={() => setShowAddModal(true)}>
                  Tambah Produk Pertama
                </Button>
              </CardContent>
            </Card>
          ) : (
            filteredProducts.map((product) => (
              <Card
                key={product.id}
                className={cn(
                  "cursor-pointer transition-all hover:shadow-md",
                  selectedProduct?.id === product.id && "ring-2 ring-purple-500"
                )}
                onClick={() => handleSelectProduct(product)}
              >
                <CardContent className="py-4">
                  <div className="flex items-center justify-between">
                    <div className="flex-1">
                      <h3 className="font-semibold text-slate-900">
                        {product.name}
                      </h3>
                      <div className="flex items-center gap-4 mt-1 text-sm text-slate-500">
                        {product.sku && <span>SKU: {product.sku}</span>}
                        {product.category && <span>{product.category}</span>}
                      </div>
                    </div>
                    <div className="text-right">
                      <p className="font-semibold text-slate-900">
                        {formatCurrency(product.unit_price)}
                      </p>
                    </div>
                  </div>
                </CardContent>
              </Card>
            ))
          )}
        </div>

        {/* Product Detail & Forecast */}
        <div className="space-y-4">
          {selectedProduct ? (
            <>
              <Card>
                <CardHeader>
                  <CardTitle className="text-lg">Detail Produk</CardTitle>
                </CardHeader>
                <CardContent className="space-y-4">
                  <div>
                    <Label className="text-slate-500">Nama Produk</Label>
                    <p className="font-semibold">{selectedProduct.name}</p>
                  </div>
                  <div className="grid grid-cols-2 gap-4">
                    <div>
                      <Label className="text-slate-500">Harga Jual</Label>
                      <p className="font-semibold">
                        {formatCurrency(selectedProduct.unit_price)}
                      </p>
                    </div>
                  </div>
                  {selectedProduct.category && (
                    <div>
                      <Label className="text-slate-500">Kategori</Label>
                      <p className="font-semibold">
                        {selectedProduct.category}
                      </p>
                    </div>
                  )}
                </CardContent>
              </Card>

              {/* Forecast */}
              <Card>
                <CardHeader>
                  <CardTitle className="text-lg flex items-center gap-2">
                    <TrendingUp className="w-5 h-5 text-purple-600" />
                    Forecast 30 Hari
                  </CardTitle>
                </CardHeader>
                <CardContent>
                  {loadingForecast ? (
                    <div className="flex justify-center py-8">
                      <Loader2 className="w-6 h-6 animate-spin text-purple-600" />
                    </div>
                  ) : forecast ? (
                    <div className="space-y-4">
                      <div className="p-4 bg-purple-50 rounded-lg">
                        <p className="text-sm text-purple-600 font-medium">
                          Prediksi Permintaan
                        </p>
                        <p className="text-3xl font-bold text-purple-700">
                          {forecast.forecast_30d} unit
                        </p>
                        <p className="text-sm text-purple-600 mt-1">
                          Confidence: {(forecast.confidence * 100).toFixed(0)}%
                        </p>
                      </div>

                      <div className="text-xs text-slate-500">
                        Algoritma: {forecast.algorithm} • Diperbarui:{" "}
                        {new Date(forecast.generated_at).toLocaleString(
                          "id-ID"
                        )}
                      </div>
                    </div>
                  ) : (
                    <p className="text-sm text-slate-500 text-center py-4">
                      Belum ada data penjualan untuk forecast
                    </p>
                  )}
                </CardContent>
              </Card>
            </>
          ) : (
            <Card>
              <CardContent className="py-12 text-center">
                <p className="text-slate-500">
                  Pilih produk untuk melihat detail dan forecast
                </p>
              </CardContent>
            </Card>
          )}
        </div>
      </div>

      {/* Add Product Modal */}
      {showAddModal && (
        <AddProductModal
          onClose={() => setShowAddModal(false)}
          onSuccess={() => {
            setShowAddModal(false);
            loadProducts();
          }}
        />
      )}
    </div>
  );
}

function AddProductModal({
  onClose,
  onSuccess,
}: {
  onClose: () => void;
  onSuccess: () => void;
}) {
  const [loading, setLoading] = useState(false);
  const [form, setForm] = useState({
    product_name: "",
    sku: "",
    category: "",
    unit_price: "",
    cost: "",
  });

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    setLoading(true);

    try {
      await api.products.create({
        product_name: form.product_name,
        sku: form.sku || undefined,
        category: form.category || undefined,
        unit_price: parseFloat(form.unit_price) || 0,
        cost: parseFloat(form.cost) || undefined,
      });
      toast({ title: "Produk berhasil ditambahkan", variant: "success" });
      onSuccess();
    } catch (err) {
      toast({
        title: "Gagal menambah produk",
        description: err instanceof Error ? err.message : "Terjadi kesalahan",
        variant: "destructive",
      });
    } finally {
      setLoading(false);
    }
  }

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50">
      <Card className="w-full max-w-md mx-4 animate-slide-in-from-bottom">
        <CardHeader>
          <CardTitle>Tambah Produk Baru</CardTitle>
        </CardHeader>
        <CardContent>
          <form onSubmit={handleSubmit} className="space-y-4">
            <div className="space-y-2">
              <Label>Nama Produk *</Label>
              <Input
                value={form.product_name}
                onChange={(e) =>
                  setForm({ ...form, product_name: e.target.value })
                }
                placeholder="Kopi Arabica 250g"
                required
              />
            </div>
            <div className="grid grid-cols-2 gap-4">
              <div className="space-y-2">
                <Label>SKU</Label>
                <Input
                  value={form.sku}
                  onChange={(e) => setForm({ ...form, sku: e.target.value })}
                  placeholder="KOP-ARB-250"
                />
              </div>
              <div className="space-y-2">
                <Label>Kategori</Label>
                <Input
                  value={form.category}
                  onChange={(e) =>
                    setForm({ ...form, category: e.target.value })
                  }
                  placeholder="Minuman"
                />
              </div>
            </div>
            <div className="space-y-2">
              <Label>Harga Jual *</Label>
              <Input
                type="number"
                value={form.unit_price}
                onChange={(e) =>
                  setForm({ ...form, unit_price: e.target.value })
                }
                placeholder="85000"
                required
              />
            </div>
            <div className="flex gap-3 pt-4">
              <Button
                type="button"
                variant="outline"
                className="flex-1"
                onClick={onClose}
              >
                Batal
              </Button>
              <Button type="submit" className="flex-1" disabled={loading}>
                {loading ? (
                  <Loader2 className="w-4 h-4 animate-spin" />
                ) : (
                  "Simpan"
                )}
              </Button>
            </div>
          </form>
        </CardContent>
      </Card>
    </div>
  );
}
