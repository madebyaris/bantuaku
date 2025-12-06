# Go 1.25 Upgrade Feature Brief

## 🎯 Context (2min)
**Problem**: Version mismatch between production Dockerfile (Go 1.25) and development environment (Go 1.22), causing inconsistency and preventing use of latest Air hot reload tool (requires Go 1.25+)

**Users**: Backend developers working on the Bantuaku platform

**Success**: 
- All environments use Go 1.25 consistently
- Air hot reload upgraded to latest version
- All builds pass successfully
- No breaking changes introduced

## 🔍 Quick Research (15min)

### Current State Analysis
- **go.mod**: Specifies `go 1.22`
- **Dockerfile (production)**: Already uses `golang:1.25-alpine` ✅
- **Dockerfile.dev (development)**: Uses `golang:1.22-alpine` ❌
- **GitHub Actions**: Security workflow already uses Go 1.25 ✅
- **Air version**: Pinned to v1.49.0 for Go 1.22 compatibility (can upgrade to latest)

### Dependency Compatibility
All major dependencies are modern and should support Go 1.25:
- `github.com/jackc/pgx/v5` v5.7.1 - Modern, supports Go 1.25
- `github.com/redis/go-redis/v9` v9.7.0 - Modern, supports Go 1.25
- `golang.org/x/*` packages - All recent versions, compatible
- `github.com/chromedp/chromedp` v0.9.5 - Should be compatible

### Go 1.25 Compatibility Notes
- **Go 1 Compatibility Promise**: Go 1.25 maintains backward compatibility - existing code will compile without changes
- **No Language Changes**: No breaking syntax changes affecting our codebase
- **macOS Requirement**: Requires macOS 12+ (not relevant for Docker/Linux builds)
- **Windows/ARM**: Final release for 32-bit Windows/ARM (not relevant for our Linux containers)

### Tech Decision
**Approach**: Upgrade all Go version references to 1.25, upgrade Air to latest

**Why**: 
- Consistency across environments
- Access to latest Air features and bug fixes
- Production already uses 1.25, so dev should match
- Go 1.25 is stable and maintains compatibility

**Avoid**: 
- Keeping version mismatch (causes confusion)
- Skipping dependency verification (could miss compatibility issues)

## ✅ Requirements (10min)

### Must-Have
1. **Update go.mod** → Change `go 1.22` to `go 1.25`
2. **Update Dockerfile.dev** → Change `golang:1.22-alpine` to `golang:1.25-alpine`
3. **Upgrade Air** → Update from `v1.49.0` to `@latest` in both Dockerfile.dev and Makefile
4. **Verify Dependencies** → Run `go mod tidy` and ensure no errors
5. **Test Builds** → Verify both Docker and local builds succeed

### Nice-to-Have
- Update any documentation mentioning Go version
- Verify CI/CD workflows are consistent (already done - security.yml uses 1.25)

## 🏗️ Implementation (5min)

### Components to Update
1. **backend/go.mod** - Version declaration
2. **backend/Dockerfile.dev** - Base image and Air version
3. **Makefile** - Air installation version for local dev
4. **Verification** - Run `go mod tidy` and test builds

### APIs
N/A - No API changes

### Data
N/A - No database changes

## 📋 Next Actions (2min)

- [ ] Update `backend/go.mod`: Change `go 1.22` → `go 1.25` (2 min)
- [ ] Update `backend/Dockerfile.dev`: Change base image to `golang:1.25-alpine` and Air to `@latest` (3 min)
- [ ] Update `Makefile`: Change Air version to `@latest` (1 min)
- [ ] Run `go mod tidy` in backend directory (1 min)
- [ ] Test Docker build: `make dev` (5 min)
- [ ] Test local build: `make dev-backend` (2 min)
- [ ] Verify no breaking changes in functionality (5 min)

**Start Coding In**: Immediately after brief approval

---
**Total Planning Time**: ~30min | **Owner**: Backend Team | **Date**: 2025-12-05

<!-- Living Document - Update as you code -->

## 🔄 Implementation Tracking

**CRITICAL**: Follow the todo-list systematically. Mark items as complete, document blockers, update progress.

### Progress
- [x] Update `backend/go.mod`: Changed `go 1.22` → `go 1.25` ✅
- [x] Update `backend/Dockerfile.dev`: Changed base image to `golang:1.25-alpine` and Air to `@latest` ✅
- [x] Update `Makefile`: Changed Air version to `@latest` ✅
- [x] Fix Air repository migration: Updated from `github.com/cosmtrek/air` to `github.com/air-verse/air` ✅
- [x] Run `go mod tidy` in backend directory ✅
- [ ] Test Docker build: `make dev` (pending user test)
- [ ] Test local build: `make dev-backend` (pending user test)
- [ ] Verify no breaking changes in functionality (pending user test)

### Blockers
- ~~Air repository migration issue~~ - Fixed: Updated to new repository `github.com/air-verse/air`

**See**: [.sdd/IMPLEMENTATION_GUIDE.md](mdc:.sdd/IMPLEMENTATION_GUIDE.md) for detailed execution rules.

## 📝 Research Notes

### Files Requiring Updates
1. `backend/go.mod` - Line 3: `go 1.22` → `go 1.25`
2. `backend/Dockerfile.dev` - Line 2: `golang:1.22-alpine` → `golang:1.25-alpine`
3. `backend/Dockerfile.dev` - Line 9: `air@v1.49.0` → `air@latest`
4. `Makefile` - Line 28: `air@v1.49.0` → `air@latest`

### Verification Steps
1. `cd backend && go mod tidy` - Should complete without errors
2. `cd backend && go build .` - Should build successfully
3. `make dev` - Docker build should succeed
4. `make dev-backend` - Local build with Air should work

### Risk Assessment
- **Low Risk**: Go 1.25 maintains compatibility promise
- **Dependencies**: All modern, should be compatible
- **Testing**: Quick verification builds will catch any issues
