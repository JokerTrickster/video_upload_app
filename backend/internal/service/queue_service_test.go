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

// MockQueueRepository implements repository.QueueRepository
type MockQueueRepository struct {
	mock.Mock
}

func (m *MockQueueRepository) Create(ctx context.Context, item *domain.UploadQueueItem) error {
	args := m.Called(ctx, item)
	if item.QueueID == uuid.Nil {
		item.QueueID = uuid.New()
	}
	return args.Error(0)
}

func (m *MockQueueRepository) FindPendingByUserID(ctx context.Context, userID string) ([]domain.UploadQueueItem, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.UploadQueueItem), args.Error(1)
}

func (m *MockQueueRepository) FindPendingAll(ctx context.Context, limit int) ([]domain.UploadQueueItem, error) {
	args := m.Called(ctx, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.UploadQueueItem), args.Error(1)
}

func (m *MockQueueRepository) Update(ctx context.Context, item *domain.UploadQueueItem) error {
	args := m.Called(ctx, item)
	return args.Error(0)
}

func (m *MockQueueRepository) FindByID(ctx context.Context, queueID string) (*domain.UploadQueueItem, error) {
	args := m.Called(ctx, queueID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.UploadQueueItem), args.Error(1)
}

func (m *MockQueueRepository) FindByUserID(ctx context.Context, userID string, limit, offset int) ([]domain.UploadQueueItem, int64, error) {
	args := m.Called(ctx, userID, limit, offset)
	if args.Get(0) == nil {
		return nil, 0, args.Error(2)
	}
	return args.Get(0).([]domain.UploadQueueItem), args.Get(1).(int64), args.Error(2)
}

func (m *MockQueueRepository) DeleteByID(ctx context.Context, queueID string) error {
	args := m.Called(ctx, queueID)
	return args.Error(0)
}

func (m *MockQueueRepository) GetOrCreateDailyQuota(ctx context.Context, date string) (*domain.DailyQuota, error) {
	args := m.Called(ctx, date)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.DailyQuota), args.Error(1)
}

func (m *MockQueueRepository) UpdateDailyQuota(ctx context.Context, quota *domain.DailyQuota) error {
	args := m.Called(ctx, quota)
	return args.Error(0)
}

func newTestQueueService(
	queueRepo *MockQueueRepository,
	mediaRepo *MockMediaRepository,
	tokenRepo *MockTokenRepo,
	tokenSvc *MockTokenSvc,
	ytClient *MockYouTubeClient,
) QueueService {
	return NewQueueService(queueRepo, mediaRepo, tokenRepo, tokenSvc, ytClient)
}

// --- AddToQueue Tests ---

func TestAddToQueue_Success(t *testing.T) {
	ctx := context.Background()
	queueRepo := new(MockQueueRepository)
	svc := newTestQueueService(queueRepo, new(MockMediaRepository), new(MockTokenRepo), new(MockTokenSvc), new(MockYouTubeClient))

	userID := uuid.New()
	queueRepo.On("Create", ctx, mock.AnythingOfType("*domain.UploadQueueItem")).Return(nil)

	item, err := svc.AddToQueue(ctx, userID, "/path/to/video.mp4", "video.mp4", 1024*1024, "My Video", "Description")

	require.NoError(t, err)
	assert.NotNil(t, item)
	assert.Equal(t, userID, item.UserID)
	assert.Equal(t, "/path/to/video.mp4", item.FilePath)
	assert.Equal(t, "video.mp4", item.Filename)
	assert.Equal(t, int64(1024*1024), item.FileSizeBytes)
	assert.Equal(t, "My Video", item.Title)
	assert.Equal(t, "Description", item.Description)
	assert.Equal(t, domain.QueueStatusPending, item.QueueStatus)
	queueRepo.AssertExpectations(t)
}

func TestAddToQueue_RepoError(t *testing.T) {
	ctx := context.Background()
	queueRepo := new(MockQueueRepository)
	svc := newTestQueueService(queueRepo, new(MockMediaRepository), new(MockTokenRepo), new(MockTokenSvc), new(MockYouTubeClient))

	queueRepo.On("Create", ctx, mock.AnythingOfType("*domain.UploadQueueItem")).
		Return(fmt.Errorf("database error"))

	_, err := svc.AddToQueue(ctx, uuid.New(), "/path", "file.mp4", 100, "", "")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to add to queue")
}

// --- GetQueueItems Tests ---

func TestGetQueueItems_Success(t *testing.T) {
	ctx := context.Background()
	queueRepo := new(MockQueueRepository)
	svc := newTestQueueService(queueRepo, new(MockMediaRepository), new(MockTokenRepo), new(MockTokenSvc), new(MockYouTubeClient))

	userID := uuid.New()
	items := []domain.UploadQueueItem{
		{QueueID: uuid.New(), UserID: userID, Filename: "a.mp4", QueueStatus: domain.QueueStatusPending},
		{QueueID: uuid.New(), UserID: userID, Filename: "b.mp4", QueueStatus: domain.QueueStatusCompleted},
	}

	queueRepo.On("FindByUserID", ctx, userID.String(), 50, 0).Return(items, int64(2), nil)

	result, total, err := svc.GetQueueItems(ctx, userID, 1, 50)

	require.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Equal(t, int64(2), total)
	queueRepo.AssertExpectations(t)
}

func TestGetQueueItems_DefaultValues(t *testing.T) {
	ctx := context.Background()
	queueRepo := new(MockQueueRepository)
	svc := newTestQueueService(queueRepo, new(MockMediaRepository), new(MockTokenRepo), new(MockTokenSvc), new(MockYouTubeClient))

	userID := uuid.New()

	// Page 0 → 1, Limit 0 → 50
	queueRepo.On("FindByUserID", ctx, userID.String(), 50, 0).
		Return([]domain.UploadQueueItem{}, int64(0), nil)

	_, _, err := svc.GetQueueItems(ctx, userID, 0, 0)
	require.NoError(t, err)
	queueRepo.AssertExpectations(t)
}

func TestGetQueueItems_LimitCapped(t *testing.T) {
	ctx := context.Background()
	queueRepo := new(MockQueueRepository)
	svc := newTestQueueService(queueRepo, new(MockMediaRepository), new(MockTokenRepo), new(MockTokenSvc), new(MockYouTubeClient))

	userID := uuid.New()

	// Limit > 100 → capped to 50
	queueRepo.On("FindByUserID", ctx, userID.String(), 50, 0).
		Return([]domain.UploadQueueItem{}, int64(0), nil)

	_, _, err := svc.GetQueueItems(ctx, userID, 1, 200)
	require.NoError(t, err)
	queueRepo.AssertExpectations(t)
}

func TestGetQueueItems_Pagination(t *testing.T) {
	ctx := context.Background()
	queueRepo := new(MockQueueRepository)
	svc := newTestQueueService(queueRepo, new(MockMediaRepository), new(MockTokenRepo), new(MockTokenSvc), new(MockYouTubeClient))

	userID := uuid.New()

	// Page 3, Limit 10 → offset = 20
	queueRepo.On("FindByUserID", ctx, userID.String(), 10, 20).
		Return([]domain.UploadQueueItem{}, int64(25), nil)

	_, total, err := svc.GetQueueItems(ctx, userID, 3, 10)
	require.NoError(t, err)
	assert.Equal(t, int64(25), total)
}

// --- RemoveFromQueue Tests ---

func TestRemoveFromQueue_Success(t *testing.T) {
	ctx := context.Background()
	queueRepo := new(MockQueueRepository)
	svc := newTestQueueService(queueRepo, new(MockMediaRepository), new(MockTokenRepo), new(MockTokenSvc), new(MockYouTubeClient))

	queueID := uuid.New()
	item := &domain.UploadQueueItem{
		QueueID:     queueID,
		QueueStatus: domain.QueueStatusPending,
	}

	queueRepo.On("FindByID", ctx, queueID.String()).Return(item, nil)
	queueRepo.On("DeleteByID", ctx, queueID.String()).Return(nil)

	err := svc.RemoveFromQueue(ctx, queueID)
	require.NoError(t, err)
	queueRepo.AssertExpectations(t)
}

func TestRemoveFromQueue_NotPending(t *testing.T) {
	ctx := context.Background()
	queueRepo := new(MockQueueRepository)
	svc := newTestQueueService(queueRepo, new(MockMediaRepository), new(MockTokenRepo), new(MockTokenSvc), new(MockYouTubeClient))

	queueID := uuid.New()
	item := &domain.UploadQueueItem{
		QueueID:     queueID,
		QueueStatus: domain.QueueStatusProcessing,
	}

	queueRepo.On("FindByID", ctx, queueID.String()).Return(item, nil)

	err := svc.RemoveFromQueue(ctx, queueID)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "can only remove PENDING items")
}

func TestRemoveFromQueue_NotFound(t *testing.T) {
	ctx := context.Background()
	queueRepo := new(MockQueueRepository)
	svc := newTestQueueService(queueRepo, new(MockMediaRepository), new(MockTokenRepo), new(MockTokenSvc), new(MockYouTubeClient))

	queueID := uuid.New()
	queueRepo.On("FindByID", ctx, queueID.String()).
		Return(nil, domain.ErrMediaAssetNotFound)

	err := svc.RemoveFromQueue(ctx, queueID)
	require.Error(t, err)
}

func TestRemoveFromQueue_CompletedItem(t *testing.T) {
	ctx := context.Background()
	queueRepo := new(MockQueueRepository)
	svc := newTestQueueService(queueRepo, new(MockMediaRepository), new(MockTokenRepo), new(MockTokenSvc), new(MockYouTubeClient))

	queueID := uuid.New()
	item := &domain.UploadQueueItem{
		QueueID:     queueID,
		QueueStatus: domain.QueueStatusCompleted,
	}

	queueRepo.On("FindByID", ctx, queueID.String()).Return(item, nil)

	err := svc.RemoveFromQueue(ctx, queueID)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "can only remove PENDING items")
	assert.Contains(t, err.Error(), "COMPLETED")
}

func TestRemoveFromQueue_FailedItem(t *testing.T) {
	ctx := context.Background()
	queueRepo := new(MockQueueRepository)
	svc := newTestQueueService(queueRepo, new(MockMediaRepository), new(MockTokenRepo), new(MockTokenSvc), new(MockYouTubeClient))

	queueID := uuid.New()
	item := &domain.UploadQueueItem{
		QueueID:     queueID,
		QueueStatus: domain.QueueStatusFailed,
	}

	queueRepo.On("FindByID", ctx, queueID.String()).Return(item, nil)

	err := svc.RemoveFromQueue(ctx, queueID)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "can only remove PENDING items")
}

// --- GetQuotaStatus Tests ---

func TestGetQuotaStatus_Success(t *testing.T) {
	ctx := context.Background()
	queueRepo := new(MockQueueRepository)
	svc := newTestQueueService(queueRepo, new(MockMediaRepository), new(MockTokenRepo), new(MockTokenSvc), new(MockYouTubeClient))

	today := time.Now().Format("2006-01-02")
	quota := &domain.DailyQuota{
		ID:        uuid.New(),
		Date:      today,
		UnitsUsed: 3200,
		UnitsMax:  10000,
		Uploads:   2,
	}

	queueRepo.On("GetOrCreateDailyQuota", ctx, today).Return(quota, nil)

	result, err := svc.GetQuotaStatus(ctx)

	require.NoError(t, err)
	assert.Equal(t, today, result.Date)
	assert.Equal(t, 3200, result.UnitsUsed)
	assert.Equal(t, 2, result.Uploads)
	assert.True(t, result.CanUpload())
	assert.Equal(t, 4, result.RemainingUploads()) // (10000-3200)/1600 = 4
}

func TestGetQuotaStatus_Exhausted(t *testing.T) {
	ctx := context.Background()
	queueRepo := new(MockQueueRepository)
	svc := newTestQueueService(queueRepo, new(MockMediaRepository), new(MockTokenRepo), new(MockTokenSvc), new(MockYouTubeClient))

	today := time.Now().Format("2006-01-02")
	quota := &domain.DailyQuota{
		ID:        uuid.New(),
		Date:      today,
		UnitsUsed: 9600, // 6 uploads * 1600
		UnitsMax:  10000,
		Uploads:   6,
	}

	queueRepo.On("GetOrCreateDailyQuota", ctx, today).Return(quota, nil)

	result, err := svc.GetQuotaStatus(ctx)

	require.NoError(t, err)
	assert.False(t, result.CanUpload())
	assert.Equal(t, 0, result.RemainingUploads())
}

// --- ProcessQueue Tests ---

func TestProcessQueue_Success(t *testing.T) {
	ctx := context.Background()
	queueRepo := new(MockQueueRepository)
	mediaRepo := new(MockMediaRepository)
	tokenRepo := new(MockTokenRepo)
	tokenSvc := new(MockTokenSvc)
	ytClient := new(MockYouTubeClient)

	svc := newTestQueueService(queueRepo, mediaRepo, tokenRepo, tokenSvc, ytClient)

	today := time.Now().Format("2006-01-02")
	userID := uuid.New()

	quota := &domain.DailyQuota{
		ID: uuid.New(), Date: today, UnitsUsed: 0, UnitsMax: 10000, Uploads: 0,
	}

	items := []domain.UploadQueueItem{
		{
			QueueID: uuid.New(), UserID: userID,
			FilePath: "/path/video.mp4", Filename: "video.mp4",
			FileSizeBytes: 1024, QueueStatus: domain.QueueStatusPending,
		},
	}

	token := &domain.Token{
		UserID:               userID,
		EncryptedAccessToken: "encrypted-token",
		ExpiresAt:            time.Now().Add(1 * time.Hour),
	}

	uploadResp := &youtube.UploadVideoResponse{
		VideoID: "yt-video-id", Title: "video.mp4", UploadedBytes: 1024,
	}

	queueRepo.On("GetOrCreateDailyQuota", ctx, today).Return(quota, nil)
	queueRepo.On("FindPendingAll", ctx, 6).Return(items, nil)
	queueRepo.On("Update", ctx, mock.AnythingOfType("*domain.UploadQueueItem")).Return(nil)
	queueRepo.On("UpdateDailyQuota", ctx, mock.AnythingOfType("*domain.DailyQuota")).Return(nil)

	tokenRepo.On("FindByUserID", ctx, userID.String()).Return(token, nil)
	tokenSvc.On("DecryptToken", ctx, "encrypted-token").Return("decrypted-access-token", nil)

	mediaRepo.On("Create", ctx, mock.AnythingOfType("*domain.MediaAsset")).Return(nil)
	mediaRepo.On("Update", ctx, mock.AnythingOfType("*domain.MediaAsset")).Return(nil)

	ytClient.On("UploadVideo", ctx, "decrypted-access-token", mock.AnythingOfType("*youtube.UploadVideoRequest")).
		Return(uploadResp, nil)

	err := svc.ProcessQueue(ctx)

	require.NoError(t, err)
	queueRepo.AssertExpectations(t)
	tokenRepo.AssertExpectations(t)
	ytClient.AssertExpectations(t)
}

func TestProcessQueue_QuotaExhausted(t *testing.T) {
	ctx := context.Background()
	queueRepo := new(MockQueueRepository)
	svc := newTestQueueService(queueRepo, new(MockMediaRepository), new(MockTokenRepo), new(MockTokenSvc), new(MockYouTubeClient))

	today := time.Now().Format("2006-01-02")
	quota := &domain.DailyQuota{
		ID: uuid.New(), Date: today, UnitsUsed: 9600, UnitsMax: 10000, Uploads: 6,
	}

	queueRepo.On("GetOrCreateDailyQuota", ctx, today).Return(quota, nil)

	err := svc.ProcessQueue(ctx)

	require.NoError(t, err)
	// FindPendingAll should NOT be called
	queueRepo.AssertNotCalled(t, "FindPendingAll")
}

func TestProcessQueue_NoPendingItems(t *testing.T) {
	ctx := context.Background()
	queueRepo := new(MockQueueRepository)
	svc := newTestQueueService(queueRepo, new(MockMediaRepository), new(MockTokenRepo), new(MockTokenSvc), new(MockYouTubeClient))

	today := time.Now().Format("2006-01-02")
	quota := &domain.DailyQuota{
		ID: uuid.New(), Date: today, UnitsUsed: 0, UnitsMax: 10000,
	}

	queueRepo.On("GetOrCreateDailyQuota", ctx, today).Return(quota, nil)
	queueRepo.On("FindPendingAll", ctx, 6).Return([]domain.UploadQueueItem{}, nil)

	err := svc.ProcessQueue(ctx)

	require.NoError(t, err)
	queueRepo.AssertExpectations(t)
}

func TestProcessQueue_TokenExpired(t *testing.T) {
	ctx := context.Background()
	queueRepo := new(MockQueueRepository)
	mediaRepo := new(MockMediaRepository)
	tokenRepo := new(MockTokenRepo)
	tokenSvc := new(MockTokenSvc)
	ytClient := new(MockYouTubeClient)

	svc := newTestQueueService(queueRepo, mediaRepo, tokenRepo, tokenSvc, ytClient)

	today := time.Now().Format("2006-01-02")
	userID := uuid.New()

	quota := &domain.DailyQuota{
		ID: uuid.New(), Date: today, UnitsUsed: 0, UnitsMax: 10000,
	}

	items := []domain.UploadQueueItem{
		{QueueID: uuid.New(), UserID: userID, Filename: "video.mp4",
			FilePath: "/path", QueueStatus: domain.QueueStatusPending},
	}

	// Token is expired
	expiredToken := &domain.Token{
		UserID:    userID,
		ExpiresAt: time.Now().Add(-1 * time.Hour),
	}

	queueRepo.On("GetOrCreateDailyQuota", ctx, today).Return(quota, nil)
	queueRepo.On("FindPendingAll", ctx, 6).Return(items, nil)
	queueRepo.On("Update", ctx, mock.AnythingOfType("*domain.UploadQueueItem")).Return(nil)
	queueRepo.On("UpdateDailyQuota", ctx, mock.AnythingOfType("*domain.DailyQuota")).Return(nil)

	tokenRepo.On("FindByUserID", ctx, userID.String()).Return(expiredToken, nil)

	err := svc.ProcessQueue(ctx)

	require.NoError(t, err)
	queueRepo.AssertExpectations(t)
	// YouTube upload should NOT be called since token was expired
	ytClient.AssertNotCalled(t, "UploadVideo")
}

func TestProcessQueue_UploadFails_RetryPending(t *testing.T) {
	ctx := context.Background()
	queueRepo := new(MockQueueRepository)
	mediaRepo := new(MockMediaRepository)
	tokenRepo := new(MockTokenRepo)
	tokenSvc := new(MockTokenSvc)
	ytClient := new(MockYouTubeClient)

	svc := newTestQueueService(queueRepo, mediaRepo, tokenRepo, tokenSvc, ytClient)

	today := time.Now().Format("2006-01-02")
	userID := uuid.New()

	quota := &domain.DailyQuota{
		ID: uuid.New(), Date: today, UnitsUsed: 0, UnitsMax: 10000,
	}

	items := []domain.UploadQueueItem{
		{QueueID: uuid.New(), UserID: userID, Filename: "video.mp4",
			FilePath: "/path", QueueStatus: domain.QueueStatusPending, RetryCount: 0},
	}

	token := &domain.Token{
		UserID:               userID,
		EncryptedAccessToken: "enc-token",
		ExpiresAt:            time.Now().Add(1 * time.Hour),
	}

	queueRepo.On("GetOrCreateDailyQuota", ctx, today).Return(quota, nil)
	queueRepo.On("FindPendingAll", ctx, 6).Return(items, nil)
	queueRepo.On("Update", ctx, mock.AnythingOfType("*domain.UploadQueueItem")).Return(nil)

	tokenRepo.On("FindByUserID", ctx, userID.String()).Return(token, nil)
	tokenSvc.On("DecryptToken", ctx, "enc-token").Return("access-token", nil)

	mediaRepo.On("Create", ctx, mock.AnythingOfType("*domain.MediaAsset")).Return(nil)
	mediaRepo.On("Update", ctx, mock.AnythingOfType("*domain.MediaAsset")).Return(nil)

	// Upload fails
	ytClient.On("UploadVideo", ctx, "access-token", mock.AnythingOfType("*youtube.UploadVideoRequest")).
		Return(nil, fmt.Errorf("upload timeout"))
	queueRepo.On("UpdateDailyQuota", ctx, mock.AnythingOfType("*domain.DailyQuota")).Return(nil)

	err := svc.ProcessQueue(ctx)

	require.NoError(t, err)
	queueRepo.AssertExpectations(t)
}

func TestProcessQueue_MaxRetriesExceeded(t *testing.T) {
	ctx := context.Background()
	queueRepo := new(MockQueueRepository)
	mediaRepo := new(MockMediaRepository)
	tokenRepo := new(MockTokenRepo)
	tokenSvc := new(MockTokenSvc)
	ytClient := new(MockYouTubeClient)

	svc := newTestQueueService(queueRepo, mediaRepo, tokenRepo, tokenSvc, ytClient)

	today := time.Now().Format("2006-01-02")
	userID := uuid.New()

	quota := &domain.DailyQuota{
		ID: uuid.New(), Date: today, UnitsUsed: 0, UnitsMax: 10000,
	}

	items := []domain.UploadQueueItem{
		{QueueID: uuid.New(), UserID: userID, Filename: "video.mp4",
			FilePath: "/path", QueueStatus: domain.QueueStatusPending,
			RetryCount: domain.MaxRetryAttempts - 1}, // Already at max-1, next failure = max
	}

	token := &domain.Token{
		UserID:               userID,
		EncryptedAccessToken: "enc-token",
		ExpiresAt:            time.Now().Add(1 * time.Hour),
	}

	queueRepo.On("GetOrCreateDailyQuota", ctx, today).Return(quota, nil)
	queueRepo.On("FindPendingAll", ctx, 6).Return(items, nil)
	queueRepo.On("Update", ctx, mock.AnythingOfType("*domain.UploadQueueItem")).Return(nil)

	tokenRepo.On("FindByUserID", ctx, userID.String()).Return(token, nil)
	tokenSvc.On("DecryptToken", ctx, "enc-token").Return("access-token", nil)

	mediaRepo.On("Create", ctx, mock.AnythingOfType("*domain.MediaAsset")).Return(nil)
	mediaRepo.On("Update", ctx, mock.AnythingOfType("*domain.MediaAsset")).Return(nil)

	ytClient.On("UploadVideo", ctx, "access-token", mock.AnythingOfType("*youtube.UploadVideoRequest")).
		Return(nil, fmt.Errorf("upload error"))
	queueRepo.On("UpdateDailyQuota", ctx, mock.AnythingOfType("*domain.DailyQuota")).Return(nil)

	err := svc.ProcessQueue(ctx)

	require.NoError(t, err)
	queueRepo.AssertExpectations(t)
}

func TestProcessQueue_MultipleItems(t *testing.T) {
	ctx := context.Background()
	queueRepo := new(MockQueueRepository)
	mediaRepo := new(MockMediaRepository)
	tokenRepo := new(MockTokenRepo)
	tokenSvc := new(MockTokenSvc)
	ytClient := new(MockYouTubeClient)

	svc := newTestQueueService(queueRepo, mediaRepo, tokenRepo, tokenSvc, ytClient)

	today := time.Now().Format("2006-01-02")
	userID := uuid.New()

	quota := &domain.DailyQuota{
		ID: uuid.New(), Date: today, UnitsUsed: 0, UnitsMax: 10000,
	}

	items := []domain.UploadQueueItem{
		{QueueID: uuid.New(), UserID: userID, Filename: "video1.mp4", FilePath: "/path1", QueueStatus: domain.QueueStatusPending},
		{QueueID: uuid.New(), UserID: userID, Filename: "video2.mp4", FilePath: "/path2", QueueStatus: domain.QueueStatusPending},
	}

	token := &domain.Token{
		UserID:               userID,
		EncryptedAccessToken: "enc",
		ExpiresAt:            time.Now().Add(1 * time.Hour),
	}

	uploadResp := &youtube.UploadVideoResponse{VideoID: "yt-id", Title: "video", UploadedBytes: 1024}

	queueRepo.On("GetOrCreateDailyQuota", ctx, today).Return(quota, nil)
	queueRepo.On("FindPendingAll", ctx, 6).Return(items, nil)
	queueRepo.On("Update", ctx, mock.AnythingOfType("*domain.UploadQueueItem")).Return(nil)
	queueRepo.On("UpdateDailyQuota", ctx, mock.AnythingOfType("*domain.DailyQuota")).Return(nil)

	tokenRepo.On("FindByUserID", ctx, userID.String()).Return(token, nil)
	tokenSvc.On("DecryptToken", ctx, "enc").Return("access", nil)

	mediaRepo.On("Create", ctx, mock.AnythingOfType("*domain.MediaAsset")).Return(nil)
	mediaRepo.On("Update", ctx, mock.AnythingOfType("*domain.MediaAsset")).Return(nil)

	ytClient.On("UploadVideo", ctx, "access", mock.AnythingOfType("*youtube.UploadVideoRequest")).
		Return(uploadResp, nil)

	err := svc.ProcessQueue(ctx)

	require.NoError(t, err)
	// Verify YouTube upload was called twice (once per item)
	ytClient.AssertNumberOfCalls(t, "UploadVideo", 2)
}

func TestProcessQueue_QuotaError(t *testing.T) {
	ctx := context.Background()
	queueRepo := new(MockQueueRepository)
	svc := newTestQueueService(queueRepo, new(MockMediaRepository), new(MockTokenRepo), new(MockTokenSvc), new(MockYouTubeClient))

	today := time.Now().Format("2006-01-02")
	queueRepo.On("GetOrCreateDailyQuota", ctx, today).
		Return(nil, fmt.Errorf("db error"))

	err := svc.ProcessQueue(ctx)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get daily quota")
}
