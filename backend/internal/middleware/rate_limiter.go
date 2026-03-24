package middleware

import (
	"fmt"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"

	"github.com/JokerTrickster/video-upload-backend/internal/handler"
	redisUtil "github.com/JokerTrickster/video-upload-backend/internal/pkg/redis"
)

// RateLimiterConfig defines rate limiter configuration
type RateLimiterConfig struct {
	RequestsPerMinute int           // Maximum requests allowed per minute
	BurstSize         int           // Maximum burst size
	WindowSize        time.Duration // Time window for rate limiting
}

// DefaultRateLimiterConfig returns default rate limiter settings
func DefaultRateLimiterConfig() *RateLimiterConfig {
	return &RateLimiterConfig{
		RequestsPerMinute: 60,              // 60 requests per minute
		BurstSize:         10,              // Allow burst of 10 requests
		WindowSize:        1 * time.Minute, // 1 minute window
	}
}

// RateLimiterMiddleware creates rate limiter middleware with sliding window algorithm
func RateLimiterMiddleware(redisClient *redis.Client, config *RateLimiterConfig) gin.HandlerFunc {
	if config == nil {
		config = DefaultRateLimiterConfig()
	}

	return func(c *gin.Context) {
		// Get client IP address
		clientIP := c.ClientIP()

		// Create Redis key for this IP
		key := redisUtil.BuildKey("rate_limit", clientIP)

		// Get current request count
		val, err := redisClient.Get(c.Request.Context(), key).Result()
		if err != nil && err != redis.Nil {
			// Redis error - allow request but log error
			c.Next()
			return
		}

		var currentCount int
		if val != "" {
			currentCount, _ = strconv.Atoi(val)
		}

		// Check if rate limit exceeded
		if currentCount >= config.RequestsPerMinute {
			// Get TTL to inform client when they can retry
			ttl, err := redisClient.TTL(c.Request.Context(), key).Result()
			if err != nil {
				ttl = config.WindowSize
			}

			// Set Retry-After header
			c.Header("Retry-After", fmt.Sprintf("%d", int(ttl.Seconds())))
			c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", config.RequestsPerMinute))
			c.Header("X-RateLimit-Remaining", "0")
			c.Header("X-RateLimit-Reset", fmt.Sprintf("%d", time.Now().Add(ttl).Unix()))

			handler.RespondTooManyRequests(c, "Rate limit exceeded. Please try again later.")
			c.Abort()
			return
		}

		// Increment counter
		pipe := redisClient.Pipeline()
		pipe.Incr(c.Request.Context(), key)
		if currentCount == 0 {
			// Set expiry only on first request in window
			pipe.Expire(c.Request.Context(), key, config.WindowSize)
		}
		_, err = pipe.Exec(c.Request.Context())
		if err != nil {
			// Redis error - allow request but log error
			c.Next()
			return
		}

		// Set rate limit headers
		remaining := config.RequestsPerMinute - (currentCount + 1)
		if remaining < 0 {
			remaining = 0
		}

		ttl, _ := redisClient.TTL(c.Request.Context(), key).Result()
		resetTime := time.Now().Add(ttl)

		c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", config.RequestsPerMinute))
		c.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", remaining))
		c.Header("X-RateLimit-Reset", fmt.Sprintf("%d", resetTime.Unix()))

		c.Next()
	}
}
