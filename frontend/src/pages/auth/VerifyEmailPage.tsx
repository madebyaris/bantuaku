import { useState } from 'react'
import { useNavigate, useSearchParams } from 'react-router-dom'
import { Mail, Loader2, CheckCircle2 } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { api } from '@/lib/api'
import { toast } from '@/components/ui/toaster'

export function VerifyEmailPage() {
  const navigate = useNavigate()
  const [searchParams] = useSearchParams()
  const emailFromParams = searchParams.get('email') || ''
  
  const [email, setEmail] = useState(emailFromParams)
  const [otp, setOtp] = useState('')
  const [loading, setLoading] = useState(false)
  const [resending, setResending] = useState(false)
  const [verified, setVerified] = useState(false)

  // Validate OTP input (only numbers, exactly 5 digits)
  function handleOtpChange(e: React.ChangeEvent<HTMLInputElement>) {
    const value = e.target.value.replace(/\D/g, '').slice(0, 5)
    setOtp(value)
  }

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    
    if (!email) {
      toast({
        title: 'Email diperlukan',
        description: 'Silakan masukkan email Anda',
        variant: 'destructive',
      })
      return
    }

    if (otp.length !== 5) {
      toast({
        title: 'Kode tidak valid',
        description: 'Kode verifikasi harus 5 digit',
        variant: 'destructive',
      })
      return
    }

    setLoading(true)

    try {
      await api.auth.verifyEmail(email, otp)
      setVerified(true)
      toast({ title: 'Email berhasil diverifikasi!', variant: 'success' })
      
      // Redirect to login after 2 seconds
      setTimeout(() => {
        navigate('/login')
      }, 2000)
    } catch (err) {
      toast({
        title: 'Verifikasi gagal',
        description: err instanceof Error ? err.message : 'Kode verifikasi tidak valid atau sudah kedaluwarsa',
        variant: 'destructive',
      })
    } finally {
      setLoading(false)
    }
  }

  async function handleResend() {
    if (!email) {
      toast({
        title: 'Email diperlukan',
        description: 'Silakan masukkan email Anda',
        variant: 'destructive',
      })
      return
    }

    setResending(true)

    try {
      await api.auth.resendVerification(email)
      toast({ 
        title: 'Email verifikasi dikirim ulang', 
        description: 'Silakan periksa inbox email Anda',
        variant: 'success' 
      })
    } catch (err) {
      toast({
        title: 'Gagal mengirim email',
        description: err instanceof Error ? err.message : 'Terjadi kesalahan saat mengirim email',
        variant: 'destructive',
      })
    } finally {
      setResending(false)
    }
  }

  if (verified) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-black p-4">
        <div className="fixed inset-0 bg-[radial-gradient(ellipse_at_top,_var(--tw-gradient-stops))] from-emerald-900/20 via-black to-black pointer-events-none" />
        
        <Card className="w-full max-w-md relative animate-fade-in border-white/10 bg-white/5 backdrop-blur-xl">
          <CardHeader className="text-center pb-2">
            <div className="flex justify-center mb-4">
              <div className="flex items-center justify-center w-16 h-16 rounded-full bg-gradient-to-br from-emerald-400 to-emerald-600 shadow-[0_0_20px_rgba(16,185,129,0.3)]">
                <CheckCircle2 className="w-8 h-8 text-black fill-black" />
              </div>
            </div>
            <CardTitle className="text-2xl font-display text-slate-100">Email Terverifikasi!</CardTitle>
            <CardDescription className="text-slate-400">
              Mengarahkan ke halaman login...
            </CardDescription>
          </CardHeader>
        </Card>
      </div>
    )
  }

  return (
    <div className="min-h-screen flex items-center justify-center bg-black p-4">
      <div className="fixed inset-0 bg-[radial-gradient(ellipse_at_top,_var(--tw-gradient-stops))] from-emerald-900/20 via-black to-black pointer-events-none" />

      <Card className="w-full max-w-md relative animate-fade-in border-white/10 bg-white/5 backdrop-blur-xl">
        <CardHeader className="text-center pb-2">
          <div className="flex justify-center mb-4">
            <div className="flex items-center justify-center w-16 h-16 rounded-full bg-gradient-to-br from-emerald-400 to-emerald-600 shadow-[0_0_20px_rgba(16,185,129,0.3)]">
              <Mail className="w-8 h-8 text-black fill-black" />
            </div>
          </div>
          <CardTitle className="text-2xl font-display text-slate-100">Verifikasi Email</CardTitle>
          <CardDescription className="text-slate-400">
            Masukkan kode 5 digit yang dikirim ke email Anda
          </CardDescription>
        </CardHeader>
        <CardContent>
          <form onSubmit={handleSubmit} className="space-y-4">
            <div className="space-y-2">
              <Label htmlFor="email" className="text-slate-200">Email</Label>
              <Input
                id="email"
                type="email"
                placeholder="nama@toko.com"
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                required
                className="bg-white/5 border-white/10 text-slate-100 placeholder:text-slate-500 focus:border-emerald-500/50 focus:ring-emerald-500/20"
              />
            </div>
            
            <div className="space-y-2">
              <Label htmlFor="otp" className="text-slate-200">Kode Verifikasi (5 digit)</Label>
              <Input
                id="otp"
                type="text"
                inputMode="numeric"
                placeholder="12345"
                value={otp}
                onChange={handleOtpChange}
                required
                maxLength={5}
                className="bg-white/5 border-white/10 text-slate-100 placeholder:text-slate-500 focus:border-emerald-500/50 focus:ring-emerald-500/20 text-center text-2xl font-mono tracking-widest"
              />
              <p className="text-xs text-slate-400 text-center">
                Kode akan kedaluwarsa dalam 1 jam
              </p>
            </div>

            <Button 
              type="submit" 
              className="w-full bg-gradient-to-r from-emerald-600 to-emerald-400 hover:from-emerald-500 hover:to-emerald-300 text-black font-semibold border-0" 
              disabled={loading || otp.length !== 5}
            >
              {loading ? (
                <>
                  <Loader2 className="w-4 h-4 mr-2 animate-spin" />
                  Memverifikasi...
                </>
              ) : (
                'Verifikasi Email'
              )}
            </Button>
          </form>

          <div className="mt-6 space-y-3">
            <Button
              type="button"
              variant="outline"
              onClick={handleResend}
              disabled={resending || !email}
              className="w-full border-white/10 text-slate-200 hover:bg-white/10"
            >
              {resending ? (
                <>
                  <Loader2 className="w-4 h-4 mr-2 animate-spin" />
                  Mengirim...
                </>
              ) : (
                'Kirim Ulang Kode'
              )}
            </Button>
            
            <p className="text-xs text-slate-400 text-center">
              Tidak menerima email? Periksa folder spam atau{' '}
              <button
                onClick={handleResend}
                disabled={resending}
                className="text-emerald-400 hover:text-emerald-300 underline"
              >
                kirim ulang
              </button>
            </p>
          </div>
        </CardContent>
      </Card>
    </div>
  )
}

