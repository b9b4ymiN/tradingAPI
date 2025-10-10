#!/bin/bash

# Crypto Trading API - Monitoring Script
# Usage: ./monitor.sh

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

API_URL="http://localhost:8080"
CONTAINER_NAME="crypto-trading-api"

clear
echo -e "${GREEN}╔════════════════════════════════════════╗${NC}"
echo -e "${GREEN}║   Crypto Trading API Monitor v1.0     ║${NC}"
echo -e "${GREEN}╚════════════════════════════════════════╝${NC}"
echo ""

# 1. Service Status
echo -e "${BLUE}🔍 Service Status${NC}"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
if docker ps | grep -q ${CONTAINER_NAME}; then
    echo -e "${GREEN}✅ Container Status: Running${NC}"
    
    # Get container uptime
    UPTIME=$(docker inspect -f '{{.State.StartedAt}}' ${CONTAINER_NAME})
    echo -e "   Started: ${UPTIME}"
else
    echo -e "${RED}❌ Container Status: Stopped${NC}"
    echo -e "${YELLOW}   Run: docker-compose up -d${NC}"
    exit 1
fi
echo ""

# 2. Health Check
echo -e "${BLUE}🏥 Health Check${NC}"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
HEALTH_RESPONSE=$(curl -s -w "\n%{http_code}" ${API_URL}/health 2>/dev/null)
HTTP_CODE=$(echo "$HEALTH_RESPONSE" | tail -n1)
RESPONSE_BODY=$(echo "$HEALTH_RESPONSE" | head -n-1)

if [ "$HTTP_CODE" == "200" ]; then
    echo -e "${GREEN}✅ API Health: Healthy${NC}"
    echo -e "   Response: ${RESPONSE_BODY}"
else
    echo -e "${RED}❌ API Health: Unhealthy (HTTP ${HTTP_CODE})${NC}"
fi
echo ""

# 3. Resource Usage
echo -e "${BLUE}💻 Resource Usage${NC}"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
docker stats ${CONTAINER_NAME} --no-stream --format "table {{.Container}}\t{{.CPUPerc}}\t{{.MemUsage}}\t{{.MemPerc}}\t{{.NetIO}}"
echo ""

# 4. Recent Logs
echo -e "${BLUE}📝 Recent Logs (Last 10 lines)${NC}"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
docker logs ${CONTAINER_NAME} --tail 10 2>&1
echo ""

# 5. Network Status
echo -e "${BLUE}🌐 Network Status${NC}"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
PORTS=$(docker port ${CONTAINER_NAME} 2>/dev/null)
if [ ! -z "$PORTS" ]; then
    echo -e "${GREEN}✅ Port Bindings:${NC}"
    echo "$PORTS"
else
    echo -e "${YELLOW}⚠️  No port bindings found${NC}"
fi
echo ""

# 6. Disk Usage
echo -e "${BLUE}💾 Disk Usage${NC}"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
df -h | grep -E '^Filesystem|/$' | awk '{print $1"\t"$3"/"$2" ("$5")"}'
echo ""

# 7. Docker Images
echo -e "${BLUE}🐳 Docker Images${NC}"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
docker images | grep -E 'REPOSITORY|crypto'
echo ""

# 8. Response Time Test
echo -e "${BLUE}⚡ Response Time Test${NC}"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
for i in {1..3}; do
    START_TIME=$(date +%s%N)
    curl -s ${API_URL}/health > /dev/null
    END_TIME=$(date +%s%N)
    RESPONSE_TIME=$(( (END_TIME - START_TIME) / 1000000 ))
    echo -e "   Test ${i}: ${RESPONSE_TIME}ms"
done
echo ""

# 9. Memory Details
echo -e "${BLUE}🧠 Container Memory Details${NC}"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
docker inspect ${CONTAINER_NAME} | jq -r '.[0].HostConfig | "Memory Limit: \(.Memory // "unlimited")\nCPU Quota: \(.CpuQuota // "unlimited")"'
echo ""

# 10. System Info
echo -e "${BLUE}📊 System Information${NC}"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo -e "Hostname: $(hostname)"
echo -e "OS: $(uname -s) $(uname -r)"
echo -e "Architecture: $(uname -m)"
echo -e "Load Average: $(uptime | awk -F'load average:' '{print $2}')"
echo ""

# Summary
echo -e "${GREEN}╔════════════════════════════════════════╗${NC}"
echo -e "${GREEN}║          Monitoring Complete           ║${NC}"
echo -e "${GREEN}╚════════════════════════════════════════╝${NC}"
echo ""
echo -e "${YELLOW}Commands:${NC}"
echo -e "  View live logs:    ${BLUE}docker-compose logs -f crypto-api${NC}"
echo -e "  Restart service:   ${BLUE}docker-compose restart crypto-api${NC}"
echo -e "  View all stats:    ${BLUE}docker stats${NC}"
echo ""
