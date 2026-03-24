# Plan: 유튜브 기반 무제한 미디어 자동 백업 시스템

## 1. Executive Summary

**Feature Name**: Media Backup System (미디어 백업 시스템)
**Feature Type**: Core Product
**Priority**: High
**Estimated Complexity**: High

**Overview**:
스마트폰의 디스크 공간 부족 문제를 해결하기 위한 자동 미디어 백업 시스템. 대용량 영상을 유튜브 비공개 채널로 자동 업로드하고, 사진은 S3 호환 스토리지(오라클 클라우드 Free Tier)에 저장하여 원본을 안전하게 삭제함으로써 무제한 저장 공간을 제공합니다.

**Target Users**:
- 매일 기가바이트 단위의 고화질 영상을 촬영하는 사용자
- 반려동물, 가족 등 일상을 지속적으로 기록하는 사용자
- 스마트폰 저장 공간 부족으로 고민하는 사용자

---

## 2. Problem Statement

### Current Pain Points
1. **디스크 공간 부족**: 고화질 영상 촬영으로 인한 빠른 저장 공간 소진
2. **수동 백업의 번거로움**: 정기적인 백업 작업의 시간 소요 및 불편함
3. **클라우드 비용 부담**: 기존 클라우드 서비스의 저장 용량 제한 및 과금
4. **원본 삭제 불안감**: 백업 검증 없이 원본을 삭제하는 것에 대한 두려움

### Business Impact
- 사용자의 일상 기록 활동이 저장 공간 제약으로 제한됨
- 수동 백업 프로세스로 인한 사용자 경험 저하
- 클라우드 저장 비용으로 인한 서비스 지속성 문제

---

## 3. Goals & Success Metrics

### Primary Goals
1. **자동화된 백업**: 사용자 개입 없이 영상/사진 자동 업로드
2. **무제한 저장**: 유튜브 비공개 채널 활용으로 무료 무제한 영상 저장
3. **안전한 원본 삭제**: 백업 검증 후 원본 파일 자동 삭제
4. **저장 공간 확보**: 스마트폰 디스크 공간 자동 관리

### Success Metrics
- **업로드 성공률**: 98% 이상
- **동기화 완료 시간**: 1GB 영상 기준 WiFi 환경에서 10분 이내
- **저장 공간 확보율**: 원본 대비 90% 이상 공간 확보
- **사용자 만족도**: 4.5/5.0 이상

### Key Performance Indicators (KPIs)
- Daily Active Upload Users
- Total Media Size Uploaded (GB/day)
- Storage Space Freed (GB/user/month)
- Upload Success Rate
- Average Upload Time per GB

---

## 4. Technical Scope

### 4.1 Technology Stack

#### Mobile Client
- **Framework**: Flutter (Dart)
- **State Management**: Provider 또는 Riverpod
- **Local Database**: SQLite (sqflite 패키지)
- **HTTP Client**: dio 패키지
- **Background Processing**: workmanager 패키지

#### Backend API
- **Language**: Go (Golang) 1.21+
- **Web Framework**: Gin 또는 Echo
- **Database**: PostgreSQL 15+
- **ORM**: GORM
- **Authentication**: JWT (golang-jwt/jwt)
- **YouTube API**: google.golang.org/api/youtube/v3
- **S3 SDK**: aws-sdk-go-v2 (Oracle Cloud S3 호환)

#### Infrastructure
- **Database**: PostgreSQL (Managed Service 권장)
- **Storage**:
  - Video: YouTube Data API v3 (무제한, 무료)
  - Image: Oracle Cloud Object Storage (200GB Free Tier)
- **Deployment**: Docker + Kubernetes 또는 단일 서버 배포

### 4.2 Core Features

#### Phase 1: MVP (Minimum Viable Product)
1. **사용자 인증**
   - Google OAuth 2.0 로그인
   - 유튜브 API 권한 획득 (youtube.upload 스코프)
   - JWT 기반 세션 관리

2. **영상 자동 백업**
   - 기기 내 영상 파일 자동 감지
   - WiFi 연결 시 유튜브 비공개 업로드
   - 업로드 진행 상태 표시
   - 업로드 완료 후 원본 삭제

3. **백업 상태 관리**
   - 업로드 대기 큐 관리
   - 실패한 업로드 재시도 로직
   - 백업 히스토리 조회

#### Phase 2: Enhanced Features (선택 확장)
1. **사진 백업**
   - S3 호환 스토리지 연동 (Oracle Cloud)
   - 이미지 압축 옵션
   - 앨범별 분류 저장

2. **스마트 백업 설정**
   - 업로드 시간대 설정 (야간 자동 업로드)
   - 네트워크 조건 설정 (WiFi 전용 / 데이터 허용)
   - 파일 크기 필터 (최소 크기 설정)

3. **복원 기능**
   - 백업된 영상 다운로드
   - 선택적 복원 (특정 날짜/파일)

### 4.3 Database Schema

#### users 테이블
```sql
CREATE TABLE users (
    user_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    google_account_id VARCHAR(255) UNIQUE NOT NULL,
    email VARCHAR(255) NOT NULL,
    youtube_channel_id VARCHAR(255),
    refresh_token TEXT, -- YouTube API 갱신 토큰
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    last_login_at TIMESTAMP
);

CREATE INDEX idx_users_google_account ON users(google_account_id);
```

#### media_assets 테이블
```sql
CREATE TABLE media_assets (
    asset_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
    youtube_video_id VARCHAR(255) UNIQUE, -- 영상의 경우
    s3_object_key VARCHAR(512), -- 사진의 경우
    original_filename VARCHAR(512) NOT NULL,
    file_size_bytes BIGINT NOT NULL,
    media_type VARCHAR(10) NOT NULL CHECK (media_type IN ('VIDEO', 'IMAGE')),
    sync_status VARCHAR(20) NOT NULL DEFAULT 'PENDING'
        CHECK (sync_status IN ('PENDING', 'UPLOADING', 'COMPLETED', 'FAILED')),
    upload_started_at TIMESTAMP,
    upload_completed_at TIMESTAMP,
    error_message TEXT,
    retry_count INT DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_media_assets_user ON media_assets(user_id);
CREATE INDEX idx_media_assets_status ON media_assets(sync_status);
CREATE INDEX idx_media_assets_created ON media_assets(created_at DESC);
```

#### upload_sessions 테이블 (업로드 세션 추적)
```sql
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
    started_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    completed_at TIMESTAMP
);

CREATE INDEX idx_upload_sessions_user ON upload_sessions(user_id);
CREATE INDEX idx_upload_sessions_status ON upload_sessions(session_status);
```

### 4.4 API Endpoints

#### Authentication
- `POST /api/v1/auth/google` - Google OAuth 로그인
- `POST /api/v1/auth/refresh` - JWT 토큰 갱신
- `GET /api/v1/auth/youtube/status` - 유튜브 연동 상태 확인

#### Media Upload
- `POST /api/v1/media/upload/initiate` - 업로드 세션 시작
- `POST /api/v1/media/upload/video` - 영상 업로드 (멀티파트)
- `POST /api/v1/media/upload/image` - 사진 업로드 (멀티파트)
- `GET /api/v1/media/upload/status/:session_id` - 업로드 진행 상태 조회
- `POST /api/v1/media/upload/complete` - 업로드 완료 알림

#### Media Management
- `GET /api/v1/media/list` - 백업된 미디어 목록 조회 (페이지네이션)
- `GET /api/v1/media/:asset_id` - 특정 미디어 상세 정보
- `DELETE /api/v1/media/:asset_id` - 백업 삭제 (원본 삭제 아님)
- `POST /api/v1/media/:asset_id/restore` - 미디어 복원 요청

#### User Settings
- `GET /api/v1/settings` - 사용자 설정 조회
- `PUT /api/v1/settings` - 사용자 설정 업데이트

---

## 5. Architecture Overview

### 5.1 System Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                     Flutter Mobile App                       │
│  ┌────────────────┐  ┌────────────────┐  ┌────────────────┐ │
│  │  UI Layer      │  │  State Mgmt    │  │  Local DB      │ │
│  │  (Widgets)     │  │  (Provider)    │  │  (SQLite)      │ │
│  └────────────────┘  └────────────────┘  └────────────────┘ │
│           │                   │                   │          │
│  ┌────────────────────────────────────────────────────────┐ │
│  │         Background Service (WorkManager)                │ │
│  │  - File Watcher  - Upload Queue  - Retry Logic         │ │
│  └────────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────┘
                              │ HTTPS/REST API
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                      Go Backend API                          │
│  ┌────────────────┐  ┌────────────────┐  ┌────────────────┐ │
│  │  HTTP Handlers │  │  Service Layer │  │  Repository    │ │
│  │  (Gin/Echo)    │  │  (Business)    │  │  (GORM)        │ │
│  └────────────────┘  └────────────────┘  └────────────────┘ │
│           │                   │                   │          │
│  ┌────────────────────────────────────────────────────────┐ │
│  │         Middleware (Auth, Logging, Error)               │ │
│  └────────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────┘
                │                              │
                ▼                              ▼
    ┌──────────────────────┐      ┌──────────────────────┐
    │  YouTube Data API v3 │      │  Oracle Cloud S3     │
    │  (Video Upload)      │      │  (Image Storage)     │
    └──────────────────────┘      └──────────────────────┘
                │
                ▼
    ┌──────────────────────┐
    │  PostgreSQL Database │
    │  (Metadata Storage)  │
    └──────────────────────┘
```

### 5.2 Data Flow

#### Upload Flow (영상)
1. **File Detection**: 앱이 새 영상 파일 감지 (FileSystemWatcher)
2. **Queue Addition**: 로컬 SQLite에 업로드 대기 항목 추가
3. **Network Check**: WiFi 연결 확인 (설정에 따라 데이터 네트워크 허용)
4. **Upload Initiation**: 백엔드 API에 업로드 세션 시작 요청
5. **Chunked Upload**: 영상 파일을 청크 단위로 유튜브에 업로드
6. **Metadata Storage**: 업로드 완료 후 `media_assets` 테이블에 메타데이터 저장
7. **Verification**: 유튜브 비디오 ID 확인 및 상태 업데이트
8. **Local Deletion**: 백엔드 확인 후 앱에서 원본 파일 삭제

#### Error Handling Flow
1. **Upload Failure**: 네트워크 오류, API 에러 등
2. **Retry Queue**: 실패한 항목을 재시도 큐에 추가 (최대 5회)
3. **Exponential Backoff**: 재시도 간격 점진적 증가 (1분, 5분, 15분, 1시간, 24시간)
4. **User Notification**: 최종 실패 시 사용자에게 알림

---

## 6. Development Phases

### Phase 1: Foundation (Week 1-2)
- [ ] 프로젝트 구조 설정 (Flutter + Go)
- [ ] PostgreSQL 데이터베이스 스키마 생성
- [ ] Google OAuth 2.0 인증 구현
- [ ] YouTube API 연동 테스트
- [ ] 기본 CRUD API 엔드포인트 구현

### Phase 2: Core Upload (Week 3-4)
- [ ] Flutter 파일 시스템 접근 권한 처리
- [ ] 영상 파일 자동 감지 기능
- [ ] 업로드 큐 관리 로직
- [ ] 유튜브 비공개 업로드 구현
- [ ] 업로드 진행 상태 UI

### Phase 3: Background & Sync (Week 5-6)
- [ ] WorkManager 기반 백그라운드 서비스
- [ ] 네트워크 상태 감지 및 조건부 업로드
- [ ] 재시도 로직 및 에러 핸들링
- [ ] 원본 파일 안전 삭제 로직

### Phase 4: Polish & Testing (Week 7-8)
- [ ] UI/UX 개선
- [ ] 종단 간 테스트 (E2E)
- [ ] 성능 최적화
- [ ] 베타 테스트 및 피드백 수렴

### Phase 5: Optional Extensions (Week 9+)
- [ ] 사진 백업 (S3 연동)
- [ ] 복원 기능
- [ ] 고급 설정 (시간대, 필터)

---

## 7. Quality Assurance Plan

### 7.1 Testing Strategy

#### Unit Tests
- **Backend (Go)**:
  - Service layer 비즈니스 로직 테스트 (coverage >80%)
  - Repository layer 데이터베이스 연동 테스트
  - Middleware 테스트 (인증, 에러 핸들링)
  - Testing Framework: `testing` package + `testify`

- **Mobile (Flutter)**:
  - Widget 테스트 (UI 컴포넌트)
  - 비즈니스 로직 테스트 (Provider/State)
  - SQLite 데이터베이스 테스트
  - Testing Framework: `flutter_test`

#### Integration Tests
- API 엔드포인트 통합 테스트 (httptest)
- YouTube API 모의 연동 테스트
- 데이터베이스 트랜잭션 테스트

#### E2E Tests
- 로그인부터 업로드 완료까지 전체 플로우
- 네트워크 오류 시나리오 테스트
- 재시도 로직 검증

### 7.2 Code Quality Tools

#### Backend (Go)
- **Linter**: `golangci-lint` (gofmt, govet, staticcheck 포함)
- **Configuration**: `.golangci.yml`
```yaml
linters:
  enable:
    - gofmt
    - govet
    - staticcheck
    - errcheck
    - gosimple
    - ineffassign
linters-settings:
  govet:
    check-shadowing: true
  gofmt:
    simplify: true
```

- **Pre-commit Hook**: `make lint` 자동 실행
- **CI/CD**: GitHub Actions에서 자동 lint 검증

#### Mobile (Flutter)
- **Linter**: `flutter analyze` (analysis_options.yaml)
- **Formatter**: `dart format`
- **Configuration**: `analysis_options.yaml`
```yaml
linter:
  rules:
    - always_declare_return_types
    - avoid_print
    - prefer_const_constructors
    - use_key_in_widget_constructors
```

- **Pre-commit Hook**: `flutter analyze && dart format --set-exit-if-changed .`

### 7.3 Continuous Integration

#### GitHub Actions Workflow
```yaml
name: CI

on: [push, pull_request]

jobs:
  backend-test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
        with:
          go-version: '1.21'
      - run: make lint
      - run: make test
      - run: make build

  mobile-test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: subosito/flutter-action@v2
        with:
          flutter-version: '3.16.0'
      - run: flutter pub get
      - run: flutter analyze
      - run: flutter test
```

### 7.4 Code Review Guidelines
- 모든 PR은 최소 1명의 리뷰어 승인 필요
- 린트 에러 0건 (CI 통과 필수)
- 테스트 커버리지 80% 이상 유지
- 변경 사항에 대한 명확한 설명 (커밋 메시지, PR 설명)

---

## 8. Security Considerations

### 8.1 Authentication & Authorization
- Google OAuth 2.0 기반 안전한 인증
- JWT 토큰 만료 시간 설정 (Access: 1시간, Refresh: 7일)
- Refresh Token은 암호화하여 데이터베이스 저장
- YouTube API 권한 최소화 (youtube.upload 스코프만)

### 8.2 Data Protection
- HTTPS 전용 통신 (TLS 1.3)
- 데이터베이스 연결 암호화
- 민감 정보 (API Key, Secret) 환경 변수 관리 (.env)
- 로그에 개인정보 노출 금지

### 8.3 API Security
- Rate Limiting (사용자당 분당 100 요청)
- CORS 정책 설정 (허용된 오리진만)
- SQL Injection 방지 (GORM 파라미터 바인딩)
- Input Validation (파일 크기, 형식 검증)

---

## 9. Risks & Mitigation

| Risk | Impact | Probability | Mitigation |
|------|--------|-------------|------------|
| YouTube API 정책 변경 | High | Medium | 정기적인 API 문서 모니터링, 대체 저장소 검토 |
| 대용량 업로드 시 네트워크 타임아웃 | Medium | High | 청크 업로드, 재시도 로직, 타임아웃 설정 |
| 원본 삭제 후 백업 손실 | High | Low | 2단계 검증 (유튜브 비디오 ID 확인 + 재생 가능 여부) |
| 배터리 소모 증가 | Medium | Medium | WiFi 전용 업로드, 충전 중 우선 업로드 옵션 |
| Google 계정 탈퇴 시 데이터 손실 | High | Low | 사용자 데이터 export 기능 제공 |

---

## 10. Deployment Plan

### 10.1 Environment Setup

#### Development
- Local PostgreSQL (Docker Compose)
- Mock YouTube API (테스트용)
- Flutter Debug Build

#### Staging
- Managed PostgreSQL (AWS RDS 또는 Digital Ocean)
- 실제 YouTube API 연동 (테스트 채널)
- Flutter Release Build (Internal Testing)

#### Production
- Managed PostgreSQL (고가용성 설정)
- 유튜브 정식 채널 연동
- Flutter Release Build (App Store / Play Store)

### 10.2 Deployment Checklist
- [ ] 환경 변수 설정 확인 (.env.production)
- [ ] 데이터베이스 마이그레이션 실행
- [ ] API 엔드포인트 헬스체크 통과
- [ ] SSL 인증서 설정
- [ ] 백업 및 롤백 계획 수립
- [ ] 모니터링 설정 (로그, 에러 트래킹)

---

## 11. Success Criteria

### Must Have (MVP)
✅ Google 계정으로 로그인 가능
✅ 유튜브 비공개 채널에 영상 자동 업로드
✅ 업로드 완료 후 원본 파일 자동 삭제
✅ 업로드 실패 시 재시도 로직 동작
✅ 백업된 영상 목록 조회 가능

### Should Have (Enhanced)
🔲 사진 백업 (S3)
🔲 업로드 시간대 설정
🔲 WiFi 전용 / 데이터 허용 선택
🔲 백업 복원 기능

### Could Have (Future)
🔲 다중 클라우드 지원 (Google Drive, OneDrive)
🔲 AI 기반 중요 영상 자동 분류
🔲 가족/친구와 백업 공유 기능

---

## 12. Next Steps

1. **Design Phase로 이동**:
   ```bash
   /pdca design media-backup-system
   ```

2. **프로젝트 초기화**:
   - Flutter 프로젝트 생성: `flutter create video_backup_app`
   - Go 백엔드 프로젝트 구조 설정
   - PostgreSQL 데이터베이스 스키마 생성

3. **Skills 생성**:
   - Flutter 개발 가이드 스킬
   - Go 백엔드 가이드 스킬 (이미 존재하는 경우 활용)

4. **개발 환경 설정**:
   - Google Cloud Console에서 OAuth 2.0 클라이언트 생성
   - YouTube Data API v3 활성화
   - Oracle Cloud 계정 생성 (Free Tier)

---

**Document Version**: 1.0
**Last Updated**: 2026-03-24
**Owner**: Development Team
**Status**: Ready for Design Phase
