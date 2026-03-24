package repository

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/JokerTrickster/video-upload-backend/internal/domain"
)

// tokenRepository implements TokenRepository interface
type tokenRepository struct {
	db *gorm.DB
}

// NewTokenRepository creates a new token repository instance
func NewTokenRepository(db *gorm.DB) TokenRepository {
	return &tokenRepository{db: db}
}

// Create creates a new token
func (r *tokenRepository) Create(ctx context.Context, token *domain.Token) error {
	if err := r.db.WithContext(ctx).Create(token).Error; err != nil {
		return err
	}
	return nil
}

// FindByUserID retrieves a token by user ID
func (r *tokenRepository) FindByUserID(ctx context.Context, userID string) (*domain.Token, error) {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, domain.ErrInvalidInput
	}

	var token domain.Token
	if err := r.db.WithContext(ctx).
		Where("user_id = ?", uid).
		First(&token).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrTokenNotFound
		}
		return nil, err
	}

	return &token, nil
}

// Update updates a token
func (r *tokenRepository) Update(ctx context.Context, token *domain.Token) error {
	token.UpdatedAt = time.Now()

	if err := r.db.WithContext(ctx).
		Model(token).
		Where("user_id = ?", token.UserID).
		Updates(map[string]interface{}{
			"encrypted_access_token":  token.EncryptedAccessToken,
			"encrypted_refresh_token": token.EncryptedRefreshToken,
			"token_type":              token.TokenType,
			"expires_at":              token.ExpiresAt,
			"updated_at":              token.UpdatedAt,
		}).Error; err != nil {
		return err
	}

	return nil
}

// Delete deletes a token
func (r *tokenRepository) Delete(ctx context.Context, userID string) error {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return domain.ErrInvalidInput
	}

	if err := r.db.WithContext(ctx).
		Where("user_id = ?", uid).
		Delete(&domain.Token{}).Error; err != nil {
		return err
	}

	return nil
}
