package youtube

import (
	"bytes"
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewClient(t *testing.T) {
	client := NewClient()
	assert.NotNil(t, client)
}

func TestUploadVideo_ValidationErrors(t *testing.T) {
	client := NewClient()
	ctx := context.Background()

	tests := []struct {
		name        string
		accessToken string
		req         *UploadVideoRequest
		wantErr     string
	}{
		{
			name:        "empty access token",
			accessToken: "",
			req: &UploadVideoRequest{
				FilePath: "/tmp/test.mp4",
			},
			wantErr: "access token is required",
		},
		{
			name:        "empty file path",
			accessToken: "test-token",
			req: &UploadVideoRequest{
				FilePath: "",
			},
			wantErr: "file path is required",
		},
		{
			name:        "file does not exist",
			accessToken: "test-token",
			req: &UploadVideoRequest{
				FilePath: "/nonexistent/file.mp4",
			},
			wantErr: "failed to stat file",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := client.UploadVideo(ctx, tt.accessToken, tt.req)
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.wantErr)
		})
	}
}

func TestUploadVideo_FileSizeLimit(t *testing.T) {
	client := NewClient()
	ctx := context.Background()

	// Create a temporary file
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.mp4")

	// Create empty file (0 bytes, under the limit)
	file, err := os.Create(testFile)
	require.NoError(t, err)
	defer file.Close()

	req := &UploadVideoRequest{
		FilePath: testFile,
	}

	// This will fail at the YouTube API call level (invalid token)
	// but should pass the file size validation since file is 0 bytes
	_, err = client.UploadVideo(ctx, "test-token", req)
	assert.Error(t, err)
	// Should NOT be a file size error
	assert.NotContains(t, err.Error(), "file size exceeds maximum")
}

func TestUploadVideo_DefaultValues(t *testing.T) {
	// Verify default values are set correctly by testing through validation
	// We can't test actual upload without YouTube API, but we verify the request flows through
	client := NewClient()
	ctx := context.Background()

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "my_video.mp4")
	err := os.WriteFile(testFile, []byte("fake video content"), 0644)
	require.NoError(t, err)

	req := &UploadVideoRequest{
		FilePath: testFile,
		// Title, Description, PrivacyStatus left empty to test defaults
	}

	// Will fail at YouTube API call but that's expected
	_, err = client.UploadVideo(ctx, "test-token", req)
	assert.Error(t, err)
	// Should reach YouTube API call (not validation error)
	assert.Contains(t, err.Error(), "failed to")
}

func TestUploadVideo_WithMetadata(t *testing.T) {
	client := NewClient()
	ctx := context.Background()

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.mp4")
	err := os.WriteFile(testFile, []byte("fake video"), 0644)
	require.NoError(t, err)

	req := &UploadVideoRequest{
		FilePath:      testFile,
		Title:         "Custom Title",
		Description:   "Custom Description",
		PrivacyStatus: "unlisted",
	}

	// Will fail at YouTube API call but should pass validation
	_, err = client.UploadVideo(ctx, "test-token", req)
	assert.Error(t, err)
	assert.NotContains(t, err.Error(), "access token is required")
	assert.NotContains(t, err.Error(), "file path is required")
	assert.NotContains(t, err.Error(), "failed to stat file")
}

func TestGetVideoStatus_ValidationErrors(t *testing.T) {
	client := NewClient()
	ctx := context.Background()

	tests := []struct {
		name        string
		accessToken string
		videoID     string
		wantErr     string
	}{
		{
			name:        "empty access token",
			accessToken: "",
			videoID:     "test-video-id",
			wantErr:     "access token is required",
		},
		{
			name:        "empty video ID",
			accessToken: "test-token",
			videoID:     "",
			wantErr:     "video ID is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := client.GetVideoStatus(ctx, tt.accessToken, tt.videoID)
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.wantErr)
		})
	}
}

func TestGetVideoStatus_InvalidToken(t *testing.T) {
	// Test that invalid token fails at API level (not validation)
	client := NewClient()
	ctx := context.Background()

	_, err := client.GetVideoStatus(ctx, "invalid-token", "test-video-id")
	require.Error(t, err)
	// Should reach YouTube API call and fail with API error
	assert.NotContains(t, err.Error(), "access token is required")
	assert.NotContains(t, err.Error(), "video ID is required")
}

func TestDeleteVideo_ValidationErrors(t *testing.T) {
	client := NewClient()
	ctx := context.Background()

	tests := []struct {
		name        string
		accessToken string
		videoID     string
		wantErr     string
	}{
		{
			name:        "empty access token",
			accessToken: "",
			videoID:     "test-video-id",
			wantErr:     "access token is required",
		},
		{
			name:        "empty video ID",
			accessToken: "test-token",
			videoID:     "",
			wantErr:     "video ID is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := client.DeleteVideo(ctx, tt.accessToken, tt.videoID)
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.wantErr)
		})
	}
}

func TestDeleteVideo_InvalidToken(t *testing.T) {
	client := NewClient()
	ctx := context.Background()

	err := client.DeleteVideo(ctx, "invalid-token", "test-video-id")
	require.Error(t, err)
	assert.NotContains(t, err.Error(), "access token is required")
	assert.NotContains(t, err.Error(), "video ID is required")
}

func TestProgressReader(t *testing.T) {
	data := []byte("test data for progress tracking")
	reader := &progressReader{
		reader: io.NopCloser(newBytesReader(data)),
		onProgress: func(n int64) {
			assert.Greater(t, n, int64(0))
		},
	}

	buf := make([]byte, len(data))
	n, err := reader.Read(buf)
	assert.NoError(t, err)
	assert.Equal(t, len(data), n)
	assert.Equal(t, data, buf)
}

func TestProgressReader_TracksTotalBytes(t *testing.T) {
	data := []byte("hello world from progress reader test")
	var totalRead int64

	pr := &progressReader{
		reader: bytes.NewReader(data),
		onProgress: func(n int64) {
			totalRead += n
		},
	}

	// Read in small chunks
	buf := make([]byte, 5)
	for {
		_, err := pr.Read(buf)
		if err == io.EOF {
			break
		}
		require.NoError(t, err)
	}

	assert.Equal(t, int64(len(data)), totalRead,
		"total bytes tracked should equal input data length")
}

func TestProgressReader_NilCallback(t *testing.T) {
	data := []byte("test data without callback")
	pr := &progressReader{
		reader:     bytes.NewReader(data),
		onProgress: nil,
	}

	buf := make([]byte, len(data))
	n, err := pr.Read(buf)
	assert.NoError(t, err)
	assert.Equal(t, len(data), n)
	assert.Equal(t, data, buf)
}

func TestProgressReader_EOF(t *testing.T) {
	data := []byte("short")
	callCount := 0
	pr := &progressReader{
		reader: bytes.NewReader(data),
		onProgress: func(n int64) {
			callCount++
		},
	}

	// Read all data
	buf := make([]byte, 100)
	n, err := pr.Read(buf)
	assert.NoError(t, err)
	assert.Equal(t, len(data), n)
	assert.Equal(t, 1, callCount)

	// Next read should return EOF
	n, err = pr.Read(buf)
	assert.Equal(t, io.EOF, err)
	assert.Equal(t, 0, n)
	// onProgress should not be called for 0 bytes
	assert.Equal(t, 1, callCount)
}

func TestProgressReader_EmptyReader(t *testing.T) {
	pr := &progressReader{
		reader: strings.NewReader(""),
		onProgress: func(n int64) {
			t.Error("should not be called for empty reader")
		},
	}

	buf := make([]byte, 10)
	n, err := pr.Read(buf)
	assert.Equal(t, io.EOF, err)
	assert.Equal(t, 0, n)
}

func TestConstants(t *testing.T) {
	assert.Equal(t, int64(2*1024*1024*1024), int64(MaxFileSize), "MaxFileSize should be 2GB")
	assert.Equal(t, 10*1024*1024, ChunkSize, "ChunkSize should be 10MB")
}

func TestUploadVideoRequest_Fields(t *testing.T) {
	var progressCalled bool
	req := &UploadVideoRequest{
		FilePath:      "/path/to/video.mp4",
		Title:         "My Video",
		Description:   "A test video",
		PrivacyStatus: "public",
		OnProgress: func(uploaded, total int64) {
			progressCalled = true
		},
	}

	assert.Equal(t, "/path/to/video.mp4", req.FilePath)
	assert.Equal(t, "My Video", req.Title)
	assert.Equal(t, "A test video", req.Description)
	assert.Equal(t, "public", req.PrivacyStatus)
	assert.NotNil(t, req.OnProgress)

	req.OnProgress(100, 200)
	assert.True(t, progressCalled)
}

func TestUploadVideoResponse_Fields(t *testing.T) {
	resp := &UploadVideoResponse{
		VideoID:       "abc123",
		Title:         "Test Video",
		ThumbnailURL:  "https://img.youtube.com/vi/abc123/default.jpg",
		UploadedBytes: 1024 * 1024,
	}

	assert.Equal(t, "abc123", resp.VideoID)
	assert.Equal(t, "Test Video", resp.Title)
	assert.Equal(t, "https://img.youtube.com/vi/abc123/default.jpg", resp.ThumbnailURL)
	assert.Equal(t, int64(1024*1024), resp.UploadedBytes)
}

func TestVideoStatus_Fields(t *testing.T) {
	tests := []struct {
		name   string
		status VideoStatus
	}{
		{
			name: "processed and playable",
			status: VideoStatus{
				VideoID:       "v1",
				Status:        "processed",
				UploadStatus:  "processed",
				PrivacyStatus: "private",
				Playable:      true,
			},
		},
		{
			name: "uploaded and playable",
			status: VideoStatus{
				VideoID:       "v2",
				Status:        "uploaded",
				UploadStatus:  "uploaded",
				PrivacyStatus: "unlisted",
				Playable:      true,
			},
		},
		{
			name: "failed with reason",
			status: VideoStatus{
				VideoID:       "v3",
				Status:        "failed",
				UploadStatus:  "failed",
				PrivacyStatus: "private",
				Playable:      false,
				FailureReason: "codec not supported",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.NotEmpty(t, tt.status.VideoID)
			assert.NotEmpty(t, tt.status.Status)
			if tt.status.FailureReason != "" {
				assert.False(t, tt.status.Playable)
			}
		})
	}
}

func TestUploadVideo_ContextCancellation(t *testing.T) {
	client := NewClient()
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.mp4")
	err := os.WriteFile(testFile, []byte("fake"), 0644)
	require.NoError(t, err)

	req := &UploadVideoRequest{
		FilePath: testFile,
	}

	_, err = client.UploadVideo(ctx, "test-token", req)
	assert.Error(t, err, "should fail with cancelled context")
}

func TestGetVideoStatus_ContextCancellation(t *testing.T) {
	client := NewClient()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := client.GetVideoStatus(ctx, "test-token", "video-id")
	assert.Error(t, err, "should fail with cancelled context")
}

func TestDeleteVideo_ContextCancellation(t *testing.T) {
	client := NewClient()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := client.DeleteVideo(ctx, "test-token", "video-id")
	assert.Error(t, err, "should fail with cancelled context")
}

// bytesReader wraps []byte to implement io.Reader
type bytesReader struct {
	data []byte
	pos  int
}

func newBytesReader(data []byte) *bytesReader {
	return &bytesReader{data: data}
}

func (r *bytesReader) Read(p []byte) (int, error) {
	if r.pos >= len(r.data) {
		return 0, io.EOF
	}
	n := copy(p, r.data[r.pos:])
	r.pos += n
	return n, nil
}

func (r *bytesReader) Close() error {
	return nil
}
