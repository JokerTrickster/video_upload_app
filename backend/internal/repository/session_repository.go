package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/JokerTrickster/video-upload-backend/internal/domain"
)

// sessionRepository implements SessionRepository interface
type sessionRepository struct {
	db *gorm.DB
}

// NewSessionRepository creates a new session repository instance
func NewSessionRepository(db *gorm.DB) SessionRepository {
	return &sessionRepository{db: db}
}

// Create creates a new upload session
func (r *sessionRepository) Create(ctx context.Context, session *domain.UploadSession) error {
	if err := r.db.WithContext(ctx).Create(session).Error; err != nil {
		return err
	}
	return nil
}

// FindByID retrieves an upload session by ID
func (r *sessionRepository) FindByID(ctx context.Context, sessionID string) (*domain.UploadSession, error) {
	id, err := uuid.Parse(sessionID)
	if err != nil {
		return nil, domain.ErrInvalidInput
	}

	var session domain.UploadSession
	if err := r.db.WithContext(ctx).
		Where("session_id = ?", id).
		First(&session).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrSessionNotFound
		}
		return nil, err
	}

	return &session, nil
}

// FindByUserID retrieves upload sessions for a user with pagination
func (r *sessionRepository) FindByUserID(ctx context.Context, userID string, limit, offset int) ([]domain.UploadSession, int64, error) {
	id, err := uuid.Parse(userID)
	if err != nil {
		return nil, 0, domain.ErrInvalidInput
	}

	var sessions []domain.UploadSession
	var total int64

	// Get total count
	if err := r.db.WithContext(ctx).
		Model(&domain.UploadSession{}).
		Where("user_id = ?", id).
		Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get paginated results
	if err := r.db.WithContext(ctx).
		Where("user_id = ?", id).
		Order("started_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&sessions).Error; err != nil {
		return nil, 0, err
	}

	return sessions, total, nil
}

// FindActiveByUserID retrieves active upload sessions for a user
func (r *sessionRepository) FindActiveByUserID(ctx context.Context, userID string) ([]domain.UploadSession, error) {
	id, err := uuid.Parse(userID)
	if err != nil {
		return nil, domain.ErrInvalidInput
	}

	var sessions []domain.UploadSession
	if err := r.db.WithContext(ctx).
		Where("user_id = ? AND session_status = ?", id, domain.SessionStatusActive).
		Order("started_at DESC").
		Find(&sessions).Error; err != nil {
		return nil, err
	}

	return sessions, nil
}

// Update updates an upload session
func (r *sessionRepository) Update(ctx context.Context, session *domain.UploadSession) error {
	if err := r.db.WithContext(ctx).
		Model(session).
		Where("session_id = ?", session.SessionID).
		Updates(session).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domain.ErrSessionNotFound
		}
		return err
	}

	return nil
}

// Delete deletes an upload session
func (r *sessionRepository) Delete(ctx context.Context, sessionID string) error {
	id, err := uuid.Parse(sessionID)
	if err != nil {
		return domain.ErrInvalidInput
	}

	result := r.db.WithContext(ctx).
		Where("session_id = ?", id).
		Delete(&domain.UploadSession{})

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return domain.ErrSessionNotFound
	}

	return nil
}
