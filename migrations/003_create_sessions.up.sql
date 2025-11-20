-- ============================================================
-- User Sessions (Station Check-ins) - Updated version
-- ============================================================
-- Drop existing table if migrating from old schema
DROP TABLE IF EXISTS user_sessions CASCADE;

CREATE TABLE user_sessions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),

    -- References
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    device_id VARCHAR(255) NOT NULL,
    station_id VARCHAR(50) NOT NULL,

    -- Timing
    entered_at TIMESTAMP WITH TIME ZONE NOT NULL,
    exited_at TIMESTAMP WITH TIME ZONE,
    duration_seconds INTEGER,

    -- Location
    entry_coordinates GEOGRAPHY(POINT, 4326),
    exit_coordinates GEOGRAPHY(POINT, 4326),

    -- Capabilities
    capabilities JSONB,

    -- P2P Stats
    p2p_chats_created INTEGER DEFAULT 0,
    p2p_messages_sent INTEGER DEFAULT 0,
    p2p_content_shared INTEGER DEFAULT 0,
    encounters INTEGER DEFAULT 0,

    -- Matrix Stats
    matrix_user_id VARCHAR(255),
    matrix_room_id VARCHAR(255),
    matrix_messages_sent INTEGER DEFAULT 0,
    matrix_joined_at TIMESTAMP WITH TIME ZONE,
    matrix_left_at TIMESTAMP WITH TIME ZONE,

    -- Metadata
    client_version VARCHAR(20),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_sessions_user_id ON user_sessions(user_id);
CREATE INDEX idx_sessions_station_id ON user_sessions(station_id);
CREATE INDEX idx_sessions_entered_at ON user_sessions(entered_at DESC);
CREATE INDEX idx_sessions_device_id ON user_sessions(device_id);
CREATE INDEX idx_sessions_matrix_room ON user_sessions(matrix_room_id);
CREATE INDEX idx_sessions_active ON user_sessions(exited_at) WHERE exited_at IS NULL;

-- ============================================================
-- Station Analytics (Daily Aggregates)
-- ============================================================
CREATE TABLE station_analytics (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    station_id VARCHAR(50) NOT NULL,
    date DATE NOT NULL,

    -- Daily Metrics
    total_entries INTEGER DEFAULT 0,
    total_exits INTEGER DEFAULT 0,
    avg_duration_seconds INTEGER,
    peak_hour INTEGER CHECK (peak_hour BETWEEN 0 AND 23),
    peak_users INTEGER,

    -- P2P Engagement
    p2p_chats_created INTEGER DEFAULT 0,
    p2p_content_shared INTEGER DEFAULT 0,

    -- Matrix Engagement
    matrix_messages_sent INTEGER DEFAULT 0,
    matrix_active_members INTEGER DEFAULT 0,
    matrix_peak_members INTEGER DEFAULT 0,

    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),

    UNIQUE(station_id, date)
);

CREATE INDEX idx_analytics_station_date ON station_analytics(station_id, date DESC);
CREATE INDEX idx_analytics_date ON station_analytics(date DESC);
