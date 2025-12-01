import { useState } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { Sparkles, Loader2 } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { api } from '@/lib/api'
import { useAuthStore } from '@/state/auth'
import { toast } from '@/components/ui/toaster'

export function LoginPage() {
  const navigate = useNavigate()
  const login = useAuthStore((state) => state.login)
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [loading, setLoading] = useState(false)

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    setLoading(true)

    try {
      const data = await api.auth.login(email, password)
      login(data)
      toast({ title: 'Berhasil masuk!', variant: 'success' })
      navigate('/dashboard')
    } catch (err) {
      toast({
        title: 'Gagal masuk',
        description: err instanceof Error ? err.message : 'Email atau password salah',
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
          <CardTitle className="text-2xl font-display">Selamat Datang</CardTitle>
          <CardDescription>
            Masuk ke akun Bantuaku Anda
          </CardDescription>
        </CardHeader>
        <CardContent>
          <form onSubmit={handleSubmit} className="space-y-4">
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
                placeholder="••••••••"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                required
                autoComplete="current-password"
              />
            </div>
            <Button type="submit" className="w-full" disabled={loading}>
              {loading ? (
                <>
                  <Loader2 className="w-4 h-4 mr-2 animate-spin" />
                  Memproses...
                </>
              ) : (
                'Masuk'
              )}
            </Button>
          </form>

          <div className="mt-6 text-center text-sm">
            <span className="text-slate-600">Belum punya akun? </span>
            <Link
              to="/register"
              className="font-medium text-purple-600 hover:text-purple-700"
            >
              Daftar gratis
            </Link>
          </div>

          {/* Demo credentials hint */}
          <div className="mt-4 p-3 rounded-lg bg-slate-50 text-sm">
            <p className="font-medium text-slate-700">Demo Account:</p>
            <p className="text-slate-600">Email: demo@bantuaku.id</p>
            <p className="text-slate-600">Password: demo123</p>
          </div>
        </CardContent>
      </Card>
    </div>
  )
}
