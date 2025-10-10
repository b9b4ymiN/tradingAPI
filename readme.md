# 🚀 Crypto Trading API Webhook System

High-performance automated cryptocurrency trading API built with Go, designed for Oracle Cloud free tier with Docker deployment.

## ✨ Features

- **High Performance**: Built with Go for maximum efficiency and low latency
- **Automated Trading**: Execute trades automatically via Binance Futures API
- **Real-time Data**: Store and sync trade data with Firebase Realtime Database
- **Smart Order Management**: Automatic Stop Loss and Take Profit orders
- **Leverage Support**: Configurable leverage (1x-125x)
- **Rate Limiting**: Built-in protection against API abuse
- **Docker Ready**: Optimized for containerized deployment
- **Health Monitoring**: Built-in health checks and monitoring

## 📋 System Requirements

- Oracle Cloud Free Tier (or any Linux server)
- Docker & Docker Compose installed
- Binance Futures account with API keys
- Firebase Realtime Database configured

## 🏗️ Architecture

```
User Request → API Gateway → Binance API → Execute Trade
                    ↓
              Firebase Database (Store Trade Data)
```

## Project Structure
crypto-trading-api/
├── cmd/
│   └── server/
│       └── main.go
├── internal/
│   ├── api/
│   │   ├── handler.go
│   │   ├── middleware.go
│   │   └── routes.go
│   ├── binance/
│   │   ├── client.go
│   │   └── trade.go
│   ├── firebase/
│   │   └── client.go
│   └── models/
│       └── trade.go
├── config/
│   └── config.go
├── Dockerfile
├── docker-compose.yml
├── go.mod
└── go.sum

## 📦 Installation

### 1. Clone the Repository

```bash
git clone <your-repo-url>
cd crypto-trading-api
```

### 2. Setup Environment Variables

```bash
cp .env.example .env
nano .env
```

Configure the following:

```bash
# API Security
API_KEY=your-strong-random-api-key

# Binance API (from Binance account)
BINANCE_API_KEY=your_binance_api_key
BINANCE_SECRET_KEY=your_binance_secret_key

# Firebase Database URL
FIREBASE_DATABASE_URL=https://your-project-id-default-rtdb.firebaseio.com
```

### 3. Setup Firebase Credentials

Download your Firebase service account key:
1. Go to Firebase Console → Project Settings → Service Accounts
2. Click "Generate New Private Key"
3. Save as `config/firebase-credentials.json`

### 4. Build and Deploy

```bash
# Build the Docker image
docker-compose build

# Start the service
docker-compose up -d

# Check logs
docker-compose logs -f crypto-api
```

## 🔧 API Documentation

### Base URL
```
http://your-server-ip:8080
```

### Authentication
All requests require API key in header:
```
X-API-Key: your-api-key
```
Or:
```
Authorization: Bearer your-api-key
```

### Endpoints

#### 1. Health Check
```bash
GET /health
```

**Response:**
```json
{
  "status": "healthy",
  "time": 1704067200
}
```

---

#### 2. Create Trade Order

```bash
POST /api/trade
```

**Headers:**
```
Content-Type: application/json
X-API-Key: your-api-key
```

**Request Body:**
```json
{
  "userId": "user123",
  "symbol": "BTCUSDT",
  "side": "BUY",
  "entryPrice": 45000.00,
  "stopLoss": 44000.00,
  "takeProfit": 47000.00,
  "leverage": 10,
  "size": 100.00
}
```

**Parameters:**
- `userId`: Unique user identifier
- `symbol`: Trading pair (e.g., BTCUSDT, ETHUSDT)
- `side`: "BUY" or "SELL"
- `entryPrice`: Target entry price (current market price)
- `stopLoss`: Stop loss price
- `takeProfit`: Take profit price
- `leverage`: Leverage multiplier (1-125)
- `size`: Position size in USDT

**Response (Success):**
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
    "entryPrice": 45000.00,
    "executedPrice": 45010.50,
    "stopLoss": 44000.00,
    "takeProfit": 47000.00,
    "leverage": 10,
    "size": 100.00,
    "status": "ACTIVE",
    "orderId": 123456789,
    "createdAt": 1704067200,
    "executedAt": 1704067205
  },
  "timestamp": 1704067205
}
```

---

#### 3. Get User Trades

```bash
GET /api/trades/:userId
```

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
      "status": "FILLED",
      ...
    }
  ],
  "timestamp": 1704067200
}
```

---

#### 4. Get Single Trade

```bash
GET /api/trade/:tradeId
```

**Response:**
```json
{
  "success": true,
  "message": "Trade fetched successfully",
  "data": {
    "id": "trade-id",
    "userId": "user123",
    ...
  },
  "timestamp": 1704067200
}
```

## 🧪 Testing

### Test with cURL

```bash
# Health check
curl http://localhost:8080/health

# Create trade
curl -X POST http://localhost:8080/api/trade \
  -H "Content-Type: application/json" \
  -H "X-API-Key: your-api-key" \
  -d '{
    "userId": "test-user",
    "symbol": "BTCUSDT",
    "side": "BUY",
    "entryPrice": 45000,
    "stopLoss": 44000,
    "takeProfit": 47000,
    "leverage": 10,
    "size": 100
  }'
```

## 🔒 Security Best Practices

1. **API Key**: Use strong random API keys
2. **Binance API**: Enable IP whitelist on Binance
3. **Firewall**: Configure Oracle Cloud security rules
4. **HTTPS**: Use Nginx with SSL certificate (Let's Encrypt)
5. **Rate Limiting**: Built-in (100 req/min per IP)
6. **Firebase Rules**: Secure database access rules

## 🚀 Production Deployment

### Oracle Cloud Setup

1. **Create Instance** (Free Tier)
   - Shape: VM.Standard.A1.Flex (ARM)
   - RAM: 6-24 GB
   - Storage: 50-200 GB

2. **Configure Firewall**
```bash
sudo firewall-cmd --permanent --add-port=8080/tcp
sudo firewall-cmd --reload
```

3. **Install Docker**
```bash
sudo yum update -y
sudo yum install -y docker
sudo systemctl start docker
sudo systemctl enable docker
sudo usermod -aG docker $USER
```

4. **Deploy Application**
```bash
git clone <your-repo>
cd crypto-trading-api
docker-compose up -d
```

## 📊 Monitoring

### Check Logs
```bash
docker-compose logs -f crypto-api
```

### Resource Usage
```bash
docker stats crypto-trading-api
```

### Container Health
```bash
docker ps
docker inspect crypto-trading-api | grep Health
```

## 🐛 Troubleshooting

### Container won't start
```bash
docker-compose logs crypto-api
docker-compose down
docker-compose up --build
```

### Firebase connection issues
- Check credentials file path
- Verify database URL
- Check Firebase console for errors

### Binance API errors
- Verify API keys
- Check IP whitelist
- Ensure futures trading enabled

## 📈 Performance

- **Response Time**: < 100ms average
- **Memory Usage**: ~50-100MB
- **CPU Usage**: < 5% idle, < 20% under load
- **Container Size**: ~15MB

## 📝 License

MIT License

## 🤝 Support

For issues and questions, please create an issue in the repository.

---

**⚠️ Disclaimer**: Cryptocurrency trading involves risk. This software is provided as-is without any warranty. Use at your own risk.
