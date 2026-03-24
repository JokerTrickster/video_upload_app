package domain

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMediaAsset_TableName(t *testing.T) {
	asset := &MediaAsset{}
	assert.Equal(t, "media_assets", asset.TableName())
}

func TestMediaAsset_BeforeCreate(t *testing.T) {
	tests := []struct {
		name    string
		asset   *MediaAsset
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid video asset",
			asset: &MediaAsset{
				UserID:           uuid.New(),
				OriginalFilename: "video.mp4",
				FileSizeBytes:    1024000,
				MediaType:        MediaTypeVideo,
			},
			wantErr: false,
		},
		{
			name: "valid image asset",
			asset: &MediaAsset{
				UserID:           uuid.New(),
				OriginalFilename: "photo.jpg",
				FileSizeBytes:    512000,
				MediaType:        MediaTypeImage,
			},
			wantErr: false,
		},
		{
			name: "missing user_id",
			asset: &MediaAsset{
				OriginalFilename: "video.mp4",
				FileSizeBytes:    1024000,
				MediaType:        MediaTypeVideo,
			},
			wantErr: true,
			errMsg:  "user_id is required",
		},
		{
			name: "missing filename",
			asset: &MediaAsset{
				UserID:        uuid.New(),
				FileSizeBytes: 1024000,
				MediaType:     MediaTypeVideo,
			},
			wantErr: true,
			errMsg:  "original_filename is required",
		},
		{
			name: "invalid file size",
			asset: &MediaAsset{
				UserID:           uuid.New(),
				OriginalFilename: "video.mp4",
				FileSizeBytes:    0,
				MediaType:        MediaTypeVideo,
			},
			wantErr: true,
			errMsg:  "file_size_bytes must be greater than 0",
		},
		{
			name: "invalid media type",
			asset: &MediaAsset{
				UserID:           uuid.New(),
				OriginalFilename: "file.txt",
				FileSizeBytes:    1024,
				MediaType:        "DOCUMENT",
			},
			wantErr: true,
			errMsg:  "media_type must be VIDEO or IMAGE",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.asset.BeforeCreate()

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				require.NoError(t, err)
				assert.NotEqual(t, uuid.Nil, tt.asset.AssetID)
				assert.Equal(t, SyncStatusPending, tt.asset.SyncStatus)
				assert.False(t, tt.asset.CreatedAt.IsZero())
				assert.False(t, tt.asset.UpdatedAt.IsZero())
			}
		})
	}
}

func TestMediaAsset_Validate(t *testing.T) {
	tests := []struct {
		name    string
		asset   *MediaAsset
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid pending asset",
			asset: &MediaAsset{
				UserID:           uuid.New(),
				OriginalFilename: "video.mp4",
				FileSizeBytes:    1024000,
				MediaType:        MediaTypeVideo,
				SyncStatus:       SyncStatusPending,
			},
			wantErr: false,
		},
		{
			name: "completed video without youtube_video_id",
			asset: &MediaAsset{
				UserID:           uuid.New(),
				OriginalFilename: "video.mp4",
				FileSizeBytes:    1024000,
				MediaType:        MediaTypeVideo,
				SyncStatus:       SyncStatusCompleted,
			},
			wantErr: true,
			errMsg:  "youtube_video_id is required for completed video uploads",
		},
		{
			name: "completed image without s3_object_key",
			asset: &MediaAsset{
				UserID:           uuid.New(),
				OriginalFilename: "photo.jpg",
				FileSizeBytes:    512000,
				MediaType:        MediaTypeImage,
				SyncStatus:       SyncStatusCompleted,
			},
			wantErr: true,
			errMsg:  "s3_object_key is required for completed image uploads",
		},
		{
			name: "negative retry count",
			asset: &MediaAsset{
				UserID:           uuid.New(),
				OriginalFilename: "video.mp4",
				FileSizeBytes:    1024000,
				MediaType:        MediaTypeVideo,
				SyncStatus:       SyncStatusPending,
				RetryCount:       -1,
			},
			wantErr: true,
			errMsg:  "retry_count cannot be negative",
		},
		{
			name: "invalid sync status",
			asset: &MediaAsset{
				UserID:           uuid.New(),
				OriginalFilename: "video.mp4",
				FileSizeBytes:    1024000,
				MediaType:        MediaTypeVideo,
				SyncStatus:       "UNKNOWN",
			},
			wantErr: true,
			errMsg:  "invalid sync_status",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.asset.Validate()

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestMediaAsset_CanRetry(t *testing.T) {
	tests := []struct {
		name       string
		syncStatus SyncStatus
		retryCount int
		want       bool
	}{
		{
			name:       "failed with 0 retries",
			syncStatus: SyncStatusFailed,
			retryCount: 0,
			want:       true,
		},
		{
			name:       "failed with 4 retries",
			syncStatus: SyncStatusFailed,
			retryCount: 4,
			want:       true,
		},
		{
			name:       "failed with 5 retries (max)",
			syncStatus: SyncStatusFailed,
			retryCount: 5,
			want:       false,
		},
		{
			name:       "failed with 6 retries (exceeded)",
			syncStatus: SyncStatusFailed,
			retryCount: 6,
			want:       false,
		},
		{
			name:       "completed asset",
			syncStatus: SyncStatusCompleted,
			retryCount: 0,
			want:       false,
		},
		{
			name:       "pending asset",
			syncStatus: SyncStatusPending,
			retryCount: 0,
			want:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			asset := &MediaAsset{
				SyncStatus: tt.syncStatus,
				RetryCount: tt.retryCount,
			}
			assert.Equal(t, tt.want, asset.CanRetry())
		})
	}
}

func TestMediaAsset_MarkAsUploading(t *testing.T) {
	asset := &MediaAsset{
		UserID:           uuid.New(),
		OriginalFilename: "video.mp4",
		FileSizeBytes:    1024000,
		MediaType:        MediaTypeVideo,
		SyncStatus:       SyncStatusPending,
	}

	asset.MarkAsUploading()

	assert.Equal(t, SyncStatusUploading, asset.SyncStatus)
	assert.NotNil(t, asset.UploadStartedAt)
	assert.False(t, asset.UploadStartedAt.IsZero())

	// Mark as uploading again - should not change UploadStartedAt
	firstStartTime := *asset.UploadStartedAt
	time.Sleep(10 * time.Millisecond)
	asset.MarkAsUploading()
	assert.Equal(t, firstStartTime, *asset.UploadStartedAt)
}

func TestMediaAsset_MarkAsCompleted(t *testing.T) {
	tests := []struct {
		name      string
		mediaType MediaType
		storageID string
		wantErr   bool
		errMsg    string
	}{
		{
			name:      "video with youtube_video_id",
			mediaType: MediaTypeVideo,
			storageID: "dQw4w9WgXcQ",
			wantErr:   false,
		},
		{
			name:      "image with s3_object_key",
			mediaType: MediaTypeImage,
			storageID: "images/user123/photo.jpg",
			wantErr:   false,
		},
		{
			name:      "empty storage_id",
			mediaType: MediaTypeVideo,
			storageID: "",
			wantErr:   true,
			errMsg:    "storage_id is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			asset := &MediaAsset{
				UserID:           uuid.New(),
				OriginalFilename: "file.ext",
				FileSizeBytes:    1024000,
				MediaType:        tt.mediaType,
				SyncStatus:       SyncStatusUploading,
			}

			err := asset.MarkAsCompleted(tt.storageID)

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				require.NoError(t, err)
				assert.Equal(t, SyncStatusCompleted, asset.SyncStatus)
				assert.NotNil(t, asset.UploadCompletedAt)
				assert.False(t, asset.UploadCompletedAt.IsZero())
				assert.Nil(t, asset.ErrorMessage)

				if tt.mediaType == MediaTypeVideo {
					require.NotNil(t, asset.YouTubeVideoID)
					assert.Equal(t, tt.storageID, *asset.YouTubeVideoID)
				} else if tt.mediaType == MediaTypeImage {
					require.NotNil(t, asset.S3ObjectKey)
					assert.Equal(t, tt.storageID, *asset.S3ObjectKey)
				}
			}
		})
	}
}

func TestMediaAsset_MarkAsFailed(t *testing.T) {
	asset := &MediaAsset{
		UserID:           uuid.New(),
		OriginalFilename: "video.mp4",
		FileSizeBytes:    1024000,
		MediaType:        MediaTypeVideo,
		SyncStatus:       SyncStatusUploading,
		RetryCount:       0,
	}

	errMsg := "network timeout"
	asset.MarkAsFailed(errMsg)

	assert.Equal(t, SyncStatusFailed, asset.SyncStatus)
	require.NotNil(t, asset.ErrorMessage)
	assert.Equal(t, errMsg, *asset.ErrorMessage)
	assert.Equal(t, 1, asset.RetryCount)

	// Mark as failed again
	asset.MarkAsFailed("quota exceeded")
	assert.Equal(t, 2, asset.RetryCount)
}

func TestMediaAsset_StatusCheckers(t *testing.T) {
	tests := []struct {
		name       string
		syncStatus SyncStatus
		checks     map[string]bool
	}{
		{
			name:       "pending status",
			syncStatus: SyncStatusPending,
			checks: map[string]bool{
				"IsPending":   true,
				"IsUploading": false,
				"IsCompleted": false,
				"IsFailed":    false,
			},
		},
		{
			name:       "uploading status",
			syncStatus: SyncStatusUploading,
			checks: map[string]bool{
				"IsPending":   false,
				"IsUploading": true,
				"IsCompleted": false,
				"IsFailed":    false,
			},
		},
		{
			name:       "completed status",
			syncStatus: SyncStatusCompleted,
			checks: map[string]bool{
				"IsPending":   false,
				"IsUploading": false,
				"IsCompleted": true,
				"IsFailed":    false,
			},
		},
		{
			name:       "failed status",
			syncStatus: SyncStatusFailed,
			checks: map[string]bool{
				"IsPending":   false,
				"IsUploading": false,
				"IsCompleted": false,
				"IsFailed":    true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			asset := &MediaAsset{SyncStatus: tt.syncStatus}

			assert.Equal(t, tt.checks["IsPending"], asset.IsPending(), "IsPending mismatch")
			assert.Equal(t, tt.checks["IsUploading"], asset.IsUploading(), "IsUploading mismatch")
			assert.Equal(t, tt.checks["IsCompleted"], asset.IsCompleted(), "IsCompleted mismatch")
			assert.Equal(t, tt.checks["IsFailed"], asset.IsFailed(), "IsFailed mismatch")
		})
	}
}

func TestMediaAsset_GetStorageID(t *testing.T) {
	videoID := "dQw4w9WgXcQ"
	s3Key := "images/user123/photo.jpg"

	tests := []struct {
		name      string
		mediaType MediaType
		videoID   *string
		s3Key     *string
		want      *string
	}{
		{
			name:      "video with youtube_video_id",
			mediaType: MediaTypeVideo,
			videoID:   &videoID,
			s3Key:     nil,
			want:      &videoID,
		},
		{
			name:      "image with s3_object_key",
			mediaType: MediaTypeImage,
			videoID:   nil,
			s3Key:     &s3Key,
			want:      &s3Key,
		},
		{
			name:      "pending video without storage",
			mediaType: MediaTypeVideo,
			videoID:   nil,
			s3Key:     nil,
			want:      nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			asset := &MediaAsset{
				MediaType:      tt.mediaType,
				YouTubeVideoID: tt.videoID,
				S3ObjectKey:    tt.s3Key,
			}

			result := asset.GetStorageID()

			if tt.want == nil {
				assert.Nil(t, result)
			} else {
				require.NotNil(t, result)
				assert.Equal(t, *tt.want, *result)
			}
		})
	}
}
