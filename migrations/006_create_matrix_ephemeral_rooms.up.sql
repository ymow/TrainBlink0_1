-- Create matrix_ephemeral_rooms table for temporary 1-on-1 DM rooms
CREATE TABLE IF NOT EXISTS matrix_ephemeral_rooms (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

  room_id VARCHAR(255) NOT NULL UNIQUE,  -- !dm_xyz:trainblink.org

  -- Participants (via trips, not direct user_id for privacy)
  trip1_id UUID NOT NULL REFERENCES trips(id) ON DELETE CASCADE,
  trip2_id UUID NOT NULL REFERENCES trips(id) ON DELETE CASCADE,

  -- Anonymous IDs for tracking (no user_id!)
  anonymous_id_1 VARCHAR(50) NOT NULL,   -- TB_abc123
  anonymous_id_2 VARCHAR(50) NOT NULL,   -- TB_def456

  -- MLS encryption group
  mls_group_id VARCHAR(255),

  -- Ephemeral settings
  expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
  auto_delete_queued BOOLEAN DEFAULT false,
  deleted_at TIMESTAMP WITH TIME ZONE,

  -- Metadata
  message_count INTEGER DEFAULT 0,
  last_message_at TIMESTAMP WITH TIME ZONE,

  -- Timestamps
  created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Indexes for queries
CREATE INDEX idx_ephemeral_rooms_expires ON matrix_ephemeral_rooms(expires_at) WHERE deleted_at IS NULL;
CREATE INDEX idx_ephemeral_rooms_trip1 ON matrix_ephemeral_rooms(trip1_id);
CREATE INDEX idx_ephemeral_rooms_trip2 ON matrix_ephemeral_rooms(trip2_id);
CREATE INDEX idx_ephemeral_rooms_room_id ON matrix_ephemeral_rooms(room_id);
CREATE INDEX idx_ephemeral_rooms_mls ON matrix_ephemeral_rooms(mls_group_id);
CREATE INDEX idx_ephemeral_rooms_auto_delete ON matrix_ephemeral_rooms(auto_delete_queued, expires_at) WHERE deleted_at IS NULL;

-- Check constraint to ensure trip1_id != trip2_id
ALTER TABLE matrix_ephemeral_rooms
ADD CONSTRAINT check_different_trips CHECK (trip1_id != trip2_id);

-- Comments for documentation
COMMENT ON TABLE matrix_ephemeral_rooms IS 'Ephemeral 1-on-1 DM rooms that auto-delete after trip ends';
COMMENT ON COLUMN matrix_ephemeral_rooms.expires_at IS 'Room will be auto-deleted after this time (typically 24h after creation)';
COMMENT ON COLUMN matrix_ephemeral_rooms.auto_delete_queued IS 'True if room is queued for deletion by cleanup service';
COMMENT ON COLUMN matrix_ephemeral_rooms.trip1_id IS 'Reference to first user trip (for privacy, not direct user_id)';
COMMENT ON COLUMN matrix_ephemeral_rooms.trip2_id IS 'Reference to second user trip (for privacy, not direct user_id)';
