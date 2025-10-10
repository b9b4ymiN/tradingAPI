# 🔥 Firebase Realtime Database Setup Guide

## ✅ Issue Fixed: Regional Database URL Support

Your error has been fixed! The API now supports **both legacy and regional Firebase Realtime Database URLs**.

### ✨ What Changed:
- ✅ Now supports regional URLs: `https://PROJECT-ID.REGION.firebasedatabase.app`
- ✅ Still supports legacy URLs: `https://PROJECT-ID.firebaseio.com`
- ✅ Uses Firebase REST API directly for better compatibility
- ✅ Automatic authentication with service account credentials

---

## 🚀 Quick Setup

### Step 1: Get Your Database URL

1. Go to [Firebase Console](https://console.firebase.google.com/)
2. Select your project
3. Click **Realtime Database** in the left menu
4. Copy the URL from the **Data** tab (top of page)

**Example URLs:**
```
Legacy:   https://solanathp-default-rtdb.firebaseio.com
Regional: https://solanathp-default-rtdb.asia-southeast1.firebasedatabase.app
```

### Step 2: Download Service Account Credentials

1. In Firebase Console, click the **⚙️ gear icon** → **Project settings**
2. Go to **Service accounts** tab
3. Click **Generate new private key**
4. Save the downloaded JSON file as `config/firebase-credentials.json`

### Step 3: Update Your .env File

```env
FIREBASE_DATABASE_URL=https://solanathp-default-rtdb.asia-southeast1.firebasedatabase.app
FIREBASE_CREDENTIALS_FILE=./config/firebase-credentials.json
```

### Step 4: Test Connection

```bash
# Rebuild (if needed)
go build -o bin/server.exe ./cmd/server

# Run server
./bin/server.exe

# You should see:
# ✅ Firebase client initialized successfully
#    Database URL: https://solanathp-default-rtdb.asia-southeast1.firebasedatabase.app
```

---

## 🔒 Firebase Security Rules

### For Development (Open Access)

```json
{
  "rules": {
    ".read": true,
    ".write": true
  }
}
```

### For Production (Authenticated Only)

```json
{
  "rules": {
    ".read": "auth != null",
    ".write": "auth != null",
    "trades": {
      "$tradeId": {
        ".read": true,
        ".write": true
      }
    },
    "users": {
      "$userId": {
        ".read": true,
        ".write": true
      }
    }
  }
}
```

**To update rules:**
1. Go to Firebase Console → Realtime Database
2. Click **Rules** tab
3. Paste the rules above
4. Click **Publish**

---

## 📊 Database Structure

Your data will be organized like this:

```
firebase-root/
├── trades/
│   ├── trade-id-1/
│   │   ├── id: "trade-id-1"
│   │   ├── userId: "user123"
│   │   ├── symbol: "BTCUSDT"
│   │   ├── side: "BUY"
│   │   ├── entryPrice: 45000
│   │   ├── stopLoss: 44000
│   │   ├── takeProfit: 47000
│   │   ├── leverage: 10
│   │   ├── size: 100
│   │   ├── status: "ACTIVE"
│   │   ├── orderId: 123456789
│   │   ├── createdAt: 1696857600
│   │   └── executedAt: 1696857601
│   └── trade-id-2/
│       └── ...
│
└── users/
    └── user123/
        └── trades/
            ├── trade-id-1/
            │   └── (same data as above)
            └── trade-id-2/
                └── ...
```

---

## 🧪 Testing Firebase Connection

### Test 1: Health Check

```bash
curl http://localhost:8080/health

# Expected:
# {"status":"healthy","time":1696857600}
```

### Test 2: Place a Trade

```bash
curl -X POST http://localhost:8080/api/trade \
  -H "X-API-Key: YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "userId": "testuser",
    "symbol": "BTCUSDT",
    "side": "BUY",
    "entryPrice": 45000,
    "stopLoss": 44000,
    "takeProfit": 47000,
    "leverage": 10,
    "size": 100
  }'
```

### Test 3: Verify in Firebase Console

1. Go to Firebase Console → Realtime Database → Data
2. You should see the trade under `trades/` and `users/testuser/trades/`

---

## ⚠️ Common Issues & Solutions

### Issue 1: "Failed to initialize Firebase"

**Symptom:**
```
Failed to initialize Firebase: error initializing firebase app: ...
```

**Solutions:**
1. Check `FIREBASE_CREDENTIALS_FILE` path is correct
2. Verify the JSON file is valid (not corrupted)
3. Ensure file permissions allow reading

**Fix:**
```bash
# Check file exists
ls -la config/firebase-credentials.json

# Verify it's valid JSON
cat config/firebase-credentials.json | jq .

# Fix permissions (if needed)
chmod 644 config/firebase-credentials.json
```

### Issue 2: "Invalid database URL" (FIXED)

**Symptom:**
```
want host: "firebaseio.com"
```

**Solution:**
✅ This has been fixed! The new implementation supports regional URLs.

### Issue 3: "Permission denied"

**Symptom:**
```
firebase request failed with status 401: Permission denied
```

**Solutions:**
1. Update Firebase security rules (see above)
2. Verify service account credentials are correct
3. Check credentials JSON has the right permissions

**Fix in Firebase Console:**
1. Realtime Database → Rules
2. Set development rules (see above)
3. Click **Publish**

### Issue 4: "Failed to save trade"

**Symptom:**
```
failed to save trade: firebase request failed with status 403
```

**Solutions:**
1. Check Firebase security rules allow write access
2. Verify database URL is correct
3. Check network/firewall allows HTTPS to Firebase

---

## 🔧 Advanced Configuration

### Using Environment Variables for Credentials

Instead of a file, you can use inline credentials:

```env
FIREBASE_DATABASE_URL=https://your-project.firebaseio.com
FIREBASE_CREDENTIALS_JSON={"type":"service_account","project_id":"..."}
```

(Note: This requires code modification)

### Using Multiple Environments

```bash
# Development
FIREBASE_DATABASE_URL=https://dev-project.firebaseio.com
FIREBASE_CREDENTIALS_FILE=./config/firebase-dev.json

# Production
FIREBASE_DATABASE_URL=https://prod-project.firebaseio.com
FIREBASE_CREDENTIALS_FILE=./config/firebase-prod.json
```

### Monitoring Firebase Usage

1. Go to Firebase Console → Usage and billing
2. Monitor:
   - **Concurrent connections**
   - **GB downloaded**
   - **GB stored**

**Free Tier Limits:**
- 100 simultaneous connections
- 1 GB stored
- 10 GB/month downloaded

---

## 📈 Performance Tips

1. **Use Indexed Queries** - Add `.indexOn` rules for better performance
2. **Limit Data Downloads** - Use Firebase queries to filter data
3. **Use Connection Pooling** - Already implemented via HTTP client
4. **Cache Reads** - Implement caching for frequently accessed data
5. **Batch Writes** - Group multiple writes when possible

---

## 🆘 Support Resources

- **Firebase Documentation:** https://firebase.google.com/docs/database
- **REST API Reference:** https://firebase.google.com/docs/reference/rest/database
- **Security Rules Guide:** https://firebase.google.com/docs/database/security
- **Project Issues:** [GitHub Issues](https://github.com/your-repo/issues)

---

## ✅ Verification Checklist

Before deploying to production:

- [ ] Firebase Realtime Database created
- [ ] Service account credentials downloaded
- [ ] Credentials saved as `config/firebase-credentials.json`
- [ ] `.env` file updated with correct `FIREBASE_DATABASE_URL`
- [ ] Security rules configured
- [ ] Server starts without errors
- [ ] Can write test trade to database
- [ ] Can read trade from database
- [ ] Data visible in Firebase Console

---

**Status:** ✅ Regional database URL support added
**Updated:** 2025-10-09
**Version:** 1.1.0
