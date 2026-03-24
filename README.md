# 유튜브 기반 무제한 미디어 자동 백업 시스템

스마트폰의 디스크 공간 부족 문제를 해결하는 자동 미디어 백업 애플리케이션입니다. 대용량 영상을 유튜브 비공개 채널로 자동 업로드하고 원본을 안전하게 삭제하여 무제한 저장 공간을 제공합니다.

## 📋 프로젝트 개요

- **영상 백업**: 유튜브 비공개 채널 (무료, 무제한)
- **사진 백업**: Oracle Cloud Object Storage (200GB Free Tier)
- **모바일**: Flutter (iOS/Android)
- **백엔드**: Go (Golang)
- **데이터베이스**: PostgreSQL

## 🏗️ 프로젝트 구조

```
video_upload_app/
├── docs/                           # PDCA 문서
│   ├── 01-plan/
│   │   └── features/
│   │       └── media-backup-system.plan.md
│   ├── 02-design/
│   ├── 03-analysis/
│   └── 04-report/
│
├── mobile/                         # Flutter 모바일 앱 (생성 예정)
│   ├── lib/
│   │   ├── features/
│   │   ├── core/
│   │   ├── data/
│   │   └── services/
│   ├── test/
│   └── pubspec.yaml
│
├── backend/                        # Go 백엔드 API (생성 예정)
│   ├── cmd/
│   │   └── api/
│   ├── internal/
│   │   ├── handler/
│   │   ├── service/
│   │   ├── repository/
│   │   └── domain/
│   ├── migrations/
│   └── go.mod
│
├── .claude/                        # Claude Code 설정
│   └── skills/
│       ├── flutter-mobile-guidelines/
│       ├── flutter-testing-guidelines/
│       ├── go-backend-guidelines/
│       └── go-testing-guidelines/
│
└── README.md
```

## 🚀 빠른 시작

### 사전 요구사항

#### 모바일 개발
- Flutter SDK 3.16.0+
- Dart 3.2.0+
- Android Studio / Xcode

#### 백엔드 개발
- Go 1.21+
- PostgreSQL 15+
- Docker (선택사항)

### 설치 방법

#### 1. Flutter 프로젝트 생성
```bash
# mobile 디렉토리 생성
flutter create mobile
cd mobile

# 의존성 설치
flutter pub get

# 분석 및 포맷 검증
flutter analyze
dart format .
```

#### 2. Go 백엔드 생성
```bash
# backend 디렉토리 생성
mkdir -p backend/cmd/api backend/internal/{handler,service,repository,domain}

cd backend

# Go 모듈 초기화
go mod init github.com/yourusername/video-upload-backend

# 필수 패키지 설치
go get -u github.com/gin-gonic/gin
go get -u gorm.io/gorm
go get -u gorm.io/driver/postgres
go get -u github.com/google/uuid
go get -u google.golang.org/api/youtube/v3
```

#### 3. PostgreSQL 데이터베이스 설정
```bash
# Docker로 PostgreSQL 실행 (선택사항)
docker run --name video-backup-db \
  -e POSTGRES_PASSWORD=yourpassword \
  -e POSTGRES_DB=video_backup \
  -p 5432:5432 \
  -d postgres:15-alpine

# 또는 로컬 PostgreSQL 사용
createdb video_backup
```

## 📚 개발 가이드

### PDCA 워크플로우

이 프로젝트는 PDCA (Plan-Do-Check-Act) 사이클을 따릅니다:

1. **Plan**: `/pdca plan media-backup-system` ✅ 완료
2. **Design**: `/pdca design media-backup-system` (다음 단계)
3. **Do**: 구현 시작
4. **Check**: 갭 분석 `/pdca analyze media-backup-system`
5. **Act**: 개선 반복 `/pdca iterate media-backup-system`

### 코드 품질 규칙

#### Flutter
- **린트**: `flutter analyze` (에러 0개 필수)
- **포맷**: `dart format .` (자동 포맷팅)
- **테스트**: `flutter test --coverage` (커버리지 80% 이상)
- **컨벤션**: `analysis_options.yaml` 준수

#### Go
- **린트**: `golangci-lint run` (에러 0개 필수)
- **포맷**: `gofmt -s -w .`
- **테스트**: `go test -cover ./...` (커버리지 80% 이상)
- **컨벤션**: Clean Architecture 패턴

### 테스트 실행

#### Flutter
```bash
# 전체 테스트 실행
flutter test

# 커버리지 포함
flutter test --coverage

# 특정 테스트
flutter test test/features/media/media_provider_test.dart

# 통합 테스트
flutter drive --target=integration_test/app_flow_test.dart
```

#### Go
```bash
# 전체 테스트 실행
go test ./...

# 커버리지 포함
go test -cover ./...

# 상세 출력
go test -v ./...

# Race 감지
go test -race ./...
```

## 🔧 환경 설정

### Flutter 환경 변수
```bash
# .env.development
API_BASE_URL=http://localhost:8080/api/v1
YOUTUBE_API_KEY=your_youtube_api_key
GOOGLE_OAUTH_CLIENT_ID=your_google_oauth_client_id
```

### Go 환경 변수
```bash
# .env
DATABASE_URL=postgresql://user:password@localhost:5432/video_backup
JWT_SECRET=your_jwt_secret_key
GOOGLE_CLIENT_ID=your_google_oauth_client_id
GOOGLE_CLIENT_SECRET=your_google_oauth_client_secret
YOUTUBE_API_KEY=your_youtube_api_key
ORACLE_S3_ENDPOINT=https://objectstorage.ap-seoul-1.oraclecloud.com
ORACLE_S3_ACCESS_KEY=your_s3_access_key
ORACLE_S3_SECRET_KEY=your_s3_secret_key
```

## 📝 API 문서

### 주요 엔드포인트 (예정)

#### Authentication
- `POST /api/v1/auth/google` - Google OAuth 로그인
- `POST /api/v1/auth/refresh` - JWT 토큰 갱신

#### Media Upload
- `POST /api/v1/media/upload/initiate` - 업로드 세션 시작
- `POST /api/v1/media/upload/video` - 영상 업로드
- `GET /api/v1/media/upload/status/:session_id` - 업로드 상태 조회

#### Media Management
- `GET /api/v1/media/list` - 백업 미디어 목록
- `GET /api/v1/media/:asset_id` - 미디어 상세 정보
- `DELETE /api/v1/media/:asset_id` - 백업 삭제

## 🧪 개발 스킬 가이드

프로젝트에는 다음 개발 스킬이 설정되어 있습니다:

- **flutter-mobile-guidelines**: Flutter 모바일 개발 가이드
- **flutter-testing-guidelines**: Flutter 테스팅 가이드
- **go-backend-guidelines**: Go 백엔드 개발 가이드
- **go-testing-guidelines**: Go 테스팅 가이드

Claude Code를 사용하여 이 스킬들을 자동으로 적용할 수 있습니다.

## 📊 진행 상황

### Phase 1: Foundation (Week 1-2)
- [x] 프로젝트 구조 설정
- [x] PDCA Plan 문서 작성
- [x] Flutter/Go 스킬 생성
- [ ] Google OAuth 2.0 인증 구현
- [ ] YouTube API 연동 테스트
- [ ] PostgreSQL 스키마 생성

### Phase 2: Core Upload (Week 3-4)
- [ ] Flutter 파일 시스템 접근
- [ ] 영상 파일 자동 감지
- [ ] 업로드 큐 관리
- [ ] 유튜브 비공개 업로드
- [ ] 업로드 진행 상태 UI

### Phase 3: Background & Sync (Week 5-6)
- [ ] WorkManager 백그라운드 서비스
- [ ] 네트워크 상태 감지
- [ ] 재시도 로직
- [ ] 원본 파일 안전 삭제

### Phase 4: Polish & Testing (Week 7-8)
- [ ] UI/UX 개선
- [ ] E2E 테스트
- [ ] 성능 최적화
- [ ] 베타 테스트

## 🛠️ 개발 명령어

### Makefile 명령어 (생성 예정)

```bash
# 백엔드
make backend-lint       # Go 린트 실행
make backend-test       # Go 테스트 실행
make backend-build      # 백엔드 빌드
make backend-run        # 백엔드 실행

# 모바일
make mobile-lint        # Flutter 린트 실행
make mobile-test        # Flutter 테스트 실행
make mobile-build-apk   # Android APK 빌드
make mobile-build-ios   # iOS 빌드

# 전체
make lint               # 전체 린트 실행
make test               # 전체 테스트 실행
```

## 🐛 문제 해결

### Flutter 관련
- **분석 에러**: `flutter clean && flutter pub get`
- **빌드 실패**: `flutter doctor` 로 환경 확인
- **테스트 실패**: Mock 설정 확인

### Go 관련
- **컴파일 에러**: `go mod tidy` 실행
- **테스트 실패**: 데이터베이스 연결 확인
- **린트 에러**: `golangci-lint run --fix`

## 📄 라이선스

MIT License

## 👥 기여

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## 📞 연락처

프로젝트 관리자: [Your Name]
이메일: [your.email@example.com]
GitHub: [https://github.com/yourusername/video-upload-app](https://github.com/yourusername/video-upload-app)

---

**마지막 업데이트**: 2026-03-24
**현재 단계**: Plan Phase 완료 ✅
**다음 단계**: `/pdca design media-backup-system` 실행
