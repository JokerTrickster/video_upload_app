package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/JokerTrickster/video-upload-backend/internal/domain"
)

// RespondJSON sends a JSON response
func RespondJSON(c *gin.Context, status int, data interface{}) {
	c.JSON(status, data)
}

// RespondSuccess sends a success response
func RespondSuccess(c *gin.Context, message string, data interface{}) {
	RespondJSON(c, http.StatusOK, SuccessResponse{
		Success: true,
		Message: message,
		Data:    data,
	})
}

// RespondError sends an error response
func RespondError(c *gin.Context, status int, err error, details interface{}) {
	errorCode := "internal_server_error"
	message := err.Error()

	// Map domain errors to HTTP status codes and error codes
	switch err {
	case domain.ErrUserNotFound:
		status = http.StatusNotFound
		errorCode = "user_not_found"
	case domain.ErrUserAlreadyExists:
		status = http.StatusConflict
		errorCode = "user_already_exists"
	case domain.ErrInvalidUserData:
		status = http.StatusBadRequest
		errorCode = "invalid_user_data"
	case domain.ErrTokenNotFound:
		status = http.StatusNotFound
		errorCode = "token_not_found"
	case domain.ErrTokenExpired:
		status = http.StatusUnauthorized
		errorCode = "token_expired"
	case domain.ErrTokenInvalid:
		status = http.StatusUnauthorized
		errorCode = "token_invalid"
	case domain.ErrTokenBlacklisted:
		status = http.StatusUnauthorized
		errorCode = "token_blacklisted"
	case domain.ErrUnauthorized:
		status = http.StatusUnauthorized
		errorCode = "unauthorized"
	case domain.ErrInvalidCredentials:
		status = http.StatusUnauthorized
		errorCode = "invalid_credentials"
	case domain.ErrInvalidState:
		status = http.StatusUnauthorized
		errorCode = "invalid_state"
	case domain.ErrOAuthFailed:
		status = http.StatusUnauthorized
		errorCode = "oauth_failed"
	case domain.ErrInvalidInput:
		status = http.StatusBadRequest
		errorCode = "invalid_input"
	}

	response := ErrorResponse{
		Error:   errorCode,
		Message: message,
		Details: details,
	}

	RespondJSON(c, status, response)
}

// RespondBadRequest sends a 400 Bad Request response
func RespondBadRequest(c *gin.Context, message string, details interface{}) {
	RespondJSON(c, http.StatusBadRequest, ErrorResponse{
		Error:   "bad_request",
		Message: message,
		Details: details,
	})
}

// RespondUnauthorized sends a 401 Unauthorized response
func RespondUnauthorized(c *gin.Context, message string) {
	RespondJSON(c, http.StatusUnauthorized, ErrorResponse{
		Error:   "unauthorized",
		Message: message,
	})
}

// RespondNotFound sends a 404 Not Found response
func RespondNotFound(c *gin.Context, message string) {
	RespondJSON(c, http.StatusNotFound, ErrorResponse{
		Error:   "not_found",
		Message: message,
	})
}

// RespondInternalServerError sends a 500 Internal Server Error response
func RespondInternalServerError(c *gin.Context, message string) {
	RespondJSON(c, http.StatusInternalServerError, ErrorResponse{
		Error:   "internal_server_error",
		Message: message,
	})
}

// RespondTooManyRequests sends a 429 Too Many Requests response
func RespondTooManyRequests(c *gin.Context, message string) {
	RespondJSON(c, http.StatusTooManyRequests, ErrorResponse{
		Error:   "rate_limit_exceeded",
		Message: message,
	})
}
