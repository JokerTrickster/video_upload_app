# Video Upload App - Project Context

**Last Updated**: 2026-03-25
**Status**: MVP Complete (Backend + Flutter App)
**No uncommitted changes** — all work pushed to `origin/main`

---

## Architecture

```
video_upload_app/
├── backend/          # Go 1.26 + Gin + GORM + PostgreSQL + Redis
│   ├── cmd/api/      # Entry point (main.go)
│   ├── internal/
│   │   ├── domain/       # Models: User, Token, MediaAsset, UploadSession, UploadQueueItem, DailyQuota
│   │   ├── repository/   # GORM repos: user, token, media, session, queue
│   │   ├── service/      # Business logic: auth, token, upload, youtube, queue, scheduler
│   │   ├── handler/      # HTTP handlers: auth, media, queue
│   │   ├── middleware/    # auth JWT, CORS, rate limiter, error handler
│   │   ├── router/       # Gin route setup
│   │   ├── config/       # Env-based config
│   │   └── pkg/          # database, redis, logger, youtube client
│   └── Dockerfile
├── app/              # Flutter 3.27.4 (iOS + Android)
│   └── lib/
│       ├── core/         # ApiClient (Dio + JWT auto-inject), router, responsive util
│       ├── features/     # auth, media, upload, queue (data + presentation)
│       └── shared/       # models, widgets (LoadingOverlay, error_snackbar)
├── docker-compose.yml
└── .env.example
```

## Backend API Endpoints (16 total)

### Auth (5)
- GET  /api/v1/auth/google/url
- POST /api/v1/auth/google/callback
- POST /api/v1/auth/refresh
- GET  /api/v1/auth/me
- POST /api/v1/auth/logout

### Media (8)
- POST /api/v1/media/upload/initiate
- POST /api/v1/media/upload/video
- GET  /api/v1/media/upload/status/:session_id
- POST /api/v1/media/upload/complete
- POST /api/v1/media/upload/cancel
- GET  /api/v1/media/list
- GET  /api/v1/media/:asset_id
- DELETE /api/v1/media/:asset_id

### Queue - Auto Upload (4) — NEW this session
- POST /api/v1/queue/add
- GET  /api/v1/queue
- DELETE /api/v1/queue/:queue_id
- GET  /api/v1/queue/quota

## Key Decisions Made This Session

1. **Broken test mocks fixed**: upload_service_test.go and media_handler_test.go had mock interfaces that didn't match actual repo interfaces (wrong method names: GetByID vs FindByID, wrong param types: uuid.UUID vs string). Completely rewrote both.

2. **OAuth token flow fixed**: media_handler.go was trying `c.Get("access_token")` which was never set. Changed to: tokenRepo.FindByUserID → token.IsExpired check → tokenService.DecryptToken → use decrypted token.

3. **Retry backoff values**: Plan said 1s→30s, implementation had 1min→24h. Changed to match Plan: 1s, 2s, 5s, 15s, 30s.

4. **Soft delete**: Added DeletedAt field to MediaAsset for GORM soft delete support.

5. **Repository filtering**: FindByUserID updated to accept mediaType, syncStatus, sort params with proper GORM Where/Order clauses.

6. **Auto upload queue**: Scheduler runs hourly, processes PENDING items within YouTube API quota (10,000 units/day = 6 uploads max). Failed items return to PENDING for next-day retry.

7. **Flutter responsive**: Galaxy S22+ (412x915 dp) as design reference. All sizes use proportional scaling with clamp() for edge cases.

8. **Flutter SDK location**: ~/development/flutter/bin (added to ~/.zshrc)

## Test Coverage

- Backend: 260+ tests ALL PASS (go test ./internal/...)
- Flutter: 41 tests ALL PASS (flutter test)
- PDCA Match Rate: 93% (after iteration 1)

## Pending / Future Work

### S3 Image Upload (Blocked — awaiting credentials)
- Domain model has S3ObjectKey field ready
- Plan marks as "Phase 3 Out of Scope"
- User will provide AWS bucket info later

### Backend Optional Improvements
- Integration tests (placeholder — needs real DB)
- YouTube client test coverage (47.9% — needs mock HTTP server)
- YouTube delete implementation (deleteFromYouTube=true needs access token chain)
- Graceful shutdown for in-progress uploads

### Flutter Optional
- Deep link for OAuth callback (currently opens external browser)
- Dark mode theme
- Push notifications for upload completion
- Offline queue support

## How to Run

```bash
# Backend (Docker)
cp .env.example .env  # Fill in GOOGLE_CLIENT_ID/SECRET
docker compose up

# Backend (local dev)
cd backend && go run ./cmd/api

# Flutter
export PATH="$HOME/development/flutter/bin:$PATH"
cd app && flutter run

# Tests
cd backend && go test ./internal/... -count=1
cd app && flutter test
```
