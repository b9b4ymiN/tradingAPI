package firebase

import (
	"context"
	"crypto-trading-api/internal/models"
	"encoding/json"
	"fmt"
	"log"
	"time"
)

// GetAllTrades - Get all trades from Firebase
func (f *Client) GetAllTrades(ctx context.Context) ([]*models.Trade, error) {
	path := "/trades"
	respBody, err := f.makeRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get all trades: %v", err)
	}

	if string(respBody) == "null" || string(respBody) == "" {
		return []*models.Trade{}, nil
	}

	var tradesMap map[string]*models.Trade
	if err := json.Unmarshal(respBody, &tradesMap); err != nil {
		return nil, fmt.Errorf("failed to unmarshal trades: %v", err)
	}

	trades := make([]*models.Trade, 0, len(tradesMap))
	for _, trade := range tradesMap {
		trades = append(trades, trade)
	}

	return trades, nil
}

// GetTradesByStatus - Get trades filtered by status
func (f *Client) GetTradesByStatus(ctx context.Context, status string) ([]*models.Trade, error) {
	// Firebase REST API query by child
	path := fmt.Sprintf("/trades?orderBy=\"status\"&equalTo=\"%s\"", status)
	respBody, err := f.makeRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get trades by status: %v", err)
	}

	if string(respBody) == "null" || string(respBody) == "" {
		return []*models.Trade{}, nil
	}

	var tradesMap map[string]*models.Trade
	if err := json.Unmarshal(respBody, &tradesMap); err != nil {
		return nil, fmt.Errorf("failed to unmarshal trades: %v", err)
	}

	trades := make([]*models.Trade, 0, len(tradesMap))
	for _, trade := range tradesMap {
		trades = append(trades, trade)
	}

	return trades, nil
}

// UpdateTradePnL - Update trade PnL
func (f *Client) UpdateTradePnL(ctx context.Context, tradeID string, pnl float64, userID string) error {
	// Get the trade first
	trade, err := f.GetTrade(ctx, tradeID)
	if err != nil {
		return err
	}

	// Update fields
	trade.PnL = pnl
	trade.Status = "CLOSED"
	trade.ClosedAt = getCurrentTimestamp()

	// Save updated trade
	return f.UpdateTrade(ctx, trade)
}

// GetUserStats - Get user statistics
func (f *Client) GetUserStats(ctx context.Context, userID string) (map[string]interface{}, error) {
	path := fmt.Sprintf("/users/%s/stats", userID)
	respBody, err := f.makeRequest(ctx, "GET", path, nil)
	if err != nil || string(respBody) == "null" || string(respBody) == "" {
		// Return default stats if not found
		return map[string]interface{}{
			"totalTrades":  0,
			"activeTrades": 0,
			"totalPnL":     0.0,
			"winRate":      0.0,
		}, nil
	}

	var stats map[string]interface{}
	if err := json.Unmarshal(respBody, &stats); err != nil {
		return nil, fmt.Errorf("failed to unmarshal stats: %v", err)
	}

	return stats, nil
}

// UpdateUserStats - Update user statistics
func (f *Client) UpdateUserStats(ctx context.Context, userID string, stats map[string]interface{}) error {
	path := fmt.Sprintf("/users/%s/stats", userID)
	_, err := f.makeRequest(ctx, "PUT", path, stats)
	if err != nil {
		return fmt.Errorf("failed to update user stats: %v", err)
	}
	return nil
}

// SaveSystemStats - Save system-wide statistics
func (f *Client) SaveSystemStats(ctx context.Context, stats map[string]interface{}) error {
	stats["lastUpdate"] = getCurrentTimestamp()
	path := "/system/stats"
	_, err := f.makeRequest(ctx, "PUT", path, stats)
	if err != nil {
		return fmt.Errorf("failed to save system stats: %v", err)
	}
	return nil
}

// GetSystemStats - Get system statistics
func (f *Client) GetSystemStats(ctx context.Context) (map[string]interface{}, error) {
	path := "/system/stats"
	respBody, err := f.makeRequest(ctx, "GET", path, nil)
	if err != nil || string(respBody) == "null" || string(respBody) == "" {
		return map[string]interface{}{
			"totalTrades":  0,
			"activeTrades": 0,
			"totalVolume":  0.0,
			"lastUpdate":   getCurrentTimestamp(),
		}, nil
	}

	var stats map[string]interface{}
	if err := json.Unmarshal(respBody, &stats); err != nil {
		return nil, fmt.Errorf("failed to unmarshal system stats: %v", err)
	}

	return stats, nil
}

// BatchUpdateTrades - Update multiple trades at once
func (f *Client) BatchUpdateTrades(ctx context.Context, trades []*models.Trade) error {
	for _, trade := range trades {
		if err := f.UpdateTrade(ctx, trade); err != nil {
			log.Printf("Error updating trade %s: %v", trade.ID, err)
		}
	}
	return nil
}

// CalculateUserStatistics - Calculate and update user statistics
func (f *Client) CalculateUserStatistics(ctx context.Context, userID string) error {
	trades, err := f.GetUserTrades(ctx, userID)
	if err != nil {
		return err
	}

	totalTrades := len(trades)
	activeTrades := 0
	totalPnL := 0.0
	winningTrades := 0

	for _, trade := range trades {
		if trade.Status == "ACTIVE" {
			activeTrades++
		}
		totalPnL += trade.PnL
		if trade.PnL > 0 {
			winningTrades++
		}
	}

	winRate := 0.0
	if totalTrades > 0 {
		winRate = (float64(winningTrades) / float64(totalTrades)) * 100
	}

	stats := map[string]interface{}{
		"totalTrades":  totalTrades,
		"activeTrades": activeTrades,
		"totalPnL":     totalPnL,
		"winRate":      winRate,
		"lastUpdate":   getCurrentTimestamp(),
	}

	return f.UpdateUserStats(ctx, userID, stats)
}

// Helper functions
func getCurrentTimestamp() int64 {
	return currentTimeMillis() / 1000
}

func currentTimeMillis() int64 {
	return currentTime().UnixNano() / 1000000
}

func currentTime() time.Time {
	return time.Now()
}
