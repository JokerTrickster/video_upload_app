package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewYouTubeService(t *testing.T) {
	svc := NewYouTubeService()
	assert.NotNil(t, svc)
}

func TestYouTubeService_GetChannelInfo_EmptyToken(t *testing.T) {
	svc := NewYouTubeService()
	ctx := context.Background()

	_, err := svc.GetChannelInfo(ctx, "")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "access token is required")
}

func TestYouTubeService_GetUserProfile_EmptyToken(t *testing.T) {
	svc := NewYouTubeService()
	ctx := context.Background()

	_, err := svc.GetUserProfile(ctx, "")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "access token is required")
}

func TestYouTubeService_GetChannelInfo_InvalidToken(t *testing.T) {
	svc := NewYouTubeService()
	ctx := context.Background()

	_, err := svc.GetChannelInfo(ctx, "invalid-token-value")
	require.Error(t, err)
	// Should reach API call and fail (not a validation error)
	assert.NotContains(t, err.Error(), "access token is required")
}

func TestYouTubeService_GetUserProfile_InvalidToken(t *testing.T) {
	svc := NewYouTubeService()
	ctx := context.Background()

	_, err := svc.GetUserProfile(ctx, "invalid-token-value")
	require.Error(t, err)
	assert.NotContains(t, err.Error(), "access token is required")
}

func TestYouTubeService_GetChannelInfo_CancelledContext(t *testing.T) {
	svc := NewYouTubeService()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := svc.GetChannelInfo(ctx, "test-token")
	require.Error(t, err)
}

func TestYouTubeService_GetUserProfile_CancelledContext(t *testing.T) {
	svc := NewYouTubeService()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := svc.GetUserProfile(ctx, "test-token")
	require.Error(t, err)
}

func TestChannelInfo_Fields(t *testing.T) {
	info := &ChannelInfo{
		ChannelID:   "UC123",
		ChannelName: "Test Channel",
		Thumbnail:   "https://example.com/thumb.jpg",
	}

	assert.Equal(t, "UC123", info.ChannelID)
	assert.Equal(t, "Test Channel", info.ChannelName)
	assert.Equal(t, "https://example.com/thumb.jpg", info.Thumbnail)
}

func TestUserProfile_Fields(t *testing.T) {
	profile := &UserProfile{
		GoogleID: "google-123",
		Email:    "test@example.com",
		Name:     "Test User",
		Picture:  "https://example.com/pic.jpg",
	}

	assert.Equal(t, "google-123", profile.GoogleID)
	assert.Equal(t, "test@example.com", profile.Email)
	assert.Equal(t, "Test User", profile.Name)
	assert.Equal(t, "https://example.com/pic.jpg", profile.Picture)
}
