# Plan: Media Backup System (Retrospective)

**Feature**: media-backup-system
**Status**: Implemented (Retrospective Documentation)
**Created**: 2026-03-25
**Type**: Backend API

---

## Executive Summary

| Perspective | Description |
|------------|-------------|
| **Problem** | Users need a reliable way to backup their YouTube videos to cloud storage without losing metadata, with support for large files (up to 2GB) and resumable uploads. |
| **Solution** | RESTful API service providing YouTube video upload management with session-based tracking, resumable uploads, and comprehensive error handling. |
| **Functional UX Effect** | Users can initiate upload sessions, upload multiple videos with progress tracking, monitor session status, and manage their media assets through a clean REST API. |
| **Core Value** | Automated video backup with 90%+ reliability through exponential backoff retry logic, session management for batch uploads, and complete audit trail of all upload activities. |

---

## Context Anchor

| Dimension | Content |
|-----------|---------|
| **WHY** | Enable reliable YouTube video backup to cloud storage with support for large files and batch operations |
| **WHO** | YouTube content creators needing automated backup solutions |
| **RISK** | YouTube API quota limits, network failures during large uploads, OAuth token expiration |
| **SUCCESS** | 90%+ upload success rate, support for 2GB files, session-based batch uploads with resume capability |
| **SCOPE** | Upload sessions, video upload API, media asset management, YouTube API integration - excludes video editing, transcoding, or analytics features |

---

## 1. Feature Overview

### 1.1 Purpose
Provide a backend API service for uploading and managing video backups to YouTube with robust error handling, session management, and progress tracking.

### 1.2 Target Users
- YouTube content creators
- Media archival services
- Content backup automation systems

### 1.3 Business Value
- Reliable video backup solution
- Support for large file uploads (up to 2GB)
- Batch upload capability with session management
- Complete upload history and audit trail

---

## 2. Requirements

### 2.1 Functional Requirements

**FR-1: Upload Session Management**
- Users can initiate upload sessions with total file count and total bytes
- Sessions track progress (completed files, failed files, uploaded bytes)
- Sessions support multiple video uploads within a single batch
- Sessions can be completed or cancelled

**FR-2: Video Upload**
- Support video files up to 2GB
- Resumable upload protocol (10MB chunks)
- Progress tracking via callback
- Automatic retry with exponential backoff (5 attempts, 1s→30s)
- Video verification after upload (playability check)

**FR-3: Media Asset Management**
- List media assets with pagination (default 50, max 100 per page)
- Filter by media type (VIDEO/IMAGE)
- Filter by sync status (PENDING/UPLOADING/COMPLETED/FAILED)
- Sort by created_at (desc/asc) or size (desc)
- Get single asset details
- Delete asset records (soft delete)

**FR-4: YouTube Integration**
- OAuth 2.0 per-request token injection
- Resumable upload protocol
- Video status verification
- Error classification (8 distinct error types)

### 2.2 Non-Functional Requirements

**NFR-1: Performance**
- Chunk size: 10MB for resumable uploads
- Request timeout: 10s read, 10s write
- Idle timeout: 120s
- Max header bytes: 1MB

**NFR-2: Reliability**
- Exponential backoff retry logic (base 1s, max 30s, jitter)
- 5 retry attempts for transient failures
- Session-based progress tracking
- Error classification system

**NFR-3: Security**
- Per-request OAuth token validation
- JWT-based authentication
- Rate limiting (inherited from Phase 1)
- AES-256-GCM token encryption (inherited from Phase 1)

**NFR-4: Scalability**
- Paginated media asset listing
- Session-based batch uploads
- Database connection pooling (inherited)
- Redis caching (inherited)

---

## 3. Success Criteria

| ID | Criterion | Target | Measurement |
|----|-----------|--------|-------------|
| SC-1 | Upload Success Rate | ≥ 90% | (Completed uploads / Total uploads) × 100 |
| SC-2 | File Size Support | 2GB max | Max file size successfully uploaded |
| SC-3 | API Response Time | < 500ms | p95 response time (excluding actual upload) |
| SC-4 | Error Classification | 8 types | Distinct error codes with HTTP status mapping |
| SC-5 | Test Coverage | ≥ 80% | YouTube client test coverage |
| SC-6 | Session Tracking | 100% | All uploads linked to valid sessions |

---

## 4. Scope

### 4.1 In Scope
- Upload session initiation and management
- Video upload to YouTube with resumable protocol
- Media asset CRUD operations
- YouTube API integration (upload, status, delete)
- Error handling and retry logic
- Progress tracking callbacks
- Pagination and filtering

### 4.2 Out of Scope
- Video transcoding or format conversion
- Thumbnail generation
- Video analytics or statistics
- Multi-platform upload (only YouTube)
- Video editing or manipulation
- Subtitle or caption management
- Scheduled uploads or cron jobs
- S3/cloud storage integration (Phase 3)

---

## 5. Assumptions

1. **Authentication**: Users are already authenticated via Phase 1 OAuth flow
2. **YouTube Access**: Users have valid YouTube channel access
3. **OAuth Tokens**: Access tokens are provided per-request (no refresh logic in upload service)
4. **File Storage**: Video files are provided as local file paths (temporary storage managed by handler)
5. **Database**: PostgreSQL and Redis are already configured and running
6. **YouTube Quota**: YouTube API quota is sufficient for intended usage volume

---

## 6. Constraints

1. **File Size**: Maximum 2GB per file (YouTube API limit)
2. **Chunk Size**: 10MB chunks for resumable upload
3. **Retry Limit**: Maximum 5 retry attempts
4. **Timeout**: 10s read/write, 120s idle
5. **Privacy**: All uploads default to "private" status
6. **Temporary Storage**: Uploaded files stored in `/tmp/media-backup-uploads`

---

## 7. Dependencies

### 7.1 External Dependencies
- **YouTube Data API v3**: Video upload, status check, delete operations
- **Google OAuth 2.0**: Access token validation (Phase 1)
- **PostgreSQL**: Media asset and session persistence
- **Redis**: Token caching and rate limiting (Phase 1)

### 7.2 Internal Dependencies
- **Phase 1 Authentication**: JWT and OAuth token management
- **Database Layer**: GORM ORM with migration support
- **Logger**: Structured logging with slog
- **Configuration**: Environment-based config management

### 7.3 Go Packages
```go
"google.golang.org/api/youtube/v3"
"google.golang.org/api/googleapi"
"golang.org/x/oauth2"
"github.com/gin-gonic/gin"
"github.com/google/uuid"
"gorm.io/gorm"
```

---

## 8. Risks and Mitigation

| Risk | Impact | Probability | Mitigation |
|------|--------|-------------|------------|
| YouTube API quota exhaustion | High | Medium | Implement rate limiting, use exponential backoff, monitor quota usage |
| Network failures during upload | High | High | Resumable upload protocol, retry logic with exponential backoff |
| OAuth token expiration mid-upload | Medium | Low | Per-request token validation, clear error messages for token refresh |
| Database connection pool exhaustion | Medium | Low | Connection pooling configured (max 100 conns), proper connection cleanup |
| Disk space for temp files | Medium | Medium | Immediate cleanup after upload, `/tmp` cleanup via defer |
| Large file memory usage | Medium | Low | Streaming upload with 10MB chunks, progress reader wrapper |

---

## 9. Implementation Notes

### 9.1 Architecture
- **Clean Architecture**: Domain → Repository → Service → Handler
- **Dependency Injection**: Constructor-based injection
- **Error Handling**: Domain-specific error types with HTTP mapping

### 9.2 Key Components
1. **Domain Models**: `MediaAsset`, `UploadSession`
2. **Repositories**: `MediaRepository`, `SessionRepository`
3. **Services**: `UploadService`, `YouTubeClient`
4. **Handlers**: `MediaHandler` (7 REST endpoints)

### 9.3 API Endpoints (7 total)
- `POST /api/v1/media/upload/initiate` - Create upload session
- `POST /api/v1/media/upload/video` - Upload video file
- `GET /api/v1/media/upload/status/:session_id` - Get session status
- `POST /api/v1/media/upload/complete` - Complete session
- `GET /api/v1/media/list` - List media assets (paginated)
- `GET /api/v1/media/:asset_id` - Get asset details
- `DELETE /api/v1/media/:asset_id` - Delete asset

### 9.4 Error Types (8 distinct)
1. `AUTH_INVALID` (401)
2. `FILE_TOO_LARGE` (413)
3. `SESSION_NOT_FOUND` (404)
4. `UPLOAD_FAILED` (500)
5. `UPLOAD_RETRY` (500 with retry flag)
6. `ASSET_NOT_FOUND` (404)
7. `INVALID_REQUEST` (400)
8. `RATE_LIMIT_EXCEEDED` (429, inherited)

---

## 10. Acceptance Criteria

### AC-1: Upload Session Lifecycle
- [ ] User can initiate session with total files and bytes
- [ ] Session tracks progress (completed, failed, uploaded bytes)
- [ ] User can query session status at any time
- [ ] User can complete or cancel session

### AC-2: Video Upload
- [ ] Supports files up to 2GB
- [ ] Uses resumable upload (10MB chunks)
- [ ] Provides progress callbacks
- [ ] Retries on failure (up to 5 times with exponential backoff)
- [ ] Verifies video is playable after upload

### AC-3: Media Asset Management
- [ ] List assets with pagination (default 50, max 100)
- [ ] Filter by media type and sync status
- [ ] Sort by created_at or size
- [ ] Get single asset details
- [ ] Delete asset records

### AC-4: Error Handling
- [ ] All 8 error types have proper HTTP status codes
- [ ] Retry logic triggers for transient errors
- [ ] Non-retryable errors fail immediately
- [ ] Error messages are clear and actionable

### AC-5: Testing
- [ ] YouTube client has ≥80% test coverage
- [ ] Upload service has unit tests with mocks
- [ ] Handler tests cover all 7 endpoints
- [ ] Integration tests verify end-to-end flow

---

## 11. Future Enhancements (Phase 3+)

1. **S3 Storage Integration**: Store videos in S3 before/after YouTube upload
2. **Webhook Notifications**: Notify on upload completion/failure
3. **Batch Operations**: Bulk upload from S3 or external sources
4. **Video Analytics**: Track views, watch time, engagement
5. **Scheduled Uploads**: Cron-based automated backups
6. **Multi-Platform**: Support Vimeo, Dailymotion, etc.
7. **Thumbnail Management**: Custom thumbnail upload
8. **Caption Support**: Subtitle/caption file upload

---

## Appendix

### A. Database Schema

**media_assets table**:
- `asset_id` (UUID, PK)
- `user_id` (UUID, FK)
- `youtube_video_id` (VARCHAR)
- `s3_object_key` (VARCHAR)
- `original_filename` (VARCHAR)
- `file_size_bytes` (BIGINT)
- `media_type` (VARCHAR: VIDEO/IMAGE)
- `sync_status` (VARCHAR: PENDING/UPLOADING/COMPLETED/FAILED)
- `upload_started_at` (TIMESTAMP)
- `upload_completed_at` (TIMESTAMP)
- `error_message` (TEXT)
- `retry_count` (INTEGER)
- `created_at` (TIMESTAMP)
- `updated_at` (TIMESTAMP)

**upload_sessions table**:
- `session_id` (UUID, PK)
- `user_id` (UUID, FK)
- `total_files` (INTEGER)
- `completed_files` (INTEGER)
- `failed_files` (INTEGER)
- `total_bytes` (BIGINT)
- `uploaded_bytes` (BIGINT)
- `session_status` (VARCHAR: ACTIVE/COMPLETED/CANCELLED)
- `started_at` (TIMESTAMP)
- `completed_at` (TIMESTAMP)
- `created_at` (TIMESTAMP)
- `updated_at` (TIMESTAMP)

### B. Implementation Statistics
- **Total Files**: 10 files (4 domain, 2 repositories, 2 services, 2 handlers + YouTube client)
- **Total Lines**: ~2,400 lines of code
- **Test Coverage**: YouTube Client 47.9% (6/6 tests PASS)
- **API Endpoints**: 7 REST endpoints
- **Error Types**: 8 distinct error codes

### C. References
- [YouTube Data API v3 Documentation](https://developers.google.com/youtube/v3/docs)
- [Resumable Upload Protocol](https://developers.google.com/youtube/v3/guides/using_resumable_upload_protocol)
- [OAuth 2.0 for Google APIs](https://developers.google.com/identity/protocols/oauth2)
