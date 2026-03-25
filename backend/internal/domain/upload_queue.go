package domain

import (
	"time"

	"github.com/google/uuid"
)

// QueueStatus represents the status of a queued upload
type QueueStatus string

const (
	QueueStatusPending    QueueStatus = "PENDING"
	QueueStatusProcessing QueueStatus = "PROCESSING"
	QueueStatusCompleted  QueueStatus = "COMPLETED"
	QueueStatusFailed     QueueStatus = "FAILED"
	QueueStatusSkipped    QueueStatus = "SKIPPED" // skipped due to quota
)

// UploadQueueItem represents a video queued for automatic upload
type UploadQueueItem struct {
	QueueID       uuid.UUID   `json:"queue_id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID        uuid.UUID   `json:"user_id" gorm:"type:uuid;not null;index"`
	FilePath      string      `json:"file_path" gorm:"type:text;not null"`
	Filename      string      `json:"filename" gorm:"type:varchar(512);not null"`
	FileSizeBytes int64       `json:"file_size_bytes" gorm:"not null"`
	Title         string      `json:"title" gorm:"type:varchar(255)"`
	Description   string      `json:"description" gorm:"type:text"`
	QueueStatus   QueueStatus `json:"queue_status" gorm:"type:varchar(20);not null;default:PENDING;index"`
	Priority      int         `json:"priority" gorm:"not null;default:0"` // higher = process first
	AssetID       *uuid.UUID  `json:"asset_id,omitempty" gorm:"type:uuid"`
	ErrorMessage  *string     `json:"error_message,omitempty" gorm:"type:text"`
	RetryCount    int         `json:"retry_count" gorm:"not null;default:0"`
	ScheduledAt   *time.Time  `json:"scheduled_at,omitempty" gorm:""`
	ProcessedAt   *time.Time  `json:"processed_at,omitempty" gorm:""`
	CreatedAt     time.Time   `json:"created_at" gorm:"not null;default:now()"`
	UpdatedAt     time.Time   `json:"updated_at" gorm:"not null;default:now()"`
}

func (UploadQueueItem) TableName() string {
	return "upload_queue"
}

func (q *UploadQueueItem) BeforeCreate() error {
	if q.QueueID == uuid.Nil {
		q.QueueID = uuid.New()
	}
	now := time.Now()
	q.CreatedAt = now
	q.UpdatedAt = now
	if q.QueueStatus == "" {
		q.QueueStatus = QueueStatusPending
	}
	return nil
}

// DailyQuota tracks API usage per day
type DailyQuota struct {
	ID        uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Date      string    `json:"date" gorm:"type:varchar(10);uniqueIndex;not null"` // YYYY-MM-DD
	UnitsUsed int       `json:"units_used" gorm:"not null;default:0"`
	UnitsMax  int       `json:"units_max" gorm:"not null;default:10000"`
	Uploads   int       `json:"uploads" gorm:"not null;default:0"`
	CreatedAt time.Time `json:"created_at" gorm:"not null;default:now()"`
	UpdatedAt time.Time `json:"updated_at" gorm:"not null;default:now()"`
}

func (DailyQuota) TableName() string {
	return "daily_quotas"
}

const (
	// YouTubeUploadCost is the API quota cost per video upload
	YouTubeUploadCost = 1600
	// DailyQuotaLimit is the default YouTube API daily quota
	DailyQuotaLimit = 10000
	// MaxDailyUploads is the max uploads per day (10000/1600 = 6)
	MaxDailyUploads = 6
)

// CanUpload checks if there's enough quota remaining for one upload
func (d *DailyQuota) CanUpload() bool {
	return d.UnitsUsed+YouTubeUploadCost <= d.UnitsMax
}

// RemainingUploads returns how many more uploads are possible today
func (d *DailyQuota) RemainingUploads() int {
	remaining := (d.UnitsMax - d.UnitsUsed) / YouTubeUploadCost
	if remaining < 0 {
		return 0
	}
	return remaining
}
