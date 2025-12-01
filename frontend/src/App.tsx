import { Routes, Route, Navigate } from 'react-router-dom'
import { useAuthStore } from '@/state/auth'
import { Layout } from '@/components/layout/Layout'
import { LoginPage } from '@/pages/auth/LoginPage'
import { RegisterPage } from '@/pages/auth/RegisterPage'
import { DashboardPage } from '@/pages/DashboardPage'
import { ProductsPage } from '@/pages/ProductsPage'
import { DataInputPage } from '@/pages/DataInputPage'
import { IntegrationsPage } from '@/pages/IntegrationsPage'
import { AIChatPage } from '@/pages/AIChatPage'
import { Toaster } from '@/components/ui/toaster'

function ProtectedRoute({ children }: { children: React.ReactNode }) {
  const isAuthenticated = useAuthStore((state) => state.isAuthenticated)
  
  if (!isAuthenticated) {
    return <Navigate to="/login" replace />
  }
  
  return <>{children}</>
}

function App() {
  return (
    <>
      <Routes>
        {/* Public routes */}
        <Route path="/login" element={<LoginPage />} />
        <Route path="/register" element={<RegisterPage />} />
        
        {/* Protected routes */}
        <Route
          path="/"
          element={
            <ProtectedRoute>
              <Layout />
            </ProtectedRoute>
          }
        >
          <Route index element={<Navigate to="/dashboard" replace />} />
          <Route path="dashboard" element={<DashboardPage />} />
          <Route path="products" element={<ProductsPage />} />
          <Route path="data-input" element={<DataInputPage />} />
          <Route path="integrations" element={<IntegrationsPage />} />
          <Route path="ai-chat" element={<AIChatPage />} />
        </Route>
        
        {/* Fallback */}
        <Route path="*" element={<Navigate to="/dashboard" replace />} />
      </Routes>
      <Toaster />
    </>
  )
}

export default App
