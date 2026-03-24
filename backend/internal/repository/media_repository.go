package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/JokerTrickster/video-upload-backend/internal/domain"
)

// mediaRepository implements MediaRepository interface
type mediaRepository struct {
	db *gorm.DB
}

// NewMediaRepository creates a new media repository instance
func NewMediaRepository(db *gorm.DB) MediaRepository {
	return &mediaRepository{db: db}
}

// Create creates a new media asset
func (r *mediaRepository) Create(ctx context.Context, asset *domain.MediaAsset) error {
	if err := r.db.WithContext(ctx).Create(asset).Error; err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return domain.ErrMediaAssetAlreadyExists
		}
		return err
	}
	return nil
}

// FindByID retrieves a media asset by ID
func (r *mediaRepository) FindByID(ctx context.Context, assetID string) (*domain.MediaAsset, error) {
	id, err := uuid.Parse(assetID)
	if err != nil {
		return nil, domain.ErrInvalidInput
	}

	var asset domain.MediaAsset
	if err := r.db.WithContext(ctx).
		Where("asset_id = ?", id).
		First(&asset).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrMediaAssetNotFound
		}
		return nil, err
	}

	return &asset, nil
}

// FindByUserID retrieves media assets for a user with pagination
func (r *mediaRepository) FindByUserID(ctx context.Context, userID string, limit, offset int) ([]domain.MediaAsset, int64, error) {
	id, err := uuid.Parse(userID)
	if err != nil {
		return nil, 0, domain.ErrInvalidInput
	}

	var assets []domain.MediaAsset
	var total int64

	// Get total count
	if err := r.db.WithContext(ctx).
		Model(&domain.MediaAsset{}).
		Where("user_id = ?", id).
		Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get paginated results
	if err := r.db.WithContext(ctx).
		Where("user_id = ?", id).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&assets).Error; err != nil {
		return nil, 0, err
	}

	return assets, total, nil
}

// FindPendingUploads retrieves all pending uploads for a user
func (r *mediaRepository) FindPendingUploads(ctx context.Context, userID string) ([]domain.MediaAsset, error) {
	id, err := uuid.Parse(userID)
	if err != nil {
		return nil, domain.ErrInvalidInput
	}

	var assets []domain.MediaAsset
	if err := r.db.WithContext(ctx).
		Where("user_id = ? AND sync_status = ?", id, domain.SyncStatusPending).
		Order("created_at ASC").
		Find(&assets).Error; err != nil {
		return nil, err
	}

	return assets, nil
}

// Update updates a media asset
func (r *mediaRepository) Update(ctx context.Context, asset *domain.MediaAsset) error {
	if err := r.db.WithContext(ctx).
		Model(asset).
		Where("asset_id = ?", asset.AssetID).
		Updates(asset).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domain.ErrMediaAssetNotFound
		}
		return err
	}

	return nil
}

// Delete deletes a media asset
func (r *mediaRepository) Delete(ctx context.Context, assetID string) error {
	id, err := uuid.Parse(assetID)
	if err != nil {
		return domain.ErrInvalidInput
	}

	result := r.db.WithContext(ctx).
		Where("asset_id = ?", id).
		Delete(&domain.MediaAsset{})

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return domain.ErrMediaAssetNotFound
	}

	return nil
}

// CountByUserID returns the total count of media assets for a user
func (r *mediaRepository) CountByUserID(ctx context.Context, userID string) (int64, error) {
	id, err := uuid.Parse(userID)
	if err != nil {
		return 0, domain.ErrInvalidInput
	}

	var count int64
	if err := r.db.WithContext(ctx).
		Model(&domain.MediaAsset{}).
		Where("user_id = ?", id).
		Count(&count).Error; err != nil {
		return 0, err
	}

	return count, nil
}

// FindByYouTubeVideoID retrieves a media asset by YouTube video ID
func (r *mediaRepository) FindByYouTubeVideoID(ctx context.Context, videoID string) (*domain.MediaAsset, error) {
	if videoID == "" {
		return nil, domain.ErrInvalidInput
	}

	var asset domain.MediaAsset
	if err := r.db.WithContext(ctx).
		Where("youtube_video_id = ?", videoID).
		First(&asset).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrMediaAssetNotFound
		}
		return nil, err
	}

	return &asset, nil
}
