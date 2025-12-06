import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { Loader2, Shield } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { useAuthStore } from '@/state/auth'
import { api } from '@/lib/api'
import { toast } from '@/components/ui/toaster'

export function LoginPage() {
  const navigate = useNavigate()
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [loading, setLoading] = useState(false)
  const { login } = useAuthStore()

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    setLoading(true)

    try {
      const response = await api.auth.login(email, password)
      
      // Decode JWT to get role
      const decoded = useAuthStore.getState().decodeToken(response.token)
      const role = decoded?.role || 'user'
      
      // Check if user has admin role
      if (role !== 'admin' && role !== 'super_admin') {
        toast({
          title: 'Access Denied',
          description: 'Only administrators can access this panel',
          variant: 'destructive',
        })
        return
      }
      
      login(response.token, {
        id: response.user_id,
        email: email,
        role: role,
      })
      
      toast({
        title: 'Login berhasil',
        description: 'Selamat datang di Admin Panel',
        variant: 'success',
      })
      
      navigate('/dashboard')
    } catch (error) {
      toast({
        title: 'Login gagal',
        description: error instanceof Error ? error.message : 'Email atau password salah',
        variant: 'destructive',
      })
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="min-h-screen w-full flex items-center justify-center bg-black p-4 relative overflow-hidden font-sans">
      {/* Background Gradients */}
      <div className="absolute inset-0 bg-[radial-gradient(circle_at_top_right,_var(--tw-gradient-stops))] from-emerald-900/20 via-black to-black pointer-events-none" />
      <div className="absolute bottom-0 left-0 w-full h-1/2 bg-gradient-to-t from-emerald-900/10 via-transparent to-transparent pointer-events-none" />

      {/* Main Card Container */}
      <div className="w-full max-w-md rounded-3xl overflow-hidden border border-white/10 bg-black/40 backdrop-blur-2xl shadow-[0_0_50px_-12px_rgba(16,185,129,0.25)] animate-fade-in-up z-10 relative">
        <div className="p-8 lg:p-12 flex flex-col justify-center relative z-20 bg-black/20">
          <div className="mb-8">
            <div className="flex items-center gap-3 mb-6">
              <div className="flex items-center justify-center w-10 h-10 rounded-xl bg-gradient-to-br from-emerald-400 to-emerald-600 shadow-[0_0_15px_rgba(16,185,129,0.4)]">
                <Shield className="w-6 h-6 text-black fill-black" />
              </div>
              <span className="font-display text-2xl font-bold text-slate-100 tracking-tight">Bantuaku Admin</span>
            </div>
            <h1 className="text-3xl font-display font-bold text-slate-100 mb-2">Admin Access</h1>
            <p className="text-slate-400">Masuk dengan akun admin untuk mengelola sistem.</p>
          </div>

          <form onSubmit={handleSubmit} className="space-y-5">
            <div className="space-y-2">
              <Label htmlFor="email" className="text-slate-300">Email Address</Label>
              <Input
                id="email"
                type="email"
                placeholder="admin@bantuaku.id"
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                required
                autoComplete="email"
                className="h-12 bg-white/5 border-white/10 text-slate-100 placeholder:text-slate-600 focus:border-emerald-500/50 focus:ring-emerald-500/20 rounded-xl transition-all hover:bg-white/10"
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="password" className="text-slate-300">Password</Label>
              <Input
                id="password"
                type="password"
                placeholder="••••••••"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                required
                autoComplete="current-password"
                className="h-12 bg-white/5 border-white/10 text-slate-100 placeholder:text-slate-600 focus:border-emerald-500/50 focus:ring-emerald-500/20 rounded-xl transition-all hover:bg-white/10"
              />
            </div>

            <Button 
              type="submit" 
              className="w-full h-12 bg-gradient-to-r from-emerald-600 to-emerald-400 hover:from-emerald-500 hover:to-emerald-300 text-black font-bold text-base rounded-xl shadow-[0_0_20px_rgba(16,185,129,0.3)] hover:shadow-[0_0_30px_rgba(16,185,129,0.5)] transition-all duration-300 border-0 mt-2" 
              disabled={loading}
            >
              {loading ? (
                <>
                  <Loader2 className="w-5 h-5 mr-2 animate-spin" />
                  Processing...
                </>
              ) : (
                'Sign In'
              )}
            </Button>
          </form>

          {/* Demo Account Hint */}
          <div className="mt-8 p-4 rounded-xl bg-emerald-900/10 border border-emerald-500/10 backdrop-blur-sm">
            <div className="flex items-center gap-2 mb-2">
              <Shield className="w-4 h-4 text-emerald-400" />
              <span className="text-sm font-semibold text-emerald-400">Demo Admin Account</span>
            </div>
            <div className="grid grid-cols-2 gap-4 text-sm">
              <div>
                <span className="block text-xs text-slate-500 mb-0.5">Email</span>
                <code className="text-emerald-300 bg-white/5 px-2 py-1 rounded text-xs">admin@bantuaku.id</code>
              </div>
              <div>
                <span className="block text-xs text-slate-500 mb-0.5">Password</span>
                <code className="text-emerald-300 bg-white/5 px-2 py-1 rounded text-xs">demo123</code>
              </div>
            </div>
            <p className="text-xs text-slate-400 mt-3">
              Role: <span className="text-emerald-400 font-semibold">super_admin</span> - Full access to all admin features
            </p>
          </div>

          {/* Admin Access Notice */}
          <div className="mt-4 p-4 rounded-xl bg-amber-900/10 border border-amber-500/10 backdrop-blur-sm">
            <div className="flex items-center gap-2 mb-2">
              <Shield className="w-4 h-4 text-amber-400" />
              <span className="text-sm font-semibold text-amber-400">Admin Access Only</span>
            </div>
            <p className="text-xs text-slate-400">
              Hanya pengguna dengan role admin atau super_admin yang dapat mengakses panel ini.
            </p>
          </div>
        </div>
      </div>
    </div>
  )
}

