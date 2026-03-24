package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// SessionStatus represents the status of an upload session
type SessionStatus string

const (
	SessionStatusActive    SessionStatus = "ACTIVE"
	SessionStatusCompleted SessionStatus = "COMPLETED"
	SessionStatusCancelled SessionStatus = "CANCELLED"
)

// UploadSession represents a batch upload session tracking multiple file uploads
type UploadSession struct {
	SessionID      uuid.UUID     `json:"session_id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID         uuid.UUID     `json:"user_id" gorm:"type:uuid;not null;index"`
	TotalFiles     int           `json:"total_files" gorm:"not null;default:0"`
	CompletedFiles int           `json:"completed_files" gorm:"not null;default:0"`
	FailedFiles    int           `json:"failed_files" gorm:"not null;default:0"`
	TotalBytes     int64         `json:"total_bytes" gorm:"not null;default:0"`
	UploadedBytes  int64         `json:"uploaded_bytes" gorm:"not null;default:0"`
	SessionStatus  SessionStatus `json:"session_status" gorm:"type:varchar(20);not null;default:ACTIVE"`
	StartedAt      time.Time     `json:"started_at" gorm:"not null;default:now()"`
	CompletedAt    *time.Time    `json:"completed_at,omitempty" gorm:""`
	CreatedAt      time.Time     `json:"created_at" gorm:"not null;default:now()"`
	UpdatedAt      time.Time     `json:"updated_at" gorm:"not null;default:now()"`
}

// TableName specifies the table name for GORM
func (UploadSession) TableName() string {
	return "upload_sessions"
}

// BeforeCreate hook to set default values
func (s *UploadSession) BeforeCreate() error {
	if s.SessionID == uuid.Nil {
		s.SessionID = uuid.New()
	}
	now := time.Now()
	s.CreatedAt = now
	s.UpdatedAt = now
	s.StartedAt = now

	// Set default session status if not specified
	if s.SessionStatus == "" {
		s.SessionStatus = SessionStatusActive
	}

	return s.Validate()
}

// BeforeUpdate hook to update timestamp
func (s *UploadSession) BeforeUpdate() error {
	s.UpdatedAt = time.Now()
	return s.Validate()
}

// Validate checks if the upload session data is valid
func (s *UploadSession) Validate() error {
	if s.UserID == uuid.Nil {
		return errors.New("user_id is required")
	}

	if s.TotalFiles < 0 {
		return errors.New("total_files cannot be negative")
	}

	if s.CompletedFiles < 0 {
		return errors.New("completed_files cannot be negative")
	}

	if s.FailedFiles < 0 {
		return errors.New("failed_files cannot be negative")
	}

	if s.CompletedFiles > s.TotalFiles {
		return errors.New("completed_files cannot exceed total_files")
	}

	if s.FailedFiles > s.TotalFiles {
		return errors.New("failed_files cannot exceed total_files")
	}

	if s.TotalBytes < 0 {
		return errors.New("total_bytes cannot be negative")
	}

	if s.UploadedBytes < 0 {
		return errors.New("uploaded_bytes cannot be negative")
	}

	if s.UploadedBytes > s.TotalBytes {
		return errors.New("uploaded_bytes cannot exceed total_bytes")
	}

	// Validate session status
	validStatuses := map[SessionStatus]bool{
		SessionStatusActive:    true,
		SessionStatusCompleted: true,
		SessionStatusCancelled: true,
	}
	if !validStatuses[s.SessionStatus] {
		return errors.New("invalid session_status")
	}

	return nil
}

// IsActive returns true if the session is currently active
func (s *UploadSession) IsActive() bool {
	return s.SessionStatus == SessionStatusActive
}

// IsCompleted returns true if the session is completed
func (s *UploadSession) IsCompleted() bool {
	return s.SessionStatus == SessionStatusCompleted
}

// IsCancelled returns true if the session is cancelled
func (s *UploadSession) IsCancelled() bool {
	return s.SessionStatus == SessionStatusCancelled
}

// CalculateProgress returns the upload progress as a percentage (0-100)
func (s *UploadSession) CalculateProgress() float64 {
	if s.TotalBytes == 0 {
		return 0.0
	}
	return float64(s.UploadedBytes) / float64(s.TotalBytes) * 100.0
}

// GetPendingFiles returns the number of files still pending upload
func (s *UploadSession) GetPendingFiles() int {
	return s.TotalFiles - s.CompletedFiles - s.FailedFiles
}

// IncrementCompleted increments the completed files counter and uploaded bytes
func (s *UploadSession) IncrementCompleted(fileBytes int64) error {
	if fileBytes < 0 {
		return errors.New("file_bytes cannot be negative")
	}

	s.CompletedFiles++
	s.UploadedBytes += fileBytes

	// Auto-complete session if all files are processed
	if s.CompletedFiles+s.FailedFiles == s.TotalFiles && s.IsActive() {
		s.Complete()
	}

	return nil
}

// IncrementFailed increments the failed files counter
func (s *UploadSession) IncrementFailed() {
	s.FailedFiles++

	// Auto-complete session if all files are processed
	if s.CompletedFiles+s.FailedFiles == s.TotalFiles && s.IsActive() {
		s.Complete()
	}
}

// Complete marks the session as completed
func (s *UploadSession) Complete() {
	if s.SessionStatus == SessionStatusActive {
		s.SessionStatus = SessionStatusCompleted
		now := time.Now()
		s.CompletedAt = &now
	}
}

// Cancel marks the session as cancelled
func (s *UploadSession) Cancel() {
	if s.SessionStatus == SessionStatusActive {
		s.SessionStatus = SessionStatusCancelled
		now := time.Now()
		s.CompletedAt = &now
	}
}

// GetSuccessRate returns the success rate as a percentage (0-100)
func (s *UploadSession) GetSuccessRate() float64 {
	processedFiles := s.CompletedFiles + s.FailedFiles
	if processedFiles == 0 {
		return 0.0
	}
	return float64(s.CompletedFiles) / float64(processedFiles) * 100.0
}

// GetDuration returns the duration of the session
func (s *UploadSession) GetDuration() time.Duration {
	endTime := time.Now()
	if s.CompletedAt != nil {
		endTime = *s.CompletedAt
	}
	return endTime.Sub(s.StartedAt)
}
