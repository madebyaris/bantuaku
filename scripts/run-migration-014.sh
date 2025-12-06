#!/bin/bash
# Run migration 014_add_settings_table.sql manually

set -e

echo "Running migration 014_add_settings_table.sql..."

# Try docker compose first, fallback to docker-compose
if command -v docker-compose &> /dev/null; then
    DOCKER_COMPOSE="docker-compose"
else
    DOCKER_COMPOSE="docker compose"
fi

# Run the migration
$DOCKER_COMPOSE exec -T db psql -U bantuaku -d bantuaku_dev -f /docker-entrypoint-initdb.d/014_add_settings_table.sql

echo "Migration completed successfully!"
