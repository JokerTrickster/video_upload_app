package domain

import "errors"

var (
	// User errors
	ErrUserNotFound      = errors.New("user not found")
	ErrUserAlreadyExists = errors.New("user already exists")
	ErrInvalidUserData   = errors.New("invalid user data")

	// Token errors
	ErrTokenNotFound    = errors.New("token not found")
	ErrTokenExpired     = errors.New("token expired")
	ErrTokenInvalid     = errors.New("token invalid")
	ErrTokenBlacklisted = errors.New("token blacklisted")

	// Auth errors
	ErrUnauthorized       = errors.New("unauthorized")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrInvalidState       = errors.New("invalid state parameter")
	ErrOAuthFailed        = errors.New("oauth authentication failed")

	// General errors
	ErrInternalServer = errors.New("internal server error")
	ErrInvalidInput   = errors.New("invalid input")

	// Media Asset errors
	ErrMediaAssetNotFound      = errors.New("media asset not found")
	ErrMediaAssetAlreadyExists = errors.New("media asset already exists")
	ErrInvalidMediaType        = errors.New("invalid media type")
	ErrInvalidSyncStatus       = errors.New("invalid sync status")
	ErrUploadInProgress        = errors.New("upload already in progress")
	ErrMaxRetriesExceeded      = errors.New("maximum retry attempts exceeded")

	// Upload Session errors
	ErrSessionNotFound       = errors.New("upload session not found")
	ErrSessionAlreadyEnded   = errors.New("session already completed or cancelled")
	ErrInvalidSessionStatus  = errors.New("invalid session status")
	ErrSessionMismatch       = errors.New("asset does not belong to this session")
)
