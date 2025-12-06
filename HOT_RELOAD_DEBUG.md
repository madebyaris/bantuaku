# Hot Reload Debug Guide

## Issue: 500 errors persist, hot reload may not be working

### Step 1: Rebuild Docker Container
The Air config was updated to use `-mod=mod` flag. You need to rebuild:

```bash
# Stop current containers
make down

# Rebuild and start with new config
make dev
```

### Step 2: Check if Air is Running
Once containers are up, check backend logs:

```bash
# View backend logs
make logs-backend

# Or directly
docker-compose logs -f backend
```

**Look for:**
- `[Air]` messages showing file watching
- `[Air] Building...` when you save a file
- `[Air] Running...` after build completes
- Any build errors in the logs

### Step 3: Check Actual 500 Error
The backend logs will show the actual error causing the 500:

```bash
docker-compose logs backend | grep -i error
```

**Common issues:**
- Database connection errors
- Missing environment variables
- Panic in handler code
- SQL query errors

### Step 4: Test Hot Reload
1. Make a small change to `backend/handlers/chat.go` (add a comment)
2. Save the file
3. Watch backend logs - you should see Air rebuilding
4. The server should restart automatically

### Step 5: Verify Changes Applied
After Air rebuilds, check if your changes are live by:
- Making a test API call
- Checking the response matches your code changes

## If Hot Reload Still Not Working

### Option A: Force Rebuild
```bash
docker-compose build --no-cache backend
docker-compose up backend
```

### Option B: Check Volume Mounts
Verify in `docker-compose.yml` that volumes are mounted:
```yaml
volumes:
  - ./backend:/app  # This should mount your code
```

### Option C: Manual Restart
If Air isn't detecting changes, manually restart:
```bash
docker-compose restart backend
```

## Current Fixes Applied
1. ✅ Updated Air config to use `-mod=mod` (ignores vendor issues)
2. ✅ Fixed GetConversations query to properly get last_message_at
3. ✅ Added database connection checks
4. ✅ Improved error logging

## Next Steps
1. Rebuild containers: `make down && make dev`
2. Check logs: `make logs-backend`
3. Look for the actual error message in logs
4. Share the error message if it persists
