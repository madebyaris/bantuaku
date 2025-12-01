import { useState } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { Sparkles, Loader2 } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { api } from '@/lib/api'
import { useAuthStore } from '@/state/auth'
import { toast } from '@/components/ui/toaster'

const industries = [
  { value: 'retail', label: 'Retail / Toko' },
  { value: 'food', label: 'Makanan & Minuman' },
  { value: 'fashion', label: 'Fashion & Pakaian' },
  { value: 'beauty', label: 'Kecantikan & Skincare' },
  { value: 'electronics', label: 'Elektronik & Gadget' },
  { value: 'other', label: 'Lainnya' },
]

export function RegisterPage() {
  const navigate = useNavigate()
  const login = useAuthStore((state) => state.login)
  const [storeName, setStoreName] = useState('')
  const [industry, setIndustry] = useState('')
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [loading, setLoading] = useState(false)

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    setLoading(true)

    try {
      const data = await api.auth.register(email, password, storeName, industry)
      login(data)
      toast({ title: 'Akun berhasil dibuat!', variant: 'success' })
      navigate('/dashboard')
    } catch (err) {
      toast({
        title: 'Gagal mendaftar',
        description: err instanceof Error ? err.message : 'Terjadi kesalahan',
        variant: 'destructive',
      })
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="min-h-screen flex items-center justify-center bg-gradient-to-br from-purple-600 via-indigo-600 to-blue-700 p-4">
      {/* Background pattern */}
      <div className="absolute inset-0 bg-[url('data:image/svg+xml;base64,PHN2ZyB3aWR0aD0iNjAiIGhlaWdodD0iNjAiIHZpZXdCb3g9IjAgMCA2MCA2MCIgeG1sbnM9Imh0dHA6Ly93d3cudzMub3JnLzIwMDAvc3ZnIj48ZyBmaWxsPSJub25lIiBmaWxsLXJ1bGU9ImV2ZW5vZGQiPjxnIGZpbGw9IiNmZmYiIGZpbGwtb3BhY2l0eT0iMC4xIj48Y2lyY2xlIGN4PSIzMCIgY3k9IjMwIiByPSIyIi8+PC9nPjwvZz48L3N2Zz4=')] opacity-30" />

      <Card className="w-full max-w-md relative animate-fade-in">
        <CardHeader className="text-center pb-2">
          <div className="flex justify-center mb-4">
            <div className="flex items-center justify-center w-16 h-16 rounded-2xl bg-gradient-to-br from-purple-600 to-indigo-600 shadow-xl shadow-purple-500/30">
              <Sparkles className="w-8 h-8 text-white" />
            </div>
          </div>
          <CardTitle className="text-2xl font-display">Mulai Gratis</CardTitle>
          <CardDescription>
            Buat akun Bantuaku untuk toko Anda
          </CardDescription>
        </CardHeader>
        <CardContent>
          <form onSubmit={handleSubmit} className="space-y-4">
            <div className="space-y-2">
              <Label htmlFor="storeName">Nama Toko</Label>
              <Input
                id="storeName"
                placeholder="Toko Berkah Jaya"
                value={storeName}
                onChange={(e) => setStoreName(e.target.value)}
                required
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="industry">Kategori Bisnis</Label>
              <Select value={industry} onValueChange={setIndustry}>
                <SelectTrigger>
                  <SelectValue placeholder="Pilih kategori..." />
                </SelectTrigger>
                <SelectContent>
                  {industries.map((ind) => (
                    <SelectItem key={ind.value} value={ind.value}>
                      {ind.label}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
            <div className="space-y-2">
              <Label htmlFor="email">Email</Label>
              <Input
                id="email"
                type="email"
                placeholder="nama@toko.com"
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                required
                autoComplete="email"
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="password">Password</Label>
              <Input
                id="password"
                type="password"
                placeholder="Minimal 6 karakter"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                required
                minLength={6}
                autoComplete="new-password"
              />
            </div>
            <Button type="submit" className="w-full" disabled={loading}>
              {loading ? (
                <>
                  <Loader2 className="w-4 h-4 mr-2 animate-spin" />
                  Memproses...
                </>
              ) : (
                'Daftar Gratis'
              )}
            </Button>
          </form>

          <div className="mt-6 text-center text-sm">
            <span className="text-slate-600">Sudah punya akun? </span>
            <Link
              to="/login"
              className="font-medium text-purple-600 hover:text-purple-700"
            >
              Masuk
            </Link>
          </div>
        </CardContent>
      </Card>
    </div>
  )
}
