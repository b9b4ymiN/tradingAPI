package api

import (
	"crypto-trading-api/internal/binance"
	"crypto-trading-api/internal/models"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// Global WebSocket manager
var wsManager *binance.WebSocketManager

// InitWebSocketManager initializes the WebSocket manager
func InitWebSocketManager(bn *binance.Client) {
	wsManager = binance.NewWebSocketManager(bn)
}

// StartWebSocketHandler - Start WebSocket user data stream
// @Summary      Start WebSocket user data stream
// @Description  Start real-time WebSocket stream for order updates and account changes
// @Tags         WebSocket
// @Produce      json
// @Security     ApiKeyAuth
// @Success      200  {object}  models.TradeResponse  "WebSocket started successfully"
// @Failure      401  {object}  models.TradeResponse  "Unauthorized"
// @Failure      500  {object}  models.TradeResponse  "Failed to start WebSocket"
// @Router       /api/websocket/start [post]
func StartWebSocketHandler(bn *binance.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		if wsManager == nil {
			InitWebSocketManager(bn)
		}

		// Start user data stream
		err := wsManager.StartUserDataStream(
			// Order update callback
			func(event *binance.OrderUpdateEvent) {
				// Log order updates
				// In production, you might want to update Firebase here
			},
			// Account update callback
			func(event *binance.AccountUpdateEvent) {
				// Log account updates
			},
		)

		if err != nil {
			c.JSON(http.StatusInternalServerError, models.TradeResponse{
				Success:   false,
				Message:   "Failed to start WebSocket stream",
				Error:     err.Error(),
				Timestamp: time.Now().Unix(),
			})
			return
		}

		c.JSON(http.StatusOK, models.TradeResponse{
			Success:   true,
			Message:   "WebSocket user data stream started successfully",
			Timestamp: time.Now().Unix(),
		})
	}
}

// WebSocketStatusHandler - Get WebSocket connection status
// @Summary      Get WebSocket status
// @Description  Check the status of all active WebSocket connections
// @Tags         WebSocket
// @Produce      json
// @Security     ApiKeyAuth
// @Success      200  {object}  models.TradeResponse  "WebSocket status retrieved"
// @Failure      401  {object}  models.TradeResponse  "Unauthorized"
// @Router       /api/websocket/status [get]
func WebSocketStatusHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		if wsManager == nil {
			c.JSON(http.StatusOK, models.TradeResponse{
				Success:   true,
				Message:   "WebSocket not initialized",
				Data:      gin.H{"status": "not_initialized"},
				Timestamp: time.Now().Unix(),
			})
			return
		}

		status := wsManager.GetStreamStatus()

		c.JSON(http.StatusOK, models.TradeResponse{
			Success:   true,
			Message:   "WebSocket status retrieved",
			Data:      status,
			Timestamp: time.Now().Unix(),
		})
	}
}

// FundingRateHandler - Get current funding rate
// @Summary      Get funding rate
// @Description  Get current funding rate for a symbol
// @Tags         Funding
// @Produce      json
// @Security     ApiKeyAuth
// @Param        symbol  query     string  true  "Trading symbol" example("BTCUSDT")
// @Success      200     {object}  models.TradeResponse{data=binance.FundingRateInfo}  "Funding rate retrieved"
// @Failure      400     {object}  models.TradeResponse  "Missing symbol parameter"
// @Failure      401     {object}  models.TradeResponse  "Unauthorized"
// @Failure      500     {object}  models.TradeResponse  "Failed to get funding rate"
// @Router       /api/funding/rate [get]
func FundingRateHandler(bn *binance.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		symbol := c.Query("symbol")
		if symbol == "" {
			c.JSON(http.StatusBadRequest, models.TradeResponse{
				Success:   false,
				Message:   "Missing symbol parameter",
				Error:     "symbol is required",
				Timestamp: time.Now().Unix(),
			})
			return
		}

		fundingRate, err := bn.GetFundingRate(symbol)
		if err != nil {
			c.JSON(http.StatusInternalServerError, models.TradeResponse{
				Success:   false,
				Message:   "Failed to get funding rate",
				Error:     err.Error(),
				Timestamp: time.Now().Unix(),
			})
			return
		}

		c.JSON(http.StatusOK, models.TradeResponse{
			Success:   true,
			Message:   "Funding rate retrieved successfully",
			Data:      fundingRate,
			Timestamp: time.Now().Unix(),
		})
	}
}

// FundingRateHistoryHandler - Get funding rate history
// @Summary      Get funding rate history
// @Description  Get historical funding rates for a symbol
// @Tags         Funding
// @Produce      json
// @Security     ApiKeyAuth
// @Param        symbol     query  string  true   "Trading symbol" example("BTCUSDT")
// @Param        limit      query  int     false  "Number of records (default: 100, max: 1000)" example(100)
// @Param        startTime  query  int64   false  "Start timestamp (seconds)" example(1640000000)
// @Param        endTime    query  int64   false  "End timestamp (seconds)" example(1650000000)
// @Success      200        {object}  models.TradeResponse{data=[]binance.FundingRateHistory}  "Funding rate history retrieved"
// @Failure      400        {object}  models.TradeResponse  "Missing symbol parameter"
// @Failure      401        {object}  models.TradeResponse  "Unauthorized"
// @Failure      500        {object}  models.TradeResponse  "Failed to get funding rate history"
// @Router       /api/funding/history [get]
func FundingRateHistoryHandler(bn *binance.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		symbol := c.Query("symbol")
		if symbol == "" {
			c.JSON(http.StatusBadRequest, models.TradeResponse{
				Success:   false,
				Message:   "Missing symbol parameter",
				Error:     "symbol is required",
				Timestamp: time.Now().Unix(),
			})
			return
		}

		limit, _ := strconv.Atoi(c.DefaultQuery("limit", "100"))
		startTime, _ := strconv.ParseInt(c.Query("startTime"), 10, 64)
		endTime, _ := strconv.ParseInt(c.Query("endTime"), 10, 64)

		history, err := bn.GetFundingRateHistory(symbol, limit, startTime, endTime)
		if err != nil {
			c.JSON(http.StatusInternalServerError, models.TradeResponse{
				Success:   false,
				Message:   "Failed to get funding rate history",
				Error:     err.Error(),
				Timestamp: time.Now().Unix(),
			})
			return
		}

		c.JSON(http.StatusOK, models.TradeResponse{
			Success:   true,
			Message:   "Funding rate history retrieved successfully",
			Data:      history,
			Timestamp: time.Now().Unix(),
		})
	}
}

// LiquidationRiskHandler - Get liquidation risk for a position
// @Summary      Get liquidation risk
// @Description  Calculate liquidation risk and distance to liquidation for a position
// @Tags         Risk Management
// @Produce      json
// @Security     ApiKeyAuth
// @Param        symbol  query     string  true  "Trading symbol" example("BTCUSDT")
// @Success      200     {object}  models.TradeResponse{data=binance.LiquidationRisk}  "Liquidation risk calculated"
// @Failure      400     {object}  models.TradeResponse  "Missing symbol parameter"
// @Failure      401     {object}  models.TradeResponse  "Unauthorized"
// @Failure      404     {object}  models.TradeResponse  "No position found"
// @Failure      500     {object}  models.TradeResponse  "Failed to calculate risk"
// @Router       /api/risk/liquidation [get]
func LiquidationRiskHandler(bn *binance.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		symbol := c.Query("symbol")
		if symbol == "" {
			c.JSON(http.StatusBadRequest, models.TradeResponse{
				Success:   false,
				Message:   "Missing symbol parameter",
				Error:     "symbol is required",
				Timestamp: time.Now().Unix(),
			})
			return
		}

		risk, err := bn.GetLiquidationRisk(symbol)
		if err != nil {
			statusCode := http.StatusInternalServerError
			if err.Error() == "no position found for "+symbol ||
				err.Error() == "no open position for "+symbol {
				statusCode = http.StatusNotFound
			}

			c.JSON(statusCode, models.TradeResponse{
				Success:   false,
				Message:   "Failed to calculate liquidation risk",
				Error:     err.Error(),
				Timestamp: time.Now().Unix(),
			})
			return
		}

		c.JSON(http.StatusOK, models.TradeResponse{
			Success:   true,
			Message:   "Liquidation risk calculated successfully",
			Data:      risk,
			Timestamp: time.Now().Unix(),
		})
	}
}

// TimeSyncHandler - Get Binance server time and sync status
// @Summary      Check time synchronization
// @Description  Get Binance server time and check if local time is synchronized
// @Tags         System
// @Produce      json
// @Security     ApiKeyAuth
// @Success      200  {object}  models.TradeResponse  "Time sync status retrieved"
// @Failure      401  {object}  models.TradeResponse  "Unauthorized"
// @Failure      500  {object}  models.TradeResponse  "Failed to sync time"
// @Router       /api/system/time [get]
func TimeSyncHandler(bn *binance.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		isInSync, offset, err := bn.CheckTimeSyncStatus()
		if err != nil {
			c.JSON(http.StatusInternalServerError, models.TradeResponse{
				Success:   false,
				Message:   "Failed to sync time",
				Error:     err.Error(),
				Timestamp: time.Now().Unix(),
			})
			return
		}

		serverTime, _ := bn.GetBinanceServerTime()

		data := gin.H{
			"isInSync":       isInSync,
			"offsetMs":       offset,
			"serverTime":     serverTime,
			"localTime":      time.Now().UnixMilli(),
			"recommendation": "",
		}

		if !isInSync {
			data["recommendation"] = "Clock drift detected. Sync your system clock using NTP: ntpdate pool.ntp.org"
		}

		c.JSON(http.StatusOK, models.TradeResponse{
			Success:   true,
			Message:   "Time sync status retrieved",
			Data:      data,
			Timestamp: time.Now().Unix(),
		})
	}
}

// ServerTimeHandler - Get Binance server time
// @Summary      Get server time
// @Description  Get current Binance server timestamp
// @Tags         System
// @Produce      json
// @Success      200  {object}  models.TradeResponse  "Server time retrieved"
// @Failure      500  {object}  models.TradeResponse  "Failed to get server time"
// @Router       /api/system/server-time [get]
func ServerTimeHandler(bn *binance.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		serverTime, err := bn.GetBinanceServerTime()
		if err != nil {
			c.JSON(http.StatusInternalServerError, models.TradeResponse{
				Success:   false,
				Message:   "Failed to get server time",
				Error:     err.Error(),
				Timestamp: time.Now().Unix(),
			})
			return
		}

		c.JSON(http.StatusOK, models.TradeResponse{
			Success:   true,
			Message:   "Server time retrieved",
			Data: gin.H{
				"serverTime": serverTime,
				"localTime":  time.Now().UnixMilli(),
			},
			Timestamp: time.Now().Unix(),
		})
	}
}
