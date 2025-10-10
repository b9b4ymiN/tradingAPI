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

	// 3. Place market order
	order, err := b.client.NewCreateOrderService().
		Symbol(trade.Symbol).
		Side(futures.SideType(trade.Side)).
		Type(futures.OrderTypeMarket).
		Quantity(quantity).
		Do(ctx)
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
	quantity := (size * float64(leverage)) / price
	return fmt.Sprintf("%.3f", quantity)
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
