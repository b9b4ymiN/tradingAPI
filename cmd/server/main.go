package main

import (
	"context"
	"crypto-trading-api/config"
	_ "crypto-trading-api/docs" // Import generated Swagger docs
	"crypto-trading-api/internal/api"
	"crypto-trading-api/internal/binance"
	"crypto-trading-api/internal/firebase"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
)

// @title           Crypto Trading API
// @version         1.0
// @description     Professional cryptocurrency trading API with automated Stop Loss and Take Profit
// @description     Supports Binance Futures trading with real-time monitoring and Firebase logging

// @contact.name   API Support
// @contact.email  support@cryptotradingapi.com

// @license.name  MIT
// @license.url   https://opensource.org/licenses/MIT

// @host      localhost:8080
// @BasePath  /

// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name X-API-Key
// @description Enter your API key to access protected endpoints

// @tag.name Health
// @tag.description Health check endpoints

// @tag.name Trading
// @tag.description Trading operations (place, view, close trades)

// @tag.name Account
// @tag.description Account and balance information

// @tag.name Positions
// @tag.description Position and order management

// @tag.name Orders
// @tag.description Order management and cancellation

// @tag.name System
// @tag.description System status and monitoring

// @tag.name Analytics
// @tag.description Trading analytics and statistics

func main() {
	// Load configuration
	cfg := config.Load()

	// Set Gin mode
	gin.SetMode(cfg.GinMode)

	// Initialize Firebase
	firebaseClient, err := firebase.InitClient()
	if err != nil {
		log.Fatalf("Failed to initialize Firebase: %v", err)
	}
	defer firebaseClient.Close()

	// Initialize Binance client
	binanceClient := binance.InitClient()

	// Setup router
	router := api.SetupRouter(firebaseClient, binanceClient)

	// Server configuration
	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// Start server in goroutine
	go func() {
		log.Printf("🚀 Server starting on port %s", cfg.Port)
		log.Printf("📄 Swagger docs: http://localhost:%s/swagger/index.html", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("🛑 Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	log.Println("✅ Server exited")
}
