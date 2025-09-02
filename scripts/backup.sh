#!/bin/bash

# Backup script for Fund Rebalance Application
# Usage: ./scripts/backup.sh

set -e

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
BACKUP_DIR="/opt/fund-rebalance/backups"
APP_DATA_DIR="/opt/fund-rebalance/data"
TIMESTAMP=$(date +"%Y%m%d_%H%M%S")
BACKUP_NAME="fund-rebalance-backup-$TIMESTAMP"

echo -e "${GREEN}💾 Starting backup process${NC}"
echo -e "${YELLOW}Backup name: $BACKUP_NAME${NC}"

# Create backup directory
mkdir -p "$BACKUP_DIR"

# Create backup archive
echo -e "${YELLOW}📦 Creating backup archive...${NC}"
tar -czf "$BACKUP_DIR/$BACKUP_NAME.tar.gz" \
    -C "/opt/fund-rebalance" \
    data \
    logs \
    docker-compose.prod.yml \
    nginx

# Verify backup
if [ -f "$BACKUP_DIR/$BACKUP_NAME.tar.gz" ]; then
    BACKUP_SIZE=$(du -sh "$BACKUP_DIR/$BACKUP_NAME.tar.gz" | cut -f1)
    echo -e "${GREEN}✅ Backup created successfully!${NC}"
    echo -e "${BLUE}📄 File: $BACKUP_DIR/$BACKUP_NAME.tar.gz${NC}"
    echo -e "${BLUE}📊 Size: $BACKUP_SIZE${NC}"
else
    echo -e "${RED}❌ Backup creation failed!${NC}"
    exit 1
fi

# Clean up old backups (keep last 7 days)
echo -e "${YELLOW}🧹 Cleaning up old backups...${NC}"
find "$BACKUP_DIR" -name "fund-rebalance-backup-*.tar.gz" -mtime +7 -delete

# List current backups
echo -e "${BLUE}📋 Current backups:${NC}"
ls -la "$BACKUP_DIR"/fund-rebalance-backup-*.tar.gz 2>/dev/null || echo "No backups found"

echo -e "${GREEN}🎉 Backup process completed!${NC}"