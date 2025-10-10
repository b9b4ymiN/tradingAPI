package models

// Trade represents a trading position
type Trade struct {
	ID            string  `json:"id"`
	UserID        string  `json:"userId"`
	Symbol        string  `json:"symbol"`
	Side          string  `json:"side"`
	EntryPrice    float64 `json:"entryPrice"`
	ExecutedPrice float64 `json:"executedPrice,omitempty"`
	StopLoss      float64 `json:"stopLoss"`
	TakeProfit    float64 `json:"takeProfit"`
	Leverage      int     `json:"leverage"`
	Size          float64 `json:"size"`
	Status        string  `json:"status"` // PENDING, ACTIVE, FILLED, CANCELED, FAILED
	OrderID       int64   `json:"orderId,omitempty"`
	Error         string  `json:"error,omitempty"`
	CreatedAt     int64   `json:"createdAt"`
	ExecutedAt    int64   `json:"executedAt,omitempty"`
	ClosedAt      int64   `json:"closedAt,omitempty"`
	PnL           float64 `json:"pnl,omitempty"`
}

// TradeRequest represents incoming trade order
type TradeRequest struct {
	UserID     string  `json:"userId" binding:"required"`
	Symbol     string  `json:"symbol" binding:"required"`    // e.g., "BTCUSDT"
	Side       string  `json:"side" binding:"required"`      // "BUY" or "SELL"
	EntryPrice float64 `json:"entryPrice" binding:"required"`
	StopLoss   float64 `json:"stopLoss" binding:"required"`
	TakeProfit float64 `json:"takeProfit" binding:"required"`
	Leverage   int     `json:"leverage" binding:"required,min=1,max=125"`
	Size       float64 `json:"size" binding:"required,gt=0"` // Position size in USDT
}

// TradeResponse represents API response
type TradeResponse struct {
	Success   bool        `json:"success"`
	TradeID   string      `json:"tradeId,omitempty"`
	Message   string      `json:"message"`
	Data      interface{} `json:"data,omitempty"`
	Error     string      `json:"error,omitempty"`
	Timestamp int64       `json:"timestamp"`
}
