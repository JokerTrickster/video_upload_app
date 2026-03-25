package repository

import (
	"context"

	"github.com/JokerTrickster/video-upload-backend/internal/domain"
)

// UserRepository defines the interface for user data access
type UserRepository interface {
	// Create creates a new user
	Create(ctx context.Context, user *domain.User) error

	// FindByID retrieves a user by ID
	FindByID(ctx context.Context, id string) (*domain.User, error)

	// FindByEmail retrieves a user by email
	FindByEmail(ctx context.Context, email string) (*domain.User, error)

	// FindByGoogleID retrieves a user by Google ID
	FindByGoogleID(ctx context.Context, googleID string) (*domain.User, error)

	// Update updates a user
	Update(ctx context.Context, user *domain.User) error

	// Delete soft deletes a user
	Delete(ctx context.Context, id string) error
}

// TokenRepository defines the interface for token data access
type TokenRepository interface {
	// Create creates a new token
	Create(ctx context.Context, token *domain.Token) error

	// FindByUserID retrieves a token by user ID
	FindByUserID(ctx context.Context, userID string) (*domain.Token, error)

	// Update updates a token
	Update(ctx context.Context, token *domain.Token) error

	// Delete deletes a token
	Delete(ctx context.Context, userID string) error
}

// MediaRepository defines the interface for media asset data access
type MediaRepository interface {
	// Create creates a new media asset
	Create(ctx context.Context, asset *domain.MediaAsset) error

	// FindByID retrieves a media asset by ID
	FindByID(ctx context.Context, assetID string) (*domain.MediaAsset, error)

	// FindByUserID retrieves media assets for a user with pagination, filtering, and sorting
	FindByUserID(ctx context.Context, userID string, limit, offset int, mediaType, syncStatus, sort string) ([]domain.MediaAsset, int64, error)

	// FindPendingUploads retrieves all pending uploads for a user
	FindPendingUploads(ctx context.Context, userID string) ([]domain.MediaAsset, error)

	// Update updates a media asset
	Update(ctx context.Context, asset *domain.MediaAsset) error

	// Delete deletes a media asset
	Delete(ctx context.Context, assetID string) error

	// CountByUserID returns the total count of media assets for a user
	CountByUserID(ctx context.Context, userID string) (int64, error)

	// FindByYouTubeVideoID retrieves a media asset by YouTube video ID
	FindByYouTubeVideoID(ctx context.Context, videoID string) (*domain.MediaAsset, error)
}

// SessionRepository defines the interface for upload session data access
type SessionRepository interface {
	// Create creates a new upload session
	Create(ctx context.Context, session *domain.UploadSession) error

	// FindByID retrieves an upload session by ID
	FindByID(ctx context.Context, sessionID string) (*domain.UploadSession, error)

	// FindByUserID retrieves upload sessions for a user with pagination
	FindByUserID(ctx context.Context, userID string, limit, offset int) ([]domain.UploadSession, int64, error)

	// FindActiveByUserID retrieves active upload sessions for a user
	FindActiveByUserID(ctx context.Context, userID string) ([]domain.UploadSession, error)

	// Update updates an upload session
	Update(ctx context.Context, session *domain.UploadSession) error

	// Delete deletes an upload session
	Delete(ctx context.Context, sessionID string) error
}
