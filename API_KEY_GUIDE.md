# 🔐 API Key Generation Guide

## Overview

The `API_KEY` in your `.env` file is used to authenticate requests to your trading API. This key should be:
- **Long** (at least 32 characters)
- **Random** (cryptographically secure)
- **Unique** (different from any other service)
- **Secret** (never committed to version control)

---

## 🚀 Quick Generation Methods

### **Method 1: Using PowerShell (Windows)** ⭐ Recommended

```powershell
# Run the provided script
.\script\generate_api_key.ps1

# Or generate directly:
$bytes = New-Object byte[] 48
[System.Security.Cryptography.RandomNumberGenerator]::Create().GetBytes($bytes)
$apiKey = [Convert]::ToBase64String($bytes)
Write-Host "API_KEY=$apiKey"
```

### **Method 2: Using Bash/Git Bash (Linux/Mac/Windows)**

```bash
# Run the provided script
./script/generate_api_key.sh

# Or generate directly with OpenSSL:
openssl rand -base64 48

# Or with /dev/urandom:
cat /dev/urandom | tr -dc 'a-zA-Z0-9' | fold -w 64 | head -n 1
```

### **Method 3: Using Node.js**

```bash
node -e "console.log(require('crypto').randomBytes(48).toString('base64'))"
```

### **Method 4: Using Python**

```bash
python -c "import secrets; print(secrets.token_urlsafe(48))"

# Or:
python -c "import os, base64; print(base64.b64encode(os.urandom(48)).decode())"
```

### **Method 5: Using Go**

```bash
go run -e 'package main; import("crypto/rand";"encoding/base64";"fmt"); func main(){b:=make([]byte,48);rand.Read(b);fmt.Println(base64.StdEncoding.EncodeToString(b))}'
```

### **Method 6: Online Generators** ⚠️ Use with caution

- https://www.random.org/strings/ (Generate 64 random alphanumeric)
- https://www.uuidgenerator.net/ (Generate 2 UUIDs and concatenate)

**Note:** Online generators are less secure. Use offline methods for production.

---

## 📝 How to Add API Key to .env

### Option 1: Automatic (PowerShell)

```powershell
# Generate and add to .env in one command
.\script\generate_api_key.ps1
# Copy one of the keys, then:
Add-Content -Path .env -Value "API_KEY=YOUR_KEY_HERE"
```

### Option 2: Automatic (Bash)

```bash
# Generate and add to .env
echo "API_KEY=$(openssl rand -base64 48)" >> .env
```

### Option 3: Manual

1. Run any generation method above
2. Copy the generated key
3. Open `.env` file
4. Replace `your-secret-api-key-here` with your key:

```env
API_KEY=Xp9K2mN5vQ8rT1wY6zC3bF0hJ4lM7nR9sV2xA5dG8kP1qW4tE6yU3iO7pL0mN
```

---

## ✅ Example Keys (Different Formats)

### Base64 Format (Recommended)
```
Xp9K2mN5vQ8rT1wY6zC3bF0hJ4lM7nR9sV2xA5dG8kP1qW4tE6yU3iO7pL0mN5vQ8rT1wY
```

### Hex Format
```
3a7b8c9d4e5f6a1b2c3d4e5f6a7b8c9d4e5f6a7b8c9d4e5f6a7b8c9d4e5f6a7b8c9d
```

### Alphanumeric Format
```
aB3cD5eF7gH9jK2mN4pQ6rS8tU1vW3xY5zA7bC9dE2fG4hJ6kL8mN1oP3qR5sT7uV
```

### UUID-based Format
```
550e8400-e29b-41d4-a716-446655440000-7c9e6679-7425-40de-944b-e07fc1f90ae7
```

---

## 🔒 Security Best Practices

### ✅ DO:
- ✅ Generate keys using cryptographically secure methods
- ✅ Use at least 32 bytes (256 bits) of randomness
- ✅ Store keys in `.env` file (which is in `.gitignore`)
- ✅ Use different keys for development and production
- ✅ Rotate keys periodically (every 90 days recommended)
- ✅ Use environment variables in production
- ✅ Limit key access to only necessary personnel

### ❌ DON'T:
- ❌ Use simple passwords or dictionary words
- ❌ Use predictable patterns (e.g., "123456", "password123")
- ❌ Commit `.env` file to Git
- ❌ Share keys in Slack, email, or other unsecured channels
- ❌ Use the same key across multiple environments
- ❌ Store keys in your code
- ❌ Use online generators for production keys

---

## 🧪 Testing Your API Key

After setting your API key, test it:

```bash
# Start your server
./bin/server.exe

# Test health endpoint (no auth required)
curl http://localhost:8080/health

# Test authenticated endpoint
curl http://localhost:8080/api/trade \
  -H "X-API-Key: YOUR_ACTUAL_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{"userId":"test","symbol":"BTCUSDT","side":"BUY","entryPrice":45000,"stopLoss":44000,"takeProfit":47000,"leverage":10,"size":100}'
```

### Expected Results:
- ✅ **With valid key:** HTTP 200 (or 400 if invalid trade data)
- ❌ **Without key:** HTTP 401 "Missing API key"
- ❌ **With invalid key:** HTTP 401 "Invalid API key"

---

## 🔄 Key Rotation Guide

### When to Rotate:
- Every 90 days (recommended)
- After a security incident
- When team member leaves
- If key may have been exposed

### How to Rotate:

1. **Generate new key:**
   ```powershell
   .\script\generate_api_key.ps1
   ```

2. **Update `.env` file:**
   ```env
   API_KEY=NEW_KEY_HERE
   OLD_API_KEY=OLD_KEY_HERE  # Temporary dual support
   ```

3. **Update code to support both keys temporarily** (optional)

4. **Notify all API consumers** to update their keys

5. **After transition period, remove old key:**
   ```env
   API_KEY=NEW_KEY_HERE
   ```

6. **Restart server:**
   ```bash
   docker-compose restart
   ```

---

## 💾 Storing Keys in Production

### Environment Variables (Recommended)

**Docker:**
```bash
docker run -e API_KEY="$API_KEY" crypto-trading-api
```

**Kubernetes Secret:**
```yaml
apiVersion: v1
kind: Secret
metadata:
  name: trading-api-secret
type: Opaque
stringData:
  api-key: "YOUR_KEY_HERE"
```

**Oracle Cloud / AWS / Azure:**
Use their respective secret management services:
- AWS Secrets Manager
- Azure Key Vault
- Oracle Cloud Vault
- Google Secret Manager

---

## 📊 Key Strength Comparison

| Method | Bits of Entropy | Strength | Recommended |
|--------|----------------|----------|-------------|
| 32-byte base64 | 256 bits | ⭐⭐⭐⭐⭐ Strong | ✅ Yes |
| 48-byte base64 | 384 bits | ⭐⭐⭐⭐⭐ Very Strong | ✅ Yes |
| UUID (single) | 122 bits | ⭐⭐⭐ Moderate | ⚠️ OK |
| UUID (double) | 244 bits | ⭐⭐⭐⭐ Strong | ✅ Yes |
| 16-char alphanumeric | ~95 bits | ⭐⭐ Weak | ❌ No |
| Simple password | <64 bits | ⭐ Very Weak | ❌ Never |

---

## 🛡️ Additional Security Measures

Beyond a strong API key, consider:

1. **IP Whitelisting** - Limit access to known IPs
2. **Rate Limiting** - Already implemented (100 req/min)
3. **HTTPS Only** - Use SSL/TLS in production
4. **Request Signing** - Add HMAC signatures
5. **Short-lived Tokens** - Implement JWT with expiration
6. **Audit Logging** - Log all API access
7. **Monitoring** - Alert on suspicious activity

---

## 📞 Support

If you have questions about API key security:
1. Review this guide
2. Check [SETUP.md](SETUP.md) for deployment details
3. Review [CLAUDE.md](CLAUDE.md) for architecture

---

**Remember:** Your API key is the master password to your trading system. Treat it like your bank password! 🔐
