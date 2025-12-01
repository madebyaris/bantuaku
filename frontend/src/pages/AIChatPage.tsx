import { useState, useRef, useEffect } from 'react'
import { Send, Loader2, Sparkles, User } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { api } from '@/lib/api'
import { cn } from '@/lib/utils'

interface ChatMessage {
  id: string
  role: 'user' | 'assistant'
  text: string
  confidence?: number
  dataSources?: string[]
  timestamp: Date
}

const suggestedQuestions = [
  'Apa yang harus saya order bulan depan?',
  'Mengapa penjualan menurun minggu ini?',
  'Produk apa yang sedang trending?',
  'Bagaimana cara mengoptimalkan stok?',
]

export function AIChatPage() {
  const [messages, setMessages] = useState<ChatMessage[]>([])
  const [input, setInput] = useState('')
  const [loading, setLoading] = useState(false)
  const messagesEndRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' })
  }, [messages])

  async function sendMessage(text: string) {
    if (!text.trim() || loading) return

    const userMessage: ChatMessage = {
      id: Date.now().toString(),
      role: 'user',
      text: text.trim(),
      timestamp: new Date(),
    }

    setMessages((prev) => [...prev, userMessage])
    setInput('')
    setLoading(true)

    try {
      const response = await api.ai.analyze(text)
      
      const assistantMessage: ChatMessage = {
        id: (Date.now() + 1).toString(),
        role: 'assistant',
        text: response.answer,
        confidence: response.confidence,
        dataSources: response.data_sources,
        timestamp: new Date(),
      }

      setMessages((prev) => [...prev, assistantMessage])
    } catch (err) {
      const errorMessage: ChatMessage = {
        id: (Date.now() + 1).toString(),
        role: 'assistant',
        text: 'Maaf, terjadi kesalahan. Silakan coba lagi.',
        timestamp: new Date(),
      }
      setMessages((prev) => [...prev, errorMessage])
    } finally {
      setLoading(false)
    }
  }

  function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    sendMessage(input)
  }

  return (
    <div className="flex flex-col h-[calc(100vh-12rem)] animate-fade-in">
      {/* Chat Messages */}
      <div className="flex-1 overflow-y-auto px-4 py-6 space-y-4">
        {messages.length === 0 ? (
          <div className="flex flex-col items-center justify-center h-full text-center">
            <div className="p-4 bg-gradient-to-br from-purple-600 to-indigo-600 rounded-2xl shadow-xl shadow-purple-500/20 mb-6">
              <Sparkles className="w-10 h-10 text-white" />
            </div>
            <h3 className="text-2xl font-display font-bold text-slate-900 mb-2">
              Halo! Saya Asisten Bantuaku
            </h3>
            <p className="text-slate-500 max-w-md mb-8">
              Tanyakan apa saja tentang bisnis Anda. Saya bisa membantu dengan
              forecasting, rekomendasi stok, dan analisis penjualan.
            </p>
            
            {/* Suggested Questions */}
            <div className="grid grid-cols-1 sm:grid-cols-2 gap-3 w-full max-w-lg">
              {suggestedQuestions.map((question, i) => (
                <button
                  key={i}
                  onClick={() => sendMessage(question)}
                  className="p-3 text-left text-sm bg-white border border-slate-200 rounded-lg hover:border-purple-300 hover:bg-purple-50 transition-colors"
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
                <div className="flex-shrink-0 w-8 h-8 rounded-full bg-gradient-to-br from-purple-600 to-indigo-600 flex items-center justify-center">
                  <Sparkles className="w-4 h-4 text-white" />
                </div>
              )}
              
              <div
                className={cn(
                  'max-w-[70%] rounded-2xl px-4 py-3',
                  message.role === 'user'
                    ? 'bg-purple-600 text-white'
                    : 'bg-white border border-slate-200 shadow-sm'
                )}
              >
                <div
                  className={cn(
                    'text-sm whitespace-pre-wrap',
                    message.role === 'assistant' && 'text-slate-700'
                  )}
                >
                  {message.text}
                </div>
                
                {message.confidence && (
                  <div className="mt-2 pt-2 border-t border-slate-100 flex items-center gap-4 text-xs text-slate-500">
                    <span>
                      Confidence: {(message.confidence * 100).toFixed(0)}%
                    </span>
                    {message.dataSources && message.dataSources.length > 0 && (
                      <span>
                        Sumber: {message.dataSources.join(', ')}
                      </span>
                    )}
                  </div>
                )}
              </div>

              {message.role === 'user' && (
                <div className="flex-shrink-0 w-8 h-8 rounded-full bg-slate-200 flex items-center justify-center">
                  <User className="w-4 h-4 text-slate-600" />
                </div>
              )}
            </div>
          ))
        )}
        
        {loading && (
          <div className="flex gap-3 justify-start">
            <div className="flex-shrink-0 w-8 h-8 rounded-full bg-gradient-to-br from-purple-600 to-indigo-600 flex items-center justify-center">
              <Sparkles className="w-4 h-4 text-white" />
            </div>
            <div className="bg-white border border-slate-200 shadow-sm rounded-2xl px-4 py-3">
              <div className="flex items-center gap-2 text-slate-500">
                <Loader2 className="w-4 h-4 animate-spin" />
                <span className="text-sm">Menganalisis...</span>
              </div>
            </div>
          </div>
        )}
        
        <div ref={messagesEndRef} />
      </div>

      {/* Input */}
      <div className="border-t bg-white p-4">
        <form onSubmit={handleSubmit} className="flex gap-3 max-w-3xl mx-auto">
          <Input
            value={input}
            onChange={(e) => setInput(e.target.value)}
            placeholder="Ketik pertanyaan Anda..."
            disabled={loading}
            className="flex-1"
            autoFocus
          />
          <Button type="submit" disabled={loading || !input.trim()}>
            {loading ? (
              <Loader2 className="w-4 h-4 animate-spin" />
            ) : (
              <Send className="w-4 h-4" />
            )}
          </Button>
        </form>
        <p className="text-xs text-slate-400 text-center mt-2">
          AI dapat membuat kesalahan. Verifikasi informasi penting.
        </p>
      </div>
    </div>
  )
}
