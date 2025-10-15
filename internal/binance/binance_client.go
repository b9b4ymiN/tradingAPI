package binance

import (
	"context"
	"crypto-trading-api/internal/models"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
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

	// 0. Get symbol precision info
	symbolInfo, err := b.getSymbolInfo(trade.Symbol)
	if err != nil {
		return nil, fmt.Errorf("failed to get symbol info: %v", err)
	}
	log.Printf("📊 Symbol Info - %s: PricePrecision=%d, QuantityPrecision=%d, MinNotional=%s",
		trade.Symbol, symbolInfo.PricePrecision, symbolInfo.QuantityPrecision, symbolInfo.MinNotional)

	// 1. Set margin type (default to ISOLATED if not specified)
	marginType := trade.MarginType
	if marginType == "" {
		marginType = "ISOLATED"
	}

	err = b.client.NewChangeMarginTypeService().
		Symbol(trade.Symbol).
		MarginType(futures.MarginType(marginType)).
		Do(ctx)
	if err != nil {
		// Ignore error if margin type is already set to desired type
		// Error -4046 means "No need to change margin type"
		errStr := err.Error()
		if !strings.Contains(errStr, "-4046") && !strings.Contains(errStr, "No need to change margin type") {
			log.Printf("Warning: Failed to set margin type to %s: %v", marginType, err)
		} else {
			log.Printf("Margin type already set to %s for %s", marginType, trade.Symbol)
		}
	} else {
		log.Printf("✅ Margin type set to %s for %s", marginType, trade.Symbol)
	}

	// 2. Set leverage
	_, err = b.client.NewChangeLeverageService().
		Symbol(trade.Symbol).
		Leverage(trade.Leverage).
		Do(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to set leverage: %v", err)
	}

	// 3. Get current price for MARKET orders (for accurate notional calculation)
	priceForCalculation := trade.EntryPrice
	if trade.OrderType == "" || trade.OrderType == "MARKET" {
		currentPrice, err := b.GetPrice(trade.Symbol)
		if err != nil {
			log.Printf("⚠️ Failed to get current price, using entry price: %v", err)
		} else {
			priceForCalculation = currentPrice
			log.Printf("📊 Using current market price for calculation: %.8f", currentPrice)
		}
	}

	// 3.1 Calculate quantity
	quantity := b.calculateQuantity(trade.Size, priceForCalculation, trade.Leverage, symbolInfo.QuantityPrecision, symbolInfo.StepSize)
	log.Printf("📊 Calculated quantity: %s %s", quantity, trade.Symbol)

	// 3.2 Validate quantity is not zero
	parsedQty, _ := strconv.ParseFloat(quantity, 64)
	if parsedQty == 0 {
		return nil, fmt.Errorf("calculated quantity is zero. Please increase Size. Current: Size=%.2f USDT, Leverage=%dx, Price=%.2f",
			trade.Size, trade.Leverage, priceForCalculation)
	}

	// 3.3 Validate minimum quantity
	minQty, _ := strconv.ParseFloat(symbolInfo.MinQuantity, 64)
	if parsedQty < minQty {
		return nil, fmt.Errorf("quantity (%.8f) is below minimum (%.8f) for %s. Please increase Size from %.2f USDT",
			parsedQty, minQty, trade.Symbol, trade.Size)
	}

	// 3.4 Validate maximum quantity
	maxQty, _ := strconv.ParseFloat(symbolInfo.MaxQuantity, 64)
	if maxQty > 0 && parsedQty > maxQty {
		return nil, fmt.Errorf("quantity (%.8f) exceeds maximum (%.8f) for %s. Please decrease Size",
			parsedQty, maxQty, trade.Symbol)
	}

	// 3.5 Validate minimum notional value (position size)
	minNotional, _ := strconv.ParseFloat(symbolInfo.MinNotional, 64)
	notionalValue := parsedQty * priceForCalculation
	if notionalValue < minNotional {
		return nil, fmt.Errorf("order value (%.2f USDT) is below minimum notional (%.2f USDT) for %s. Please increase Size or Leverage",
			notionalValue, minNotional, trade.Symbol)
	}
	log.Printf("✅ Validation passed - Quantity: %s, Notional: %.2f USDT (min: %.2f USDT)", quantity, notionalValue, minNotional)

	// 3. Place order (MARKET or LIMIT)
	orderService := b.client.NewCreateOrderService().
		Symbol(trade.Symbol).
		Side(futures.SideType(trade.Side)).
		Quantity(quantity)

	// Choose order type based on trade.OrderType
	if trade.OrderType == "LIMIT" {
		// LIMIT order: Wait for specific entry price
		// Format entry price with correct precision
		formattedEntryPrice := b.formatPrice(trade.EntryPrice, symbolInfo.PricePrecision)
		orderService.Type(futures.OrderTypeLimit).
			Price(formattedEntryPrice).
			TimeInForce(futures.TimeInForceTypeGTC) // Good Till Cancel
		log.Printf("📌 Placing LIMIT order: Symbol=%s, Price=%s, Quantity=%s", trade.Symbol, formattedEntryPrice, quantity)
	} else {
		// MARKET order (default): Execute immediately at current price
		orderService.Type(futures.OrderTypeMarket)
		log.Printf("📌 Placing MARKET order: Symbol=%s, Quantity=%s", trade.Symbol, quantity)
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
	log.Printf("📌 Placing Stop Loss order for %s...", trade.Symbol)
	slOrderID, err := b.placeStopLoss(trade.Symbol, trade.Side, quantity, trade.StopLoss, symbolInfo.PricePrecision)
	if err != nil {
		log.Printf("❌ Failed to place SL order: %v", err)
		// Don't fail the entire trade, just log the error
	} else {
		result.SLOrderID = slOrderID
	}

	// 6. Place Take Profit order
	log.Printf("📌 Placing Take Profit order for %s...", trade.Symbol)
	tpOrderID, err := b.placeTakeProfit(trade.Symbol, trade.Side, quantity, trade.TakeProfit, symbolInfo.PricePrecision)
	if err != nil {
		log.Printf("❌ Failed to place TP order: %v", err)
		// Don't fail the entire trade, just log the error
	} else {
		result.TPOrderID = tpOrderID
	}

	return result, nil
}

// Place Stop Loss order
func (b *Client) placeStopLoss(symbol, side, quantity string, stopPrice float64, pricePrecision int) (int64, error) {
	ctx := context.Background()

	// Reverse side for closing position
	closeSide := futures.SideTypeSell
	if side == "SELL" {
		closeSide = futures.SideTypeBuy
	}

	// Format stop price with correct precision
	formattedStopPrice := b.formatPrice(stopPrice, pricePrecision)

	// Use ClosePosition(true) to automatically close the entire position
	// Do NOT specify Quantity when using ClosePosition
	order, err := b.client.NewCreateOrderService().
		Symbol(symbol).
		Side(closeSide).
		Type(futures.OrderTypeStopMarket).
		StopPrice(formattedStopPrice).
		ClosePosition(true).
		Do(ctx)

	if err != nil {
		return 0, fmt.Errorf("failed to place SL order: %v", err)
	}

	log.Printf("✅ Stop Loss order placed: OrderID=%d, Symbol=%s, StopPrice=%s", order.OrderID, symbol, formattedStopPrice)
	return order.OrderID, nil
}

// Place Take Profit order
func (b *Client) placeTakeProfit(symbol, side, quantity string, tpPrice float64, pricePrecision int) (int64, error) {
	ctx := context.Background()

	// Reverse side for closing position
	closeSide := futures.SideTypeSell
	if side == "SELL" {
		closeSide = futures.SideTypeBuy
	}

	// Format TP price with correct precision
	formattedTPPrice := b.formatPrice(tpPrice, pricePrecision)

	// Use ClosePosition(true) to automatically close the entire position
	// Do NOT specify Quantity when using ClosePosition
	order, err := b.client.NewCreateOrderService().
		Symbol(symbol).
		Side(closeSide).
		Type(futures.OrderTypeTakeProfitMarket).
		StopPrice(formattedTPPrice).
		ClosePosition(true).
		Do(ctx)

	if err != nil {
		return 0, fmt.Errorf("failed to place TP order: %v", err)
	}

	log.Printf("✅ Take Profit order placed: OrderID=%d, Symbol=%s, TPPrice=%s", order.OrderID, symbol, formattedTPPrice)
	return order.OrderID, nil
}

// getSymbolInfo - Get symbol precision information
func (b *Client) getSymbolInfo(symbol string) (*SymbolInfo, error) {
	exchangeInfo, err := b.GetExchangeInfo(symbol)
	if err != nil {
		return nil, err
	}

	if len(exchangeInfo.Symbols) == 0 {
		return nil, fmt.Errorf("symbol %s not found", symbol)
	}

	return &exchangeInfo.Symbols[0], nil
}

// formatPrice - Format price with correct precision
func (b *Client) formatPrice(price float64, precision int) string {
	formatStr := fmt.Sprintf("%%.%df", precision)
	return fmt.Sprintf(formatStr, price)
}

// Calculate position quantity based on size and leverage
func (b *Client) calculateQuantity(size, price float64, leverage int, quantityPrecision int, stepSize string) string {
	// Calculate quantity: (position size in USDT * leverage) / price
	quantity := (size * float64(leverage)) / price

	// Parse step size
	step, _ := strconv.ParseFloat(stepSize, 64)
	if step <= 0 {
		step = 1.0 / float64(pow10(quantityPrecision))
	}

	// Round quantity to nearest step size
	// Example: if stepSize=0.001, quantity=0.0018 → 0.002
	quantity = roundToStepSize(quantity, step)

	// Calculate the minimum quantity based on step size
	minQuantity := step

	// If quantity is less than minimum, round UP to minimum
	if quantity < minQuantity {
		quantity = minQuantity
		log.Printf("⚠️ Quantity too small (%.8f), rounded up to minimum: %.8f", (size*float64(leverage))/price, quantity)
	}

	// Format with symbol's quantity precision
	formatStr := fmt.Sprintf("%%.%df", quantityPrecision)
	formattedQty := fmt.Sprintf(formatStr, quantity)

	// Parse back to verify it's not zero
	parsedQty, _ := strconv.ParseFloat(formattedQty, 64)
	if parsedQty == 0 {
		// Force to minimum quantity
		return fmt.Sprintf(formatStr, minQuantity)
	}

	return formattedQty
}

// roundToStepSize rounds a value to the nearest step size
func roundToStepSize(value, stepSize float64) float64 {
	if stepSize == 0 {
		return value
	}
	return float64(int64(value/stepSize+0.5)) * stepSize
}

// Helper function to calculate 10^n
func pow10(n int) int {
	result := 1
	for i := 0; i < n; i++ {
		result *= 10
	}
	return result
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
