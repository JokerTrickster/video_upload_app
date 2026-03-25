package config

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfig_Validate(t *testing.T) {
	validConfig := func() *Config {
		return &Config{
			Database: DatabaseConfig{URL: "postgres://localhost/test"},
			Google: GoogleConfig{
				ClientID:     "test-client-id",
				ClientSecret: "test-client-secret",
				RedirectURL:  "http://localhost/callback",
			},
			JWT: JWTConfig{
				Secret: "this-is-a-very-long-secret-key-32",
			},
			Security: SecurityConfig{
				EncryptionKey: "12345678901234567890123456789012", // exactly 32 chars
			},
		}
	}

	tests := []struct {
		name    string
		modify  func(*Config)
		wantErr string
	}{
		{
			name:   "valid config",
			modify: func(c *Config) {},
		},
		{
			name:    "missing DATABASE_URL",
			modify:  func(c *Config) { c.Database.URL = "" },
			wantErr: "DATABASE_URL is required",
		},
		{
			name:    "missing GOOGLE_CLIENT_ID",
			modify:  func(c *Config) { c.Google.ClientID = "" },
			wantErr: "GOOGLE_CLIENT_ID is required",
		},
		{
			name:    "missing GOOGLE_CLIENT_SECRET",
			modify:  func(c *Config) { c.Google.ClientSecret = "" },
			wantErr: "GOOGLE_CLIENT_SECRET is required",
		},
		{
			name:    "missing GOOGLE_REDIRECT_URL",
			modify:  func(c *Config) { c.Google.RedirectURL = "" },
			wantErr: "GOOGLE_REDIRECT_URL is required",
		},
		{
			name:    "empty JWT_SECRET",
			modify:  func(c *Config) { c.JWT.Secret = "" },
			wantErr: "JWT_SECRET must be at least 32 characters",
		},
		{
			name:    "short JWT_SECRET",
			modify:  func(c *Config) { c.JWT.Secret = "short" },
			wantErr: "JWT_SECRET must be at least 32 characters",
		},
		{
			name:    "empty ENCRYPTION_KEY",
			modify:  func(c *Config) { c.Security.EncryptionKey = "" },
			wantErr: "ENCRYPTION_KEY must be exactly 32 characters",
		},
		{
			name:    "wrong length ENCRYPTION_KEY",
			modify:  func(c *Config) { c.Security.EncryptionKey = "too-short" },
			wantErr: "ENCRYPTION_KEY must be exactly 32 characters",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := validConfig()
			tt.modify(cfg)

			err := cfg.Validate()
			if tt.wantErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestGetEnv(t *testing.T) {
	tests := []struct {
		name         string
		key          string
		envValue     string
		defaultValue string
		want         string
	}{
		{
			name:         "returns env value when set",
			key:          "TEST_GET_ENV_1",
			envValue:     "from-env",
			defaultValue: "default",
			want:         "from-env",
		},
		{
			name:         "returns default when env not set",
			key:          "TEST_GET_ENV_2",
			envValue:     "",
			defaultValue: "default-value",
			want:         "default-value",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.envValue != "" {
				os.Setenv(tt.key, tt.envValue)
				defer os.Unsetenv(tt.key)
			} else {
				os.Unsetenv(tt.key)
			}

			got := getEnv(tt.key, tt.defaultValue)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestGetEnvInt(t *testing.T) {
	tests := []struct {
		name         string
		key          string
		envValue     string
		defaultValue int
		want         int
	}{
		{
			name:         "returns parsed int from env",
			key:          "TEST_GET_ENV_INT_1",
			envValue:     "42",
			defaultValue: 10,
			want:         42,
		},
		{
			name:         "returns default for invalid int",
			key:          "TEST_GET_ENV_INT_2",
			envValue:     "not-a-number",
			defaultValue: 10,
			want:         10,
		},
		{
			name:         "returns default when env not set",
			key:          "TEST_GET_ENV_INT_3",
			envValue:     "",
			defaultValue: 99,
			want:         99,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.envValue != "" {
				os.Setenv(tt.key, tt.envValue)
				defer os.Unsetenv(tt.key)
			} else {
				os.Unsetenv(tt.key)
			}

			got := getEnvInt(tt.key, tt.defaultValue)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestGetEnvDuration(t *testing.T) {
	tests := []struct {
		name         string
		key          string
		envValue     string
		defaultValue time.Duration
		want         time.Duration
	}{
		{
			name:         "returns parsed duration from env",
			key:          "TEST_GET_ENV_DUR_1",
			envValue:     "30m",
			defaultValue: time.Hour,
			want:         30 * time.Minute,
		},
		{
			name:         "returns default for invalid duration",
			key:          "TEST_GET_ENV_DUR_2",
			envValue:     "invalid",
			defaultValue: time.Hour,
			want:         time.Hour,
		},
		{
			name:         "returns default when env not set",
			key:          "TEST_GET_ENV_DUR_3",
			envValue:     "",
			defaultValue: 5 * time.Minute,
			want:         5 * time.Minute,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.envValue != "" {
				os.Setenv(tt.key, tt.envValue)
				defer os.Unsetenv(tt.key)
			} else {
				os.Unsetenv(tt.key)
			}

			got := getEnvDuration(tt.key, tt.defaultValue)
			assert.Equal(t, tt.want, got)
		})
	}
}
