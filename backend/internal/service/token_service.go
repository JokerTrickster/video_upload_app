package service

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/JokerTrickster/video-upload-backend/internal/config"
	"github.com/JokerTrickster/video-upload-backend/internal/domain"
	redisUtil "github.com/JokerTrickster/video-upload-backend/internal/pkg/redis"
)

// TokenService defines token management operations
type TokenService interface {
	// EncryptToken encrypts token using AES-256-GCM
	EncryptToken(ctx context.Context, plainText string) (string, error)

	// DecryptToken decrypts token using AES-256-GCM
	DecryptToken(ctx context.Context, cipherText string) (string, error)

	// AddToBlacklist adds token to blacklist in Redis
	AddToBlacklist(ctx context.Context, token string, expirySeconds int64) error

	// IsBlacklisted checks if token is blacklisted
	IsBlacklisted(ctx context.Context, token string) (bool, error)

	// SaveOAuthState saves OAuth state to Redis with TTL
	SaveOAuthState(ctx context.Context, state string, ttlSeconds int64) error

	// ValidateOAuthState validates and removes OAuth state from Redis
	ValidateOAuthState(ctx context.Context, state string) (bool, error)
}

// tokenService implements TokenService interface
type tokenService struct {
	redisClient   *redis.Client
	encryptionKey []byte
}

// NewTokenService creates a new token service instance
func NewTokenService(redisClient *redis.Client, cfg *config.Config) TokenService {
	return &tokenService{
		redisClient:   redisClient,
		encryptionKey: []byte(cfg.Security.EncryptionKey),
	}
}

// EncryptToken encrypts plaintext using AES-256-GCM
func (s *tokenService) EncryptToken(ctx context.Context, plainText string) (string, error) {
	if plainText == "" {
		return "", errors.New("plaintext cannot be empty")
	}

	// Create AES cipher block
	block, err := aes.NewCipher(s.encryptionKey)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}

	// Create GCM mode
	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %w", err)
	}

	// Create nonce
	nonce := make([]byte, aesGCM.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("failed to generate nonce: %w", err)
	}

	// Encrypt
	ciphertext := aesGCM.Seal(nonce, nonce, []byte(plainText), nil)

	// Encode to base64
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// DecryptToken decrypts ciphertext using AES-256-GCM
func (s *tokenService) DecryptToken(ctx context.Context, cipherText string) (string, error) {
	if cipherText == "" {
		return "", errors.New("ciphertext cannot be empty")
	}

	// Decode from base64
	data, err := base64.StdEncoding.DecodeString(cipherText)
	if err != nil {
		return "", fmt.Errorf("failed to decode base64: %w", err)
	}

	// Create AES cipher block
	block, err := aes.NewCipher(s.encryptionKey)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}

	// Create GCM mode
	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %w", err)
	}

	// Get nonce size
	nonceSize := aesGCM.NonceSize()
	if len(data) < nonceSize {
		return "", errors.New("ciphertext too short")
	}

	// Split nonce and ciphertext
	nonce, ciphertext := data[:nonceSize], data[nonceSize:]

	// Decrypt
	plaintext, err := aesGCM.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt: %w", err)
	}

	return string(plaintext), nil
}

// AddToBlacklist adds token to blacklist in Redis
func (s *tokenService) AddToBlacklist(ctx context.Context, token string, expirySeconds int64) error {
	if token == "" {
		return errors.New("token cannot be empty")
	}

	key := redisUtil.BuildKey("jwt", "blacklist", token)
	expiry := time.Duration(expirySeconds) * time.Second

	if err := s.redisClient.Set(ctx, key, "1", expiry).Err(); err != nil {
		return fmt.Errorf("failed to add token to blacklist: %w", err)
	}

	return nil
}

// IsBlacklisted checks if token is blacklisted
func (s *tokenService) IsBlacklisted(ctx context.Context, token string) (bool, error) {
	if token == "" {
		return false, errors.New("token cannot be empty")
	}

	key := redisUtil.BuildKey("jwt", "blacklist", token)

	exists, err := redisUtil.Exists(ctx, s.redisClient, key)
	if err != nil {
		return false, fmt.Errorf("failed to check blacklist: %w", err)
	}

	return exists, nil
}

// SaveOAuthState saves OAuth state to Redis with TTL
func (s *tokenService) SaveOAuthState(ctx context.Context, state string, ttlSeconds int64) error {
	if state == "" {
		return errors.New("state cannot be empty")
	}

	key := redisUtil.BuildKey("oauth", "state", state)
	expiry := time.Duration(ttlSeconds) * time.Second

	if err := s.redisClient.Set(ctx, key, "1", expiry).Err(); err != nil {
		return fmt.Errorf("failed to save OAuth state: %w", err)
	}

	return nil
}

// ValidateOAuthState validates and removes OAuth state from Redis
func (s *tokenService) ValidateOAuthState(ctx context.Context, state string) (bool, error) {
	if state == "" {
		return false, domain.ErrInvalidState
	}

	key := redisUtil.BuildKey("oauth", "state", state)

	// Check if state exists
	exists, err := redisUtil.Exists(ctx, s.redisClient, key)
	if err != nil {
		return false, fmt.Errorf("failed to validate state: %w", err)
	}

	if !exists {
		return false, nil
	}

	// Delete state after validation (one-time use)
	if err := redisUtil.Delete(ctx, s.redisClient, key); err != nil {
		return false, fmt.Errorf("failed to delete state: %w", err)
	}

	return true, nil
}
