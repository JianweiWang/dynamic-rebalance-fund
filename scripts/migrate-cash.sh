#!/bin/bash

# Migration script to update short-term bucket names from "货币基金" to "现金"
# Usage: ./scripts/migrate-cash.sh

set -e

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${GREEN}🔄 Migrating short-term bucket names from 货币基金 to 现金${NC}"

# Check if database exists
DB_PATH="./fund_data.db"
if [ ! -f "$DB_PATH" ]; then
    echo -e "${YELLOW}⚠️  Database file not found: $DB_PATH${NC}"
    echo -e "${BLUE}💡 This migration is only needed for existing installations${NC}"
    exit 0
fi

# Backup database
BACKUP_PATH="./fund_data_backup_$(date +%Y%m%d_%H%M%S).db"
echo -e "${YELLOW}📦 Creating backup: $BACKUP_PATH${NC}"
cp "$DB_PATH" "$BACKUP_PATH"

# Update bucket names
echo -e "${YELLOW}🔄 Updating bucket names...${NC}"
sqlite3 "$DB_PATH" << 'EOF'
-- Update short-term bucket names
UPDATE buckets 
SET name = '短期桶（现金）', updated_at = CURRENT_TIMESTAMP 
WHERE name LIKE '%货币基金%';

-- Update fund names (optional - only if they are generic money market fund names)
UPDATE funds 
SET name = '现金管理', code = 'CASH001', updated_at = CURRENT_TIMESTAMP 
WHERE name = '易方达货币A' AND code = '000009';

-- Show updated results
SELECT 'Updated buckets:' as info;
SELECT id, name, portfolio_id FROM buckets WHERE name LIKE '%现金%';

SELECT 'Updated funds:' as info;
SELECT id, name, code, bucket_id FROM funds WHERE name = '现金管理';
EOF

echo -e "${GREEN}✅ Migration completed successfully!${NC}"
echo -e "${BLUE}📄 Backup saved as: $BACKUP_PATH${NC}"
echo -e "${YELLOW}🔄 Please restart the application to see the changes${NC}"