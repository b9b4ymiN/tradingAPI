#!/bin/bash

# Crypto Trading API - Deployment Script for Oracle Cloud
# Usage: ./deploy.sh [production|development]

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

ENV=${1:-development}

echo -e "${GREEN}🚀 Starting deployment for ${ENV} environment...${NC}"

# Check if .env file exists
if [ ! -f .env ]; then
    echo -e "${RED}❌ .env file not found!${NC}"
    echo -e "${YELLOW}Creating from .env.example...${NC}"
    cp .env.example .env
    echo -e "${YELLOW}⚠️  Please edit .env file with your credentials${NC}"
    exit 1
fi

# Check if Firebase credentials exist
if [ ! -f config/firebase-credentials.json ]; then
    echo -e "${RED}❌ Firebase credentials not found!${NC}"
    echo -e "${YELLOW}Please add config/firebase-credentials.json${NC}"
    exit 1
fi

# Create necessary directories
echo -e "${GREEN}📁 Creating directories...${NC}"
mkdir -p logs
mkdir -p config

# Stop existing containers
echo -e "${GREEN}🛑 Stopping existing containers...${NC}"
docker-compose down || true

# Pull latest code (if git repo)
if [ -d .git ]; then
    echo -e "${GREEN}📥 Pulling latest code...${NC}"
    git pull
fi

# Build Docker image
echo -e "${GREEN}🔨 Building Docker image...${NC}"
docker-compose build --no-cache

# Start services
echo -e "${GREEN}🚀 Starting services...${NC}"
docker-compose up -d

# Wait for service to be ready
echo -e "${GREEN}⏳ Waiting for service to start...${NC}"
sleep 10

# Health check
echo -e "${GREEN}🏥 Running health check...${NC}"
for i in {1..5}; do
    if curl -f http://localhost:8080/health > /dev/null 2>&1; then
        echo -e "${GREEN}✅ Service is healthy!${NC}"
        break
    else
        if [ $i -eq 5 ]; then
            echo -e "${RED}❌ Health check failed!${NC}"
            docker-compose logs --tail=50 crypto-api
            exit 1
        fi
        echo -e "${YELLOW}Attempt $i/5 - Waiting...${NC}"
        sleep 5
    fi
done

# Show running containers
echo -e "${GREEN}📊 Running containers:${NC}"
docker-compose ps

# Show logs
echo -e "${GREEN}📝 Recent logs:${NC}"
docker-compose logs --tail=20 crypto-api

echo -e "${GREEN}✅ Deployment completed successfully!${NC}"
echo -e "${GREEN}🌐 API is running at: http://localhost:8080${NC}"
echo -e "${GREEN}📊 Health check: http://localhost:8080/health${NC}"
echo -e "${YELLOW}📋 View logs: docker-compose logs -f crypto-api${NC}"
