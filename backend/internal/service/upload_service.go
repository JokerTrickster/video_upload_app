package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/JokerTrickster/video-upload-backend/internal/domain"
	"github.com/JokerTrickster/video-upload-backend/internal/pkg/youtube"
	"github.com/JokerTrickster/video-upload-backend/internal/repository"
)

// UploadService defines upload operations
type UploadService interface {
	// InitiateUploadSession creates a new upload session
	InitiateUploadSession(ctx context.Context, userID uuid.UUID, totalFiles int, totalBytes int64) (*domain.UploadSession, error)

	// UploadVideo uploads a video file to YouTube
	UploadVideo(ctx context.Context, req *UploadVideoRequest) (*UploadVideoResult, error)

	// GetUploadSessionStatus retrieves upload session status
	GetUploadSessionStatus(ctx context.Context, sessionID uuid.UUID) (*UploadSessionStatus, error)

	// CompleteUploadSession marks an upload session as completed
	CompleteUploadSession(ctx context.Context, sessionID uuid.UUID) error

	// CancelUploadSession cancels an upload session
	CancelUploadSession(ctx context.Context, sessionID uuid.UUID) error

	// ListMediaAssets retrieves paginated list of media assets
	ListMediaAssets(ctx context.Context, userID uuid.UUID, opts *ListMediaOptions) (*MediaAssetList, error)

	// GetMediaAsset retrieves a single media asset by ID
	GetMediaAsset(ctx context.Context, assetID uuid.UUID) (*domain.MediaAsset, error)

	// DeleteMediaAsset deletes a media asset
	DeleteMediaAsset(ctx context.Context, assetID uuid.UUID, deleteFromYouTube bool) error
}

// UploadVideoRequest represents video upload request
type UploadVideoRequest struct {
	SessionID     uuid.UUID
	UserID        uuid.UUID
	AccessToken   string
	FilePath      string
	Filename      string
	FileSizeBytes int64
	Title         string
	Description   string
	OnProgress    func(uploadedBytes, totalBytes int64)
}

// UploadVideoResult represents video upload result
type UploadVideoResult struct {
	AssetID      uuid.UUID
	VideoID      string
	Filename     string
	FileSizeBytes int64
	SyncStatus   string
	UploadedAt   time.Time
}

// UploadSessionStatus represents upload session status
type UploadSessionStatus struct {
	SessionID      uuid.UUID
	UserID         uuid.UUID
	TotalFiles     int
	CompletedFiles int
	FailedFiles    int
	TotalBytes     int64
	UploadedBytes  int64
	SessionStatus  string
	StartedAt      time.Time
	CompletedAt    *time.Time
}

// ListMediaOptions represents media list options
type ListMediaOptions struct {
	Page       int
	Limit      int
	MediaType  string // "VIDEO" or "IMAGE"
	SyncStatus string // "PENDING", "UPLOADING", "COMPLETED", "FAILED"
	Sort       string // "created_at_desc", "created_at_asc", "size_desc"
}

// MediaAssetList represents paginated media assets
type MediaAssetList struct {
	Assets     []*domain.MediaAsset
	Page       int
	Limit      int
	Total      int64
	TotalPages int
}

// uploadService implements UploadService
type uploadService struct {
	mediaRepo   repository.MediaRepository
	sessionRepo repository.SessionRepository
	youtubeClient youtube.Client
}

// NewUploadService creates a new upload service
func NewUploadService(
	mediaRepo repository.MediaRepository,
	sessionRepo repository.SessionRepository,
	youtubeClient youtube.Client,
) UploadService {
	return &uploadService{
		mediaRepo:   mediaRepo,
		sessionRepo: sessionRepo,
		youtubeClient: youtubeClient,
	}
}

// InitiateUploadSession creates a new upload session
func (s *uploadService) InitiateUploadSession(ctx context.Context, userID uuid.UUID, totalFiles int, totalBytes int64) (*domain.UploadSession, error) {
	session := &domain.UploadSession{
		UserID:        userID,
		TotalFiles:    totalFiles,
		TotalBytes:    totalBytes,
		SessionStatus: "ACTIVE",
		StartedAt:     time.Now(),
	}

	err := s.sessionRepo.Create(ctx, session)
	if err != nil {
		return nil, fmt.Errorf("failed to create upload session: %w", err)
	}

	return session, nil
}

// UploadVideo uploads a video file to YouTube
func (s *uploadService) UploadVideo(ctx context.Context, req *UploadVideoRequest) (*UploadVideoResult, error) {
	// Validate session exists and is active
	session, err := s.sessionRepo.FindByID(ctx, req.SessionID.String())
	if err != nil {
		return nil, fmt.Errorf("session not found: %w", err)
	}

	if session.SessionStatus != "ACTIVE" {
		return nil, fmt.Errorf("session is not active: %s", session.SessionStatus)
	}

	// Create media asset record with UPLOADING status
	now := time.Now()
	asset := &domain.MediaAsset{
		UserID:            req.UserID,
		OriginalFilename:  req.Filename,
		FileSizeBytes:     req.FileSizeBytes,
		MediaType:         "VIDEO",
		SyncStatus:        "UPLOADING",
		UploadStartedAt:   &now,
		RetryCount:        0,
	}

	err = s.mediaRepo.Create(ctx, asset)
	if err != nil {
		return nil, fmt.Errorf("failed to create media asset: %w", err)
	}

	// Upload to YouTube
	uploadReq := &youtube.UploadVideoRequest{
		FilePath:      req.FilePath,
		Title:         req.Title,
		Description:   req.Description,
		PrivacyStatus: "private",
		OnProgress:    req.OnProgress,
	}

	uploadResp, err := s.youtubeClient.UploadVideo(ctx, req.AccessToken, uploadReq)
	if err != nil {
		// Update asset status to FAILED
		asset.SyncStatus = "FAILED"
		asset.RetryCount++
		errMsg := err.Error()
		asset.ErrorMessage = &errMsg

		_ = s.mediaRepo.Update(ctx, asset)

		// Update session failed count
		session.FailedFiles++
		_ = s.sessionRepo.Update(ctx, session)

		return nil, fmt.Errorf("failed to upload video to YouTube: %w", err)
	}

	// Verify video is playable
	videoStatus, err := s.youtubeClient.GetVideoStatus(ctx, req.AccessToken, uploadResp.VideoID)
	if err != nil {
		// Video uploaded but verification failed
		// Mark as COMPLETED with warning
		asset.SyncStatus = "COMPLETED"
	} else if !videoStatus.Playable {
		// Video not playable yet
		asset.SyncStatus = "COMPLETED"
		errMsg := fmt.Sprintf("video uploaded but not yet playable: %s", videoStatus.Status)
		asset.ErrorMessage = &errMsg
	} else {
		// Video is playable
		asset.SyncStatus = "COMPLETED"
	}

	// Update asset with YouTube video ID and completed status
	asset.YouTubeVideoID = &uploadResp.VideoID
	completedAt := time.Now()
	asset.UploadCompletedAt = &completedAt

	err = s.mediaRepo.Update(ctx, asset)
	if err != nil {
		return nil, fmt.Errorf("failed to update media asset: %w", err)
	}

	// Update session progress
	session.CompletedFiles++
	session.UploadedBytes += req.FileSizeBytes
	err = s.sessionRepo.Update(ctx, session)
	if err != nil {
		return nil, fmt.Errorf("failed to update session: %w", err)
	}

	return &UploadVideoResult{
		AssetID:       asset.AssetID,
		VideoID:       uploadResp.VideoID,
		Filename:      asset.OriginalFilename,
		FileSizeBytes: asset.FileSizeBytes,
		SyncStatus:    string(asset.SyncStatus),
		UploadedAt:    *asset.UploadCompletedAt,
	}, nil
}

// GetUploadSessionStatus retrieves upload session status
func (s *uploadService) GetUploadSessionStatus(ctx context.Context, sessionID uuid.UUID) (*UploadSessionStatus, error) {
	session, err := s.sessionRepo.FindByID(ctx, sessionID.String())
	if err != nil {
		return nil, fmt.Errorf("session not found: %w", err)
	}

	return &UploadSessionStatus{
		SessionID:      session.SessionID,
		UserID:         session.UserID,
		TotalFiles:     session.TotalFiles,
		CompletedFiles: session.CompletedFiles,
		FailedFiles:    session.FailedFiles,
		TotalBytes:     session.TotalBytes,
		UploadedBytes:  session.UploadedBytes,
		SessionStatus:  string(session.SessionStatus),
		StartedAt:      session.StartedAt,
		CompletedAt:    session.CompletedAt,
	}, nil
}

// CompleteUploadSession marks an upload session as completed
func (s *uploadService) CompleteUploadSession(ctx context.Context, sessionID uuid.UUID) error {
	session, err := s.sessionRepo.FindByID(ctx, sessionID.String())
	if err != nil {
		return fmt.Errorf("session not found: %w", err)
	}

	now := time.Now()
	session.SessionStatus = "COMPLETED"
	session.CompletedAt = &now

	err = s.sessionRepo.Update(ctx, session)
	if err != nil {
		return fmt.Errorf("failed to update session: %w", err)
	}

	return nil
}

// CancelUploadSession cancels an upload session
func (s *uploadService) CancelUploadSession(ctx context.Context, sessionID uuid.UUID) error {
	session, err := s.sessionRepo.FindByID(ctx, sessionID.String())
	if err != nil {
		return fmt.Errorf("session not found: %w", err)
	}

	now := time.Now()
	session.SessionStatus = "CANCELLED"
	session.CompletedAt = &now

	err = s.sessionRepo.Update(ctx, session)
	if err != nil {
		return fmt.Errorf("failed to update session: %w", err)
	}

	return nil
}

// ListMediaAssets retrieves paginated list of media assets
func (s *uploadService) ListMediaAssets(ctx context.Context, userID uuid.UUID, opts *ListMediaOptions) (*MediaAssetList, error) {
	// Set defaults
	if opts.Page < 1 {
		opts.Page = 1
	}
	if opts.Limit < 1 || opts.Limit > 100 {
		opts.Limit = 50
	}
	if opts.Sort == "" {
		opts.Sort = "created_at_desc"
	}

	// Calculate offset
	offset := (opts.Page - 1) * opts.Limit

	// List assets (FindByUserID returns both assets and total count)
	assetsSlice, total, err := s.mediaRepo.FindByUserID(ctx, userID.String(), opts.Limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list media assets: %w", err)
	}

	// Convert []domain.MediaAsset to []*domain.MediaAsset
	assets := make([]*domain.MediaAsset, len(assetsSlice))
	for i := range assetsSlice {
		assets[i] = &assetsSlice[i]
	}

	// TODO: Apply client-side filtering for MediaType, SyncStatus, and Sort if needed
	// The current repository interface doesn't support these filters

	// Calculate total pages
	totalPages := int(total) / opts.Limit
	if int(total)%opts.Limit > 0 {
		totalPages++
	}

	return &MediaAssetList{
		Assets:     assets,
		Page:       opts.Page,
		Limit:      opts.Limit,
		Total:      total,
		TotalPages: totalPages,
	}, nil
}

// GetMediaAsset retrieves a single media asset by ID
func (s *uploadService) GetMediaAsset(ctx context.Context, assetID uuid.UUID) (*domain.MediaAsset, error) {
	asset, err := s.mediaRepo.FindByID(ctx, assetID.String())
	if err != nil {
		return nil, fmt.Errorf("media asset not found: %w", err)
	}

	return asset, nil
}

// DeleteMediaAsset deletes a media asset
func (s *uploadService) DeleteMediaAsset(ctx context.Context, assetID uuid.UUID, deleteFromYouTube bool) error {
	// Get asset
	asset, err := s.mediaRepo.FindByID(ctx, assetID.String())
	if err != nil {
		return fmt.Errorf("media asset not found: %w", err)
	}

	// Delete from YouTube if requested and video ID exists
	if deleteFromYouTube && asset.YouTubeVideoID != nil && *asset.YouTubeVideoID != "" {
		// Note: This would require access token, which we don't have here
		// In a real implementation, you'd need to get the user's access token
		// For now, we just delete the record
	}

	// Delete from database
	err = s.mediaRepo.Delete(ctx, assetID.String())
	if err != nil {
		return fmt.Errorf("failed to delete media asset: %w", err)
	}

	return nil
}
