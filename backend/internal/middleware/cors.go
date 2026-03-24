package middleware

import (
	"github.com/gin-gonic/gin"
)

// CORSConfig defines CORS configuration
type CORSConfig struct {
	AllowOrigins     []string
	AllowMethods     []string
	AllowHeaders     []string
	ExposeHeaders    []string
	AllowCredentials bool
	MaxAge           int // in seconds
}

// DefaultCORSConfig returns default CORS settings
func DefaultCORSConfig() *CORSConfig {
	return &CORSConfig{
		AllowOrigins: []string{"http://localhost:3000", "http://localhost:5173"}, // React/Vite dev servers
		AllowMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders: []string{
			"Origin",
			"Content-Type",
			"Accept",
			"Authorization",
			"X-Request-ID",
		},
		ExposeHeaders: []string{
			"X-RateLimit-Limit",
			"X-RateLimit-Remaining",
			"X-RateLimit-Reset",
		},
		AllowCredentials: true,
		MaxAge:           86400, // 24 hours
	}
}

// CORSMiddleware creates CORS middleware
func CORSMiddleware(config *CORSConfig) gin.HandlerFunc {
	if config == nil {
		config = DefaultCORSConfig()
	}

	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")

		// Check if origin is allowed
		allowOrigin := ""
		for _, allowed := range config.AllowOrigins {
			if allowed == "*" || allowed == origin {
				allowOrigin = allowed
				break
			}
		}

		// Set CORS headers
		if allowOrigin != "" {
			c.Header("Access-Control-Allow-Origin", allowOrigin)
		}

		if config.AllowCredentials {
			c.Header("Access-Control-Allow-Credentials", "true")
		}

		// Handle preflight requests
		if c.Request.Method == "OPTIONS" {
			// Set allowed methods
			methods := ""
			for i, method := range config.AllowMethods {
				if i > 0 {
					methods += ", "
				}
				methods += method
			}
			c.Header("Access-Control-Allow-Methods", methods)

			// Set allowed headers
			headers := ""
			for i, header := range config.AllowHeaders {
				if i > 0 {
					headers += ", "
				}
				headers += header
			}
			c.Header("Access-Control-Allow-Headers", headers)

			// Set max age
			if config.MaxAge > 0 {
				c.Header("Access-Control-Max-Age", string(rune(config.MaxAge)))
			}

			c.AbortWithStatus(204)
			return
		}

		// Set exposed headers
		if len(config.ExposeHeaders) > 0 {
			exposeHeaders := ""
			for i, header := range config.ExposeHeaders {
				if i > 0 {
					exposeHeaders += ", "
				}
				exposeHeaders += header
			}
			c.Header("Access-Control-Expose-Headers", exposeHeaders)
		}

		c.Next()
	}
}
