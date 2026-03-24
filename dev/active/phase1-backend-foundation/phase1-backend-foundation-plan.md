# Strategic Implementation Plan: Phase 1 Backend Foundation

**Project**: media-backup-system - YouTube Unlimited Media Backup
**Phase**: 1 of 6 - Backend Foundation
**Created**: 2026-03-24
**Last Updated**: 2026-03-24
**Status**: Ready to Start
**Estimated Duration**: 1 week (40 hours)

---

## Executive Summary

This plan outlines the implementation of Phase 1 (Backend Foundation) for the media-backup-system feature. This phase establishes the core backend infrastructure required for unlimited media backup using YouTube Data API v3. The foundation will extend the existing youtube-auth-api backend with new domain entities, database schema, and repository layer for managing media assets and upload sessions.

**Key Deliverables**:
- PostgreSQL database schema for media_assets and upload_sessions tables
- Go domain entities (MediaAsset, UploadSession) with business logic
- Repository layer with GORM implementations
- Database migrations (up/down scripts)
- Unit tests for domain and repository layers

**Success Criteria**:
- All migrations run successfully
- Domain entities compile without errors
- Repository layer passes unit tests (>80% coverage)
- Clean integration with existing youtube-auth-api infrastructure

---

## Current State Analysis

### Existing Infrastructure (Reusable)

**youtube-auth-api Backend** (92% complete, archived):
```
✅ User authentication (Google OAuth + JWT)
✅ Token management (encryption, blacklist, refresh)
✅ Clean Architecture (Handler → Service → Domain → Repository)
✅ Middleware (auth, CORS, rate limiter, error handler)
✅ PostgreSQL + Redis infrastructure
✅ Database migrations framework
```

**Database Schema (Existing)**:
```sql
users table:
  - user_id (UUID PK)
  - google_account_id (VARCHAR 255, unique)
  - email (VARCHAR 255)
  - youtube_channel_id (VARCHAR 255)
  - refresh_token (TEXT, encrypted)
  - created_at, updated_at, last_login_at

user_tokens table:
  - id (UUID PK)
  - user_id (UUID FK → users)
  - encrypted_access_token (TEXT)
  - encrypted_refresh_token (TEXT)
  - token_type, expires_at, created_at, updated_at
```

### Gaps to Address

**Missing Components** (Phase 1 Scope):
- ❌ media_assets table schema
- ❌ upload_sessions table schema
- ❌ MediaAsset domain entity
- ❌ UploadSession domain entity
- ❌ MediaRepository interface + implementation
- ❌ UploadSessionRepository interface + implementation

---

## Proposed Future State

### New Database Schema

```sql
media_assets table:
  - asset_id (UUID PK, gen_random_uuid())
  - user_id (UUID FK → users, CASCADE DELETE)
  - youtube_video_id (VARCHAR 255, unique, nullable)
  - s3_object_key (VARCHAR 512, nullable)
  - original_filename (VARCHAR 512, not null)
  - file_size_bytes (BIGINT, not null)
  - media_type (VARCHAR 10, CHECK: VIDEO|IMAGE)
  - sync_status (VARCHAR 20, CHECK: PENDING|UPLOADING|COMPLETED|FAILED)
  - upload_started_at (TIMESTAMP, nullable)
  - upload_completed_at (TIMESTAMP, nullable)
  - error_message (TEXT, nullable)
  - retry_count (INT, default 0)
  - created_at (TIMESTAMP, default NOW())
  - updated_at (TIMESTAMP, default NOW())

  Indexes:
    - idx_media_assets_user (user_id)
    - idx_media_assets_status (sync_status)
    - idx_media_assets_created (created_at DESC)

upload_sessions table:
  - session_id (UUID PK, gen_random_uuid())
  - user_id (UUID FK → users, CASCADE DELETE)
  - total_files (INT, default 0)
  - completed_files (INT, default 0)
  - failed_files (INT, default 0)
  - total_bytes (BIGINT, default 0)
  - uploaded_bytes (BIGINT, default 0)
  - session_status (VARCHAR 20, CHECK: ACTIVE|COMPLETED|CANCELLED)
  - started_at (TIMESTAMP, default NOW())
  - completed_at (TIMESTAMP, nullable)

  Indexes:
    - idx_upload_sessions_user (user_id)
    - idx_upload_sessions_status (session_status)
```

### Go Domain Entities

**MediaAsset Entity**:
```go
type MediaType string
type SyncStatus string

const (
    MediaTypeVideo MediaType = "VIDEO"
    MediaTypeImage MediaType = "IMAGE"

    SyncStatusPending   SyncStatus = "PENDING"
    SyncStatusUploading SyncStatus = "UPLOADING"
    SyncStatusCompleted SyncStatus = "COMPLETED"
    SyncStatusFailed    SyncStatus = "FAILED"
)

type MediaAsset struct {
    AssetID           uuid.UUID  `gorm:"type:uuid;primaryKey"`
    UserID            uuid.UUID  `gorm:"type:uuid;not null;index"`
    YouTubeVideoID    *string    `gorm:"type:varchar(255);uniqueIndex"`
    S3ObjectKey       *string    `gorm:"type:varchar(512)"`
    OriginalFilename  string     `gorm:"type:varchar(512);not null"`
    FileSizeBytes     int64      `gorm:"not null"`
    MediaType         MediaType  `gorm:"type:varchar(10);not null"`
    SyncStatus        SyncStatus `gorm:"type:varchar(20);not null;default:PENDING"`
    UploadStartedAt   *time.Time
    UploadCompletedAt *time.Time
    ErrorMessage      *string    `gorm:"type:text"`
    RetryCount        int        `gorm:"default:0"`
    CreatedAt         time.Time  `gorm:"default:CURRENT_TIMESTAMP"`
    UpdatedAt         time.Time  `gorm:"default:CURRENT_TIMESTAMP"`

    User User `gorm:"foreignKey:UserID"`
}
```

**UploadSession Entity**:
```go
type SessionStatus string

const (
    SessionStatusActive    SessionStatus = "ACTIVE"
    SessionStatusCompleted SessionStatus = "COMPLETED"
    SessionStatusCancelled SessionStatus = "CANCELLED"
)

type UploadSession struct {
    SessionID      uuid.UUID      `gorm:"type:uuid;primaryKey"`
    UserID         uuid.UUID      `gorm:"type:uuid;not null;index"`
    TotalFiles     int            `gorm:"not null;default:0"`
    CompletedFiles int            `gorm:"not null;default:0"`
    FailedFiles    int            `gorm:"not null;default:0"`
    TotalBytes     int64          `gorm:"not null;default:0"`
    UploadedBytes  int64          `gorm:"not null;default:0"`
    SessionStatus  SessionStatus  `gorm:"type:varchar(20);not null;default:ACTIVE"`
    StartedAt      time.Time      `gorm:"default:CURRENT_TIMESTAMP"`
    CompletedAt    *time.Time

    User User `gorm:"foreignKey:UserID"`
}
```

---

## Implementation Phases

### Phase 1.1: Database Migrations (Priority: HIGH)

**Objective**: Create PostgreSQL migration scripts for new tables

**Tasks**:

1. **Create media_assets UP migration** [Effort: S]
   - File: `backend/migrations/000003_create_media_assets_table.up.sql`
   - Define complete table schema with all columns
   - Add CHECK constraints for enums
   - Create 3 indexes (user_id, sync_status, created_at DESC)
   - **Acceptance Criteria**:
     - SQL syntax valid for PostgreSQL 15+
     - All constraints properly defined
     - Indexes created for query optimization
     - Foreign key references users(user_id) with CASCADE DELETE

2. **Create media_assets DOWN migration** [Effort: S]
   - File: `backend/migrations/000003_create_media_assets_table.down.sql`
   - Drop table with IF EXISTS check
   - **Acceptance Criteria**:
     - Cleanly removes table and all constraints
     - Reversible migration for rollback scenarios

3. **Create upload_sessions UP migration** [Effort: S]
   - File: `backend/migrations/000004_create_upload_sessions_table.up.sql`
   - Define complete table schema
   - Add CHECK constraints for session_status enum
   - Create 2 indexes (user_id, session_status)
   - **Acceptance Criteria**:
     - Same quality standards as media_assets migration
     - Foreign key to users table

4. **Create upload_sessions DOWN migration** [Effort: S]
   - File: `backend/migrations/000004_create_upload_sessions_table.down.sql`
   - Drop table with IF EXISTS check
   - **Acceptance Criteria**:
     - Cleanly removes table

5. **Test migrations locally** [Effort: S]
   - Run: `migrate -database $DATABASE_URL -path migrations up`
   - Verify tables created with correct schema
   - Run: `migrate -database $DATABASE_URL -path migrations down 2`
   - Verify tables removed cleanly
   - Re-run UP migration to confirm repeatability
   - **Acceptance Criteria**:
     - Migrations run without errors
     - Schema matches design specification
     - UP and DOWN migrations are reversible

**Dependencies**: None (can start immediately)

---

### Phase 1.2: Domain Entities (Priority: HIGH)

**Objective**: Implement Go domain entities with business logic

**Tasks**:

1. **Create MediaAsset entity** [Effort: M]
   - File: `backend/internal/domain/media_asset.go`
   - Define MediaAsset struct with GORM tags
   - Define MediaType and SyncStatus enums
   - Implement `TableName()` method returning "media_assets"
   - Implement `BeforeCreate()` hook for UUID generation
   - Implement `BeforeUpdate()` hook for updated_at timestamp
   - Implement validation methods:
     - `Validate() error` - check required fields
     - `CanRetry() bool` - check if retry_count < 5
     - `MarkAsUploading()` - set status and upload_started_at
     - `MarkAsCompleted(videoID string)` - set status and upload_completed_at
     - `MarkAsFailed(errMsg string)` - set status, error_message, increment retry_count
   - **Acceptance Criteria**:
     - Compiles without errors
     - All GORM tags correctly set
     - Business logic methods implemented
     - Follows existing User entity patterns

2. **Create UploadSession entity** [Effort: M]
   - File: `backend/internal/domain/upload_session.go`
   - Define UploadSession struct with GORM tags
   - Define SessionStatus enum
   - Implement `TableName()` method
   - Implement hooks (BeforeCreate, BeforeUpdate)
   - Implement business methods:
     - `Validate() error`
     - `CalculateProgress() float64` - return percentage
     - `IncrementCompleted()` - increment completed_files
     - `IncrementFailed()` - increment failed_files
     - `UpdateUploadedBytes(bytes int64)`
     - `Complete()` - set status and completed_at
     - `Cancel()` - set status and completed_at
   - **Acceptance Criteria**:
     - Same quality standards as MediaAsset
     - Progress calculation accurate (0-100%)

3. **Update domain errors** [Effort: S]
   - File: `backend/internal/domain/errors.go`
   - Add new error types:
     - `ErrMediaAssetNotFound`
     - `ErrInvalidMediaType`
     - `ErrInvalidSyncStatus`
     - `ErrSessionNotFound`
     - `ErrSessionAlreadyCompleted`
   - **Acceptance Criteria**:
     - Errors follow existing pattern (var ErrName = errors.New("message"))
     - Clear, descriptive error messages

4. **Write domain unit tests** [Effort: M]
   - File: `backend/internal/domain/media_asset_test.go`
   - Test entity creation and validation
   - Test business logic methods (MarkAsUploading, MarkAsCompleted, etc.)
   - Test edge cases (nil pointers, boundary values)
   - File: `backend/internal/domain/upload_session_test.go`
   - Test session lifecycle methods
   - Test progress calculation
   - **Acceptance Criteria**:
     - >80% code coverage for domain entities
     - All edge cases covered
     - Tests use table-driven test pattern

**Dependencies**: None (can run in parallel with Phase 1.1)

---

### Phase 1.3: Repository Layer (Priority: HIGH)

**Objective**: Implement data access layer with GORM

**Tasks**:

1. **Define repository interfaces** [Effort: M]
   - File: `backend/internal/repository/interfaces.go`
   - Add MediaRepository interface:
     ```go
     type MediaRepository interface {
         Create(ctx context.Context, asset *domain.MediaAsset) error
         FindByID(ctx context.Context, assetID string) (*domain.MediaAsset, error)
         FindByUserID(ctx context.Context, userID string, limit, offset int) ([]domain.MediaAsset, int64, error)
         FindPendingUploads(ctx context.Context, userID string) ([]domain.MediaAsset, error)
         Update(ctx context.Context, asset *domain.MediaAsset) error
         Delete(ctx context.Context, assetID string) error
     }
     ```
   - Add UploadSessionRepository interface:
     ```go
     type UploadSessionRepository interface {
         Create(ctx context.Context, session *domain.UploadSession) error
         FindByID(ctx context.Context, sessionID string) (*domain.UploadSession, error)
         FindActiveByUserID(ctx context.Context, userID string) (*domain.UploadSession, error)
         Update(ctx context.Context, session *domain.UploadSession) error
         Complete(ctx context.Context, sessionID string) error
         Cancel(ctx context.Context, sessionID string) error
     }
     ```
   - **Acceptance Criteria**:
     - Interfaces follow existing repository patterns
     - All CRUD operations included
     - Context passed for cancellation support

2. **Implement MediaRepository** [Effort: L]
   - File: `backend/internal/repository/media_repository.go`
   - Implement all interface methods with GORM
   - Handle UUID parsing with proper error mapping
   - Implement pagination in FindByUserID
   - Map database errors to domain errors
   - Use Preload for User relationship when needed
   - **Acceptance Criteria**:
     - All methods implemented correctly
     - Proper error handling and mapping
     - UUID parsing robust
     - Follows existing UserRepository patterns

3. **Implement UploadSessionRepository** [Effort: M]
   - File: `backend/internal/repository/session_repository.go`
   - Implement all interface methods
   - Handle session lifecycle (create, update, complete, cancel)
   - Map errors to domain errors
   - **Acceptance Criteria**:
     - All methods work correctly
     - Proper transaction handling for state changes

4. **Write repository unit tests** [Effort: L]
   - File: `backend/internal/repository/media_repository_test.go`
   - Use sqlmock or testcontainers for database testing
   - Test all CRUD operations
   - Test error scenarios (not found, duplicate, etc.)
   - Test pagination
   - File: `backend/internal/repository/session_repository_test.go`
   - Test session lifecycle
   - Test concurrent access scenarios
   - **Acceptance Criteria**:
     - >80% code coverage
     - All error paths tested
     - Tests can run in parallel

**Dependencies**:
- Phase 1.1 (migrations must be created first)
- Phase 1.2 (domain entities must exist)

---

### Phase 1.4: Integration and Wiring (Priority: MEDIUM)

**Objective**: Integrate new components into existing application

**Tasks**:

1. **Update main.go initialization** [Effort: M]
   - File: `backend/cmd/api/main.go`
   - Initialize media repository: `mediaRepo := repository.NewMediaRepository(db)`
   - Initialize session repository: `sessionRepo := repository.NewSessionRepository(db)`
   - Add to dependency injection chain
   - **Acceptance Criteria**:
     - Application compiles successfully
     - Repositories properly initialized
     - No circular dependencies

2. **Run database migrations in production flow** [Effort: S]
   - Verify migration sequence (000001, 000002, 000003, 000004)
   - Document migration commands in README
   - **Acceptance Criteria**:
     - Migrations run cleanly on fresh database
     - Migrations are idempotent (can run multiple times safely)

3. **Update AutoMigrate for development** [Effort: S]
   - File: `backend/internal/pkg/database/postgres.go`
   - Add MediaAsset and UploadSession to AutoMigrate call
   - **Acceptance Criteria**:
     - Dev environment auto-creates tables
     - No schema drift between migrations and AutoMigrate

4. **Integration testing** [Effort: M]
   - Create integration test file: `backend/tests/integration/media_test.go`
   - Test full stack: HTTP request → Repository → Database
   - Use real PostgreSQL (testcontainers)
   - Test transactions and rollbacks
   - **Acceptance Criteria**:
     - End-to-end flow works
     - Database constraints enforced
     - Foreign key relationships work correctly

**Dependencies**: All previous phases must be complete

---

### Phase 1.5: Documentation and Verification (Priority: LOW)

**Objective**: Document changes and verify phase completion

**Tasks**:

1. **Update backend README** [Effort: S]
   - File: `backend/README.md`
   - Document new database tables
   - Add migration commands
   - Update project structure section
   - **Acceptance Criteria**:
     - Clear instructions for running migrations
     - Schema documented with examples

2. **Create Phase 1 completion checklist** [Effort: S]
   - Verify all migrations work
   - Verify all tests pass: `go test ./internal/domain ./internal/repository -v`
   - Verify test coverage: `go test -cover ./internal/domain ./internal/repository`
   - Verify no linting errors: `golangci-lint run`
   - **Acceptance Criteria**:
     - All tests pass
     - Coverage >80% for new code
     - No linting errors

3. **Update PDCA status** [Effort: S]
   - Update `.pdca-status.json`
   - Mark Phase 1 as complete
   - Update match rate (expected: 15-20%)
   - **Acceptance Criteria**:
     - PDCA tracking updated
     - Ready for Phase 2

**Dependencies**: All implementation phases complete

---

## Risk Assessment and Mitigation Strategies

### High Risk Items

| Risk | Probability | Impact | Mitigation |
|------|------------|--------|------------|
| Migration breaks existing data | Low | High | Test migrations on database dump first, have rollback plan |
| UUID generation conflicts | Low | High | Use PostgreSQL gen_random_uuid(), test with concurrent inserts |
| Foreign key constraint violations | Medium | Medium | Thoroughly test CASCADE DELETE behavior, add database constraints |

### Medium Risk Items

| Risk | Probability | Impact | Mitigation |
|------|------------|--------|------------|
| GORM tag errors causing schema mismatch | Medium | Medium | Compare AutoMigrate schema vs manual migrations, add validation tests |
| Repository error mapping incomplete | Medium | Low | Comprehensive error handling tests, use exhaustive switch statements |
| Performance issues with indexes | Low | Medium | Add EXPLAIN ANALYZE tests, monitor query performance |

### Low Risk Items

| Risk | Probability | Impact | Mitigation |
|------|------------|--------|------------|
| Test coverage insufficient | Medium | Low | Set coverage threshold in CI/CD, require reviews |
| Documentation outdated | Medium | Low | Update docs as part of PR process, automated checks |

---

## Success Metrics

### Quantitative Metrics

1. **Code Quality**
   - Test Coverage: >80% for domain and repository layers
   - Linting Score: 0 errors, 0 warnings
   - Build Time: <30 seconds for full backend

2. **Database Performance**
   - Migration Time: <5 seconds for both UP migrations
   - Query Performance: <100ms for single record retrieval
   - Index Effectiveness: >95% index usage in queries

3. **Development Velocity**
   - Phase Completion: Within 1 week (40 hours)
   - Rework Rate: <10% of initial implementation
   - PR Review Cycles: <2 iterations per component

### Qualitative Metrics

1. **Architectural Alignment**
   - Follows Clean Architecture principles
   - Consistent with existing codebase patterns
   - Minimal coupling between layers

2. **Maintainability**
   - Code is self-documenting with clear naming
   - Complex logic has inline comments
   - Error messages are actionable

3. **Reliability**
   - All migrations are reversible
   - No data loss scenarios
   - Graceful error handling

---

## Required Resources and Dependencies

### Technical Requirements

**Development Environment**:
- Go 1.21+ installed
- PostgreSQL 15+ running locally
- Redis 6+ running (for existing auth system)
- migrate CLI tool installed
- golangci-lint installed

**Database Access**:
- PostgreSQL connection with CREATE TABLE privileges
- Ability to run migrations
- Backup of existing database for testing

### External Dependencies

**Go Packages** (already installed):
```
github.com/google/uuid
gorm.io/gorm
gorm.io/driver/postgres
github.com/gin-gonic/gin
```

**Testing Packages**:
```
github.com/stretchr/testify
github.com/DATA-DOG/go-sqlmock (for repository tests)
```

### Knowledge Requirements

- Understanding of PostgreSQL advanced features (CHECK constraints, CASCADE)
- GORM ORM framework patterns
- Go struct tags and reflection
- Repository pattern implementation
- Database migration best practices

---

## Timeline Estimates

### Detailed Breakdown

| Phase | Task Count | Effort | Duration |
|-------|-----------|--------|----------|
| 1.1 Database Migrations | 5 | 8h | Day 1 |
| 1.2 Domain Entities | 4 | 12h | Day 1-2 |
| 1.3 Repository Layer | 4 | 16h | Day 3-4 |
| 1.4 Integration | 4 | 6h | Day 4-5 |
| 1.5 Documentation | 3 | 2h | Day 5 |
| **Total** | **20** | **44h** | **5 days** |

### Critical Path

```
Day 1: Migrations (1.1) → Domain Entities (1.2 partial)
Day 2: Domain Entities complete (1.2) → Domain Tests
Day 3: Repository Layer (1.3) → Repository Implementation
Day 4: Repository Tests (1.3) → Integration (1.4)
Day 5: Integration Tests (1.4) → Documentation (1.5) → Phase Complete
```

### Milestones

- **Day 1 EOD**: Database migrations run successfully
- **Day 2 EOD**: Domain entities pass all unit tests
- **Day 3 EOD**: Repository layer implemented
- **Day 4 EOD**: All repository tests pass
- **Day 5 EOD**: Phase 1 complete, ready for Phase 2

---

## Validation and Testing Strategy

### Unit Testing Approach

**Domain Layer Tests**:
```go
// Example test structure
func TestMediaAsset_MarkAsCompleted(t *testing.T) {
    tests := []struct {
        name      string
        asset     *domain.MediaAsset
        videoID   string
        wantErr   bool
        wantStatus domain.SyncStatus
    }{
        {
            name: "successful completion",
            asset: &domain.MediaAsset{
                SyncStatus: domain.SyncStatusUploading,
            },
            videoID: "test-video-123",
            wantErr: false,
            wantStatus: domain.SyncStatusCompleted,
        },
        // ... more test cases
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Test logic
        })
    }
}
```

**Repository Layer Tests**:
```go
// Use sqlmock for database mocking
func TestMediaRepository_Create(t *testing.T) {
    db, mock, err := sqlmock.New()
    require.NoError(t, err)
    defer db.Close()

    gormDB, err := gorm.Open(postgres.New(postgres.Config{
        Conn: db,
    }), &gorm.Config{})
    require.NoError(t, err)

    repo := repository.NewMediaRepository(gormDB)

    // Test create operation with mock expectations
}
```

### Integration Testing Approach

**Database Integration Tests**:
```go
// Use testcontainers for real PostgreSQL
func TestMediaRepository_Integration(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping integration test")
    }

    // Start PostgreSQL container
    postgres, err := testcontainers.GenericContainer(...)
    require.NoError(t, err)
    defer postgres.Terminate(context.Background())

    // Run tests against real database
}
```

### Manual Testing Checklist

- [ ] Run migrations on fresh database
- [ ] Create media asset record manually
- [ ] Verify foreign key constraints work
- [ ] Test CASCADE DELETE (delete user, verify assets deleted)
- [ ] Verify indexes created (check with `\d+ media_assets`)
- [ ] Test concurrent inserts (no UUID conflicts)
- [ ] Verify GORM AutoMigrate matches manual migrations

---

## Handoff to Phase 2

### Prerequisites for Phase 2

- ✅ All Phase 1 tasks completed
- ✅ All tests passing (unit + integration)
- ✅ Database migrations deployed
- ✅ Code merged to main branch
- ✅ Documentation updated

### Phase 2 Preview

**Phase 2: YouTube Integration** will build on this foundation:
- Extend YouTubeService with resumable upload
- Create UploadService (uses MediaRepository and SessionRepository)
- Implement MediaHandler with 6 API endpoints
- Add route configuration
- Integration testing for upload flow

**Key Interfaces Phase 2 Will Use**:
- `MediaRepository.Create()` - save uploaded media metadata
- `MediaRepository.FindPendingUploads()` - retry failed uploads
- `UploadSessionRepository.Create()` - track upload progress
- `UploadSessionRepository.Update()` - update progress

---

## Appendices

### A. File Structure After Phase 1

```
backend/
├── cmd/api/main.go                          [MODIFIED]
├── internal/
│   ├── domain/
│   │   ├── user.go                          [EXISTING]
│   │   ├── token.go                         [EXISTING]
│   │   ├── errors.go                        [MODIFIED]
│   │   ├── media_asset.go                   [NEW]
│   │   ├── media_asset_test.go              [NEW]
│   │   ├── upload_session.go                [NEW]
│   │   └── upload_session_test.go           [NEW]
│   ├── repository/
│   │   ├── interfaces.go                    [MODIFIED]
│   │   ├── user_repository.go               [EXISTING]
│   │   ├── token_repository.go              [EXISTING]
│   │   ├── media_repository.go              [NEW]
│   │   ├── media_repository_test.go         [NEW]
│   │   ├── session_repository.go            [NEW]
│   │   └── session_repository_test.go       [NEW]
│   └── pkg/database/postgres.go             [MODIFIED]
├── migrations/
│   ├── 000001_create_users_table.*.sql      [EXISTING]
│   ├── 000002_create_user_tokens_table.*.sql [EXISTING]
│   ├── 000003_create_media_assets_table.up.sql    [NEW]
│   ├── 000003_create_media_assets_table.down.sql  [NEW]
│   ├── 000004_create_upload_sessions_table.up.sql [NEW]
│   └── 000004_create_upload_sessions_table.down.sql [NEW]
└── tests/integration/
    └── media_test.go                        [NEW]
```

### B. SQL Query Examples for Testing

```sql
-- Verify media_assets table structure
SELECT column_name, data_type, is_nullable, column_default
FROM information_schema.columns
WHERE table_name = 'media_assets'
ORDER BY ordinal_position;

-- Verify indexes
SELECT indexname, indexdef
FROM pg_indexes
WHERE tablename = 'media_assets';

-- Test foreign key constraint
INSERT INTO media_assets (user_id, original_filename, file_size_bytes, media_type)
VALUES ('00000000-0000-0000-0000-000000000000', 'test.mp4', 1000, 'VIDEO');
-- Should fail with foreign key violation

-- Verify CASCADE DELETE works
DELETE FROM users WHERE user_id = 'valid-uuid';
-- Should also delete related media_assets
```

### C. Code Review Checklist

**Before Submitting PR**:
- [ ] All tests pass locally
- [ ] Migrations tested (up and down)
- [ ] golangci-lint passes with no errors
- [ ] Code follows existing patterns (check user.go, user_repository.go)
- [ ] GORM tags match database schema exactly
- [ ] Error handling comprehensive
- [ ] No TODO comments in production code
- [ ] Documentation updated (README, inline comments)
- [ ] Test coverage >80% (check with `go test -cover`)

**During Code Review**:
- [ ] Domain entities follow DDD principles
- [ ] Repository pattern correctly implemented
- [ ] No business logic in repository layer
- [ ] Proper separation of concerns
- [ ] Error messages are user-friendly
- [ ] SQL injection prevented (using GORM parameterization)
- [ ] Race conditions considered (especially for sessions)

---

**End of Plan Document**

**Last Updated**: 2026-03-24
**Next Review**: After Phase 1 completion
**Plan Owner**: Development Team
**Approved By**: Pending implementation start
