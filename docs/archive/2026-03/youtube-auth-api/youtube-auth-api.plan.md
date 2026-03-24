# Plan: YouTube 계정 로그인 API

## 1. Feature Overview

### 1.1 Feature Name
**YouTube 계정 인증 및 로그인 API**

### 1.2 Feature Description
유튜브 계정별 사용자 인증을 위한 백엔드 API 구현. Google OAuth 2.0을 사용하여 사용자가 자신의 유튜브 계정으로 로그인하고, 해당 계정으로 영상을 업로드할 수 있도록 인증 토큰을 관리합니다.

### 1.3 Business Value
- **사용자별 유튜브 계정 관리**: 각 사용자가 자신의 유튜브 계정에 영상을 비공개로 업로드
- **보안성**: OAuth 2.0 표준을 사용한 안전한 인증
- **확장성**: 다중 계정 지원으로 여러 유튜브 계정 관리 가능
- **자동화**: 액세스 토큰 자동 갱신으로 끊김 없는 서비스 제공

### 1.4 Target Users
- 스마트폰 저장 공간 부족 문제를 해결하려는 일반 사용자
- 대용량 영상을 안전하게 백업하고자 하는 사용자
- 자동 백업 시스템을 원하는 사용자

## 2. Requirements

### 2.1 Functional Requirements

#### FR-01: Google OAuth 2.0 인증
- Google OAuth 2.0 웹 서버 플로우를 사용한 사용자 인증
- YouTube Data API v3 스코프 권한 요청
- 사용자 동의 화면 제공
- 인증 코드를 액세스 토큰으로 교환

**Acceptance Criteria**:
- [ ] Google OAuth 2.0 클라이언트 설정 완료
- [ ] 인증 URL 생성 API 제공 (`GET /api/v1/auth/google/url`)
- [ ] 인증 코드를 액세스 토큰으로 교환하는 콜백 API (`POST /api/v1/auth/google/callback`)
- [ ] 필요한 YouTube 스코프 포함 (`https://www.googleapis.com/auth/youtube.upload`, `https://www.googleapis.com/auth/youtube`)

#### FR-02: 액세스 토큰 및 리프레시 토큰 저장
- 사용자별 OAuth 토큰 안전하게 저장
- 리프레시 토큰을 사용한 액세스 토큰 자동 갱신
- 암호화된 토큰 저장

**Acceptance Criteria**:
- [ ] PostgreSQL에 사용자 인증 정보 저장
- [ ] 액세스 토큰, 리프레시 토큰, 만료 시간 저장
- [ ] 토큰 암호화 처리 (AES-256)
- [ ] 토큰 만료 시 자동 갱신 로직

#### FR-03: JWT 기반 세션 관리
- JWT 토큰을 사용한 API 인증
- 액세스 토큰 (짧은 수명: 7일)
- 리프레시 토큰 (긴 수명: 365일)

**Acceptance Criteria**:
- [ ] JWT 액세스 토큰 생성 및 검증
- [ ] JWT 리프레시 토큰 생성 및 검증
- [ ] 토큰 갱신 API (`POST /api/v1/auth/refresh`)
- [ ] 토큰 만료 처리 (401 Unauthorized)

#### FR-04: 사용자 정보 조회
- 인증된 사용자의 유튜브 채널 정보 조회
- 사용자 프로필 정보 반환

**Acceptance Criteria**:
- [ ] 현재 로그인한 사용자 정보 조회 API (`GET /api/v1/auth/me`)
- [ ] 유튜브 채널 ID, 이름, 프로필 이미지 반환
- [ ] 사용자 이메일 반환

#### FR-05: 로그아웃
- 사용자 세션 무효화
- JWT 토큰 블랙리스트 처리

**Acceptance Criteria**:
- [ ] 로그아웃 API (`POST /api/v1/auth/logout`)
- [ ] JWT 토큰 블랙리스트에 추가
- [ ] Redis를 사용한 토큰 블랙리스트 관리

### 2.2 Non-Functional Requirements

#### NFR-01: 보안
- **토큰 암호화**: OAuth 토큰은 AES-256으로 암호화하여 저장
- **HTTPS 전용**: 모든 인증 API는 HTTPS로만 접근 가능
- **CSRF 보호**: State 파라미터를 사용한 CSRF 공격 방어
- **Rate Limiting**: 인증 API는 IP당 분당 10회 제한

#### NFR-02: 성능
- **응답 시간**: 인증 API 응답 시간 < 500ms (P95)
- **동시 접속**: 1,000 동시 사용자 지원
- **토큰 캐싱**: 액세스 토큰을 Redis에 캐싱하여 빠른 검증

#### NFR-03: 가용성
- **업타임**: 99.9% 가용성 목표
- **장애 복구**: 데이터베이스 연결 실패 시 자동 재시도 (3회)
- **모니터링**: 인증 실패율, 토큰 갱신 성공률 모니터링

#### NFR-04: 확장성
- **수평 확장**: Stateless 아키텍처로 서버 수평 확장 가능
- **데이터베이스 인덱싱**: 사용자 ID, 이메일에 인덱스 적용

### 2.3 Technical Requirements

#### Tech Stack
- **Language**: Go 1.21+
- **Framework**: Gin Web Framework
- **Database**: PostgreSQL 15+
- **Cache**: Redis 7+
- **OAuth Library**: `golang.org/x/oauth2`
- **YouTube API**: `google.golang.org/api/youtube/v3`
- **JWT Library**: `github.com/golang-jwt/jwt/v5`
- **Encryption**: `crypto/aes` (AES-256-GCM)

#### Database Schema
```sql
-- users 테이블
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) UNIQUE NOT NULL,
    google_id VARCHAR(255) UNIQUE NOT NULL,
    youtube_channel_id VARCHAR(255),
    youtube_channel_name VARCHAR(255),
    profile_image_url TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP
);

-- user_tokens 테이블 (OAuth 토큰)
CREATE TABLE user_tokens (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    encrypted_access_token TEXT NOT NULL,
    encrypted_refresh_token TEXT NOT NULL,
    token_type VARCHAR(50) NOT NULL DEFAULT 'Bearer',
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    UNIQUE(user_id)
);

-- 인덱스
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_google_id ON users(google_id);
CREATE INDEX idx_user_tokens_user_id ON user_tokens(user_id);
CREATE INDEX idx_user_tokens_expires_at ON user_tokens(expires_at);
```

#### Environment Variables
```bash
# Google OAuth
GOOGLE_CLIENT_ID=your_google_oauth_client_id
GOOGLE_CLIENT_SECRET=your_google_oauth_client_secret
GOOGLE_REDIRECT_URL=http://localhost:8080/api/v1/auth/google/callback

# JWT
JWT_SECRET=your_jwt_secret_key
JWT_ACCESS_TOKEN_EXPIRY=15m
JWT_REFRESH_TOKEN_EXPIRY=168h

# Database
DATABASE_URL=postgresql://user:password@localhost:5432/video_backup

# Redis
REDIS_URL=redis://localhost:6379/0

# Encryption
ENCRYPTION_KEY=your_32_byte_encryption_key

# Server
PORT=8080
ENV=development
```

## 3. Architecture

### 3.1 System Architecture

```
┌─────────────┐
│   Flutter   │
│  Mobile App │
└──────┬──────┘
       │ HTTPS
       ├─ POST /api/v1/auth/google/url
       ├─ POST /api/v1/auth/google/callback
       ├─ POST /api/v1/auth/refresh
       ├─ GET  /api/v1/auth/me
       └─ POST /api/v1/auth/logout
       ↓
┌──────────────────────────────────────┐
│         Go Backend (Gin)             │
│  ┌────────────────────────────────┐  │
│  │  Handler Layer                 │  │
│  │  - AuthHandler                 │  │
│  └────────────┬───────────────────┘  │
│               ↓                      │
│  ┌────────────────────────────────┐  │
│  │  Service Layer                 │  │
│  │  - AuthService                 │  │
│  │  - TokenService                │  │
│  │  - YouTubeService              │  │
│  └────────────┬───────────────────┘  │
│               ↓                      │
│  ┌────────────────────────────────┐  │
│  │  Repository Layer              │  │
│  │  - UserRepository              │  │
│  │  - TokenRepository             │  │
│  └────────────┬───────────────────┘  │
└───────────────┼──────────────────────┘
                ↓
    ┌───────────────────┐
    │   PostgreSQL      │
    │   - users         │
    │   - user_tokens   │
    └───────────────────┘
                ↓
        ┌───────────────┐
        │     Redis     │
        │   (Cache)     │
        └───────────────┘
```

### 3.2 Clean Architecture Layers

#### Handler Layer (`internal/handler/auth_handler.go`)
- HTTP 요청/응답 처리
- 입력 검증
- 에러 핸들링
- JSON 직렬화/역직렬화

#### Service Layer (`internal/service/`)
- **AuthService**: 인증 비즈니스 로직
  - Google OAuth 플로우 처리
  - JWT 토큰 생성/검증
  - 사용자 등록/로그인
- **TokenService**: 토큰 관리
  - 액세스 토큰 갱신
  - 토큰 암호화/복호화
  - 토큰 블랙리스트 관리
- **YouTubeService**: YouTube API 연동
  - 채널 정보 조회
  - 사용자 프로필 조회

#### Repository Layer (`internal/repository/`)
- **UserRepository**: 사용자 데이터 CRUD
- **TokenRepository**: OAuth 토큰 데이터 CRUD

#### Domain Layer (`internal/domain/`)
- **User**: 사용자 엔티티
- **Token**: 토큰 엔티티
- **Errors**: 도메인 에러 정의

### 3.3 API Endpoints

| Method | Endpoint | Description | Auth Required |
|--------|----------|-------------|---------------|
| GET | `/api/v1/auth/google/url` | Google OAuth 인증 URL 생성 | No |
| POST | `/api/v1/auth/google/callback` | OAuth 콜백 처리 및 JWT 발급 | No |
| POST | `/api/v1/auth/refresh` | JWT 토큰 갱신 | No |
| GET | `/api/v1/auth/me` | 현재 사용자 정보 조회 | Yes |
| POST | `/api/v1/auth/logout` | 로그아웃 (토큰 무효화) | Yes |

### 3.4 Authentication Flow

```
┌────────┐                                 ┌────────┐                    ┌────────┐
│ Client │                                 │ Backend│                    │ Google │
└───┬────┘                                 └───┬────┘                    └───┬────┘
    │                                          │                             │
    │ 1. GET /auth/google/url                 │                             │
    │─────────────────────────────────────────>│                             │
    │                                          │ 2. Generate OAuth URL       │
    │                                          │────────────────────────────>│
    │ 3. Return OAuth URL                     │                             │
    │<─────────────────────────────────────────│                             │
    │                                          │                             │
    │ 4. User clicks OAuth URL (Browser)      │                             │
    │──────────────────────────────────────────────────────────────────────>│
    │                                          │                             │
    │                                          │ 5. User grants permission   │
    │                                          │                             │
    │ 6. Redirect to callback with auth code  │                             │
    │<──────────────────────────────────────────────────────────────────────│
    │                                          │                             │
    │ 7. POST /auth/google/callback {code}    │                             │
    │─────────────────────────────────────────>│                             │
    │                                          │ 8. Exchange code for tokens │
    │                                          │────────────────────────────>│
    │                                          │ 9. Return access/refresh    │
    │                                          │<────────────────────────────│
    │                                          │ 10. Fetch user profile      │
    │                                          │────────────────────────────>│
    │                                          │ 11. Return profile          │
    │                                          │<────────────────────────────│
    │                                          │ 12. Save user & tokens (DB) │
    │                                          │ 13. Generate JWT            │
    │ 14. Return JWT access/refresh tokens    │                             │
    │<─────────────────────────────────────────│                             │
    │                                          │                             │
    │ 15. Store JWT locally                   │                             │
    │                                          │                             │
    │ 16. Subsequent requests with JWT        │                             │
    │─────────────────────────────────────────>│                             │
    │                                          │ 17. Validate JWT            │
    │                                          │ 18. Return protected data   │
    │<─────────────────────────────────────────│                             │
```

## 4. Implementation Plan

### 4.1 Phase 1: 프로젝트 초기화 (Day 1)
- [ ] Go 모듈 초기화 및 디렉토리 구조 생성
- [ ] 필수 의존성 설치
  - `github.com/gin-gonic/gin`
  - `gorm.io/gorm`
  - `gorm.io/driver/postgres`
  - `github.com/golang-jwt/jwt/v5`
  - `golang.org/x/oauth2`
  - `google.golang.org/api/youtube/v3`
  - `github.com/go-redis/redis/v8`
- [ ] 환경 변수 설정 (`.env` 파일)
- [ ] PostgreSQL 데이터베이스 생성
- [ ] Redis 설치 및 실행

### 4.2 Phase 2: 데이터베이스 스키마 (Day 1)
- [ ] `users` 테이블 마이그레이션 생성
- [ ] `user_tokens` 테이블 마이그레이션 생성
- [ ] 인덱스 추가
- [ ] 마이그레이션 실행 및 검증

### 4.3 Phase 3: Domain Layer (Day 2)
- [ ] `internal/domain/user.go` - User 엔티티 정의
- [ ] `internal/domain/token.go` - Token 엔티티 정의
- [ ] `internal/domain/errors.go` - 도메인 에러 정의

### 4.4 Phase 4: Repository Layer (Day 2-3)
- [ ] `internal/repository/user_repository.go`
  - `Create(user *domain.User) error`
  - `FindByID(id string) (*domain.User, error)`
  - `FindByEmail(email string) (*domain.User, error)`
  - `FindByGoogleID(googleID string) (*domain.User, error)`
  - `Update(user *domain.User) error`
- [ ] `internal/repository/token_repository.go`
  - `Create(token *domain.Token) error`
  - `FindByUserID(userID string) (*domain.Token, error)`
  - `Update(token *domain.Token) error`
  - `Delete(userID string) error`
- [ ] Repository 인터페이스 정의
- [ ] GORM 구현체 작성

### 4.5 Phase 5: Service Layer (Day 3-5)
- [ ] `internal/service/auth_service.go`
  - Google OAuth 설정
  - `GenerateAuthURL(state string) string`
  - `HandleCallback(code string) (*domain.User, string, string, error)`
  - `GenerateJWT(userID string) (accessToken, refreshToken string, error)`
  - `ValidateJWT(token string) (*JWTClaims, error)`
  - `RefreshAccessToken(refreshToken string) (string, error)`
- [ ] `internal/service/token_service.go`
  - `EncryptToken(plainText string) (string, error)`
  - `DecryptToken(cipherText string) (string, error)`
  - `RefreshGoogleToken(userID string) error`
  - `AddToBlacklist(token string, expiry time.Duration) error`
  - `IsBlacklisted(token string) (bool, error)`
- [ ] `internal/service/youtube_service.go`
  - `GetChannelInfo(accessToken string) (*ChannelInfo, error)`
  - `GetUserProfile(accessToken string) (*UserProfile, error)`

### 4.6 Phase 6: Handler Layer (Day 5-6)
- [ ] `internal/handler/auth_handler.go`
  - `GetGoogleAuthURL(c *gin.Context)`
  - `HandleGoogleCallback(c *gin.Context)`
  - `RefreshToken(c *gin.Context)`
  - `GetCurrentUser(c *gin.Context)`
  - `Logout(c *gin.Context)`
- [ ] 입력 검증 (Validator)
- [ ] 에러 응답 표준화

### 4.7 Phase 7: Middleware (Day 6)
- [ ] `internal/middleware/auth_middleware.go`
  - JWT 토큰 검증
  - 블랙리스트 확인
  - 사용자 컨텍스트 주입
- [ ] `internal/middleware/rate_limiter.go`
  - Redis 기반 Rate Limiting
- [ ] `internal/middleware/cors.go`
  - CORS 설정

### 4.8 Phase 8: Router 설정 (Day 7)
- [ ] `cmd/api/main.go` - 서버 엔트리포인트
- [ ] `internal/router/router.go` - 라우트 설정
- [ ] Health check 엔드포인트 (`GET /health`)
- [ ] API 버저닝 (`/api/v1/*`)

### 4.9 Phase 9: 테스트 (Day 7-8)
- [ ] Unit Test (커버리지 80% 이상)
  - Repository 테스트
  - Service 테스트
  - Handler 테스트
- [ ] Integration Test
  - 인증 플로우 E2E 테스트
  - 토큰 갱신 테스트
- [ ] Mock 객체 생성 (gomock)

### 4.10 Phase 10: 문서화 (Day 8)
- [ ] API 문서 작성 (Swagger/OpenAPI)
- [ ] README 업데이트
- [ ] 환경 설정 가이드
- [ ] 로컬 개발 가이드

## 5. Testing Strategy

### 5.1 Unit Tests

#### Repository Layer
```go
// internal/repository/user_repository_test.go
func TestUserRepository_Create(t *testing.T)
func TestUserRepository_FindByEmail(t *testing.T)
func TestUserRepository_FindByGoogleID(t *testing.T)
```

#### Service Layer
```go
// internal/service/auth_service_test.go
func TestAuthService_GenerateAuthURL(t *testing.T)
func TestAuthService_HandleCallback(t *testing.T)
func TestAuthService_GenerateJWT(t *testing.T)
func TestAuthService_ValidateJWT(t *testing.T)

// internal/service/token_service_test.go
func TestTokenService_EncryptToken(t *testing.T)
func TestTokenService_DecryptToken(t *testing.T)
func TestTokenService_RefreshGoogleToken(t *testing.T)
```

### 5.2 Integration Tests

```go
// test/integration/auth_flow_test.go
func TestAuthFlow_CompleteOAuthFlow(t *testing.T) {
    // 1. Get OAuth URL
    // 2. Simulate Google callback
    // 3. Verify JWT tokens returned
    // 4. Verify user created in DB
    // 5. Verify OAuth tokens stored
}

func TestAuthFlow_RefreshToken(t *testing.T) {
    // 1. Create user with refresh token
    // 2. Refresh access token
    // 3. Verify new access token valid
}

func TestAuthFlow_Logout(t *testing.T) {
    // 1. Login user
    // 2. Logout
    // 3. Verify token blacklisted
    // 4. Verify subsequent requests fail
}
```

### 5.3 Test Coverage Goal
- **Overall**: 80% 이상
- **Service Layer**: 90% 이상 (핵심 비즈니스 로직)
- **Handler Layer**: 70% 이상
- **Repository Layer**: 85% 이상

## 6. Security Considerations

### 6.1 Token Security
- **OAuth 토큰 암호화**: AES-256-GCM 사용
- **Encryption Key**: 32바이트 랜덤 키 사용, 환경 변수로 관리
- **Key Rotation**: 정기적인 암호화 키 교체 (분기별)

### 6.2 CSRF Protection
- **State 파라미터**: OAuth 플로우에 랜덤 state 사용
- **State 검증**: 콜백 시 state 일치 확인
- **State 저장**: Redis에 state 저장 (TTL 10분)

### 6.3 JWT Security
- **Short-lived Access Token**: 15분 수명
- **Secure Refresh Token**: 7일 수명, HttpOnly 쿠키로 전송 권장
- **Token Signing**: HMAC-SHA256 알고리즘 사용
- **Blacklist**: Redis를 사용한 로그아웃 토큰 블랙리스트

### 6.4 Rate Limiting
- **인증 엔드포인트**: IP당 분당 10회 제한
- **일반 API**: IP당 분당 100회 제한
- **Redis 기반**: 분산 환경에서도 동작

### 6.5 Input Validation
- **이메일 형식**: RFC 5322 준수
- **토큰 형식**: JWT 표준 검증
- **SQL Injection 방지**: GORM 파라미터 바인딩 사용

## 7. Monitoring & Logging

### 7.1 Key Metrics
- **인증 성공률**: Google OAuth 인증 성공 비율
- **토큰 갱신 성공률**: 액세스 토큰 자동 갱신 성공 비율
- **API 응답 시간**: P50, P95, P99 레이턴시
- **에러율**: 5xx 에러 발생 비율

### 7.2 Logging
- **구조화된 로깅**: JSON 형식 로그
- **로그 레벨**: DEBUG, INFO, WARN, ERROR
- **민감 정보 마스킹**: 액세스 토큰, 리프레시 토큰, 이메일 일부 마스킹
- **Request ID**: 모든 요청에 고유 ID 부여하여 추적

### 7.3 Alerting
- **인증 실패율 > 10%**: 즉시 알림
- **API 응답 시간 P95 > 1초**: 경고
- **데이터베이스 연결 실패**: 즉시 알림

## 8. Dependencies

### 8.1 Go Packages
```go
require (
    github.com/gin-gonic/gin v1.9.1
    github.com/golang-jwt/jwt/v5 v5.2.0
    github.com/go-redis/redis/v8 v8.11.5
    golang.org/x/oauth2 v0.15.0
    google.golang.org/api v0.156.0
    gorm.io/gorm v1.25.5
    gorm.io/driver/postgres v1.5.4
    github.com/google/uuid v1.5.0
    github.com/joho/godotenv v1.5.1
    github.com/stretchr/testify v1.8.4
    github.com/golang/mock v1.6.0
)
```

### 8.2 External Services
- **Google OAuth 2.0**: 사용자 인증
- **YouTube Data API v3**: 채널 정보 조회
- **PostgreSQL**: 사용자 및 토큰 데이터 저장
- **Redis**: 토큰 캐싱 및 블랙리스트

## 9. Risks & Mitigation

### 9.1 Risk: Google API 할당량 초과
**Impact**: High
**Probability**: Medium
**Mitigation**:
- YouTube Data API 일일 할당량 모니터링
- 백오프 전략 구현 (Exponential Backoff)
- 사용자별 API 호출 제한

### 9.2 Risk: OAuth 토큰 만료
**Impact**: High
**Probability**: Medium
**Mitigation**:
- 자동 토큰 갱신 로직 구현
- 리프레시 토큰 만료 7일 전 사용자에게 알림
- 토큰 갱신 실패 시 재인증 요청

### 9.3 Risk: 데이터베이스 연결 실패
**Impact**: Critical
**Probability**: Low
**Mitigation**:
- 자동 재연결 로직 (3회 재시도)
- 데이터베이스 커넥션 풀 사용
- Health check 엔드포인트로 DB 상태 모니터링

### 9.4 Risk: JWT Secret 노출
**Impact**: Critical
**Probability**: Low
**Mitigation**:
- 환경 변수로 Secret 관리 (절대 코드에 하드코딩 금지)
- Secret Rotation 정책 (분기별)
- Secret Manager 사용 고려 (AWS Secrets Manager, HashiCorp Vault)

## 10. Success Criteria

### 10.1 Functional Criteria
- [ ] 사용자가 Google 계정으로 로그인 가능
- [ ] 로그인 후 JWT 토큰 발급
- [ ] JWT 토큰으로 보호된 API 접근 가능
- [ ] 액세스 토큰 만료 시 리프레시 토큰으로 자동 갱신
- [ ] 로그아웃 시 토큰 무효화
- [ ] 사용자 정보 조회 가능

### 10.2 Non-Functional Criteria
- [ ] 인증 API 응답 시간 P95 < 500ms
- [ ] 테스트 커버리지 80% 이상
- [ ] 코드 린트 에러 0개 (`golangci-lint run`)
- [ ] 보안 취약점 0개
- [ ] API 문서 완성도 100%

### 10.3 Quality Criteria
- [ ] Clean Architecture 패턴 준수
- [ ] SOLID 원칙 준수
- [ ] 모든 함수에 단위 테스트 작성
- [ ] 통합 테스트로 E2E 플로우 검증
- [ ] 에러 핸들링 표준화

## 11. Timeline

### Sprint 1 (Week 1)
- Day 1-2: 프로젝트 초기화, 데이터베이스 스키마, Domain Layer
- Day 3-5: Repository Layer, Service Layer
- Day 6-7: Handler Layer, Middleware, Router

### Sprint 2 (Week 2)
- Day 1-3: 테스트 작성 (Unit + Integration)
- Day 4-5: 문서화, 린트 및 코드 리뷰
- Day 6-7: 버그 수정, 최적화, 배포 준비

**Total Estimated Time**: 2 weeks (10 working days)

## 12. Next Steps

1. **Plan Review**: 이 Plan 문서 검토 및 승인
2. **Design Phase**: `/pdca design youtube-auth-api` 실행하여 상세 설계 문서 작성
3. **Implementation**: 설계 문서 기반으로 구현 시작
4. **Testing**: 구현 완료 후 `/pdca analyze youtube-auth-api` 실행하여 갭 분석
5. **Iteration**: 필요 시 `/pdca iterate youtube-auth-api` 실행하여 개선

---

**Document Version**: 1.0
**Created**: 2026-03-24
**Author**: Claude Code PDCA System
**Status**: Draft
