-- Rollback migration for media_assets table
-- Part of media-backup-system feature (Phase 1.1)

-- Drop trigger first
DROP TRIGGER IF EXISTS trigger_media_assets_updated_at ON media_assets;

-- Drop trigger function
DROP FUNCTION IF EXISTS update_media_assets_updated_at();

-- Drop indexes
DROP INDEX IF EXISTS idx_media_assets_user_status;
DROP INDEX IF EXISTS idx_media_assets_created_at;
DROP INDEX IF EXISTS idx_media_assets_sync_status;
DROP INDEX IF EXISTS idx_media_assets_user_id;

-- Drop table (CASCADE will remove foreign key references if any)
DROP TABLE IF EXISTS media_assets CASCADE;
