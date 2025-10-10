package api

import (
	"crypto-trading-api/internal/binance"
	"crypto-trading-api/internal/firebase"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// SetupRouter configures all routes and middleware
func SetupRouter(fb *firebase.Client, bn *binance.Client) *gin.Engine {
	router := gin.Default()

	// Middleware
	router.Use(gin.Recovery())
	router.Use(CORSMiddleware())
	router.Use(RateLimitMiddleware())

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "healthy",
			"time":   time.Now().Unix(),
		})
	})

	// Basic API routes
	apiGroup := router.Group("/api")
	apiGroup.Use(AuthMiddleware())
	{
		// Core trading endpoints
		apiGroup.POST("/trade", TradeHandler(fb, bn))
		apiGroup.GET("/trades/:userId", GetTradesHandler(fb))
		apiGroup.GET("/trade/:tradeId", GetTradeHandler(fb))

		// Advanced endpoints
		apiGroup.GET("/status", SystemStatusHandler(fb, bn))              // System status
		apiGroup.GET("/balance", AccountBalanceHandler(bn))                // Account balance
		apiGroup.GET("/positions", OpenPositionsHandler(bn))              // Open positions
		apiGroup.GET("/orders", PendingOrdersHandler(bn))                 // Pending orders
		apiGroup.POST("/orders/cancel", CancelOrdersHandler(bn))          // Cancel orders
		apiGroup.POST("/position/close", ClosePositionHandler(bn, fb))    // Close position
		apiGroup.GET("/summary", TradingSummaryHandler(fb, bn))           // Trading summary
	}

	return router
}
