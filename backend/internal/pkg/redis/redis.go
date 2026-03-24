package redis

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/redis/go-redis/v9"

	"github.com/JokerTrickster/video-upload-backend/internal/config"
)

// NewRedisClient creates a new Redis client
func NewRedisClient(cfg *config.Config) (*redis.Client, error) {
	// Parse Redis URL
	opt, err := redis.ParseURL(cfg.Redis.URL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse Redis URL: %w", err)
	}

	// Override with config if password is set
	if cfg.Redis.Password != "" {
		opt.Password = cfg.Redis.Password
	}
	if cfg.Redis.DB != 0 {
		opt.DB = cfg.Redis.DB
	}

	// Create client
	client := redis.NewClient(opt)

	// Ping to verify connection
	ctx := context.Background()
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	log.Println("✅ Redis connection established")

	return client, nil
}

// CloseRedis closes the Redis connection
func CloseRedis(client *redis.Client) error {
	if err := client.Close(); err != nil {
		return fmt.Errorf("failed to close Redis: %w", err)
	}

	log.Println("✅ Redis connection closed")
	return nil
}

// Helper functions for common Redis operations

// SetWithExpiry sets a key-value pair with expiration
func SetWithExpiry(ctx context.Context, client *redis.Client, key string, value interface{}, expiry int64) error {
	return client.Set(ctx, key, value, 0).Err()
}

// Get retrieves a value by key
func Get(ctx context.Context, client *redis.Client, key string) (string, error) {
	val, err := client.Get(ctx, key).Result()
	if err == redis.Nil {
		return "", nil
	}
	return val, err
}

// Exists checks if a key exists
func Exists(ctx context.Context, client *redis.Client, key string) (bool, error) {
	count, err := client.Exists(ctx, key).Result()
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// Delete deletes a key
func Delete(ctx context.Context, client *redis.Client, keys ...string) error {
	return client.Del(ctx, keys...).Err()
}

// Keys returns all keys matching pattern
func Keys(ctx context.Context, client *redis.Client, pattern string) ([]string, error) {
	return client.Keys(ctx, pattern).Result()
}

// BuildKey builds a Redis key with namespace
func BuildKey(parts ...string) string {
	return strings.Join(parts, ":")
}
