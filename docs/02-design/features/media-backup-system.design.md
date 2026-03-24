# media-backup-system Design Document

> **Summary**: YouTube-based unlimited media auto-backup system design with Flutter mobile app and Go backend
>
> **Project**: video_upload_app
> **Version**: 1.0.0
> **Author**: Development Team
> **Date**: 2026-03-24
> **Status**: Draft
> **Planning Doc**: [media-backup-system.plan.md](../../01-plan/features/media-backup-system.plan.md)

### Pipeline References

| Phase | Document | Status |
|-------|----------|--------|
| Phase 1 | Schema Definition | ✅ (in Plan) |
| Phase 2 | Coding Conventions | N/A |
| Phase 3 | Mockup | N/A |
| Phase 4 | API Spec | ✅ (below) |

---

## 1. Overview

### 1.1 Design Goals

- **Automated Background Sync**: Zero user intervention for media backup
- **Scalable Architecture**: Handle thousands of concurrent uploads
- **Resilient Upload**: Automatic retry with exponential backoff
- **Storage Optimization**: 90%+ disk space recovery after backup
- **Cross-Platform**: Flutter for iOS and Android with single codebase

### 1.2 Design Principles

- **Clean Architecture**: Clear separation between Domain, Application, and Infrastructure layers
- **Dependency Inversion**: High-level policies don't depend on low-level details
- **Single Responsibility**: Each component has one reason to change
- **Idempotency**: Upload operations are safe to retry without duplication
- **Fail-Safe**: Never delete original file without verified backup

---

## 2. Architecture

### 2.1 Component Diagram

```
┌──────────────────────────────────────────────────────────────┐
│                    Flutter Mobile App                         │
│  ┌─────────────────┐  ┌──────────────┐  ┌─────────────────┐ │
│  │ Presentation    │  │ Application  │  │ Infrastructure  │ │
│  │ - Widgets       │  │ - UseCase    │  │ - SQLite        │ │
│  │ - Screens       │  │ - Services   │  │ - HTTP Client   │ │
│  │ - Provider      │  │ - State Mgmt │  │ - File System   │ │
│  └─────────────────┘  └──────────────┘  └─────────────────┘ │
│           │                   │                   │          │
│  ┌──────────────────────────────────────────────────────────┐│
│  │  WorkManager Background Service                          ││
│  │  - File Watcher  - Upload Queue  - Network Monitor      ││
│  └──────────────────────────────────────────────────────────┘│
└──────────────────────────────────────────────────────────────┘
                              │ HTTPS/REST API
                              ▼
┌──────────────────────────────────────────────────────────────┐
│                      Go Backend API                           │
│  ┌─────────────────┐  ┌──────────────┐  ┌─────────────────┐ │
│  │ Handler         │  │ Service      │  │ Repository      │ │
│  │ (HTTP Layer)    │  │ (Business)   │  │ (Data Access)   │ │
│  └─────────────────┘  └──────────────┘  └─────────────────┘ │
│           │                   │                   │          │
│  ┌──────────────────────────────────────────────────────────┐│
│  │  Middleware: Auth, Logging, Error, Rate Limiter          ││
│  └──────────────────────────────────────────────────────────┘│
└──────────────────────────────────────────────────────────────┘
                │                              │
                ▼                              ▼
    ┌──────────────────────┐      ┌──────────────────────┐
    │  YouTube Data API v3 │      │  Oracle Cloud S3     │
    │  (Video Storage)     │      │  (Image Storage)     │
    └──────────────────────┘      └──────────────────────┘
                │
                ▼
    ┌──────────────────────┐
    │  PostgreSQL Database │
    │  (Metadata Storage)  │
    └──────────────────────┘
```

### 2.2 Data Flow

#### Video Upload Flow

```
1. File Detection
   FileSystemWatcher → New video detected → Add to local SQLite queue

2. Network Check
   WiFi Monitor → Network available → Trigger upload service

3. Upload Initiation
   Mobile App → POST /api/v1/media/upload/initiate → Backend
   Backend → Create upload_session → Return session_id

4. Video Upload
   Mobile App → POST /api/v1/media/upload/video (multipart) → Backend
   Backend → YouTube API (resumable upload) → Get video_id
   Backend → Save to media_assets table → Return success

5. Verification
   Backend → GET YouTube video status → Verify playable
   Backend → Update media_assets.sync_status = COMPLETED

6. Cleanup
   Mobile App → Receive success → Delete local file
   Mobile App → Update local SQLite → Mark as synced
```

#### Error Handling Flow

```
Upload Failure → Log error → Update retry_count → Exponential backoff
  ├─ retry_count < 5 → Add to retry queue → Retry after delay
  └─ retry_count >= 5 → Mark as FAILED → Notify user
```

### 2.3 Dependencies

| Component | Depends On | Purpose |
|-----------|-----------|---------|
| Flutter App | Go Backend API | Media upload and metadata sync |
| Go Backend | YouTube Data API v3 | Video upload and storage |
| Go Backend | PostgreSQL | Metadata persistence |
| Go Backend | Oracle Cloud S3 | Image storage (Phase 2) |
| WorkManager | Android/iOS System | Background job scheduling |

---

## 3. Data Model

### 3.1 Backend Entities (Go)

#### User Entity

```go
// Domain entity
type User struct {
    UserID           uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
    GoogleAccountID  string     `gorm:"type:varchar(255);uniqueIndex;not null"`
    Email            string     `gorm:"type:varchar(255);not null"`
    YouTubeChannelID *string    `gorm:"type:varchar(255)"`
    RefreshToken     *string    `gorm:"type:text"` // YouTube API refresh token (encrypted)
    CreatedAt        time.Time  `gorm:"default:CURRENT_TIMESTAMP"`
    UpdatedAt        time.Time  `gorm:"default:CURRENT_TIMESTAMP"`
    LastLoginAt      *time.Time
}
```

#### MediaAsset Entity

```go
type MediaAsset struct {
    AssetID             uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
    UserID              uuid.UUID  `gorm:"type:uuid;not null;index"`
    YouTubeVideoID      *string    `gorm:"type:varchar(255);uniqueIndex"` // For VIDEO
    S3ObjectKey         *string    `gorm:"type:varchar(512)"`             // For IMAGE
    OriginalFilename    string     `gorm:"type:varchar(512);not null"`
    FileSizeBytes       int64      `gorm:"not null"`
    MediaType           string     `gorm:"type:varchar(10);not null"` // VIDEO or IMAGE
    SyncStatus          string     `gorm:"type:varchar(20);not null;default:PENDING;index"` // PENDING, UPLOADING, COMPLETED, FAILED
    UploadStartedAt     *time.Time
    UploadCompletedAt   *time.Time
    ErrorMessage        *string    `gorm:"type:text"`
    RetryCount          int        `gorm:"default:0"`
    CreatedAt           time.Time  `gorm:"default:CURRENT_TIMESTAMP;index:idx_created_desc"`
    UpdatedAt           time.Time  `gorm:"default:CURRENT_TIMESTAMP"`

    // Relationships
    User User `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
}
```

#### UploadSession Entity

```go
type UploadSession struct {
    SessionID      uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
    UserID         uuid.UUID  `gorm:"type:uuid;not null;index"`
    TotalFiles     int        `gorm:"not null;default:0"`
    CompletedFiles int        `gorm:"not null;default:0"`
    FailedFiles    int        `gorm:"not null;default:0"`
    TotalBytes     int64      `gorm:"not null;default:0"`
    UploadedBytes  int64      `gorm:"not null;default:0"`
    SessionStatus  string     `gorm:"type:varchar(20);not null;default:ACTIVE;index"` // ACTIVE, COMPLETED, CANCELLED
    StartedAt      time.Time  `gorm:"default:CURRENT_TIMESTAMP"`
    CompletedAt    *time.Time

    // Relationships
    User User `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
}
```

### 3.2 Mobile App Entities (Flutter/Dart)

#### LocalMediaAsset Model

```dart
class LocalMediaAsset {
  final String id;               // UUID v4
  final String filePath;         // Local file path
  final String filename;
  final int fileSizeBytes;
  final MediaType mediaType;     // enum: video, image
  final SyncStatus syncStatus;   // enum: pending, uploading, completed, failed
  final String? youtubeVideoId;
  final String? s3ObjectKey;
  final String? errorMessage;
  final int retryCount;
  final DateTime createdAt;
  final DateTime? uploadStartedAt;
  final DateTime? uploadCompletedAt;

  LocalMediaAsset({
    required this.id,
    required this.filePath,
    required this.filename,
    required this.fileSizeBytes,
    required this.mediaType,
    required this.syncStatus,
    this.youtubeVideoId,
    this.s3ObjectKey,
    this.errorMessage,
    this.retryCount = 0,
    required this.createdAt,
    this.uploadStartedAt,
    this.uploadCompletedAt,
  });

  // Factory for SQLite deserialization
  factory LocalMediaAsset.fromMap(Map<String, dynamic> map) { ... }

  // Method for SQLite serialization
  Map<String, dynamic> toMap() { ... }
}
```

### 3.3 Entity Relationships

```
[User] 1 ──── N [MediaAsset]
   │
   └── 1 ──── N [UploadSession]
```

### 3.4 Database Schema (PostgreSQL)

```sql
-- Users table
CREATE TABLE users (
    user_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    google_account_id VARCHAR(255) UNIQUE NOT NULL,
    email VARCHAR(255) NOT NULL,
    youtube_channel_id VARCHAR(255),
    refresh_token TEXT, -- Encrypted with AES-256-GCM
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    last_login_at TIMESTAMP
);

CREATE INDEX idx_users_google_account ON users(google_account_id);

-- Media assets table
CREATE TABLE media_assets (
    asset_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
    youtube_video_id VARCHAR(255) UNIQUE,
    s3_object_key VARCHAR(512),
    original_filename VARCHAR(512) NOT NULL,
    file_size_bytes BIGINT NOT NULL,
    media_type VARCHAR(10) NOT NULL CHECK (media_type IN ('VIDEO', 'IMAGE')),
    sync_status VARCHAR(20) NOT NULL DEFAULT 'PENDING'
        CHECK (sync_status IN ('PENDING', 'UPLOADING', 'COMPLETED', 'FAILED')),
    upload_started_at TIMESTAMP,
    upload_completed_at TIMESTAMP,
    error_message TEXT,
    retry_count INT DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_media_assets_user ON media_assets(user_id);
CREATE INDEX idx_media_assets_status ON media_assets(sync_status);
CREATE INDEX idx_media_assets_created ON media_assets(created_at DESC);

-- Upload sessions table
CREATE TABLE upload_sessions (
    session_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
    total_files INT NOT NULL DEFAULT 0,
    completed_files INT NOT NULL DEFAULT 0,
    failed_files INT NOT NULL DEFAULT 0,
    total_bytes BIGINT NOT NULL DEFAULT 0,
    uploaded_bytes BIGINT NOT NULL DEFAULT 0,
    session_status VARCHAR(20) NOT NULL DEFAULT 'ACTIVE'
        CHECK (session_status IN ('ACTIVE', 'COMPLETED', 'CANCELLED')),
    started_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    completed_at TIMESTAMP
);

CREATE INDEX idx_upload_sessions_user ON upload_sessions(user_id);
CREATE INDEX idx_upload_sessions_status ON upload_sessions(session_status);
```

---

## 4. API Specification

### 4.1 Endpoint List

| Method | Path | Description | Auth |
|--------|------|-------------|------|
| POST | /api/v1/auth/google | Google OAuth login | No |
| POST | /api/v1/auth/refresh | JWT token refresh | No (refresh token) |
| GET | /api/v1/auth/youtube/status | YouTube integration status | Required |
| POST | /api/v1/media/upload/initiate | Start upload session | Required |
| POST | /api/v1/media/upload/video | Upload video file | Required |
| POST | /api/v1/media/upload/image | Upload image file | Required |
| GET | /api/v1/media/upload/status/:session_id | Get upload progress | Required |
| POST | /api/v1/media/upload/complete | Mark session complete | Required |
| GET | /api/v1/media/list | List backed up media | Required |
| GET | /api/v1/media/:asset_id | Get media details | Required |
| DELETE | /api/v1/media/:asset_id | Delete backup record | Required |
| POST | /api/v1/media/:asset_id/restore | Request media restoration | Required |
| GET | /api/v1/settings | Get user settings | Required |
| PUT | /api/v1/settings | Update user settings | Required |

### 4.2 Detailed Specification

#### `POST /api/v1/media/upload/initiate`

**Purpose**: Start a new upload session for batch upload tracking

**Request:**
```json
{
  "total_files": 10,
  "total_bytes": 2147483648,
  "media_types": {
    "video": 5,
    "image": 5
  }
}
```

**Response (201 Created):**
```json
{
  "success": true,
  "data": {
    "session_id": "uuid",
    "started_at": "2026-03-24T10:00:00Z"
  }
}
```

**Error Responses:**
- `401 Unauthorized`: JWT invalid or expired
- `429 Too Many Requests`: Rate limit exceeded

#### `POST /api/v1/media/upload/video`

**Purpose**: Upload video file and store metadata

**Request (multipart/form-data):**
```
Content-Type: multipart/form-data; boundary=----WebKitFormBoundary

------WebKitFormBoundary
Content-Disposition: form-data; name="session_id"

uuid
------WebKitFormBoundary
Content-Disposition: form-data; name="file"; filename="video.mp4"
Content-Type: video/mp4

[binary data]
------WebKitFormBoundary--
```

**Response (201 Created):**
```json
{
  "success": true,
  "message": "Video uploaded successfully",
  "data": {
    "asset_id": "uuid",
    "youtube_video_id": "dQw4w9WgXcQ",
    "original_filename": "video.mp4",
    "file_size_bytes": 104857600,
    "sync_status": "COMPLETED",
    "upload_completed_at": "2026-03-24T10:05:30Z"
  }
}
```

**Error Responses:**
- `400 Bad Request`: Invalid file format or size
- `401 Unauthorized`: Authentication required
- `413 Payload Too Large`: File exceeds max size (2GB)
- `500 Internal Server Error`: YouTube API error
- `503 Service Unavailable`: YouTube API quota exceeded

#### `GET /api/v1/media/list`

**Purpose**: List backed up media with pagination

**Query Parameters:**
- `page`: Page number (default: 1)
- `limit`: Items per page (default: 50, max: 100)
- `media_type`: Filter by VIDEO or IMAGE
- `sync_status`: Filter by sync status
- `sort`: created_at_desc (default), created_at_asc, size_desc

**Request:**
```
GET /api/v1/media/list?page=1&limit=50&media_type=VIDEO&sort=created_at_desc
Authorization: Bearer <jwt_token>
```

**Response (200 OK):**
```json
{
  "success": true,
  "data": {
    "assets": [
      {
        "asset_id": "uuid",
        "youtube_video_id": "dQw4w9WgXcQ",
        "original_filename": "video.mp4",
        "file_size_bytes": 104857600,
        "media_type": "VIDEO",
        "sync_status": "COMPLETED",
        "created_at": "2026-03-24T10:00:00Z",
        "upload_completed_at": "2026-03-24T10:05:30Z"
      }
    ],
    "pagination": {
      "page": 1,
      "limit": 50,
      "total": 120,
      "total_pages": 3
    }
  }
}
```

---

## 5. UI/UX Design

### 5.1 Screen Layout (Mobile App)

#### Main Screen: Dashboard

```
┌────────────────────────────────────┐
│  ☰  Media Backup     [Settings] 🔔 │
├────────────────────────────────────┤
│                                    │
│  📊 Storage Status                 │
│  ━━━━━━━━━━━━━━━━━━━ 75%          │
│  15 GB freed / 20 GB total         │
│                                    │
│  🔄 Sync Status                    │
│  ━━━━━━━━━━━━━━━━━━━ 85%          │
│  85 / 100 files synced             │
│  [Pause] [Resume]                  │
│                                    │
│  📁 Recent Backups                 │
│  ┌──────────────────────────────┐ │
│  │ 🎥 video_2026_03_24.mp4      │ │
│  │    ✅ 120 MB - 2h ago        │ │
│  └──────────────────────────────┘ │
│  ┌──────────────────────────────┐ │
│  │ 🎥 morning_walk.mp4          │ │
│  │    🔄 Uploading... 45%       │ │
│  └──────────────────────────────┘ │
│                                    │
├────────────────────────────────────┤
│  [Home] [History] [Settings]       │
└────────────────────────────────────┘
```

#### History Screen

```
┌────────────────────────────────────┐
│  ← Back     Backup History         │
├────────────────────────────────────┤
│  Filters: [All] [Videos] [Images]  │
│  Sort: [Latest First ▼]            │
│                                    │
│  Today                             │
│  ┌──────────────────────────────┐ │
│  │ 🎥 video_001.mp4             │ │
│  │    ✅ 250 MB - 10:30 AM      │ │
│  │    [View] [Restore] [Delete] │ │
│  └──────────────────────────────┘ │
│                                    │
│  Yesterday                         │
│  ┌──────────────────────────────┐ │
│  │ 📷 photo_123.jpg             │ │
│  │    ✅ 5 MB - 8:15 PM         │ │
│  │    [View] [Restore] [Delete] │ │
│  └──────────────────────────────┘ │
│                                    │
│  [Load More...]                    │
└────────────────────────────────────┘
```

### 5.2 User Flow

```
App Launch → Check Auth Status
  ├─ Not Logged In → Google Login → Grant YouTube Permission → Dashboard
  └─ Logged In → Dashboard

Dashboard → Background Service Running
  ├─ New File Detected → Add to Queue → Show Notification
  └─ WiFi Available → Start Upload → Update Progress → Complete → Delete Local File

User Actions:
  ├─ View History → List Screen → Select Media → Details Screen
  ├─ Restore Media → Confirmation Dialog → Download → Save to Device
  └─ Settings → Configure (WiFi only, Upload schedule, Auto-delete)
```

### 5.3 Component List

| Component | Location | Responsibility |
|-----------|----------|----------------|
| DashboardScreen | lib/screens/dashboard_screen.dart | Main app screen with sync status |
| HistoryScreen | lib/screens/history_screen.dart | List backed up media |
| SettingsScreen | lib/screens/settings_screen.dart | User preferences |
| MediaCard | lib/widgets/media_card.dart | Display individual media item |
| UploadProgressIndicator | lib/widgets/upload_progress.dart | Show upload progress |
| SyncStatusWidget | lib/widgets/sync_status.dart | Display sync statistics |
| AuthProvider | lib/providers/auth_provider.dart | Manage authentication state |
| UploadProvider | lib/providers/upload_provider.dart | Manage upload queue state |
| BackgroundService | lib/services/background_service.dart | WorkManager integration |
| ApiClient | lib/services/api_client.dart | HTTP communication |
| DatabaseHelper | lib/services/database_helper.dart | SQLite operations |

---

## 6. Error Handling

### 6.1 Error Code Definition

| Code | Message | Cause | Handling |
|------|---------|-------|----------|
| AUTH_001 | Invalid credentials | OAuth failure | Redirect to login |
| AUTH_002 | JWT expired | Token timeout | Refresh token |
| UPLOAD_001 | File too large | Size > 2GB | Show size limit error |
| UPLOAD_002 | Invalid file format | Unsupported type | Show format error |
| UPLOAD_003 | YouTube quota exceeded | API limit hit | Retry after 24h |
| UPLOAD_004 | Network timeout | Connection lost | Add to retry queue |
| STORAGE_001 | Insufficient storage | Disk full | Notify user |
| SYNC_001 | Verification failed | Video not playable | Mark as failed, notify user |

### 6.2 Error Response Format

```json
{
  "success": false,
  "error": {
    "code": "UPLOAD_003",
    "message": "YouTube API quota exceeded. Please try again tomorrow.",
    "details": {
      "quota_reset_at": "2026-03-25T00:00:00Z"
    }
  }
}
```

### 6.3 Retry Strategy

**Exponential Backoff:**
- Retry 1: 1 minute
- Retry 2: 5 minutes
- Retry 3: 15 minutes
- Retry 4: 1 hour
- Retry 5: 24 hours
- Max retries: 5

**Retry Conditions:**
- Network timeout
- 5xx server errors
- YouTube API rate limits
- Temporary authentication failures

**No Retry:**
- 400 Bad Request (invalid input)
- 401 Unauthorized (requires re-login)
- 413 Payload Too Large (file size issue)

---

## 7. Security Considerations

- [x] Input validation (file size, format, filename sanitization)
- [x] JWT-based authentication (15min access, 7 day refresh)
- [x] Encrypted storage for YouTube refresh tokens (AES-256-GCM)
- [x] HTTPS enforcement (TLS 1.3)
- [x] Rate limiting (100 req/min per user, 1000 req/min global)
- [x] CORS policy (mobile app origins only)
- [x] SQL injection prevention (GORM parameterized queries)
- [x] XSS prevention (no HTML rendering in mobile app)
- [x] Secure file upload (virus scanning optional in Phase 2)
- [x] API key protection (environment variables, never committed)

---

## 8. Test Plan

### 8.1 Test Scope

| Type | Target | Tool | Coverage Goal |
|------|--------|------|---------------|
| Unit Test | Service layer (Go) | `testing` + `testify` | >80% |
| Unit Test | Repository layer (Go) | `testing` + `sqlmock` | >80% |
| Unit Test | Widgets (Flutter) | `flutter_test` | >70% |
| Integration Test | API endpoints (Go) | `httptest` | All endpoints |
| Integration Test | YouTube API | Mock server | All upload scenarios |
| E2E Test | Full upload flow | Manual + Flutter integration_test | Critical path |

### 8.2 Test Cases (Key)

**Backend (Go)**:
- [x] Happy path: User uploads video → YouTube upload succeeds → Metadata saved
- [x] Error scenario: YouTube API returns 403 → Error logged → Retry scheduled
- [x] Edge case: Video upload interrupted → Resumable upload continues
- [x] Edge case: Duplicate video_id → Unique constraint error → Return 409
- [x] Security: Unauthenticated request → Return 401
- [x] Security: JWT expired → Return 401 with refresh hint

**Mobile (Flutter)**:
- [x] Happy path: New video detected → Added to queue → Uploaded → Deleted locally
- [x] Error scenario: Network disconnected mid-upload → Pause → Resume on WiFi
- [x] Edge case: User logs out during upload → Cancel upload → Clear queue
- [x] Edge case: App killed during upload → Resume on restart
- [x] UI: Upload progress displayed correctly (0% → 100%)
- [x] UI: History list pagination works (50 items per page)

---

## 9. Clean Architecture

### 9.1 Backend Layer Structure (Go)

| Layer | Responsibility | Location |
|-------|---------------|----------|
| **Handler** | HTTP request/response, validation | `internal/handler/` |
| **Service** | Business logic, orchestration | `internal/service/` |
| **Domain** | Entities, business rules | `internal/domain/` |
| **Repository** | Data access, persistence | `internal/repository/` |
| **Infrastructure** | External services (YouTube, S3) | `internal/pkg/` |

### 9.2 Mobile Layer Structure (Flutter)

| Layer | Responsibility | Location |
|-------|---------------|----------|
| **Presentation** | Widgets, screens, UI | `lib/screens/`, `lib/widgets/` |
| **Application** | State management, use cases | `lib/providers/`, `lib/usecases/` |
| **Domain** | Models, business rules | `lib/models/` |
| **Infrastructure** | API client, SQLite, file system | `lib/services/` |

### 9.3 Dependency Rules

```
┌─────────────────────────────────────────────────────────────┐
│                    Backend (Go)                              │
├─────────────────────────────────────────────────────────────┤
│   Handler ──→ Service ──→ Domain ←── Repository             │
│                  │                         │                 │
│                  └──→ Infrastructure ←─────┘                 │
│                        (YouTube, S3)                         │
└─────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────┐
│                    Mobile (Flutter)                          │
├─────────────────────────────────────────────────────────────┤
│   Presentation ──→ Application ──→ Domain ←── Infrastructure│
│                         │                         │          │
│                         └──→ Services ←───────────┘          │
│                          (API, SQLite, FileSystem)           │
└─────────────────────────────────────────────────────────────┘
```

### 9.4 This Feature's Layer Assignment

#### Backend Components

| Component | Layer | Location |
|-----------|-------|----------|
| MediaHandler | Handler | `internal/handler/media_handler.go` |
| UploadService | Service | `internal/service/upload_service.go` |
| YouTubeService | Service | `internal/service/youtube_service.go` |
| MediaAsset (entity) | Domain | `internal/domain/media_asset.go` |
| MediaRepository | Repository | `internal/repository/media_repository.go` |
| YouTubeClient | Infrastructure | `internal/pkg/youtube/client.go` |

#### Mobile Components

| Component | Layer | Location |
|-----------|-------|----------|
| DashboardScreen | Presentation | `lib/screens/dashboard_screen.dart` |
| HistoryScreen | Presentation | `lib/screens/history_screen.dart` |
| UploadProvider | Application | `lib/providers/upload_provider.dart` |
| BackgroundUploadUseCase | Application | `lib/usecases/background_upload.dart` |
| LocalMediaAsset | Domain | `lib/models/local_media_asset.dart` |
| ApiClient | Infrastructure | `lib/services/api_client.dart` |
| DatabaseHelper | Infrastructure | `lib/services/database_helper.dart` |
| BackgroundService | Infrastructure | `lib/services/background_service.dart` |

---

## 10. Coding Convention Reference

### 10.1 Naming Conventions

#### Backend (Go)

| Target | Rule | Example |
|--------|------|---------|
| Structs | PascalCase | `MediaAsset`, `UploadSession` |
| Functions | camelCase (exported: PascalCase) | `uploadVideo()`, `CreateMediaAsset()` |
| Variables | camelCase | `mediaAsset`, `uploadService` |
| Constants | UPPER_SNAKE_CASE | `MAX_FILE_SIZE`, `YOUTUBE_SCOPE` |
| Interfaces | PascalCase + "er" suffix | `MediaRepository`, `Uploader` |
| Files | snake_case | `media_handler.go`, `upload_service.go` |
| Packages | lowercase | `handler`, `service`, `domain` |

#### Mobile (Flutter/Dart)

| Target | Rule | Example |
|--------|------|---------|
| Classes | PascalCase | `DashboardScreen`, `UploadProvider` |
| Functions | camelCase | `uploadVideo()`, `initializeDatabase()` |
| Variables | camelCase | `mediaAsset`, `uploadProgress` |
| Constants | lowerCamelCase | `maxFileSize`, `apiBaseUrl` |
| Private members | \_camelCase | `_uploadQueue`, `_initializeState()` |
| Files | snake_case | `dashboard_screen.dart`, `upload_provider.dart` |

### 10.2 Import Order

#### Go

```go
// 1. Standard library
import (
    "context"
    "fmt"
    "time"
)

// 2. External packages
import (
    "github.com/gin-gonic/gin"
    "google.golang.org/api/youtube/v3"
)

// 3. Internal packages
import (
    "github.com/yourusername/video-backup/internal/domain"
    "github.com/yourusername/video-backup/internal/service"
)
```

#### Dart

```dart
// 1. Dart SDK
import 'dart:async';
import 'dart:io';

// 2. Flutter framework
import 'package:flutter/material.dart';

// 3. External packages
import 'package:provider/provider.dart';
import 'package:dio/dio.dart';

// 4. Internal imports
import 'package:video_backup_app/models/local_media_asset.dart';
import 'package:video_backup_app/services/api_client.dart';

// 5. Relative imports
import '../widgets/media_card.dart';
```

### 10.3 Environment Variables

#### Backend (.env)

```bash
# Server
PORT=8080
ENV=development

# Database
DATABASE_URL=postgresql://user:pass@localhost:5432/video_backup?sslmode=disable

# Google OAuth
GOOGLE_CLIENT_ID=your-client-id.apps.googleusercontent.com
GOOGLE_CLIENT_SECRET=your-secret
GOOGLE_REDIRECT_URL=http://localhost:8080/api/v1/auth/google/callback

# YouTube API
YOUTUBE_API_KEY=your-youtube-api-key
YOUTUBE_UPLOAD_SCOPE=https://www.googleapis.com/auth/youtube.upload

# JWT
JWT_SECRET=your-32-character-secret-key
JWT_ACCESS_EXPIRY=15m
JWT_REFRESH_EXPIRY=168h

# Encryption
ENCRYPTION_KEY=your-exactly-32-character-key!

# Rate Limiting
RATE_LIMIT_GENERAL=100
RATE_LIMIT_UPLOAD=10
```

#### Mobile (.env)

```bash
API_BASE_URL=https://api.yourdomain.com
GOOGLE_CLIENT_ID=your-client-id.apps.googleusercontent.com
```

### 10.4 This Feature's Conventions

| Item | Convention Applied |
|------|-------------------|
| Component naming | PascalCase for widgets/screens, camelCase for functions |
| File organization | Feature-based structure (screens/, widgets/, services/) |
| State management | Provider pattern for app-wide state |
| Error handling | Custom exception classes with error codes |
| Logging | Structured logging with log levels (debug, info, warn, error) |
| Testing | Test file naming: `{filename}_test.go` or `{filename}_test.dart` |

---

## 11. Implementation Guide

### 11.1 Backend File Structure

```
backend/
├── cmd/
│   └── api/
│       └── main.go                      # Application entry point
├── internal/
│   ├── domain/
│   │   ├── user.go                      # User entity
│   │   ├── media_asset.go               # MediaAsset entity
│   │   ├── upload_session.go            # UploadSession entity
│   │   └── errors.go                    # Domain errors
│   ├── repository/
│   │   ├── interfaces.go                # Repository contracts
│   │   ├── user_repository.go
│   │   ├── media_repository.go
│   │   └── session_repository.go
│   ├── service/
│   │   ├── auth_service.go              # OAuth + JWT
│   │   ├── upload_service.go            # Upload orchestration
│   │   ├── youtube_service.go           # YouTube API integration
│   │   └── s3_service.go                # S3 integration (Phase 2)
│   ├── handler/
│   │   ├── auth_handler.go
│   │   ├── media_handler.go
│   │   ├── dto.go                       # Request/Response DTOs
│   │   └── response.go                  # Response helpers
│   ├── middleware/
│   │   ├── auth.go
│   │   ├── rate_limiter.go
│   │   ├── cors.go
│   │   └── error_handler.go
│   ├── router/
│   │   └── router.go                    # Route configuration
│   └── pkg/
│       ├── database/
│       │   └── postgres.go
│       ├── logger/
│       │   └── logger.go
│       └── youtube/
│           └── client.go
├── migrations/
│   ├── 000001_create_users_table.up.sql
│   ├── 000001_create_users_table.down.sql
│   ├── 000002_create_media_assets_table.up.sql
│   ├── 000002_create_media_assets_table.down.sql
│   ├── 000003_create_upload_sessions_table.up.sql
│   └── 000003_create_upload_sessions_table.down.sql
├── .env
├── go.mod
└── README.md
```

### 11.2 Mobile File Structure

```
lib/
├── main.dart                           # App entry point
├── app.dart                            # MaterialApp configuration
├── screens/
│   ├── splash_screen.dart
│   ├── login_screen.dart
│   ├── dashboard_screen.dart
│   ├── history_screen.dart
│   ├── media_detail_screen.dart
│   └── settings_screen.dart
├── widgets/
│   ├── media_card.dart
│   ├── upload_progress.dart
│   ├── sync_status.dart
│   └── custom_app_bar.dart
├── providers/
│   ├── auth_provider.dart
│   ├── upload_provider.dart
│   └── settings_provider.dart
├── models/
│   ├── local_media_asset.dart
│   ├── upload_session.dart
│   └── user.dart
├── services/
│   ├── api_client.dart
│   ├── database_helper.dart
│   ├── background_service.dart
│   ├── file_service.dart
│   └── notification_service.dart
├── usecases/
│   ├── background_upload.dart
│   ├── file_detection.dart
│   └── sync_status.dart
├── utils/
│   ├── constants.dart
│   ├── validators.dart
│   └── formatters.dart
└── config/
    └── app_config.dart
```

### 11.3 Implementation Order

#### Phase 1: Backend Foundation (Week 1)

1. [x] Project setup (Go modules, folder structure)
2. [x] Database migrations (users, media_assets, upload_sessions tables)
3. [x] Domain entities (User, MediaAsset, UploadSession)
4. [x] Repository layer (interfaces + GORM implementations)
5. [x] Google OAuth + JWT authentication
6. [x] Basic HTTP server + middleware

#### Phase 2: YouTube Integration (Week 2)

1. [ ] YouTube API client setup
2. [ ] Upload service implementation (resumable upload)
3. [ ] Media handler endpoints (initiate, upload, status)
4. [ ] Error handling and retry logic
5. [ ] Integration tests

#### Phase 3: Mobile Foundation (Week 3)

1. [ ] Flutter project setup
2. [ ] SQLite database schema
3. [ ] API client implementation (dio)
4. [ ] Authentication flow (OAuth + JWT)
5. [ ] Basic UI screens (Login, Dashboard)

#### Phase 4: File Detection & Upload (Week 4)

1. [ ] File system watcher implementation
2. [ ] Local queue management (SQLite)
3. [ ] Upload provider (state management)
4. [ ] Network status monitoring
5. [ ] Background service setup (WorkManager)

#### Phase 5: Background Sync (Week 5)

1. [ ] Background upload service
2. [ ] Retry logic with exponential backoff
3. [ ] Progress tracking and notifications
4. [ ] Local file deletion after verification
5. [ ] Upload history UI

#### Phase 6: Testing & Polish (Week 6)

1. [ ] Unit tests (backend + mobile)
2. [ ] Integration tests
3. [ ] E2E tests (critical path)
4. [ ] UI/UX improvements
5. [ ] Performance optimization

#### Phase 7: Phase 2 Features (Week 7+)

1. [ ] S3 integration for images
2. [ ] Restore functionality
3. [ ] Advanced settings (schedule, filters)
4. [ ] Beta testing

---

## Version History

| Version | Date | Changes | Author |
|---------|------|---------|--------|
| 0.1 | 2026-03-24 | Initial draft | Development Team |
