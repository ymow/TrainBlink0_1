-- Enable required extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- ============================================================
-- Users Table (Mobile App Users)
-- ============================================================
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),

    -- Authentication
    firebase_uid VARCHAR(128) UNIQUE NOT NULL,
    matrix_user_id VARCHAR(255),

    -- Profile
    display_name VARCHAR(100),
    avatar_emoji VARCHAR(10),
    avatar_color VARCHAR(7) CHECK (avatar_color ~ '^#[0-9A-Fa-f]{6}$'),
    status_text VARCHAR(36),

    -- Settings
    preferences JSONB DEFAULT '{}',

    -- Status
    is_active BOOLEAN DEFAULT true,
    is_banned BOOLEAN DEFAULT false,
    ban_reason TEXT,
    banned_at TIMESTAMP WITH TIME ZONE,
    banned_by UUID,
    ban_until TIMESTAMP WITH TIME ZONE,

    -- Statistics
    total_sessions INTEGER DEFAULT 0,
    total_encounters INTEGER DEFAULT 0,
    total_messages_sent INTEGER DEFAULT 0,
    total_content_shared INTEGER DEFAULT 0,

    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    last_seen_at TIMESTAMP WITH TIME ZONE,

    -- Privacy
    data_retention_days INTEGER DEFAULT 1 CHECK (data_retention_days BETWEEN 1 AND 30)
);

CREATE INDEX idx_users_firebase_uid ON users(firebase_uid);
CREATE INDEX idx_users_matrix_id ON users(matrix_user_id);
CREATE INDEX idx_users_is_active ON users(is_active) WHERE is_active = true;
CREATE INDEX idx_users_is_banned ON users(is_banned) WHERE is_banned = true;
CREATE INDEX idx_users_last_seen ON users(last_seen_at DESC NULLS LAST);
CREATE INDEX idx_users_created_at ON users(created_at DESC);

-- Auto-update updated_at trigger function
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_users_updated_at BEFORE UPDATE ON users
FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
