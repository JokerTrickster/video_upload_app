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

func (m *MockMediaRepository) FindByUserID(ctx context.Context, userID string, limit, offset int, mediaType, syncStatus, sort string) ([]domain.MediaAsset, int64, error) {
	args := m.Called(ctx, userID, limit, offset, mediaType, syncStatus, sort)
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

// MockTokenRepo implements repository.TokenRepository for tests
type MockTokenRepo struct {
	mock.Mock
}

func (m *MockTokenRepo) Create(ctx context.Context, token *domain.Token) error {
	args := m.Called(ctx, token)
	return args.Error(0)
}

func (m *MockTokenRepo) FindByUserID(ctx context.Context, userID string) (*domain.Token, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Token), args.Error(1)
}

func (m *MockTokenRepo) Update(ctx context.Context, token *domain.Token) error {
	args := m.Called(ctx, token)
	return args.Error(0)
}

func (m *MockTokenRepo) Delete(ctx context.Context, userID string) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

// MockTokenSvc implements TokenService for tests
type MockTokenSvc struct {
	mock.Mock
}

func (m *MockTokenSvc) EncryptToken(ctx context.Context, plainText string) (string, error) {
	args := m.Called(ctx, plainText)
	return args.String(0), args.Error(1)
}

func (m *MockTokenSvc) DecryptToken(ctx context.Context, cipherText string) (string, error) {
	args := m.Called(ctx, cipherText)
	return args.String(0), args.Error(1)
}

func (m *MockTokenSvc) AddToBlacklist(ctx context.Context, token string, expirySeconds int64) error {
	args := m.Called(ctx, token, expirySeconds)
	return args.Error(0)
}

func (m *MockTokenSvc) IsBlacklisted(ctx context.Context, token string) (bool, error) {
	args := m.Called(ctx, token)
	return args.Bool(0), args.Error(1)
}

func (m *MockTokenSvc) SaveOAuthState(ctx context.Context, state string, ttlSeconds int64) error {
	args := m.Called(ctx, state, ttlSeconds)
	return args.Error(0)
}

func (m *MockTokenSvc) ValidateOAuthState(ctx context.Context, state string) (bool, error) {
	args := m.Called(ctx, state)
	return args.Bool(0), args.Error(1)
}

// newTestUploadService creates an UploadService with all mocks
func newTestUploadService(mediaRepo *MockMediaRepository, sessionRepo *MockSessionRepository, ytClient *MockYouTubeClient) UploadService {
	return NewUploadService(mediaRepo, sessionRepo, new(MockTokenRepo), new(MockTokenSvc), ytClient)
}

// --- Tests ---

func TestInitiateUploadSession(t *testing.T) {
	ctx := context.Background()
	mockMediaRepo := new(MockMediaRepository)
	mockSessionRepo := new(MockSessionRepository)
	mockYouTube := new(MockYouTubeClient)

	svc := newTestUploadService(mockMediaRepo, mockSessionRepo, mockYouTube)

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

	svc := newTestUploadService(mockMediaRepo, mockSessionRepo, mockYouTube)

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

	svc := newTestUploadService(mockMediaRepo, mockSessionRepo, mockYouTube)

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

	svc := newTestUploadService(mockMediaRepo, mockSessionRepo, mockYouTube)

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

	svc := newTestUploadService(mockMediaRepo, mockSessionRepo, mockYouTube)

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

	svc := newTestUploadService(mockMediaRepo, mockSessionRepo, mockYouTube)

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

	svc := newTestUploadService(mockMediaRepo, mockSessionRepo, mockYouTube)

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

	svc := newTestUploadService(mockMediaRepo, mockSessionRepo, mockYouTube)

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

	svc := newTestUploadService(mockMediaRepo, mockSessionRepo, mockYouTube)

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

	svc := newTestUploadService(mockMediaRepo, mockSessionRepo, mockYouTube)

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

	svc := newTestUploadService(mockMediaRepo, mockSessionRepo, mockYouTube)

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

	svc := newTestUploadService(mockMediaRepo, mockSessionRepo, mockYouTube)

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

	mockMediaRepo.On("FindByUserID", ctx, userID.String(), 50, 0, "", "", "created_at_desc").Return(assets, int64(2), nil)

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

	svc := newTestUploadService(mockMediaRepo, mockSessionRepo, mockYouTube)

	userID := uuid.New()

	// Page 0 and Limit 0 should default to 1 and 50
	opts := &ListMediaOptions{
		Page:  0,
		Limit: 0,
	}

	mockMediaRepo.On("FindByUserID", ctx, userID.String(), 50, 0, "", "", "created_at_desc").
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

	svc := newTestUploadService(mockMediaRepo, mockSessionRepo, mockYouTube)

	userID := uuid.New()

	// Limit > 100 should be capped to 50
	opts := &ListMediaOptions{
		Page:  1,
		Limit: 200,
	}

	mockMediaRepo.On("FindByUserID", ctx, userID.String(), 50, 0, "", "", "created_at_desc").
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

	svc := newTestUploadService(mockMediaRepo, mockSessionRepo, mockYouTube)

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

	svc := newTestUploadService(mockMediaRepo, mockSessionRepo, mockYouTube)

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

	svc := newTestUploadService(mockMediaRepo, mockSessionRepo, mockYouTube)

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

	svc := newTestUploadService(mockMediaRepo, mockSessionRepo, mockYouTube)

	assetID := uuid.New()
	mockMediaRepo.On("FindByID", ctx, assetID.String()).
		Return(nil, domain.ErrMediaAssetNotFound)

	err := svc.DeleteMediaAsset(ctx, assetID, false)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "media asset not found")
}

// --- Delete with YouTube ---

func TestDeleteMediaAsset_WithYouTubeDelete(t *testing.T) {
	ctx := context.Background()
	mockMediaRepo := new(MockMediaRepository)
	mockSessionRepo := new(MockSessionRepository)
	mockYouTube := new(MockYouTubeClient)
	mockTokenRepo := new(MockTokenRepo)
	mockTokenSvc := new(MockTokenSvc)

	svc := NewUploadService(mockMediaRepo, mockSessionRepo, mockTokenRepo, mockTokenSvc, mockYouTube)

	assetID := uuid.New()
	userID := uuid.New()
	videoID := "yt-video-123"
	asset := &domain.MediaAsset{
		AssetID:        assetID,
		UserID:         userID,
		YouTubeVideoID: &videoID,
		SyncStatus:     "COMPLETED",
	}

	token := &domain.Token{
		UserID:               userID,
		EncryptedAccessToken: "encrypted-token",
		ExpiresAt:            time.Now().Add(1 * time.Hour),
	}

	mockMediaRepo.On("FindByID", ctx, assetID.String()).Return(asset, nil)
	mockTokenRepo.On("FindByUserID", ctx, userID.String()).Return(token, nil)
	mockTokenSvc.On("DecryptToken", ctx, "encrypted-token").Return("decrypted-token", nil)
	mockYouTube.On("DeleteVideo", ctx, "decrypted-token", videoID).Return(nil)
	mockMediaRepo.On("Delete", ctx, assetID.String()).Return(nil)

	err := svc.DeleteMediaAsset(ctx, assetID, true)

	require.NoError(t, err)
	mockYouTube.AssertCalled(t, "DeleteVideo", ctx, "decrypted-token", videoID)
	mockMediaRepo.AssertExpectations(t)
}

func TestDeleteMediaAsset_WithYouTubeDelete_NoVideoID(t *testing.T) {
	ctx := context.Background()
	mockMediaRepo := new(MockMediaRepository)
	mockSessionRepo := new(MockSessionRepository)
	mockYouTube := new(MockYouTubeClient)

	svc := newTestUploadService(mockMediaRepo, mockSessionRepo, mockYouTube)

	assetID := uuid.New()
	asset := &domain.MediaAsset{
		AssetID:        assetID,
		YouTubeVideoID: nil, // No video ID
		SyncStatus:     "FAILED",
	}

	mockMediaRepo.On("FindByID", ctx, assetID.String()).Return(asset, nil)
	mockMediaRepo.On("Delete", ctx, assetID.String()).Return(nil)

	err := svc.DeleteMediaAsset(ctx, assetID, true)

	require.NoError(t, err)
	// YouTube delete should not be called
	mockYouTube.AssertNotCalled(t, "DeleteVideo")
}

func TestDeleteMediaAsset_WithYouTubeDelete_EmptyVideoID(t *testing.T) {
	ctx := context.Background()
	mockMediaRepo := new(MockMediaRepository)
	mockSessionRepo := new(MockSessionRepository)
	mockYouTube := new(MockYouTubeClient)

	svc := newTestUploadService(mockMediaRepo, mockSessionRepo, mockYouTube)

	assetID := uuid.New()
	emptyID := ""
	asset := &domain.MediaAsset{
		AssetID:        assetID,
		YouTubeVideoID: &emptyID,
		SyncStatus:     "COMPLETED",
	}

	mockMediaRepo.On("FindByID", ctx, assetID.String()).Return(asset, nil)
	mockMediaRepo.On("Delete", ctx, assetID.String()).Return(nil)

	err := svc.DeleteMediaAsset(ctx, assetID, true)

	require.NoError(t, err)
	mockYouTube.AssertNotCalled(t, "DeleteVideo")
}

func TestDeleteMediaAsset_WithYouTubeDelete_ExpiredToken(t *testing.T) {
	ctx := context.Background()
	mockMediaRepo := new(MockMediaRepository)
	mockSessionRepo := new(MockSessionRepository)
	mockYouTube := new(MockYouTubeClient)
	mockTokenRepo := new(MockTokenRepo)
	mockTokenSvc := new(MockTokenSvc)

	svc := NewUploadService(mockMediaRepo, mockSessionRepo, mockTokenRepo, mockTokenSvc, mockYouTube)

	assetID := uuid.New()
	userID := uuid.New()
	videoID := "yt-video-123"
	asset := &domain.MediaAsset{
		AssetID:        assetID,
		UserID:         userID,
		YouTubeVideoID: &videoID,
		SyncStatus:     "COMPLETED",
	}

	expiredToken := &domain.Token{
		UserID:    userID,
		ExpiresAt: time.Now().Add(-1 * time.Hour), // Expired
	}

	mockMediaRepo.On("FindByID", ctx, assetID.String()).Return(asset, nil)
	mockTokenRepo.On("FindByUserID", ctx, userID.String()).Return(expiredToken, nil)
	mockMediaRepo.On("Delete", ctx, assetID.String()).Return(nil)

	err := svc.DeleteMediaAsset(ctx, assetID, true)

	require.NoError(t, err)
	// YouTube delete not called because token expired
	mockYouTube.AssertNotCalled(t, "DeleteVideo")
}

// --- Video Status Verification ---

func TestUploadVideo_VerificationFails(t *testing.T) {
	ctx := context.Background()
	mockMediaRepo := new(MockMediaRepository)
	mockSessionRepo := new(MockSessionRepository)
	mockYouTube := new(MockYouTubeClient)

	svc := newTestUploadService(mockMediaRepo, mockSessionRepo, mockYouTube)

	sessionID := uuid.New()
	session := &domain.UploadSession{SessionID: sessionID, SessionStatus: "ACTIVE"}

	uploadResp := &youtube.UploadVideoResponse{
		VideoID: "test-vid", Title: "Test", UploadedBytes: 1024,
	}

	req := &UploadVideoRequest{
		SessionID: sessionID, UserID: uuid.New(), AccessToken: "token",
		FilePath: "/tmp/test.mp4", Filename: "test.mp4", FileSizeBytes: 1024,
	}

	mockSessionRepo.On("FindByID", ctx, sessionID.String()).Return(session, nil)
	mockMediaRepo.On("Create", ctx, mock.AnythingOfType("*domain.MediaAsset")).Return(nil)
	mockYouTube.On("UploadVideo", ctx, "token", mock.AnythingOfType("*youtube.UploadVideoRequest")).
		Return(uploadResp, nil)
	// Verification fails
	mockYouTube.On("GetVideoStatus", ctx, "token", "test-vid").
		Return(nil, fmt.Errorf("verification timeout"))
	mockMediaRepo.On("Update", ctx, mock.AnythingOfType("*domain.MediaAsset")).Return(nil)
	mockSessionRepo.On("Update", ctx, mock.AnythingOfType("*domain.UploadSession")).Return(nil)

	result, err := svc.UploadVideo(ctx, req)

	require.NoError(t, err)
	assert.Equal(t, "COMPLETED", result.SyncStatus)
}

func TestUploadVideo_VideoNotPlayable(t *testing.T) {
	ctx := context.Background()
	mockMediaRepo := new(MockMediaRepository)
	mockSessionRepo := new(MockSessionRepository)
	mockYouTube := new(MockYouTubeClient)

	svc := newTestUploadService(mockMediaRepo, mockSessionRepo, mockYouTube)

	sessionID := uuid.New()
	session := &domain.UploadSession{SessionID: sessionID, SessionStatus: "ACTIVE"}

	uploadResp := &youtube.UploadVideoResponse{
		VideoID: "test-vid", Title: "Test", UploadedBytes: 1024,
	}

	videoStatus := &youtube.VideoStatus{
		VideoID:  "test-vid",
		Status:   "processing",
		Playable: false,
	}

	req := &UploadVideoRequest{
		SessionID: sessionID, UserID: uuid.New(), AccessToken: "token",
		FilePath: "/tmp/test.mp4", Filename: "test.mp4", FileSizeBytes: 1024,
	}

	mockSessionRepo.On("FindByID", ctx, sessionID.String()).Return(session, nil)
	mockMediaRepo.On("Create", ctx, mock.AnythingOfType("*domain.MediaAsset")).Return(nil)
	mockYouTube.On("UploadVideo", ctx, "token", mock.AnythingOfType("*youtube.UploadVideoRequest")).
		Return(uploadResp, nil)
	mockYouTube.On("GetVideoStatus", ctx, "token", "test-vid").
		Return(videoStatus, nil)
	mockMediaRepo.On("Update", ctx, mock.AnythingOfType("*domain.MediaAsset")).Return(nil)
	mockSessionRepo.On("Update", ctx, mock.AnythingOfType("*domain.UploadSession")).Return(nil)

	result, err := svc.UploadVideo(ctx, req)

	require.NoError(t, err)
	assert.Equal(t, "COMPLETED", result.SyncStatus) // Still COMPLETED even if not yet playable
}

// --- Retry Logic Tests ---

func TestUploadVideo_NonRetryableError(t *testing.T) {
	ctx := context.Background()
	mockMediaRepo := new(MockMediaRepository)
	mockSessionRepo := new(MockSessionRepository)
	mockYouTube := new(MockYouTubeClient)

	svc := newTestUploadService(mockMediaRepo, mockSessionRepo, mockYouTube)

	sessionID := uuid.New()
	session := &domain.UploadSession{SessionID: sessionID, SessionStatus: "ACTIVE"}

	req := &UploadVideoRequest{
		SessionID: sessionID, UserID: uuid.New(), AccessToken: "token",
		FilePath: "/tmp/test.mp4", Filename: "test.mp4", FileSizeBytes: 1024,
	}

	mockSessionRepo.On("FindByID", ctx, sessionID.String()).Return(session, nil)
	mockMediaRepo.On("Create", ctx, mock.AnythingOfType("*domain.MediaAsset")).Return(nil)
	// Non-retryable error (e.g., file too large)
	mockYouTube.On("UploadVideo", ctx, "token", mock.AnythingOfType("*youtube.UploadVideoRequest")).
		Return(nil, domain.ErrFileTooLarge)
	mockMediaRepo.On("Update", ctx, mock.AnythingOfType("*domain.MediaAsset")).Return(nil)
	mockSessionRepo.On("Update", ctx, mock.AnythingOfType("*domain.UploadSession")).Return(nil)

	_, err := svc.UploadVideo(ctx, req)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to upload video to YouTube")
	// Should only be called once (no retry for non-retryable errors)
	mockYouTube.AssertNumberOfCalls(t, "UploadVideo", 1)
}

func TestUploadVideo_MediaAssetCreateFails(t *testing.T) {
	ctx := context.Background()
	mockMediaRepo := new(MockMediaRepository)
	mockSessionRepo := new(MockSessionRepository)
	mockYouTube := new(MockYouTubeClient)

	svc := newTestUploadService(mockMediaRepo, mockSessionRepo, mockYouTube)

	sessionID := uuid.New()
	session := &domain.UploadSession{SessionID: sessionID, SessionStatus: "ACTIVE"}

	req := &UploadVideoRequest{
		SessionID: sessionID, UserID: uuid.New(), AccessToken: "token",
		FilePath: "/tmp/test.mp4", Filename: "test.mp4", FileSizeBytes: 1024,
	}

	mockSessionRepo.On("FindByID", ctx, sessionID.String()).Return(session, nil)
	mockMediaRepo.On("Create", ctx, mock.AnythingOfType("*domain.MediaAsset")).
		Return(fmt.Errorf("db error"))

	_, err := svc.UploadVideo(ctx, req)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create media asset")
}

func TestUploadVideo_SessionUpdateProgress(t *testing.T) {
	ctx := context.Background()
	mockMediaRepo := new(MockMediaRepository)
	mockSessionRepo := new(MockSessionRepository)
	mockYouTube := new(MockYouTubeClient)

	svc := newTestUploadService(mockMediaRepo, mockSessionRepo, mockYouTube)

	sessionID := uuid.New()
	userID := uuid.New()
	session := &domain.UploadSession{
		SessionID:      sessionID,
		UserID:         userID,
		TotalFiles:     2,
		CompletedFiles: 0,
		FailedFiles:    0,
		TotalBytes:     2048,
		UploadedBytes:  0,
		SessionStatus:  "ACTIVE",
	}

	uploadResp := &youtube.UploadVideoResponse{
		VideoID: "vid-1", Title: "Test", UploadedBytes: 1024,
	}

	videoStatus := &youtube.VideoStatus{VideoID: "vid-1", Playable: true, Status: "processed"}

	req := &UploadVideoRequest{
		SessionID: sessionID, UserID: userID, AccessToken: "token",
		FilePath: "/tmp/test.mp4", Filename: "test.mp4", FileSizeBytes: 1024,
	}

	mockSessionRepo.On("FindByID", ctx, sessionID.String()).Return(session, nil)
	mockMediaRepo.On("Create", ctx, mock.AnythingOfType("*domain.MediaAsset")).Return(nil)
	mockYouTube.On("UploadVideo", ctx, "token", mock.AnythingOfType("*youtube.UploadVideoRequest")).
		Return(uploadResp, nil)
	mockYouTube.On("GetVideoStatus", ctx, "token", "vid-1").Return(videoStatus, nil)
	mockMediaRepo.On("Update", ctx, mock.AnythingOfType("*domain.MediaAsset")).Return(nil)
	mockSessionRepo.On("Update", ctx, mock.AnythingOfType("*domain.UploadSession")).Return(nil)

	result, err := svc.UploadVideo(ctx, req)

	require.NoError(t, err)
	assert.Equal(t, "vid-1", result.VideoID)
	// Session should be updated: CompletedFiles = 1, UploadedBytes = 1024
	assert.Equal(t, 1, session.CompletedFiles)
	assert.Equal(t, int64(1024), session.UploadedBytes)
}

func TestListMediaAssets_WithFilters(t *testing.T) {
	ctx := context.Background()
	mockMediaRepo := new(MockMediaRepository)
	mockSessionRepo := new(MockSessionRepository)
	mockYouTube := new(MockYouTubeClient)

	svc := newTestUploadService(mockMediaRepo, mockSessionRepo, mockYouTube)

	userID := uuid.New()
	opts := &ListMediaOptions{
		Page:       1,
		Limit:      20,
		MediaType:  "VIDEO",
		SyncStatus: "COMPLETED",
		Sort:       "size_desc",
	}

	mockMediaRepo.On("FindByUserID", ctx, userID.String(), 20, 0, "VIDEO", "COMPLETED", "size_desc").
		Return([]domain.MediaAsset{}, int64(0), nil)

	result, err := svc.ListMediaAssets(ctx, userID, opts)

	require.NoError(t, err)
	assert.Empty(t, result.Assets)
	assert.Equal(t, 0, result.TotalPages)
}

func TestListMediaAssets_TotalPagesCalculation(t *testing.T) {
	ctx := context.Background()
	mockMediaRepo := new(MockMediaRepository)
	mockSessionRepo := new(MockSessionRepository)
	mockYouTube := new(MockYouTubeClient)

	svc := newTestUploadService(mockMediaRepo, mockSessionRepo, mockYouTube)

	userID := uuid.New()
	opts := &ListMediaOptions{Page: 1, Limit: 10}

	// 25 total items / 10 per page = 3 pages (2 full + 1 partial)
	mockMediaRepo.On("FindByUserID", ctx, userID.String(), 10, 0, "", "", "created_at_desc").
		Return([]domain.MediaAsset{}, int64(25), nil)

	result, err := svc.ListMediaAssets(ctx, userID, opts)

	require.NoError(t, err)
	assert.Equal(t, 3, result.TotalPages)
	assert.Equal(t, int64(25), result.Total)
}

func TestCancelUploadSession_SessionAlreadyCancelled(t *testing.T) {
	ctx := context.Background()
	mockMediaRepo := new(MockMediaRepository)
	mockSessionRepo := new(MockSessionRepository)
	mockYouTube := new(MockYouTubeClient)

	svc := newTestUploadService(mockMediaRepo, mockSessionRepo, mockYouTube)

	sessionID := uuid.New()
	now := time.Now()
	session := &domain.UploadSession{
		SessionID:     sessionID,
		SessionStatus: "CANCELLED",
		CompletedAt:   &now,
	}

	mockSessionRepo.On("FindByID", ctx, sessionID.String()).Return(session, nil)
	mockSessionRepo.On("Update", ctx, mock.AnythingOfType("*domain.UploadSession")).Return(nil)

	// Cancelling an already cancelled session still works (idempotent)
	err := svc.CancelUploadSession(ctx, sessionID)
	require.NoError(t, err)
}
