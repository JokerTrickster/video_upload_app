package domain

import (
	"time"

	"github.com/google/uuid"
)

// User represents a user in the system
type User struct {
	ID                 uuid.UUID  `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Email              string     `json:"email" gorm:"type:varchar(255);uniqueIndex;not null"`
	GoogleID           string     `json:"google_id" gorm:"type:varchar(255);uniqueIndex;not null"`
	YouTubeChannelID   *string    `json:"youtube_channel_id,omitempty" gorm:"type:varchar(255)"`
	YouTubeChannelName *string    `json:"youtube_channel_name,omitempty" gorm:"type:varchar(255)"`
	ProfileImageURL    *string    `json:"profile_image_url,omitempty" gorm:"type:text"`
	CreatedAt          time.Time  `json:"created_at" gorm:"not null;default:now()"`
	UpdatedAt          time.Time  `json:"updated_at" gorm:"not null;default:now()"`
	DeletedAt          *time.Time `json:"deleted_at,omitempty" gorm:"index"`
}

// TableName specifies the table name for GORM
func (User) TableName() string {
	return "users"
}

// BeforeCreate hook to set default values
func (u *User) BeforeCreate() error {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	now := time.Now()
	u.CreatedAt = now
	u.UpdatedAt = now
	return nil
}

// BeforeUpdate hook to update timestamp
func (u *User) BeforeUpdate() error {
	u.UpdatedAt = time.Now()
	return nil
}
