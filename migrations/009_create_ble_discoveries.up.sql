-- ============================================================
-- BLE Discovery-based Messaging Permissions
-- ============================================================
-- This table tracks BLE proximity discoveries between users.
-- Users can ONLY message others they've discovered via BLE
-- within the last 10 minutes (enforces 50-100m proximity requirement).

CREATE TABLE IF NOT EXISTS ble_discoveries (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    -- Who discovered whom
    discoverer_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    discovered_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,

    -- Discovery context (for analytics and debugging)
    trip_id UUID REFERENCES trips(id) ON DELETE SET NULL,

    -- BLE metadata
    rssi INTEGER,                                    -- Signal strength in dBm (typically -100 to -50)
    distance_estimate VARCHAR(20),                   -- "Close (2-10m)", "Medium (10-50m)", "Far (50-100m)"

    -- Discovery timestamp (CRITICAL for TTL validation)
    discovered_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),

    -- Expiry timestamp (10 minutes from discovery)
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT (NOW() + INTERVAL '10 minutes'),

    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),

    -- Prevent duplicate discoveries within same trip
    CONSTRAINT unique_discovery_pair UNIQUE (discoverer_id, discovered_id, trip_id)
);

-- ============================================================
-- CRITICAL INDEXES for permission checks (must be fast!)
-- ============================================================

-- Primary index for checking if A can message B
-- This is the MOST IMPORTANT index - used in every permission check
CREATE INDEX idx_ble_discoveries_permission_check
    ON ble_discoveries(discoverer_id, discovered_id, expires_at DESC)
    WHERE expires_at > NOW();

-- Index for bidirectional discovery check
-- Useful for finding mutual discoveries
CREATE INDEX idx_ble_discoveries_bidirectional
    ON ble_discoveries(
        LEAST(discoverer_id, discovered_id),
        GREATEST(discoverer_id, discovered_id),
        expires_at DESC
    )
    WHERE expires_at > NOW();

-- Index for cleanup of expired discoveries
-- Used by background worker to remove old records
CREATE INDEX idx_ble_discoveries_expired
    ON ble_discoveries(expires_at)
    WHERE expires_at <= NOW();

-- Index for user's active discoveries
-- Used to show "nearby users" in mobile app
CREATE INDEX idx_ble_discoveries_active
    ON ble_discoveries(discoverer_id, expires_at DESC)
    WHERE expires_at > NOW();

-- Index for trip-based discovery queries
CREATE INDEX idx_ble_discoveries_trip
    ON ble_discoveries(trip_id, discovered_at DESC)
    WHERE trip_id IS NOT NULL;

-- ============================================================
-- Comments for documentation
-- ============================================================

COMMENT ON TABLE ble_discoveries IS 'BLE proximity discoveries - grants 10-minute messaging permission between users';
COMMENT ON COLUMN ble_discoveries.discoverer_id IS 'User who performed the BLE scan and discovered another user';
COMMENT ON COLUMN ble_discoveries.discovered_id IS 'User who was discovered via BLE broadcast';
COMMENT ON COLUMN ble_discoveries.expires_at IS 'Discovery expires 10 minutes after discovered_at - after expiry, users cannot message';
COMMENT ON COLUMN ble_discoveries.rssi IS 'BLE RSSI signal strength in dBm (typically -100 to -50, lower = farther)';
COMMENT ON COLUMN ble_discoveries.distance_estimate IS 'Human-readable distance estimate based on RSSI';
