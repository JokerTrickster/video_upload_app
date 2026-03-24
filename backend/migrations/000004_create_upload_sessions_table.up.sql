-- Create upload_sessions table for tracking batch upload operations
-- Part of media-backup-system feature (Phase 1.1)

CREATE TABLE IF NOT EXISTS upload_sessions (
    -- Primary identifier
    session_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    -- User ownership
    user_id UUID NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,

    -- File count tracking
    total_files INT NOT NULL DEFAULT 0 CHECK (total_files >= 0),
    completed_files INT NOT NULL DEFAULT 0 CHECK (completed_files >= 0),
    failed_files INT NOT NULL DEFAULT 0 CHECK (failed_files >= 0),

    -- Byte count tracking
    total_bytes BIGINT NOT NULL DEFAULT 0 CHECK (total_bytes >= 0),
    uploaded_bytes BIGINT NOT NULL DEFAULT 0 CHECK (uploaded_bytes >= 0),

    -- Session status
    session_status VARCHAR(20) NOT NULL DEFAULT 'ACTIVE'
        CHECK (session_status IN ('ACTIVE', 'COMPLETED', 'CANCELLED')),

    -- Timing information
    started_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    completed_at TIMESTAMP,

    -- Audit timestamps
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

    -- Data integrity constraints
    CONSTRAINT chk_completed_files_lte_total CHECK (completed_files <= total_files),
    CONSTRAINT chk_failed_files_lte_total CHECK (failed_files <= total_files),
    CONSTRAINT chk_uploaded_bytes_lte_total CHECK (uploaded_bytes <= total_bytes)
);

-- Create indexes for efficient queries
CREATE INDEX IF NOT EXISTS idx_upload_sessions_user_id ON upload_sessions(user_id);
CREATE INDEX IF NOT EXISTS idx_upload_sessions_status ON upload_sessions(session_status);
CREATE INDEX IF NOT EXISTS idx_upload_sessions_started_at ON upload_sessions(started_at DESC);
CREATE INDEX IF NOT EXISTS idx_upload_sessions_user_status ON upload_sessions(user_id, session_status);

-- Create trigger for updated_at timestamp
CREATE OR REPLACE FUNCTION update_upload_sessions_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_upload_sessions_updated_at
    BEFORE UPDATE ON upload_sessions
    FOR EACH ROW
    EXECUTE FUNCTION update_upload_sessions_updated_at();

-- Add comment for documentation
COMMENT ON TABLE upload_sessions IS 'Tracks batch upload sessions with progress and status';
COMMENT ON COLUMN upload_sessions.session_status IS 'Session state: ACTIVE, COMPLETED, CANCELLED';
COMMENT ON COLUMN upload_sessions.total_bytes IS 'Total bytes to upload in this session';
COMMENT ON COLUMN upload_sessions.uploaded_bytes IS 'Bytes successfully uploaded so far';
