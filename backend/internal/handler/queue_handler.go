package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/JokerTrickster/video-upload-backend/internal/domain"
	"github.com/JokerTrickster/video-upload-backend/internal/service"
)

type QueueHandler struct {
	queueService service.QueueService
}

func NewQueueHandler(queueService service.QueueService) *QueueHandler {
	return &QueueHandler{queueService: queueService}
}

// AddToQueue handles POST /api/v1/queue/add
func (h *QueueHandler) AddToQueue(c *gin.Context) {
	userIDStr, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error: domain.ErrorCodeAuthInvalid, Message: "User not authenticated",
		})
		return
	}
	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "INVALID_USER_ID", Message: "Invalid user ID"})
		return
	}

	var req struct {
		FilePath    string `json:"file_path" binding:"required"`
		Filename    string `json:"filename" binding:"required"`
		FileSizeBytes int64 `json:"file_size_bytes" binding:"required,min=1"`
		Title       string `json:"title"`
		Description string `json:"description"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "INVALID_REQUEST", Message: err.Error()})
		return
	}

	item, err := h.queueService.AddToQueue(
		c.Request.Context(), userID,
		req.FilePath, req.Filename, req.FileSizeBytes,
		req.Title, req.Description,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "QUEUE_ADD_FAILED", Message: err.Error()})
		return
	}

	c.JSON(http.StatusCreated, SuccessResponse{
		Success: true,
		Message: "Added to upload queue",
		Data: map[string]interface{}{
			"queue_id":   item.QueueID.String(),
			"filename":   item.Filename,
			"status":     string(item.QueueStatus),
			"created_at": item.CreatedAt,
		},
	})
}

// GetQueue handles GET /api/v1/queue
func (h *QueueHandler) GetQueue(c *gin.Context) {
	userIDStr, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error: domain.ErrorCodeAuthInvalid, Message: "User not authenticated",
		})
		return
	}
	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "INVALID_USER_ID", Message: "Invalid user ID"})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))

	items, total, err := h.queueService.GetQueueItems(c.Request.Context(), userID, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "QUEUE_FETCH_FAILED", Message: err.Error()})
		return
	}

	queueItems := make([]map[string]interface{}, len(items))
	for i, item := range items {
		queueItems[i] = map[string]interface{}{
			"queue_id":       item.QueueID.String(),
			"filename":       item.Filename,
			"file_size_bytes": item.FileSizeBytes,
			"title":          item.Title,
			"status":         string(item.QueueStatus),
			"retry_count":    item.RetryCount,
			"error_message":  item.ErrorMessage,
			"created_at":     item.CreatedAt,
			"processed_at":   item.ProcessedAt,
		}
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Success: true,
		Data: map[string]interface{}{
			"items":       queueItems,
			"total":       total,
			"page":        page,
			"limit":       limit,
		},
	})
}

// RemoveFromQueue handles DELETE /api/v1/queue/:queue_id
func (h *QueueHandler) RemoveFromQueue(c *gin.Context) {
	queueIDStr := c.Param("queue_id")
	queueID, err := uuid.Parse(queueIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "INVALID_QUEUE_ID", Message: "Invalid queue ID"})
		return
	}

	if err := h.queueService.RemoveFromQueue(c.Request.Context(), queueID); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "QUEUE_REMOVE_FAILED", Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{Success: true, Message: "Removed from queue"})
}

// GetQuotaStatus handles GET /api/v1/queue/quota
func (h *QueueHandler) GetQuotaStatus(c *gin.Context) {
	quota, err := h.queueService.GetQuotaStatus(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "QUOTA_FETCH_FAILED", Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Success: true,
		Data: map[string]interface{}{
			"date":              quota.Date,
			"units_used":        quota.UnitsUsed,
			"units_max":         quota.UnitsMax,
			"uploads_today":     quota.Uploads,
			"remaining_uploads": quota.RemainingUploads(),
			"can_upload":        quota.CanUpload(),
		},
	})
}
