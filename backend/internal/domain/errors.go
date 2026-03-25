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

	// Upload errors (API error codes)
	ErrFileTooLarge         = errors.New("file size exceeds maximum allowed size")
	ErrInvalidFileFormat    = errors.New("unsupported file format")
	ErrYouTubeQuotaExceeded = errors.New("youtube API quota exceeded")
	ErrNetworkTimeout       = errors.New("network timeout during upload")
	ErrInsufficientStorage  = errors.New("insufficient storage space")
	ErrVerificationFailed   = errors.New("video verification failed")
)

// Error codes for API responses
const (
	// Authentication errors
	ErrorCodeAuthInvalid       = "AUTH_001"
	ErrorCodeTokenExpired      = "AUTH_002"

	// Upload errors
	ErrorCodeFileTooLarge      = "UPLOAD_001"
	ErrorCodeInvalidFormat     = "UPLOAD_002"
	ErrorCodeQuotaExceeded     = "UPLOAD_003"
	ErrorCodeNetworkTimeout    = "UPLOAD_004"

	// Storage errors
	ErrorCodeInsufficientSpace = "STORAGE_001"

	// Sync errors
	ErrorCodeVerificationFailed = "SYNC_001"
)

// Retry configuration
const (
	// MaxRetryAttempts is the maximum number of retry attempts
	MaxRetryAttempts = 5

	// Retry delays (exponential backoff)
	RetryDelay1 = 1 * 60        // 1 minute
	RetryDelay2 = 5 * 60        // 5 minutes
	RetryDelay3 = 15 * 60       // 15 minutes
	RetryDelay4 = 60 * 60       // 1 hour
	RetryDelay5 = 24 * 60 * 60  // 24 hours
)

// GetRetryDelay returns the retry delay for a given attempt number
func GetRetryDelay(attempt int) int {
	switch attempt {
	case 1:
		return RetryDelay1
	case 2:
		return RetryDelay2
	case 3:
		return RetryDelay3
	case 4:
		return RetryDelay4
	case 5:
		return RetryDelay5
	default:
		return 0
	}
}

// ShouldRetry determines if an error is retryable
func ShouldRetry(err error) bool {
	if err == nil {
		return false
	}

	// Retry on specific errors
	switch {
	case errors.Is(err, ErrNetworkTimeout):
		return true
	case errors.Is(err, ErrYouTubeQuotaExceeded):
		return true
	case errors.Is(err, ErrInternalServer):
		return true
	default:
		return false
	}
}
