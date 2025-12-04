import { Routes, Route, Navigate } from "react-router-dom";
import { useAuthStore } from "@/state/auth";
import { ErrorBoundary } from "@/components/ui/ErrorBoundary";
import { Layout } from "@/components/layout/Layout";
import { LoginPage } from "@/pages/auth/LoginPage";
import { RegisterPage } from "@/pages/auth/RegisterPage";
import { VerifyEmailPage } from "@/pages/auth/VerifyEmailPage";
import { ForgotPasswordPage } from "@/pages/auth/ForgotPasswordPage";
import { ResetPasswordPage } from "@/pages/auth/ResetPasswordPage";
import { DashboardPage } from "@/pages/DashboardPage";
import { ForecastPage } from "@/pages/ForecastPage";
import { MarketPredictionPage } from "@/pages/MarketPredictionPage";
import { MarketingPage } from "@/pages/MarketingPage";
import { RegulationPage } from "@/pages/RegulationPage";
import { AIChatPage } from "@/pages/AIChatPage";
import { Toaster } from "@/components/ui/toaster";

function ProtectedRoute({ children }: { children: React.ReactNode }) {
  const isAuthenticated = useAuthStore((state) => state.isAuthenticated);

  if (!isAuthenticated) {
    return <Navigate to="/login" replace />;
  }

  return <>{children}</>;
}

function App() {
  return (
    <ErrorBoundary>
      <Routes>
        {/* Public routes */}
        <Route path="/login" element={<LoginPage />} />
        <Route path="/register" element={<RegisterPage />} />
        <Route path="/verify-email" element={<VerifyEmailPage />} />
        <Route path="/forgot-password" element={<ForgotPasswordPage />} />
        <Route path="/reset-password" element={<ResetPasswordPage />} />

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
          <Route path="forecast" element={<ForecastPage />} />
          <Route path="market-prediction" element={<MarketPredictionPage />} />
          <Route path="marketing" element={<MarketingPage />} />
          <Route path="regulation" element={<RegulationPage />} />
          <Route path="ai-chat" element={<AIChatPage />} />
        </Route>

        {/* Fallback */}
        <Route path="*" element={<Navigate to="/dashboard" replace />} />
      </Routes>
      <Toaster />
    </ErrorBoundary>
  );
}

export default App;
