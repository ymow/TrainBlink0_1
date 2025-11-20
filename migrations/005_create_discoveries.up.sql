-- Create discoveries table for anonymized analytics
CREATE TABLE IF NOT EXISTS discoveries (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

  -- Anonymized trip information (NO user_id for privacy!)
  trip_route VARCHAR(200) NOT NULL,      -- "Tokyo → Osaka"

  -- Discovery metadata
  discovered_user_anonymous_id VARCHAR(50) NOT NULL,
  distance_estimate VARCHAR(20),         -- "Close (2-10m)", "Far (50-100m)", etc.
  rssi INTEGER,                          -- BLE signal strength in dBm

  -- Optional demographics (anonymized, opt-in)
  discoverer_age_range VARCHAR(20),      -- "25-30"
  discovered_age_range VARCHAR(20),      -- "25-30"
  discoverer_gender CHAR(1),             -- M/F/O
  discovered_gender CHAR(1),             -- M/F/O

  -- Timestamp
  discovered_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Indexes for analytics queries
CREATE INDEX idx_discoveries_route ON discoveries(trip_route);
CREATE INDEX idx_discoveries_timestamp ON discoveries(discovered_at);
CREATE INDEX idx_discoveries_distance ON discoveries(distance_estimate);
CREATE INDEX idx_discoveries_date ON discoveries(DATE(discovered_at));

-- Comments for documentation
COMMENT ON TABLE discoveries IS 'Anonymized BLE discovery analytics - no user_id for privacy';
COMMENT ON COLUMN discoveries.trip_route IS 'Trip route for analytics, not linked to specific user';
COMMENT ON COLUMN discoveries.rssi IS 'BLE RSSI signal strength in dBm (typically -100 to -50)';
COMMENT ON COLUMN discoveries.distance_estimate IS 'Human-readable distance category';
