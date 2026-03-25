package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/JokerTrickster/video-upload-backend/internal/config"
	"github.com/JokerTrickster/video-upload-backend/internal/pkg/logger"
)

func setupLogger() {
	cfg := &config.Config{
		Server: config.ServerConfig{LogLevel: "error"},
	}
	logger.Init(cfg)
}

func TestErrorHandlerMiddleware_NoPanic(t *testing.T) {
	setupLogger()

	router := gin.New()
	router.Use(ErrorHandlerMiddleware())
	router.GET("/ok", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	req := httptest.NewRequest(http.MethodGet, "/ok", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestErrorHandlerMiddleware_RecoversPanic(t *testing.T) {
	setupLogger()

	router := gin.New()
	router.Use(ErrorHandlerMiddleware())
	router.GET("/panic", func(c *gin.Context) {
		panic("something broke")
	})

	req := httptest.NewRequest(http.MethodGet, "/panic", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, "internal_server_error", resp["error"])
}

func TestRequestLoggerMiddleware_NoPanic(t *testing.T) {
	setupLogger()

	router := gin.New()
	router.Use(RequestLoggerMiddleware())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestDefaultRateLimiterConfig(t *testing.T) {
	cfg := DefaultRateLimiterConfig()

	assert.NotNil(t, cfg)
	assert.Equal(t, 60, cfg.RequestsPerMinute)
	assert.Equal(t, 10, cfg.BurstSize)
	assert.Greater(t, cfg.WindowSize.Seconds(), float64(0))
}
