package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/JokerTrickster/video-upload-backend/internal/domain"
	"github.com/JokerTrickster/video-upload-backend/internal/pkg/logger"
	"github.com/JokerTrickster/video-upload-backend/internal/pkg/youtube"
	"github.com/JokerTrickster/video-upload-backend/internal/repository"
)

type QueueService interface {
	// AddToQueue adds a video to the upload queue
	AddToQueue(ctx context.Context, userID uuid.UUID, filePath, filename string, fileSizeBytes int64, title, description string) (*domain.UploadQueueItem, error)

	// GetQueueItems returns queued items for a user
	GetQueueItems(ctx context.Context, userID uuid.UUID, page, limit int) ([]domain.UploadQueueItem, int64, error)

	// RemoveFromQueue removes an item from the queue (only if PENDING)
	RemoveFromQueue(ctx context.Context, queueID uuid.UUID) error

	// GetQuotaStatus returns today's quota usage
	GetQuotaStatus(ctx context.Context) (*domain.DailyQuota, error)

	// ProcessQueue processes pending queue items within daily quota
	// This is called by the scheduler, not by API handlers
	ProcessQueue(ctx context.Context) error
}

type queueService struct {
	queueRepo     repository.QueueRepository
	mediaRepo     repository.MediaRepository
	tokenRepo     repository.TokenRepository
	tokenService  TokenService
	youtubeClient youtube.Client
}

func NewQueueService(
	queueRepo repository.QueueRepository,
	mediaRepo repository.MediaRepository,
	tokenRepo repository.TokenRepository,
	tokenService TokenService,
	youtubeClient youtube.Client,
) QueueService {
	return &queueService{
		queueRepo:     queueRepo,
		mediaRepo:     mediaRepo,
		tokenRepo:     tokenRepo,
		tokenService:  tokenService,
		youtubeClient: youtubeClient,
	}
}

func (s *queueService) AddToQueue(ctx context.Context, userID uuid.UUID, filePath, filename string, fileSizeBytes int64, title, description string) (*domain.UploadQueueItem, error) {
	item := &domain.UploadQueueItem{
		UserID:        userID,
		FilePath:      filePath,
		Filename:      filename,
		FileSizeBytes: fileSizeBytes,
		Title:         title,
		Description:   description,
		QueueStatus:   domain.QueueStatusPending,
	}

	if err := s.queueRepo.Create(ctx, item); err != nil {
		return nil, fmt.Errorf("failed to add to queue: %w", err)
	}
	return item, nil
}

func (s *queueService) GetQueueItems(ctx context.Context, userID uuid.UUID, page, limit int) ([]domain.UploadQueueItem, int64, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 50
	}
	offset := (page - 1) * limit
	return s.queueRepo.FindByUserID(ctx, userID.String(), limit, offset)
}

func (s *queueService) RemoveFromQueue(ctx context.Context, queueID uuid.UUID) error {
	item, err := s.queueRepo.FindByID(ctx, queueID.String())
	if err != nil {
		return err
	}
	if item.QueueStatus != domain.QueueStatusPending {
		return fmt.Errorf("can only remove PENDING items, current status: %s", item.QueueStatus)
	}
	return s.queueRepo.DeleteByID(ctx, queueID.String())
}

func (s *queueService) GetQuotaStatus(ctx context.Context) (*domain.DailyQuota, error) {
	today := time.Now().Format("2006-01-02")
	return s.queueRepo.GetOrCreateDailyQuota(ctx, today)
}

func (s *queueService) ProcessQueue(ctx context.Context) error {
	today := time.Now().Format("2006-01-02")

	// Get or create today's quota
	quota, err := s.queueRepo.GetOrCreateDailyQuota(ctx, today)
	if err != nil {
		return fmt.Errorf("failed to get daily quota: %w", err)
	}

	if !quota.CanUpload() {
		logger.Info("Daily quota exhausted, skipping queue processing",
			"date", today, "units_used", quota.UnitsUsed, "units_max", quota.UnitsMax)
		return nil
	}

	remainingUploads := quota.RemainingUploads()
	logger.Info("Processing upload queue",
		"date", today, "remaining_uploads", remainingUploads)

	// Get pending items (limited by remaining quota)
	items, err := s.queueRepo.FindPendingAll(ctx, remainingUploads)
	if err != nil {
		return fmt.Errorf("failed to get pending items: %w", err)
	}

	if len(items) == 0 {
		logger.Info("No pending items in queue")
		return nil
	}

	logger.Info("Found pending items", "count", len(items))

	for _, item := range items {
		// Re-check quota before each upload
		if !quota.CanUpload() {
			logger.Info("Quota exhausted during processing, stopping")
			break
		}

		if err := s.processItem(ctx, &item, quota); err != nil {
			logger.Error("Failed to process queue item",
				"queue_id", item.QueueID, "error", err)
			continue
		}

		// Update quota
		quota.UnitsUsed += domain.YouTubeUploadCost
		quota.Uploads++
		if err := s.queueRepo.UpdateDailyQuota(ctx, quota); err != nil {
			logger.Error("Failed to update quota", "error", err)
		}
	}

	logger.Info("Queue processing complete",
		"date", today, "uploads_today", quota.Uploads, "units_used", quota.UnitsUsed)
	return nil
}

func (s *queueService) processItem(ctx context.Context, item *domain.UploadQueueItem, quota *domain.DailyQuota) error {
	// Mark as processing
	item.QueueStatus = domain.QueueStatusProcessing
	now := time.Now()
	item.ProcessedAt = &now
	if err := s.queueRepo.Update(ctx, item); err != nil {
		return err
	}

	// Get user's OAuth token
	token, err := s.tokenRepo.FindByUserID(ctx, item.UserID.String())
	if err != nil {
		return s.markFailed(ctx, item, "OAuth token not found: "+err.Error())
	}

	if token.IsExpired() {
		return s.markFailed(ctx, item, "OAuth token expired, user needs to re-authenticate")
	}

	// Decrypt access token
	accessToken, err := s.tokenService.DecryptToken(ctx, token.EncryptedAccessToken)
	if err != nil {
		return s.markFailed(ctx, item, "Failed to decrypt token: "+err.Error())
	}

	// Create media asset record
	asset := &domain.MediaAsset{
		UserID:           item.UserID,
		OriginalFilename: item.Filename,
		FileSizeBytes:    item.FileSizeBytes,
		MediaType:        "VIDEO",
		SyncStatus:       "UPLOADING",
		UploadStartedAt:  &now,
	}
	if err := s.mediaRepo.Create(ctx, asset); err != nil {
		return s.markFailed(ctx, item, "Failed to create asset: "+err.Error())
	}

	// Upload to YouTube
	title := item.Title
	if title == "" {
		title = item.Filename
	}
	uploadReq := &youtube.UploadVideoRequest{
		FilePath:      item.FilePath,
		Title:         title,
		Description:   item.Description,
		PrivacyStatus: "private",
	}

	uploadResp, err := s.youtubeClient.UploadVideo(ctx, accessToken, uploadReq)
	if err != nil {
		// Mark asset as failed
		asset.SyncStatus = "FAILED"
		errMsg := err.Error()
		asset.ErrorMessage = &errMsg
		_ = s.mediaRepo.Update(ctx, asset)

		item.RetryCount++
		if item.RetryCount >= domain.MaxRetryAttempts {
			return s.markFailed(ctx, item, "Max retries exceeded: "+err.Error())
		}
		// Return to pending for retry next day
		item.QueueStatus = domain.QueueStatusPending
		return s.queueRepo.Update(ctx, item)
	}

	// Success - update asset
	asset.YouTubeVideoID = &uploadResp.VideoID
	if uploadResp.ThumbnailURL != "" {
		asset.ThumbnailURL = &uploadResp.ThumbnailURL
	}
	asset.SyncStatus = "COMPLETED"
	completedAt := time.Now()
	asset.UploadCompletedAt = &completedAt
	_ = s.mediaRepo.Update(ctx, asset)

	// Mark queue item as completed
	item.QueueStatus = domain.QueueStatusCompleted
	item.AssetID = &asset.AssetID
	return s.queueRepo.Update(ctx, item)
}

func (s *queueService) markFailed(ctx context.Context, item *domain.UploadQueueItem, errMsg string) error {
	item.QueueStatus = domain.QueueStatusFailed
	item.ErrorMessage = &errMsg
	return s.queueRepo.Update(ctx, item)
}
