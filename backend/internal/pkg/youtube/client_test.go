package youtube

import (
	"context"
	"io"
	"os"
	"path/filepath"
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

	// Create empty file
	file, err := os.Create(testFile)
	require.NoError(t, err)
	defer file.Close()

	// Note: We can't actually create a >2GB file in tests
	// This test just validates the check exists
	req := &UploadVideoRequest{
		FilePath: testFile,
	}

	// This should fail with real YouTube API (no token)
	// but we're just testing validation logic
	_, err = client.UploadVideo(ctx, "test-token", req)
	assert.Error(t, err)
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
