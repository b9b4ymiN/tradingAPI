package api

import (
	"crypto-trading-api/internal/binance"
	"crypto-trading-api/internal/firebase"
	"crypto-trading-api/internal/models"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

var serverStartTime = time.Now().Unix()

// SystemStatusHandler - Get system status
func SystemStatusHandler(fb *firebase.Client, bn *binance.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		// Get system stats
		activeTrades, err := fb.GetActiveTrades(ctx)
		if err != nil {
			c.JSON(http.StatusInternalServerError, models.TradeResponse{
				Success:   false,
				Message:   "Failed to get active trades",
				Error:     err.Error(),
				Timestamp: time.Now().Unix(),
			})
			return
		}

		// Get Binance server time (to check connection)
		serverTime, err := bn.GetServerTime()
		if err != nil {
			c.JSON(http.StatusInternalServerError, models.TradeResponse{
				Success:   false,
				Message:   "Failed to connect to Binance",
				Error:     err.Error(),
				Timestamp: time.Now().Unix(),
			})
			return
		}

		// Get account status
		account, err := bn.GetAccountInfo()
		if err != nil {
			c.JSON(http.StatusInternalServerError, models.TradeResponse{
				Success:   false,
				Message:   "Failed to get account info",
				Error:     err.Error(),
				Timestamp: time.Now().Unix(),
			})
			return
		}

		status := gin.H{
			"server": gin.H{
				"status":    "online",
				"uptime":    time.Now().Unix() - serverStartTime,
				"timestamp": time.Now().Unix(),
				"version":   "1.1.0",
			},
			"binance": gin.H{
				"status":      "connected",
				"serverTime":  serverTime,
				"canTrade":    account.CanTrade,
				"canDeposit":  account.CanDeposit,
				"canWithdraw": account.CanWithdraw,
			},
			"firebase": gin.H{
				"status":       "connected",
				"activeTrades": len(activeTrades),
			},
		}

		c.JSON(http.StatusOK, models.TradeResponse{
			Success:   true,
			Message:   "System status retrieved successfully",
			Data:      status,
			Timestamp: time.Now().Unix(),
		})
	}
}

// AccountBalanceHandler - Get account balance
func AccountBalanceHandler(bn *binance.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		account, err := bn.GetAccountInfo()
		if err != nil {
			c.JSON(http.StatusInternalServerError, models.TradeResponse{
				Success:   false,
				Message:   "Failed to get account balance",
				Error:     err.Error(),
				Timestamp: time.Now().Unix(),
			})
			return
		}

		// Calculate total balance
		balance := bn.CalculateBalance(account)

		c.JSON(http.StatusOK, models.TradeResponse{
			Success:   true,
			Message:   "Account balance retrieved successfully",
			Data:      balance,
			Timestamp: time.Now().Unix(),
		})
	}
}

// OpenPositionsHandler - Get open positions with PnL
func OpenPositionsHandler(bn *binance.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		positions, err := bn.GetOpenPositions()
		if err != nil {
			c.JSON(http.StatusInternalServerError, models.TradeResponse{
				Success:   false,
				Message:   "Failed to get open positions",
				Error:     err.Error(),
				Timestamp: time.Now().Unix(),
			})
			return
		}

		// Calculate total PNL
		totalPnL := 0.0
		totalPositions := 0
		positionDetails := []gin.H{}

		for _, pos := range positions {
			if pos.PositionAmt != 0 {
				totalPositions++
				totalPnL += pos.UnrealizedProfit

				positionDetails = append(positionDetails, gin.H{
					"symbol":           pos.Symbol,
					"side":             pos.PositionSide,
					"positionAmt":      pos.PositionAmt,
					"entryPrice":       pos.EntryPrice,
					"markPrice":        pos.MarkPrice,
					"unrealizedProfit": pos.UnrealizedProfit,
					"leverage":         pos.Leverage,
					"liquidationPrice": pos.LiquidationPrice,
					"marginType":       pos.MarginType,
				})
			}
		}

		data := gin.H{
			"totalPositions": totalPositions,
			"totalPnL":       totalPnL,
			"positions":      positionDetails,
		}

		c.JSON(http.StatusOK, models.TradeResponse{
			Success:   true,
			Message:   "Open positions retrieved successfully",
			Data:      data,
			Timestamp: time.Now().Unix(),
		})
	}
}

// PendingOrdersHandler - Get pending orders
func PendingOrdersHandler(bn *binance.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		symbol := c.Query("symbol") // Optional: filter by symbol

		orders, err := bn.GetOpenOrders(symbol)
		if err != nil {
			c.JSON(http.StatusInternalServerError, models.TradeResponse{
				Success:   false,
				Message:   "Failed to get pending orders",
				Error:     err.Error(),
				Timestamp: time.Now().Unix(),
			})
			return
		}

		orderDetails := []gin.H{}
		for _, order := range orders {
			orderDetails = append(orderDetails, gin.H{
				"orderId":       order.OrderID,
				"symbol":        order.Symbol,
				"side":          order.Side,
				"type":          order.Type,
				"price":         order.Price,
				"stopPrice":     order.StopPrice,
				"quantity":      order.OrigQuantity,
				"status":        order.Status,
				"timeInForce":   order.TimeInForce,
				"createdTime":   order.Time,
				"reduceOnly":    order.ReduceOnly,
				"closePosition": order.ClosePosition,
			})
		}

		data := gin.H{
			"totalOrders": len(orderDetails),
			"orders":      orderDetails,
		}

		c.JSON(http.StatusOK, models.TradeResponse{
			Success:   true,
			Message:   "Pending orders retrieved successfully",
			Data:      data,
			Timestamp: time.Now().Unix(),
		})
	}
}

// CancelOrdersHandler - Cancel pending orders
func CancelOrdersHandler(bn *binance.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Symbol  string `json:"symbol"`  // Optional: cancel by symbol
			OrderID int64  `json:"orderId"` // Optional: cancel specific order
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			// If no body, cancel all orders
			req.Symbol = ""
			req.OrderID = 0
		}

		var cancelResults []gin.H
		var errors []string
		cancelledCount := 0

		if req.OrderID != 0 && req.Symbol != "" {
			// Cancel specific order
			err := bn.CancelOrder(req.Symbol, req.OrderID)
			if err != nil {
				errors = append(errors, err.Error())
			} else {
				cancelledCount++
				cancelResults = append(cancelResults, gin.H{
					"symbol":  req.Symbol,
					"orderId": req.OrderID,
					"status":  "cancelled",
				})
			}
		} else if req.Symbol != "" {
			// Cancel all orders for symbol
			result, err := bn.CancelAllOrders(req.Symbol)
			if err != nil {
				errors = append(errors, err.Error())
			} else {
				cancelledCount = result
				cancelResults = append(cancelResults, gin.H{
					"symbol":          req.Symbol,
					"cancelledOrders": result,
					"status":          "success",
				})
			}
		} else {
			// Cancel all orders (all symbols)
			symbols, err := bn.GetActiveSymbols()
			if err != nil {
				c.JSON(http.StatusInternalServerError, models.TradeResponse{
					Success:   false,
					Message:   "Failed to get active symbols",
					Error:     err.Error(),
					Timestamp: time.Now().Unix(),
				})
				return
			}

			for _, symbol := range symbols {
				result, err := bn.CancelAllOrders(symbol)
				if err != nil {
					errors = append(errors, err.Error())
				} else {
					cancelledCount += result
					if result > 0 {
						cancelResults = append(cancelResults, gin.H{
							"symbol":          symbol,
							"cancelledOrders": result,
						})
					}
				}
			}
		}

		data := gin.H{
			"totalCancelled": cancelledCount,
			"results":        cancelResults,
		}

		if len(errors) > 0 {
			data["errors"] = errors
		}

		c.JSON(http.StatusOK, models.TradeResponse{
			Success:   true,
			Message:   "Orders cancelled",
			Data:      data,
			Timestamp: time.Now().Unix(),
		})
	}
}

// ClosePositionHandler - Close a position
func ClosePositionHandler(bn *binance.Client, fb *firebase.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Symbol  string `json:"symbol" binding:"required"`
			TradeID string `json:"tradeId"` // Optional: link to Firebase trade
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, models.TradeResponse{
				Success:   false,
				Message:   "Invalid request",
				Error:     err.Error(),
				Timestamp: time.Now().Unix(),
			})
			return
		}

		// Close position on Binance
		result, err := bn.ClosePosition(req.Symbol)
		if err != nil {
			c.JSON(http.StatusInternalServerError, models.TradeResponse{
				Success:   false,
				Message:   "Failed to close position",
				Error:     err.Error(),
				Timestamp: time.Now().Unix(),
			})
			return
		}

		// Update trade in Firebase if tradeId provided
		if req.TradeID != "" {
			trade, err := fb.GetTrade(c.Request.Context(), req.TradeID)
			if err == nil {
				trade.Status = "CLOSED"
				trade.ClosedAt = time.Now().Unix()
				trade.PnL = result.RealizedProfit
				fb.UpdateTrade(c.Request.Context(), trade)
			}
		}

		c.JSON(http.StatusOK, models.TradeResponse{
			Success:   true,
			Message:   "Position closed successfully",
			Data:      result,
			Timestamp: time.Now().Unix(),
		})
	}
}

// TradingSummaryHandler - Get trading summary for period
func TradingSummaryHandler(fb *firebase.Client, bn *binance.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		period := c.DefaultQuery("period", "1d") // 1d, 7d, 1w, 1m
		userID := c.Query("userId")              // Optional: filter by user

		// Calculate time range
		now := time.Now()
		var startTime int64

		switch period {
		case "1d":
			startTime = now.AddDate(0, 0, -1).Unix()
		case "7d":
			startTime = now.AddDate(0, 0, -7).Unix()
		case "1w":
			startTime = now.AddDate(0, 0, -7).Unix()
		case "1m":
			startTime = now.AddDate(0, -1, 0).Unix()
		default:
			startTime = now.AddDate(0, 0, -1).Unix()
		}

		// Get trades from Firebase
		var trades []*models.Trade
		var err error

		if userID != "" {
			trades, err = fb.GetUserTrades(c.Request.Context(), userID)
		} else {
			trades, err = fb.GetAllTrades(c.Request.Context())
		}

		if err != nil {
			c.JSON(http.StatusInternalServerError, models.TradeResponse{
				Success:   false,
				Message:   "Failed to get trades",
				Error:     err.Error(),
				Timestamp: time.Now().Unix(),
			})
			return
		}

		// Calculate statistics
		summary := calculateTradingSummary(trades, startTime)

		// Get current account PnL from Binance
		accountPnL, _ := bn.GetAccountPnL()
		summary["currentAccountPnL"] = accountPnL

		c.JSON(http.StatusOK, models.TradeResponse{
			Success:   true,
			Message:   "Trading summary retrieved successfully",
			Data:      summary,
			Timestamp: time.Now().Unix(),
		})
	}
}

// Helper function to calculate trading summary
func calculateTradingSummary(trades []*models.Trade, startTime int64) gin.H {
	totalTrades := 0
	winningTrades := 0
	losingTrades := 0
	totalPnL := 0.0
	totalVolume := 0.0
	bestTrade := 0.0
	worstTrade := 0.0

	symbolStats := make(map[string]int)

	for _, trade := range trades {
		if trade.CreatedAt < startTime {
			continue
		}

		totalTrades++
		totalVolume += trade.Size

		if trade.PnL > 0 {
			winningTrades++
		} else if trade.PnL < 0 {
			losingTrades++
		}

		totalPnL += trade.PnL

		if trade.PnL > bestTrade {
			bestTrade = trade.PnL
		}
		if trade.PnL < worstTrade {
			worstTrade = trade.PnL
		}

		symbolStats[trade.Symbol]++
	}

	winRate := 0.0
	avgPnL := 0.0
	if totalTrades > 0 {
		winRate = (float64(winningTrades) / float64(totalTrades)) * 100
		avgPnL = totalPnL / float64(totalTrades)
	}

	return gin.H{
		"totalTrades":   totalTrades,
		"winningTrades": winningTrades,
		"losingTrades":  losingTrades,
		"winRate":       winRate,
		"totalPnL":      totalPnL,
		"totalVolume":   totalVolume,
		"bestTrade":     bestTrade,
		"worstTrade":    worstTrade,
		"averagePnL":    avgPnL,
		"symbolStats":   symbolStats,
	}
}
