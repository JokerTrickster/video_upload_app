package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/JokerTrickster/video-upload-backend/internal/handler"
	"github.com/JokerTrickster/video-upload-backend/internal/service"
)

// AuthMiddleware creates JWT authentication middleware
func AuthMiddleware(authService service.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extract Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			handler.RespondUnauthorized(c, "Authorization header missing")
			c.Abort()
			return
		}

		// Parse Bearer token
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			handler.RespondBadRequest(c, "Invalid authorization header format", "Expected format: Bearer <token>")
			c.Abort()
			return
		}

		token := parts[1]

		// Validate JWT token
		claims, err := authService.ValidateJWT(c.Request.Context(), token)
		if err != nil {
			handler.RespondUnauthorized(c, "Invalid or expired token")
			c.Abort()
			return
		}

		// Ensure it's an access token, not a refresh token
		if claims.TokenType != "access" {
			handler.RespondUnauthorized(c, "Invalid token type")
			c.Abort()
			return
		}

		// Set user information in context
		c.Set("user_id", claims.UserID)
		c.Set("user_email", claims.Email)

		c.Next()
	}
}
