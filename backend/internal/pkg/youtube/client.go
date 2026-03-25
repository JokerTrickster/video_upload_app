package youtube

import (
	"context"
	"fmt"
	"io"
	"os"

	"golang.org/x/oauth2"
	"google.golang.org/api/googleapi"
	"google.golang.org/api/option"
	"google.golang.org/api/youtube/v3"
)

const (
	// MaxFileSize is the maximum file size for upload (2GB)
	MaxFileSize = 2 * 1024 * 1024 * 1024

	// ChunkSize for resumable upload (10MB chunks)
	ChunkSize = 10 * 1024 * 1024
)

// Client wraps YouTube API operations
type Client interface {
	// UploadVideo uploads a video file to YouTube with resumable upload
	UploadVideo(ctx context.Context, accessToken string, req *UploadVideoRequest) (*UploadVideoResponse, error)

	// GetVideoStatus checks if a video is playable and retrieves its status
	GetVideoStatus(ctx context.Context, accessToken string, videoID string) (*VideoStatus, error)

	// DeleteVideo deletes a video from YouTube
	DeleteVideo(ctx context.Context, accessToken string, videoID string) error
}

// UploadVideoRequest represents video upload request
type UploadVideoRequest struct {
	FilePath    string
	Title       string
	Description string
	PrivacyStatus string // "private", "unlisted", "public"
	OnProgress  func(uploadedBytes, totalBytes int64) // Progress callback
}

// UploadVideoResponse represents video upload response
type UploadVideoResponse struct {
	VideoID       string
	Title         string
	ThumbnailURL  string
	UploadedBytes int64
}

// VideoStatus represents YouTube video status
type VideoStatus struct {
	VideoID        string
	Status         string // "uploaded", "processed", "failed"
	UploadStatus   string
	PrivacyStatus  string
	Playable       bool
	FailureReason  string
}

// client implements Client interface
type client struct{}

// NewClient creates a new YouTube client
func NewClient() Client {
	return &client{}
}

// UploadVideo uploads a video file to YouTube with resumable upload
func (c *client) UploadVideo(ctx context.Context, accessToken string, req *UploadVideoRequest) (*UploadVideoResponse, error) {
	if accessToken == "" {
		return nil, fmt.Errorf("access token is required")
	}

	if req.FilePath == "" {
		return nil, fmt.Errorf("file path is required")
	}

	// Validate file exists
	fileInfo, err := os.Stat(req.FilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to stat file: %w", err)
	}

	if fileInfo.Size() > MaxFileSize {
		return nil, fmt.Errorf("file size exceeds maximum allowed size of 2GB")
	}

	// Open file
	file, err := os.Open(req.FilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

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

	// Set default values
	title := req.Title
	if title == "" {
		title = fileInfo.Name()
	}

	description := req.Description
	if description == "" {
		description = "Uploaded via Media Backup System"
	}

	privacyStatus := req.PrivacyStatus
	if privacyStatus == "" {
		privacyStatus = "private"
	}

	// Create video object
	video := &youtube.Video{
		Snippet: &youtube.VideoSnippet{
			Title:       title,
			Description: description,
		},
		Status: &youtube.VideoStatus{
			PrivacyStatus: privacyStatus,
		},
	}

	// Create insert call
	call := youtubeService.Videos.Insert([]string{"snippet", "status"}, video)

	// Create progress reader
	totalBytes := fileInfo.Size()
	var uploadedBytes int64

	progressReader := &progressReader{
		reader: file,
		onProgress: func(n int64) {
			uploadedBytes += n
			if req.OnProgress != nil {
				req.OnProgress(uploadedBytes, totalBytes)
			}
		},
	}

	// Set media body with chunk size for resumable upload
	call = call.Media(progressReader, googleapi.ChunkSize(ChunkSize))

	// Execute upload
	uploadedVideo, err := call.Do()
	if err != nil {
		return nil, fmt.Errorf("failed to upload video: %w", err)
	}

	// Get thumbnail URL
	thumbnailURL := ""
	if uploadedVideo.Snippet != nil && uploadedVideo.Snippet.Thumbnails != nil {
		if uploadedVideo.Snippet.Thumbnails.Default != nil {
			thumbnailURL = uploadedVideo.Snippet.Thumbnails.Default.Url
		}
	}

	return &UploadVideoResponse{
		VideoID:       uploadedVideo.Id,
		Title:         uploadedVideo.Snippet.Title,
		ThumbnailURL:  thumbnailURL,
		UploadedBytes: uploadedBytes,
	}, nil
}

// GetVideoStatus checks if a video is playable and retrieves its status
func (c *client) GetVideoStatus(ctx context.Context, accessToken string, videoID string) (*VideoStatus, error) {
	if accessToken == "" {
		return nil, fmt.Errorf("access token is required")
	}

	if videoID == "" {
		return nil, fmt.Errorf("video ID is required")
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

	// Get video details
	call := youtubeService.Videos.List([]string{"status", "processingDetails"})
	call = call.Id(videoID)

	response, err := call.Do()
	if err != nil {
		return nil, fmt.Errorf("failed to get video status: %w", err)
	}

	if len(response.Items) == 0 {
		return nil, fmt.Errorf("video not found: %s", videoID)
	}

	video := response.Items[0]
	status := &VideoStatus{
		VideoID:       video.Id,
		PrivacyStatus: video.Status.PrivacyStatus,
	}

	// Check upload status
	if video.Status.UploadStatus != "" {
		status.UploadStatus = video.Status.UploadStatus
		status.Status = video.Status.UploadStatus
	}

	// Check if video is playable
	// Video is playable if upload status is "processed" or "uploaded"
	if video.Status.UploadStatus == "processed" || video.Status.UploadStatus == "uploaded" {
		status.Playable = true
	}

	// Check for failure
	if video.Status.FailureReason != "" {
		status.Status = "failed"
		status.FailureReason = video.Status.FailureReason
		status.Playable = false
	}

	// Check processing details if available
	if video.ProcessingDetails != nil {
		if video.ProcessingDetails.ProcessingStatus != "" {
			status.Status = video.ProcessingDetails.ProcessingStatus
		}
		if video.ProcessingDetails.ProcessingFailureReason != "" {
			status.FailureReason = video.ProcessingDetails.ProcessingFailureReason
			status.Playable = false
		}
	}

	return status, nil
}

// DeleteVideo deletes a video from YouTube
func (c *client) DeleteVideo(ctx context.Context, accessToken string, videoID string) error {
	if accessToken == "" {
		return fmt.Errorf("access token is required")
	}

	if videoID == "" {
		return fmt.Errorf("video ID is required")
	}

	// Create OAuth2 token
	token := &oauth2.Token{
		AccessToken: accessToken,
		TokenType:   "Bearer",
	}

	// Create YouTube service
	youtubeService, err := youtube.NewService(ctx, option.WithTokenSource(oauth2.StaticTokenSource(token)))
	if err != nil {
		return fmt.Errorf("failed to create YouTube service: %w", err)
	}

	// Delete video
	call := youtubeService.Videos.Delete(videoID)
	err = call.Do()
	if err != nil {
		return fmt.Errorf("failed to delete video: %w", err)
	}

	return nil
}

// progressReader wraps an io.Reader to track read progress
type progressReader struct {
	reader     io.Reader
	onProgress func(n int64)
}

// Read implements io.Reader interface
func (pr *progressReader) Read(p []byte) (int, error) {
	n, err := pr.reader.Read(p)
	if n > 0 && pr.onProgress != nil {
		pr.onProgress(int64(n))
	}
	return n, err
}
