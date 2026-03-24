package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/JokerTrickster/video-upload-backend/internal/config"
	"github.com/JokerTrickster/video-upload-backend/internal/handler"
	"github.com/JokerTrickster/video-upload-backend/internal/pkg/database"
	"github.com/JokerTrickster/video-upload-backend/internal/pkg/logger"
	redisUtil "github.com/JokerTrickster/video-upload-backend/internal/pkg/redis"
	"github.com/JokerTrickster/video-upload-backend/internal/repository"
	"github.com/JokerTrickster/video-upload-backend/internal/router"
	"github.com/JokerTrickster/video-upload-backend/internal/service"
)

func main() {
	// Load configuration first
	cfg, err := config.Load()
	if err != nil {
		fmt.Printf("Failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	if err := cfg.Validate(); err != nil {
		fmt.Printf("Invalid configuration: %v\n", err)
		os.Exit(1)
	}

	// Initialize logger with config
	logger.Init(cfg)
	logger.Info("Starting video-upload-backend service")
	logger.Info("Configuration loaded successfully",
		"environment", cfg.Server.Env,
		"port", cfg.Server.Port,
	)

	// Initialize database
	db, err := database.NewPostgresDB(cfg)
	if err != nil {
		logger.Fatal("Failed to connect to database", "error", err)
	}
	defer database.CloseDB(db)

	// Run migrations
	if err := database.AutoMigrate(db); err != nil {
		logger.Fatal("Failed to run migrations", "error", err)
	}
	logger.Info("Database migrations completed successfully")

	// Initialize Redis
	redisClient, err := redisUtil.NewRedisClient(cfg)
	if err != nil {
		logger.Fatal("Failed to connect to Redis", "error", err)
	}
	defer redisClient.Close()
	logger.Info("Redis connection established")

	// Initialize repositories
	userRepo := repository.NewUserRepository(db)
	tokenRepo := repository.NewTokenRepository(db)
	logger.Info("Repositories initialized")

	// Initialize services
	tokenService := service.NewTokenService(redisClient, cfg)
	youtubeService := service.NewYouTubeService()
	authService := service.NewAuthService(userRepo, tokenRepo, tokenService, youtubeService, cfg)
	logger.Info("Services initialized")

	// Initialize handlers
	authHandler := handler.NewAuthHandler(authService)
	logger.Info("Handlers initialized")

	// Setup router
	r := router.SetupRouter(cfg, authHandler, authService, redisClient)
	logger.Info("Router configured")

	// Create HTTP server
	srv := &http.Server{
		Addr:           fmt.Sprintf(":%s", cfg.Server.Port),
		Handler:        r,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		IdleTimeout:    120 * time.Second,
		MaxHeaderBytes: 1 << 20, // 1 MB
	}

	// Start server in a goroutine
	go func() {
		logger.Info("Starting HTTP server", "address", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Failed to start server", "error", err)
		}
	}()

	logger.Info("Video upload backend service started successfully",
		"port", cfg.Server.Port,
		"environment", cfg.Server.Env,
	)

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")

	// Graceful shutdown with 5 second timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Fatal("Server forced to shutdown", "error", err)
	}

	logger.Info("Server exited successfully")
}
