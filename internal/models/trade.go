package models

// Trade represents a trading position
type Trade struct {
	ID            string  `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	UserID        string  `json:"userId" example:"user123"`
	Symbol        string  `json:"symbol" example:"BTCUSDT"`
	Side          string  `json:"side" example:"BUY"`
	OrderType     string  `json:"orderType,omitempty" example:"MARKET"` // MARKET or LIMIT
	MarginType    string  `json:"marginType,omitempty" example:"ISOLATED"` // ISOLATED or CROSSED (default: ISOLATED)
	EntryPrice    float64 `json:"entryPrice" example:"50000.00"`
	ExecutedPrice float64 `json:"executedPrice,omitempty" example:"50100.50"`
	StopLoss      float64 `json:"stopLoss" example:"49000.00"`
	TakeProfit    float64 `json:"takeProfit" example:"52000.00"`
	Leverage      int     `json:"leverage" example:"10"`
	Size          float64 `json:"size" example:"1000.00"`
	Status        string  `json:"status" example:"ACTIVE"` // PENDING, ACTIVE, FILLED, CANCELED, FAILED
	OrderID       int64   `json:"orderId,omitempty" example:"123456789"`
	SLOrderID     int64   `json:"slOrderId,omitempty" example:"123456790"` // Stop Loss order ID
	TPOrderID     int64   `json:"tpOrderId,omitempty" example:"123456791"` // Take Profit order ID
	Error         string  `json:"error,omitempty" example:""`
	CreatedAt     int64   `json:"createdAt" example:"1640995200"`
	ExecutedAt    int64   `json:"executedAt,omitempty" example:"1640995260"`
	ClosedAt      int64   `json:"closedAt,omitempty" example:"1640999800"`
	PnL           float64 `json:"pnl,omitempty" example:"250.75"`
}

// TradeRequest represents incoming trade order
type TradeRequest struct {
	UserID     string  `json:"userId" binding:"required" example:"user123"`
	Symbol     string  `json:"symbol" binding:"required" example:"BTCUSDT"`         // e.g., "BTCUSDT"
	Side       string  `json:"side" binding:"required" example:"BUY"`               // "BUY" or "SELL"
	EntryPrice float64 `json:"entryPrice" binding:"required" example:"50000.00"`    // Entry price
	StopLoss   float64 `json:"stopLoss" binding:"required" example:"49000.00"`      // Stop loss price
	TakeProfit float64 `json:"takeProfit" binding:"required" example:"52000.00"`    // Take profit price
	Leverage   int     `json:"leverage" binding:"required,min=1,max=125" example:"10"` // Leverage (1-125x)
	Size       float64 `json:"size" binding:"required,gt=0" example:"1000.00"`      // Position size in USDT
	OrderType  string  `json:"orderType,omitempty" example:"MARKET"`                // "MARKET" or "LIMIT" (default: MARKET)
	MarginType string  `json:"marginType,omitempty" example:"ISOLATED"`             // "ISOLATED" or "CROSSED" (default: ISOLATED)
	APIKey     string  `json:"apiKey,omitempty" example:"your-api-key-here"`        // Optional: API key for authentication (useful for TradingView alerts)
}

// TradeResponse represents API response
type TradeResponse struct {
	Success   bool        `json:"success" example:"true"`
	TradeID   string      `json:"tradeId,omitempty" example:"550e8400-e29b-41d4-a716-446655440000"`
	Message   string      `json:"message" example:"Trade executed successfully"`
	Data      interface{} `json:"data,omitempty"`
	Error     string      `json:"error,omitempty" example:""`
	Timestamp int64       `json:"timestamp" example:"1640995200"`
}

// CancelOrderRequest represents order cancellation request
type CancelOrderRequest struct {
	Symbol  string `json:"symbol,omitempty" example:"BTCUSDT"`    // Optional: cancel by symbol
	OrderID int64  `json:"orderId,omitempty" example:"123456789"` // Optional: cancel specific order
}

// ClosePositionRequest represents position closure request
type ClosePositionRequest struct {
	Symbol  string `json:"symbol" binding:"required" example:"BTCUSDT"`
	TradeID string `json:"tradeId,omitempty" example:"550e8400-e29b-41d4-a716-446655440000"` // Optional: link to Firebase trade
}
