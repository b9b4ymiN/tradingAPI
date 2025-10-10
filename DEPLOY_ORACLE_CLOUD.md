# 🚀 Oracle Cloud Deployment Guide

**Crypto Trading API - Docker Deployment for Oracle Cloud Free Tier**

---

## 📋 Prerequisites

### Oracle Cloud Free Tier Includes:
- ✅ 2 AMD-based Compute VMs (1/8 OCPU + 1 GB RAM each)
- ✅ OR 4 Arm-based Ampere A1 cores + 24 GB RAM
- ✅ 200 GB Block Volume Storage
- ✅ 10 TB Outbound Data Transfer per month
- ✅ Free forever (no credit card charges)

### What You Need:
1. Oracle Cloud Account (free tier)
2. Ubuntu 22.04 VM instance created
3. SSH access to your VM
4. Domain name (optional, for SSL)

---

## 🔧 Step 1: Prepare Your Local Files

### Required Files for Deployment:

```
tradingAPI/
├── cmd/server/main.go              ✅ Required
├── internal/                       ✅ Required
│   ├── api/
│   ├── binance/
│   ├── firebase/
│   └── models/
├── config/
│   ├── config.go                   ✅ Required
│   └── firebase-credentials.json   ✅ Required (your file)
├── go.mod                          ✅ Required
├── go.sum                          ✅ Required
├── Dockerfile                      ✅ Required
├── docker-compose.yml              ✅ Required
├── .dockerignore                   ✅ Required
├── .env.example                    ✅ Required
└── .gitignore                      ✅ Required
```

### Files Already Removed (Not Needed):
- ❌ bin/server.exe (local binary)
- ❌ Test scripts (script/test_*.sh)
- ❌ Test documentation (TEST_*.md)
- ❌ Analysis reports
- ❌ Logs (*.log)
- ❌ .claude/ directory

---

## 🖥️ Step 2: Create Oracle Cloud VM

### 2.1 Login to Oracle Cloud

1. Go to: https://cloud.oracle.com
2. Login to your account
3. Navigate to: **Compute → Instances**

### 2.2 Create Instance

**Recommended Configuration:**

| Setting | Value |
|---------|-------|
| **Name** | crypto-trading-api |
| **Image** | Ubuntu 22.04 Minimal |
| **Shape** | VM.Standard.E2.1.Micro (Free Tier) |
| **OCPU** | 1 |
| **Memory** | 1 GB RAM |
| **Boot Volume** | 50 GB |
| **Network** | Default VCN |
| **Public IP** | Assign public IP |
| **SSH Keys** | Add your SSH public key |

### 2.3 Open Firewall Ports

**In Oracle Cloud Console:**

1. Navigate to: **Networking → Virtual Cloud Networks**
2. Select your VCN
3. Click: **Security Lists → Default Security List**
4. Click: **Add Ingress Rules**

**Add these rules:**

| Port | Protocol | Source | Description |
|------|----------|--------|-------------|
| 22 | TCP | 0.0.0.0/0 | SSH |
| 80 | TCP | 0.0.0.0/0 | HTTP |
| 443 | TCP | 0.0.0.0/0 | HTTPS (optional) |
| 8080 | TCP | 0.0.0.0/0 | API (temporary, for testing) |

**In Ubuntu Firewall (after SSH):**

```bash
# Allow ports through Ubuntu firewall
sudo iptables -I INPUT 6 -m state --state NEW -p tcp --dport 80 -j ACCEPT
sudo iptables -I INPUT 6 -m state --state NEW -p tcp --dport 443 -j ACCEPT
sudo iptables -I INPUT 6 -m state --state NEW -p tcp --dport 8080 -j ACCEPT
sudo netfilter-persistent save
```

---

## 📦 Step 3: Install Required Software on VM

### 3.1 SSH to Your VM

```bash
ssh ubuntu@YOUR_VM_IP
```

### 3.2 Update System

```bash
sudo apt update && sudo apt upgrade -y
```

### 3.3 Install Docker

```bash
# Install Docker
curl -fsSL https://get.docker.com -o get-docker.sh
sudo sh get-docker.sh

# Add user to docker group
sudo usermod -aG docker ubuntu

# Enable Docker service
sudo systemctl enable docker
sudo systemctl start docker

# Verify installation
docker --version
```

### 3.4 Install Docker Compose

```bash
# Install Docker Compose
sudo apt install docker-compose-plugin -y

# Verify installation
docker compose version
```

---

## 📤 Step 4: Upload Project Files

### Option 1: Using Git (Recommended)

```bash
# On VM
cd ~
git clone YOUR_REPOSITORY_URL tradingAPI
cd tradingAPI
```

### Option 2: Using SCP (Direct Upload)

```bash
# On your local machine
cd /c/Programing/go/tradingAPI

# Create tar archive (excludes unnecessary files)
tar --exclude='.git' \
    --exclude='bin' \
    --exclude='*.log' \
    --exclude='.claude' \
    -czf tradingAPI.tar.gz .

# Upload to VM
scp tradingAPI.tar.gz ubuntu@YOUR_VM_IP:~/

# On VM - Extract
ssh ubuntu@YOUR_VM_IP
cd ~
mkdir -p tradingAPI
tar -xzf tradingAPI.tar.gz -C tradingAPI/
cd tradingAPI
```

### Option 3: Using rsync (Recommended for Updates)

```bash
# On your local machine
cd /c/Programing/go/tradingAPI

# Sync files to VM (excludes via .gitignore)
rsync -avz --progress \
    --exclude='.git' \
    --exclude='bin/' \
    --exclude='*.log' \
    --exclude='.claude/' \
    ./ ubuntu@YOUR_VM_IP:~/tradingAPI/
```

---

## ⚙️ Step 5: Configure Environment

### 5.1 Create .env File

```bash
# On VM
cd ~/tradingAPI

# Copy example
cp .env.example .env

# Edit with your values
nano .env
```

### 5.2 Configure .env

```bash
# Server Configuration
PORT=8080
GIN_MODE=release

# API Security (generate new key!)
API_KEY=your-generated-api-key-here

# Binance Configuration
BINANCE_TESTNET=false
BINANCE_API_KEY=your-binance-api-key
BINANCE_SECRET_KEY=your-binance-secret-key

# Firebase Configuration
FIREBASE_DATABASE_URL=https://your-project.firebaseio.com
FIREBASE_CREDENTIALS_FILE=./config/firebase-credentials.json
```

**Generate API Key:**
```bash
# On VM
openssl rand -base64 48
```

### 5.3 Upload Firebase Credentials

**From local machine:**
```bash
scp config/firebase-credentials.json ubuntu@YOUR_VM_IP:~/tradingAPI/config/
```

**Verify:**
```bash
# On VM
ls -la ~/tradingAPI/config/firebase-credentials.json
```

---

## 🐳 Step 6: Build and Deploy with Docker

### 6.1 Build Docker Image

```bash
cd ~/tradingAPI

# Build image
docker compose build

# Or build manually
docker build -t crypto-trading-api:latest .
```

**Expected build time:** 2-5 minutes

### 6.2 Start Services

```bash
# Start in detached mode
docker compose up -d

# Check status
docker compose ps
```

**Expected output:**
```
NAME                    IMAGE                      STATUS
crypto-trading-api      tradingapi-crypto-api     Up 30 seconds
```

### 6.3 Verify Deployment

```bash
# Check logs
docker compose logs -f

# Expected output:
# ✅ Firebase client initialized successfully
# ✅ Binance client initialized successfully
# 🚀 Server starting on port 8080
```

**Test API:**
```bash
# Health check
curl http://localhost:8080/health

# Expected: {"status":"healthy","time":...}
```

---

## 🔍 Step 7: Test Your API

### Test from VM

```bash
# Set API key
export API_KEY="your-api-key"

# Test health
curl http://localhost:8080/health

# Test with authentication
curl -H "X-API-Key: $API_KEY" http://localhost:8080/api/status

# Test balance
curl -H "X-API-Key: $API_KEY" http://localhost:8080/api/balance
```

### Test from Your Computer

```bash
# Replace YOUR_VM_IP with actual IP
export API_URL="http://YOUR_VM_IP:8080"
export API_KEY="your-api-key"

# Test endpoints
curl $API_URL/health
curl -H "X-API-Key: $API_KEY" $API_URL/api/status
curl -H "X-API-Key: $API_KEY" $API_URL/api/balance
```

---

## 🔒 Step 8: Security Hardening (Recommended)

### 8.1 Disable Direct Port 8080 Access

After testing, only allow access through Nginx (port 80/443):

```bash
# Remove port 8080 from Oracle Cloud Security List
# (via console as done in Step 2.3)

# Update docker-compose.yml to not expose 8080
# (only Nginx should access it internally)
```

### 8.2 Setup SSL with Let's Encrypt (Optional)

```bash
# Install Certbot
sudo apt install certbot python3-certbot-nginx -y

# Get certificate (requires domain)
sudo certbot --nginx -d yourdomain.com
```

### 8.3 Enable Nginx (Optional but Recommended)

Create `nginx/nginx.conf`:

```nginx
events {
    worker_connections 1024;
}

http {
    upstream api {
        server crypto-api:8080;
    }

    server {
        listen 80;
        server_name yourdomain.com;

        location / {
            proxy_pass http://api;
            proxy_set_header Host $host;
            proxy_set_header X-Real-IP $remote_addr;
        }
    }
}
```

**Start Nginx:**
```bash
# Uncomment nginx section in docker-compose.yml
docker compose up -d
```

---

## 📊 Step 9: Monitoring & Maintenance

### Check Status

```bash
# Container status
docker compose ps

# View logs
docker compose logs -f

# View last 100 lines
docker compose logs --tail=100

# Check resource usage
docker stats
```

### Restart Services

```bash
# Restart
docker compose restart

# Stop
docker compose stop

# Start
docker compose start

# Rebuild and restart
docker compose up -d --build
```

### Update Deployment

```bash
# Pull latest code
cd ~/tradingAPI
git pull

# Or rsync from local
# rsync -avz ./ ubuntu@YOUR_VM_IP:~/tradingAPI/

# Rebuild and restart
docker compose up -d --build
```

---

## 🚨 Troubleshooting

### Issue: Container Won't Start

```bash
# Check logs
docker compose logs

# Common issues:
# 1. Missing .env file
# 2. Missing firebase-credentials.json
# 3. Invalid API keys

# Check config
cat .env
ls -la config/firebase-credentials.json
```

### Issue: Can't Access API from Outside

```bash
# Check firewall
sudo iptables -L -n | grep 8080

# Check if container is running
docker compose ps

# Check port binding
netstat -tlnp | grep 8080

# Check Oracle Cloud Security List
# (via console)
```

### Issue: Out of Memory

```bash
# Check memory usage
free -h
docker stats

# Restart container
docker compose restart

# Consider upgrading to Ampere A1 (4 cores, 24GB RAM, still free)
```

### Issue: Binance API Errors

```bash
# Check API keys are correct
docker compose exec crypto-api env | grep BINANCE

# Check testnet vs production
# Make sure BINANCE_TESTNET matches your keys

# Test Binance connectivity
curl https://fapi.binance.com/fapi/v1/ping
# OR for testnet
curl https://testnet.binancefuture.com/fapi/v1/ping
```

---

## 📁 Directory Structure on VM

```
~/tradingAPI/
├── cmd/                    # Source code
├── internal/               # Source code
├── config/
│   ├── config.go
│   └── firebase-credentials.json  # Your credentials
├── .env                    # Your configuration
├── docker-compose.yml
├── Dockerfile
└── logs/                   # Created by Docker
```

---

## 🔄 Backup & Recovery

### Backup Important Files

```bash
# Backup configuration
cd ~
tar -czf tradingAPI-backup-$(date +%Y%m%d).tar.gz \
    tradingAPI/.env \
    tradingAPI/config/firebase-credentials.json

# Download backup
# On local machine:
scp ubuntu@YOUR_VM_IP:~/tradingAPI-backup-*.tar.gz ./
```

### Restore from Backup

```bash
# Upload backup
scp tradingAPI-backup-*.tar.gz ubuntu@YOUR_VM_IP:~/

# Extract
ssh ubuntu@YOUR_VM_IP
tar -xzf tradingAPI-backup-*.tar.gz -C ~/
```

---

## 💰 Oracle Cloud Free Tier Limits

### Free Forever Includes:

- ✅ 2 AMD VMs (1GB RAM each) **OR** 4 ARM cores + 24GB RAM
- ✅ 200 GB Block Storage
- ✅ 10 TB Outbound Data Transfer/month
- ✅ Always Free (no expiration)

### Monitoring Usage:

```bash
# Check disk usage
df -h

# Check memory
free -h

# Check network usage (Oracle Console)
# Go to: Observability & Management → Monitoring
```

---

## 🎯 Production Checklist

Before going live with real trading:

- [ ] Oracle Cloud VM created and configured
- [ ] Docker and Docker Compose installed
- [ ] Project files uploaded
- [ ] .env configured with production keys
- [ ] Firebase credentials uploaded
- [ ] Firewall ports opened (22, 80, 443, 8080)
- [ ] Docker containers running
- [ ] Health check passing
- [ ] All API endpoints tested
- [ ] Binance production keys configured
- [ ] Firebase rules configured
- [ ] SSL certificate installed (optional)
- [ ] Nginx reverse proxy configured (optional)
- [ ] Monitoring setup
- [ ] Backups configured
- [ ] Test trade executed successfully

---

## 📚 Quick Commands Reference

```bash
# Deploy
cd ~/tradingAPI
docker compose up -d --build

# Check logs
docker compose logs -f

# Restart
docker compose restart

# Stop
docker compose down

# Update
git pull && docker compose up -d --build

# Test API
curl http://localhost:8080/health

# View resources
docker stats

# Clean up
docker system prune -a
```

---

## 🆘 Support

### Documentation:
- API Endpoints: [API_ENDPOINTS.md](API_ENDPOINTS.md)
- Quick Start: [QUICK_START.md](QUICK_START.md)
- Setup Guide: [SETUP.md](SETUP.md)

### Oracle Cloud:
- Console: https://cloud.oracle.com
- Documentation: https://docs.oracle.com/en-us/iaas/

### Docker:
- Documentation: https://docs.docker.com/
- Compose: https://docs.docker.com/compose/

---

**Deployment Checklist:** ✅ Ready for Oracle Cloud Free Tier
**Estimated Setup Time:** 30-45 minutes
**Difficulty:** Intermediate

🚀 **Your API is now running 24/7 on Oracle Cloud for FREE!**
