-- Create media_assets table for storing video and image backup metadata
-- Part of media-backup-system feature (Phase 1.1)

CREATE TABLE IF NOT EXISTS media_assets (
    -- Primary identifier
    asset_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    -- User ownership
    user_id UUID NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,

    -- Storage location identifiers
    youtube_video_id VARCHAR(255) UNIQUE,  -- YouTube video ID (for videos)
    s3_object_key VARCHAR(512),             -- S3 object key (for images)

    -- File metadata
    original_filename VARCHAR(512) NOT NULL,
    file_size_bytes BIGINT NOT NULL CHECK (file_size_bytes > 0),
    media_type VARCHAR(10) NOT NULL CHECK (media_type IN ('VIDEO', 'IMAGE')),

    -- Sync status tracking
    sync_status VARCHAR(20) NOT NULL DEFAULT 'PENDING'
        CHECK (sync_status IN ('PENDING', 'UPLOADING', 'COMPLETED', 'FAILED')),

    -- Timing information
    upload_started_at TIMESTAMP,
    upload_completed_at TIMESTAMP,

    -- Error handling
    error_message TEXT,
    retry_count INT NOT NULL DEFAULT 0 CHECK (retry_count >= 0),

    -- Audit timestamps
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for efficient queries
CREATE INDEX IF NOT EXISTS idx_media_assets_user_id ON media_assets(user_id);
CREATE INDEX IF NOT EXISTS idx_media_assets_sync_status ON media_assets(sync_status);
CREATE INDEX IF NOT EXISTS idx_media_assets_created_at ON media_assets(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_media_assets_user_status ON media_assets(user_id, sync_status);

-- Create trigger for updated_at timestamp
CREATE OR REPLACE FUNCTION update_media_assets_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_media_assets_updated_at
    BEFORE UPDATE ON media_assets
    FOR EACH ROW
    EXECUTE FUNCTION update_media_assets_updated_at();

-- Add comment for documentation
COMMENT ON TABLE media_assets IS 'Stores metadata for backed up videos (YouTube) and images (S3)';
COMMENT ON COLUMN media_assets.youtube_video_id IS 'YouTube video ID for videos uploaded to private channel';
COMMENT ON COLUMN media_assets.s3_object_key IS 'Oracle Cloud S3 object key for images';
COMMENT ON COLUMN media_assets.sync_status IS 'Upload status: PENDING, UPLOADING, COMPLETED, FAILED';
