package middleware

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/JokerTrickster/video-upload-backend/internal/handler"
	"github.com/JokerTrickster/video-upload-backend/internal/pkg/logger"
)

// ErrorHandlerMiddleware creates error recovery middleware
func ErrorHandlerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// Log the panic
				logger.Error("Panic recovered",
					"error", fmt.Sprintf("%v", err),
					"path", c.Request.URL.Path,
					"method", c.Request.Method,
					"client_ip", c.ClientIP(),
				)

				// Return 500 error
				handler.RespondInternalServerError(c, "Internal server error")
			}
		}()

		c.Next()

		// Check if there were any errors during request processing
		if len(c.Errors) > 0 {
			err := c.Errors.Last()

			// Log the error
			logger.Error("Request error",
				"error", err.Error(),
				"path", c.Request.URL.Path,
				"method", c.Request.Method,
				"client_ip", c.ClientIP(),
			)

			// If response hasn't been written yet, send error response
			if !c.Writer.Written() {
				handler.RespondError(c, http.StatusInternalServerError, err.Err, nil)
			}
		}
	}
}

// RequestLoggerMiddleware creates request logging middleware
func RequestLoggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Log request
		logger.Info("Incoming request",
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
			"client_ip", c.ClientIP(),
			"user_agent", c.Request.UserAgent(),
		)

		// Process request
		c.Next()

		// Log response
		logger.Info("Request completed",
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
			"status", c.Writer.Status(),
			"client_ip", c.ClientIP(),
		)
	}
}
