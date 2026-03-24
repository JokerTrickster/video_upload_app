-- Create user_tokens table
CREATE TABLE IF NOT EXISTS user_tokens (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    encrypted_access_token TEXT NOT NULL,
    encrypted_refresh_token TEXT NOT NULL,
    token_type VARCHAR(50) NOT NULL DEFAULT 'Bearer',
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    UNIQUE(user_id)
);

-- Create indexes
CREATE INDEX IF NOT EXISTS idx_user_tokens_user_id ON user_tokens(user_id);
CREATE INDEX IF NOT EXISTS idx_user_tokens_expires_at ON user_tokens(expires_at);

-- Add comments
COMMENT ON TABLE user_tokens IS 'Stores encrypted OAuth tokens';
COMMENT ON COLUMN user_tokens.encrypted_access_token IS 'AES-256 encrypted Google access token';
COMMENT ON COLUMN user_tokens.encrypted_refresh_token IS 'AES-256 encrypted Google refresh token';
COMMENT ON COLUMN user_tokens.expires_at IS 'Access token expiration timestamp';
