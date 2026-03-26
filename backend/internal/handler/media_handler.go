package handler

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/JokerTrickster/video-upload-backend/internal/domain"
	"github.com/JokerTrickster/video-upload-backend/internal/repository"
	"github.com/JokerTrickster/video-upload-backend/internal/service"
)

const (
	// MaxUploadSize is the maximum upload size (2GB)
	MaxUploadSize = 2 * 1024 * 1024 * 1024
)

// getTempUploadDir returns the temporary upload directory from env or OS default
func getTempUploadDir() string {
	if dir := os.Getenv("UPLOAD_TEMP_DIR"); dir != "" {
		return dir
	}
	return filepath.Join(os.TempDir(), "media-backup-uploads")
}

// MediaHandler handles media-related HTTP requests
type MediaHandler struct {
	uploadService service.UploadService
	tokenService  service.TokenService
	tokenRepo     repository.TokenRepository
}

// NewMediaHandler creates a new media handler
func NewMediaHandler(uploadService service.UploadService, tokenService service.TokenService, tokenRepo repository.TokenRepository) *MediaHandler {
	// Ensure temp upload directory exists
	os.MkdirAll(getTempUploadDir(), 0755)

	return &MediaHandler{
		uploadService: uploadService,
		tokenService:  tokenService,
		tokenRepo:     tokenRepo,
	}
}

// InitiateUpload handles POST /api/v1/media/upload/initiate
func (h *MediaHandler) InitiateUpload(c *gin.Context) {
	// Get user ID from context (set by auth middleware)
	userIDStr, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error:   domain.ErrorCodeAuthInvalid,
			Message: "User not authenticated",
		})
		return
	}

	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "INVALID_USER_ID",
			Message: "Invalid user ID",
		})
		return
	}

	// Parse request
	var req InitiateUploadRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "INVALID_REQUEST",
			Message: err.Error(),
		})
		return
	}

	// Create upload session
	session, err := h.uploadService.InitiateUploadSession(c.Request.Context(), userID, req.TotalFiles, req.TotalBytes)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "SESSION_CREATE_FAILED",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, SuccessResponse{
		Success: true,
		Data: InitiateUploadResponse{
			SessionID: session.SessionID.String(),
			StartedAt: session.StartedAt,
		},
	})
}

// UploadVideo handles POST /api/v1/media/upload/video
func (h *MediaHandler) UploadVideo(c *gin.Context) {
	// Get user ID from context
	userIDStr, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error:   domain.ErrorCodeAuthInvalid,
			Message: "User not authenticated",
		})
		return
	}

	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "INVALID_USER_ID",
			Message: "Invalid user ID",
		})
		return
	}

	// Get session ID from form
	sessionIDStr := c.PostForm("session_id")
	if sessionIDStr == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "MISSING_SESSION_ID",
			Message: "Session ID is required",
		})
		return
	}

	sessionID, err := uuid.Parse(sessionIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "INVALID_SESSION_ID",
			Message: "Invalid session ID",
		})
		return
	}

	// Get uploaded file
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "FILE_MISSING",
			Message: "File is required",
		})
		return
	}
	defer file.Close()

	// Check file size
	if header.Size > MaxUploadSize {
		c.JSON(http.StatusRequestEntityTooLarge, ErrorResponse{
			Error:   domain.ErrorCodeFileTooLarge,
			Message: "File size exceeds 2GB limit",
		})
		return
	}

	// Save file to temporary location
	tempFilePath := filepath.Join(getTempUploadDir(), fmt.Sprintf("%s_%s", uuid.New().String(), header.Filename))
	tempFile, err := os.Create(tempFilePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "FILE_SAVE_FAILED",
			Message: "Failed to save uploaded file",
		})
		return
	}
	defer func() {
		tempFile.Close()
		os.Remove(tempFilePath) // Clean up temp file
	}()

	// Copy uploaded file to temp file
	_, err = io.Copy(tempFile, file)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "FILE_COPY_FAILED",
			Message: "Failed to copy uploaded file",
		})
		return
	}

	// Retrieve user's OAuth access token from database
	userToken, err := h.tokenRepo.FindByUserID(c.Request.Context(), userID.String())
	if err != nil {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error:   domain.ErrorCodeAuthInvalid,
			Message: "OAuth token not found. Please re-authenticate with Google.",
		})
		return
	}

	// Check if OAuth token is expired
	if userToken.IsExpired() {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error:   domain.ErrorCodeTokenExpired,
			Message: "OAuth token expired. Please re-authenticate with Google.",
		})
		return
	}

	// Decrypt the stored OAuth access token
	accessToken, err := h.tokenService.DecryptToken(c.Request.Context(), userToken.EncryptedAccessToken)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "TOKEN_DECRYPT_FAILED",
			Message: "Failed to decrypt access token",
		})
		return
	}

	// Upload to YouTube
	uploadReq := &service.UploadVideoRequest{
		SessionID:     sessionID,
		UserID:        userID,
		AccessToken:   accessToken,
		FilePath:      tempFilePath,
		Filename:      header.Filename,
		FileSizeBytes: header.Size,
		Title:         c.PostForm("title"),
		Description:   c.PostForm("description"),
	}

	result, err := h.uploadService.UploadVideo(c.Request.Context(), uploadReq)
	if err != nil {
		// Determine error code
		errorCode := "UPLOAD_FAILED"
		if domain.ShouldRetry(err) {
			errorCode = "UPLOAD_RETRY"
		}

		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   errorCode,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, SuccessResponse{
		Success: true,
		Message: "Video uploaded successfully",
		Data: UploadVideoResponse{
			AssetID:           result.AssetID.String(),
			YouTubeVideoID:    result.VideoID,
			OriginalFilename:  result.Filename,
			FileSizeBytes:     result.FileSizeBytes,
			SyncStatus:        result.SyncStatus,
			UploadCompletedAt: result.UploadedAt,
		},
	})
}

// GetUploadSessionStatus handles GET /api/v1/media/upload/status/:session_id
func (h *MediaHandler) GetUploadSessionStatus(c *gin.Context) {
	// Get user ID from context
	userIDStr, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error:   domain.ErrorCodeAuthInvalid,
			Message: "User not authenticated",
		})
		return
	}

	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "INVALID_USER_ID",
			Message: "Invalid user ID",
		})
		return
	}

	// Get session ID from URL
	sessionIDStr := c.Param("session_id")
	sessionID, err := uuid.Parse(sessionIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "INVALID_SESSION_ID",
			Message: "Invalid session ID",
		})
		return
	}

	// Get session status
	status, err := h.uploadService.GetUploadSessionStatus(c.Request.Context(), sessionID)
	if err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse{
			Error:   "SESSION_NOT_FOUND",
			Message: err.Error(),
		})
		return
	}

	// Verify session belongs to authenticated user
	if status.UserID != userID {
		c.JSON(http.StatusForbidden, ErrorResponse{
			Error:   "SESSION_ACCESS_DENIED",
			Message: "You do not have access to this session",
		})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Success: true,
		Data: UploadSessionStatusResponse{
			SessionID:      status.SessionID.String(),
			UserID:         status.UserID.String(),
			TotalFiles:     status.TotalFiles,
			CompletedFiles: status.CompletedFiles,
			FailedFiles:    status.FailedFiles,
			TotalBytes:     status.TotalBytes,
			UploadedBytes:  status.UploadedBytes,
			SessionStatus:  status.SessionStatus,
			StartedAt:      status.StartedAt,
			CompletedAt:    status.CompletedAt,
		},
	})
}

// CompleteUploadSession handles POST /api/v1/media/upload/complete
func (h *MediaHandler) CompleteUploadSession(c *gin.Context) {
	// Get session ID from request body
	var req struct {
		SessionID string `json:"session_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "INVALID_REQUEST",
			Message: err.Error(),
		})
		return
	}

	sessionID, err := uuid.Parse(req.SessionID)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "INVALID_SESSION_ID",
			Message: "Invalid session ID",
		})
		return
	}

	// Complete session
	err = h.uploadService.CompleteUploadSession(c.Request.Context(), sessionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "SESSION_COMPLETE_FAILED",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Success: true,
		Message: "Upload session completed successfully",
	})
}

// CancelUploadSession handles POST /api/v1/media/upload/cancel
func (h *MediaHandler) CancelUploadSession(c *gin.Context) {
	var req struct {
		SessionID string `json:"session_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "INVALID_REQUEST",
			Message: err.Error(),
		})
		return
	}

	sessionID, err := uuid.Parse(req.SessionID)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "INVALID_SESSION_ID",
			Message: "Invalid session ID",
		})
		return
	}

	err = h.uploadService.CancelUploadSession(c.Request.Context(), sessionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "SESSION_CANCEL_FAILED",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Success: true,
		Message: "Upload session cancelled successfully",
	})
}

// ListMediaAssets handles GET /api/v1/media/list
func (h *MediaHandler) ListMediaAssets(c *gin.Context) {
	// Get user ID from context
	userIDStr, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error:   domain.ErrorCodeAuthInvalid,
			Message: "User not authenticated",
		})
		return
	}

	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "INVALID_USER_ID",
			Message: "Invalid user ID",
		})
		return
	}

	// Parse query parameters with validation
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 50
	}
	mediaType := c.Query("media_type")
	syncStatus := c.Query("sync_status")
	sort := c.DefaultQuery("sort", "created_at_desc")

	opts := &service.ListMediaOptions{
		Page:       page,
		Limit:      limit,
		MediaType:  mediaType,
		SyncStatus: syncStatus,
		Sort:       sort,
	}

	// List assets
	result, err := h.uploadService.ListMediaAssets(c.Request.Context(), userID, opts)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "LIST_FAILED",
			Message: err.Error(),
		})
		return
	}

	// Convert to response DTOs
	assets := make([]MediaAssetResponse, len(result.Assets))
	for i, asset := range result.Assets {
		assets[i] = MediaAssetResponse{
			AssetID:           asset.AssetID.String(),
			YouTubeVideoID:    asset.YouTubeVideoID,
			S3ObjectKey:       asset.S3ObjectKey,
			ThumbnailURL:      asset.ThumbnailURL,
			OriginalFilename:  asset.OriginalFilename,
			FileSizeBytes:     asset.FileSizeBytes,
			MediaType:         string(asset.MediaType),
			SyncStatus:        string(asset.SyncStatus),
			CreatedAt:         asset.CreatedAt,
			UploadStartedAt:   asset.UploadStartedAt,
			UploadCompletedAt: asset.UploadCompletedAt,
			ErrorMessage:      asset.ErrorMessage,
			RetryCount:        asset.RetryCount,
		}
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Success: true,
		Data: MediaAssetListResponse{
			Assets: assets,
			Pagination: PaginationResponse{
				Page:       result.Page,
				Limit:      result.Limit,
				Total:      result.Total,
				TotalPages: result.TotalPages,
			},
		},
	})
}

// GetMediaAsset handles GET /api/v1/media/:asset_id
func (h *MediaHandler) GetMediaAsset(c *gin.Context) {
	// Get asset ID from URL
	assetIDStr := c.Param("asset_id")
	assetID, err := uuid.Parse(assetIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "INVALID_ASSET_ID",
			Message: "Invalid asset ID",
		})
		return
	}

	// Get asset
	asset, err := h.uploadService.GetMediaAsset(c.Request.Context(), assetID)
	if err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse{
			Error:   "ASSET_NOT_FOUND",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Success: true,
		Data: MediaAssetResponse{
			AssetID:           asset.AssetID.String(),
			YouTubeVideoID:    asset.YouTubeVideoID,
			S3ObjectKey:       asset.S3ObjectKey,
			ThumbnailURL:      asset.ThumbnailURL,
			OriginalFilename:  asset.OriginalFilename,
			FileSizeBytes:     asset.FileSizeBytes,
			MediaType:         string(asset.MediaType),
			SyncStatus:        string(asset.SyncStatus),
			CreatedAt:         asset.CreatedAt,
			UploadStartedAt:   asset.UploadStartedAt,
			UploadCompletedAt: asset.UploadCompletedAt,
			ErrorMessage:      asset.ErrorMessage,
			RetryCount:        asset.RetryCount,
		},
	})
}

// DeleteMediaAsset handles DELETE /api/v1/media/:asset_id
func (h *MediaHandler) DeleteMediaAsset(c *gin.Context) {
	// Get asset ID from URL
	assetIDStr := c.Param("asset_id")
	assetID, err := uuid.Parse(assetIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "INVALID_ASSET_ID",
			Message: "Invalid asset ID",
		})
		return
	}

	// Delete asset (not from YouTube, just the record)
	err = h.uploadService.DeleteMediaAsset(c.Request.Context(), assetID, false)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "DELETE_FAILED",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Success: true,
		Message: "Media asset deleted successfully",
	})
}
