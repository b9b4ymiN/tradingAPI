package binance

import (
	"context"
	"fmt"
	"strconv"

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
