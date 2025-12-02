# Docker Troubleshooting Guide

## PostgreSQL Volume Mount Error

If you encounter this error after upgrading PostgreSQL:
```
error mounting "/var/lib/docker/volumes/bantuaku_db_data/_data" to rootfs at "/var/lib/postgresql/data"
```

### Solution

This happens when upgrading PostgreSQL versions (e.g., 16 → 18) because the old volume format may be incompatible.

**Option 1: Clean and recreate (recommended for development)**
```bash
# Stop containers and remove volumes
make clean-db

# Or manually:
docker compose down -v
docker volume rm bantuaku_db_data 2>/dev/null || true

# Start fresh
make dev
```

**Option 2: Keep data (if you have important data)**
```bash
# Backup first
docker compose exec db pg_dump -U bantuaku bantuaku_dev > backup.sql

# Remove volume
docker compose down -v
docker volume rm bantuaku_db_data

# Start fresh
make dev

# Restore data
docker compose exec -T db psql -U bantuaku bantuaku_dev < backup.sql
```

## Frontend Hot Reload

The frontend is configured for hot reload in Docker with:
- Volume mount: `./frontend:/app` (excludes node_modules)
- Vite watch mode with polling enabled
- `CHOKIDAR_USEPOLLING=true` environment variable

### If hot reload doesn't work:

1. **Check file permissions**:
   ```bash
   ls -la frontend/
   ```

2. **Restart frontend container**:
   ```bash
   docker compose restart frontend
   ```

3. **Check logs**:
   ```bash
   make logs-frontend
   ```

4. **Verify Vite is watching**:
   Look for `VITE` messages in logs showing file changes

### Manual Development (without Docker)

For faster hot reload during development, run frontend locally:
```bash
make dev-frontend
```

This runs `npm run dev` directly on your machine (faster file watching).

## Common Issues

### Port Already in Use
```bash
# Find process using port
lsof -i :3000  # Frontend
lsof -i :8080  # Backend
lsof -i :5432  # PostgreSQL

# Kill process
kill -9 <PID>
```

### Container Won't Start
```bash
# Check logs
docker compose logs <service-name>

# Rebuild without cache
docker compose build --no-cache

# Full reset
make reset
```

### Database Connection Issues
```bash
# Check if database is healthy
docker compose ps

# Check database logs
docker compose logs db

# Test connection manually
docker compose exec db psql -U bantuaku -d bantuaku_dev
```
