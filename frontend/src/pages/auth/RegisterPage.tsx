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
    <div className="min-h-screen flex items-center justify-center bg-black p-4">
      {/* Background pattern */}
      <div className="fixed inset-0 bg-[radial-gradient(ellipse_at_top,_var(--tw-gradient-stops))] from-emerald-900/20 via-black to-black pointer-events-none" />

      <Card className="w-full max-w-md relative animate-fade-in border-white/10 bg-white/5 backdrop-blur-xl">
        <CardHeader className="text-center pb-2">
          <div className="flex justify-center mb-4">
            <div className="flex items-center justify-center w-16 h-16 rounded-full bg-gradient-to-br from-emerald-400 to-emerald-600 shadow-[0_0_20px_rgba(16,185,129,0.3)]">
              <Sparkles className="w-8 h-8 text-black fill-black" />
            </div>
          </div>
          <CardTitle className="text-2xl font-display text-slate-100">Mulai Gratis</CardTitle>
          <CardDescription className="text-slate-400">
            Buat akun Bantuaku untuk toko Anda
          </CardDescription>
        </CardHeader>
        <CardContent>
          <form onSubmit={handleSubmit} className="space-y-4">
            <div className="space-y-2">
              <Label htmlFor="storeName" className="text-slate-200">Nama Toko</Label>
              <Input
                id="storeName"
                placeholder="Toko Berkah Jaya"
                value={storeName}
                onChange={(e) => setStoreName(e.target.value)}
                required
                className="bg-white/5 border-white/10 text-slate-100 placeholder:text-slate-500 focus:border-emerald-500/50 focus:ring-emerald-500/20"
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="industry" className="text-slate-200">Kategori Bisnis</Label>
              <Select value={industry} onValueChange={setIndustry}>
                <SelectTrigger className="bg-white/5 border-white/10 text-slate-100 focus:border-emerald-500/50 focus:ring-emerald-500/20">
                  <SelectValue placeholder="Pilih kategori..." />
                </SelectTrigger>
                <SelectContent className="bg-black/90 border-white/10 text-slate-100">
                  {industries.map((ind) => (
                    <SelectItem key={ind.value} value={ind.value} className="focus:bg-emerald-500/20 focus:text-emerald-400">
                      {ind.label}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
            <div className="space-y-2">
              <Label htmlFor="email" className="text-slate-200">Email</Label>
              <Input
                id="email"
                type="email"
                placeholder="nama@toko.com"
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                required
                autoComplete="email"
                className="bg-white/5 border-white/10 text-slate-100 placeholder:text-slate-500 focus:border-emerald-500/50 focus:ring-emerald-500/20"
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="password" classname="text-slate-200">Password</Label>
              <Input
                id="password"
                type="password"
                placeholder="Minimal 6 karakter"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                required
                minLength={6}
                autoComplete="new-password"
                className="bg-white/5 border-white/10 text-slate-100 placeholder:text-slate-500 focus:border-emerald-500/50 focus:ring-emerald-500/20"
              />
            </div>
            <Button type="submit" className="w-full bg-gradient-to-r from-emerald-600 to-emerald-400 hover:from-emerald-500 hover:to-emerald-300 text-black font-semibold border-0" disabled={loading}>
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
            <span className="text-slate-400">Sudah punya akun? </span>
            <Link
              to="/login"
              className="font-medium text-emerald-400 hover:text-emerald-300 transition-colors"
            >
              Masuk
            </Link>
          </div>
        </CardContent>
      </Card>
    </div>
  )
}
