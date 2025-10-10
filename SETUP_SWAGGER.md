# 📄 Swagger Documentation Setup Guide

## 🎯 What is Swagger?

Swagger (OpenAPI) provides:
- **Interactive API Documentation** - Test endpoints directly in your browser
- **Auto-generated Docs** - Always up-to-date with your code
- **API Client Generation** - Auto-create client libraries
- **API Testing** - Test all endpoints without Postman/curl

---

## 🚀 Quick Setup

### Step 1: Install Swagger CLI Tool

```bash
# On your local Windows machine (PowerShell as Admin)
go install github.com/swaggo/swag/cmd/swag@latest

# Add Go bin to PATH if not already
# Add this to your system PATH:
# C:\Users\YourUsername\go\bin

# Verify installation
swag --version
```

### Step 2: Download Dependencies

```bash
cd /c/Programing/go/tradingAPI

# Download new dependencies
go mod download
go mod tidy
```

### Step 3: Generate Swagger Docs

```bash
# Generate Swagger documentation
swag init -g cmd/server/main.go --output docs

# This creates:
# - docs/docs.go
# - docs/swagger.json
# - docs/swagger.yaml
```

### Step 4: Build and Run

```bash
# Build
go build -o bin/server.exe ./cmd/server

# Run
./bin/server.exe

# Or with Docker
docker compose up --build
```

### Step 5: Access Swagger UI

Open your browser:
```
http://localhost:8080/swagger/index.html
```

🎉 **You should see interactive API documentation!**

---

## 📚 What's Included

I've added Swagger annotations to document:

### ✅ All 11 API Endpoints:

1. **GET /health** - Health check
2. **POST /api/trade** - Place new trade
3. **GET /api/trades/:userId** - Get user trades
4. **GET /api/trade/:tradeId** - Get specific trade
5. **GET /api/status** - System status
6. **GET /api/balance** - Account balance
7. **GET /api/positions** - Open positions
8. **GET /api/orders** - Pending orders
9. **POST /api/orders/cancel** - Cancel orders
10. **POST /api/position/close** - Close position
11. **GET /api/summary** - Trading summary

### 📋 Features:

- ✅ Request/Response examples
- ✅ Parameter validation
- ✅ Error responses
- ✅ Authentication (X-API-Key)
- ✅ Try it out functionality
- ✅ Model schemas
- ✅ Tags and grouping

---

## 🔧 Files Modified/Created

### Modified:
1. **go.mod** - Added Swagger dependencies
2. **cmd/server/main.go** - Added Swagger metadata
3. **internal/api/routes.go** - Added Swagger route
4. **internal/api/handler.go** - Added Swagger annotations
5. **internal/api/advanced_handlers.go** - Added Swagger annotations

### Created:
6. **internal/api/swagger_handlers.go** - Swagger helper functions
7. **docs/** directory - Auto-generated (after running `swag init`)

---

## 🎨 Swagger UI Features

### Interactive Testing:

1. **Click "Authorize"** button
2. **Enter your API_KEY**
3. **Click any endpoint**
4. **Click "Try it out"**
5. **Fill in parameters**
6. **Click "Execute"**
7. **See live response!**

### Features:

- ✅ **Models** - View all data structures
- ✅ **Schemas** - See request/response formats
- ✅ **Try it out** - Test directly in browser
- ✅ **Download** - Get OpenAPI spec (JSON/YAML)
- ✅ **Code Gen** - Generate client code

---

## 📖 Swagger Annotations Guide

### Basic Endpoint:

```go
// GetBalance godoc
// @Summary      Get account balance
// @Description  Retrieve current Binance Futures account balance
// @Tags         Account
// @Security     ApiKeyAuth
// @Produce      json
// @Success      200  {object}  BalanceResponse
// @Failure      401  {object}  ErrorResponse
// @Failure      500  {object}  ErrorResponse
// @Router       /api/balance [get]
func GetBalance(c *gin.Context) {
    // Handler code...
}
```

### Endpoint with Parameters:

```go
// PlaceTrade godoc
// @Summary      Place new trade
// @Description  Execute a Futures trade with automatic SL/TP
// @Tags         Trading
// @Security     ApiKeyAuth
// @Accept       json
// @Produce      json
// @Param        trade  body      TradeRequest  true  "Trade parameters"
// @Success      200    {object}  TradeResponse
// @Failure      400    {object}  ErrorResponse
// @Failure      500    {object}  ErrorResponse
// @Router       /api/trade [post]
func PlaceTrade(c *gin.Context) {
    // Handler code...
}
```

---

## 🌐 Access Swagger

### Local Development:
```
http://localhost:8080/swagger/index.html
```

### Production Server:
```
http://YOUR_SERVER_IP:8080/swagger/index.html
```

### With Domain:
```
https://yourdomain.com/swagger/index.html
```

---

## 🔐 Security in Swagger

### Using API Key:

1. Open Swagger UI
2. Click **"Authorize"** button (🔒 icon)
3. Enter your **API_KEY**
4. Click **"Authorize"**
5. Click **"Close"**

Now all protected endpoints will include your API key!

---

## 📥 Export API Specification

### JSON Format:
```
http://localhost:8080/swagger/doc.json
```

### YAML Format:
```
# After generating docs
cat docs/swagger.yaml
```

### Use Cases:
- Import into Postman
- Generate client libraries
- Share with frontend team
- API documentation portal

---

## 🛠️ Updating Documentation

**After changing any handler annotations:**

```bash
# Regenerate docs
swag init -g cmd/server/main.go --output docs

# Restart server
./bin/server.exe

# Or with Docker
docker compose up --build
```

**The docs update automatically!**

---

## 📱 Generate API Clients

### JavaScript/TypeScript:
```bash
swagger-codegen generate -i docs/swagger.json -l typescript-axios -o client/typescript
```

### Python:
```bash
swagger-codegen generate -i docs/swagger.json -l python -o client/python
```

### Go:
```bash
swagger-codegen generate -i docs/swagger.json -l go -o client/go
```

---

## 🎯 Example Swagger Screenshots

### Home Page:
Shows all endpoints grouped by tags:
- Health
- Trading
- Account
- Positions
- Analytics

### Endpoint Detail:
- Parameters
- Request body schema
- Response codes
- Example values
- Try it out button

### Authorization:
- API key input
- Bearer token support
- OAuth2 (if configured)

---

## ✅ Verification Checklist

After setup, verify:

- [ ] `swag --version` works
- [ ] `go mod tidy` completes successfully
- [ ] `swag init` generates docs/ folder
- [ ] Server starts without errors
- [ ] Can access http://localhost:8080/swagger/index.html
- [ ] See all 11 endpoints listed
- [ ] Can click "Authorize" and enter API key
- [ ] Can "Try it out" on /health endpoint
- [ ] Response shows correctly

---

## 🆘 Troubleshooting

### Issue: `swag: command not found`

**Solution:**
```bash
# Install swag
go install github.com/swaggo/swag/cmd/swag@latest

# Add to PATH (Windows)
# Add: C:\Users\YourUsername\go\bin
```

### Issue: Cannot find package "crypto-trading-api/docs"

**Solution:**
```bash
# Generate docs first
swag init -g cmd/server/main.go --output docs

# Then build
go build ./cmd/server
```

### Issue: Swagger UI shows no endpoints

**Solution:**
```bash
# Regenerate with verbose output
swag init -g cmd/server/main.go --output docs --parseDependency

# Check docs/swagger.json exists
cat docs/swagger.json
```

### Issue: 404 on /swagger/index.html

**Solution:**
Check that routes.go has:
```go
router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
```

---

## 🎓 Learn More

### Official Documentation:
- **Swaggo:** https://github.com/swaggo/swag
- **OpenAPI:** https://swagger.io/specification/

### Annotation Reference:
- **General Info:** https://github.com/swaggo/swag#general-api-info
- **API Operations:** https://github.com/swaggo/swag#api-operation
- **Security:** https://github.com/swaggo/swag#security

---

## 🚀 Production Deployment

### Build with Swagger:

```dockerfile
# Already included in Dockerfile
# Swagger docs are embedded in binary
```

### Disable in Production (Optional):

```go
// In routes.go
if os.Getenv("GIN_MODE") != "release" {
    router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
}
```

### Or protect with authentication:

```go
// Only allow from localhost
swaggerGroup := router.Group("/swagger")
swaggerGroup.Use(func(c *gin.Context) {
    if c.ClientIP() != "127.0.0.1" {
        c.AbortWithStatus(403)
        return
    }
})
swaggerGroup.GET("/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
```

---

## ✅ Summary

**What you get:**

✅ Professional API documentation
✅ Interactive testing interface
✅ Always up-to-date with code
✅ Exportable OpenAPI spec
✅ Client code generation
✅ Better developer experience

**Setup time:** 10-15 minutes
**Maintenance:** Automatic!

**Next steps:**
1. Install swag CLI
2. Run `swag init`
3. Access Swagger UI
4. Enjoy! 🎉

---

**Created:** 2025-10-10
**For:** Crypto Trading API v1.0
