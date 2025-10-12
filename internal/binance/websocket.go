package binance

import (
	"context"
	"crypto-trading-api/internal/models"
	"fmt"
	"log"
	"strconv"
	"sync"
	"time"

	"github.com/adshao/go-binance/v2/futures"
)

// WebSocketManager manages WebSocket connections
type WebSocketManager struct {
	client           *Client
	userDataStream   *UserDataStream
	priceStreams     map[string]*PriceStream
	mu               sync.RWMutex
	isRunning        bool
	stopChan         chan struct{}
}

// UserDataStream represents user data WebSocket stream
type UserDataStream struct {
	ListenKey    string
	DoneC        chan struct{}
	StopC        chan struct{}
	LastPing     time.Time
	IsConnected  bool
	mu           sync.RWMutex
}

// PriceStream represents market price WebSocket stream
type PriceStream struct {
	Symbol      string
	LastPrice   float64
	LastUpdate  time.Time
	DoneC       chan struct{}
	StopC       chan struct{}
	IsConnected bool
	mu          sync.RWMutex
}

// OrderUpdateEvent represents order update from WebSocket
type OrderUpdateEvent struct {
	Symbol           string
	Side             string
	OrderType        string
	OrderID          int64
	ClientOrderID    string
	Price            string
	Quantity         string
	ExecutedQty      string
	CumulativeQty    string
	Status           string
	TimeInForce      string
	AvgPrice         string
	IsReduceOnly     bool
	WorkingType      string
	OriginalType     string
	PositionSide     string
	IsClosePosition  bool
	RealizedProfit   string
	TransactionTime  int64
}

// AccountUpdateEvent represents account update from WebSocket
type AccountUpdateEvent struct {
	Reason          string
	Balances        []BalanceUpdate
	Positions       []PositionUpdate
	TransactionTime int64
}

// BalanceUpdate represents balance change
type BalanceUpdate struct {
	Asset            string
	WalletBalance    string
	CrossWalletBalance string
	BalanceChange    string
}

// PositionUpdate represents position change
type PositionUpdate struct {
	Symbol           string
	PositionAmount   string
	EntryPrice       string
	UnrealizedPnL    string
	PositionSide     string
}

// NewWebSocketManager creates a new WebSocket manager
func NewWebSocketManager(client *Client) *WebSocketManager {
	return &WebSocketManager{
		client:       client,
		priceStreams: make(map[string]*PriceStream),
		stopChan:     make(chan struct{}),
	}
}

// StartUserDataStream starts the user data WebSocket stream
func (wsm *WebSocketManager) StartUserDataStream(onOrderUpdate func(*OrderUpdateEvent), onAccountUpdate func(*AccountUpdateEvent)) error {
	ctx := context.Background()

	// Get listen key
	listenKey, err := wsm.client.client.NewStartUserStreamService().Do(ctx)
	if err != nil {
		return fmt.Errorf("failed to start user stream: %v", err)
	}

	log.Printf("📡 WebSocket User Data Stream started (listenKey: %s...)", listenKey[:10])

	wsm.userDataStream = &UserDataStream{
		ListenKey:   listenKey,
		LastPing:    time.Now(),
		IsConnected: false,
	}

	// Start keep-alive goroutine (ping every 30 minutes)
	go wsm.keepAliveUserStream()

	// WebSocket handler
	wsHandler := func(event *futures.WsUserDataEvent) {
		wsm.userDataStream.mu.Lock()
		wsm.userDataStream.IsConnected = true
		wsm.userDataStream.mu.Unlock()

		// Handle ORDER_TRADE_UPDATE
		if event.Event == futures.UserDataEventTypeOrderTradeUpdate {
			orderUpdate := &OrderUpdateEvent{
				Symbol:          event.OrderTradeUpdate.Symbol,
				Side:            string(event.OrderTradeUpdate.Side),
				OrderType:       string(event.OrderTradeUpdate.Type),
				OrderID:         event.OrderTradeUpdate.ID,
				ClientOrderID:   event.OrderTradeUpdate.ClientOrderID,
				Price:           event.OrderTradeUpdate.OriginalPrice,
				Quantity:        event.OrderTradeUpdate.OriginalQty,
				ExecutedQty:     event.OrderTradeUpdate.AccumulatedFilledQty,
				Status:          string(event.OrderTradeUpdate.Status),
				AvgPrice:        event.OrderTradeUpdate.AveragePrice,
				IsReduceOnly:    event.OrderTradeUpdate.IsReduceOnly,
				PositionSide:    string(event.OrderTradeUpdate.PositionSide),
				IsClosePosition: false, // Field not available in this SDK version
				RealizedProfit:  event.OrderTradeUpdate.RealizedPnL,
				TransactionTime: event.OrderTradeUpdate.TradeTime,
			}

			log.Printf("🔔 Order Update: %s %s %s - Status: %s",
				orderUpdate.Symbol, orderUpdate.Side, orderUpdate.OrderType, orderUpdate.Status)

			if onOrderUpdate != nil {
				onOrderUpdate(orderUpdate)
			}
		}

		// Handle ACCOUNT_UPDATE
		if event.Event == futures.UserDataEventTypeAccountUpdate {
			balances := []BalanceUpdate{}
			for _, bal := range event.AccountUpdate.Balances {
				balances = append(balances, BalanceUpdate{
					Asset:              bal.Asset,
					WalletBalance:      bal.Balance,
					CrossWalletBalance: bal.CrossWalletBalance,
					BalanceChange:      "0", // Not available in SDK
				})
			}

			positions := []PositionUpdate{}
			for _, pos := range event.AccountUpdate.Positions {
				positions = append(positions, PositionUpdate{
					Symbol:         pos.Symbol,
					PositionAmount: pos.Amount,
					EntryPrice:     pos.EntryPrice,
					UnrealizedPnL:  pos.UnrealizedPnL,
					PositionSide:   string(pos.Side),
				})
			}

			accountUpdate := &AccountUpdateEvent{
				Reason:          string(event.AccountUpdate.Reason),
				Balances:        balances,
				Positions:       positions,
				TransactionTime: event.Time,
			}

			log.Printf("💰 Account Update: %s - Balances: %d, Positions: %d",
				accountUpdate.Reason, len(accountUpdate.Balances), len(accountUpdate.Positions))

			if onAccountUpdate != nil {
				onAccountUpdate(accountUpdate)
			}
		}
	}

	// Error handler
	errHandler := func(err error) {
		log.Printf("⚠️ WebSocket error: %v", err)
		wsm.userDataStream.mu.Lock()
		wsm.userDataStream.IsConnected = false
		wsm.userDataStream.mu.Unlock()

		// Attempt reconnection after 5 seconds
		time.Sleep(5 * time.Second)
		log.Println("🔄 Attempting to reconnect WebSocket...")
		wsm.StartUserDataStream(onOrderUpdate, onAccountUpdate)
	}

	// Start WebSocket
	doneC, stopC, err := futures.WsUserDataServe(listenKey, wsHandler, errHandler)
	if err != nil {
		return fmt.Errorf("failed to serve user data: %v", err)
	}

	wsm.userDataStream.DoneC = doneC
	wsm.userDataStream.StopC = stopC

	log.Println("✅ WebSocket User Data Stream connected")

	return nil
}

// keepAliveUserStream pings the listen key every 30 minutes
func (wsm *WebSocketManager) keepAliveUserStream() {
	ticker := time.NewTicker(30 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if wsm.userDataStream == nil {
				return
			}

			ctx := context.Background()
			err := wsm.client.client.NewKeepaliveUserStreamService().
				ListenKey(wsm.userDataStream.ListenKey).
				Do(ctx)

			if err != nil {
				log.Printf("⚠️ Failed to ping listen key: %v", err)
			} else {
				wsm.userDataStream.mu.Lock()
				wsm.userDataStream.LastPing = time.Now()
				wsm.userDataStream.mu.Unlock()
				log.Println("🏓 WebSocket keep-alive ping sent")
			}

		case <-wsm.stopChan:
			return
		}
	}
}

// StartPriceStream starts a price WebSocket stream for a symbol
func (wsm *WebSocketManager) StartPriceStream(symbol string, onPriceUpdate func(symbol string, price float64)) error {
	wsm.mu.Lock()
	defer wsm.mu.Unlock()

	// Check if already streaming
	if _, exists := wsm.priceStreams[symbol]; exists {
		return fmt.Errorf("price stream already exists for %s", symbol)
	}

	log.Printf("📈 Starting price stream for %s", symbol)

	priceStream := &PriceStream{
		Symbol:      symbol,
		IsConnected: false,
	}

	// WebSocket handler
	wsHandler := func(event *futures.WsMarkPriceEvent) {
		markPrice, _ := strconv.ParseFloat(event.MarkPrice, 64)

		priceStream.mu.Lock()
		priceStream.LastPrice = markPrice
		priceStream.LastUpdate = time.Now()
		priceStream.IsConnected = true
		priceStream.mu.Unlock()

		if onPriceUpdate != nil {
			onPriceUpdate(symbol, markPrice)
		}
	}

	// Error handler
	errHandler := func(err error) {
		log.Printf("⚠️ Price stream error for %s: %v", symbol, err)
		priceStream.mu.Lock()
		priceStream.IsConnected = false
		priceStream.mu.Unlock()
	}

	// Start WebSocket
	doneC, stopC, err := futures.WsMarkPriceServe(symbol, wsHandler, errHandler)
	if err != nil {
		return fmt.Errorf("failed to start price stream: %v", err)
	}

	priceStream.DoneC = doneC
	priceStream.StopC = stopC

	wsm.priceStreams[symbol] = priceStream

	log.Printf("✅ Price stream connected for %s", symbol)

	return nil
}

// StopPriceStream stops a price stream for a symbol
func (wsm *WebSocketManager) StopPriceStream(symbol string) {
	wsm.mu.Lock()
	defer wsm.mu.Unlock()

	if stream, exists := wsm.priceStreams[symbol]; exists {
		close(stream.StopC)
		delete(wsm.priceStreams, symbol)
		log.Printf("🛑 Price stream stopped for %s", symbol)
	}
}

// StopAllStreams stops all WebSocket streams
func (wsm *WebSocketManager) StopAllStreams() {
	wsm.mu.Lock()
	defer wsm.mu.Unlock()

	// Stop user data stream
	if wsm.userDataStream != nil {
		ctx := context.Background()
		wsm.client.client.NewCloseUserStreamService().
			ListenKey(wsm.userDataStream.ListenKey).
			Do(ctx)

		if wsm.userDataStream.StopC != nil {
			close(wsm.userDataStream.StopC)
		}
		wsm.userDataStream = nil
		log.Println("🛑 User data stream stopped")
	}

	// Stop all price streams
	for symbol, stream := range wsm.priceStreams {
		if stream.StopC != nil {
			close(stream.StopC)
		}
		log.Printf("🛑 Price stream stopped for %s", symbol)
	}
	wsm.priceStreams = make(map[string]*PriceStream)

	close(wsm.stopChan)
	log.Println("✅ All WebSocket streams stopped")
}

// GetStreamStatus returns the status of all streams
func (wsm *WebSocketManager) GetStreamStatus() map[string]interface{} {
	wsm.mu.RLock()
	defer wsm.mu.RUnlock()

	status := map[string]interface{}{
		"userDataStream": "disconnected",
		"priceStreams":   []map[string]interface{}{},
	}

	// User data stream status
	if wsm.userDataStream != nil {
		wsm.userDataStream.mu.RLock()
		if wsm.userDataStream.IsConnected {
			status["userDataStream"] = map[string]interface{}{
				"status":   "connected",
				"lastPing": wsm.userDataStream.LastPing.Format(time.RFC3339),
			}
		}
		wsm.userDataStream.mu.RUnlock()
	}

	// Price streams status
	priceStreamsStatus := []map[string]interface{}{}
	for symbol, stream := range wsm.priceStreams {
		stream.mu.RLock()
		streamStatus := map[string]interface{}{
			"symbol":     symbol,
			"connected":  stream.IsConnected,
			"lastPrice":  stream.LastPrice,
			"lastUpdate": stream.LastUpdate.Format(time.RFC3339),
		}
		stream.mu.RUnlock()
		priceStreamsStatus = append(priceStreamsStatus, streamStatus)
	}
	status["priceStreams"] = priceStreamsStatus

	return status
}

// UpdateTradeFromWebSocket updates trade from WebSocket order event
func UpdateTradeFromWebSocket(trade *models.Trade, event *OrderUpdateEvent, fb interface {
	UpdateTrade(ctx context.Context, trade *models.Trade) error
}) {
	ctx := context.Background()

	// Update trade status
	trade.Status = event.Status

	if event.Status == string(futures.OrderStatusTypeFilled) ||
		event.Status == string(futures.OrderStatusTypeCanceled) ||
		event.Status == string(futures.OrderStatusTypeExpired) {
		trade.ClosedAt = time.Now().Unix()
	}

	// Update executed price if available
	if event.AvgPrice != "" && event.AvgPrice != "0" {
		// Parse avg price
		// avgPrice, _ := strconv.ParseFloat(event.AvgPrice, 64)
		// trade.ExecutedPrice = avgPrice
	}

	// Save to Firebase
	if err := fb.UpdateTrade(ctx, trade); err != nil {
		log.Printf("⚠️ Failed to update trade from WebSocket: %v", err)
	} else {
		log.Printf("✅ Trade %s updated from WebSocket: %s", trade.ID, trade.Status)
	}
}
