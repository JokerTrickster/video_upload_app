package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/JokerTrickster/video-upload-backend/internal/domain"
	"github.com/JokerTrickster/video-upload-backend/internal/repository"
	"github.com/JokerTrickster/video-upload-backend/internal/service"
)

// MockUploadService implements service.UploadService
type MockUploadService struct {
	mock.Mock
}

func (m *MockUploadService) InitiateUploadSession(ctx context.Context, userID uuid.UUID, totalFiles int, totalBytes int64) (*domain.UploadSession, error) {
	args := m.Called(ctx, userID, totalFiles, totalBytes)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.UploadSession), args.Error(1)
}

func (m *MockUploadService) UploadVideo(ctx context.Context, req *service.UploadVideoRequest) (*service.UploadVideoResult, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*service.UploadVideoResult), args.Error(1)
}

func (m *MockUploadService) GetUploadSessionStatus(ctx context.Context, sessionID uuid.UUID) (*service.UploadSessionStatus, error) {
	args := m.Called(ctx, sessionID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*service.UploadSessionStatus), args.Error(1)
}

func (m *MockUploadService) CompleteUploadSession(ctx context.Context, sessionID uuid.UUID) error {
	args := m.Called(ctx, sessionID)
	return args.Error(0)
}

func (m *MockUploadService) CancelUploadSession(ctx context.Context, sessionID uuid.UUID) error {
	args := m.Called(ctx, sessionID)
	return args.Error(0)
}

func (m *MockUploadService) ListMediaAssets(ctx context.Context, userID uuid.UUID, opts *service.ListMediaOptions) (*service.MediaAssetList, error) {
	args := m.Called(ctx, userID, opts)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*service.MediaAssetList), args.Error(1)
}

func (m *MockUploadService) GetMediaAsset(ctx context.Context, assetID uuid.UUID) (*domain.MediaAsset, error) {
	args := m.Called(ctx, assetID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.MediaAsset), args.Error(1)
}

func (m *MockUploadService) DeleteMediaAsset(ctx context.Context, assetID uuid.UUID, deleteFromYouTube bool) error {
	args := m.Called(ctx, assetID, deleteFromYouTube)
	return args.Error(0)
}

// MockTokenService implements service.TokenService
type MockTokenService struct {
	mock.Mock
}

func (m *MockTokenService) EncryptToken(ctx context.Context, plainText string) (string, error) {
	args := m.Called(ctx, plainText)
	return args.String(0), args.Error(1)
}

func (m *MockTokenService) DecryptToken(ctx context.Context, cipherText string) (string, error) {
	args := m.Called(ctx, cipherText)
	return args.String(0), args.Error(1)
}

func (m *MockTokenService) AddToBlacklist(ctx context.Context, token string, expirySeconds int64) error {
	args := m.Called(ctx, token, expirySeconds)
	return args.Error(0)
}

func (m *MockTokenService) IsBlacklisted(ctx context.Context, token string) (bool, error) {
	args := m.Called(ctx, token)
	return args.Bool(0), args.Error(1)
}

func (m *MockTokenService) SaveOAuthState(ctx context.Context, state string, ttlSeconds int64) error {
	args := m.Called(ctx, state, ttlSeconds)
	return args.Error(0)
}

func (m *MockTokenService) ValidateOAuthState(ctx context.Context, state string) (bool, error) {
	args := m.Called(ctx, state)
	return args.Bool(0), args.Error(1)
}

// MockTokenRepository implements repository.TokenRepository
type MockTokenRepository struct {
	mock.Mock
}

func (m *MockTokenRepository) Create(ctx context.Context, token *domain.Token) error {
	args := m.Called(ctx, token)
	return args.Error(0)
}

func (m *MockTokenRepository) FindByUserID(ctx context.Context, userID string) (*domain.Token, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Token), args.Error(1)
}

func (m *MockTokenRepository) Update(ctx context.Context, token *domain.Token) error {
	args := m.Called(ctx, token)
	return args.Error(0)
}

func (m *MockTokenRepository) Delete(ctx context.Context, userID string) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

// Ensure mock implements the interface at compile time
var _ repository.TokenRepository = (*MockTokenRepository)(nil)

func setupTestRouter(mockUpload *MockUploadService, mockToken *MockTokenService) *gin.Engine {
	return setupTestRouterWithTokenRepo(mockUpload, mockToken, new(MockTokenRepository))
}

func setupTestRouterWithTokenRepo(mockUpload *MockUploadService, mockToken *MockTokenService, mockTokenRepo *MockTokenRepository) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	handler := NewMediaHandler(mockUpload, mockToken, mockTokenRepo)

	// Add middleware to set user_id
	router.Use(func(c *gin.Context) {
		c.Set("user_id", "550e8400-e29b-41d4-a716-446655440000")
		c.Set("access_token", "test-access-token")
		c.Next()
	})

	v1 := router.Group("/api/v1/media")
	{
		v1.POST("/upload/initiate", handler.InitiateUpload)
		v1.POST("/upload/video", handler.UploadVideo)
		v1.GET("/upload/status/:session_id", handler.GetUploadSessionStatus)
		v1.POST("/upload/complete", handler.CompleteUploadSession)
		v1.GET("/list", handler.ListMediaAssets)
		v1.GET("/:asset_id", handler.GetMediaAsset)
		v1.DELETE("/:asset_id", handler.DeleteMediaAsset)
	}

	return router
}

func TestInitiateUpload_Success(t *testing.T) {
	mockUpload := new(MockUploadService)
	mockToken := new(MockTokenService)
	router := setupTestRouter(mockUpload, mockToken)

	userID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")
	sessionID := uuid.New()

	session := &domain.UploadSession{
		SessionID:     sessionID,
		UserID:        userID,
		TotalFiles:    10,
		TotalBytes:    1024 * 1024 * 1024,
		SessionStatus: "ACTIVE",
		StartedAt:     time.Now(),
	}

	mockUpload.On("InitiateUploadSession", mock.Anything, userID, 10, int64(1024*1024*1024)).Return(session, nil)

	reqBody := InitiateUploadRequest{
		TotalFiles: 10,
		TotalBytes: 1024 * 1024 * 1024,
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/media/upload/initiate", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var response SuccessResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.True(t, response.Success)

	mockUpload.AssertExpectations(t)
}

func TestInitiateUpload_InvalidRequest(t *testing.T) {
	mockUpload := new(MockUploadService)
	mockToken := new(MockTokenService)
	router := setupTestRouter(mockUpload, mockToken)

	reqBody := map[string]interface{}{
		"total_files": -1, // Invalid
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/media/upload/initiate", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestGetUploadSessionStatus_Success(t *testing.T) {
	mockUpload := new(MockUploadService)
	mockToken := new(MockTokenService)
	router := setupTestRouter(mockUpload, mockToken)

	sessionID := uuid.New()
	userID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")

	status := &service.UploadSessionStatus{
		SessionID:      sessionID,
		UserID:         userID,
		TotalFiles:     10,
		CompletedFiles: 5,
		FailedFiles:    0,
		TotalBytes:     1024 * 1024 * 1024,
		UploadedBytes:  512 * 1024 * 1024,
		SessionStatus:  "ACTIVE",
		StartedAt:      time.Now(),
	}

	mockUpload.On("GetUploadSessionStatus", mock.Anything, sessionID).Return(status, nil)

	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/media/upload/status/%s", sessionID.String()), nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response SuccessResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.True(t, response.Success)

	mockUpload.AssertExpectations(t)
}

func TestGetUploadSessionStatus_InvalidUUID(t *testing.T) {
	mockUpload := new(MockUploadService)
	mockToken := new(MockTokenService)
	router := setupTestRouter(mockUpload, mockToken)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/media/upload/status/not-a-uuid", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCompleteUploadSession_Success(t *testing.T) {
	mockUpload := new(MockUploadService)
	mockToken := new(MockTokenService)
	router := setupTestRouter(mockUpload, mockToken)

	sessionID := uuid.New()

	mockUpload.On("CompleteUploadSession", mock.Anything, sessionID).Return(nil)

	reqBody := map[string]string{
		"session_id": sessionID.String(),
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/media/upload/complete", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response SuccessResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.True(t, response.Success)

	mockUpload.AssertExpectations(t)
}

func TestListMediaAssets_Success(t *testing.T) {
	mockUpload := new(MockUploadService)
	mockToken := new(MockTokenService)
	router := setupTestRouter(mockUpload, mockToken)

	userID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")
	assetID1 := uuid.New()
	videoID1 := "video-id-1"

	assets := []*domain.MediaAsset{
		{
			AssetID:          assetID1,
			UserID:           userID,
			YouTubeVideoID:   &videoID1,
			OriginalFilename: "test1.mp4",
			FileSizeBytes:    1024,
			MediaType:        "VIDEO",
			SyncStatus:       "COMPLETED",
			CreatedAt:        time.Now(),
		},
	}

	result := &service.MediaAssetList{
		Assets:     assets,
		Page:       1,
		Limit:      50,
		Total:      1,
		TotalPages: 1,
	}

	mockUpload.On("ListMediaAssets", mock.Anything, userID, mock.AnythingOfType("*service.ListMediaOptions")).Return(result, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/media/list?page=1&limit=50", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response SuccessResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.True(t, response.Success)

	mockUpload.AssertExpectations(t)
}

func TestGetMediaAsset_Success(t *testing.T) {
	mockUpload := new(MockUploadService)
	mockToken := new(MockTokenService)
	router := setupTestRouter(mockUpload, mockToken)

	assetID := uuid.New()
	videoID := "test-video-id"

	asset := &domain.MediaAsset{
		AssetID:          assetID,
		YouTubeVideoID:   &videoID,
		OriginalFilename: "test.mp4",
		FileSizeBytes:    1024,
		MediaType:        "VIDEO",
		SyncStatus:       "COMPLETED",
		CreatedAt:        time.Now(),
	}

	mockUpload.On("GetMediaAsset", mock.Anything, assetID).Return(asset, nil)

	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/media/%s", assetID.String()), nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response SuccessResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.True(t, response.Success)

	mockUpload.AssertExpectations(t)
}

func TestGetMediaAsset_InvalidUUID(t *testing.T) {
	mockUpload := new(MockUploadService)
	mockToken := new(MockTokenService)
	router := setupTestRouter(mockUpload, mockToken)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/media/not-a-uuid", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestDeleteMediaAsset_Success(t *testing.T) {
	mockUpload := new(MockUploadService)
	mockToken := new(MockTokenService)
	router := setupTestRouter(mockUpload, mockToken)

	assetID := uuid.New()

	mockUpload.On("DeleteMediaAsset", mock.Anything, assetID, false).Return(nil)

	req := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/api/v1/media/%s", assetID.String()), nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response SuccessResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.True(t, response.Success)

	mockUpload.AssertExpectations(t)
}

func TestUploadVideo_MissingFile(t *testing.T) {
	mockUpload := new(MockUploadService)
	mockToken := new(MockTokenService)
	router := setupTestRouter(mockUpload, mockToken)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	writer.WriteField("session_id", uuid.New().String())
	writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/api/v1/media/upload/video", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "FILE_MISSING", response.Error)
}

func TestUploadVideo_MissingSessionID(t *testing.T) {
	mockUpload := new(MockUploadService)
	mockToken := new(MockTokenService)
	router := setupTestRouter(mockUpload, mockToken)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	// No session_id field
	writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/api/v1/media/upload/video", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}
