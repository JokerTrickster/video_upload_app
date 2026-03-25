package service

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/JokerTrickster/video-upload-backend/internal/domain"
	"github.com/JokerTrickster/video-upload-backend/internal/pkg/youtube"
)

// MockMediaRepository implements repository.MediaRepository
type MockMediaRepository struct {
	mock.Mock
}

func (m *MockMediaRepository) Create(ctx context.Context, asset *domain.MediaAsset) error {
	args := m.Called(ctx, asset)
	if asset.AssetID == uuid.Nil {
		asset.AssetID = uuid.New()
	}
	return args.Error(0)
}

func (m *MockMediaRepository) FindByID(ctx context.Context, assetID string) (*domain.MediaAsset, error) {
	args := m.Called(ctx, assetID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.MediaAsset), args.Error(1)
}

func (m *MockMediaRepository) FindByUserID(ctx context.Context, userID string, limit, offset int) ([]domain.MediaAsset, int64, error) {
	args := m.Called(ctx, userID, limit, offset)
	if args.Get(0) == nil {
		return nil, 0, args.Error(2)
	}
	return args.Get(0).([]domain.MediaAsset), args.Get(1).(int64), args.Error(2)
}

func (m *MockMediaRepository) FindPendingUploads(ctx context.Context, userID string) ([]domain.MediaAsset, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.MediaAsset), args.Error(1)
}

func (m *MockMediaRepository) Update(ctx context.Context, asset *domain.MediaAsset) error {
	args := m.Called(ctx, asset)
	return args.Error(0)
}

func (m *MockMediaRepository) Delete(ctx context.Context, assetID string) error {
	args := m.Called(ctx, assetID)
	return args.Error(0)
}

func (m *MockMediaRepository) CountByUserID(ctx context.Context, userID string) (int64, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockMediaRepository) FindByYouTubeVideoID(ctx context.Context, videoID string) (*domain.MediaAsset, error) {
	args := m.Called(ctx, videoID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.MediaAsset), args.Error(1)
}

// MockSessionRepository implements repository.SessionRepository
type MockSessionRepository struct {
	mock.Mock
}

func (m *MockSessionRepository) Create(ctx context.Context, session *domain.UploadSession) error {
	args := m.Called(ctx, session)
	if session.SessionID == uuid.Nil {
		session.SessionID = uuid.New()
	}
	return args.Error(0)
}

func (m *MockSessionRepository) FindByID(ctx context.Context, sessionID string) (*domain.UploadSession, error) {
	args := m.Called(ctx, sessionID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.UploadSession), args.Error(1)
}

func (m *MockSessionRepository) FindByUserID(ctx context.Context, userID string, limit, offset int) ([]domain.UploadSession, int64, error) {
	args := m.Called(ctx, userID, limit, offset)
	if args.Get(0) == nil {
		return nil, 0, args.Error(2)
	}
	return args.Get(0).([]domain.UploadSession), args.Get(1).(int64), args.Error(2)
}

func (m *MockSessionRepository) FindActiveByUserID(ctx context.Context, userID string) ([]domain.UploadSession, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.UploadSession), args.Error(1)
}

func (m *MockSessionRepository) Update(ctx context.Context, session *domain.UploadSession) error {
	args := m.Called(ctx, session)
	return args.Error(0)
}

func (m *MockSessionRepository) Delete(ctx context.Context, sessionID string) error {
	args := m.Called(ctx, sessionID)
	return args.Error(0)
}

// MockYouTubeClient implements youtube.Client
type MockYouTubeClient struct {
	mock.Mock
}

func (m *MockYouTubeClient) UploadVideo(ctx context.Context, accessToken string, req *youtube.UploadVideoRequest) (*youtube.UploadVideoResponse, error) {
	args := m.Called(ctx, accessToken, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*youtube.UploadVideoResponse), args.Error(1)
}

func (m *MockYouTubeClient) GetVideoStatus(ctx context.Context, accessToken string, videoID string) (*youtube.VideoStatus, error) {
	args := m.Called(ctx, accessToken, videoID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*youtube.VideoStatus), args.Error(1)
}

func (m *MockYouTubeClient) DeleteVideo(ctx context.Context, accessToken string, videoID string) error {
	args := m.Called(ctx, accessToken, videoID)
	return args.Error(0)
}

// --- Tests ---

func TestInitiateUploadSession(t *testing.T) {
	ctx := context.Background()
	mockMediaRepo := new(MockMediaRepository)
	mockSessionRepo := new(MockSessionRepository)
	mockYouTube := new(MockYouTubeClient)

	svc := NewUploadService(mockMediaRepo, mockSessionRepo, mockYouTube)

	userID := uuid.New()

	mockSessionRepo.On("Create", ctx, mock.AnythingOfType("*domain.UploadSession")).Return(nil)

	session, err := svc.InitiateUploadSession(ctx, userID, 10, 1024*1024*1024)

	require.NoError(t, err)
	assert.NotNil(t, session)
	assert.Equal(t, userID, session.UserID)
	assert.Equal(t, 10, session.TotalFiles)
	assert.Equal(t, int64(1024*1024*1024), session.TotalBytes)
	assert.Equal(t, "ACTIVE", string(session.SessionStatus))

	mockSessionRepo.AssertExpectations(t)
}

func TestInitiateUploadSession_RepoError(t *testing.T) {
	ctx := context.Background()
	mockMediaRepo := new(MockMediaRepository)
	mockSessionRepo := new(MockSessionRepository)
	mockYouTube := new(MockYouTubeClient)

	svc := NewUploadService(mockMediaRepo, mockSessionRepo, mockYouTube)

	mockSessionRepo.On("Create", ctx, mock.AnythingOfType("*domain.UploadSession")).
		Return(fmt.Errorf("database error"))

	_, err := svc.InitiateUploadSession(ctx, uuid.New(), 5, 1000)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create upload session")
}

func TestUploadVideo_Success(t *testing.T) {
	ctx := context.Background()
	mockMediaRepo := new(MockMediaRepository)
	mockSessionRepo := new(MockSessionRepository)
	mockYouTube := new(MockYouTubeClient)

	svc := NewUploadService(mockMediaRepo, mockSessionRepo, mockYouTube)

	sessionID := uuid.New()
	userID := uuid.New()

	session := &domain.UploadSession{
		SessionID:     sessionID,
		UserID:        userID,
		SessionStatus: "ACTIVE",
	}

	videoID := "test-video-id"
	uploadResp := &youtube.UploadVideoResponse{
		VideoID:       videoID,
		Title:         "Test Video",
		ThumbnailURL:  "https://example.com/thumb.jpg",
		UploadedBytes: 1024,
	}

	videoStatus := &youtube.VideoStatus{
		VideoID:  videoID,
		Status:   "processed",
		Playable: true,
	}

	req := &UploadVideoRequest{
		SessionID:     sessionID,
		UserID:        userID,
		AccessToken:   "test-token",
		FilePath:      "/tmp/test.mp4",
		Filename:      "test.mp4",
		FileSizeBytes: 1024,
		Title:         "Test Video",
		Description:   "Test Description",
	}

	mockSessionRepo.On("FindByID", ctx, sessionID.String()).Return(session, nil)
	mockMediaRepo.On("Create", ctx, mock.AnythingOfType("*domain.MediaAsset")).Return(nil)
	mockYouTube.On("UploadVideo", ctx, "test-token", mock.AnythingOfType("*youtube.UploadVideoRequest")).Return(uploadResp, nil)
	mockYouTube.On("GetVideoStatus", ctx, "test-token", videoID).Return(videoStatus, nil)
	mockMediaRepo.On("Update", ctx, mock.AnythingOfType("*domain.MediaAsset")).Return(nil)
	mockSessionRepo.On("Update", ctx, mock.AnythingOfType("*domain.UploadSession")).Return(nil)

	result, err := svc.UploadVideo(ctx, req)

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, videoID, result.VideoID)
	assert.Equal(t, "test.mp4", result.Filename)
	assert.Equal(t, "COMPLETED", result.SyncStatus)

	mockSessionRepo.AssertExpectations(t)
	mockMediaRepo.AssertExpectations(t)
	mockYouTube.AssertExpectations(t)
}

func TestUploadVideo_SessionNotActive(t *testing.T) {
	ctx := context.Background()
	mockMediaRepo := new(MockMediaRepository)
	mockSessionRepo := new(MockSessionRepository)
	mockYouTube := new(MockYouTubeClient)

	svc := NewUploadService(mockMediaRepo, mockSessionRepo, mockYouTube)

	sessionID := uuid.New()
	session := &domain.UploadSession{
		SessionID:     sessionID,
		SessionStatus: "COMPLETED",
	}

	req := &UploadVideoRequest{
		SessionID:   sessionID,
		UserID:      uuid.New(),
		AccessToken: "test-token",
		FilePath:    "/tmp/test.mp4",
		Filename:    "test.mp4",
	}

	mockSessionRepo.On("FindByID", ctx, sessionID.String()).Return(session, nil)

	_, err := svc.UploadVideo(ctx, req)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "session is not active")
}

func TestUploadVideo_SessionNotFound(t *testing.T) {
	ctx := context.Background()
	mockMediaRepo := new(MockMediaRepository)
	mockSessionRepo := new(MockSessionRepository)
	mockYouTube := new(MockYouTubeClient)

	svc := NewUploadService(mockMediaRepo, mockSessionRepo, mockYouTube)

	sessionID := uuid.New()
	req := &UploadVideoRequest{
		SessionID: sessionID,
		UserID:    uuid.New(),
	}

	mockSessionRepo.On("FindByID", ctx, sessionID.String()).
		Return(nil, domain.ErrSessionNotFound)

	_, err := svc.UploadVideo(ctx, req)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "session not found")
}

func TestUploadVideo_YouTubeUploadFails(t *testing.T) {
	ctx := context.Background()
	mockMediaRepo := new(MockMediaRepository)
	mockSessionRepo := new(MockSessionRepository)
	mockYouTube := new(MockYouTubeClient)

	svc := NewUploadService(mockMediaRepo, mockSessionRepo, mockYouTube)

	sessionID := uuid.New()
	session := &domain.UploadSession{
		SessionID:     sessionID,
		SessionStatus: "ACTIVE",
	}

	req := &UploadVideoRequest{
		SessionID:     sessionID,
		UserID:        uuid.New(),
		AccessToken:   "test-token",
		FilePath:      "/tmp/test.mp4",
		Filename:      "test.mp4",
		FileSizeBytes: 1024,
	}

	mockSessionRepo.On("FindByID", ctx, sessionID.String()).Return(session, nil)
	mockMediaRepo.On("Create", ctx, mock.AnythingOfType("*domain.MediaAsset")).Return(nil)
	mockYouTube.On("UploadVideo", ctx, "test-token", mock.AnythingOfType("*youtube.UploadVideoRequest")).
		Return(nil, fmt.Errorf("YouTube API error"))
	mockMediaRepo.On("Update", ctx, mock.AnythingOfType("*domain.MediaAsset")).Return(nil)
	mockSessionRepo.On("Update", ctx, mock.AnythingOfType("*domain.UploadSession")).Return(nil)

	_, err := svc.UploadVideo(ctx, req)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to upload video to YouTube")
}

func TestCompleteUploadSession(t *testing.T) {
	ctx := context.Background()
	mockMediaRepo := new(MockMediaRepository)
	mockSessionRepo := new(MockSessionRepository)
	mockYouTube := new(MockYouTubeClient)

	svc := NewUploadService(mockMediaRepo, mockSessionRepo, mockYouTube)

	sessionID := uuid.New()
	session := &domain.UploadSession{
		SessionID:     sessionID,
		SessionStatus: "ACTIVE",
		StartedAt:     time.Now(),
	}

	mockSessionRepo.On("FindByID", ctx, sessionID.String()).Return(session, nil)
	mockSessionRepo.On("Update", ctx, mock.AnythingOfType("*domain.UploadSession")).Return(nil)

	err := svc.CompleteUploadSession(ctx, sessionID)

	require.NoError(t, err)
	mockSessionRepo.AssertExpectations(t)
}

func TestCompleteUploadSession_NotFound(t *testing.T) {
	ctx := context.Background()
	mockMediaRepo := new(MockMediaRepository)
	mockSessionRepo := new(MockSessionRepository)
	mockYouTube := new(MockYouTubeClient)

	svc := NewUploadService(mockMediaRepo, mockSessionRepo, mockYouTube)

	sessionID := uuid.New()
	mockSessionRepo.On("FindByID", ctx, sessionID.String()).
		Return(nil, domain.ErrSessionNotFound)

	err := svc.CompleteUploadSession(ctx, sessionID)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "session not found")
}

func TestCancelUploadSession(t *testing.T) {
	ctx := context.Background()
	mockMediaRepo := new(MockMediaRepository)
	mockSessionRepo := new(MockSessionRepository)
	mockYouTube := new(MockYouTubeClient)

	svc := NewUploadService(mockMediaRepo, mockSessionRepo, mockYouTube)

	sessionID := uuid.New()
	session := &domain.UploadSession{
		SessionID:     sessionID,
		SessionStatus: "ACTIVE",
	}

	mockSessionRepo.On("FindByID", ctx, sessionID.String()).Return(session, nil)
	mockSessionRepo.On("Update", ctx, mock.AnythingOfType("*domain.UploadSession")).Return(nil)

	err := svc.CancelUploadSession(ctx, sessionID)

	require.NoError(t, err)
	mockSessionRepo.AssertExpectations(t)
}

func TestGetUploadSessionStatus(t *testing.T) {
	ctx := context.Background()
	mockMediaRepo := new(MockMediaRepository)
	mockSessionRepo := new(MockSessionRepository)
	mockYouTube := new(MockYouTubeClient)

	svc := NewUploadService(mockMediaRepo, mockSessionRepo, mockYouTube)

	sessionID := uuid.New()
	userID := uuid.New()
	session := &domain.UploadSession{
		SessionID:      sessionID,
		UserID:         userID,
		TotalFiles:     10,
		CompletedFiles: 5,
		FailedFiles:    1,
		TotalBytes:     1024000,
		UploadedBytes:  512000,
		SessionStatus:  "ACTIVE",
		StartedAt:      time.Now(),
	}

	mockSessionRepo.On("FindByID", ctx, sessionID.String()).Return(session, nil)

	status, err := svc.GetUploadSessionStatus(ctx, sessionID)

	require.NoError(t, err)
	assert.Equal(t, sessionID, status.SessionID)
	assert.Equal(t, userID, status.UserID)
	assert.Equal(t, 10, status.TotalFiles)
	assert.Equal(t, 5, status.CompletedFiles)
	assert.Equal(t, 1, status.FailedFiles)
	assert.Equal(t, "ACTIVE", status.SessionStatus)
}

func TestGetUploadSessionStatus_NotFound(t *testing.T) {
	ctx := context.Background()
	mockMediaRepo := new(MockMediaRepository)
	mockSessionRepo := new(MockSessionRepository)
	mockYouTube := new(MockYouTubeClient)

	svc := NewUploadService(mockMediaRepo, mockSessionRepo, mockYouTube)

	sessionID := uuid.New()
	mockSessionRepo.On("FindByID", ctx, sessionID.String()).
		Return(nil, domain.ErrSessionNotFound)

	_, err := svc.GetUploadSessionStatus(ctx, sessionID)

	require.Error(t, err)
}

func TestListMediaAssets(t *testing.T) {
	ctx := context.Background()
	mockMediaRepo := new(MockMediaRepository)
	mockSessionRepo := new(MockSessionRepository)
	mockYouTube := new(MockYouTubeClient)

	svc := NewUploadService(mockMediaRepo, mockSessionRepo, mockYouTube)

	userID := uuid.New()
	assets := []domain.MediaAsset{
		{
			AssetID:          uuid.New(),
			UserID:           userID,
			OriginalFilename: "test1.mp4",
			MediaType:        "VIDEO",
			SyncStatus:       "COMPLETED",
		},
		{
			AssetID:          uuid.New(),
			UserID:           userID,
			OriginalFilename: "test2.mp4",
			MediaType:        "VIDEO",
			SyncStatus:       "COMPLETED",
		},
	}

	opts := &ListMediaOptions{
		Page:  1,
		Limit: 50,
	}

	mockMediaRepo.On("FindByUserID", ctx, userID.String(), 50, 0).Return(assets, int64(2), nil)

	result, err := svc.ListMediaAssets(ctx, userID, opts)

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result.Assets, 2)
	assert.Equal(t, 1, result.Page)
	assert.Equal(t, 50, result.Limit)
	assert.Equal(t, int64(2), result.Total)
	assert.Equal(t, 1, result.TotalPages)

	mockMediaRepo.AssertExpectations(t)
}

func TestListMediaAssets_DefaultValues(t *testing.T) {
	ctx := context.Background()
	mockMediaRepo := new(MockMediaRepository)
	mockSessionRepo := new(MockSessionRepository)
	mockYouTube := new(MockYouTubeClient)

	svc := NewUploadService(mockMediaRepo, mockSessionRepo, mockYouTube)

	userID := uuid.New()

	// Page 0 and Limit 0 should default to 1 and 50
	opts := &ListMediaOptions{
		Page:  0,
		Limit: 0,
	}

	mockMediaRepo.On("FindByUserID", ctx, userID.String(), 50, 0).
		Return([]domain.MediaAsset{}, int64(0), nil)

	result, err := svc.ListMediaAssets(ctx, userID, opts)

	require.NoError(t, err)
	assert.Equal(t, 1, result.Page)
	assert.Equal(t, 50, result.Limit)
}

func TestListMediaAssets_LimitCapped(t *testing.T) {
	ctx := context.Background()
	mockMediaRepo := new(MockMediaRepository)
	mockSessionRepo := new(MockSessionRepository)
	mockYouTube := new(MockYouTubeClient)

	svc := NewUploadService(mockMediaRepo, mockSessionRepo, mockYouTube)

	userID := uuid.New()

	// Limit > 100 should be capped to 50
	opts := &ListMediaOptions{
		Page:  1,
		Limit: 200,
	}

	mockMediaRepo.On("FindByUserID", ctx, userID.String(), 50, 0).
		Return([]domain.MediaAsset{}, int64(0), nil)

	result, err := svc.ListMediaAssets(ctx, userID, opts)

	require.NoError(t, err)
	assert.Equal(t, 50, result.Limit)
}

func TestGetMediaAsset(t *testing.T) {
	ctx := context.Background()
	mockMediaRepo := new(MockMediaRepository)
	mockSessionRepo := new(MockSessionRepository)
	mockYouTube := new(MockYouTubeClient)

	svc := NewUploadService(mockMediaRepo, mockSessionRepo, mockYouTube)

	assetID := uuid.New()
	videoID := "test-video-id"
	asset := &domain.MediaAsset{
		AssetID:        assetID,
		YouTubeVideoID: &videoID,
		MediaType:      "VIDEO",
		SyncStatus:     "COMPLETED",
	}

	mockMediaRepo.On("FindByID", ctx, assetID.String()).Return(asset, nil)

	result, err := svc.GetMediaAsset(ctx, assetID)

	require.NoError(t, err)
	assert.Equal(t, assetID, result.AssetID)
	assert.Equal(t, &videoID, result.YouTubeVideoID)
}

func TestGetMediaAsset_NotFound(t *testing.T) {
	ctx := context.Background()
	mockMediaRepo := new(MockMediaRepository)
	mockSessionRepo := new(MockSessionRepository)
	mockYouTube := new(MockYouTubeClient)

	svc := NewUploadService(mockMediaRepo, mockSessionRepo, mockYouTube)

	assetID := uuid.New()
	mockMediaRepo.On("FindByID", ctx, assetID.String()).
		Return(nil, domain.ErrMediaAssetNotFound)

	_, err := svc.GetMediaAsset(ctx, assetID)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "media asset not found")
}

func TestDeleteMediaAsset(t *testing.T) {
	ctx := context.Background()
	mockMediaRepo := new(MockMediaRepository)
	mockSessionRepo := new(MockSessionRepository)
	mockYouTube := new(MockYouTubeClient)

	svc := NewUploadService(mockMediaRepo, mockSessionRepo, mockYouTube)

	assetID := uuid.New()
	videoID := "test-video-id"
	asset := &domain.MediaAsset{
		AssetID:        assetID,
		YouTubeVideoID: &videoID,
		MediaType:      "VIDEO",
		SyncStatus:     "COMPLETED",
	}

	mockMediaRepo.On("FindByID", ctx, assetID.String()).Return(asset, nil)
	mockMediaRepo.On("Delete", ctx, assetID.String()).Return(nil)

	err := svc.DeleteMediaAsset(ctx, assetID, false)

	require.NoError(t, err)
	mockMediaRepo.AssertExpectations(t)
}

func TestDeleteMediaAsset_NotFound(t *testing.T) {
	ctx := context.Background()
	mockMediaRepo := new(MockMediaRepository)
	mockSessionRepo := new(MockSessionRepository)
	mockYouTube := new(MockYouTubeClient)

	svc := NewUploadService(mockMediaRepo, mockSessionRepo, mockYouTube)

	assetID := uuid.New()
	mockMediaRepo.On("FindByID", ctx, assetID.String()).
		Return(nil, domain.ErrMediaAssetNotFound)

	err := svc.DeleteMediaAsset(ctx, assetID, false)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "media asset not found")
}
