package redis

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBuildKey(t *testing.T) {
	tests := []struct {
		name  string
		parts []string
		want  string
	}{
		{
			name:  "single part",
			parts: []string{"users"},
			want:  "users",
		},
		{
			name:  "two parts",
			parts: []string{"jwt", "blacklist"},
			want:  "jwt:blacklist",
		},
		{
			name:  "three parts",
			parts: []string{"oauth", "state", "abc123"},
			want:  "oauth:state:abc123",
		},
		{
			name:  "rate limit key",
			parts: []string{"rate_limit", "192.168.1.1"},
			want:  "rate_limit:192.168.1.1",
		},
		{
			name:  "empty parts",
			parts: []string{},
			want:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := BuildKey(tt.parts...)
			assert.Equal(t, tt.want, got)
		})
	}
}
