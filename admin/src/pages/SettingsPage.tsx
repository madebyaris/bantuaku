import { useEffect, useState } from "react";
import { Save, Settings, ShieldAlert, Info } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Card } from "@/components/ui/card";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { api } from "@/lib/api";
import { useToast } from "@/components/ui/toaster";

export function SettingsPage() {
  const { toast } = useToast();
  const [provider, setProvider] = useState<string>("");
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);

  useEffect(() => {
    loadSettings();
  }, []);

  const loadSettings = async () => {
    try {
      setLoading(true);
      // #region agent log
      fetch('http://127.0.0.1:7242/ingest/caa1e494-1c2c-46ae-ab69-48afbc48a0f9',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify({location:'SettingsPage.tsx:loadSettings:entry',message:'loadSettings called',data:{},timestamp:Date.now(),sessionId:'debug-session',runId:'run1',hypothesisId:'A'})}).catch(()=>{});
      // #endregion
      const data = await api.admin.settings.getAIProvider();
      // #region agent log
      fetch('http://127.0.0.1:7242/ingest/caa1e494-1c2c-46ae-ab69-48afbc48a0f9',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify({location:'SettingsPage.tsx:loadSettings:success',message:'getAIProvider success',data:{provider:data.provider},timestamp:Date.now(),sessionId:'debug-session',runId:'run1',hypothesisId:'A'})}).catch(()=>{});
      // #endregion
      setProvider(data.provider || "openrouter");
    } catch (error) {
      // #region agent log
      fetch('http://127.0.0.1:7242/ingest/caa1e494-1c2c-46ae-ab69-48afbc48a0f9',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify({location:'SettingsPage.tsx:loadSettings:error',message:'getAIProvider failed',data:{error:error instanceof Error?error.message:String(error)},timestamp:Date.now(),sessionId:'debug-session',runId:'run1',hypothesisId:'A'})}).catch(()=>{});
      // #endregion
      console.error("Failed to load settings:", error);
      toast({
        title: "Error",
        description: "Gagal memuat pengaturan",
        variant: "destructive",
      });
      // Default to openrouter if load fails
      setProvider("openrouter");
    } finally {
      setLoading(false);
    }
  };

  const handleSave = async () => {
    if (!provider || (provider !== "openrouter" && provider !== "kolosal")) {
      toast({
        title: "Error",
        description: "Pilih provider yang valid",
        variant: "destructive",
      });
      return;
    }

    try {
      setSaving(true);
      await api.admin.settings.updateAIProvider(provider as "openrouter" | "kolosal");
      toast({
        title: "Success",
        description: "Pengaturan berhasil disimpan",
        variant: "success",
      });
    } catch (error) {
      console.error("Failed to save settings:", error);
      toast({
        title: "Error",
        description: "Gagal menyimpan pengaturan",
        variant: "destructive",
      });
    } finally {
      setSaving(false);
    }
  };

  return (
    <div className="space-y-6 animate-fade-in-up">
      <div>
        <h1 className="text-3xl font-display font-bold text-slate-100 mb-2">Pengaturan</h1>
        <p className="text-slate-400">Konfigurasi pengaturan aplikasi</p>
      </div>

      {/* Security Notice */}
      <Card className="p-4 border-amber-500/30 bg-amber-500/5 hover-card-effect">
        <div className="flex gap-3">
          <ShieldAlert className="h-5 w-5 text-amber-400 flex-shrink-0 mt-0.5" />
          <div>
            <h3 className="text-sm font-semibold text-amber-400 mb-1">Pengaturan Keamanan</h3>
            <p className="text-xs text-slate-400 leading-relaxed">
              API keys dan konfigurasi sensitif lainnya harus diubah melalui <code className="px-1.5 py-0.5 rounded bg-black/30 text-amber-300 font-mono text-[10px]">.env</code> file atau environment variables di server untuk alasan keamanan. 
              Pengaturan berikut yang perlu dikonfigurasi di level environment:
            </p>
            <ul className="mt-2 text-xs text-slate-500 space-y-1">
              <li className="flex items-center gap-2">
                <span className="w-1.5 h-1.5 rounded-full bg-amber-500/50"></span>
                <code className="text-slate-400 font-mono">OPENROUTER_API_KEY</code> - API key untuk OpenRouter
              </li>
              <li className="flex items-center gap-2">
                <span className="w-1.5 h-1.5 rounded-full bg-amber-500/50"></span>
                <code className="text-slate-400 font-mono">KOLOSAL_API_KEY</code> - API key untuk Kolosal
              </li>
              <li className="flex items-center gap-2">
                <span className="w-1.5 h-1.5 rounded-full bg-amber-500/50"></span>
                <code className="text-slate-400 font-mono">JWT_SECRET</code> - Secret key untuk authentication
              </li>
              <li className="flex items-center gap-2">
                <span className="w-1.5 h-1.5 rounded-full bg-amber-500/50"></span>
                <code className="text-slate-400 font-mono">DATABASE_URL</code> - Database connection string
              </li>
            </ul>
          </div>
        </div>
      </Card>

      <Card className="p-6 hover-card-effect">
        <div className="flex items-center gap-2 mb-4">
          <Settings className="h-5 w-5 text-emerald-400" />
          <h2 className="text-xl font-display font-semibold text-slate-100">AI Provider</h2>
        </div>
        <p className="text-sm text-slate-400 mb-4">
          Pilih provider AI yang akan digunakan untuk fitur chat. Perubahan akan diterapkan untuk semua percakapan baru.
        </p>
        <div className="flex items-start gap-2 p-3 rounded-lg bg-blue-500/10 border border-blue-500/20 mb-6">
          <Info className="h-4 w-4 text-blue-400 flex-shrink-0 mt-0.5" />
          <p className="text-xs text-blue-300">
            Pengaturan ini hanya memilih provider yang aktif. Pastikan API key untuk provider yang dipilih sudah dikonfigurasi di environment variables.
          </p>
        </div>

        {loading ? (
          <div className="text-slate-400">Memuat pengaturan...</div>
        ) : (
          <div className="space-y-4">
            <div className="space-y-2">
              <label className="text-sm font-medium text-slate-200">Provider</label>
              <Select value={provider} onValueChange={setProvider}>
                <SelectTrigger className="w-full max-w-md bg-white/5 border-white/10 text-slate-100 focus:border-emerald-500/50 focus:ring-emerald-500/20">
                  <SelectValue placeholder="Pilih provider" />
                </SelectTrigger>
                <SelectContent className="bg-[#0a0a0a] border-white/10">
                  <SelectItem value="openrouter" className="text-slate-100 focus:bg-emerald-500/20 focus:text-emerald-400">
                    OpenRouter
                  </SelectItem>
                  <SelectItem value="kolosal" className="text-slate-100 focus:bg-emerald-500/20 focus:text-emerald-400">
                    Kolosal
                  </SelectItem>
                </SelectContent>
              </Select>
              <p className="text-xs text-slate-500">
                {provider === "openrouter"
                  ? "OpenRouter menyediakan akses ke berbagai model AI seperti GPT-4, Claude, dan lainnya."
                  : "Kolosal adalah provider AI lokal dengan model GLM 4.6."}
              </p>
            </div>

            <div className="flex gap-2">
              <Button
                onClick={handleSave}
                disabled={saving || loading}
                className="flex items-center gap-2 bg-gradient-to-r from-emerald-600 to-emerald-400 hover:from-emerald-500 hover:to-emerald-300 text-black font-semibold"
              >
                <Save className="h-4 w-4" />
                {saving ? "Menyimpan..." : "Simpan"}
              </Button>
            </div>
          </div>
        )}
      </Card>
    </div>
  );
}
