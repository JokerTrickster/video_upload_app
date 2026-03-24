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
