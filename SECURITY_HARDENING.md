# 🔒 Security Hardening Guide

**Production Security Checklist for Real Money Trading**

---

## 🎯 Security Levels

| Level | Description | Suitable For |
|-------|-------------|--------------|
| **Basic** | Minimum security | Testing only |
| **Standard** | Good security | Small accounts ($500-5k) |
| **Advanced** | Strong security | Medium accounts ($5k-50k) |
| **Enterprise** | Maximum security | Large accounts ($50k+) |

---

## 🔐 1. API Key Security

### Level: Basic (Minimum Required)

```bash
# Generate strong API key
openssl rand -base64 48

# Store in .env (never commit to Git)
API_KEY=generated-key-here

# Verify .gitignore includes .env
echo ".env" >> .gitignore
```

### Level: Standard (Recommended)

- ✅ Different API keys for testnet vs production
- ✅ Rotate keys every 90 days
- ✅ Store backup in password manager
- ✅ Never share or expose keys

### Level: Advanced

- ✅ Use environment variable management service (HashiCorp Vault, AWS Secrets Manager)
- ✅ Implement key rotation automation
- ✅ Audit log for key usage
- ✅ Multi-user access with separate keys

---

## 🏦 2. Binance API Security

### Critical Configuration

**Required Permissions (Enable ONLY these):**
- ✅ Enable Futures
- ✅ Enable Spot & Margin Trading (if needed)

**NEVER Enable:**
- ❌ Enable Withdrawals
- ❌ Enable Internal Transfer
- ❌ Universal Transfer

**IP Whitelist (CRITICAL):**

```bash
# Get your server IP
curl ifconfig.me

# Add to Binance API whitelist:
# Format: 123.45.67.89/32
# DO NOT use 0.0.0.0/0
```

### Additional Protection

```bash
# Test API key restrictions
curl -H "X-MBX-APIKEY: YOUR_KEY" \
     https://fapi.binance.com/fapi/v2/account

# Should work from whitelisted IP
# Should fail from other IPs
```

---

## 🔥 3. Firebase Security

### Database Rules (Production)

**Replace default rules with:**

```json
{
  "rules": {
    // Default: Deny all
    ".read": false,
    ".write": false,

    // Allow authenticated access only
    "trades": {
      ".read": "auth != null",
      ".write": "auth != null",

      "$tradeId": {
        ".validate": "newData.hasChildren(['userId', 'symbol', 'side', 'entryPrice', 'leverage'])"
      }
    },

    "users": {
      "$userId": {
        ".read": "auth != null && auth.uid == $userId",
        ".write": "auth != null && auth.uid == $userId"
      }
    },

    "system": {
      ".read": "auth != null && root.child('admins').child(auth.uid).exists()",
      ".write": "auth != null && root.child('admins').child(auth.uid).exists()"
    }
  }
}
```

### Service Account Security

```bash
# Restrict service account permissions
# Go to: Firebase Console → Project Settings → Service Accounts

# Recommended: Create custom role with only:
# - Firebase Realtime Database Admin
# - Firebase Authentication Admin (if using auth)

# DO NOT use "Editor" or "Owner" roles
```

---

## 🛡️ 4. Server Security (Oracle Cloud)

### Firewall Configuration

```bash
# Ubuntu UFW
sudo ufw default deny incoming
sudo ufw default allow outgoing
sudo ufw allow 22/tcp    # SSH
sudo ufw allow 80/tcp    # HTTP
sudo ufw allow 443/tcp   # HTTPS
sudo ufw enable

# Oracle Cloud Security List
# Go to: Networking → Security Lists
# Add rules:
# - TCP 22 (SSH) from your IP only
# - TCP 80 (HTTP) from 0.0.0.0/0
# - TCP 443 (HTTPS) from 0.0.0.0/0
# - Remove TCP 8080 (use Nginx instead)
```

### SSH Hardening

```bash
# Edit SSH config
sudo nano /etc/ssh/sshd_config

# Add/modify these lines:
PermitRootLogin no
PasswordAuthentication no
PubkeyAuthentication yes
Port 2222  # Change from default 22

# Restart SSH
sudo systemctl restart sshd

# Update firewall for new port
sudo ufw delete allow 22/tcp
sudo ufw allow 2222/tcp
```

### Fail2Ban (Brute Force Protection)

```bash
# Install
sudo apt install fail2ban -y

# Configure
sudo cp /etc/fail2ban/jail.conf /etc/fail2ban/jail.local
sudo nano /etc/fail2ban/jail.local

# Set:
[sshd]
enabled = true
maxretry = 3
bantime = 3600

# Start
sudo systemctl enable fail2ban
sudo systemctl start fail2ban

# Check banned IPs
sudo fail2ban-client status sshd
```

---

## 🔐 5. SSL/TLS (HTTPS)

### Using Let's Encrypt (Free)

```bash
# Install Certbot
sudo apt install certbot python3-certbot-nginx -y

# Get certificate (requires domain)
sudo certbot certonly --standalone -d yourdomain.com

# Auto-renewal
sudo certbot renew --dry-run

# Setup cron for auto-renewal
sudo crontab -e
# Add: 0 0 * * 0 certbot renew --quiet
```

### Nginx HTTPS Configuration

Create `nginx/nginx.conf`:

```nginx
events {
    worker_connections 1024;
}

http {
    # Security headers
    add_header X-Frame-Options "SAMEORIGIN" always;
    add_header X-Content-Type-Options "nosniff" always;
    add_header X-XSS-Protection "1; mode=block" always;
    add_header Strict-Transport-Security "max-age=31536000; includeSubDomains" always;

    upstream api {
        server crypto-api:8080;
    }

    # HTTP redirect
    server {
        listen 80;
        server_name yourdomain.com;
        return 301 https://$server_name$request_uri;
    }

    # HTTPS
    server {
        listen 443 ssl http2;
        server_name yourdomain.com;

        ssl_certificate /etc/letsencrypt/live/yourdomain.com/fullchain.pem;
        ssl_certificate_key /etc/letsencrypt/live/yourdomain.com/privkey.pem;
        ssl_protocols TLSv1.2 TLSv1.3;
        ssl_ciphers HIGH:!aNULL:!MD5;
        ssl_prefer_server_ciphers on;

        # Rate limiting
        limit_req_zone $binary_remote_addr zone=api_limit:10m rate=10r/s;

        location / {
            limit_req zone=api_limit burst=20;

            proxy_pass http://api;
            proxy_set_header Host $host;
            proxy_set_header X-Real-IP $remote_addr;
            proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
            proxy_set_header X-Forwarded-Proto $scheme;

            # Timeouts
            proxy_connect_timeout 60s;
            proxy_send_timeout 60s;
            proxy_read_timeout 60s;
        }
    }
}
```

---

## 🔍 6. Monitoring & Logging

### Docker Logging

```yaml
# In docker-compose.yml
services:
  crypto-api:
    logging:
      driver: "json-file"
      options:
        max-size: "10m"
        max-file: "3"
```

### Application Logging

Logs are already implemented in `middleware.go`. Monitor with:

```bash
# Real-time logs
docker compose logs -f

# Filter errors only
docker compose logs | grep ERROR

# Export logs
docker compose logs > logs_$(date +%Y%m%d).txt
```

### System Monitoring

```bash
# Install monitoring tools
sudo apt install htop iotop nethogs -y

# Check resources
htop
docker stats

# Check network
sudo nethogs

# Check disk
df -h
```

### Alert Setup (Optional)

**Slack Webhook:**

```bash
# In .env add:
SLACK_WEBHOOK_URL=https://hooks.slack.com/services/YOUR/WEBHOOK/URL

# Create alert script
#!/bin/bash
WEBHOOK_URL=$SLACK_WEBHOOK_URL
MESSAGE="$1"

curl -X POST $WEBHOOK_URL \
  -H 'Content-Type: application/json' \
  -d "{\"text\":\"$MESSAGE\"}"
```

---

## 🚨 7. Incident Response

### Security Incident Procedure

**If you suspect a breach:**

1. **Immediate Actions:**
   ```bash
   # Stop the server
   docker compose down

   # Disable Binance API keys
   # Go to: https://www.binance.com/en/my/settings/api-management
   # Click "Delete"

   # Close all positions manually
   # Go to Binance Futures, close all
   ```

2. **Investigation:**
   ```bash
   # Check logs
   docker compose logs > incident_logs.txt

   # Check Firebase access logs
   # Firebase Console → Usage → Logs

   # Check server access logs
   sudo grep "Failed" /var/log/auth.log
   sudo last -20
   ```

3. **Recovery:**
   ```bash
   # Generate new API keys (all of them)
   openssl rand -base64 48  # API_KEY
   # New Binance keys
   # New Firebase credentials

   # Rebuild and redeploy
   docker compose up -d --build

   # Monitor closely for 48 hours
   ```

### Backup Strategy

```bash
# Daily backup script
#!/bin/bash
DATE=$(date +%Y%m%d)
BACKUP_DIR="/home/ubuntu/backups"

# Backup configuration
tar -czf $BACKUP_DIR/config-$DATE.tar.gz \
    ~/tradingAPI/.env \
    ~/tradingAPI/config/

# Backup Docker volume (if any)
docker run --rm -v tradingapi_logs:/data \
    -v $BACKUP_DIR:/backup \
    alpine tar czf /backup/logs-$DATE.tar.gz /data

# Keep only last 30 days
find $BACKUP_DIR -name "*.tar.gz" -mtime +30 -delete

# Upload to cloud (optional)
# aws s3 cp $BACKUP_DIR/config-$DATE.tar.gz s3://your-bucket/
```

**Setup automatic backups:**

```bash
# Add to crontab
crontab -e

# Daily at 2 AM
0 2 * * * /home/ubuntu/backup.sh
```

---

## 🎯 8. Security Checklist

### Pre-Production

- [ ] Strong API_KEY generated (64+ characters)
- [ ] Binance API IP whitelist configured (specific IP, not 0.0.0.0/0)
- [ ] Binance withdrawals disabled
- [ ] Firebase security rules configured
- [ ] Service account with minimal permissions
- [ ] .env not committed to Git
- [ ] .gitignore configured properly

### Server Security

- [ ] UFW firewall enabled
- [ ] SSH key-only authentication
- [ ] SSH port changed from default 22
- [ ] Fail2Ban installed and configured
- [ ] SSL certificate installed
- [ ] Nginx configured with security headers
- [ ] Regular security updates enabled
- [ ] Non-root user for Docker

### Monitoring

- [ ] Logging configured and working
- [ ] Log rotation enabled
- [ ] Alert notifications setup
- [ ] Backup script running daily
- [ ] Backup tested (restore works)
- [ ] Monitoring dashboard (optional)

### Operational

- [ ] Emergency stop procedure documented
- [ ] Incident response plan ready
- [ ] Backup administrator access
- [ ] Contact information updated
- [ ] Regular security audits scheduled

---

## 🔧 9. Security Automation

### Auto-Update Script

```bash
#!/bin/bash
# auto-update.sh

# Update system
sudo apt update
sudo apt upgrade -y

# Update Docker images
cd ~/tradingAPI
docker compose pull
docker compose up -d

# Clean old images
docker system prune -f

# Log update
echo "$(date): System updated" >> /var/log/auto-update.log
```

**Setup auto-updates:**

```bash
crontab -e

# Weekly on Sunday at 3 AM
0 3 * * 0 /home/ubuntu/auto-update.sh
```

### Health Check Script

```bash
#!/bin/bash
# health-check.sh

API_URL="http://localhost:8080"
API_KEY="your-api-key"

# Check API health
health=$(curl -s --max-time 10 $API_URL/health)

if ! echo "$health" | grep -q "healthy"; then
    # Alert
    echo "API DOWN at $(date)" | mail -s "ALERT: API Down" you@email.com

    # Restart
    cd ~/tradingAPI
    docker compose restart

    # Log
    echo "$(date): API restarted" >> /var/log/health-check.log
fi
```

---

## 📊 10. Security Audit Checklist

### Monthly Review

- [ ] Review all API keys (rotate if needed)
- [ ] Check Binance API access logs
- [ ] Review Firebase access logs
- [ ] Check server auth.log for suspicious activity
- [ ] Verify backup integrity
- [ ] Test disaster recovery procedure
- [ ] Update dependencies
- [ ] Review and update firewall rules
- [ ] Check SSL certificate expiration
- [ ] Review trading bot performance and safety

### Quarterly Review

- [ ] Full security audit
- [ ] Penetration testing (optional)
- [ ] Review and update all documentation
- [ ] Test all emergency procedures
- [ ] Review and update access controls
- [ ] Compliance check (if applicable)

---

## ⚠️ Common Security Mistakes

### ❌ DON'T DO THIS:

1. **Use 0.0.0.0/0 for IP whitelist** → Anyone can access
2. **Enable withdrawals on Binance API** → Funds can be stolen
3. **Commit .env to Git** → Keys exposed publicly
4. **Use same keys for testnet and production** → Confusion and risk
5. **Skip SSL/HTTPS** → Man-in-the-middle attacks
6. **No firewall** → Open to attacks
7. **Default SSH port 22** → Automated attacks
8. **Root login enabled** → Full system compromise
9. **No backups** → Data loss
10. **No monitoring** → Undetected issues

### ✅ DO THIS:

1. **Specific IP whitelist** → Only your server
2. **Disable withdrawals** → Can't steal funds
3. **Keep .env private** → Secrets stay secret
4. **Separate keys per environment** → Clear separation
5. **Use HTTPS** → Encrypted traffic
6. **Enable firewall** → Protected server
7. **Change SSH port** → Fewer attacks
8. **Disable root login** → Limited damage
9. **Daily backups** → Can recover
10. **Active monitoring** → Quick detection

---

## 🆘 Emergency Contacts

**In case of security incident:**

1. **Binance Support:** https://www.binance.com/en/support
2. **Oracle Cloud Support:** https://cloud.oracle.com/support
3. **Firebase Support:** https://firebase.google.com/support

**Keep these readily accessible:**
- Binance account recovery info
- Firebase project owner credentials
- Oracle Cloud admin access
- Backup administrator contact

---

## ✅ Security Status

Use this to track your security level:

| Component | Basic | Standard | Advanced | Status |
|-----------|-------|----------|----------|--------|
| API Key Security | ⚪ | ⚪ | ⚪ | ☐ |
| Binance API | ⚪ | ⚪ | ⚪ | ☐ |
| Firebase Security | ⚪ | ⚪ | ⚪ | ☐ |
| Server Firewall | ⚪ | ⚪ | ⚪ | ☐ |
| SSL/HTTPS | ⚪ | ⚪ | ⚪ | ☐ |
| SSH Security | ⚪ | ⚪ | ⚪ | ☐ |
| Monitoring | ⚪ | ⚪ | ⚪ | ☐ |
| Backups | ⚪ | ⚪ | ⚪ | ☐ |
| Incident Response | ⚪ | ⚪ | ⚪ | ☐ |

**Minimum for Production:** All "Standard" ✅
**Recommended:** Mix of Standard and Advanced
**Large Accounts:** All "Advanced" ✅

---

**Remember:** Security is ongoing, not one-time!

**Review and update regularly.**
