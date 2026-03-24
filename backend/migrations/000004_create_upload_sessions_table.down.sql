-- Rollback migration for upload_sessions table
-- Part of media-backup-system feature (Phase 1.1)

-- Drop trigger first
DROP TRIGGER IF EXISTS trigger_upload_sessions_updated_at ON upload_sessions;

-- Drop trigger function
DROP FUNCTION IF EXISTS update_upload_sessions_updated_at();

-- Drop indexes
DROP INDEX IF EXISTS idx_upload_sessions_user_status;
DROP INDEX IF EXISTS idx_upload_sessions_started_at;
DROP INDEX IF EXISTS idx_upload_sessions_status;
DROP INDEX IF EXISTS idx_upload_sessions_user_id;

-- Drop table (CASCADE will remove foreign key references if any)
DROP TABLE IF EXISTS upload_sessions CASCADE;
