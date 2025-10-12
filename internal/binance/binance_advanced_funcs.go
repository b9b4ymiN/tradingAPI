package binance

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"

	"github.com/adshao/go-binance/v2/futures"
)

// AccountInfo represents Binance account information
type AccountInfo struct {
	TotalWalletBalance   float64
	AvailableBalance     float64
	TotalUnrealizedPnL   float64
	TotalMarginBalance   float64
	TotalPositionValue   float64
	CanTrade             bool
	CanDeposit           bool
	CanWithdraw          bool
}

// PositionInfo represents position details
type PositionInfo struct {
	Symbol            string
	PositionSide      string
	PositionAmt       float64
	EntryPrice        float64
	MarkPrice         float64
	UnrealizedProfit  float64
	Leverage          int
	LiquidationPrice  float64
	MarginType        string
}

// BalanceInfo represents account balance
type BalanceInfo struct {
	TotalBalance         float64   `json:"totalBalance"`
	AvailableBalance     float64   `json:"availableBalance"`
	TotalUnrealizedPnL   float64   `json:"totalUnrealizedPnL"`
	TotalMarginBalance   float64   `json:"totalMarginBalance"`
	TotalPositionValue   float64   `json:"totalPositionValue"`
	Assets               []AssetBalance `json:"assets"`
}

type AssetBalance struct {
	Asset              string  `json:"asset"`
	WalletBalance      float64 `json:"walletBalance"`
	UnrealizedProfit   float64 `json:"unrealizedProfit"`
	MarginBalance      float64 `json:"marginBalance"`
	AvailableBalance   float64 `json:"availableBalance"`
}

// ClosePositionResult represents the result of closing a position
type ClosePositionResult struct {
	Symbol          string  `json:"symbol"`
	OrderID         int64   `json:"orderId"`
	Side            string  `json:"side"`
	PositionSide    string  `json:"positionSide"`
	Quantity        string  `json:"quantity"`
	Price           string  `json:"price"`
	Status          string  `json:"status"`
	RealizedProfit  float64 `json:"realizedProfit"`
}

// GetServerTime - Get Binance server time
func (b *Client) GetServerTime() (int64, error) {
	serverTime, err := b.client.NewServerTimeService().Do(context.Background())
	if err != nil {
		return 0, err
	}
	return serverTime, nil
}

// GetAccountInfo - Get account information
func (b *Client) GetAccountInfo() (*AccountInfo, error) {
	ctx := context.Background()
	account, err := b.client.NewGetAccountService().Do(ctx)
	if err != nil {
		return nil, err
	}

	totalWalletBalance, _ := strconv.ParseFloat(account.TotalWalletBalance, 64)
	availableBalance, _ := strconv.ParseFloat(account.AvailableBalance, 64)
	totalUnrealizedPnL, _ := strconv.ParseFloat(account.TotalUnrealizedProfit, 64)
	totalMarginBalance, _ := strconv.ParseFloat(account.TotalMarginBalance, 64)
	totalPositionValue, _ := strconv.ParseFloat(account.TotalPositionInitialMargin, 64)

	return &AccountInfo{
		TotalWalletBalance: totalWalletBalance,
		AvailableBalance:   availableBalance,
		TotalUnrealizedPnL: totalUnrealizedPnL,
		TotalMarginBalance: totalMarginBalance,
		TotalPositionValue: totalPositionValue,
		CanTrade:           account.CanTrade,
		CanDeposit:         account.CanDeposit,
		CanWithdraw:        account.CanWithdraw,
	}, nil
}

// CalculateBalance - Calculate detailed balance information
func (b *Client) CalculateBalance(account *AccountInfo) *BalanceInfo {
	ctx := context.Background()
	
	// Get all assets
	accountData, err := b.client.NewGetAccountService().Do(ctx)
	if err != nil {
		return &BalanceInfo{
			TotalBalance:       account.TotalWalletBalance,
			AvailableBalance:   account.AvailableBalance,
			TotalUnrealizedPnL: account.TotalUnrealizedPnL,
			TotalMarginBalance: account.TotalMarginBalance,
			TotalPositionValue: account.TotalPositionValue,
			Assets:             []AssetBalance{},
		}
	}

	assets := []AssetBalance{}
	for _, asset := range accountData.Assets {
		walletBalance, _ := strconv.ParseFloat(asset.WalletBalance, 64)
		unrealizedProfit, _ := strconv.ParseFloat(asset.UnrealizedProfit, 64)
		marginBalance, _ := strconv.ParseFloat(asset.MarginBalance, 64)
		availableBalance, _ := strconv.ParseFloat(asset.AvailableBalance, 64)

		if walletBalance > 0 || unrealizedProfit != 0 {
			assets = append(assets, AssetBalance{
				Asset:            asset.Asset,
				WalletBalance:    walletBalance,
				UnrealizedProfit: unrealizedProfit,
				MarginBalance:    marginBalance,
				AvailableBalance: availableBalance,
			})
		}
	}

	return &BalanceInfo{
		TotalBalance:       account.TotalWalletBalance,
		AvailableBalance:   account.AvailableBalance,
		TotalUnrealizedPnL: account.TotalUnrealizedPnL,
		TotalMarginBalance: account.TotalMarginBalance,
		TotalPositionValue: account.TotalPositionValue,
		Assets:             assets,
	}
}

// GetOpenPositions - Get all open positions
func (b *Client) GetOpenPositions() ([]*PositionInfo, error) {
	ctx := context.Background()
	positions, err := b.client.NewGetPositionRiskService().Do(ctx)
	if err != nil {
		return nil, err
	}

	result := []*PositionInfo{}
	for _, pos := range positions {
		posAmt, _ := strconv.ParseFloat(pos.PositionAmt, 64)
		if posAmt == 0 {
			continue // Skip closed positions
		}

		entryPrice, _ := strconv.ParseFloat(pos.EntryPrice, 64)
		markPrice, _ := strconv.ParseFloat(pos.MarkPrice, 64)
		unrealizedProfit, _ := strconv.ParseFloat(pos.UnRealizedProfit, 64)
		leverage, _ := strconv.Atoi(pos.Leverage)
		liquidationPrice, _ := strconv.ParseFloat(pos.LiquidationPrice, 64)

		result = append(result, &PositionInfo{
			Symbol:           pos.Symbol,
			PositionSide:     pos.PositionSide,
			PositionAmt:      posAmt,
			EntryPrice:       entryPrice,
			MarkPrice:        markPrice,
			UnrealizedProfit: unrealizedProfit,
			Leverage:         leverage,
			LiquidationPrice: liquidationPrice,
			MarginType:       pos.MarginType,
		})
	}

	return result, nil
}

// GetOpenOrders - Get all open orders (pending orders)
func (b *Client) GetOpenOrders(symbol string) ([]*futures.Order, error) {
	ctx := context.Background()
	service := b.client.NewListOpenOrdersService()
	
	if symbol != "" {
		service.Symbol(symbol)
	}

	orders, err := service.Do(ctx)
	if err != nil {
		return nil, err
	}

	return orders, nil
}

// CancelOrder - Cancel a specific order
func (b *Client) CancelOrder(symbol string, orderID int64) error {
	ctx := context.Background()
	_, err := b.client.NewCancelOrderService().
		Symbol(symbol).
		OrderID(orderID).
		Do(ctx)
	
	return err
}

// CancelAllOrders - Cancel all orders for a symbol
func (b *Client) CancelAllOrders(symbol string) (int, error) {
	ctx := context.Background()
	
	err := b.client.NewCancelAllOpenOrdersService().
		Symbol(symbol).
		Do(ctx)
	
	if err != nil {
		return 0, err
	}

	// Get count of cancelled orders (before cancellation)
	orders, _ := b.GetOpenOrders(symbol)
	return len(orders), nil
}

// GetActiveSymbols - Get list of symbols with open positions or orders
func (b *Client) GetActiveSymbols() ([]string, error) {
	ctx := context.Background()
	
	positions, err := b.client.NewGetPositionRiskService().Do(ctx)
	if err != nil {
		return nil, err
	}

	symbolMap := make(map[string]bool)
	
	for _, pos := range positions {
		posAmt, _ := strconv.ParseFloat(pos.PositionAmt, 64)
		if posAmt != 0 {
			symbolMap[pos.Symbol] = true
		}
	}

	symbols := []string{}
	for symbol := range symbolMap {
		symbols = append(symbols, symbol)
	}

	return symbols, nil
}

// ClosePosition - Close an open position
func (b *Client) ClosePosition(symbol string) (*ClosePositionResult, error) {
	ctx := context.Background()

	// Get current position
	positions, err := b.client.NewGetPositionRiskService().Symbol(symbol).Do(ctx)
	if err != nil {
		return nil, err
	}

	if len(positions) == 0 {
		return nil, fmt.Errorf("no position found for symbol %s", symbol)
	}

	position := positions[0]
	posAmt, _ := strconv.ParseFloat(position.PositionAmt, 64)

	if posAmt == 0 {
		return nil, fmt.Errorf("no open position for symbol %s", symbol)
	}

	// Determine close side (opposite of position)
	closeSide := futures.SideTypeSell
	if posAmt < 0 {
		closeSide = futures.SideTypeBuy
	}

	// Place market order to close position
	order, err := b.client.NewCreateOrderService().
		Symbol(symbol).
		Side(closeSide).
		Type(futures.OrderTypeMarket).
		Quantity(fmt.Sprintf("%.3f", absFloat(posAmt))).
		ReduceOnly(true).
		Do(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to close position: %v", err)
	}

	avgPrice, _ := strconv.ParseFloat(order.AvgPrice, 64)
	
	// Calculate realized profit
	entryPrice, _ := strconv.ParseFloat(position.EntryPrice, 64)
	realizedProfit := (avgPrice - entryPrice) * posAmt

	return &ClosePositionResult{
		Symbol:         symbol,
		OrderID:        order.OrderID,
		Side:           string(order.Side),
		PositionSide:   string(order.PositionSide),
		Quantity:       order.ExecutedQuantity,
		Price:          order.AvgPrice,
		Status:         string(order.Status),
		RealizedProfit: realizedProfit,
	}, nil
}

// GetAccountPnL - Get current account total PnL
func (b *Client) GetAccountPnL() (float64, error) {
	account, err := b.GetAccountInfo()
	if err != nil {
		return 0, err
	}
	return account.TotalUnrealizedPnL, nil
}

// Helper function to get absolute value
func absFloat(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

// GetTradeHistory - Get trade history for period
func (b *Client) GetTradeHistory(symbol string, startTime, endTime int64) ([]*futures.AccountTrade, error) {
	ctx := context.Background()
	
	service := b.client.NewListAccountTradeService().
		Symbol(symbol).
		StartTime(startTime * 1000). // Convert to milliseconds
		EndTime(endTime * 1000)

	trades, err := service.Do(ctx)
	if err != nil {
		return nil, err
	}

	return trades, nil
}

// GetIncomeHistory - Get income history (PnL history)
func (b *Client) GetIncomeHistory(symbol string, startTime, endTime int64) (float64, error) {
	ctx := context.Background()
	
	service := b.client.NewGetIncomeHistoryService().
		StartTime(startTime * 1000). // Convert to milliseconds
		EndTime(endTime * 1000).
		IncomeType("REALIZED_PNL")

	if symbol != "" {
		service.Symbol(symbol)
	}

	incomes, err := service.Do(ctx)
	if err != nil {
		return 0, err
	}

	totalPnL := 0.0
	for _, income := range incomes {
		pnl, _ := strconv.ParseFloat(income.Income, 64)
		totalPnL += pnl
	}

	return totalPnL, nil
}

// SymbolInfo represents trading rules for a symbol
type SymbolInfo struct {
	Symbol              string  `json:"symbol"`
	Status              string  `json:"status"`
	BaseAsset           string  `json:"baseAsset"`
	QuoteAsset          string  `json:"quoteAsset"`
	PricePrecision      int     `json:"pricePrecision"`
	QuantityPrecision   int     `json:"quantityPrecision"`
	MinQuantity         string  `json:"minQuantity"`
	MaxQuantity         string  `json:"maxQuantity"`
	StepSize            string  `json:"stepSize"`
	MinNotional         string  `json:"minNotional"`
	MinPrice            string  `json:"minPrice"`
	MaxPrice            string  `json:"maxPrice"`
	TickSize            string  `json:"tickSize"`
}

// ExchangeInfoResponse represents the exchange info response
type ExchangeInfoResponse struct {
	Timezone   string       `json:"timezone"`
	ServerTime int64        `json:"serverTime"`
	Symbols    []SymbolInfo `json:"symbols"`
}

// GetExchangeInfo - Get exchange trading rules and symbol information
func (b *Client) GetExchangeInfo(symbol string) (*ExchangeInfoResponse, error) {
	ctx := context.Background()

	// Get exchange info from Binance
	exchangeInfo, err := b.client.NewExchangeInfoService().Do(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get exchange info: %v", err)
	}

	response := &ExchangeInfoResponse{
		Timezone:   exchangeInfo.Timezone,
		ServerTime: exchangeInfo.ServerTime,
		Symbols:    []SymbolInfo{},
	}

	// Process symbols
	for _, s := range exchangeInfo.Symbols {
		// If specific symbol requested, filter it
		if symbol != "" && s.Symbol != symbol {
			continue
		}

		// Extract filters
		filters := make(map[string]map[string]interface{})
		for _, filter := range s.Filters {
			if filterType, ok := filter["filterType"].(string); ok {
				filters[filterType] = filter
			}
		}

		// Build symbol info
		symbolInfo := SymbolInfo{
			Symbol:            s.Symbol,
			Status:            string(s.Status),
			BaseAsset:         s.BaseAsset,
			QuoteAsset:        s.QuoteAsset,
			PricePrecision:    s.PricePrecision,
			QuantityPrecision: s.QuantityPrecision,
		}

		// Extract LOT_SIZE filter (quantity rules)
		if lotSize, ok := filters["LOT_SIZE"]; ok {
			if minQty, exists := lotSize["minQty"]; exists {
				if val, ok := minQty.(string); ok {
					symbolInfo.MinQuantity = val
				}
			}
			if maxQty, exists := lotSize["maxQty"]; exists {
				if val, ok := maxQty.(string); ok {
					symbolInfo.MaxQuantity = val
				}
			}
			if stepSize, exists := lotSize["stepSize"]; exists {
				if val, ok := stepSize.(string); ok {
					symbolInfo.StepSize = val
				}
			}
		}

		// Extract PRICE_FILTER (price rules)
		if priceFilter, ok := filters["PRICE_FILTER"]; ok {
			if minPrice, exists := priceFilter["minPrice"]; exists {
				if val, ok := minPrice.(string); ok {
					symbolInfo.MinPrice = val
				}
			}
			if maxPrice, exists := priceFilter["maxPrice"]; exists {
				if val, ok := maxPrice.(string); ok {
					symbolInfo.MaxPrice = val
				}
			}
			if tickSize, exists := priceFilter["tickSize"]; exists {
				if val, ok := tickSize.(string); ok {
					symbolInfo.TickSize = val
				}
			}
		}

		// Extract MIN_NOTIONAL filter (minimum order value)
		if minNotional, ok := filters["MIN_NOTIONAL"]; ok {
			if notional, exists := minNotional["notional"]; exists {
				if val, ok := notional.(string); ok {
					symbolInfo.MinNotional = val
				}
			}
		}

		response.Symbols = append(response.Symbols, symbolInfo)
	}

	return response, nil
}

// AccountSnapshotAsset represents asset information in snapshot
type AccountSnapshotAsset struct {
	Asset              string  `json:"asset"`
	MarginBalance      float64 `json:"marginBalance,string"`
	WalletBalance      float64 `json:"walletBalance,string"`
	UnrealizedProfit   float64 `json:"unrealizedProfit,string"`
	AvailableBalance   float64 `json:"availableBalance,string"`
	MaxWithdrawAmount  float64 `json:"maxWithdrawAmount,string"`
}

// AccountSnapshotPosition represents position information in snapshot
type AccountSnapshotPosition struct {
	Symbol           string  `json:"symbol"`
	EntryPrice       float64 `json:"entryPrice,string"`
	MarkPrice        float64 `json:"markPrice,string"`
	PositionAmt      float64 `json:"positionAmt,string"`
	UnrealizedProfit float64 `json:"unRealizedProfit,string"`
	PositionSide     string  `json:"positionSide"`
}

// AccountSnapshotData represents snapshot data for a specific time
type AccountSnapshotData struct {
	Assets    []AccountSnapshotAsset    `json:"assets"`
	Position  []AccountSnapshotPosition `json:"position"`
	UpdateTime int64                    `json:"updateTime"`
}

// AccountSnapshot represents a single snapshot entry
type AccountSnapshot struct {
	Type       string              `json:"type"`
	UpdateTime int64               `json:"updateTime"`
	Data       AccountSnapshotData `json:"data"`
}

// AccountSnapshotResponse represents the full snapshot response
type AccountSnapshotResponse struct {
	Code         int               `json:"code"`
	Msg          string            `json:"msg"`
	SnapshotVos  []AccountSnapshot `json:"snapshotVos"`
}

// FundingRateInfo represents funding rate information
type FundingRateInfo struct {
	Symbol          string  `json:"symbol"`
	FundingRate     float64 `json:"fundingRate"`
	FundingTime     int64   `json:"fundingTime"`
	NextFundingTime int64   `json:"nextFundingTime"`
	MarkPrice       float64 `json:"markPrice"`
	IndexPrice      float64 `json:"indexPrice"`
}

// FundingRateHistory represents historical funding rate
type FundingRateHistory struct {
	Symbol      string  `json:"symbol"`
	FundingRate float64 `json:"fundingRate"`
	FundingTime int64   `json:"fundingTime"`
}

// LiquidationRisk represents liquidation risk information
type LiquidationRisk struct {
	Symbol              string  `json:"symbol"`
	PositionSize        float64 `json:"positionSize"`
	EntryPrice          float64 `json:"entryPrice"`
	MarkPrice           float64 `json:"markPrice"`
	LiquidationPrice    float64 `json:"liquidationPrice"`
	MarginRatio         float64 `json:"marginRatio"`
	UnrealizedPnL       float64 `json:"unrealizedPnl"`
	Leverage            int     `json:"leverage"`
	DistanceToLiquidation float64 `json:"distanceToLiquidation"` // Percentage
	RiskLevel           string  `json:"riskLevel"` // LOW, MEDIUM, HIGH, CRITICAL
}

// GetFundingRate - Get current funding rate for a symbol
func (b *Client) GetFundingRate(symbol string) (*FundingRateInfo, error) {
	ctx := context.Background()

	premiumIndex, err := b.client.NewPremiumIndexService().
		Symbol(symbol).
		Do(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get funding rate: %v", err)
	}

	if len(premiumIndex) == 0 {
		return nil, fmt.Errorf("no funding rate data for symbol %s", symbol)
	}

	fundingRate, _ := strconv.ParseFloat(premiumIndex[0].LastFundingRate, 64)
	markPrice, _ := strconv.ParseFloat(premiumIndex[0].MarkPrice, 64)

	return &FundingRateInfo{
		Symbol:          symbol,
		FundingRate:     fundingRate,
		FundingTime:     premiumIndex[0].Time,
		NextFundingTime: premiumIndex[0].NextFundingTime,
		MarkPrice:       markPrice,
		IndexPrice:      markPrice, // Use mark price as index price
	}, nil
}

// GetFundingRateHistory - Get historical funding rates
func (b *Client) GetFundingRateHistory(symbol string, limit int, startTime, endTime int64) ([]*FundingRateHistory, error) {
	ctx := context.Background()

	service := b.client.NewFundingRateService().Symbol(symbol)

	if limit > 0 {
		service.Limit(limit)
	} else {
		service.Limit(100) // Default 100
	}

	if startTime > 0 {
		service.StartTime(startTime * 1000) // Convert to milliseconds
	}
	if endTime > 0 {
		service.EndTime(endTime * 1000)
	}

	rates, err := service.Do(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get funding rate history: %v", err)
	}

	result := []*FundingRateHistory{}
	for _, rate := range rates {
		fundingRate, _ := strconv.ParseFloat(rate.FundingRate, 64)
		result = append(result, &FundingRateHistory{
			Symbol:      rate.Symbol,
			FundingRate: fundingRate,
			FundingTime: rate.FundingTime,
		})
	}

	return result, nil
}

// CalculateFundingFee - Calculate expected funding fee
func (b *Client) CalculateFundingFee(symbol string, positionSize float64) (float64, error) {
	fundingInfo, err := b.GetFundingRate(symbol)
	if err != nil {
		return 0, err
	}

	// Funding fee = Position Value * Funding Rate
	// Position Value = Position Size * Mark Price
	positionValue := positionSize * fundingInfo.MarkPrice
	fundingFee := positionValue * fundingInfo.FundingRate

	return fundingFee, nil
}

// GetLiquidationRisk - Calculate liquidation risk for a position
func (b *Client) GetLiquidationRisk(symbol string) (*LiquidationRisk, error) {
	ctx := context.Background()

	// Get position information
	positions, err := b.client.NewGetPositionRiskService().
		Symbol(symbol).
		Do(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get position: %v", err)
	}

	if len(positions) == 0 {
		return nil, fmt.Errorf("no position found for %s", symbol)
	}

	pos := positions[0]
	posAmt, _ := strconv.ParseFloat(pos.PositionAmt, 64)

	if posAmt == 0 {
		return nil, fmt.Errorf("no open position for %s", symbol)
	}

	entryPrice, _ := strconv.ParseFloat(pos.EntryPrice, 64)
	markPrice, _ := strconv.ParseFloat(pos.MarkPrice, 64)
	liquidationPrice, _ := strconv.ParseFloat(pos.LiquidationPrice, 64)
	unrealizedPnL, _ := strconv.ParseFloat(pos.UnRealizedProfit, 64)
	leverage, _ := strconv.Atoi(pos.Leverage)

	// Calculate distance to liquidation (percentage)
	var distanceToLiquidation float64
	if liquidationPrice > 0 {
		if posAmt > 0 { // Long position
			distanceToLiquidation = ((markPrice - liquidationPrice) / markPrice) * 100
		} else { // Short position
			distanceToLiquidation = ((liquidationPrice - markPrice) / markPrice) * 100
		}
	}

	// Calculate margin ratio
	account, err := b.GetAccountInfo()
	var marginRatio float64
	if err == nil && account.TotalMarginBalance > 0 {
		marginRatio = (account.TotalMarginBalance + unrealizedPnL) / account.TotalPositionValue * 100
	}

	// Determine risk level
	riskLevel := "LOW"
	if distanceToLiquidation < 5 {
		riskLevel = "CRITICAL"
	} else if distanceToLiquidation < 10 {
		riskLevel = "HIGH"
	} else if distanceToLiquidation < 20 {
		riskLevel = "MEDIUM"
	}

	return &LiquidationRisk{
		Symbol:                symbol,
		PositionSize:          absFloat(posAmt),
		EntryPrice:            entryPrice,
		MarkPrice:             markPrice,
		LiquidationPrice:      liquidationPrice,
		MarginRatio:           marginRatio,
		UnrealizedPnL:         unrealizedPnL,
		Leverage:              leverage,
		DistanceToLiquidation: distanceToLiquidation,
		RiskLevel:             riskLevel,
	}, nil
}

// GetAccountSnapshot - Get daily account snapshot (Futures)
// This retrieves historical snapshots of your Futures account balance and positions
func (b *Client) GetAccountSnapshot(startTime, endTime int64, limit int) (*AccountSnapshotResponse, error) {
	// Get API credentials from environment
	apiKey := os.Getenv("BINANCE_API_KEY")
	secretKey := os.Getenv("BINANCE_SECRET_KEY")

	if apiKey == "" || secretKey == "" {
		return nil, fmt.Errorf("Binance API credentials not found")
	}

	// Determine base URL (testnet or production)
	baseURL := "https://api.binance.com"
	if os.Getenv("BINANCE_TESTNET") == "true" {
		// Note: Testnet uses different endpoint
		baseURL = "https://testnet.binance.vision"
	}

	// Build query parameters
	params := url.Values{}
	params.Set("type", "FUTURES")

	if limit <= 0 {
		limit = 7 // Default 7 days
	}
	if limit > 30 {
		limit = 30 // Max 30 days
	}
	params.Set("limit", strconv.Itoa(limit))

	if startTime > 0 {
		params.Set("startTime", strconv.FormatInt(startTime, 10))
	}
	if endTime > 0 {
		params.Set("endTime", strconv.FormatInt(endTime, 10))
	}

	// Add timestamp
	timestamp := time.Now().UnixMilli()
	params.Set("timestamp", strconv.FormatInt(timestamp, 10))

	// Create signature
	queryString := params.Encode()
	h := hmac.New(sha256.New, []byte(secretKey))
	h.Write([]byte(queryString))
	signature := hex.EncodeToString(h.Sum(nil))

	// Add signature to query
	params.Set("signature", signature)

	// Build full URL
	fullURL := fmt.Sprintf("%s/sapi/v1/accountSnapshot?%s", baseURL, params.Encode())

	// Create HTTP request
	req, err := http.NewRequest("GET", fullURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	// Add API key header
	req.Header.Set("X-MBX-APIKEY", apiKey)

	// Execute request
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %v", err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %v", err)
	}

	// Check for HTTP errors
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Binance API error (status %d): %s", resp.StatusCode, string(body))
	}

	// Parse response
	var result AccountSnapshotResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %v", err)
	}

	return &result, nil
}
