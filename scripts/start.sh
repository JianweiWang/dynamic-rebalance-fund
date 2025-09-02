#!/bin/bash

# Start script for Fund Rebalance Application
# Usage: ./scripts/start.sh

set -e

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${GREEN}🚀 Starting Fund Rebalance Application${NC}"

# Create necessary directories
echo -e "${YELLOW}📁 Creating directories...${NC}"
mkdir -p data logs logs/nginx

# Set proper permissions
chmod -R 755 data logs

# Start services
echo -e "${YELLOW}🐳 Starting Docker containers...${NC}"
docker-compose -f docker-compose.prod.yml up -d

# Wait for services to be ready
echo -e "${YELLOW}⏳ Waiting for services to be ready...${NC}"
sleep 15

# Check service status
echo -e "${YELLOW}📊 Checking service status...${NC}"
docker-compose -f docker-compose.prod.yml ps

# Verify services are running
if docker-compose -f docker-compose.prod.yml ps | grep -q "Up"; then
    echo -e "${GREEN}✅ All services are running!${NC}"
    
    # Show logs for the last few seconds
    echo -e "${YELLOW}📝 Recent logs:${NC}"
    docker-compose -f docker-compose.prod.yml logs --tail=20 fund-rebalance
else
    echo -e "${RED}❌ Some services failed to start!${NC}"
    docker-compose -f docker-compose.prod.yml logs
    exit 1
fi

echo -e "${GREEN}🎉 Application started successfully!${NC}"