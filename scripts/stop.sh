#!/bin/bash

# Stop script for Fund Rebalance Application
# Usage: ./scripts/stop.sh

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${YELLOW}🛑 Stopping Fund Rebalance Application${NC}"

# Check if docker-compose file exists
if [ ! -f "docker-compose.prod.yml" ]; then
    echo -e "${RED}❌ docker-compose.prod.yml not found!${NC}"
    exit 1
fi

# Stop services gracefully
echo -e "${YELLOW}🐳 Stopping Docker containers...${NC}"
docker-compose -f docker-compose.prod.yml down --timeout 30

# Check if any containers are still running
RUNNING_CONTAINERS=$(docker ps -q --filter "label=com.docker.compose.service=fund-rebalance")
if [ -n "$RUNNING_CONTAINERS" ]; then
    echo -e "${YELLOW}🔄 Force stopping remaining containers...${NC}"
    docker stop $RUNNING_CONTAINERS
fi

# Clean up unused images (optional)
read -p "Do you want to clean up unused Docker images? (y/N): " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
    echo -e "${YELLOW}🧹 Cleaning up unused images...${NC}"
    docker image prune -f
fi

echo -e "${GREEN}✅ Application stopped successfully!${NC}"