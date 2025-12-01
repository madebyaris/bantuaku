package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/bantuaku/backend/config"
	"github.com/bantuaku/backend/handlers"
	"github.com/bantuaku/backend/middleware"
	"github.com/bantuaku/backend/services/storage"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Initialize database connection
	db, err := storage.NewPostgres(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Initialize Redis connection
	redis, err := storage.NewRedis(cfg.RedisURL)
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	defer redis.Close()

	// Create handler with dependencies
	h := handlers.New(db, redis, cfg)

	// Setup router
	mux := http.NewServeMux()

	// Health check
	mux.HandleFunc("GET /healthz", h.HealthCheck)

	// Auth routes (public)
	mux.HandleFunc("POST /api/v1/auth/register", h.Register)
	mux.HandleFunc("POST /api/v1/auth/login", h.Login)

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

	// Sentiment & Market
	mux.HandleFunc("GET /api/v1/sentiment/{product_id}", middleware.Auth(cfg.JWTSecret, h.GetSentiment))
	mux.HandleFunc("GET /api/v1/market/trends", middleware.Auth(cfg.JWTSecret, h.GetMarketTrends))

	// AI Assistant
	mux.HandleFunc("POST /api/v1/ai/analyze", middleware.Auth(cfg.JWTSecret, h.AIAnalyze))

	// Dashboard
	mux.HandleFunc("GET /api/v1/dashboard/summary", middleware.Auth(cfg.JWTSecret, h.DashboardSummary))

	// Apply middleware stack
	handler := middleware.Chain(
		mux,
		middleware.RequestID,
		middleware.Logger,
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
		log.Printf("🚀 Bantuaku API server starting on port %s", cfg.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited properly")
}
