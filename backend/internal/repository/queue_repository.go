package repository

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/JokerTrickster/video-upload-backend/internal/domain"
)

type QueueRepository interface {
	// Queue item operations
	Create(ctx context.Context, item *domain.UploadQueueItem) error
	FindPendingByUserID(ctx context.Context, userID string) ([]domain.UploadQueueItem, error)
	FindPendingAll(ctx context.Context, limit int) ([]domain.UploadQueueItem, error)
	Update(ctx context.Context, item *domain.UploadQueueItem) error
	FindByID(ctx context.Context, queueID string) (*domain.UploadQueueItem, error)
	FindByUserID(ctx context.Context, userID string, limit, offset int) ([]domain.UploadQueueItem, int64, error)
	DeleteByID(ctx context.Context, queueID string) error

	// Daily quota operations
	GetOrCreateDailyQuota(ctx context.Context, date string) (*domain.DailyQuota, error)
	UpdateDailyQuota(ctx context.Context, quota *domain.DailyQuota) error
}

type queueRepository struct {
	db *gorm.DB
}

func NewQueueRepository(db *gorm.DB) QueueRepository {
	return &queueRepository{db: db}
}

func (r *queueRepository) Create(ctx context.Context, item *domain.UploadQueueItem) error {
	return r.db.WithContext(ctx).Create(item).Error
}

func (r *queueRepository) FindPendingByUserID(ctx context.Context, userID string) ([]domain.UploadQueueItem, error) {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, domain.ErrInvalidInput
	}

	var items []domain.UploadQueueItem
	err = r.db.WithContext(ctx).
		Where("user_id = ? AND queue_status = ?", uid, domain.QueueStatusPending).
		Order("priority DESC, created_at ASC").
		Find(&items).Error
	return items, err
}

func (r *queueRepository) FindPendingAll(ctx context.Context, limit int) ([]domain.UploadQueueItem, error) {
	var items []domain.UploadQueueItem
	err := r.db.WithContext(ctx).
		Where("queue_status = ?", domain.QueueStatusPending).
		Order("priority DESC, created_at ASC").
		Limit(limit).
		Find(&items).Error
	return items, err
}

func (r *queueRepository) Update(ctx context.Context, item *domain.UploadQueueItem) error {
	item.UpdatedAt = time.Now()
	return r.db.WithContext(ctx).
		Model(item).
		Where("queue_id = ?", item.QueueID).
		Updates(item).Error
}

func (r *queueRepository) FindByID(ctx context.Context, queueID string) (*domain.UploadQueueItem, error) {
	id, err := uuid.Parse(queueID)
	if err != nil {
		return nil, domain.ErrInvalidInput
	}

	var item domain.UploadQueueItem
	if err := r.db.WithContext(ctx).Where("queue_id = ?", id).First(&item).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrMediaAssetNotFound
		}
		return nil, err
	}
	return &item, nil
}

func (r *queueRepository) FindByUserID(ctx context.Context, userID string, limit, offset int) ([]domain.UploadQueueItem, int64, error) {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, 0, domain.ErrInvalidInput
	}

	var items []domain.UploadQueueItem
	var total int64

	r.db.WithContext(ctx).Model(&domain.UploadQueueItem{}).
		Where("user_id = ?", uid).Count(&total)

	err = r.db.WithContext(ctx).
		Where("user_id = ?", uid).
		Order("created_at DESC").
		Limit(limit).Offset(offset).
		Find(&items).Error
	return items, total, err
}

func (r *queueRepository) DeleteByID(ctx context.Context, queueID string) error {
	id, err := uuid.Parse(queueID)
	if err != nil {
		return domain.ErrInvalidInput
	}

	result := r.db.WithContext(ctx).Where("queue_id = ?", id).Delete(&domain.UploadQueueItem{})
	if result.RowsAffected == 0 {
		return domain.ErrMediaAssetNotFound
	}
	return result.Error
}

func (r *queueRepository) GetOrCreateDailyQuota(ctx context.Context, date string) (*domain.DailyQuota, error) {
	var quota domain.DailyQuota
	err := r.db.WithContext(ctx).Where("date = ?", date).First(&quota).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		quota = domain.DailyQuota{
			ID:        uuid.New(),
			Date:      date,
			UnitsUsed: 0,
			UnitsMax:  domain.DailyQuotaLimit,
			Uploads:   0,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		if err := r.db.WithContext(ctx).Create(&quota).Error; err != nil {
			return nil, err
		}
		return &quota, nil
	}
	return &quota, err
}

func (r *queueRepository) UpdateDailyQuota(ctx context.Context, quota *domain.DailyQuota) error {
	quota.UpdatedAt = time.Now()
	return r.db.WithContext(ctx).Save(quota).Error
}
