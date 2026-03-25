package domain

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUser_TableName(t *testing.T) {
	user := &User{}
	assert.Equal(t, "users", user.TableName())
}

func TestUser_BeforeCreate(t *testing.T) {
	tests := []struct {
		name   string
		user   *User
		wantID bool
	}{
		{
			name:   "generates UUID when ID is nil",
			user:   &User{},
			wantID: true,
		},
		{
			name: "keeps existing UUID",
			user: &User{
				ID: uuid.MustParse("550e8400-e29b-41d4-a716-446655440000"),
			},
			wantID: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			originalID := tt.user.ID
			err := tt.user.BeforeCreate()

			require.NoError(t, err)
			assert.NotEqual(t, uuid.Nil, tt.user.ID, "ID should be set")
			assert.False(t, tt.user.CreatedAt.IsZero(), "CreatedAt should be set")
			assert.False(t, tt.user.UpdatedAt.IsZero(), "UpdatedAt should be set")
			assert.Equal(t, tt.user.CreatedAt, tt.user.UpdatedAt, "CreatedAt and UpdatedAt should be equal on create")

			if !tt.wantID {
				assert.Equal(t, originalID, tt.user.ID, "existing ID should not change")
			}
		})
	}
}

func TestUser_BeforeUpdate(t *testing.T) {
	user := &User{
		ID:    uuid.New(),
		Email: "test@example.com",
	}

	// Set initial timestamps via BeforeCreate
	err := user.BeforeCreate()
	require.NoError(t, err)
	createdAt := user.CreatedAt

	// BeforeUpdate should only update UpdatedAt
	err = user.BeforeUpdate()
	require.NoError(t, err)
	assert.Equal(t, createdAt, user.CreatedAt, "CreatedAt should not change on update")
	assert.False(t, user.UpdatedAt.IsZero(), "UpdatedAt should be set")
}
