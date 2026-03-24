package service

import (
	"context"
	"fmt"

	"golang.org/x/oauth2"
	oauth2api "google.golang.org/api/oauth2/v2"
	"google.golang.org/api/option"
	"google.golang.org/api/youtube/v3"
)

// YouTubeService defines YouTube API operations
type YouTubeService interface {
	// GetChannelInfo retrieves YouTube channel information
	GetChannelInfo(ctx context.Context, accessToken string) (*ChannelInfo, error)

	// GetUserProfile retrieves user profile from Google
	GetUserProfile(ctx context.Context, accessToken string) (*UserProfile, error)
}

// ChannelInfo represents YouTube channel information
type ChannelInfo struct {
	ChannelID   string `json:"channel_id"`
	ChannelName string `json:"channel_name"`
	Thumbnail   string `json:"thumbnail"`
}

// UserProfile represents Google user profile
type UserProfile struct {
	GoogleID string `json:"google_id"`
	Email    string `json:"email"`
	Name     string `json:"name"`
	Picture  string `json:"picture"`
}

// youtubeService implements YouTubeService interface
type youtubeService struct{}

// NewYouTubeService creates a new YouTube service instance
func NewYouTubeService() YouTubeService {
	return &youtubeService{}
}

// GetChannelInfo retrieves YouTube channel information
func (s *youtubeService) GetChannelInfo(ctx context.Context, accessToken string) (*ChannelInfo, error) {
	if accessToken == "" {
		return nil, fmt.Errorf("access token is required")
	}

	// Create OAuth2 token
	token := &oauth2.Token{
		AccessToken: accessToken,
		TokenType:   "Bearer",
	}

	// Create YouTube service
	youtubeService, err := youtube.NewService(ctx, option.WithTokenSource(oauth2.StaticTokenSource(token)))
	if err != nil {
		return nil, fmt.Errorf("failed to create YouTube service: %w", err)
	}

	// Get channels
	call := youtubeService.Channels.List([]string{"snippet"})
	call.Mine(true)

	response, err := call.Do()
	if err != nil {
		return nil, fmt.Errorf("failed to get channel info: %w", err)
	}

	if len(response.Items) == 0 {
		return nil, fmt.Errorf("no YouTube channel found for this account")
	}

	channel := response.Items[0]
	thumbnail := ""
	if channel.Snippet.Thumbnails != nil && channel.Snippet.Thumbnails.Default != nil {
		thumbnail = channel.Snippet.Thumbnails.Default.Url
	}

	return &ChannelInfo{
		ChannelID:   channel.Id,
		ChannelName: channel.Snippet.Title,
		Thumbnail:   thumbnail,
	}, nil
}

// GetUserProfile retrieves user profile from Google
func (s *youtubeService) GetUserProfile(ctx context.Context, accessToken string) (*UserProfile, error) {
	if accessToken == "" {
		return nil, fmt.Errorf("access token is required")
	}

	// Create OAuth2 token
	token := &oauth2.Token{
		AccessToken: accessToken,
		TokenType:   "Bearer",
	}

	// Create OAuth2 service
	oauth2Service, err := oauth2api.NewService(ctx, option.WithTokenSource(oauth2.StaticTokenSource(token)))
	if err != nil {
		return nil, fmt.Errorf("failed to create OAuth2 service: %w", err)
	}

	// Get user info
	userInfo, err := oauth2Service.Userinfo.Get().Do()
	if err != nil {
		return nil, fmt.Errorf("failed to get user info: %w", err)
	}

	return &UserProfile{
		GoogleID: userInfo.Id,
		Email:    userInfo.Email,
		Name:     userInfo.Name,
		Picture:  userInfo.Picture,
	}, nil
}
