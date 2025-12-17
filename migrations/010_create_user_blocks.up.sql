-- ============================================================
-- User Blocking System
-- ============================================================
-- This table allows users to block others from messaging them.
-- Blocked users CANNOT send messages to their blocker, regardless
-- of BLE discovery status.

CREATE TABLE IF NOT EXISTS user_blocks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    -- Who blocked whom
    blocker_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    blocked_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,

    -- Optional reason (for user's reference, not shown to blocked user)
    reason TEXT,

    -- Timestamps
    blocked_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),

    -- Ensure user can't block same person twice
    CONSTRAINT unique_block_pair UNIQUE (blocker_id, blocked_id),

    -- User cannot block themselves
    CONSTRAINT no_self_block CHECK (blocker_id != blocked_id)
);

-- ============================================================
-- CRITICAL INDEXES for block checks (must be fast!)
-- ============================================================

-- Primary index for checking if B is blocked by A
-- This is used in every permission check
CREATE INDEX idx_user_blocks_check
    ON user_blocks(blocker_id, blocked_id);

-- Reverse lookup (who has blocked this user)
-- Useful for analytics and debugging
CREATE INDEX idx_user_blocks_reverse
    ON user_blocks(blocked_id, blocker_id);

-- Index for getting user's block list
-- Used by "Get Blocked Users" API endpoint
CREATE INDEX idx_user_blocks_blocker
    ON user_blocks(blocker_id, blocked_at DESC);

-- ============================================================
-- Comments for documentation
-- ============================================================

COMMENT ON TABLE user_blocks IS 'User blocking system - prevents messaging from blocked users';
COMMENT ON COLUMN user_blocks.blocker_id IS 'User who initiated the block';
COMMENT ON COLUMN user_blocks.blocked_id IS 'User who is blocked (cannot send messages to blocker)';
COMMENT ON COLUMN user_blocks.reason IS 'Optional reason for block (private, only visible to blocker)';
