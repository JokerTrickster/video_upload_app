package domain

import (
	"time"

	"github.com/google/uuid"
)

// Token represents OAuth tokens for a user
type Token struct {
	ID                    uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID                uuid.UUID `json:"user_id" gorm:"type:uuid;not null;uniqueIndex"`
	EncryptedAccessToken  string    `json:"-" gorm:"type:text;not null"`
	EncryptedRefreshToken string    `json:"-" gorm:"type:text;not null"`
	TokenType             string    `json:"token_type" gorm:"type:varchar(50);not null;default:'Bearer'"`
	ExpiresAt             time.Time `json:"expires_at" gorm:"not null;index"`
	CreatedAt             time.Time `json:"created_at" gorm:"not null;default:now()"`
	UpdatedAt             time.Time `json:"updated_at" gorm:"not null;default:now()"`

	// Associations
	User User `json:"user" gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
}

// TableName specifies the table name for GORM
func (Token) TableName() string {
	return "user_tokens"
}

// BeforeCreate hook to set default values
func (t *Token) BeforeCreate() error {
	if t.ID == uuid.Nil {
		t.ID = uuid.New()
	}
	now := time.Now()
	t.CreatedAt = now
	t.UpdatedAt = now
	return nil
}

// BeforeUpdate hook to update timestamp
func (t *Token) BeforeUpdate() error {
	t.UpdatedAt = time.Now()
	return nil
}

// IsExpired checks if the OAuth token is expired
func (t *Token) IsExpired() bool {
	return time.Now().After(t.ExpiresAt)
}
