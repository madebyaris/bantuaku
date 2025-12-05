import { useState, useRef, useEffect } from 'react'
import { Send, Loader2, Sparkles, User, Upload } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { cn } from '@/lib/utils'
import { useChatStore, ChatMessage } from '@/state/chat'

interface ChatInterfaceProps {
  isWidget?: boolean
  className?: string
}

const suggestedQuestions = [
  'Apa yang harus saya order bulan depan?',
  'Mengapa penjualan menurun minggu ini?',
  'Produk apa yang sedang trending?',
  'Bagaimana cara mengoptimalkan stok?',
]

export function ChatInterface({ isWidget = false, className }: ChatInterfaceProps) {
  const {
    messages,
    loading,
    currentConversationId,
    addMessage,
    setLoading,
    initializeConversation,
    loadMessages,
  } = useChatStore()
  const [input, setInput] = useState('')
  const messagesEndRef = useRef<HTMLDivElement>(null)
  const [initializing, setInitializing] = useState(true)

  // Initialize conversation on mount (single continuous chat)
  useEffect(() => {
    const init = async () => {
      try {
        await initializeConversation()
      } catch (err) {
        console.error('Failed to initialize conversation:', err)
      } finally {
        setInitializing(false)
      }
    }
    init()
  }, [initializeConversation])

  useEffect(() => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' })
  }, [messages])

  useEffect(() => {
    // Load UnicornStudio script for matrix effect
    // Only load if not already loaded or if we want to ensure it runs
    if (!document.querySelector('script[src*="unicornStudio.umd.js"]')) {
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
    } else {
        // @ts-ignore
        if (window.UnicornStudio) {
            // @ts-ignore
            window.UnicornStudio.init()
        }
    }
  }, [])

  async function sendMessage(text: string) {
    if (!text.trim() || loading || !currentConversationId) return

    const userMessage: ChatMessage = {
      id: Date.now().toString(),
      role: 'user',
      text: text.trim(),
      timestamp: new Date(),
    }

    addMessage(userMessage)
    setInput('')
    setLoading(true)

    try {
      const { api } = await import('@/lib/api')
      const response = await api.chat.sendMessage(currentConversationId, text.trim())
      
      const assistantMessage: ChatMessage = {
        id: (Date.now() + 1).toString(),
        role: 'assistant',
        text: response.assistant_reply,
        citations: response.citations,
        timestamp: new Date(),
      }

      addMessage(assistantMessage)
    } catch (err) {
      console.error('Failed to send message:', err)
      const errorMessage: ChatMessage = {
        id: (Date.now() + 1).toString(),
        role: 'assistant',
        text: 'Maaf, terjadi kesalahan. Silakan coba lagi.',
        timestamp: new Date(),
      }
      addMessage(errorMessage)
    } finally {
      setLoading(false)
    }
  }

  function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    sendMessage(input)
  }

  if (initializing) {
    return (
      <div className={cn("flex flex-col items-center justify-center h-full", className)}>
        <Loader2 className="w-8 h-8 animate-spin text-emerald-400" />
        <p className="mt-4 text-slate-400">Memuat chat...</p>
      </div>
    )
  }

  return (
    <div className={cn("flex flex-col relative overflow-hidden", isWidget ? "h-full" : "h-[calc(100vh-12rem)] rounded-xl", className)}>
      {/* Matrix Background with Blur Overlay */}
      <div className="absolute inset-0 -z-10 pointer-events-none">
        <div className="absolute inset-0" data-us-project="EET25BiXxR2StNXZvAzF"></div>
        <div className="absolute inset-0 bg-black/80 backdrop-blur-[2px]"></div>
      </div>

      {/* Chat Messages */}
      <div className="flex-1 overflow-y-auto px-4 py-6 space-y-4 relative z-10">
        {messages.length === 0 ? (
          <div className="flex flex-col items-center justify-center h-full text-center">
            <div className={cn("p-4 bg-gradient-to-br from-emerald-400 to-emerald-600 rounded-full shadow-[0_0_20px_rgba(16,185,129,0.3)] mb-6", isWidget && "p-3 mb-4")}>
              <Sparkles className={cn("text-black fill-black", isWidget ? "w-6 h-6" : "w-10 h-10")} />
            </div>
            <h3 className={cn("font-display font-bold text-slate-100 mb-2", isWidget ? "text-lg" : "text-2xl")}>
              Halo! Saya Asisten Bantuaku
            </h3>
            {!isWidget && (
                <p className="text-slate-400 max-w-md mb-8">
                Tanyakan apa saja tentang bisnis Anda. Saya bisa membantu dengan
                forecasting, rekomendasi stok, dan analisis penjualan.
                </p>
            )}
            
            {/* Suggested Questions */}
            <div className={cn("grid gap-3 w-full", isWidget ? "grid-cols-1" : "grid-cols-1 sm:grid-cols-2 max-w-lg")}>
              {suggestedQuestions.slice(0, isWidget ? 2 : 4).map((question, i) => (
                <button
                  key={i}
                  onClick={() => sendMessage(question)}
                  className="p-3 text-left text-sm bg-white/5 border border-white/10 rounded-lg hover:border-emerald-500/50 hover:bg-white/10 text-slate-300 transition-all backdrop-blur-sm"
                >
                  {question}
                </button>
              ))}
            </div>
          </div>
        ) : (
          messages.map((message) => (
            <div
              key={message.id}
              className={cn(
                'flex gap-3 chat-message-enter',
                message.role === 'user' ? 'justify-end' : 'justify-start'
              )}
            >
              {message.role === 'assistant' && (
                <div className="flex-shrink-0 w-8 h-8 rounded-full bg-gradient-to-br from-emerald-400 to-emerald-600 flex items-center justify-center shadow-[0_0_10px_rgba(16,185,129,0.3)]">
                  <Sparkles className="w-4 h-4 text-black fill-black" />
                </div>
              )}
              
              <div
                className={cn(
                  'max-w-[85%] rounded-2xl px-4 py-3 backdrop-blur-md',
                  message.role === 'user'
                    ? 'bg-emerald-600 text-white shadow-[0_0_15px_rgba(16,185,129,0.2)]'
                    : 'bg-white/5 border border-white/10 text-slate-200 shadow-sm'
                )}
              >
                <div
                  className={cn(
                    'text-sm whitespace-pre-wrap',
                    message.role === 'assistant' && 'text-slate-200'
                  )}
                >
                  {message.text}
                </div>
                
                {message.confidence && (
                  <div className="mt-2 pt-2 border-t border-white/10 flex items-center gap-4 text-xs text-slate-400">
                    <span>
                      Confidence: {(message.confidence * 100).toFixed(0)}%
                    </span>
                  </div>
                )}
                {message.citations && message.citations.length > 0 && (
                  <div className="mt-2 pt-2 border-t border-white/10 space-y-2">
                    <p className="text-[11px] uppercase tracking-wide text-slate-500">Sumber</p>
                    <ul className="space-y-1">
                      {message.citations.map((c, idx) => (
                        <li key={idx} className="text-xs text-slate-300">
                          • {c.text} <span className="text-slate-500">({c.source})</span>
                        </li>
                      ))}
                    </ul>
                  </div>
                )}
              </div>

              {message.role === 'user' && (
                <div className="flex-shrink-0 w-8 h-8 rounded-full bg-white/10 flex items-center justify-center border border-white/10">
                  <User className="w-4 h-4 text-slate-300" />
                </div>
              )}
            </div>
          ))
        )}
        
        {loading && (
          <div className="flex gap-3 justify-start">
            <div className="flex-shrink-0 w-8 h-8 rounded-full bg-gradient-to-br from-emerald-400 to-emerald-600 flex items-center justify-center shadow-[0_0_10px_rgba(16,185,129,0.3)]">
              <Sparkles className="w-4 h-4 text-black fill-black" />
            </div>
            <div className="bg-white/5 border border-white/10 shadow-sm rounded-2xl px-4 py-3 backdrop-blur-md">
              <div className="flex items-center gap-2 text-slate-400">
                <Loader2 className="w-4 h-4 animate-spin text-emerald-400" />
                <span className="text-sm">Menganalisis...</span>
              </div>
            </div>
          </div>
        )}
        
        <div ref={messagesEndRef} />
      </div>

      {/* Input */}
      <div className={cn("border-t border-white/10 bg-black/50 backdrop-blur-xl z-20", isWidget ? "p-3" : "p-4")}>
        <form onSubmit={handleSubmit} className="flex gap-3 max-w-3xl mx-auto">
          <Button 
            type="button" 
            variant="outline" 
            size="icon"
            className="shrink-0 border-white/10 bg-white/5 text-slate-400 hover:text-emerald-400 hover:border-emerald-500/30 hover:bg-emerald-500/10"
          >
            <Upload className="w-4 h-4" />
          </Button>
          <Input
            value={input}
            onChange={(e) => setInput(e.target.value)}
            placeholder="Ketik pertanyaan..."
            disabled={loading || !currentConversationId}
            className="flex-1 bg-white/5 border-white/10 text-slate-100 placeholder:text-slate-500 focus:border-emerald-500/50 focus:ring-emerald-500/20"
            autoFocus={!isWidget}
          />
          <Button type="submit" disabled={loading || !input.trim() || !currentConversationId} className="bg-emerald-500 hover:bg-emerald-400 text-black">
            {loading ? (
              <Loader2 className="w-4 h-4 animate-spin" />
            ) : (
              <Send className="w-4 h-4" />
            )}
          </Button>
        </form>
        {!isWidget && (
          <p className="text-xs text-slate-500 text-center mt-2">
            AI dapat membuat kesalahan. Verifikasi informasi penting.
          </p>
        )}
      </div>
    </div>
  )
}
