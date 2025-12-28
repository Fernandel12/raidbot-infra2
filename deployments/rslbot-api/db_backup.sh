#!/bin/bash

# Database backup script for RaidBot
# Usage: ./db_backup.sh

BACKUP_DIR="./backups"
TIMESTAMP=$(date +"%Y%m%d_%H%M%S")
BACKUP_FILE="${BACKUP_DIR}/raidbot_${TIMESTAMP}.sql"

# Create backup directory if it doesn't exist
mkdir -p "$BACKUP_DIR"

# Perform backup
echo "Creating backup: $BACKUP_FILE"
docker compose exec -T mysql mysqldump -u rslbot -p"${MYSQL_PASSWORD}" rslbot > "$BACKUP_FILE"

if [ $? -eq 0 ]; then
    echo "Backup completed successfully: $BACKUP_FILE"

    # Compress backup
    gzip "$BACKUP_FILE"
    echo "Compressed: ${BACKUP_FILE}.gz"

    # Keep only last 30 days of backups
    find "$BACKUP_DIR" -name "raidbot_*.sql.gz" -mtime +30 -delete
    echo "Old backups cleaned up"
else
    echo "Backup failed!"
    exit 1
fi
