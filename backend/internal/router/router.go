package router

import (
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"

	"github.com/JokerTrickster/video-upload-backend/internal/config"
	"github.com/JokerTrickster/video-upload-backend/internal/handler"
	"github.com/JokerTrickster/video-upload-backend/internal/middleware"
	"github.com/JokerTrickster/video-upload-backend/internal/service"
)

// SetupRouter configures and returns the Gin router with all routes and middleware
func SetupRouter(
	cfg *config.Config,
	authHandler *handler.AuthHandler,
	authService service.AuthService,
	redisClient *redis.Client,
) *gin.Engine {
	// Set Gin mode
	if cfg.Server.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}

	// Create router
	router := gin.New()

	// Global middleware
	router.Use(gin.Recovery()) // Panic recovery
	router.Use(middleware.ErrorHandlerMiddleware())
	router.Use(middleware.RequestLoggerMiddleware())
	router.Use(middleware.CORSMiddleware(middleware.DefaultCORSConfig()))

	// Health check endpoint (no auth required)
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "healthy",
			"service": "video-upload-backend",
		})
	})

	// API v1 routes
	v1 := router.Group("/api/v1")
	{
		// Rate limiter for all API endpoints
		rateLimiterConfig := &middleware.RateLimiterConfig{
			RequestsPerMinute: 60,
			BurstSize:         10,
		}
		v1.Use(middleware.RateLimiterMiddleware(redisClient, rateLimiterConfig))

		// Auth routes (no authentication required)
		auth := v1.Group("/auth")
		{
			auth.GET("/google/url", authHandler.GetGoogleAuthURL)
			auth.POST("/google/callback", authHandler.HandleGoogleCallback)
			auth.POST("/refresh", authHandler.RefreshToken)
		}

		// Protected routes (authentication required)
		protected := v1.Group("")
		protected.Use(middleware.AuthMiddleware(authService))
		{
			// User routes
			protected.GET("/auth/me", authHandler.GetCurrentUser)
			protected.POST("/auth/logout", authHandler.Logout)

			// Future routes will be added here:
			// - Video upload routes
			// - Video management routes
			// - Channel management routes
		}
	}

	return router
}
