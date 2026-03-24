-- Drop user_tokens table
DROP INDEX IF EXISTS idx_user_tokens_expires_at;
DROP INDEX IF EXISTS idx_user_tokens_user_id;
DROP TABLE IF EXISTS user_tokens;
