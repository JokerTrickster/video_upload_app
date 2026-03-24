-- Create users table
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) UNIQUE NOT NULL,
    google_id VARCHAR(255) UNIQUE NOT NULL,
    youtube_channel_id VARCHAR(255),
    youtube_channel_name VARCHAR(255),
    profile_image_url TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP
);

-- Create indexes
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_users_google_id ON users(google_id);
CREATE INDEX IF NOT EXISTS idx_users_deleted_at ON users(deleted_at);

-- Add comments
COMMENT ON TABLE users IS 'Stores user account information';
COMMENT ON COLUMN users.id IS 'Primary key UUID';
COMMENT ON COLUMN users.email IS 'User email from Google OAuth';
COMMENT ON COLUMN users.google_id IS 'Google account ID';
COMMENT ON COLUMN users.youtube_channel_id IS 'YouTube channel ID';
COMMENT ON COLUMN users.deleted_at IS 'Soft delete timestamp';
