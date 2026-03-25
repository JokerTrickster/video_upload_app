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
