# 📡 API Endpoints Documentation

## Overview

This document describes all available API endpoints for the Crypto Trading API.

**Base URL:** `http://localhost:8080`
**Authentication:** All `/api/*` endpoints require API key authentication via `X-API-Key` header

---

## 🔐 Authentication

All API endpoints (except `/health`) require authentication:

```bash
-H "X-API-Key: your-api-key-here"
```

---

## 📋 Core Endpoints

### 1. Health Check

**GET** `/health`

Check if the server is running.

**Authentication:** Not required

**Response:**
```json
{
  "status": "healthy",
  "time": 1696857600
}
```

**Example:**
```bash
curl http://localhost:8080/health
```

---

### 2. Place Trade

**POST** `/api/trade`

Place a new futures trade with automatic stop-loss and take-profit.

**Request Body:**
```json
{
  "userId": "user123",
  "symbol": "BTCUSDT",
  "side": "BUY",
  "entryPrice": 45000,
  "stopLoss": 44000,
  "takeProfit": 47000,
  "leverage": 10,
  "size": 100
}
```

**Response:**
```json
{
  "success": true,
  "tradeId": "550e8400-e29b-41d4-a716-446655440000",
  "message": "Trade executed successfully",
  "data": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "userId": "user123",
    "symbol": "BTCUSDT",
    "side": "BUY",
    "status": "ACTIVE",
    "orderId": 123456789,
    "executedPrice": 45010.5,
    "createdAt": 1696857600,
    "executedAt": 1696857601
  },
  "timestamp": 1696857601
}
```

**Example:**
```bash
curl -X POST http://localhost:8080/api/trade \
  -H "X-API-Key: your-api-key" \
  -H "Content-Type: application/json" \
  -d '{
    "userId": "user123",
    "symbol": "BTCUSDT",
    "side": "BUY",
    "entryPrice": 45000,
    "stopLoss": 44000,
    "takeProfit": 47000,
    "leverage": 10,
    "size": 100
  }'
```

---

### 3. Get User Trades

**GET** `/api/trades/:userId`

Get all trades for a specific user.

**Response:**
```json
{
  "success": true,
  "message": "Trades fetched successfully",
  "data": [
    {
      "id": "trade-id-1",
      "userId": "user123",
      "symbol": "BTCUSDT",
      "status": "ACTIVE",
      ...
    }
  ],
  "timestamp": 1696857600
}
```

**Example:**
```bash
curl http://localhost:8080/api/trades/user123 \
  -H "X-API-Key: your-api-key"
```

---

### 4. Get Single Trade

**GET** `/api/trade/:tradeId`

Get details of a specific trade.

**Response:**
```json
{
  "success": true,
  "message": "Trade fetched successfully",
  "data": {
    "id": "trade-id",
    "userId": "user123",
    "symbol": "BTCUSDT",
    ...
  },
  "timestamp": 1696857600
}
```

**Example:**
```bash
curl http://localhost:8080/api/trade/550e8400-e29b-41d4-a716-446655440000 \
  -H "X-API-Key: your-api-key"
```

---

## 🚀 Advanced Endpoints

### 5. System Status

**GET** `/api/status`

Get comprehensive system status including server, Binance, and Firebase connection status.

**Response:**
```json
{
  "success": true,
  "message": "System status retrieved successfully",
  "data": {
    "server": {
      "status": "online",
      "uptime": 3600,
      "timestamp": 1696857600,
      "version": "1.1.0"
    },
    "binance": {
      "status": "connected",
      "serverTime": 1696857600000,
      "canTrade": true,
      "canDeposit": true,
      "canWithdraw": false
    },
    "firebase": {
      "status": "connected",
      "activeTrades": 5
    }
  },
  "timestamp": 1696857600
}
```

**Example:**
```bash
curl http://localhost:8080/api/status \
  -H "X-API-Key: your-api-key"
```

---

### 6. Account Balance

**GET** `/api/balance`

Get detailed account balance information.

**Response:**
```json
{
  "success": true,
  "message": "Account balance retrieved successfully",
  "data": {
    "totalBalance": 10000.50,
    "availableBalance": 8500.25,
    "totalUnrealizedPnL": 250.75,
    "totalMarginBalance": 9500.00,
    "totalPositionValue": 1500.00,
    "assets": [
      {
        "asset": "USDT",
        "walletBalance": 10000.50,
        "unrealizedProfit": 250.75,
        "marginBalance": 9500.00,
        "availableBalance": 8500.25
      }
    ]
  },
  "timestamp": 1696857600
}
```

**Example:**
```bash
curl http://localhost:8080/api/balance \
  -H "X-API-Key: your-api-key"
```

---

### 7. Open Positions

**GET** `/api/positions`

Get all open positions with unrealized PnL.

**Response:**
```json
{
  "success": true,
  "message": "Open positions retrieved successfully",
  "data": {
    "totalPositions": 2,
    "totalPnL": 150.50,
    "positions": [
      {
        "symbol": "BTCUSDT",
        "side": "LONG",
        "positionAmt": 0.5,
        "entryPrice": 45000.0,
        "markPrice": 45500.0,
        "unrealizedProfit": 250.0,
        "leverage": 10,
        "liquidationPrice": 40500.0,
        "marginType": "isolated"
      }
    ]
  },
  "timestamp": 1696857600
}
```

**Example:**
```bash
curl http://localhost:8080/api/positions \
  -H "X-API-Key: your-api-key"
```

---

### 8. Pending Orders

**GET** `/api/orders?symbol=BTCUSDT`

Get all pending (open) orders. Optional symbol filter.

**Query Parameters:**
- `symbol` (optional): Filter by trading symbol

**Response:**
```json
{
  "success": true,
  "message": "Pending orders retrieved successfully",
  "data": {
    "totalOrders": 2,
    "orders": [
      {
        "orderId": 123456789,
        "symbol": "BTCUSDT",
        "side": "BUY",
        "type": "LIMIT",
        "price": "44000.0",
        "stopPrice": "0",
        "quantity": "0.5",
        "status": "NEW",
        "timeInForce": "GTC",
        "createdTime": 1696857600000,
        "reduceOnly": false,
        "closePosition": false
      }
    ]
  },
  "timestamp": 1696857600
}
```

**Examples:**
```bash
# Get all pending orders
curl http://localhost:8080/api/orders \
  -H "X-API-Key: your-api-key"

# Get pending orders for specific symbol
curl http://localhost:8080/api/orders?symbol=BTCUSDT \
  -H "X-API-Key: your-api-key"
```

---

### 9. Cancel Orders

**POST** `/api/orders/cancel`

Cancel pending orders. Can cancel specific order, all orders for a symbol, or all orders.

**Request Body (cancel specific order):**
```json
{
  "symbol": "BTCUSDT",
  "orderId": 123456789
}
```

**Request Body (cancel all for symbol):**
```json
{
  "symbol": "BTCUSDT"
}
```

**Request Body (cancel all orders):**
```json
{}
```

**Response:**
```json
{
  "success": true,
  "message": "Orders cancelled",
  "data": {
    "totalCancelled": 3,
    "results": [
      {
        "symbol": "BTCUSDT",
        "cancelledOrders": 2
      },
      {
        "symbol": "ETHUSDT",
        "cancelledOrders": 1
      }
    ]
  },
  "timestamp": 1696857600
}
```

**Examples:**
```bash
# Cancel specific order
curl -X POST http://localhost:8080/api/orders/cancel \
  -H "X-API-Key: your-api-key" \
  -H "Content-Type: application/json" \
  -d '{"symbol":"BTCUSDT","orderId":123456789}'

# Cancel all orders for BTCUSDT
curl -X POST http://localhost:8080/api/orders/cancel \
  -H "X-API-Key: your-api-key" \
  -H "Content-Type: application/json" \
  -d '{"symbol":"BTCUSDT"}'

# Cancel ALL orders
curl -X POST http://localhost:8080/api/orders/cancel \
  -H "X-API-Key: your-api-key" \
  -H "Content-Type: application/json" \
  -d '{}'
```

---

### 10. Close Position

**POST** `/api/position/close`

Close an open position immediately at market price.

**Request Body:**
```json
{
  "symbol": "BTCUSDT",
  "tradeId": "550e8400-e29b-41d4-a716-446655440000"
}
```

**Response:**
```json
{
  "success": true,
  "message": "Position closed successfully",
  "data": {
    "symbol": "BTCUSDT",
    "orderId": 987654321,
    "side": "SELL",
    "positionSide": "LONG",
    "quantity": "0.5",
    "price": "45500.0",
    "status": "FILLED",
    "realizedProfit": 250.0
  },
  "timestamp": 1696857600
}
```

**Example:**
```bash
curl -X POST http://localhost:8080/api/position/close \
  -H "X-API-Key: your-api-key" \
  -H "Content-Type: application/json" \
  -d '{
    "symbol": "BTCUSDT",
    "tradeId": "550e8400-e29b-41d4-a716-446655440000"
  }'
```

---

### 11. Trading Summary

**GET** `/api/summary?period=7d&userId=user123`

Get trading performance summary for a specific time period.

**Query Parameters:**
- `period` (optional): `1d`, `7d`, `1w`, `1m` (default: `1d`)
- `userId` (optional): Filter by specific user

**Response:**
```json
{
  "success": true,
  "message": "Trading summary retrieved successfully",
  "data": {
    "totalTrades": 25,
    "winningTrades": 15,
    "losingTrades": 8,
    "winRate": 60.0,
    "totalPnL": 1250.50,
    "totalVolume": 50000.0,
    "bestTrade": 500.0,
    "worstTrade": -150.0,
    "averagePnL": 50.02,
    "currentAccountPnL": 250.75,
    "symbolStats": {
      "BTCUSDT": 15,
      "ETHUSDT": 8,
      "BNBUSDT": 2
    }
  },
  "timestamp": 1696857600
}
```

**Examples:**
```bash
# Get 1-day summary for all users
curl http://localhost:8080/api/summary \
  -H "X-API-Key: your-api-key"

# Get 7-day summary
curl http://localhost:8080/api/summary?period=7d \
  -H "X-API-Key: your-api-key"

# Get 1-month summary for specific user
curl http://localhost:8080/api/summary?period=1m&userId=user123 \
  -H "X-API-Key: your-api-key"
```

---

## 📊 Summary Table

| Endpoint | Method | Auth | Description |
|----------|--------|------|-------------|
| `/health` | GET | ❌ | Health check |
| `/api/trade` | POST | ✅ | Place trade |
| `/api/trades/:userId` | GET | ✅ | Get user trades |
| `/api/trade/:tradeId` | GET | ✅ | Get single trade |
| `/api/status` | GET | ✅ | System status |
| `/api/balance` | GET | ✅ | Account balance |
| `/api/positions` | GET | ✅ | Open positions |
| `/api/orders` | GET | ✅ | Pending orders |
| `/api/orders/cancel` | POST | ✅ | Cancel orders |
| `/api/position/close` | POST | ✅ | Close position |
| `/api/summary` | GET | ✅ | Trading summary |

---

## 🔒 Error Responses

All endpoints return consistent error responses:

```json
{
  "success": false,
  "message": "Error description",
  "error": "Detailed error message",
  "timestamp": 1696857600
}
```

**Common HTTP Status Codes:**
- `200` - Success
- `400` - Bad Request (invalid parameters)
- `401` - Unauthorized (missing/invalid API key)
- `404` - Not Found
- `429` - Too Many Requests (rate limit exceeded)
- `500` - Internal Server Error

---

## ⚡ Rate Limiting

- **Limit:** 100 requests per minute per IP address
- **Response:** HTTP 429 when exceeded
- **Reset:** Every 60 seconds

---

## 💡 Tips

1. **Always check** `success` field in responses
2. **Store** `tradeId` for tracking trades
3. **Use** `/api/status` to verify system health before trading
4. **Monitor** `/api/positions` for active position PnL
5. **Review** `/api/summary` for performance analysis

---

**Version:** 1.1.0
**Last Updated:** 2025-10-09
