package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

// Config holds all application configuration
type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Redis    RedisConfig
	Google   GoogleConfig
	JWT      JWTConfig
	Security SecurityConfig
}

// ServerConfig holds server-related configuration
type ServerConfig struct {
	Port           string
	Env            string
	LogLevel       string
	AllowedOrigins []string
}

// DatabaseConfig holds database connection configuration
type DatabaseConfig struct {
	URL             string
	MaxIdleConns    int
	MaxOpenConns    int
	ConnMaxLifetime time.Duration
}

// RedisConfig holds Redis connection configuration
type RedisConfig struct {
	URL      string
	Password string
	DB       int
}

// GoogleConfig holds Google OAuth configuration
type GoogleConfig struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
}

// JWTConfig holds JWT token configuration
type JWTConfig struct {
	Secret              string
	AccessTokenExpiry   time.Duration
	RefreshTokenExpiry  time.Duration
}

// SecurityConfig holds security-related configuration
type SecurityConfig struct {
	EncryptionKey      string
	RateLimitAuth      int
	RateLimitGeneral   int
}

// Load loads configuration from environment variables
func Load() (*Config, error) {
	// Load .env file if it exists
	_ = godotenv.Load()

	config := &Config{
		Server: ServerConfig{
			Port:           getEnv("PORT", "8080"),
			Env:            getEnv("ENV", "development"),
			LogLevel:       getEnv("LOG_LEVEL", "info"),
			AllowedOrigins: []string{
				getEnv("ALLOWED_ORIGINS", "http://localhost:3000"),
			},
		},
		Database: DatabaseConfig{
			URL:             getEnv("DATABASE_URL", ""),
			MaxIdleConns:    getEnvInt("DB_MAX_IDLE_CONNS", 10),
			MaxOpenConns:    getEnvInt("DB_MAX_OPEN_CONNS", 100),
			ConnMaxLifetime: getEnvDuration("DB_CONN_MAX_LIFETIME", time.Hour),
		},
		Redis: RedisConfig{
			URL:      getEnv("REDIS_URL", "redis://localhost:6379/0"),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       getEnvInt("REDIS_DB", 0),
		},
		Google: GoogleConfig{
			ClientID:     getEnv("GOOGLE_CLIENT_ID", ""),
			ClientSecret: getEnv("GOOGLE_CLIENT_SECRET", ""),
			RedirectURL:  getEnv("GOOGLE_REDIRECT_URL", ""),
		},
		JWT: JWTConfig{
			Secret:             getEnv("JWT_SECRET", ""),
			AccessTokenExpiry:  getEnvDuration("JWT_ACCESS_TOKEN_EXPIRY", 15*time.Minute),
			RefreshTokenExpiry: getEnvDuration("JWT_REFRESH_TOKEN_EXPIRY", 168*time.Hour),
		},
		Security: SecurityConfig{
			EncryptionKey:    getEnv("ENCRYPTION_KEY", ""),
			RateLimitAuth:    getEnvInt("RATE_LIMIT_AUTH", 10),
			RateLimitGeneral: getEnvInt("RATE_LIMIT_GENERAL", 60),
		},
	}

	// Validate required configuration
	if err := config.Validate(); err != nil {
		return nil, err
	}

	return config, nil
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if c.Database.URL == "" {
		return fmt.Errorf("DATABASE_URL is required")
	}
	if c.Google.ClientID == "" {
		return fmt.Errorf("GOOGLE_CLIENT_ID is required")
	}
	if c.Google.ClientSecret == "" {
		return fmt.Errorf("GOOGLE_CLIENT_SECRET is required")
	}
	if c.Google.RedirectURL == "" {
		return fmt.Errorf("GOOGLE_REDIRECT_URL is required")
	}
	if c.JWT.Secret == "" || len(c.JWT.Secret) < 32 {
		return fmt.Errorf("JWT_SECRET must be at least 32 characters")
	}
	if c.Security.EncryptionKey == "" || len(c.Security.EncryptionKey) != 32 {
		return fmt.Errorf("ENCRYPTION_KEY must be exactly 32 characters")
	}
	return nil
}

// Helper functions
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}
