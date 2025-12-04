import { useState, useEffect } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { Sparkles, Loader2, Zap, CheckCircle2, ShieldCheck, BarChart3 } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { api } from '@/lib/api'
import { useAuthStore } from '@/state/auth'
import { toast } from '@/components/ui/toaster'

export function LoginPage() {
  const navigate = useNavigate()
  const login = useAuthStore((state) => state.login)
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [loading, setLoading] = useState(false)

  useEffect(() => {
    // Load UnicornStudio script for matrix/tech effect
    const script = document.createElement('script')
    script.src = "https://cdn.jsdelivr.net/gh/hiunicornstudio/unicornstudio.js@v1.4.29/dist/unicornStudio.umd.js"
    script.onload = () => {
      // @ts-ignore
      if (window.UnicornStudio) {
        // @ts-ignore
        window.UnicornStudio.init()
      }
    }
    document.head.appendChild(script)

    return () => {
      document.head.removeChild(script)
    }
  }, [])

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    setLoading(true)

    try {
      const data = await api.auth.login(email, password)
      login(data)
      toast({ title: 'Berhasil masuk!', variant: 'success' })
      navigate('/dashboard')
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : 'Email atau password salah'
      const isEmailNotVerified = errorMessage.toLowerCase().includes('email not verified') || 
                                 errorMessage.toLowerCase().includes('belum diverifikasi')
      
      toast({
        title: 'Gagal masuk',
        description: errorMessage,
        variant: 'destructive',
      })

      // If email not verified, show additional message
      if (isEmailNotVerified) {
        setTimeout(() => {
          toast({
            title: 'Email belum diverifikasi',
            description: 'Silakan verifikasi email Anda terlebih dahulu. Periksa inbox untuk kode 5 digit.',
            variant: 'default',
          })
        }, 500)
      }
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
      <div className="w-full max-w-5xl grid grid-cols-1 lg:grid-cols-2 gap-0 rounded-3xl overflow-hidden border border-white/10 bg-black/40 backdrop-blur-2xl shadow-[0_0_50px_-12px_rgba(16,185,129,0.25)] animate-fade-in-up z-10 relative">

        {/* Left Side: Login Form */}
        <div className="p-8 lg:p-12 flex flex-col justify-center relative z-20 bg-black/20 border-r border-white/5">
          <div className="mb-8">
            <div className="flex items-center gap-3 mb-6">
              <div className="flex items-center justify-center w-10 h-10 rounded-xl bg-gradient-to-br from-emerald-400 to-emerald-600 shadow-[0_0_15px_rgba(16,185,129,0.4)]">
                <Zap className="w-6 h-6 text-black fill-black" />
              </div>
              <span className="font-display text-2xl font-bold text-slate-100 tracking-tight">Bantuaku</span>
            </div>
            <h1 className="text-3xl font-display font-bold text-slate-100 mb-2">Welcome Back</h1>
            <p className="text-slate-400">Masuk untuk melanjutkan akses dashboard Anda.</p>
          </div>

          <form onSubmit={handleSubmit} className="space-y-5">
            <div className="space-y-2">
              <Label htmlFor="email" className="text-slate-300">Email Address</Label>
              <Input
                id="email"
                type="email"
                placeholder="nama@perusahaan.com"
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                required
                autoComplete="email"
                className="h-12 bg-white/5 border-white/10 text-slate-100 placeholder:text-slate-600 focus:border-emerald-500/50 focus:ring-emerald-500/20 rounded-xl transition-all hover:bg-white/10"
              />
            </div>
            <div className="space-y-2">
              <div className="flex items-center justify-between">
                <Label htmlFor="password" className="text-slate-300">Password</Label>
                <Link to="/forgot-password" className="text-xs font-medium text-emerald-400 hover:text-emerald-300 transition-colors">Forgot password?</Link>
              </div>
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

          <div className="mt-8 text-center">
            <p className="text-sm text-slate-500">
              Belum punya akun?{' '}
              <Link
                to="/register"
                className="font-semibold text-emerald-400 hover:text-emerald-300 transition-colors"
              >
                Daftar sekarang
              </Link>
            </p>
          </div>

          {/* Demo Hint */}
          <div className="mt-8 p-4 rounded-xl bg-emerald-900/10 border border-emerald-500/10 backdrop-blur-sm">
            <div className="flex items-center gap-2 mb-2">
              <Sparkles className="w-4 h-4 text-emerald-400" />
              <span className="text-sm font-semibold text-emerald-400">Demo Account Access</span>
            </div>
            <div className="grid grid-cols-2 gap-4 text-sm">
              <div>
                <span className="block text-xs text-slate-500 mb-0.5">Email</span>
                <code className="text-slate-300 bg-white/5 px-2 py-1 rounded">demo@bantuaku.id</code>
              </div>
              <div>
                <span className="block text-xs text-slate-500 mb-0.5">Password</span>
                <code className="text-slate-300 bg-white/5 px-2 py-1 rounded">demo123</code>
              </div>
            </div>
          </div>
        </div>

        {/* Right Side: Tech Visuals */}
        <div className="hidden lg:flex flex-col justify-center p-12 relative overflow-hidden bg-black/40">
           {/* Tech Background - Matrix Effect */}
           <div className="absolute inset-0 opacity-80 mix-blend-screen scale-110" data-us-project="EET25BiXxR2StNXZvAzF"></div>
           <div className="absolute inset-0 bg-gradient-to-t from-black/60 via-transparent to-black/60 pointer-events-none" />
           
           <div className="relative z-10">
              <div className="space-y-6 max-w-md mx-auto">
                <div className="inline-flex items-center px-3 py-1 rounded-full bg-emerald-500/10 border border-emerald-500/20 backdrop-blur-md">
                  <span className="w-2 h-2 rounded-full bg-emerald-400 animate-pulse mr-2"></span>
                  <span className="text-xs font-medium text-emerald-400 tracking-wide uppercase">AI-Powered Forecasting</span>
                </div>
                
                <h2 className="text-4xl font-display font-bold text-transparent bg-clip-text bg-gradient-to-br from-white to-slate-400 leading-tight">
                  Unlock the Future of Your Business
                </h2>
                
                <div className="space-y-4 pt-2">
                  <div className="flex items-start gap-3 text-slate-300">
                    <div className="mt-1 p-1 rounded-full bg-emerald-500/10 text-emerald-400">
                      <BarChart3 className="w-4 h-4" />
                    </div>
                    <p className="text-sm leading-relaxed">Real-time sales forecasting using advanced predictive models tailored for Indonesian markets.</p>
                  </div>
                  <div className="flex items-start gap-3 text-slate-300">
                    <div className="mt-1 p-1 rounded-full bg-emerald-500/10 text-emerald-400">
                      <ShieldCheck className="w-4 h-4" />
                    </div>
                    <p className="text-sm leading-relaxed">Enterprise-grade security with automated regulation compliance monitoring.</p>
                  </div>
                </div>

                <div className="pt-8 flex items-center gap-4">
                  <div className="flex -space-x-3">
                    {[1, 2, 3, 4].map((i) => (
                      <div key={i} className="w-10 h-10 rounded-full border-2 border-black bg-slate-800 flex items-center justify-center overflow-hidden">
                        <img src={`https://api.dicebear.com/7.x/avataaars/svg?seed=${i+20}`} alt="User" className="w-full h-full" />
                      </div>
                    ))}
                  </div>
                  <div>
                    <div className="flex items-center gap-1">
                      {[1, 2, 3, 4, 5].map((i) => (
                        <Sparkles key={i} className="w-3 h-3 text-emerald-400 fill-emerald-400" />
                      ))}
                    </div>
                    <p className="text-xs text-slate-400 mt-1"><span className="text-slate-200 font-bold">1,000+</span> UMKM joined</p>
                  </div>
                </div>
              </div>
           </div>
        </div>
      </div>
    </div>
  )
}
