package api

import (
	"context"
	"crypto-trading-api/internal/binance"
	"crypto-trading-api/internal/models"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// FirebaseInterface defines methods needed from Firebase client
type FirebaseInterface interface {
	SaveTrade(ctx context.Context, trade *models.Trade) error
	UpdateTrade(ctx context.Context, trade *models.Trade) error
	GetTrade(ctx context.Context, tradeID string) (*models.Trade, error)
	GetUserTrades(ctx context.Context, userID string) ([]*models.Trade, error)
}

// BinanceInterface defines methods needed from Binance client
type BinanceInterface interface {
	PlaceFuturesOrder(trade *models.Trade) (*binance.OrderResult, error)
	MonitorTrade(trade *models.Trade, fb interface {
		UpdateTrade(ctx context.Context, trade *models.Trade) error
	})
}

// TradeHandler - Main function to handle trade requests
func TradeHandler(fb FirebaseInterface, bn BinanceInterface) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req models.TradeRequest

		// Validate request body
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, models.TradeResponse{
				Success:   false,
				Message:   "Invalid request",
				Error:     err.Error(),
				Timestamp: time.Now().Unix(),
			})
			return
		}

		// Validate trade parameters
		if err := validateTradeParams(&req); err != nil {
			c.JSON(http.StatusBadRequest, models.TradeResponse{
				Success:   false,
				Message:   "Invalid trade parameters",
				Error:     err.Error(),
				Timestamp: time.Now().Unix(),
			})
			return
		}

		// Generate unique trade ID
		tradeID := uuid.New().String()

		// Create trade record
		trade := &models.Trade{
			ID:         tradeID,
			UserID:     req.UserID,
			Symbol:     req.Symbol,
			Side:       req.Side,
			EntryPrice: req.EntryPrice,
			StopLoss:   req.StopLoss,
			TakeProfit: req.TakeProfit,
			Leverage:   req.Leverage,
			Size:       req.Size,
			Status:     "PENDING",
			CreatedAt:  time.Now().Unix(),
		}

		// Execute trade on Binance
		orderResult, err := bn.PlaceFuturesOrder(trade)
		if err != nil {
			trade.Status = "FAILED"
			trade.Error = err.Error()
			fb.SaveTrade(c.Request.Context(), trade)

			c.JSON(http.StatusInternalServerError, models.TradeResponse{
				Success:   false,
				TradeID:   tradeID,
				Message:   "Failed to execute trade",
				Error:     err.Error(),
				Timestamp: time.Now().Unix(),
			})
			return
		}

		// Update trade with order result
		trade.Status = "ACTIVE"
		trade.OrderID = orderResult.OrderID
		trade.ExecutedPrice = orderResult.AvgPrice
		trade.ExecutedAt = time.Now().Unix()

		// Save to Firebase
		if err := fb.SaveTrade(c.Request.Context(), trade); err != nil {
			c.JSON(http.StatusInternalServerError, models.TradeResponse{
				Success:   false,
				TradeID:   tradeID,
				Message:   "Trade executed but failed to save",
				Error:     err.Error(),
				Timestamp: time.Now().Unix(),
			})
			return
		}

		// Start monitoring for SL/TP (in goroutine)
		go bn.MonitorTrade(trade, fb)

		// Success response
		c.JSON(http.StatusOK, models.TradeResponse{
			Success:   true,
			TradeID:   tradeID,
			Message:   "Trade executed successfully",
			Data:      trade,
			Timestamp: time.Now().Unix(),
		})
	}
}

// GetTradesHandler - Get trades for a user
func GetTradesHandler(fb FirebaseInterface) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.Param("userId")

		trades, err := fb.GetUserTrades(c.Request.Context(), userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, models.TradeResponse{
				Success:   false,
				Message:   "Failed to fetch trades",
				Error:     err.Error(),
				Timestamp: time.Now().Unix(),
			})
			return
		}

		c.JSON(http.StatusOK, models.TradeResponse{
			Success:   true,
			Message:   "Trades fetched successfully",
			Data:      trades,
			Timestamp: time.Now().Unix(),
		})
	}
}

// GetTradeHandler - Get single trade
func GetTradeHandler(fb FirebaseInterface) gin.HandlerFunc {
	return func(c *gin.Context) {
		tradeID := c.Param("tradeId")

		trade, err := fb.GetTrade(c.Request.Context(), tradeID)
		if err != nil {
			c.JSON(http.StatusNotFound, models.TradeResponse{
				Success:   false,
				Message:   "Trade not found",
				Error:     err.Error(),
				Timestamp: time.Now().Unix(),
			})
			return
		}

		c.JSON(http.StatusOK, models.TradeResponse{
			Success:   true,
			Message:   "Trade fetched successfully",
			Data:      trade,
			Timestamp: time.Now().Unix(),
		})
	}
}

// Validate trade parameters
func validateTradeParams(req *models.TradeRequest) error {
	if req.Side != "BUY" && req.Side != "SELL" {
		return fmt.Errorf("side must be BUY or SELL")
	}

	if req.EntryPrice <= 0 {
		return fmt.Errorf("entry price must be greater than 0")
	}

	if req.Side == "BUY" {
		if req.StopLoss >= req.EntryPrice {
			return fmt.Errorf("stop loss must be less than entry price for BUY")
		}
		if req.TakeProfit <= req.EntryPrice {
			return fmt.Errorf("take profit must be greater than entry price for BUY")
		}
	} else {
		if req.StopLoss <= req.EntryPrice {
			return fmt.Errorf("stop loss must be greater than entry price for SELL")
		}
		if req.TakeProfit >= req.EntryPrice {
			return fmt.Errorf("take profit must be less than entry price for SELL")
		}
	}

	return nil
}
