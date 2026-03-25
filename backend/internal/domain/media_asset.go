package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// MediaType represents the type of media asset
type MediaType string

const (
	MediaTypeVideo MediaType = "VIDEO"
	MediaTypeImage MediaType = "IMAGE"
)

// SyncStatus represents the synchronization status of a media asset
type SyncStatus string

const (
	SyncStatusPending   SyncStatus = "PENDING"
	SyncStatusUploading SyncStatus = "UPLOADING"
	SyncStatusCompleted SyncStatus = "COMPLETED"
	SyncStatusFailed    SyncStatus = "FAILED"
)

// MediaAsset represents a backed up media file (video or image)
type MediaAsset struct {
	AssetID           uuid.UUID  `json:"asset_id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID            uuid.UUID  `json:"user_id" gorm:"type:uuid;not null;index"`
	YouTubeVideoID    *string    `json:"youtube_video_id,omitempty" gorm:"type:varchar(255);uniqueIndex"`
	S3ObjectKey       *string    `json:"s3_object_key,omitempty" gorm:"type:varchar(512)"`
	OriginalFilename  string     `json:"original_filename" gorm:"type:varchar(512);not null"`
	FileSizeBytes     int64      `json:"file_size_bytes" gorm:"not null"`
	MediaType         MediaType  `json:"media_type" gorm:"type:varchar(10);not null"`
	SyncStatus        SyncStatus `json:"sync_status" gorm:"type:varchar(20);not null;default:PENDING"`
	UploadStartedAt   *time.Time `json:"upload_started_at,omitempty" gorm:""`
	UploadCompletedAt *time.Time `json:"upload_completed_at,omitempty" gorm:""`
	ErrorMessage      *string    `json:"error_message,omitempty" gorm:"type:text"`
	RetryCount        int        `json:"retry_count" gorm:"not null;default:0"`
	CreatedAt         time.Time  `json:"created_at" gorm:"not null;default:now()"`
	UpdatedAt         time.Time  `json:"updated_at" gorm:"not null;default:now()"`
	DeletedAt         *time.Time `json:"deleted_at,omitempty" gorm:"index"`
}

// TableName specifies the table name for GORM
func (MediaAsset) TableName() string {
	return "media_assets"
}

// BeforeCreate hook to set default values
func (m *MediaAsset) BeforeCreate() error {
	if m.AssetID == uuid.Nil {
		m.AssetID = uuid.New()
	}
	now := time.Now()
	m.CreatedAt = now
	m.UpdatedAt = now

	// Set default sync status if not specified
	if m.SyncStatus == "" {
		m.SyncStatus = SyncStatusPending
	}

	return m.Validate()
}

// BeforeUpdate hook to update timestamp
func (m *MediaAsset) BeforeUpdate() error {
	m.UpdatedAt = time.Now()
	return m.Validate()
}

// Validate checks if the media asset data is valid
func (m *MediaAsset) Validate() error {
	if m.UserID == uuid.Nil {
		return errors.New("user_id is required")
	}

	if m.OriginalFilename == "" {
		return errors.New("original_filename is required")
	}

	if m.FileSizeBytes <= 0 {
		return errors.New("file_size_bytes must be greater than 0")
	}

	// Validate media type
	if m.MediaType != MediaTypeVideo && m.MediaType != MediaTypeImage {
		return errors.New("media_type must be VIDEO or IMAGE")
	}

	// Validate sync status
	validStatuses := map[SyncStatus]bool{
		SyncStatusPending:   true,
		SyncStatusUploading: true,
		SyncStatusCompleted: true,
		SyncStatusFailed:    true,
	}
	if !validStatuses[m.SyncStatus] {
		return errors.New("invalid sync_status")
	}

	// Validate storage location based on media type
	if m.SyncStatus == SyncStatusCompleted {
		if m.MediaType == MediaTypeVideo && (m.YouTubeVideoID == nil || *m.YouTubeVideoID == "") {
			return errors.New("youtube_video_id is required for completed video uploads")
		}
		if m.MediaType == MediaTypeImage && (m.S3ObjectKey == nil || *m.S3ObjectKey == "") {
			return errors.New("s3_object_key is required for completed image uploads")
		}
	}

	if m.RetryCount < 0 {
		return errors.New("retry_count cannot be negative")
	}

	return nil
}

// CanRetry determines if the asset can be retried based on retry count and status
func (m *MediaAsset) CanRetry() bool {
	const maxRetries = 5
	return m.SyncStatus == SyncStatusFailed && m.RetryCount < maxRetries
}

// MarkAsUploading updates the asset status to uploading
func (m *MediaAsset) MarkAsUploading() {
	m.SyncStatus = SyncStatusUploading
	now := time.Now()
	if m.UploadStartedAt == nil {
		m.UploadStartedAt = &now
	}
}

// MarkAsCompleted updates the asset status to completed with storage location
func (m *MediaAsset) MarkAsCompleted(storageID string) error {
	if storageID == "" {
		return errors.New("storage_id is required")
	}

	m.SyncStatus = SyncStatusCompleted
	now := time.Now()
	m.UploadCompletedAt = &now

	// Set the appropriate storage location based on media type
	if m.MediaType == MediaTypeVideo {
		m.YouTubeVideoID = &storageID
	} else if m.MediaType == MediaTypeImage {
		m.S3ObjectKey = &storageID
	}

	// Clear error message on success
	m.ErrorMessage = nil

	return nil
}

// MarkAsFailed updates the asset status to failed with error message
func (m *MediaAsset) MarkAsFailed(errMsg string) {
	m.SyncStatus = SyncStatusFailed
	m.ErrorMessage = &errMsg
	m.RetryCount++
}

// IsCompleted returns true if the asset upload is completed
func (m *MediaAsset) IsCompleted() bool {
	return m.SyncStatus == SyncStatusCompleted
}

// IsPending returns true if the asset is pending upload
func (m *MediaAsset) IsPending() bool {
	return m.SyncStatus == SyncStatusPending
}

// IsUploading returns true if the asset is currently uploading
func (m *MediaAsset) IsUploading() bool {
	return m.SyncStatus == SyncStatusUploading
}

// IsFailed returns true if the asset upload failed
func (m *MediaAsset) IsFailed() bool {
	return m.SyncStatus == SyncStatusFailed
}

// GetStorageID returns the appropriate storage ID based on media type
func (m *MediaAsset) GetStorageID() *string {
	if m.MediaType == MediaTypeVideo {
		return m.YouTubeVideoID
	} else if m.MediaType == MediaTypeImage {
		return m.S3ObjectKey
	}
	return nil
}
