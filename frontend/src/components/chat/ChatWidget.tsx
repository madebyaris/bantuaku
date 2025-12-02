import { useState, useEffect } from 'react'
import { useLocation, useNavigate } from 'react-router-dom'
import { MessageSquareText, X, Maximize2, Sparkles } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { ChatInterface } from '@/components/chat/ChatInterface'
import { cn } from '@/lib/utils'
import { useChatStore } from '@/state/chat'

export function ChatWidget() {
  const [isOpen, setIsOpen] = useState(false)
  const location = useLocation()
  const navigate = useNavigate()
  const { messages } = useChatStore()
  const [isNewMessage, setIsNewMessage] = useState(false)

  // Hide widget on specific routes (like the full chat page)
  const isHidden = location.pathname === '/ai-chat' || location.pathname === '/login' || location.pathname === '/register'

  useEffect(() => {
    // Show notification dot if new messages arrive while closed
    if (!isOpen && messages.length > 0) {
       // Logic to detect new message could be enhanced, for now just showing dot if messages exist and closed
       // In a real app we'd compare message count
    }
  }, [messages, isOpen])

  if (isHidden) return null

  return (
    <div className="fixed bottom-6 right-6 z-50 flex flex-col items-end space-y-4">
      {/* Chat Popover */}
      <div 
        className={cn(
          "bg-black/90 border border-white/10 backdrop-blur-xl rounded-2xl shadow-2xl transition-all duration-300 overflow-hidden origin-bottom-right",
          isOpen 
            ? "w-[calc(100vw-3rem)] sm:w-[350px] h-[500px] max-h-[calc(100vh-8rem)] opacity-100 scale-100 translate-y-0" 
            : "w-0 h-0 opacity-0 scale-95 translate-y-10 pointer-events-none"
        )}
      >
        <div className="flex items-center justify-between p-3 border-b border-white/10 bg-white/5">
            <div className="flex items-center gap-2">
                <Sparkles className="w-4 h-4 text-emerald-400" />
                <span className="font-semibold text-sm text-slate-200">AI Assistant</span>
            </div>
            <div className="flex items-center gap-1">
                <Button 
                    variant="ghost" 
                    size="icon" 
                    className="h-8 w-8 text-slate-400 hover:text-emerald-400 hover:bg-white/5"
                    onClick={() => {
                        setIsOpen(false)
                        navigate('/ai-chat')
                    }}
                    title="Expand to full page"
                >
                    <Maximize2 className="w-4 h-4" />
                </Button>
                <Button 
                    variant="ghost" 
                    size="icon" 
                    className="h-8 w-8 text-slate-400 hover:text-white hover:bg-white/5"
                    onClick={() => setIsOpen(false)}
                >
                    <X className="w-4 h-4" />
                </Button>
            </div>
        </div>
        <div className="h-[calc(100%-50px)]">
            <ChatInterface isWidget className="h-full" />
        </div>
      </div>

      {/* Floating Action Button */}
      <Button
        onClick={() => setIsOpen(!isOpen)}
        className={cn(
          "rounded-full w-14 h-14 shadow-[0_0_20px_rgba(16,185,129,0.3)] transition-all duration-300",
          isOpen 
            ? "bg-slate-800 hover:bg-slate-700 text-slate-200 rotate-90" 
            : "bg-gradient-to-r from-emerald-600 to-emerald-400 hover:from-emerald-500 hover:to-emerald-300 text-black rotate-0"
        )}
      >
        {isOpen ? (
            <X className="w-6 h-6" />
        ) : (
            <MessageSquareText className="w-6 h-6 fill-current" />
        )}
      </Button>
    </div>
  )
}
