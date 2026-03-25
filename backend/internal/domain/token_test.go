package domain

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestToken_TableName(t *testing.T) {
	token := &Token{}
	assert.Equal(t, "user_tokens", token.TableName())
}

func TestToken_BeforeCreate(t *testing.T) {
	tests := []struct {
		name   string
		token  *Token
		wantID bool
	}{
		{
			name: "generates UUID when ID is nil",
			token: &Token{
				UserID: uuid.New(),
			},
			wantID: true,
		},
		{
			name: "keeps existing UUID",
			token: &Token{
				ID:     uuid.MustParse("550e8400-e29b-41d4-a716-446655440000"),
				UserID: uuid.New(),
			},
			wantID: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			originalID := tt.token.ID
			err := tt.token.BeforeCreate()

			require.NoError(t, err)
			assert.NotEqual(t, uuid.Nil, tt.token.ID, "ID should be set")
			assert.False(t, tt.token.CreatedAt.IsZero(), "CreatedAt should be set")
			assert.False(t, tt.token.UpdatedAt.IsZero(), "UpdatedAt should be set")

			if !tt.wantID {
				assert.Equal(t, originalID, tt.token.ID, "existing ID should not change")
			}
		})
	}
}

func TestToken_BeforeUpdate(t *testing.T) {
	token := &Token{
		ID:     uuid.New(),
		UserID: uuid.New(),
	}

	err := token.BeforeCreate()
	require.NoError(t, err)
	createdAt := token.CreatedAt

	err = token.BeforeUpdate()
	require.NoError(t, err)
	assert.Equal(t, createdAt, token.CreatedAt, "CreatedAt should not change on update")
	assert.False(t, token.UpdatedAt.IsZero(), "UpdatedAt should be set")
}

func TestToken_IsExpired(t *testing.T) {
	tests := []struct {
		name      string
		expiresAt time.Time
		want      bool
	}{
		{
			name:      "expired token (1 hour ago)",
			expiresAt: time.Now().Add(-1 * time.Hour),
			want:      true,
		},
		{
			name:      "valid token (1 hour from now)",
			expiresAt: time.Now().Add(1 * time.Hour),
			want:      false,
		},
		{
			name:      "just expired (1 second ago)",
			expiresAt: time.Now().Add(-1 * time.Second),
			want:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token := &Token{
				ExpiresAt: tt.expiresAt,
			}
			assert.Equal(t, tt.want, token.IsExpired())
		})
	}
}
