# media-backup-system Implementation Tasks

**Last Updated**: 2026-03-24T17:45:00Z
**Match Rate**: 0% (no implementation)
**Priority**: High (user requested)

---

## PDCA Phase Status

- ✅ **Plan Phase** - Complete (2026-03-24)
- ✅ **Design Phase** - Complete (2026-03-24)
- 🔄 **Do Phase** - In Progress (0% implementation)
- ✅ **Check Phase** - Complete (gap analysis: 0%)
- ⏳ **Act Phase** - Blocked (no code to iterate)

---

## Phase 1: Backend Foundation (Week 1)

### Database Migrations
- [ ] Create `000003_create_media_assets_table.up.sql`
  - Table: media_assets (asset_id, user_id, youtube_video_id, s3_object_key, etc.)
  - Indexes: user_id, sync_status, created_at

- [ ] Create `000003_create_media_assets_table.down.sql`
  - Drop table and indexes

- [ ] Create `000004_create_upload_sessions_table.up.sql`
  - Table: upload_sessions (session_id, user_id, total_files, completed_files, etc.)
  - Indexes: user_id, session_status

- [ ] Create `000004_create_upload_sessions_table.down.sql`
  - Drop table and indexes

- [ ] Run migrations
  ```bash
  migrate -database $DATABASE_URL -path migrations up
  ```

### Domain Entities
- [ ] Create `internal/domain/media_asset.go`
  - MediaAsset struct with GORM tags
  - MediaType enum (VIDEO, IMAGE)
  - SyncStatus enum (PENDING, UPLOADING, COMPLETED, FAILED)
  - TableName() method
  - Validation methods

- [ ] Create `internal/domain/upload_session.go`
  - UploadSession struct with GORM tags
  - SessionStatus enum (ACTIVE, COMPLETED, CANCELLED)
  - Progress calculation methods
  - TableName() method

### Repository Layer
- [ ] Update `internal/repository/interfaces.go`
  - Add MediaRepository interface (Create, FindByID, FindByUserID, Update, Delete, FindPendingUploads)
  - Add UploadSessionRepository interface

- [ ] Create `internal/repository/media_repository.go`
  - Implement MediaRepository with GORM
  - Handle UUID parsing and error mapping
  - Pagination support for FindByUserID

- [ ] Create `internal/repository/session_repository.go`
  - Implement UploadSessionRepository
  - Methods for session lifecycle management

### Initialize in main.go
- [ ] Update `cmd/api/main.go`
  - Initialize media and session repositories
  - Wire up with existing database connection
  - Add to dependency injection chain

---

## Phase 2: YouTube Integration (Week 2)

### YouTube Service Extension
- [ ] Extend `internal/service/youtube_service.go`
  - Add ResumableUpload method
  - Implement chunked upload logic
  - Add video verification (check if playable)
  - Handle YouTube API quota errors

### Upload Service
- [ ] Create `internal/service/upload_service.go`
  - Implement UploadService interface
  - InitiateSession method
  - UploadVideo method (orchestrates YouTube upload + metadata save)
  - GetUploadStatus method
  - CompleteSession method
  - Retry logic with exponential backoff

### Media Handler
- [ ] Create `internal/handler/media_handler.go`
  - InitiateUpload handler (POST /api/v1/media/upload/initiate)
  - UploadVideo handler (POST /api/v1/media/upload/video)
  - GetUploadStatus handler (GET /api/v1/media/upload/status/:session_id)
  - ListMedia handler (GET /api/v1/media/list)
  - GetMediaDetails handler (GET /api/v1/media/:asset_id)
  - DeleteMedia handler (DELETE /api/v1/media/:asset_id)

- [ ] Update `internal/handler/dto.go`
  - Add MediaUploadRequest DTO
  - Add UploadStatusResponse DTO
  - Add MediaListResponse DTO with pagination

### Router Updates
- [ ] Update `internal/router/router.go`
  - Add media routes to protected group
  - Wire up media handler
  - Configure multipart form size limits

---

## Phase 3: Mobile Foundation (Week 3)

### Flutter Project Setup
- [ ] Create Flutter project
  ```bash
  flutter create mobile_backup_app
  cd mobile_backup_app
  ```

- [ ] Update `pubspec.yaml`
  - Add provider (state management)
  - Add dio (HTTP client)
  - Add sqflite (local database)
  - Add workmanager (background processing)
  - Add google_sign_in (OAuth)
  - Add permission_handler
  - Add path_provider

### Core Services
- [ ] Create `lib/services/api_client.dart`
  - Dio configuration
  - Base URL setup
  - Auth token interceptor
  - Methods: initiateUpload, uploadVideo, getUploadStatus, listMedia

- [ ] Create `lib/services/database_helper.dart`
  - SQLite database initialization
  - Table: local_media_assets
  - CRUD operations
  - Migration support

### Models
- [ ] Create `lib/models/local_media_asset.dart`
  - LocalMediaAsset class
  - fromMap / toMap methods
  - Enum: MediaType, SyncStatus

- [ ] Create `lib/models/user.dart`
  - User model for authentication state

### Authentication
- [ ] Create `lib/providers/auth_provider.dart`
  - GoogleSignIn integration
  - Token management
  - Login/logout flows
  - Token refresh logic

---

## Phase 4: File Detection & Upload (Week 4)

### File System Access
- [ ] Create `lib/services/permission_service.dart`
  - Request storage permissions
  - Request notification permissions
  - Check permission status

- [ ] Create `lib/services/file_service.dart`
  - Detect new video files
  - Monitor DCIM/Camera directory
  - Get file metadata (size, name, date)
  - Delete local files safely

### Upload Management
- [ ] Create `lib/providers/upload_provider.dart`
  - Upload queue management
  - Progress tracking (0-100%)
  - Add to queue, remove from queue
  - Upload state (idle, uploading, paused)

- [ ] Implement upload flow
  - File detection → Add to SQLite queue
  - Check network (WiFi only)
  - Initiate upload session via API
  - Upload file with progress callbacks
  - Update local database on completion
  - Delete local file after verification

### UI Screens
- [ ] Create `lib/screens/dashboard_screen.dart`
  - Storage status widget
  - Sync status widget
  - Recent backups list
  - Upload progress indicator

- [ ] Create `lib/screens/history_screen.dart`
  - List backed up media with pagination
  - Filters (videos, images, all)
  - Sort options
  - Media detail navigation

- [ ] Create `lib/screens/settings_screen.dart`
  - WiFi only toggle
  - Auto-delete original toggle
  - Upload schedule settings

---

## Phase 5: Background Sync (Week 5)

### WorkManager Setup
- [ ] Create `lib/services/background_service.dart`
  - Initialize WorkManager
  - Register periodic upload task (15min intervals)
  - Configure constraints (WiFi, battery)
  - Callback dispatcher implementation

- [ ] Implement background upload logic
  - Check network status
  - Fetch pending uploads from SQLite
  - Upload each file sequentially
  - Update sync status
  - Show notifications for progress/completion

### Network Monitoring
- [ ] Add connectivity package
- [ ] Implement network change listener
- [ ] Pause/resume uploads based on network
- [ ] Queue failed uploads for retry

### Notification System
- [ ] Create `lib/services/notification_service.dart`
  - Show upload progress notification
  - Show completion notification
  - Show error notification
  - Handle notification tap actions

---

## Phase 6: Testing & Polish (Week 6)

### Backend Tests
- [ ] Unit tests for repositories
  - Test CRUD operations
  - Test error handling
  - Test pagination

- [ ] Unit tests for services
  - Test upload orchestration
  - Test retry logic
  - Test YouTube API integration (mocked)

- [ ] Integration tests
  - Test full upload flow
  - Test error scenarios
  - Test concurrent uploads

### Mobile Tests
- [ ] Widget tests
  - Dashboard screen
  - History screen
  - Settings screen

- [ ] Provider tests
  - Auth provider
  - Upload provider

- [ ] Integration tests
  - Login flow
  - Upload flow
  - Background service

### Documentation
- [ ] Update backend README
  - New API endpoints
  - Environment variables
  - Migration instructions

- [ ] Create mobile app README
  - Setup instructions
  - Build instructions
  - Configuration guide

- [ ] API documentation (Swagger)
  - Generate OpenAPI spec
  - Add example requests/responses

---

## Deferred Tasks (Phase 2 Features)

### Image Backup (S3)
- [ ] Implement S3Service
- [ ] Add image upload endpoint
- [ ] Mobile: image selection and upload

### Advanced Features
- [ ] Restore functionality (download from YouTube)
- [ ] Scheduled uploads (time-based)
- [ ] Compression options
- [ ] Batch operations

---

## Testing Checklist

### Backend Tests to Run
```bash
cd backend
go test ./internal/domain -v
go test ./internal/repository -v
go test ./internal/service -v
go test ./internal/handler -v
go test ./... -cover
# Target: >80% coverage
```

### Mobile Tests to Run
```bash
cd mobile_backup_app
flutter test
flutter test --coverage
# Target: >70% coverage
```

### Manual Testing
- [ ] Login flow
- [ ] Small video upload (<10MB)
- [ ] Large video upload (>100MB)
- [ ] Network interruption during upload
- [ ] Background upload while app closed
- [ ] Upload retry after failure
- [ ] Local file deletion after successful upload

---

## Known Issues & Risks

### High Risk
1. **YouTube API Quota**: Daily upload limit (10,000 quota units per day)
2. **Large File Upload**: Network stability for multi-GB files
3. **Background Service Reliability**: Android/iOS background restrictions

### Medium Risk
1. **Storage Permission**: Changes in Android 11+ storage access
2. **Battery Optimization**: May kill background service
3. **Network Detection**: Unreliable on some devices

### Mitigation
- Implement chunked resumable uploads
- Add comprehensive retry logic
- User education about permissions
- Fallback to foreground service if needed

---

## Resources & References

### Documentation
- Plan: `docs/01-plan/features/media-backup-system.plan.md`
- Design: `docs/02-design/features/media-backup-system.design.md`
- Gap Analysis: `docs/03-analysis/media-backup-system.analysis.md`

### External APIs
- YouTube Data API v3: https://developers.google.com/youtube/v3
- Oracle Cloud S3: https://docs.oracle.com/en-us/iaas/Content/Object/home.htm
- Google OAuth 2.0: https://developers.google.com/identity/protocols/oauth2

### Libraries
- Go: gin, gorm, youtube/v3, aws-sdk-go-v2
- Flutter: provider, dio, sqflite, workmanager, google_sign_in

---

## Current Blocker

**Issue**: User requested `/pdca iterate` but no implementation exists (0% match rate).

**Resolution**: The iterate command is designed for auto-fixing existing code, not creating from scratch. Must start manual implementation following the phase-by-phase guide above.

**Next Action**: Confirm with user to start Phase 1 Backend Foundation implementation.
