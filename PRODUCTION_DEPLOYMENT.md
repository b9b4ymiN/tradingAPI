# 🚀 Production Deployment Guide

**⚠️ WARNING: This guide is for deploying with REAL money!**

Read carefully and follow all security recommendations.

---

## 🎯 Production vs Testnet Differences

| Aspect | Testnet | Production |
|--------|---------|------------|
| **Money** | Fake (test funds) | REAL money 💰 |
| **API URL** | testnet.binancefuture.com | fapi.binance.com |
| **API Keys** | Testnet keys | Production keys |
| **Risk** | Zero | HIGH - real losses possible |
| **Testing** | Unlimited | Use small amounts first |
| **Data** | May reset | Permanent |

---

## ⚠️ Critical Production Requirements

### 1. **Start Small!**

**Recommended First Trade:**
- Leverage: 1x-2x (NOT 10x or higher!)
- Position Size: $10-20 USDT
- Stop Loss: 2-3% max loss
- Symbol: BTC or ETH (most liquid)

**Why?** Test with amounts you can afford to lose completely.

### 2. **Binance API Security**

**CRITICAL Setup:**

1. **Create NEW API Keys** (don't reuse testnet keys!)
   - Go to: https://www.binance.com/en/my/settings/api-management
   - Click "Create API"
   - Label: "Trading API Production"

2. **Enable ONLY Required Permissions:**
   - ✅ Enable Futures
   - ✅ Enable Spot & Margin Trading (if needed)
   - ❌ **DO NOT** enable "Enable Withdrawals"
   - ❌ **DO NOT** enable "Enable Internal Transfer"

3. **IP Whitelist (CRITICAL):**
   - Add ONLY your Oracle Cloud server IP
   - Format: `123.45.67.89/32`
   - **DO NOT** use `0.0.0.0/0` (allows anyone!)

4. **Additional Security:**
   - Enable 2FA on your Binance account
   - Set API key expiration (e.g., 90 days)
   - Set daily withdrawal limit to 0

### 3. **Firebase Production Database**

**Recommended:** Use separate Firebase project for production

1. Create new Firebase project: `your-project-prod`
2. Setup Realtime Database
3. Download NEW service account credentials
4. Configure security rules (see below)

**Security Rules for Production:**

```json
{
  "rules": {
    ".read": "auth != null",
    ".write": "auth != null",

    "trades": {
      "$tradeId": {
        ".validate": "newData.hasChildren(['userId', 'symbol', 'side', 'entryPrice'])",
        ".read": "auth != null",
        ".write": "auth != null"
      }
    },

    "users": {
      "$userId": {
        ".read": "auth != null",
        ".write": "auth != null"
      }
    }
  }
}
```

### 4. **New Production API Key**

**Generate NEW API key** (don't use testnet key):

```bash
# On server
openssl rand -base64 48
```

Save this securely - you'll need it for API calls.

---

## 📋 Pre-Production Checklist

### Security ✅

- [ ] New production API key generated
- [ ] Binance production API keys created
- [ ] Binance API IP whitelist configured (server IP only)
- [ ] Binance "Enable Withdrawals" is DISABLED
- [ ] Firebase production project created
- [ ] Firebase security rules configured
- [ ] SSL certificate installed (HTTPS)
- [ ] Server firewall configured
- [ ] Backup credentials stored securely (offline)

### Configuration ✅

- [ ] `.env.production` file created and configured
- [ ] `BINANCE_TESTNET=false` verified
- [ ] Firebase production credentials uploaded
- [ ] Docker compose configured for production
- [ ] Monitoring/logging enabled
- [ ] Alert notifications configured (optional)

### Testing ✅

- [ ] Test trade with $10-20 executed successfully
- [ ] Stop Loss triggered correctly
- [ ] Take Profit triggered correctly
- [ ] Position close works
- [ ] Order cancellation works
- [ ] Firebase logging works
- [ ] All API endpoints tested
- [ ] Error handling verified

### Infrastructure ✅

- [ ] Server has sufficient resources (CPU, RAM, disk)
- [ ] Automatic restart enabled (docker restart: unless-stopped)
- [ ] Log rotation configured
- [ ] Backup strategy implemented
- [ ] Monitoring dashboard setup (optional)
- [ ] Emergency stop procedure documented

---

## 🔧 Step-by-Step Production Deployment

### Step 1: Prepare Production Configuration

```bash
# On your local machine
cd /c/Programing/go/tradingAPI

# Create production .env from template
cp .env.production .env.production.local

# Edit with your production values
nano .env.production.local
```

**Configure these CRITICAL values:**

```bash
# Generate NEW production API key
API_KEY=$(openssl rand -base64 48)

# Set to production
BINANCE_TESTNET=false

# Your production Binance keys
BINANCE_API_KEY=your-production-key
BINANCE_SECRET_KEY=your-production-secret

# Production Firebase
FIREBASE_DATABASE_URL=https://your-prod-project.firebaseio.com
FIREBASE_CREDENTIALS_FILE=./config/firebase-credentials-production.json
```

### Step 2: Upload to Oracle Cloud

```bash
# Upload .env file
scp .env.production.local ubuntu@YOUR_VM_IP:~/tradingAPI/.env

# Upload production Firebase credentials
scp config/firebase-credentials-production.json \
    ubuntu@YOUR_VM_IP:~/tradingAPI/config/firebase-credentials.json

# Verify upload
ssh ubuntu@YOUR_VM_IP
cd ~/tradingAPI
cat .env | grep BINANCE_TESTNET
# Should output: BINANCE_TESTNET=false
```

### Step 3: Verify Binance Configuration

```bash
# On server
cd ~/tradingAPI

# Test Binance connection
curl https://fapi.binance.com/fapi/v1/ping
# Should return: {}

# Test your API key (replace with your key)
curl -H "X-MBX-APIKEY: YOUR_BINANCE_API_KEY" \
     https://fapi.binance.com/fapi/v2/account
# Should return account info (not error)
```

### Step 4: Deploy with Docker

```bash
# Stop any running containers
docker compose down

# Pull latest code (if using git)
git pull

# Rebuild with production config
docker compose up -d --build

# Check logs
docker compose logs -f
```

**Verify logs show:**
```
✅ Firebase client initialized successfully
🔧 Using Binance PRODUCTION  ← IMPORTANT!
✅ Binance client initialized successfully
🚀 Server starting on port 8080
```

### Step 5: Test Production API

**⚠️ IMPORTANT: Start with SMALL amounts!**

```bash
# Set production API key
export API_KEY="your-production-api-key"
export API_URL="http://YOUR_VM_IP:8080"

# Test health
curl $API_URL/health

# Test authentication
curl -H "X-API-Key: $API_KEY" $API_URL/api/status

# Test balance (should show REAL balance)
curl -H "X-API-Key: $API_KEY" $API_URL/api/balance

# Test positions
curl -H "X-API-Key: $API_KEY" $API_URL/api/positions
```

### Step 6: Execute Test Trade (Small Amount!)

```bash
# ⚠️ WARNING: This will use REAL money!
# Start with $10-20 USDT, 1x-2x leverage

curl -X POST $API_URL/api/trade \
  -H "X-API-Key: $API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "userId": "production_test",
    "symbol": "BTCUSDT",
    "side": "BUY",
    "entryPrice": 45000,
    "stopLoss": 44500,
    "takeProfit": 45500,
    "leverage": 1,
    "size": 10
  }'
```

**Monitor the trade:**

```bash
# Check position
curl -H "X-API-Key: $API_KEY" $API_URL/api/positions

# Check if SL/TP orders were placed
curl -H "X-API-Key: $API_KEY" $API_URL/api/orders

# Close position manually if needed
curl -X POST $API_URL/api/position/close \
  -H "X-API-Key: $API_KEY" \
  -d '{"symbol":"BTCUSDT"}'
```

### Step 7: Monitor and Verify

```bash
# Watch logs in real-time
docker compose logs -f

# Check container status
docker compose ps

# Check resource usage
docker stats

# View Firebase data
# Go to: Firebase Console → Database → Data
```

---

## 🛡️ Production Security Best Practices

### 1. SSL/HTTPS (Highly Recommended)

```bash
# Install Certbot
sudo apt install certbot -y

# Get SSL certificate (requires domain)
sudo certbot certonly --standalone -d yourdomain.com

# Certificates will be at:
# /etc/letsencrypt/live/yourdomain.com/fullchain.pem
# /etc/letsencrypt/live/yourdomain.com/privkey.pem
```

**Configure Nginx for HTTPS** (create `nginx/nginx.conf`):

```nginx
events {
    worker_connections 1024;
}

http {
    upstream api {
        server crypto-api:8080;
    }

    # Redirect HTTP to HTTPS
    server {
        listen 80;
        server_name yourdomain.com;
        return 301 https://$server_name$request_uri;
    }

    # HTTPS server
    server {
        listen 443 ssl;
        server_name yourdomain.com;

        ssl_certificate /etc/letsencrypt/live/yourdomain.com/fullchain.pem;
        ssl_certificate_key /etc/letsencrypt/live/yourdomain.com/privkey.pem;

        location / {
            proxy_pass http://api;
            proxy_set_header Host $host;
            proxy_set_header X-Real-IP $remote_addr;
            proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
            proxy_set_header X-Forwarded-Proto $scheme;
        }
    }
}
```

### 2. Firewall Configuration

```bash
# Allow only necessary ports
sudo ufw default deny incoming
sudo ufw default allow outgoing
sudo ufw allow 22/tcp    # SSH
sudo ufw allow 80/tcp    # HTTP
sudo ufw allow 443/tcp   # HTTPS
sudo ufw enable

# Remove direct access to port 8080 (use Nginx instead)
# Do this AFTER Nginx is working
```

### 3. Fail2Ban (Brute Force Protection)

```bash
# Install Fail2Ban
sudo apt install fail2ban -y

# Configure for SSH
sudo cp /etc/fail2ban/jail.conf /etc/fail2ban/jail.local
sudo systemctl enable fail2ban
sudo systemctl start fail2ban
```

### 4. Monitoring and Alerts

**Create monitoring script** (`script/production_monitor.sh`):

```bash
#!/bin/bash

API_URL="http://localhost:8080"
API_KEY="your-api-key"

while true; do
    # Check health
    health=$(curl -s $API_URL/health)

    if ! echo "$health" | grep -q "healthy"; then
        echo "⚠️  ALERT: API is down!"
        # Send alert (email, Slack, Telegram)
    fi

    # Check balance
    balance=$(curl -s -H "X-API-Key: $API_KEY" $API_URL/api/balance)
    available=$(echo $balance | grep -o '"availableBalance":[0-9.]*' | cut -d: -f2)

    # Alert if balance is low
    if (( $(echo "$available < 50" | bc -l) )); then
        echo "⚠️  ALERT: Low balance: $available USDT"
    fi

    # Check for positions at risk
    positions=$(curl -s -H "X-API-Key: $API_KEY" $API_URL/api/positions)

    sleep 60  # Check every minute
done
```

### 5. Backup Strategy

```bash
# Backup script (run daily via cron)
#!/bin/bash

DATE=$(date +%Y%m%d)
BACKUP_DIR="/home/ubuntu/backups"

# Backup configuration
tar -czf $BACKUP_DIR/config-$DATE.tar.gz \
    ~/tradingAPI/.env \
    ~/tradingAPI/config/

# Backup Firebase data (export via Firebase Console)
# Or use Firebase Admin SDK to export

# Keep last 7 days only
find $BACKUP_DIR -name "config-*.tar.gz" -mtime +7 -delete
```

**Setup cron:**

```bash
crontab -e

# Add line:
0 2 * * * /home/ubuntu/backup.sh
```

---

## 📊 Risk Management Recommendations

### 1. Trading Limits

**Implement in your trading bot:**

```python
MAX_LEVERAGE = 5  # Don't exceed 5x leverage
MAX_POSITION_SIZE = 500  # Max $500 per trade
MAX_DAILY_LOSS = 100  # Stop trading if lose $100/day
MAX_CONCURRENT_POSITIONS = 3  # Max 3 positions at once
MIN_BALANCE = 200  # Stop if balance drops below $200
```

### 2. Gradual Rollout

**Week 1:**
- Leverage: 1x-2x only
- Position size: $10-50
- Max 2-3 trades per day
- Manual monitoring

**Week 2-4:**
- Leverage: up to 3x
- Position size: $50-200
- Increase frequency
- Review results daily

**Month 2+:**
- Leverage: up to 5x
- Position size: based on account size (max 5% per trade)
- Fully automated
- Weekly performance review

### 3. Stop Loss Rules

**Always use:**
- Minimum: 1% stop loss
- Maximum: 5% stop loss
- Trailing stop for profitable trades
- Emergency stop if daily loss limit hit

---

## 🚨 Emergency Procedures

### Emergency Stop

**If something goes wrong:**

```bash
# 1. Stop the server immediately
ssh ubuntu@YOUR_VM_IP
docker compose down

# 2. Close all positions manually via Binance
# Go to: https://www.binance.com/en/futures/BTCUSDT
# Click "Close All Positions"

# 3. Disable API keys
# Go to: https://www.binance.com/en/my/settings/api-management
# Click "Delete" or "Disable"

# 4. Check Firebase for trade logs
# Analyze what went wrong

# 5. Fix issue before restarting
```

### Recovery Procedure

```bash
# 1. Review logs
docker compose logs > error_logs.txt

# 2. Check Firebase
# Export all trades for analysis

# 3. Calculate losses
# Review all executed trades

# 4. Fix root cause
# Update code or configuration

# 5. Test on testnet first
# Before going back to production
```

---

## 📈 Performance Monitoring

### Key Metrics to Track

| Metric | Target | Alert If |
|--------|--------|----------|
| **Win Rate** | > 50% | < 40% |
| **Average PnL** | Positive | Negative 3 days in a row |
| **Max Drawdown** | < 20% | > 30% |
| **Daily Loss** | < $100 | > $100 |
| **API Uptime** | > 99% | < 95% |
| **Response Time** | < 500ms | > 1000ms |

### Monitoring Commands

```bash
# Check trading summary
curl -H "X-API-Key: $API_KEY" \
     "$API_URL/api/summary?period=7"

# Check recent trades
curl -H "X-API-Key: $API_KEY" \
     "$API_URL/api/trades/your_user_id"

# Check system status
curl -H "X-API-Key: $API_KEY" \
     "$API_URL/api/status"
```

---

## ✅ Production Deployment Checklist

Print this and check off each item:

### Pre-Deployment

- [ ] Read this entire document
- [ ] Understand the risks
- [ ] Have emergency funds available
- [ ] Tested thoroughly on testnet (100+ trades)
- [ ] Reviewed all code
- [ ] Created production API keys
- [ ] Configured IP whitelist
- [ ] Disabled withdrawals on Binance API
- [ ] Generated new production API_KEY
- [ ] Setup production Firebase
- [ ] Configured Firebase security rules
- [ ] Uploaded all credentials securely
- [ ] Installed SSL certificate (optional but recommended)
- [ ] Configured firewall
- [ ] Setup monitoring/alerts
- [ ] Documented emergency procedures
- [ ] Informed anyone who needs to know

### First Deployment

- [ ] Deployed to production server
- [ ] Verified BINANCE_TESTNET=false
- [ ] Verified using production API keys
- [ ] Tested with $10-20 test trade
- [ ] Verified SL triggers correctly
- [ ] Verified TP triggers correctly
- [ ] Monitored for 24 hours
- [ ] Reviewed all logs
- [ ] No errors in Firebase
- [ ] Backup working correctly

### Ongoing

- [ ] Daily: Check trading summary
- [ ] Daily: Review any errors
- [ ] Weekly: Analyze performance
- [ ] Weekly: Update security
- [ ] Monthly: Review and optimize strategy
- [ ] Monthly: Backup configuration
- [ ] As needed: Adjust risk parameters

---

## 🆘 Support and Resources

### If Something Goes Wrong

1. **Stop trading immediately** (docker compose down)
2. **Close all positions** on Binance manually
3. **Review logs** and Firebase data
4. **Calculate actual losses**
5. **Fix the issue**
6. **Test on testnet again**
7. **Gradually resume production**

### Useful Links

- **Binance API Docs:** https://binance-docs.github.io/apidocs/futures/en/
- **Firebase Console:** https://console.firebase.google.com/
- **Oracle Cloud Console:** https://cloud.oracle.com/
- **Your Server Logs:** `docker compose logs`

---

## ⚠️ Final Warning

**PRODUCTION TRADING INVOLVES REAL FINANCIAL RISK**

- You can lose money
- Start small and scale gradually
- Never trade more than you can afford to lose
- Always use stop losses
- Monitor regularly
- Have an emergency plan

**This software is provided AS IS. Use at your own risk.**

---

**Ready for Production:** Only after completing ALL checklist items ✅

**Recommended Starting Capital:** $500-1000 minimum
**Recommended First Trade:** $10-20, 1x leverage
**Recommended Max Risk:** 2-5% per trade

🚀 **Trade responsibly and good luck!**
