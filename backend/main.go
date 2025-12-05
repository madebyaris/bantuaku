package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/bantuaku/backend/config"
	"github.com/bantuaku/backend/handlers"
	adminhandlers "github.com/bantuaku/backend/handlers/admin"
	"github.com/bantuaku/backend/logger"
	"github.com/bantuaku/backend/middleware"
	"github.com/bantuaku/backend/services/audit"
	"github.com/bantuaku/backend/services/billing"
	"github.com/bantuaku/backend/services/storage"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Initialize structured logger
	logger.InitGlobalLogger(logger.Config{
		Level:  logger.LogLevel(cfg.LogLevel),
		Format: "json",
		Output: os.Stdout,
	})

	log := logger.Default()
	log.Info("Starting Bantuaku API server", "version", "0.1.0")

	// Initialize database connection
	log.Info("Connecting to database", "url", maskDatabaseURL(cfg.DatabaseURL))
	db, err := storage.NewPostgres(cfg.DatabaseURL)
	if err != nil {
		log.Error("Failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer db.Close()
	log.Info("Database connection established")

	// Initialize Redis connection
	log.Info("Connecting to Redis", "url", maskRedisURL(cfg.RedisURL))
	redis, err := storage.NewRedis(cfg.RedisURL)
	if err != nil {
		log.Warn("Failed to connect to Redis", "error", err)
		// Continue without Redis for now
	} else {
	defer redis.Close()
		log.Info("Redis connection established")
	}

	// Create handler with dependencies
	h := handlers.New(db, redis, cfg)
	log.Info("HTTP handlers initialized")

	// Initialize audit logger
	auditLogger := audit.NewLogger(db)
	log.Info("Audit logger initialized")

	// Initialize Stripe billing service (if configured)
	var billingHandler *handlers.BillingHandler
	if cfg.StripeSecretKey != "" {
		stripeService := billing.NewStripeService(cfg.StripeSecretKey, cfg.StripeWebhookSecret, db)
		billingHandler = handlers.NewBillingHandler(stripeService, db)
		log.Info("Stripe billing service initialized")
	}

	// Setup router
	mux := http.NewServeMux()

	// Health check
	mux.HandleFunc("GET /healthz", h.HealthCheck)

	// Auth routes (public)
	mux.HandleFunc("POST /api/v1/auth/register", h.Register)
	mux.HandleFunc("POST /api/v1/auth/login", h.Login)
	mux.HandleFunc("POST /api/v1/auth/verify-email", h.VerifyEmail)
	mux.HandleFunc("POST /api/v1/auth/resend-verification", h.ResendVerification)
	mux.HandleFunc("POST /api/v1/auth/forgot-password", h.RequestPasswordReset)
	mux.HandleFunc("POST /api/v1/auth/reset-password", h.ResetPassword)

	// Protected routes
	mux.HandleFunc("GET /api/v1/products", middleware.Auth(cfg.JWTSecret, h.ListProducts))
	mux.HandleFunc("POST /api/v1/products", middleware.Auth(cfg.JWTSecret, h.CreateProduct))
	mux.HandleFunc("GET /api/v1/products/{id}", middleware.Auth(cfg.JWTSecret, h.GetProduct))
	mux.HandleFunc("PUT /api/v1/products/{id}", middleware.Auth(cfg.JWTSecret, h.UpdateProduct))
	mux.HandleFunc("DELETE /api/v1/products/{id}", middleware.Auth(cfg.JWTSecret, h.DeleteProduct))

	// Sales data input
	mux.HandleFunc("POST /api/v1/sales/manual", middleware.Auth(cfg.JWTSecret, h.RecordSale))
	mux.HandleFunc("POST /api/v1/sales/import-csv", middleware.Auth(cfg.JWTSecret, h.ImportCSV))
	mux.HandleFunc("GET /api/v1/sales", middleware.Auth(cfg.JWTSecret, h.ListSales))

	// WooCommerce integration
	mux.HandleFunc("POST /api/v1/integrations/woocommerce/connect", middleware.Auth(cfg.JWTSecret, h.WooCommerceConnect))
	mux.HandleFunc("GET /api/v1/integrations/woocommerce/sync-status", middleware.Auth(cfg.JWTSecret, h.WooCommerceSyncStatus))
	mux.HandleFunc("POST /api/v1/integrations/woocommerce/sync-now", middleware.Auth(cfg.JWTSecret, h.WooCommerceSyncNow))

	// Forecasting
	mux.HandleFunc("GET /api/v1/forecasts/{product_id}", middleware.Auth(cfg.JWTSecret, h.GetForecast))
	mux.HandleFunc("GET /api/v1/recommendations", middleware.Auth(cfg.JWTSecret, h.GetRecommendations))
	
	// Advanced Forecasting (12-month)
	mux.HandleFunc("GET /api/v1/forecasts/monthly", middleware.Auth(cfg.JWTSecret, h.GetMonthlyForecasts))
	mux.HandleFunc("POST /api/v1/forecasts/monthly/generate", middleware.Auth(cfg.JWTSecret, h.GenerateMonthlyForecast))
	mux.HandleFunc("GET /api/v1/strategies/monthly", middleware.Auth(cfg.JWTSecret, h.GetMonthlyStrategies))
	mux.HandleFunc("POST /api/v1/strategies/monthly/generate", middleware.Auth(cfg.JWTSecret, h.GenerateStrategies))

	// Sentiment & Market
	mux.HandleFunc("GET /api/v1/sentiment/{product_id}", middleware.Auth(cfg.JWTSecret, h.GetSentiment))
	mux.HandleFunc("GET /api/v1/market/trends", middleware.Auth(cfg.JWTSecret, h.GetMarketTrends))

	// AI Assistant (legacy)
	mux.HandleFunc("POST /api/v1/ai/analyze", middleware.Auth(cfg.JWTSecret, h.AIAnalyze))

	// Chat & Conversations (NEW)
	mux.HandleFunc("POST /api/v1/chat/start", middleware.Auth(cfg.JWTSecret, h.StartConversation))
	mux.HandleFunc("POST /api/v1/chat/message", middleware.Auth(cfg.JWTSecret, h.SendMessage))
	mux.HandleFunc("GET /api/v1/chat/conversations", middleware.Auth(cfg.JWTSecret, h.GetConversations))
	mux.HandleFunc("GET /api/v1/chat/messages", middleware.Auth(cfg.JWTSecret, h.GetMessages))
	mux.HandleFunc("POST /api/v1/chat/feedback", middleware.Auth(cfg.JWTSecret, h.SubmitFeedback))

	// File Uploads (NEW)
	mux.HandleFunc("POST /api/v1/files/upload", middleware.Auth(cfg.JWTSecret, h.UploadFile))
	mux.HandleFunc("GET /api/v1/files/{id}", middleware.Auth(cfg.JWTSecret, h.GetFile))

	// Insights (NEW - Four Outcome Types)
	mux.HandleFunc("POST /api/v1/insights/forecast", middleware.Auth(cfg.JWTSecret, h.GenerateForecastInsight))
	mux.HandleFunc("POST /api/v1/insights/market", middleware.Auth(cfg.JWTSecret, h.GenerateMarketInsight))
	mux.HandleFunc("POST /api/v1/insights/marketing", middleware.Auth(cfg.JWTSecret, h.GenerateMarketingInsight))
	mux.HandleFunc("POST /api/v1/insights/regulation", middleware.Auth(cfg.JWTSecret, h.GenerateRegulationInsight))
	mux.HandleFunc("GET /api/v1/insights", middleware.Auth(cfg.JWTSecret, h.GetInsights))

	// Dashboard
	mux.HandleFunc("GET /api/v1/dashboard/summary", middleware.Auth(cfg.JWTSecret, h.DashboardSummary))

	// Regulations scraper (admin endpoints) - with rate limiting
	scrapingRateLimit := middleware.RateLimiter(redis, middleware.DefaultRateLimitConfigs.Scraping)
	mux.Handle("POST /api/v1/regulations/scrape", scrapingRateLimit(middleware.Auth(cfg.JWTSecret, h.TriggerScraping)))
	mux.HandleFunc("GET /api/v1/regulations/status", middleware.Auth(cfg.JWTSecret, h.GetScrapingStatus))
	mux.HandleFunc("GET /api/v1/regulations", middleware.Auth(cfg.JWTSecret, h.ListRegulations))

	// Embeddings & Vectorization
	mux.HandleFunc("POST /api/v1/embeddings/index", middleware.Auth(cfg.JWTSecret, h.IndexChunks))
	mux.HandleFunc("GET /api/v1/regulations/search", middleware.Auth(cfg.JWTSecret, h.SearchRegulations))

	// Google Trends - with rate limiting
	trendsRateLimit := middleware.RateLimiter(redis, middleware.DefaultRateLimitConfigs.Trends)
	mux.Handle("POST /api/v1/trends/keywords", trendsRateLimit(middleware.Auth(cfg.JWTSecret, h.CreateKeyword)))
	mux.HandleFunc("GET /api/v1/trends/keywords", middleware.Auth(cfg.JWTSecret, h.ListKeywords))
	mux.HandleFunc("DELETE /api/v1/trends/keywords", middleware.Auth(cfg.JWTSecret, h.DeleteKeyword))
	mux.HandleFunc("GET /api/v1/trends/series", middleware.Auth(cfg.JWTSecret, h.GetTimeSeries))
	mux.Handle("POST /api/v1/trends/ingest", trendsRateLimit(middleware.Auth(cfg.JWTSecret, h.TriggerIngestion)))

	// Admin routes (RBAC protected) - with rate limiting
	adminHandler := adminhandlers.NewAdminHandler(db, cfg.JWTSecret, auditLogger)
	adminRateLimit := middleware.RateLimiter(redis, middleware.DefaultRateLimitConfigs.Admin)
	adminAuth := func(handler http.HandlerFunc) http.Handler {
		return adminRateLimit(middleware.Auth(cfg.JWTSecret, middleware.RequireAdmin(handler)))
	}
	
	mux.Handle("GET /api/v1/admin/users", adminAuth(adminHandler.ListUsers))
	mux.Handle("GET /api/v1/admin/users/{id}", adminAuth(adminHandler.GetUser))
	mux.Handle("POST /api/v1/admin/users", adminAuth(adminHandler.CreateUser))
	mux.Handle("PUT /api/v1/admin/users/{id}", adminAuth(adminHandler.UpdateUser))
	mux.Handle("PUT /api/v1/admin/users/{id}/role", adminAuth(adminHandler.UpdateUserRole))
	mux.Handle("PUT /api/v1/admin/users/{id}/status", adminAuth(adminHandler.UpdateUserStatus))
	mux.Handle("PUT /api/v1/admin/users/{id}/upgrade-subscription", adminAuth(adminHandler.UpgradeUserSubscription))
	mux.Handle("DELETE /api/v1/admin/users/{id}", adminAuth(adminHandler.DeleteUser))
	mux.Handle("GET /api/v1/admin/stats", adminAuth(adminHandler.GetStats))
	
	mux.Handle("GET /api/v1/admin/subscriptions", adminAuth(adminHandler.ListSubscriptions))
	mux.Handle("GET /api/v1/admin/subscriptions/{id}", adminAuth(adminHandler.GetSubscription))
	mux.Handle("POST /api/v1/admin/subscriptions", adminAuth(adminHandler.CreateSubscription))
	mux.Handle("PUT /api/v1/admin/subscriptions/{id}/status", adminAuth(adminHandler.UpdateSubscriptionStatus))
	
	mux.Handle("GET /api/v1/admin/audit-logs", adminAuth(adminHandler.ListAuditLogs))

	// Billing routes (if Stripe is configured)
	if billingHandler != nil {
		mux.HandleFunc("POST /api/v1/billing/checkout", middleware.Auth(cfg.JWTSecret, billingHandler.CreateCheckoutSession))
		mux.HandleFunc("GET /api/v1/billing/subscription", middleware.Auth(cfg.JWTSecret, billingHandler.GetSubscription))
		mux.HandleFunc("GET /api/v1/billing/plans", billingHandler.ListPlans) // Public endpoint
		mux.HandleFunc("POST /api/v1/billing/webhook", billingHandler.HandleWebhook) // Webhook doesn't need auth
	}

	// Apply middleware stack
	handler := middleware.Chain(
		mux,
		middleware.RequestID,
		middleware.StructuredLogger,
		middleware.ErrorHandler,
		middleware.CORS(cfg.CORSOrigin),
		middleware.Recover,
	)

	// Create server
	server := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      handler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in goroutine
	go func() {
		log.Info("Starting HTTP server", "port", cfg.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error("Server error", "error", err)
			os.Exit(1)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("Shutting down server gracefully...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Error("Server forced to shutdown", "error", err)
		os.Exit(1)
	}

	log.Info("Server exited properly")
}

// maskDatabaseURL masks sensitive information in database URL for logging
func maskDatabaseURL(url string) string {
	if url == "" {
		return "empty"
	}

	// For PostgreSQL URLs, mask the password
	if len(url) > 10 && url[:10] == "postgres://" {
		// Find the user:password part (between // and @)
		startIdx := len("postgres://")
		atIdx := -1
		for i := startIdx; i < len(url); i++ {
			if url[i] == '@' {
				atIdx = i
				break
			}
		}

		if atIdx != -1 {
			// Find the colon in the user:password part
			colonIdx := -1
			for i := startIdx; i < atIdx; i++ {
				if url[i] == ':' {
					colonIdx = i
					break
				}
			}

			if colonIdx != -1 {
				// Return masked URL
				return url[:colonIdx+1] + "****" + url[atIdx:]
			}
		}
	}

	return "****" // Fallback
}

// maskRedisURL masks sensitive information in Redis URL for logging
func maskRedisURL(url string) string {
	if url == "" {
		return "empty"
	}

	// For Redis URLs, mask the password
	if len(url) > 8 && url[:8] == "redis://" {
		return "redis://****"
	}

	return "****" // Fallback
}
