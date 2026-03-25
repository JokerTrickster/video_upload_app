package domain

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetRetryDelay(t *testing.T) {
	tests := []struct {
		attempt int
		want    int
	}{
		{attempt: 1, want: RetryDelay1},
		{attempt: 2, want: RetryDelay2},
		{attempt: 3, want: RetryDelay3},
		{attempt: 4, want: RetryDelay4},
		{attempt: 5, want: RetryDelay5},
		{attempt: 0, want: 0},
		{attempt: 6, want: 0},
		{attempt: -1, want: 0},
		{attempt: 100, want: 0},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("attempt_%d", tt.attempt), func(t *testing.T) {
			got := GetRetryDelay(tt.attempt)
			assert.Equal(t, tt.want, got, "GetRetryDelay(%d) = %d, want %d", tt.attempt, got, tt.want)
		})
	}

	// Verify delays are in increasing order
	t.Run("delays_increase_monotonically", func(t *testing.T) {
		prev := 0
		for attempt := 1; attempt <= 5; attempt++ {
			delay := GetRetryDelay(attempt)
			assert.Greater(t, delay, prev, "delay for attempt %d should be greater than attempt %d", attempt, attempt-1)
			prev = delay
		}
	})
}

func TestShouldRetry(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "nil error",
			err:  nil,
			want: false,
		},
		{
			name: "network timeout - retryable",
			err:  ErrNetworkTimeout,
			want: true,
		},
		{
			name: "youtube quota exceeded - retryable",
			err:  ErrYouTubeQuotaExceeded,
			want: true,
		},
		{
			name: "internal server error - retryable",
			err:  ErrInternalServer,
			want: true,
		},
		{
			name: "wrapped network timeout - retryable",
			err:  fmt.Errorf("upload failed: %w", ErrNetworkTimeout),
			want: true,
		},
		{
			name: "file too large - not retryable",
			err:  ErrFileTooLarge,
			want: false,
		},
		{
			name: "invalid file format - not retryable",
			err:  ErrInvalidFileFormat,
			want: false,
		},
		{
			name: "user not found - not retryable",
			err:  ErrUserNotFound,
			want: false,
		},
		{
			name: "token expired - not retryable",
			err:  ErrTokenExpired,
			want: false,
		},
		{
			name: "generic error - not retryable",
			err:  errors.New("something went wrong"),
			want: false,
		},
		{
			name: "verification failed - not retryable",
			err:  ErrVerificationFailed,
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ShouldRetry(tt.err)
			assert.Equal(t, tt.want, got, "ShouldRetry(%v) = %v, want %v", tt.err, got, tt.want)
		})
	}
}

func TestErrorConstants(t *testing.T) {
	// Verify error codes are unique
	codes := []string{
		ErrorCodeAuthInvalid,
		ErrorCodeTokenExpired,
		ErrorCodeFileTooLarge,
		ErrorCodeInvalidFormat,
		ErrorCodeQuotaExceeded,
		ErrorCodeNetworkTimeout,
		ErrorCodeInsufficientSpace,
		ErrorCodeVerificationFailed,
	}

	seen := make(map[string]bool)
	for _, code := range codes {
		assert.False(t, seen[code], "duplicate error code: %s", code)
		assert.NotEmpty(t, code, "error code should not be empty")
		seen[code] = true
	}
}

func TestRetryConstants(t *testing.T) {
	assert.Equal(t, 5, MaxRetryAttempts, "MaxRetryAttempts should be 5")
	assert.Equal(t, 1, RetryDelay1, "RetryDelay1 should be 1 second")
	assert.Equal(t, 2, RetryDelay2, "RetryDelay2 should be 2 seconds")
	assert.Equal(t, 5, RetryDelay3, "RetryDelay3 should be 5 seconds")
	assert.Equal(t, 15, RetryDelay4, "RetryDelay4 should be 15 seconds")
	assert.Equal(t, 30, RetryDelay5, "RetryDelay5 should be 30 seconds")
}
