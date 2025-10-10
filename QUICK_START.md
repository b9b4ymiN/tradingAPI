# ⚡ Quick Start Guide

## 🚀 Get Running in 5 Minutes

### Step 1: Generate API Key (Choose one method)

**Windows PowerShell:**
```powershell
.\script\generate_api_key.ps1
```

**Linux/Mac/Git Bash:**
```bash
./script/generate_api_key.sh
```

**Or generate directly:**
```bash
# Using OpenSSL
openssl rand -base64 48

# Using Python
python -c "import secrets; print(secrets.token_urlsafe(48))"

# Using Node.js
node -e "console.log(require('crypto').randomBytes(48).toString('base64'))"
```

### Step 2: Setup Environment

```bash
# Copy template
cp .env.example .env

# Edit .env and add your keys
nano .env  # or use any text editor
```

Required values:
```env
API_KEY=<PASTE_GENERATED_KEY_HERE>
BINANCE_API_KEY=<YOUR_BINANCE_KEY>
BINANCE_SECRET_KEY=<YOUR_BINANCE_SECRET>
FIREBASE_DATABASE_URL=https://your-project.firebaseio.com
```

### Step 3: Setup Firebase

```bash
# Create config directory
mkdir -p config

# Download firebase-credentials.json from Firebase Console
# Save it to: config/firebase-credentials.json
```

### Step 4: Run

**Option A: Use pre-compiled binary**
```bash
./bin/server.exe
```

**Option B: Build and run**
```bash
go build -o bin/server ./cmd/server
./bin/server
```

**Option C: Use Docker**
```bash
docker-compose up -d
```

### Step 5: Test

```bash
# Health check
curl http://localhost:8080/health

# Should return:
# {"status":"healthy","time":1234567890}
```

---

## 📚 Full Documentation

- **Setup Guide:** [SETUP.md](SETUP.md) - Complete deployment instructions
- **API Key Guide:** [API_KEY_GUIDE.md](API_KEY_GUIDE.md) - Security & generation
- **Architecture:** [CLAUDE.md](CLAUDE.md) - System design
- **Git Ignore:** [.gitignore](.gitignore) - What's excluded from Git

---

## 🔑 Example: Complete Setup

```bash
# 1. Generate API key
openssl rand -base64 48
# Output: habWXyQmkhtN+6jYxPNwoVRCq6Td1UKO0KJsAlcAbG/2ap5KFfyMznzQhXIU12Dw

# 2. Create .env file
cat > .env << 'EOF'
PORT=8080
GIN_MODE=release
API_KEY=habWXyQmkhtN+6jYxPNwoVRCq6Td1UKO0KJsAlcAbG/2ap5KFfyMznzQhXIU12Dw
BINANCE_API_KEY=your_binance_key_here
BINANCE_SECRET_KEY=your_binance_secret_here
FIREBASE_DATABASE_URL=https://your-project.firebaseio.com
FIREBASE_CREDENTIALS_FILE=./config/firebase-credentials.json
TZ=Asia/Bangkok
EOF

# 3. Setup Firebase
mkdir -p config
# Place your firebase-credentials.json in config/

# 4. Run
./bin/server.exe

# 5. Test
curl http://localhost:8080/health
```

---

## ⚠️ Common Issues

**Issue:** "API_KEY environment variable is required"
- **Fix:** Make sure .env file exists and has API_KEY set

**Issue:** "Failed to initialize Firebase"
- **Fix:** Check firebase-credentials.json exists in config/
- **Fix:** Verify FIREBASE_DATABASE_URL is correct

**Issue:** "Failed to connect to Binance"
- **Fix:** Check BINANCE_API_KEY and BINANCE_SECRET_KEY
- **Fix:** Ensure Futures API is enabled on Binance account
- **Fix:** Whitelist your IP in Binance API settings

**Issue:** "Cannot find package"
- **Fix:** Run `go mod tidy` and rebuild

---

## 🎯 Next Steps After Running

1. Test with Binance Testnet first
2. Read [SETUP.md](SETUP.md) for production deployment
3. Configure Nginx for SSL/HTTPS
4. Setup monitoring and alerts
5. Test all API endpoints

---

**Status:** ✅ Ready to run
**Time to setup:** ~5 minutes
**Documentation:** Complete
