#!/bin/bash

# API Testing Examples for Crypto Trading API
# Make sure to set your API_KEY environment variable first
# export API_KEY="your-api-key-here"

API_URL="http://localhost:8080"
API_KEY="${API_KEY:-your-api-key-here}"

echo "🧪 Crypto Trading API - Testing Suite"
echo "======================================"

# 1. Health Check
echo ""
echo "1️⃣ Health Check"
echo "----------------"
curl -X GET "${API_URL}/health" \
  -H "Content-Type: application/json" \
  | jq '.'

# 2. Create BUY Trade
echo ""
echo "2️⃣ Create BUY Trade (Long Position)"
echo "------------------------------------"
curl -X POST "${API_URL}/api/trade" \
  -H "Content-Type: application/json" \
  -H "X-API-Key: ${API_KEY}" \
  -d '{
    "userId": "user123",
    "symbol": "BTCUSDT",
    "side": "BUY",
    "entryPrice": 45000.00,
    "stopLoss": 44000.00,
    "takeProfit": 47000.00,
    "leverage": 10,
    "size": 100.00
  }' | jq '.'

# 3. Create SELL Trade
echo ""
echo "3️⃣ Create SELL Trade (Short Position)"
echo "--------------------------------------"
curl -X POST "${API_URL}/api/trade" \
  -H "Content-Type: application/json" \
  -H "X-API-Key: ${API_KEY}" \
  -d '{
    "userId": "user123",
    "symbol": "ETHUSDT",
    "side": "SELL",
    "entryPrice": 2500.00,
    "stopLoss": 2550.00,
    "takeProfit": 2400.00,
    "leverage": 5,
    "size": 50.00
  }' | jq '.'

# 4. Get User Trades
echo ""
echo "4️⃣ Get All Trades for User"
echo "----------------------------"
curl -X GET "${API_URL}/api/trades/user123" \
  -H "Content-Type: application/json" \
  -H "X-API-Key: ${API_KEY}" \
  | jq '.'

# 5. Get Single Trade (replace with actual trade ID)
echo ""
echo "5️⃣ Get Single Trade"
echo "--------------------"
TRADE_ID="550e8400-e29b-41d4-a716-446655440000"
curl -X GET "${API_URL}/api/trade/${TRADE_ID}" \
  -H "Content-Type: application/json" \
  -H "X-API-Key: ${API_KEY}" \
  | jq '.'

# 6. Test Invalid API Key
echo ""
echo "6️⃣ Test Invalid API Key (Should Fail)"
echo "---------------------------------------"
curl -X POST "${API_URL}/api/trade" \
  -H "Content-Type: application/json" \
  -H "X-API-Key: invalid-key" \
  -d '{
    "userId": "user123",
    "symbol": "BTCUSDT",
    "side": "BUY",
    "entryPrice": 45000.00,
    "stopLoss": 44000.00,
    "takeProfit": 47000.00,
    "leverage": 10,
    "size": 100.00
  }' | jq '.'

# 7. Test Invalid Trade Parameters
echo ""
echo "7️⃣ Test Invalid Parameters (Should Fail)"
echo "------------------------------------------"
curl -X POST "${API_URL}/api/trade" \
  -H "Content-Type: application/json" \
  -H "X-API-Key: ${API_KEY}" \
  -d '{
    "userId": "user123",
    "symbol": "BTCUSDT",
    "side": "BUY",
    "entryPrice": 45000.00,
    "stopLoss": 46000.00,
    "takeProfit": 47000.00,
    "leverage": 10,
    "size": 100.00
  }' | jq '.'

echo ""
echo "✅ Testing completed!"
