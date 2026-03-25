package domain

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUploadQueueItem_TableName(t *testing.T) {
	item := &UploadQueueItem{}
	assert.Equal(t, "upload_queue", item.TableName())
}

func TestUploadQueueItem_BeforeCreate(t *testing.T) {
	item := &UploadQueueItem{
		UserID:   uuid.New(),
		FilePath: "/tmp/test.mp4",
		Filename: "test.mp4",
	}

	err := item.BeforeCreate()
	require.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, item.QueueID)
	assert.Equal(t, QueueStatusPending, item.QueueStatus)
	assert.False(t, item.CreatedAt.IsZero())
}

func TestDailyQuota_TableName(t *testing.T) {
	q := &DailyQuota{}
	assert.Equal(t, "daily_quotas", q.TableName())
}

func TestDailyQuota_CanUpload(t *testing.T) {
	tests := []struct {
		name      string
		unitsUsed int
		unitsMax  int
		want      bool
	}{
		{"fresh quota", 0, 10000, true},
		{"after 5 uploads", 8000, 10000, true},
		{"after 6 uploads", 9600, 10000, false},
		{"exactly at limit", 10000, 10000, false},
		{"over limit", 11000, 10000, false},
		{"one upload remaining", 8400, 10000, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q := &DailyQuota{UnitsUsed: tt.unitsUsed, UnitsMax: tt.unitsMax}
			assert.Equal(t, tt.want, q.CanUpload())
		})
	}
}

func TestDailyQuota_RemainingUploads(t *testing.T) {
	tests := []struct {
		name      string
		unitsUsed int
		want      int
	}{
		{"fresh", 0, 6},
		{"after 1 upload", 1600, 5},
		{"after 5 uploads", 8000, 1},
		{"after 6 uploads", 9600, 0},
		{"over limit", 11000, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q := &DailyQuota{UnitsUsed: tt.unitsUsed, UnitsMax: 10000}
			assert.Equal(t, tt.want, q.RemainingUploads())
		})
	}
}

func TestConstants(t *testing.T) {
	assert.Equal(t, 1600, YouTubeUploadCost)
	assert.Equal(t, 10000, DailyQuotaLimit)
	assert.Equal(t, 6, MaxDailyUploads)
}
