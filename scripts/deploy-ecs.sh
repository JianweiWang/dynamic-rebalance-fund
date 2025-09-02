#!/bin/bash

# ECS deployment script for Fund Rebalance Application
# Usage: ./scripts/deploy-ecs.sh [environment] [tag]

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
APP_NAME="fund-rebalance"
REGISTRY="your-registry.com"  # Replace with your Docker registry
DEFAULT_ENV="production"
DEFAULT_TAG="latest"

# Get parameters
ENVIRONMENT=${1:-$DEFAULT_ENV}
TAG=${2:-$DEFAULT_TAG}
IMAGE_NAME="$REGISTRY/$APP_NAME:$TAG"

echo -e "${GREEN}🚀 Deploying Fund Rebalance Application to ECS${NC}"
echo -e "${YELLOW}Environment: $ENVIRONMENT${NC}"
echo -e "${YELLOW}Image: $IMAGE_NAME${NC}"

# Check if required tools are installed
check_dependencies() {
    local deps=("docker" "scp" "ssh")
    for dep in "${deps[@]}"; do
        if ! command -v $dep &> /dev/null; then
            echo -e "${RED}❌ $dep is not installed. Please install it first.${NC}"
            exit 1
        fi
    done
}

# Load environment variables
load_config() {
    local config_file="config/${ENVIRONMENT}.env"
    if [ -f "$config_file" ]; then
        echo -e "${YELLOW}📋 Loading configuration from $config_file${NC}"
        source "$config_file"
    else
        echo -e "${RED}❌ Configuration file $config_file not found!${NC}"
        echo -e "${BLUE}💡 Please create the configuration file with the following variables:${NC}"
        echo "ECS_HOST=your-ecs-server.com"
        echo "ECS_USER=root"
        echo "ECS_SSH_KEY=/path/to/ssh/key"
        echo "APP_PORT=80"
        echo "DOCKER_COMPOSE_FILE=docker-compose.prod.yml"
        exit 1
    fi
}

# Copy files to ECS server
deploy_files() {
    echo -e "${YELLOW}📤 Copying files to ECS server...${NC}"
    
    # Create deployment package
    local temp_dir="/tmp/fund-rebalance-deploy-$(date +%s)"
    mkdir -p "$temp_dir"
    
    # Copy necessary files
    cp docker-compose.yml "$temp_dir/docker-compose.yml"
    cp -r nginx "$temp_dir/"
    cp scripts/start.sh "$temp_dir/"
    cp scripts/stop.sh "$temp_dir/"
    cp scripts/backup.sh "$temp_dir/"
    
    # Create environment-specific docker-compose
    sed "s|latest|$TAG|g" docker-compose.yml > "$temp_dir/docker-compose.prod.yml"
    
    # Upload to server
    scp -i "$ECS_SSH_KEY" -r "$temp_dir"/* "$ECS_USER@$ECS_HOST:/opt/fund-rebalance/"
    
    # Cleanup
    rm -rf "$temp_dir"
}

# Execute deployment on ECS server
execute_deployment() {
    echo -e "${YELLOW}🚀 Executing deployment on ECS server...${NC}"
    
    ssh -i "$ECS_SSH_KEY" "$ECS_USER@$ECS_HOST" << EOF
set -e

# Navigate to application directory
cd /opt/fund-rebalance

# Make scripts executable
chmod +x *.sh

# Pull latest images
docker-compose -f docker-compose.prod.yml pull

# Stop existing containers
./stop.sh

# Start new containers
./start.sh

# Verify deployment
sleep 10
if docker-compose -f docker-compose.prod.yml ps | grep -q "Up"; then
    echo "✅ Deployment successful!"
    
    # Test health endpoint
    if curl -f http://localhost/health >/dev/null 2>&1; then
        echo "✅ Health check passed!"
    else
        echo "⚠️  Health check failed, but containers are running"
    fi
else
    echo "❌ Deployment failed!"
    exit 1
fi
EOF
}

# Main deployment process
main() {
    echo -e "${BLUE}=== Starting Deployment Process ===${NC}"
    
    check_dependencies
    load_config
    deploy_files
    execute_deployment
    
    echo -e "${GREEN}🎉 Deployment completed successfully!${NC}"
    echo -e "${BLUE}🔗 Application URL: http://$ECS_HOST${NC}"
}

# Run main function
main