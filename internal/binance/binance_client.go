package binance

import (
	"context"
	"crypto-trading-api/internal/models"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/adshao/go-binance/v2/futures"
)

type Client struct {
	client *futures.Client
}

// OrderResult represents the result of a futures order
type OrderResult struct {
	OrderID     int64
	AvgPrice    float64
	ExecutedQty string
	Status      string
	SLOrderID   int64
	TPOrderID   int64
}

func InitClient() *Client {
	apiKey := os.Getenv("BINANCE_API_KEY")
	secretKey := os.Getenv("BINANCE_SECRET_KEY")
	useTestnet := os.Getenv("BINANCE_TESTNET") // Add testnet support

	if apiKey == "" || secretKey == "" {
		log.Fatal("BINANCE_API_KEY and BINANCE_SECRET_KEY must be set")
	}

	// Enable testnet if configured
	if useTestnet == "true" || useTestnet == "1" {
		futures.UseTestnet = true
		log.Println("🔧 Using Binance TESTNET")
	} else {
		log.Println("🔧 Using Binance PRODUCTION")
	}

	client := futures.NewClient(apiKey, secretKey)

	// Test connection
	if err := testBinanceConnection(client); err != nil {
		log.Fatalf("Failed to connect to Binance: %v", err)
	}

	log.Println("✅ Binance client initialized successfully")

	return &Client{client: client}
}

func testBinanceConnection(client *futures.Client) error {
	_, err := client.NewExchangeInfoService().Do(context.Background())
	return err
}

// PlaceFuturesOrder - Execute market order with SL/TP
func (b *Client) PlaceFuturesOrder(trade *models.Trade) (*OrderResult, error) {
	ctx := context.Background()

	// 1. Set leverage
	_, err := b.client.NewChangeLeverageService().
		Symbol(trade.Symbol).
		Leverage(trade.Leverage).
		Do(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to set leverage: %v", err)
	}

	// 2. Calculate quantity
	quantity := b.calculateQuantity(trade.Size, trade.EntryPrice, trade.Leverage)

	// 2.1 Validate minimum notional value (position size)
	// Different symbols have different minimums:
	// - BTCUSDT, ETHUSDT: $100 minimum
	// - XRP, TRX, ADA, BNB, etc.: $5 minimum
	// - Check Binance docs for specific symbol requirements

	// 3. Place order (MARKET or LIMIT)
	orderService := b.client.NewCreateOrderService().
		Symbol(trade.Symbol).
		Side(futures.SideType(trade.Side)).
		Quantity(quantity)

	// Choose order type based on trade.OrderType
	if trade.OrderType == "LIMIT" {
		// LIMIT order: Wait for specific entry price
		orderService.Type(futures.OrderTypeLimit).
			Price(fmt.Sprintf("%.8f", trade.EntryPrice)).
			TimeInForce(futures.TimeInForceTypeGTC) // Good Till Cancel
	} else {
		// MARKET order (default): Execute immediately at current price
		orderService.Type(futures.OrderTypeMarket)
	}

	order, err := orderService.Do(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to place order: %v", err)
	}

	// 4. Get executed price
	avgPrice, _ := strconv.ParseFloat(order.AvgPrice, 64)

	result := &OrderResult{
		OrderID:     order.OrderID,
		AvgPrice:    avgPrice,
		ExecutedQty: order.ExecutedQuantity,
		Status:      string(order.Status),
	}

	// 5. Place Stop Loss order
	slOrderID, err := b.placeStopLoss(trade.Symbol, trade.Side, quantity, trade.StopLoss)
	if err != nil {
		log.Printf("Warning: Failed to place SL order: %v", err)
	} else {
		result.SLOrderID = slOrderID
	}

	// 6. Place Take Profit order
	tpOrderID, err := b.placeTakeProfit(trade.Symbol, trade.Side, quantity, trade.TakeProfit)
	if err != nil {
		log.Printf("Warning: Failed to place TP order: %v", err)
	} else {
		result.TPOrderID = tpOrderID
	}

	return result, nil
}

// Place Stop Loss order
func (b *Client) placeStopLoss(symbol, side, quantity string, stopPrice float64) (int64, error) {
	ctx := context.Background()

	// Reverse side for closing position
	closeSide := futures.SideTypeSell
	if side == "SELL" {
		closeSide = futures.SideTypeBuy
	}

	order, err := b.client.NewCreateOrderService().
		Symbol(symbol).
		Side(closeSide).
		Type(futures.OrderTypeStopMarket).
		StopPrice(fmt.Sprintf("%.8f", stopPrice)).
		Quantity(quantity).
		ClosePosition(true).
		Do(ctx)

	if err != nil {
		return 0, err
	}

	return order.OrderID, nil
}

// Place Take Profit order
func (b *Client) placeTakeProfit(symbol, side, quantity string, tpPrice float64) (int64, error) {
	ctx := context.Background()

	// Reverse side for closing position
	closeSide := futures.SideTypeSell
	if side == "SELL" {
		closeSide = futures.SideTypeBuy
	}

	order, err := b.client.NewCreateOrderService().
		Symbol(symbol).
		Side(closeSide).
		Type(futures.OrderTypeTakeProfitMarket).
		StopPrice(fmt.Sprintf("%.8f", tpPrice)).
		Quantity(quantity).
		ClosePosition(true).
		Do(ctx)

	if err != nil {
		return 0, err
	}

	return order.OrderID, nil
}

// Calculate position quantity based on size and leverage
func (b *Client) calculateQuantity(size, price float64, leverage int) string {
	// Calculate quantity: (position size in USDT * leverage) / price
	quantity := (size * float64(leverage)) / price

	// Round to reasonable precision based on quantity size
	// Different symbols have different precision requirements:
	// - BTC: 3 decimals (0.001)
	// - XRP, ADA: 1 decimal (0.1)
	// - TRX: 0 decimals (1)
	var precision int
	if quantity < 1 {
		precision = 3 // Small quantities (BTC, ETH)
	} else if quantity < 100 {
		precision = 1 // Medium quantities (XRP, ADA, BNB)
	} else {
		precision = 0 // Large quantities (TRX, DOGE)
	}

	// Format with determined precision
	formatStr := fmt.Sprintf("%%.%df", precision)
	return fmt.Sprintf(formatStr, quantity)
}

// MonitorTrade - Monitor trade and update status in Firebase
// Note: fb should be interface or concrete type from firebase package
func (b *Client) MonitorTrade(trade *models.Trade, fb interface {
	UpdateTrade(ctx context.Context, trade *models.Trade) error
}) {
	ctx := context.Background()
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		// Check order status
		order, err := b.client.NewGetOrderService().
			Symbol(trade.Symbol).
			OrderID(trade.OrderID).
			Do(ctx)

		if err != nil {
			log.Printf("Error checking order status: %v", err)
			continue
		}

		// Update trade status
		if order.Status != futures.OrderStatusTypeNew &&
			order.Status != futures.OrderStatusTypePartiallyFilled {

			trade.Status = string(order.Status)
			trade.ClosedAt = time.Now().Unix()

			if err := fb.UpdateTrade(ctx, trade); err != nil {
				log.Printf("Error updating trade: %v", err)
			}

			// Stop monitoring if trade is closed
			if order.Status == futures.OrderStatusTypeFilled ||
				order.Status == futures.OrderStatusTypeCanceled {
				log.Printf("Trade %s closed with status: %s", trade.ID, order.Status)
				return
			}
		}
	}
}

// GetPrice - Get current price
func (b *Client) GetPrice(symbol string) (float64, error) {
	prices, err := b.client.NewListPricesService().Symbol(symbol).Do(context.Background())
	if err != nil {
		return 0, err
	}

	if len(prices) == 0 {
		return 0, fmt.Errorf("no price data for symbol %s", symbol)
	}

	price, err := strconv.ParseFloat(prices[0].Price, 64)
	return price, err
}

// GetBinanceServerTime - Get Binance server time
func (b *Client) GetBinanceServerTime() (int64, error) {
	ctx := context.Background()
	serverTime, err := b.client.NewServerTimeService().Do(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to get server time: %v", err)
	}
	return serverTime, nil
}

// SyncTime - Sync local time with Binance server and return offset
func (b *Client) SyncTime() (int64, error) {
	localTime := time.Now().UnixMilli()

	serverTime, err := b.GetBinanceServerTime()
	if err != nil {
		return 0, err
	}

	offset := serverTime - localTime

	log.Printf("⏰ Time sync: Local=%d, Server=%d, Offset=%dms", localTime, serverTime, offset)

	if absInt64(offset) > 1000 {
		log.Printf("⚠️ Clock drift detected: %dms. Consider syncing system clock.", offset)
	}

	return offset, nil
}

// CheckTimeSyncStatus - Check if time is within acceptable range
func (b *Client) CheckTimeSyncStatus() (bool, int64, error) {
	offset, err := b.SyncTime()
	if err != nil {
		return false, 0, err
	}

	// Binance recommends recvWindow <= 5000ms
	// Clock drift should be less than 1000ms
	isInSync := absInt64(offset) < 1000

	if !isInSync {
		log.Printf("❌ Time not in sync! Offset: %dms (max recommended: 1000ms)", offset)
	} else {
		log.Printf("✅ Time in sync. Offset: %dms", offset)
	}

	return isInSync, offset, nil
}

// Helper function to get absolute value of int64
func absInt64(x int64) int64 {
	if x < 0 {
		return -x
	}
	return x
}
