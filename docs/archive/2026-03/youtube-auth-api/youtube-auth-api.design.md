# Design: YouTube 계정 로그인 API

## Document Control

| Property | Value |
|----------|-------|
| **Feature** | youtube-auth-api |
| **Phase** | Design |
| **Version** | 1.0 |
| **Created** | 2026-03-24 |
| **Status** | Draft |
| **Plan Reference** | [youtube-auth-api.plan.md](../../01-plan/features/youtube-auth-api.plan.md) |

---

## Table of Contents

1. [System Architecture](#1-system-architecture)
2. [Data Models](#2-data-models)
3. [API Specification](#3-api-specification)
4. [Service Layer Design](#4-service-layer-design)
5. [Security Implementation](#5-security-implementation)
6. [Error Handling](#6-error-handling)
7. [Implementation Order](#7-implementation-order)
8. [Testing Specifications](#8-testing-specifications)

---

## 1. System Architecture

### 1.1 Directory Structure

```
backend/
├── cmd/
│   └── api/
│       └── main.go                    # Application entry point
├── internal/
│   ├── config/
│   │   └── config.go                  # Configuration management
│   ├── domain/
│   │   ├── user.go                    # User entity
│   │   ├── token.go                   # Token entity
│   │   └── errors.go                  # Domain errors
│   ├── repository/
│   │   ├── interfaces.go              # Repository interfaces
│   │   ├── user_repository.go         # User data access
│   │   └── token_repository.go        # Token data access
│   ├── service/
│   │   ├── auth_service.go            # Authentication business logic
│   │   ├── token_service.go           # Token management
│   │   └── youtube_service.go         # YouTube API integration
│   ├── handler/
│   │   ├── auth_handler.go            # HTTP handlers
│   │   └── response.go                # Response helpers
│   ├── middleware/
│   │   ├── auth.go                    # JWT authentication
│   │   ├── rate_limiter.go            # Rate limiting
│   │   └── cors.go                    # CORS configuration
│   ├── router/
│   │   └── router.go                  # Route definitions
│   └── pkg/
│       ├── database/
│       │   └── postgres.go            # PostgreSQL connection
│       ├── redis/
│       │   └── redis.go               # Redis connection
│       └── logger/
│           └── logger.go              # Structured logging
├── migrations/
│   ├── 000001_create_users_table.up.sql
│   ├── 000001_create_users_table.down.sql
│   ├── 000002_create_user_tokens_table.up.sql
│   └── 000002_create_user_tokens_table.down.sql
├── test/
│   ├── integration/
│   │   └── auth_flow_test.go
│   └── mocks/
│       └── mock_*.go                  # Generated mocks
├── .env.example
├── go.mod
├── go.sum
├── Makefile
└── README.md
```

### 1.2 Component Diagram

```
┌─────────────────────────────────────────────────────────────┐
│                     External Systems                         │
│  ┌──────────────┐         ┌──────────────┐                  │
│  │ Google OAuth │         │ YouTube API  │                  │
│  └──────────────┘         └──────────────┘                  │
└──────────────┬────────────────────┬─────────────────────────┘
               │                    │
┌──────────────┴────────────────────┴─────────────────────────┐
│                    API Gateway (Gin)                         │
│  ┌────────────────────────────────────────────────────────┐ │
│  │              Middleware Layer                          │ │
│  │  - CORS                                                │ │
│  │  - Rate Limiter                                        │ │
│  │  - JWT Auth                                            │ │
│  │  - Request Logger                                      │ │
│  └────────────────────────────────────────────────────────┘ │
└──────────────┬──────────────────────────────────────────────┘
               ↓
┌──────────────────────────────────────────────────────────────┐
│                      Handler Layer                           │
│  ┌────────────────────────────────────────────────────────┐ │
│  │  AuthHandler                                           │ │
│  │  - GetGoogleAuthURL()                                  │ │
│  │  - HandleGoogleCallback()                              │ │
│  │  - RefreshToken()                                      │ │
│  │  - GetCurrentUser()                                    │ │
│  │  - Logout()                                            │ │
│  └────────────────────────────────────────────────────────┘ │
└──────────────┬──────────────────────────────────────────────┘
               ↓
┌──────────────────────────────────────────────────────────────┐
│                      Service Layer                           │
│  ┌───────────────┐  ┌───────────────┐  ┌───────────────┐   │
│  │  AuthService  │  │ TokenService  │  │YouTubeService │   │
│  │               │  │               │  │               │   │
│  │ - OAuth Flow  │  │ - Encrypt     │  │ - GetChannel  │   │
│  │ - JWT Mgmt    │  │ - Decrypt     │  │ - GetProfile  │   │
│  │ - User Mgmt   │  │ - Blacklist   │  │               │   │
│  └───────────────┘  └───────────────┘  └───────────────┘   │
└──────────────┬──────────────────────────────────────────────┘
               ↓
┌──────────────────────────────────────────────────────────────┐
│                    Repository Layer                          │
│  ┌──────────────────┐         ┌──────────────────┐          │
│  │  UserRepository  │         │ TokenRepository  │          │
│  │                  │         │                  │          │
│  │  - Create        │         │  - Create        │          │
│  │  - FindByID      │         │  - FindByUserID  │          │
│  │  - FindByEmail   │         │  - Update        │          │
│  │  - Update        │         │  - Delete        │          │
│  └──────────────────┘         └──────────────────┘          │
└──────────────┬────────────────────────┬─────────────────────┘
               ↓                        ↓
    ┌──────────────────┐    ┌──────────────────┐
    │   PostgreSQL     │    │      Redis       │
    │                  │    │                  │
    │  - users         │    │  - jwt_blacklist │
    │  - user_tokens   │    │  - oauth_states  │
    └──────────────────┘    └──────────────────┘
```

### 1.3 Request Flow Sequence

```
Client → Middleware → Handler → Service → Repository → Database
   ↑                                           ↓
   └──────────────← Response ←────────────────┘
```

---

## 2. Data Models

### 2.1 Domain Entities

#### 2.1.1 User Entity (`internal/domain/user.go`)

```go
package domain

import (
    "time"
    "github.com/google/uuid"
)

// User represents a user in the system
type User struct {
    ID                 uuid.UUID  `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
    Email              string     `json:"email" gorm:"type:varchar(255);uniqueIndex;not null"`
    GoogleID           string     `json:"google_id" gorm:"type:varchar(255);uniqueIndex;not null"`
    YouTubeChannelID   *string    `json:"youtube_channel_id" gorm:"type:varchar(255)"`
    YouTubeChannelName *string    `json:"youtube_channel_name" gorm:"type:varchar(255)"`
    ProfileImageURL    *string    `json:"profile_image_url" gorm:"type:text"`
    CreatedAt          time.Time  `json:"created_at" gorm:"not null;default:now()"`
    UpdatedAt          time.Time  `json:"updated_at" gorm:"not null;default:now()"`
    DeletedAt          *time.Time `json:"deleted_at,omitempty" gorm:"index"`
}

// TableName specifies the table name for GORM
func (User) TableName() string {
    return "users"
}

// BeforeCreate hook
func (u *User) BeforeCreate() error {
    if u.ID == uuid.Nil {
        u.ID = uuid.New()
    }
    now := time.Now()
    u.CreatedAt = now
    u.UpdatedAt = now
    return nil
}

// BeforeUpdate hook
func (u *User) BeforeUpdate() error {
    u.UpdatedAt = time.Now()
    return nil
}
```

#### 2.1.2 Token Entity (`internal/domain/token.go`)

```go
package domain

import (
    "time"
    "github.com/google/uuid"
)

// Token represents OAuth tokens for a user
type Token struct {
    ID                    uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
    UserID                uuid.UUID `json:"user_id" gorm:"type:uuid;not null;uniqueIndex"`
    EncryptedAccessToken  string    `json:"-" gorm:"type:text;not null"`
    EncryptedRefreshToken string    `json:"-" gorm:"type:text;not null"`
    TokenType             string    `json:"token_type" gorm:"type:varchar(50);not null;default:'Bearer'"`
    ExpiresAt             time.Time `json:"expires_at" gorm:"not null;index"`
    CreatedAt             time.Time `json:"created_at" gorm:"not null;default:now()"`
    UpdatedAt             time.Time `json:"updated_at" gorm:"not null;default:now()"`

    // Associations
    User User `json:"user" gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
}

// TableName specifies the table name for GORM
func (Token) TableName() string {
    return "user_tokens"
}

// BeforeCreate hook
func (t *Token) BeforeCreate() error {
    if t.ID == uuid.Nil {
        t.ID = uuid.New()
    }
    now := time.Now()
    t.CreatedAt = now
    t.UpdatedAt = now
    return nil
}

// BeforeUpdate hook
func (t *Token) BeforeUpdate() error {
    t.UpdatedAt = time.Now()
    return nil
}

// IsExpired checks if the OAuth token is expired
func (t *Token) IsExpired() bool {
    return time.Now().After(t.ExpiresAt)
}
```

#### 2.1.3 Domain Errors (`internal/domain/errors.go`)

```go
package domain

import "errors"

var (
    // User errors
    ErrUserNotFound       = errors.New("user not found")
    ErrUserAlreadyExists  = errors.New("user already exists")
    ErrInvalidUserData    = errors.New("invalid user data")

    // Token errors
    ErrTokenNotFound      = errors.New("token not found")
    ErrTokenExpired       = errors.New("token expired")
    ErrTokenInvalid       = errors.New("token invalid")
    ErrTokenBlacklisted   = errors.New("token blacklisted")

    // Auth errors
    ErrUnauthorized       = errors.New("unauthorized")
    ErrInvalidCredentials = errors.New("invalid credentials")
    ErrInvalidState       = errors.New("invalid state parameter")
    ErrOAuthFailed        = errors.New("oauth authentication failed")

    // General errors
    ErrInternalServer     = errors.New("internal server error")
    ErrInvalidInput       = errors.New("invalid input")
)
```

### 2.2 Database Schema

#### 2.2.1 Migration: Create Users Table

**File**: `migrations/000001_create_users_table.up.sql`

```sql
-- Create users table
CREATE TABLE IF NOT EXISTS users (
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

-- Create indexes
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_users_google_id ON users(google_id);
CREATE INDEX IF NOT EXISTS idx_users_deleted_at ON users(deleted_at);

-- Add comments
COMMENT ON TABLE users IS 'Stores user account information';
COMMENT ON COLUMN users.id IS 'Primary key UUID';
COMMENT ON COLUMN users.email IS 'User email from Google OAuth';
COMMENT ON COLUMN users.google_id IS 'Google account ID';
COMMENT ON COLUMN users.youtube_channel_id IS 'YouTube channel ID';
COMMENT ON COLUMN users.deleted_at IS 'Soft delete timestamp';
```

**File**: `migrations/000001_create_users_table.down.sql`

```sql
-- Drop users table
DROP INDEX IF EXISTS idx_users_deleted_at;
DROP INDEX IF EXISTS idx_users_google_id;
DROP INDEX IF EXISTS idx_users_email;
DROP TABLE IF EXISTS users;
```

#### 2.2.2 Migration: Create User Tokens Table

**File**: `migrations/000002_create_user_tokens_table.up.sql`

```sql
-- Create user_tokens table
CREATE TABLE IF NOT EXISTS user_tokens (
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

-- Create indexes
CREATE INDEX IF NOT EXISTS idx_user_tokens_user_id ON user_tokens(user_id);
CREATE INDEX IF NOT EXISTS idx_user_tokens_expires_at ON user_tokens(expires_at);

-- Add comments
COMMENT ON TABLE user_tokens IS 'Stores encrypted OAuth tokens';
COMMENT ON COLUMN user_tokens.encrypted_access_token IS 'AES-256 encrypted Google access token';
COMMENT ON COLUMN user_tokens.encrypted_refresh_token IS 'AES-256 encrypted Google refresh token';
COMMENT ON COLUMN user_tokens.expires_at IS 'Access token expiration timestamp';
```

**File**: `migrations/000002_create_user_tokens_table.down.sql`

```sql
-- Drop user_tokens table
DROP INDEX IF EXISTS idx_user_tokens_expires_at;
DROP INDEX IF EXISTS idx_user_tokens_user_id;
DROP TABLE IF EXISTS user_tokens;
```

### 2.3 DTO (Data Transfer Objects)

#### 2.3.1 Request DTOs (`internal/handler/dto.go`)

```go
package handler

// GetAuthURLRequest represents the request to get OAuth URL
type GetAuthURLRequest struct {
    RedirectURL string `json:"redirect_url,omitempty"` // Optional custom redirect
}

// GoogleCallbackRequest represents OAuth callback parameters
type GoogleCallbackRequest struct {
    Code  string `json:"code" binding:"required"`
    State string `json:"state" binding:"required"`
}

// RefreshTokenRequest represents token refresh request
type RefreshTokenRequest struct {
    RefreshToken string `json:"refresh_token" binding:"required"`
}
```

#### 2.3.2 Response DTOs (`internal/handler/response.go`)

```go
package handler

import (
    "time"
    "github.com/google/uuid"
)

// AuthURLResponse represents OAuth URL response
type AuthURLResponse struct {
    AuthURL string `json:"auth_url"`
    State   string `json:"state"`
}

// AuthResponse represents authentication response with JWT tokens
type AuthResponse struct {
    AccessToken  string       `json:"access_token"`
    RefreshToken string       `json:"refresh_token"`
    ExpiresIn    int          `json:"expires_in"`
    TokenType    string       `json:"token_type"`
    User         UserResponse `json:"user"`
}

// UserResponse represents user information response
type UserResponse struct {
    ID                 uuid.UUID `json:"id"`
    Email              string    `json:"email"`
    GoogleID           string    `json:"google_id"`
    YouTubeChannelID   *string   `json:"youtube_channel_id,omitempty"`
    YouTubeChannelName *string   `json:"youtube_channel_name,omitempty"`
    ProfileImageURL    *string   `json:"profile_image_url,omitempty"`
    CreatedAt          time.Time `json:"created_at"`
}

// ErrorResponse represents error response
type ErrorResponse struct {
    Error   string      `json:"error"`
    Message string      `json:"message"`
    Details interface{} `json:"details,omitempty"`
}

// SuccessResponse represents generic success response
type SuccessResponse struct {
    Success bool        `json:"success"`
    Message string      `json:"message"`
    Data    interface{} `json:"data,omitempty"`
}
```

---

## 3. API Specification

### 3.1 API Endpoints Overview

| Method | Endpoint | Description | Auth | Rate Limit |
|--------|----------|-------------|------|------------|
| GET | `/api/v1/auth/google/url` | Get OAuth URL | No | 10/min |
| POST | `/api/v1/auth/google/callback` | OAuth callback | No | 10/min |
| POST | `/api/v1/auth/refresh` | Refresh JWT | No | 20/min |
| GET | `/api/v1/auth/me` | Get current user | Yes | 60/min |
| POST | `/api/v1/auth/logout` | Logout | Yes | 20/min |
| GET | `/health` | Health check | No | 100/min |

### 3.2 Detailed API Specifications

#### 3.2.1 GET /api/v1/auth/google/url

**Description**: Generate Google OAuth authentication URL

**Request**:
```http
GET /api/v1/auth/google/url HTTP/1.1
Host: api.example.com
```

**Query Parameters**: None

**Response (200 OK)**:
```json
{
  "auth_url": "https://accounts.google.com/o/oauth2/v2/auth?client_id=...&redirect_uri=...&response_type=code&scope=...&state=...",
  "state": "random_state_string_12345"
}
```

**Error Responses**:
- `500 Internal Server Error`: Failed to generate OAuth URL
```json
{
  "error": "internal_server_error",
  "message": "Failed to generate authentication URL"
}
```

**Rate Limit**: 10 requests per minute per IP

---

#### 3.2.2 POST /api/v1/auth/google/callback

**Description**: Handle Google OAuth callback and issue JWT tokens

**Request**:
```http
POST /api/v1/auth/google/callback HTTP/1.1
Host: api.example.com
Content-Type: application/json

{
  "code": "4/0AX4XfWh...",
  "state": "random_state_string_12345"
}
```

**Request Body**:
```json
{
  "code": "string (required)",
  "state": "string (required)"
}
```

**Response (200 OK)**:
```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "expires_in": 900,
  "token_type": "Bearer",
  "user": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "email": "user@example.com",
    "google_id": "1234567890",
    "youtube_channel_id": "UCxxxxxxxx",
    "youtube_channel_name": "My Channel",
    "profile_image_url": "https://...",
    "created_at": "2026-03-24T10:00:00Z"
  }
}
```

**Error Responses**:

- `400 Bad Request`: Invalid request body
```json
{
  "error": "bad_request",
  "message": "Invalid request parameters",
  "details": {
    "code": "required field",
    "state": "required field"
  }
}
```

- `401 Unauthorized`: Invalid state or OAuth code
```json
{
  "error": "unauthorized",
  "message": "Invalid state parameter or authorization code"
}
```

- `500 Internal Server Error`: OAuth exchange failed
```json
{
  "error": "internal_server_error",
  "message": "Failed to exchange authorization code for tokens"
}
```

**Rate Limit**: 10 requests per minute per IP

---

#### 3.2.3 POST /api/v1/auth/refresh

**Description**: Refresh JWT access token using refresh token

**Request**:
```http
POST /api/v1/auth/refresh HTTP/1.1
Host: api.example.com
Content-Type: application/json

{
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

**Request Body**:
```json
{
  "refresh_token": "string (required)"
}
```

**Response (200 OK)**:
```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "expires_in": 900,
  "token_type": "Bearer"
}
```

**Error Responses**:

- `400 Bad Request`: Missing refresh token
```json
{
  "error": "bad_request",
  "message": "Refresh token is required"
}
```

- `401 Unauthorized`: Invalid or expired refresh token
```json
{
  "error": "unauthorized",
  "message": "Invalid or expired refresh token"
}
```

**Rate Limit**: 20 requests per minute per IP

---

#### 3.2.4 GET /api/v1/auth/me

**Description**: Get current authenticated user information

**Request**:
```http
GET /api/v1/auth/me HTTP/1.1
Host: api.example.com
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

**Headers**:
- `Authorization: Bearer <access_token>` (required)

**Response (200 OK)**:
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "email": "user@example.com",
  "google_id": "1234567890",
  "youtube_channel_id": "UCxxxxxxxx",
  "youtube_channel_name": "My Channel",
  "profile_image_url": "https://...",
  "created_at": "2026-03-24T10:00:00Z"
}
```

**Error Responses**:

- `401 Unauthorized`: Missing or invalid JWT token
```json
{
  "error": "unauthorized",
  "message": "Missing or invalid authentication token"
}
```

- `404 Not Found`: User not found
```json
{
  "error": "not_found",
  "message": "User not found"
}
```

**Rate Limit**: 60 requests per minute per user

---

#### 3.2.5 POST /api/v1/auth/logout

**Description**: Logout user and blacklist JWT token

**Request**:
```http
POST /api/v1/auth/logout HTTP/1.1
Host: api.example.com
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

**Headers**:
- `Authorization: Bearer <access_token>` (required)

**Response (200 OK)**:
```json
{
  "success": true,
  "message": "Logged out successfully"
}
```

**Error Responses**:

- `401 Unauthorized`: Missing or invalid JWT token
```json
{
  "error": "unauthorized",
  "message": "Missing or invalid authentication token"
}
```

**Rate Limit**: 20 requests per minute per user

---

#### 3.2.6 GET /health

**Description**: Health check endpoint

**Request**:
```http
GET /health HTTP/1.1
Host: api.example.com
```

**Response (200 OK)**:
```json
{
  "status": "healthy",
  "timestamp": "2026-03-24T10:00:00Z",
  "services": {
    "database": "up",
    "redis": "up"
  }
}
```

**Rate Limit**: 100 requests per minute per IP

---

## 4. Service Layer Design

### 4.1 AuthService Interface (`internal/service/auth_service.go`)

```go
package service

import (
    "context"
    "github.com/yourusername/video-upload-backend/internal/domain"
)

// AuthService defines authentication business logic
type AuthService interface {
    // GenerateAuthURL generates Google OAuth authentication URL
    GenerateAuthURL(ctx context.Context) (authURL string, state string, err error)

    // HandleCallback handles OAuth callback and creates/updates user
    HandleCallback(ctx context.Context, code, state string) (*domain.User, string, string, error)

    // GenerateJWT generates JWT access and refresh tokens
    GenerateJWT(ctx context.Context, userID string) (accessToken, refreshToken string, err error)

    // ValidateJWT validates JWT token and returns claims
    ValidateJWT(ctx context.Context, token string) (*JWTClaims, error)

    // RefreshAccessToken refreshes access token using refresh token
    RefreshAccessToken(ctx context.Context, refreshToken string) (accessToken string, err error)

    // GetUserByID retrieves user by ID
    GetUserByID(ctx context.Context, userID string) (*domain.User, error)

    // Logout blacklists JWT token
    Logout(ctx context.Context, token string, userID string) error
}

// JWTClaims represents JWT token claims
type JWTClaims struct {
    UserID    string `json:"user_id"`
    Email     string `json:"email"`
    TokenType string `json:"token_type"` // "access" or "refresh"
    IssuedAt  int64  `json:"iat"`
    ExpiresAt int64  `json:"exp"`
}
```

### 4.2 TokenService Interface (`internal/service/token_service.go`)

```go
package service

import (
    "context"
)

// TokenService defines token management operations
type TokenService interface {
    // EncryptToken encrypts token using AES-256-GCM
    EncryptToken(ctx context.Context, plainText string) (string, error)

    // DecryptToken decrypts token using AES-256-GCM
    DecryptToken(ctx context.Context, cipherText string) (string, error)

    // RefreshGoogleToken refreshes Google OAuth token
    RefreshGoogleToken(ctx context.Context, userID string) error

    // AddToBlacklist adds token to blacklist in Redis
    AddToBlacklist(ctx context.Context, token string, expirySeconds int64) error

    // IsBlacklisted checks if token is blacklisted
    IsBlacklisted(ctx context.Context, token string) (bool, error)

    // SaveOAuthState saves OAuth state to Redis with TTL
    SaveOAuthState(ctx context.Context, state string, ttlSeconds int64) error

    // ValidateOAuthState validates and removes OAuth state from Redis
    ValidateOAuthState(ctx context.Context, state string) (bool, error)
}
```

### 4.3 YouTubeService Interface (`internal/service/youtube_service.go`)

```go
package service

import (
    "context"
)

// YouTubeService defines YouTube API operations
type YouTubeService interface {
    // GetChannelInfo retrieves YouTube channel information
    GetChannelInfo(ctx context.Context, accessToken string) (*ChannelInfo, error)

    // GetUserProfile retrieves user profile from Google
    GetUserProfile(ctx context.Context, accessToken string) (*UserProfile, error)
}

// ChannelInfo represents YouTube channel information
type ChannelInfo struct {
    ChannelID   string `json:"channel_id"`
    ChannelName string `json:"channel_name"`
    Thumbnail   string `json:"thumbnail"`
}

// UserProfile represents Google user profile
type UserProfile struct {
    GoogleID    string `json:"google_id"`
    Email       string `json:"email"`
    Name        string `json:"name"`
    Picture     string `json:"picture"`
}
```

### 4.4 AuthService Implementation Example

```go
package service

import (
    "context"
    "crypto/rand"
    "encoding/base64"
    "fmt"
    "time"

    "github.com/golang-jwt/jwt/v5"
    "golang.org/x/oauth2"
    "golang.org/x/oauth2/google"
    "google.golang.org/api/youtube/v3"

    "github.com/yourusername/video-upload-backend/internal/domain"
    "github.com/yourusername/video-upload-backend/internal/repository"
)

type authService struct {
    userRepo      repository.UserRepository
    tokenRepo     repository.TokenRepository
    tokenService  TokenService
    youtubeService YouTubeService
    oauthConfig   *oauth2.Config
    jwtSecret     string
    accessTokenExpiry  time.Duration
    refreshTokenExpiry time.Duration
}

// NewAuthService creates new AuthService instance
func NewAuthService(
    userRepo repository.UserRepository,
    tokenRepo repository.TokenRepository,
    tokenService TokenService,
    youtubeService YouTubeService,
    clientID, clientSecret, redirectURL, jwtSecret string,
    accessExpiry, refreshExpiry time.Duration,
) AuthService {
    return &authService{
        userRepo:      userRepo,
        tokenRepo:     tokenRepo,
        tokenService:  tokenService,
        youtubeService: youtubeService,
        oauthConfig: &oauth2.Config{
            ClientID:     clientID,
            ClientSecret: clientSecret,
            RedirectURL:  redirectURL,
            Scopes: []string{
                "https://www.googleapis.com/auth/userinfo.email",
                "https://www.googleapis.com/auth/userinfo.profile",
                "https://www.googleapis.com/auth/youtube.upload",
                "https://www.googleapis.com/auth/youtube",
            },
            Endpoint: google.Endpoint,
        },
        jwtSecret:          jwtSecret,
        accessTokenExpiry:  accessExpiry,
        refreshTokenExpiry: refreshExpiry,
    }
}

// GenerateAuthURL generates OAuth URL with random state
func (s *authService) GenerateAuthURL(ctx context.Context) (string, string, error) {
    state, err := generateRandomState(32)
    if err != nil {
        return "", "", fmt.Errorf("failed to generate state: %w", err)
    }

    // Save state to Redis with 10-minute TTL
    if err := s.tokenService.SaveOAuthState(ctx, state, 600); err != nil {
        return "", "", fmt.Errorf("failed to save state: %w", err)
    }

    authURL := s.oauthConfig.AuthCodeURL(state, oauth2.AccessTypeOffline, oauth2.ApprovalForce)
    return authURL, state, nil
}

// generateRandomState generates cryptographically secure random state
func generateRandomState(length int) (string, error) {
    b := make([]byte, length)
    if _, err := rand.Read(b); err != nil {
        return "", err
    }
    return base64.URLEncoding.EncodeToString(b), nil
}

// HandleCallback handles OAuth callback
func (s *authService) HandleCallback(ctx context.Context, code, state string) (*domain.User, string, string, error) {
    // Validate state parameter
    valid, err := s.tokenService.ValidateOAuthState(ctx, state)
    if err != nil || !valid {
        return nil, "", "", domain.ErrInvalidState
    }

    // Exchange code for tokens
    token, err := s.oauthConfig.Exchange(ctx, code)
    if err != nil {
        return nil, "", "", fmt.Errorf("failed to exchange code: %w", err)
    }

    // Get user profile
    profile, err := s.youtubeService.GetUserProfile(ctx, token.AccessToken)
    if err != nil {
        return nil, "", "", fmt.Errorf("failed to get user profile: %w", err)
    }

    // Get YouTube channel info
    channelInfo, _ := s.youtubeService.GetChannelInfo(ctx, token.AccessToken)

    // Find or create user
    user, err := s.userRepo.FindByGoogleID(ctx, profile.GoogleID)
    if err == domain.ErrUserNotFound {
        // Create new user
        user = &domain.User{
            Email:    profile.Email,
            GoogleID: profile.GoogleID,
            ProfileImageURL: &profile.Picture,
        }

        if channelInfo != nil {
            user.YouTubeChannelID = &channelInfo.ChannelID
            user.YouTubeChannelName = &channelInfo.ChannelName
        }

        if err := s.userRepo.Create(ctx, user); err != nil {
            return nil, "", "", fmt.Errorf("failed to create user: %w", err)
        }
    } else if err != nil {
        return nil, "", "", err
    } else {
        // Update existing user
        if channelInfo != nil {
            user.YouTubeChannelID = &channelInfo.ChannelID
            user.YouTubeChannelName = &channelInfo.ChannelName
        }
        if err := s.userRepo.Update(ctx, user); err != nil {
            return nil, "", "", fmt.Errorf("failed to update user: %w", err)
        }
    }

    // Encrypt and save OAuth tokens
    encryptedAccess, err := s.tokenService.EncryptToken(ctx, token.AccessToken)
    if err != nil {
        return nil, "", "", fmt.Errorf("failed to encrypt access token: %w", err)
    }

    encryptedRefresh, err := s.tokenService.EncryptToken(ctx, token.RefreshToken)
    if err != nil {
        return nil, "", "", fmt.Errorf("failed to encrypt refresh token: %w", err)
    }

    userToken := &domain.Token{
        UserID:                user.ID,
        EncryptedAccessToken:  encryptedAccess,
        EncryptedRefreshToken: encryptedRefresh,
        TokenType:             "Bearer",
        ExpiresAt:             token.Expiry,
    }

    // Save or update token
    existingToken, err := s.tokenRepo.FindByUserID(ctx, user.ID.String())
    if err == domain.ErrTokenNotFound {
        if err := s.tokenRepo.Create(ctx, userToken); err != nil {
            return nil, "", "", fmt.Errorf("failed to save token: %w", err)
        }
    } else if err != nil {
        return nil, "", "", err
    } else {
        userToken.ID = existingToken.ID
        if err := s.tokenRepo.Update(ctx, userToken); err != nil {
            return nil, "", "", fmt.Errorf("failed to update token: %w", err)
        }
    }

    // Generate JWT tokens
    accessToken, refreshToken, err := s.GenerateJWT(ctx, user.ID.String())
    if err != nil {
        return nil, "", "", fmt.Errorf("failed to generate JWT: %w", err)
    }

    return user, accessToken, refreshToken, nil
}

// GenerateJWT generates JWT access and refresh tokens
func (s *authService) GenerateJWT(ctx context.Context, userID string) (string, string, error) {
    now := time.Now()

    // Generate access token
    accessClaims := jwt.MapClaims{
        "user_id":    userID,
        "token_type": "access",
        "iat":        now.Unix(),
        "exp":        now.Add(s.accessTokenExpiry).Unix(),
    }
    accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
    accessTokenString, err := accessToken.SignedString([]byte(s.jwtSecret))
    if err != nil {
        return "", "", fmt.Errorf("failed to sign access token: %w", err)
    }

    // Generate refresh token
    refreshClaims := jwt.MapClaims{
        "user_id":    userID,
        "token_type": "refresh",
        "iat":        now.Unix(),
        "exp":        now.Add(s.refreshTokenExpiry).Unix(),
    }
    refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
    refreshTokenString, err := refreshToken.SignedString([]byte(s.jwtSecret))
    if err != nil {
        return "", "", fmt.Errorf("failed to sign refresh token: %w", err)
    }

    return accessTokenString, refreshTokenString, nil
}

// ValidateJWT validates JWT token and returns claims
func (s *authService) ValidateJWT(ctx context.Context, tokenString string) (*JWTClaims, error) {
    // Check if token is blacklisted
    blacklisted, err := s.tokenService.IsBlacklisted(ctx, tokenString)
    if err != nil {
        return nil, err
    }
    if blacklisted {
        return nil, domain.ErrTokenBlacklisted
    }

    token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
        if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
            return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
        }
        return []byte(s.jwtSecret), nil
    })

    if err != nil {
        return nil, domain.ErrTokenInvalid
    }

    if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
        return &JWTClaims{
            UserID:    claims["user_id"].(string),
            TokenType: claims["token_type"].(string),
            IssuedAt:  int64(claims["iat"].(float64)),
            ExpiresAt: int64(claims["exp"].(float64)),
        }, nil
    }

    return nil, domain.ErrTokenInvalid
}

// RefreshAccessToken refreshes access token
func (s *authService) RefreshAccessToken(ctx context.Context, refreshToken string) (string, error) {
    claims, err := s.ValidateJWT(ctx, refreshToken)
    if err != nil {
        return "", err
    }

    if claims.TokenType != "refresh" {
        return "", domain.ErrTokenInvalid
    }

    // Generate new access token
    accessToken, _, err := s.GenerateJWT(ctx, claims.UserID)
    if err != nil {
        return "", err
    }

    return accessToken, nil
}

// GetUserByID retrieves user by ID
func (s *authService) GetUserByID(ctx context.Context, userID string) (*domain.User, error) {
    return s.userRepo.FindByID(ctx, userID)
}

// Logout blacklists JWT token
func (s *authService) Logout(ctx context.Context, token string, userID string) error {
    claims, err := s.ValidateJWT(ctx, token)
    if err != nil {
        return err
    }

    // Calculate TTL (time until token expires)
    ttl := claims.ExpiresAt - time.Now().Unix()
    if ttl <= 0 {
        return nil // Token already expired
    }

    return s.tokenService.AddToBlacklist(ctx, token, ttl)
}
```

---

## 5. Security Implementation

### 5.1 AES-256-GCM Encryption (`internal/service/token_service.go`)

```go
package service

import (
    "crypto/aes"
    "crypto/cipher"
    "crypto/rand"
    "encoding/base64"
    "errors"
    "io"
)

// EncryptToken encrypts plaintext using AES-256-GCM
func (s *tokenService) EncryptToken(ctx context.Context, plainText string) (string, error) {
    block, err := aes.NewCipher([]byte(s.encryptionKey))
    if err != nil {
        return "", err
    }

    aesGCM, err := cipher.NewGCM(block)
    if err != nil {
        return "", err
    }

    nonce := make([]byte, aesGCM.NonceSize())
    if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
        return "", err
    }

    ciphertext := aesGCM.Seal(nonce, nonce, []byte(plainText), nil)
    return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// DecryptToken decrypts ciphertext using AES-256-GCM
func (s *tokenService) DecryptToken(ctx context.Context, cipherText string) (string, error) {
    data, err := base64.StdEncoding.DecodeString(cipherText)
    if err != nil {
        return "", err
    }

    block, err := aes.NewCipher([]byte(s.encryptionKey))
    if err != nil {
        return "", err
    }

    aesGCM, err := cipher.NewGCM(block)
    if err != nil {
        return "", err
    }

    nonceSize := aesGCM.NonceSize()
    if len(data) < nonceSize {
        return "", errors.New("ciphertext too short")
    }

    nonce, ciphertext := data[:nonceSize], data[nonceSize:]
    plaintext, err := aesGCM.Open(nil, nonce, ciphertext, nil)
    if err != nil {
        return "", err
    }

    return string(plaintext), nil
}
```

### 5.2 Rate Limiter Middleware (`internal/middleware/rate_limiter.go`)

```go
package middleware

import (
    "fmt"
    "net/http"
    "time"

    "github.com/gin-gonic/gin"
    "github.com/go-redis/redis/v8"
)

// RateLimiter creates rate limiting middleware
func RateLimiter(redisClient *redis.Client, maxRequests int, window time.Duration) gin.HandlerFunc {
    return func(c *gin.Context) {
        ip := c.ClientIP()
        key := fmt.Sprintf("rate_limit:%s", ip)

        ctx := c.Request.Context()

        // Increment counter
        count, err := redisClient.Incr(ctx, key).Result()
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "rate_limit_error"})
            c.Abort()
            return
        }

        // Set expiry on first request
        if count == 1 {
            redisClient.Expire(ctx, key, window)
        }

        // Check limit
        if count > int64(maxRequests) {
            c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", maxRequests))
            c.Header("X-RateLimit-Remaining", "0")
            c.Header("X-RateLimit-Reset", fmt.Sprintf("%d", time.Now().Add(window).Unix()))

            c.JSON(http.StatusTooManyRequests, gin.H{
                "error":   "rate_limit_exceeded",
                "message": "Too many requests. Please try again later.",
            })
            c.Abort()
            return
        }

        // Set rate limit headers
        c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", maxRequests))
        c.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", maxRequests-int(count)))

        c.Next()
    }
}
```

### 5.3 JWT Authentication Middleware (`internal/middleware/auth.go`)

```go
package middleware

import (
    "net/http"
    "strings"

    "github.com/gin-gonic/gin"
    "github.com/yourusername/video-upload-backend/internal/service"
)

// AuthMiddleware validates JWT tokens
func AuthMiddleware(authService service.AuthService) gin.HandlerFunc {
    return func(c *gin.Context) {
        authHeader := c.GetHeader("Authorization")
        if authHeader == "" {
            c.JSON(http.StatusUnauthorized, gin.H{
                "error":   "unauthorized",
                "message": "Missing authorization header",
            })
            c.Abort()
            return
        }

        parts := strings.SplitN(authHeader, " ", 2)
        if len(parts) != 2 || parts[0] != "Bearer" {
            c.JSON(http.StatusUnauthorized, gin.H{
                "error":   "unauthorized",
                "message": "Invalid authorization header format",
            })
            c.Abort()
            return
        }

        token := parts[1]
        claims, err := authService.ValidateJWT(c.Request.Context(), token)
        if err != nil {
            c.JSON(http.StatusUnauthorized, gin.H{
                "error":   "unauthorized",
                "message": err.Error(),
            })
            c.Abort()
            return
        }

        // Set user ID in context
        c.Set("user_id", claims.UserID)
        c.Set("token", token)

        c.Next()
    }
}
```

---

## 6. Error Handling

### 6.1 Error Response Format

All error responses follow this structure:

```json
{
  "error": "error_code",
  "message": "Human-readable error message",
  "details": {
    // Optional: Additional error details
  }
}
```

### 6.2 HTTP Status Codes

| Status Code | Description | Use Case |
|-------------|-------------|----------|
| 200 OK | Success | Successful requests |
| 400 Bad Request | Invalid input | Validation errors, malformed JSON |
| 401 Unauthorized | Authentication failed | Invalid/expired token, missing auth |
| 404 Not Found | Resource not found | User not found |
| 429 Too Many Requests | Rate limit exceeded | Too many requests from IP |
| 500 Internal Server Error | Server error | Unexpected errors |

### 6.3 Error Handling Middleware

```go
package middleware

import (
    "net/http"

    "github.com/gin-gonic/gin"
    "github.com/yourusername/video-upload-backend/internal/pkg/logger"
)

// ErrorHandler handles panics and errors
func ErrorHandler() gin.HandlerFunc {
    return func(c *gin.Context) {
        defer func() {
            if err := recover(); err != nil {
                logger.Error("Panic recovered", "error", err)
                c.JSON(http.StatusInternalServerError, gin.H{
                    "error":   "internal_server_error",
                    "message": "An unexpected error occurred",
                })
                c.Abort()
            }
        }()

        c.Next()

        // Check for errors in context
        if len(c.Errors) > 0 {
            err := c.Errors.Last()
            logger.Error("Request error", "error", err.Err)

            // Return appropriate error response
            c.JSON(http.StatusInternalServerError, gin.H{
                "error":   "internal_server_error",
                "message": "An error occurred processing your request",
            })
        }
    }
}
```

---

## 7. Implementation Order

### 7.1 Day-by-Day Implementation Plan

#### **Day 1: Project Setup & Database**

1. Initialize Go module
```bash
mkdir -p backend
cd backend
go mod init github.com/yourusername/video-upload-backend
```

2. Create directory structure
```bash
mkdir -p cmd/api internal/{config,domain,repository,service,handler,middleware,router,pkg/{database,redis,logger}} migrations test/{integration,mocks}
```

3. Install dependencies
```bash
go get github.com/gin-gonic/gin
go get gorm.io/gorm
go get gorm.io/driver/postgres
go get github.com/golang-jwt/jwt/v5
go get golang.org/x/oauth2
go get google.golang.org/api/youtube/v3
go get github.com/go-redis/redis/v8
go get github.com/google/uuid
go get github.com/joho/godotenv
go get github.com/stretchr/testify
go get github.com/golang/mock/mockgen
```

4. Create environment variables file (`.env`)
5. Implement database migrations
6. Implement domain entities (`user.go`, `token.go`, `errors.go`)

**Deliverables**:
- ✅ Project structure created
- ✅ Dependencies installed
- ✅ Domain models defined
- ✅ Database migrations ready

---

#### **Day 2: Repository Layer**

1. Define repository interfaces (`internal/repository/interfaces.go`)
2. Implement UserRepository with GORM
3. Implement TokenRepository with GORM
4. Write unit tests for repositories

**Deliverables**:
- ✅ Repository interfaces defined
- ✅ Repository implementations complete
- ✅ Unit tests passing (80%+ coverage)

**Files to create**:
- `internal/repository/interfaces.go`
- `internal/repository/user_repository.go`
- `internal/repository/token_repository.go`
- `internal/repository/user_repository_test.go`
- `internal/repository/token_repository_test.go`

---

#### **Day 3: Token & YouTube Services**

1. Implement TokenService
   - AES-256-GCM encryption/decryption
   - Redis blacklist operations
   - OAuth state management
2. Implement YouTubeService
   - Channel info retrieval
   - User profile retrieval
3. Write unit tests

**Deliverables**:
- ✅ TokenService implementation complete
- ✅ YouTubeService implementation complete
- ✅ Unit tests passing (90%+ coverage)

**Files to create**:
- `internal/service/token_service.go`
- `internal/service/youtube_service.go`
- `internal/service/token_service_test.go`
- `internal/service/youtube_service_test.go`

---

#### **Day 4-5: AuthService**

1. Implement AuthService methods:
   - `GenerateAuthURL`
   - `HandleCallback`
   - `GenerateJWT`
   - `ValidateJWT`
   - `RefreshAccessToken`
   - `GetUserByID`
   - `Logout`
2. Write comprehensive unit tests
3. Test OAuth flow manually

**Deliverables**:
- ✅ AuthService implementation complete
- ✅ OAuth flow working
- ✅ JWT generation/validation working
- ✅ Unit tests passing (90%+ coverage)

**Files to create**:
- `internal/service/auth_service.go`
- `internal/service/auth_service_test.go`

---

#### **Day 6: Handler & Middleware**

1. Implement HTTP handlers
   - `GetGoogleAuthURL`
   - `HandleGoogleCallback`
   - `RefreshToken`
   - `GetCurrentUser`
   - `Logout`
2. Implement middleware
   - JWT authentication
   - Rate limiter
   - CORS
   - Error handler
3. Write handler tests

**Deliverables**:
- ✅ All handlers implemented
- ✅ Middleware complete
- ✅ Handler tests passing (70%+ coverage)

**Files to create**:
- `internal/handler/auth_handler.go`
- `internal/handler/response.go`
- `internal/handler/dto.go`
- `internal/handler/auth_handler_test.go`
- `internal/middleware/auth.go`
- `internal/middleware/rate_limiter.go`
- `internal/middleware/cors.go`
- `internal/middleware/error_handler.go`

---

#### **Day 7: Router & Integration**

1. Implement router with all routes
2. Implement main.go entry point
3. Implement health check endpoint
4. Test all endpoints with Postman/curl
5. Write integration tests

**Deliverables**:
- ✅ Router configured
- ✅ Server running
- ✅ All endpoints tested
- ✅ Integration tests passing

**Files to create**:
- `internal/router/router.go`
- `cmd/api/main.go`
- `test/integration/auth_flow_test.go`

---

#### **Day 8: Testing & Documentation**

1. Achieve 80%+ test coverage
2. Write API documentation (Swagger)
3. Update README
4. Write deployment guide
5. Code review and cleanup

**Deliverables**:
- ✅ Test coverage ≥ 80%
- ✅ API documentation complete
- ✅ README updated
- ✅ Code linted and formatted

---

## 8. Testing Specifications

### 8.1 Unit Test Coverage Goals

| Layer | Target Coverage | Priority |
|-------|-----------------|----------|
| Domain | 90% | High |
| Repository | 85% | High |
| Service | 90% | Critical |
| Handler | 70% | Medium |
| Middleware | 80% | High |

### 8.2 Integration Test Scenarios

#### Test 1: Complete OAuth Flow
```go
func TestAuthFlow_CompleteOAuthFlow(t *testing.T) {
    // 1. Get OAuth URL
    // 2. Simulate Google callback with valid code
    // 3. Verify JWT tokens returned
    // 4. Verify user created in database
    // 5. Verify OAuth tokens stored and encrypted
}
```

#### Test 2: Token Refresh
```go
func TestAuthFlow_RefreshToken(t *testing.T) {
    // 1. Create user with valid refresh token
    // 2. Call refresh endpoint
    // 3. Verify new access token generated
    // 4. Verify old access token still valid (not blacklisted)
}
```

#### Test 3: Logout
```go
func TestAuthFlow_Logout(t *testing.T) {
    // 1. Login user and get access token
    // 2. Call logout endpoint
    // 3. Verify token added to blacklist
    // 4. Verify subsequent requests with token fail
}
```

#### Test 4: Token Expiry
```go
func TestAuthFlow_TokenExpiry(t *testing.T) {
    // 1. Create user with expired access token
    // 2. Attempt to access protected endpoint
    // 3. Verify 401 Unauthorized response
    // 4. Refresh token and retry
    // 5. Verify success
}
```

### 8.3 Test Environment Setup

```go
// test/integration/setup.go
package integration

import (
    "context"
    "testing"

    "github.com/testcontainers/testcontainers-go"
    "github.com/testcontainers/testcontainers-go/wait"
)

func SetupTestDatabase(t *testing.T) *testcontainers.Container {
    ctx := context.Background()

    req := testcontainers.ContainerRequest{
        Image:        "postgres:15-alpine",
        ExposedPorts: []string{"5432/tcp"},
        Env: map[string]string{
            "POSTGRES_PASSWORD": "test",
            "POSTGRES_DB":       "test_db",
        },
        WaitingFor: wait.ForLog("database system is ready to accept connections"),
    }

    postgres, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
        ContainerRequest: req,
        Started:          true,
    })

    if err != nil {
        t.Fatal(err)
    }

    return &postgres
}

func SetupTestRedis(t *testing.T) *testcontainers.Container {
    ctx := context.Background()

    req := testcontainers.ContainerRequest{
        Image:        "redis:7-alpine",
        ExposedPorts: []string{"6379/tcp"},
        WaitingFor:   wait.ForLog("Ready to accept connections"),
    }

    redis, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
        ContainerRequest: req,
        Started:          true,
    })

    if err != nil {
        t.Fatal(err)
    }

    return &redis
}
```

---

## 9. Success Criteria

### 9.1 Functional Requirements

- ✅ OAuth URL generation works
- ✅ OAuth callback creates/updates users
- ✅ JWT tokens generated correctly
- ✅ Token refresh works
- ✅ Protected endpoints require authentication
- ✅ Logout blacklists tokens
- ✅ User information retrieved correctly

### 9.2 Non-Functional Requirements

- ✅ API response time P95 < 500ms
- ✅ Test coverage ≥ 80%
- ✅ Zero linting errors
- ✅ All migrations reversible
- ✅ Proper error handling
- ✅ Structured logging implemented

### 9.3 Security Requirements

- ✅ OAuth tokens encrypted with AES-256-GCM
- ✅ JWT tokens signed with HMAC-SHA256
- ✅ CSRF protection via state parameter
- ✅ Rate limiting active
- ✅ Blacklist prevents token reuse after logout
- ✅ No secrets in code (all in env vars)

---

## 10. Next Steps

After completing this design phase:

1. **Review Design**: Ensure all team members understand the architecture
2. **Start Implementation**: Follow the day-by-day implementation plan
3. **Daily Standups**: Track progress against timeline
4. **Code Reviews**: Review all code before merging
5. **Testing**: Run tests continuously during development
6. **Documentation**: Update docs as implementation progresses

To start implementation:
```bash
/pdca do youtube-auth-api
```

---

**Document Version**: 1.0
**Last Updated**: 2026-03-24
**Author**: Claude Code PDCA System
**Status**: Ready for Implementation
