# Video Upload Backend API

YouTube Authentication and Video Upload Backend Service built with Go, Gin, PostgreSQL, and Redis.

## Features

- **Google OAuth 2.0 Authentication**: Secure YouTube account login
- **JWT Token Management**: Access tokens (15min) and refresh tokens (7 days)
- **Token Security**: AES-256-GCM encryption for OAuth tokens
- **Rate Limiting**: Redis-based sliding window rate limiter
- **CORS Support**: Configurable cross-origin resource sharing
- **Clean Architecture**: Domain → Repository → Service → Handler layers
- **Health Checks**: Built-in health check endpoint
- **Graceful Shutdown**: Proper cleanup of resources

## Tech Stack

- **Language**: Go 1.21+
- **Web Framework**: Gin
- **Database**: PostgreSQL with GORM ORM
- **Cache**: Redis
- **Authentication**: Google OAuth 2.0, JWT
- **Security**: AES-256-GCM encryption, CSRF protection

## Project Structure

```
backend/
├── cmd/
│   └── api/
│       └── main.go                 # Application entry point
├── internal/
│   ├── config/
│   │   └── config.go              # Configuration management
│   ├── domain/
│   │   ├── user.go                # User entity
│   │   ├── token.go               # Token entity
│   │   └── errors.go              # Domain errors
│   ├── repository/
│   │   ├── interfaces.go          # Repository interfaces
│   │   ├── user_repository.go     # User data access
│   │   └── token_repository.go    # Token data access
│   ├── service/
│   │   ├── auth_service.go        # Authentication business logic
│   │   ├── token_service.go       # Token encryption & blacklist
│   │   └── youtube_service.go     # YouTube API integration
│   ├── handler/
│   │   ├── auth_handler.go        # HTTP handlers
│   │   ├── dto.go                 # Request/Response DTOs
│   │   └── response.go            # Response helpers
│   ├── middleware/
│   │   ├── auth.go                # JWT authentication
│   │   ├── rate_limiter.go        # Rate limiting
│   │   ├── cors.go                # CORS configuration
│   │   └── error_handler.go       # Error recovery & logging
│   ├── router/
│   │   └── router.go              # Route configuration
│   └── pkg/
│       ├── database/
│       │   └── postgres.go        # Database connection
│       ├── redis/
│       │   └── redis.go           # Redis client
│       └── logger/
│           └── logger.go          # Structured logging
├── migrations/
│   ├── 000001_create_users_table.up.sql
│   ├── 000001_create_users_table.down.sql
│   ├── 000002_create_user_tokens_table.up.sql
│   └── 000002_create_user_tokens_table.down.sql
├── .env.example                    # Environment variables template
├── go.mod                          # Go module dependencies
└── README.md                       # This file
```

## Setup

### Prerequisites

- Go 1.21 or higher
- PostgreSQL 14+
- Redis 6+
- Google OAuth 2.0 credentials

### 1. Clone and Setup

```bash
cd backend
cp .env.example .env
```

### 2. Configure Environment Variables

Edit `.env` file with your credentials:

```env
# Server Configuration
PORT=8080
ENV=development
LOG_LEVEL=info
ALLOWED_ORIGINS=http://localhost:3000

# Database Configuration
DATABASE_URL=postgresql://user:password@localhost:5432/video_upload?sslmode=disable
DB_MAX_IDLE_CONNS=10
DB_MAX_OPEN_CONNS=100
DB_CONN_MAX_LIFETIME=1h

# Redis Configuration
REDIS_URL=redis://localhost:6379/0
REDIS_PASSWORD=
REDIS_DB=0

# Google OAuth Configuration
GOOGLE_CLIENT_ID=your-client-id.apps.googleusercontent.com
GOOGLE_CLIENT_SECRET=your-client-secret
GOOGLE_REDIRECT_URL=http://localhost:8080/api/v1/auth/google/callback

# JWT Configuration
JWT_SECRET=your-32-character-secret-key-here
JWT_ACCESS_TOKEN_EXPIRY=15m
JWT_REFRESH_TOKEN_EXPIRY=168h

# Security Configuration
ENCRYPTION_KEY=your-exactly-32-character-key!
RATE_LIMIT_AUTH=10
RATE_LIMIT_GENERAL=60
```

### 3. Install Dependencies

```bash
go mod download
```

### 4. Run Database Migrations

```bash
# Using golang-migrate CLI
migrate -database "postgresql://user:password@localhost:5432/video_upload?sslmode=disable" \
        -path migrations up

# Or let the application auto-migrate on startup
```

### 5. Build and Run

```bash
# Build
go build -o bin/api ./cmd/api

# Run
./bin/api

# Or run directly
go run cmd/api/main.go
```

## API Endpoints

### Health Check

```
GET /health
```

### Authentication Endpoints

#### 1. Get Google OAuth URL

```
GET /api/v1/auth/google/url
```

**Request Body** (optional):
```json
{
  "redirect_url": "http://localhost:3000/auth/callback"
}
```

**Response**:
```json
{
  "success": true,
  "message": "OAuth URL generated successfully",
  "data": {
    "auth_url": "https://accounts.google.com/o/oauth2/v2/auth?...",
    "state": "random-state-string"
  }
}
```

#### 2. Handle OAuth Callback

```
POST /api/v1/auth/google/callback
```

**Request Body**:
```json
{
  "code": "oauth-authorization-code",
  "state": "state-from-step-1"
}
```

**Response**:
```json
{
  "success": true,
  "message": "Authentication successful",
  "data": {
    "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "expires_in": 900,
    "token_type": "Bearer",
    "user": {
      "id": "uuid",
      "email": "user@example.com",
      "google_id": "google-id",
      "youtube_channel_id": "channel-id",
      "youtube_channel_name": "Channel Name",
      "profile_image_url": "https://...",
      "created_at": "2024-01-01T00:00:00Z"
    }
  }
}
```

#### 3. Refresh Access Token

```
POST /api/v1/auth/refresh
```

**Request Body**:
```json
{
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

**Response**:
```json
{
  "success": true,
  "message": "Token refreshed successfully",
  "data": {
    "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "expires_in": 900,
    "token_type": "Bearer"
  }
}
```

#### 4. Get Current User (Protected)

```
GET /api/v1/auth/me
Authorization: Bearer <access_token>
```

**Response**:
```json
{
  "success": true,
  "message": "User information retrieved successfully",
  "data": {
    "id": "uuid",
    "email": "user@example.com",
    "google_id": "google-id",
    "youtube_channel_id": "channel-id",
    "youtube_channel_name": "Channel Name",
    "profile_image_url": "https://...",
    "created_at": "2024-01-01T00:00:00Z"
  }
}
```

#### 5. Logout (Protected)

```
POST /api/v1/auth/logout
Authorization: Bearer <access_token>
```

**Response**:
```json
{
  "success": true,
  "message": "Logged out successfully"
}
```

## Error Responses

All error responses follow this format:

```json
{
  "error": "error_code",
  "message": "Human-readable error message",
  "details": {} // optional additional details
}
```

### HTTP Status Codes

- `200` - Success
- `400` - Bad Request (invalid input)
- `401` - Unauthorized (invalid/expired token)
- `404` - Not Found
- `429` - Too Many Requests (rate limit exceeded)
- `500` - Internal Server Error

### Rate Limiting Headers

```
X-RateLimit-Limit: 60
X-RateLimit-Remaining: 45
X-RateLimit-Reset: 1704067200
Retry-After: 60  // Only present when rate limited
```

## Security Features

### JWT Token Security

- **Access Tokens**: Short-lived (15 minutes) for API requests
- **Refresh Tokens**: Long-lived (7 days) for obtaining new access tokens
- **Token Blacklisting**: Redis-based blacklist for logout functionality
- **Token Type Validation**: Ensures refresh tokens aren't used as access tokens

### OAuth Token Security

- **AES-256-GCM Encryption**: OAuth tokens encrypted before database storage
- **CSRF Protection**: State parameter validation in OAuth flow
- **One-Time State**: OAuth state tokens are single-use with 10-minute TTL

### API Security

- **Rate Limiting**: IP-based sliding window rate limiter
- **CORS**: Configurable allowed origins
- **Request Logging**: All requests logged with IP and user agent
- **Panic Recovery**: Automatic recovery from panics with error logging

## Development

### Run Tests

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run tests with verbose output
go test -v ./...
```

### Run with Auto-Reload

```bash
# Install air
go install github.com/cosmtrek/air@latest

# Run with hot reload
air
```

### Database Migrations

```bash
# Create new migration
migrate create -ext sql -dir migrations -seq migration_name

# Apply migrations
migrate -database $DATABASE_URL -path migrations up

# Rollback migration
migrate -database $DATABASE_URL -path migrations down 1

# Check migration version
migrate -database $DATABASE_URL -path migrations version
```

## Production Deployment

### 1. Build for Production

```bash
CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o bin/api ./cmd/api
```

### 2. Environment Variables

Set `ENV=production` and ensure:
- Strong JWT secret (32+ characters)
- Strong encryption key (exactly 32 characters)
- Secure database credentials
- HTTPS redirect URLs
- Proper CORS origins

### 3. Database Connection Pool

Recommended settings for production:
```env
DB_MAX_IDLE_CONNS=25
DB_MAX_OPEN_CONNS=100
DB_CONN_MAX_LIFETIME=1h
```

### 4. Redis Connection

Use Redis with persistence (AOF or RDB) for production.

## Monitoring

### Health Check

```bash
curl http://localhost:8080/health
```

### Logs

Structured logs with levels (DEBUG, INFO, WARN, ERROR, FATAL):

```
[2024-01-01 12:00:00] INFO: Starting video-upload-backend service
[2024-01-01 12:00:01] INFO: Database migrations completed successfully
[2024-01-01 12:00:02] INFO: Video upload backend service started successfully port=8080 environment=production
```

## License

MIT

## Support

For issues and questions, please create an issue in the repository.
