package api

import (
	"crypto-trading-api/internal/binance"
	"crypto-trading-api/internal/firebase"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// SetupRouter configures all routes and middleware
func SetupRouter(fb *firebase.Client, bn *binance.Client) *gin.Engine {
	router := gin.Default()

	// Middleware
	router.Use(gin.Recovery())
	router.Use(CORSMiddleware())
	router.Use(RateLimitMiddleware())

	// Swagger documentation
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Health check
	router.GET("/health", HealthCheck)

	// Basic API routes
	apiGroup := router.Group("/api")
	apiGroup.Use(AuthMiddleware())
	{
		// Core trading endpoints
		apiGroup.POST("/trade", TradeHandler(fb, bn))
		apiGroup.GET("/trades/:userId", GetTradesHandler(fb))
		apiGroup.GET("/trade/:tradeId", GetTradeHandler(fb))

		// Advanced endpoints
		apiGroup.GET("/status", SystemStatusHandler(fb, bn))           // System status
		apiGroup.GET("/balance", AccountBalanceHandler(bn))            // Account balance
		apiGroup.GET("/positions", OpenPositionsHandler(bn))           // Open positions
		apiGroup.GET("/orders", PendingOrdersHandler(bn))              // Pending orders
		apiGroup.POST("/orders/cancel", CancelOrdersHandler(bn))       // Cancel orders
		apiGroup.POST("/position/close", ClosePositionHandler(bn, fb)) // Close position
		apiGroup.GET("/summary", TradingSummaryHandler(fb, bn))        // Trading summary
		apiGroup.GET("/exchange/info", ExchangeInfoHandler(bn))        // Exchange info (min trade sizes, etc.)
		apiGroup.GET("/account/snapshot", AccountSnapshotHandler(bn))  // Daily account snapshot

		// 🆕 CRITICAL FEATURES - WebSocket, Funding, Risk, Time Sync
		// WebSocket endpoints
		apiGroup.POST("/websocket/start", StartWebSocketHandler(bn))   // Start WebSocket stream
		apiGroup.GET("/websocket/status", WebSocketStatusHandler())    // WebSocket status

		// Funding rate endpoints
		apiGroup.GET("/funding/rate", FundingRateHandler(bn))          // Current funding rate
		apiGroup.GET("/funding/history", FundingRateHistoryHandler(bn)) // Funding rate history

		// Risk management endpoints
		apiGroup.GET("/risk/liquidation", LiquidationRiskHandler(bn))  // Liquidation risk analysis

		// System/Time sync endpoints
		apiGroup.GET("/system/time", TimeSyncHandler(bn))              // Time synchronization check
		apiGroup.GET("/system/server-time", ServerTimeHandler(bn))     // Binance server time
	}

	return router
}
