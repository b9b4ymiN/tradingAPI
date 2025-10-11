# Binance Futures Trading API

A production-ready RESTful API service for automated cryptocurrency futures trading on Binance, built with Go for high performance and reliability.

**Version:** 1.2.0
**Status:** Production Ready
**API Documentation:** http://localhost:8080/swagger/index.html

---

## Overview

This API service provides a comprehensive solution for executing and managing Binance Futures trades programmatically. It features real-time order execution, automated risk management through stop-loss and take-profit orders, historical account tracking, and complete position management capabilities.

### Key Features

- **Order Execution**: Market and limit orders with configurable leverage (1x-125x)
- **Risk Management**: Automatic stop-loss and take-profit order placement
- **Account Monitoring**: Real-time balance tracking and historical account snapshots
- **Exchange Information**: Query trading requirements and symbol specifications
- **Position Management**: Open, monitor, and close positions programmatically
- **Data Persistence**: Firebase integration for trade history and analytics
- **API Documentation**: Interactive Swagger/OpenAPI specification
- **Production Ready**: Containerized deployment with Docker

---

## Technical Stack

- **Language**: Go 1.21+
- **Exchange**: Binance Futures API
- **Database**: Firebase Realtime Database
- **Documentation**: Swagger/OpenAPI 3.0
- **Deployment**: Docker & Docker Compose
- **Architecture**: RESTful API with middleware-based authentication

---

## API Endpoints

| Endpoint | Method | Description | Authentication |
|----------|--------|-------------|----------------|
| `/health` | GET | Service health check | No |
| `/api/balance` | GET | Retrieve account balance | Required |
| `/api/positions` | GET | List open positions | Required |
| `/api/orders` | GET | List pending orders | Required |
| `/api/trade` | POST | Execute trade order | Required |
| `/api/position/close` | POST | Close open position | Required |
| `/api/orders/cancel` | POST | Cancel pending orders | Required |
| `/api/exchange/info` | GET | Query symbol requirements | Required |
| `/api/account/snapshot` | GET | Historical account data | Required |
| `/api/summary` | GET | Trading statistics | Required |

Complete API documentation available at: `/swagger/index.html`

---

## Installation

### Prerequisites

- Go 1.21 or higher
- Docker and Docker Compose
- Binance Futures account with API credentials
- Firebase project with Realtime Database

### Configuration

1. **Clone the repository**
   ```bash
   git clone <repository-url>
   cd tradingAPI
   ```

2. **Configure environment variables**
   ```bash
   cp .env.example .env
   ```

   Edit `.env` with your credentials:
   ```env
   # API Authentication
   API_KEY=<your-secure-api-key>

   # Binance Configuration
   BINANCE_TESTNET=false
   BINANCE_API_KEY=<your-binance-api-key>
   BINANCE_SECRET_KEY=<your-binance-secret-key>

   # Firebase Configuration
   FIREBASE_DATABASE_URL=https://<project-id>.firebaseio.com
   FIREBASE_CREDENTIALS_FILE=./config/firebase-credentials.json
   ```

3. **Setup Firebase credentials**
   - Download service account key from Firebase Console
   - Place file at `./config/firebase-credentials.json`

4. **Deploy the service**
   ```bash
   docker-compose up -d --build
   ```

5. **Verify deployment**
   ```bash
   docker-compose logs -f crypto-api
   ```

---

## Usage

### Authentication

All API requests (except `/health`) require authentication via API key in the request header:

```http
X-API-Key: <your-api-key>
```

### Execute Market Order

```bash
curl -X POST http://localhost:8080/api/trade \
  -H "X-API-Key: <your-api-key>" \
  -H "Content-Type: application/json" \
  -d '{
    "userId": "user_identifier",
    "symbol": "BTCUSDT",
    "side": "BUY",
    "entryPrice": 50000.00,
    "stopLoss": 49000.00,
    "takeProfit": 52000.00,
    "leverage": 10,
    "size": 100.00,
    "orderType": "MARKET"
  }'
```

### Execute Limit Order

```bash
curl -X POST http://localhost:8080/api/trade \
  -H "X-API-Key: <your-api-key>" \
  -H "Content-Type: application/json" \
  -d '{
    "userId": "user_identifier",
    "symbol": "BTCUSDT",
    "side": "BUY",
    "entryPrice": 48000.00,
    "stopLoss": 47000.00,
    "takeProfit": 51000.00,
    "leverage": 10,
    "size": 100.00,
    "orderType": "LIMIT"
  }'
```

### Query Symbol Requirements

```bash
curl -X GET "http://localhost:8080/api/exchange/info?symbol=BTCUSDT" \
  -H "X-API-Key: <your-api-key>"
```

**Response:**
```json
{
  "success": true,
  "data": {
    "symbols": [{
      "symbol": "BTCUSDT",
      "minNotional": "100",
      "minQuantity": "0.001",
      "quantityPrecision": 3,
      "pricePrecision": 2
    }]
  }
}
```

### Retrieve Account Snapshot

```bash
curl -X GET "http://localhost:8080/api/account/snapshot?limit=7" \
  -H "X-API-Key: <your-api-key>"
```

---

## Binance API Configuration

### Required Permissions

1. Navigate to Binance → API Management
2. Create new API key with the following settings:
   - **Enable**: "Enable Futures", "Enable Reading"
   - **Disable**: "Enable Withdrawals" (recommended for security)
3. Configure IP whitelist with your server's IP address
4. Enable two-factor authentication for API key management

### Security Recommendations

- Never enable withdrawal permissions on API keys used for automated trading
- Use IP whitelist restrictions
- Rotate API keys periodically
- Monitor API usage through Binance dashboard
- Implement rate limiting in your application

---

## Trading Requirements

### Minimum Position Sizes (Notional)

| Symbol | Minimum Position | Typical Use Case |
|--------|-----------------|------------------|
| BTCUSDT | $100 | Bitcoin trading |
| ETHUSDT | $100 | Ethereum trading |
| XRPUSDT | $5 | Low-cost testing |
| ADAUSDT | $5 | Altcoin trading |
| BNBUSDT | $5 | Binance Coin |

Query requirements for any symbol: `GET /api/exchange/info?symbol=<SYMBOL>`

---

## Architecture

```
tradingAPI/
├── cmd/
│   └── server/
│       └── main.go                 # Application entry point
├── internal/
│   ├── api/
│   │   ├── handler.go             # Core trade handlers
│   │   ├── advanced_handlers.go   # Extended functionality
│   │   ├── middleware.go          # Authentication & rate limiting
│   │   └── routes.go              # Route configuration
│   ├── binance/
│   │   ├── binance_client.go      # Binance API integration
│   │   └── binance_advanced_funcs.go
│   ├── firebase/
│   │   └── client.go              # Firebase integration
│   └── models/
│       └── trade.go               # Data models
├── docs/                          # Swagger documentation
├── Dockerfile                     # Container configuration
├── docker-compose.yml             # Service orchestration
└── .env                           # Environment configuration
```

---

## Development

### Local Development

```bash
go run cmd/server/main.go
```

### Generate Swagger Documentation

```bash
swag init -g cmd/server/main.go -o docs
```

### Run Tests

```bash
go test ./...
```

### Build Binary

```bash
go build -o bin/server cmd/server/main.go
```

---

## Troubleshooting

### Precision Errors

**Issue**: "Precision is over the maximum defined"
**Solution**: Fixed in version 1.1.0. Rebuild the application.

### Position Size Errors

**Issue**: "Position size too small"
**Solution**: Query minimum requirements via `/api/exchange/info?symbol=<SYMBOL>`
- BTCUSDT requires minimum $100 position
- XRPUSDT requires minimum $5 position

### Connection Issues

**Issue**: Service not responding
**Solution**:
```bash
docker-compose ps                    # Check service status
docker-compose logs crypto-api       # View logs
docker-compose restart               # Restart service
```

### Swagger Documentation Not Updated

**Solution**:
```bash
swag init -g cmd/server/main.go -o docs
docker-compose up -d --build
```

---

## Production Deployment

### Pre-deployment Checklist

- [ ] Configure production Binance API credentials
- [ ] Set `BINANCE_TESTNET=false` in environment
- [ ] Add production server IP to Binance API whitelist
- [ ] Upload Firebase service account credentials
- [ ] Generate secure API authentication key
- [ ] Configure firewall rules
- [ ] Setup HTTPS reverse proxy (recommended: Nginx)
- [ ] Implement monitoring and alerting
- [ ] Test with minimal position sizes
- [ ] Review and configure rate limits

### Monitoring

```bash
# View real-time logs
docker-compose logs -f crypto-api

# Check service health
curl http://localhost:8080/health

# Verify environment configuration
docker-compose exec crypto-api env | grep BINANCE
```

---

## Version History

| Version | Date | Changes |
|---------|------|---------|
| 1.2.0 | 2025-01-11 | Added account snapshot API, MARKET/LIMIT order types |
| 1.1.0 | 2025-01-10 | Added exchange info API, fixed quantity precision |
| 1.0.0 | 2025-01-08 | Initial production release |

---

## Security Considerations

### Application Security

- Environment variables not committed to version control
- API key authentication required for all trading endpoints
- Rate limiting implemented on all routes
- Input validation on all user-supplied data

### Trading Security

- Withdrawal permissions should be disabled on Binance API keys
- IP whitelist restrictions recommended
- Two-factor authentication required for API key management
- Start with minimal position sizes for testing
- Implement position size limits based on account balance

---

## Disclaimer

This software is provided for educational and research purposes. Cryptocurrency trading carries substantial risk of loss. Users are solely responsible for:

- All trading decisions and outcomes
- Compliance with local regulations
- Security of API credentials
- Proper risk management
- Testing in non-production environments before live deployment

The authors and contributors assume no liability for financial losses incurred through use of this software.

---

## License

MIT License - See LICENSE file for details

---

## Support

**Documentation**: http://localhost:8080/swagger/index.html
**Issues**: Submit via repository issue tracker
**API Reference**: Binance Futures API Documentation

---

**Built with Go, Docker, Binance Futures API, and Firebase**
