-- Create trips table for MVP
CREATE TABLE IF NOT EXISTS trips (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

  user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,

  -- Trip details
  route VARCHAR(200) NOT NULL,           -- "Tokyo → Osaka"
  train_number VARCHAR(50),              -- "Nozomi 123" (optional)

  -- Timing
  departure_time TIMESTAMP WITH TIME ZONE NOT NULL,
  estimated_arrival TIMESTAMP WITH TIME ZONE NOT NULL,
  actual_end_time TIMESTAMP WITH TIME ZONE,

  -- BLE Discovery
  discovery_enabled BOOLEAN DEFAULT true,
  ble_anonymous_id VARCHAR(50) NOT NULL UNIQUE, -- "TB_abc123"

  -- Status
  status VARCHAR(20) DEFAULT 'active' CHECK (status IN ('active', 'ended', 'cancelled')),

  -- Timestamps
  created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
  updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Indexes for performance
CREATE INDEX idx_trips_user_id ON trips(user_id);
CREATE INDEX idx_trips_status ON trips(status);
CREATE INDEX idx_trips_ble_id ON trips(ble_anonymous_id);
CREATE INDEX idx_trips_departure ON trips(departure_time);
CREATE INDEX idx_trips_route ON trips(route);

-- Function to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_trips_updated_at()
RETURNS TRIGGER AS $$
BEGIN
  NEW.updated_at = NOW();
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Trigger to automatically update updated_at
CREATE TRIGGER trips_updated_at_trigger
  BEFORE UPDATE ON trips
  FOR EACH ROW
  EXECUTE FUNCTION update_trips_updated_at();

-- Comments for documentation
COMMENT ON TABLE trips IS 'User trips for TrainBlink MVP - manual trip selection';
COMMENT ON COLUMN trips.route IS 'Trip route in format "Origin → Destination"';
COMMENT ON COLUMN trips.ble_anonymous_id IS 'Anonymous ID for BLE broadcasting (TB_xxx)';
COMMENT ON COLUMN trips.status IS 'Trip status: active (ongoing), ended (completed), cancelled';
