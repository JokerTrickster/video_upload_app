package domain

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUploadSession_TableName(t *testing.T) {
	session := &UploadSession{}
	assert.Equal(t, "upload_sessions", session.TableName())
}

func TestUploadSession_BeforeCreate(t *testing.T) {
	tests := []struct {
		name    string
		session *UploadSession
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid session",
			session: &UploadSession{
				UserID:     uuid.New(),
				TotalFiles: 10,
				TotalBytes: 1024000,
			},
			wantErr: false,
		},
		{
			name: "missing user_id",
			session: &UploadSession{
				TotalFiles: 10,
				TotalBytes: 1024000,
			},
			wantErr: true,
			errMsg:  "user_id is required",
		},
		{
			name: "negative total_files",
			session: &UploadSession{
				UserID:     uuid.New(),
				TotalFiles: -1,
				TotalBytes: 1024000,
			},
			wantErr: true,
			errMsg:  "total_files cannot be negative",
		},
		{
			name: "completed_files exceeds total_files",
			session: &UploadSession{
				UserID:         uuid.New(),
				TotalFiles:     10,
				CompletedFiles: 15,
				TotalBytes:     1024000,
			},
			wantErr: true,
			errMsg:  "completed_files cannot exceed total_files",
		},
		{
			name: "uploaded_bytes exceeds total_bytes",
			session: &UploadSession{
				UserID:        uuid.New(),
				TotalFiles:    10,
				TotalBytes:    1024000,
				UploadedBytes: 2048000,
			},
			wantErr: true,
			errMsg:  "uploaded_bytes cannot exceed total_bytes",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.session.BeforeCreate()

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				require.NoError(t, err)
				assert.NotEqual(t, uuid.Nil, tt.session.SessionID)
				assert.Equal(t, SessionStatusActive, tt.session.SessionStatus)
				assert.False(t, tt.session.CreatedAt.IsZero())
				assert.False(t, tt.session.UpdatedAt.IsZero())
				assert.False(t, tt.session.StartedAt.IsZero())
			}
		})
	}
}

func TestUploadSession_Validate(t *testing.T) {
	tests := []struct {
		name    string
		session *UploadSession
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid active session",
			session: &UploadSession{
				UserID:         uuid.New(),
				TotalFiles:     10,
				CompletedFiles: 3,
				FailedFiles:    1,
				TotalBytes:     1024000,
				UploadedBytes:  300000,
				SessionStatus:  SessionStatusActive,
			},
			wantErr: false,
		},
		{
			name: "invalid session status",
			session: &UploadSession{
				UserID:        uuid.New(),
				TotalFiles:    10,
				TotalBytes:    1024000,
				SessionStatus: "UNKNOWN",
			},
			wantErr: true,
			errMsg:  "invalid session_status",
		},
		{
			name: "failed_files exceeds total_files",
			session: &UploadSession{
				UserID:        uuid.New(),
				TotalFiles:    10,
				FailedFiles:   15,
				TotalBytes:    1024000,
				SessionStatus: SessionStatusActive,
			},
			wantErr: true,
			errMsg:  "failed_files cannot exceed total_files",
		},
		{
			name: "negative uploaded_bytes",
			session: &UploadSession{
				UserID:        uuid.New(),
				TotalFiles:    10,
				TotalBytes:    1024000,
				UploadedBytes: -100,
				SessionStatus: SessionStatusActive,
			},
			wantErr: true,
			errMsg:  "uploaded_bytes cannot be negative",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.session.Validate()

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestUploadSession_StatusCheckers(t *testing.T) {
	tests := []struct {
		name          string
		sessionStatus SessionStatus
		checks        map[string]bool
	}{
		{
			name:          "active status",
			sessionStatus: SessionStatusActive,
			checks: map[string]bool{
				"IsActive":    true,
				"IsCompleted": false,
				"IsCancelled": false,
			},
		},
		{
			name:          "completed status",
			sessionStatus: SessionStatusCompleted,
			checks: map[string]bool{
				"IsActive":    false,
				"IsCompleted": true,
				"IsCancelled": false,
			},
		},
		{
			name:          "cancelled status",
			sessionStatus: SessionStatusCancelled,
			checks: map[string]bool{
				"IsActive":    false,
				"IsCompleted": false,
				"IsCancelled": true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			session := &UploadSession{SessionStatus: tt.sessionStatus}

			assert.Equal(t, tt.checks["IsActive"], session.IsActive(), "IsActive mismatch")
			assert.Equal(t, tt.checks["IsCompleted"], session.IsCompleted(), "IsCompleted mismatch")
			assert.Equal(t, tt.checks["IsCancelled"], session.IsCancelled(), "IsCancelled mismatch")
		})
	}
}

func TestUploadSession_CalculateProgress(t *testing.T) {
	tests := []struct {
		name          string
		totalBytes    int64
		uploadedBytes int64
		want          float64
	}{
		{
			name:          "0% progress",
			totalBytes:    1024000,
			uploadedBytes: 0,
			want:          0.0,
		},
		{
			name:          "50% progress",
			totalBytes:    1024000,
			uploadedBytes: 512000,
			want:          50.0,
		},
		{
			name:          "100% progress",
			totalBytes:    1024000,
			uploadedBytes: 1024000,
			want:          100.0,
		},
		{
			name:          "zero total bytes",
			totalBytes:    0,
			uploadedBytes: 0,
			want:          0.0,
		},
		{
			name:          "33.33% progress",
			totalBytes:    3000000,
			uploadedBytes: 1000000,
			want:          33.333333333333336,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			session := &UploadSession{
				TotalBytes:    tt.totalBytes,
				UploadedBytes: tt.uploadedBytes,
			}
			assert.InDelta(t, tt.want, session.CalculateProgress(), 0.0001)
		})
	}
}

func TestUploadSession_GetPendingFiles(t *testing.T) {
	tests := []struct {
		name           string
		totalFiles     int
		completedFiles int
		failedFiles    int
		want           int
	}{
		{
			name:           "all pending",
			totalFiles:     10,
			completedFiles: 0,
			failedFiles:    0,
			want:           10,
		},
		{
			name:           "some completed",
			totalFiles:     10,
			completedFiles: 5,
			failedFiles:    0,
			want:           5,
		},
		{
			name:           "some failed",
			totalFiles:     10,
			completedFiles: 3,
			failedFiles:    2,
			want:           5,
		},
		{
			name:           "all processed",
			totalFiles:     10,
			completedFiles: 7,
			failedFiles:    3,
			want:           0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			session := &UploadSession{
				TotalFiles:     tt.totalFiles,
				CompletedFiles: tt.completedFiles,
				FailedFiles:    tt.failedFiles,
			}
			assert.Equal(t, tt.want, session.GetPendingFiles())
		})
	}
}

func TestUploadSession_IncrementCompleted(t *testing.T) {
	tests := []struct {
		name           string
		initialSession *UploadSession
		fileBytes      int64
		wantErr        bool
		wantCompleted  int
		wantUploaded   int64
		wantStatus     SessionStatus
	}{
		{
			name: "normal increment",
			initialSession: &UploadSession{
				UserID:         uuid.New(),
				TotalFiles:     10,
				CompletedFiles: 3,
				FailedFiles:    0,
				TotalBytes:     1024000,
				UploadedBytes:  300000,
				SessionStatus:  SessionStatusActive,
			},
			fileBytes:     100000,
			wantErr:       false,
			wantCompleted: 4,
			wantUploaded:  400000,
			wantStatus:    SessionStatusActive,
		},
		{
			name: "auto-complete when all files processed",
			initialSession: &UploadSession{
				UserID:         uuid.New(),
				TotalFiles:     10,
				CompletedFiles: 9,
				FailedFiles:    0,
				TotalBytes:     1024000,
				UploadedBytes:  924000,
				SessionStatus:  SessionStatusActive,
			},
			fileBytes:     100000,
			wantErr:       false,
			wantCompleted: 10,
			wantUploaded:  1024000,
			wantStatus:    SessionStatusCompleted,
		},
		{
			name: "negative file bytes",
			initialSession: &UploadSession{
				UserID:         uuid.New(),
				TotalFiles:     10,
				CompletedFiles: 3,
				SessionStatus:  SessionStatusActive,
			},
			fileBytes:     -100,
			wantErr:       true,
			wantCompleted: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.initialSession.IncrementCompleted(tt.fileBytes)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.wantCompleted, tt.initialSession.CompletedFiles)
				assert.Equal(t, tt.wantUploaded, tt.initialSession.UploadedBytes)
				assert.Equal(t, tt.wantStatus, tt.initialSession.SessionStatus)

				if tt.wantStatus == SessionStatusCompleted {
					assert.NotNil(t, tt.initialSession.CompletedAt)
				}
			}
		})
	}
}

func TestUploadSession_IncrementFailed(t *testing.T) {
	tests := []struct {
		name           string
		initialSession *UploadSession
		wantFailed     int
		wantStatus     SessionStatus
	}{
		{
			name: "normal increment",
			initialSession: &UploadSession{
				UserID:         uuid.New(),
				TotalFiles:     10,
				CompletedFiles: 3,
				FailedFiles:    1,
				SessionStatus:  SessionStatusActive,
			},
			wantFailed: 2,
			wantStatus: SessionStatusActive,
		},
		{
			name: "auto-complete when all files processed",
			initialSession: &UploadSession{
				UserID:         uuid.New(),
				TotalFiles:     10,
				CompletedFiles: 7,
				FailedFiles:    2,
				SessionStatus:  SessionStatusActive,
			},
			wantFailed: 3,
			wantStatus: SessionStatusCompleted,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.initialSession.IncrementFailed()

			assert.Equal(t, tt.wantFailed, tt.initialSession.FailedFiles)
			assert.Equal(t, tt.wantStatus, tt.initialSession.SessionStatus)

			if tt.wantStatus == SessionStatusCompleted {
				assert.NotNil(t, tt.initialSession.CompletedAt)
			}
		})
	}
}

func TestUploadSession_Complete(t *testing.T) {
	session := &UploadSession{
		UserID:        uuid.New(),
		TotalFiles:    10,
		SessionStatus: SessionStatusActive,
	}

	session.Complete()

	assert.Equal(t, SessionStatusCompleted, session.SessionStatus)
	require.NotNil(t, session.CompletedAt)
	assert.False(t, session.CompletedAt.IsZero())

	// Complete again - should not change CompletedAt
	firstCompleteTime := *session.CompletedAt
	time.Sleep(10 * time.Millisecond)
	session.Complete()
	assert.Equal(t, firstCompleteTime, *session.CompletedAt)
}

func TestUploadSession_Cancel(t *testing.T) {
	session := &UploadSession{
		UserID:        uuid.New(),
		TotalFiles:    10,
		SessionStatus: SessionStatusActive,
	}

	session.Cancel()

	assert.Equal(t, SessionStatusCancelled, session.SessionStatus)
	require.NotNil(t, session.CompletedAt)
	assert.False(t, session.CompletedAt.IsZero())

	// Cancel again - should not change CompletedAt
	firstCancelTime := *session.CompletedAt
	time.Sleep(10 * time.Millisecond)
	session.Cancel()
	assert.Equal(t, firstCancelTime, *session.CompletedAt)
}

func TestUploadSession_GetSuccessRate(t *testing.T) {
	tests := []struct {
		name           string
		completedFiles int
		failedFiles    int
		want           float64
	}{
		{
			name:           "100% success",
			completedFiles: 10,
			failedFiles:    0,
			want:           100.0,
		},
		{
			name:           "0% success (all failed)",
			completedFiles: 0,
			failedFiles:    10,
			want:           0.0,
		},
		{
			name:           "50% success",
			completedFiles: 5,
			failedFiles:    5,
			want:           50.0,
		},
		{
			name:           "70% success",
			completedFiles: 7,
			failedFiles:    3,
			want:           70.0,
		},
		{
			name:           "no files processed yet",
			completedFiles: 0,
			failedFiles:    0,
			want:           0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			session := &UploadSession{
				CompletedFiles: tt.completedFiles,
				FailedFiles:    tt.failedFiles,
			}
			assert.InDelta(t, tt.want, session.GetSuccessRate(), 0.0001)
		})
	}
}

func TestUploadSession_GetDuration(t *testing.T) {
	now := time.Now()
	fiveMinutesAgo := now.Add(-5 * time.Minute)

	tests := []struct {
		name        string
		startedAt   time.Time
		completedAt *time.Time
		minDuration time.Duration
		maxDuration time.Duration
	}{
		{
			name:        "active session (5 minutes old)",
			startedAt:   fiveMinutesAgo,
			completedAt: nil,
			minDuration: 4*time.Minute + 59*time.Second,
			maxDuration: 5*time.Minute + 1*time.Second,
		},
		{
			name:        "completed session (1 hour duration)",
			startedAt:   fiveMinutesAgo,
			completedAt: func() *time.Time { t := fiveMinutesAgo.Add(1 * time.Hour); return &t }(),
			minDuration: 59 * time.Minute,
			maxDuration: 61 * time.Minute,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			session := &UploadSession{
				StartedAt:   tt.startedAt,
				CompletedAt: tt.completedAt,
			}

			duration := session.GetDuration()

			assert.GreaterOrEqual(t, duration, tt.minDuration)
			assert.LessOrEqual(t, duration, tt.maxDuration)
		})
	}
}
