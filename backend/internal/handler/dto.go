package handler

import (
	"time"

	"github.com/google/uuid"
)

// Request DTOs

// GetAuthURLRequest represents the request to get OAuth URL
type GetAuthURLRequest struct {
	RedirectURL string `json:"redirect_url,omitempty"`
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

// Response DTOs

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

// TokenRefreshResponse represents token refresh response
type TokenRefreshResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
	TokenType   string `json:"token_type"`
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

// Media Upload DTOs

// InitiateUploadRequest represents upload session initiation request
type InitiateUploadRequest struct {
	TotalFiles int                   `json:"total_files" binding:"required,min=1"`
	TotalBytes int64                 `json:"total_bytes" binding:"required,min=1"`
	MediaTypes map[string]int        `json:"media_types,omitempty"`
}

// InitiateUploadResponse represents upload session initiation response
type InitiateUploadResponse struct {
	SessionID string    `json:"session_id"`
	StartedAt time.Time `json:"started_at"`
}

// UploadVideoResponse represents video upload response
type UploadVideoResponse struct {
	AssetID           string    `json:"asset_id"`
	YouTubeVideoID    string    `json:"youtube_video_id"`
	OriginalFilename  string    `json:"original_filename"`
	FileSizeBytes     int64     `json:"file_size_bytes"`
	SyncStatus        string    `json:"sync_status"`
	UploadCompletedAt time.Time `json:"upload_completed_at"`
}

// UploadSessionStatusResponse represents upload session status
type UploadSessionStatusResponse struct {
	SessionID      string     `json:"session_id"`
	UserID         string     `json:"user_id"`
	TotalFiles     int        `json:"total_files"`
	CompletedFiles int        `json:"completed_files"`
	FailedFiles    int        `json:"failed_files"`
	TotalBytes     int64      `json:"total_bytes"`
	UploadedBytes  int64      `json:"uploaded_bytes"`
	SessionStatus  string     `json:"session_status"`
	StartedAt      time.Time  `json:"started_at"`
	CompletedAt    *time.Time `json:"completed_at,omitempty"`
}

// MediaAssetResponse represents a single media asset
type MediaAssetResponse struct {
	AssetID           string     `json:"asset_id"`
	YouTubeVideoID    *string    `json:"youtube_video_id,omitempty"`
	S3ObjectKey       *string    `json:"s3_object_key,omitempty"`
	OriginalFilename  string     `json:"original_filename"`
	FileSizeBytes     int64      `json:"file_size_bytes"`
	MediaType         string     `json:"media_type"`
	SyncStatus        string     `json:"sync_status"`
	CreatedAt         time.Time  `json:"created_at"`
	UploadStartedAt   *time.Time `json:"upload_started_at,omitempty"`
	UploadCompletedAt *time.Time `json:"upload_completed_at,omitempty"`
	ErrorMessage      *string    `json:"error_message,omitempty"`
	RetryCount        int        `json:"retry_count"`
}

// MediaAssetListResponse represents paginated media assets
type MediaAssetListResponse struct {
	Assets []MediaAssetResponse `json:"assets"`
	Pagination PaginationResponse `json:"pagination"`
}

// PaginationResponse represents pagination information
type PaginationResponse struct {
	Page       int   `json:"page"`
	Limit      int   `json:"limit"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"total_pages"`
}
