#!/bin/bash

# Build script for Fund Rebalance Application
# Usage: ./scripts/build.sh [tag]

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Configuration
APP_NAME="fund-rebalance"
REGISTRY="your-registry.com"  # Replace with your Docker registry
DEFAULT_TAG="latest"

# Get tag from parameter or use default
TAG=${1:-$DEFAULT_TAG}
IMAGE_NAME="$REGISTRY/$APP_NAME:$TAG"

echo -e "${GREEN}🔨 Building Fund Rebalance Application${NC}"
echo -e "${YELLOW}Image: $IMAGE_NAME${NC}"

# Check if Docker is running
if ! docker info >/dev/null 2>&1; then
    echo -e "${RED}❌ Docker is not running. Please start Docker first.${NC}"
    exit 1
fi

# Create necessary directories
echo -e "${YELLOW}📁 Creating directories...${NC}"
mkdir -p data logs logs/nginx

# Build Docker image
echo -e "${YELLOW}🐳 Building Docker image...${NC}"
docker build -t $IMAGE_NAME .

# Verify build
if [ $? -eq 0 ]; then
    echo -e "${GREEN}✅ Build completed successfully!${NC}"
    echo -e "${GREEN}Image: $IMAGE_NAME${NC}"
    
    # Show image size
    echo -e "${YELLOW}📊 Image information:${NC}"
    docker images $IMAGE_NAME
    
    # Optional: Push to registry
    read -p "Do you want to push to registry? (y/N): " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        echo -e "${YELLOW}📤 Pushing to registry...${NC}"
        docker push $IMAGE_NAME
        echo -e "${GREEN}✅ Push completed!${NC}"
    fi
else
    echo -e "${RED}❌ Build failed!${NC}"
    exit 1
fi

echo -e "${GREEN}🎉 All done!${NC}"