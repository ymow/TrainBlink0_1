-- Update matrix_users table for MVP requirements
-- This table might not exist yet, so we create it if needed

CREATE TABLE IF NOT EXISTS matrix_users (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

  user_id UUID NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE,
  matrix_user_id VARCHAR(255) NOT NULL UNIQUE,  -- @trainblink_xxx:trainblink.org
  display_name VARCHAR(100),
  avatar_url TEXT,

  -- Authentication
  access_token TEXT NOT NULL,                    -- Encrypted Matrix access token
  device_id VARCHAR(255),

  -- Status
  is_active BOOLEAN DEFAULT true,
  last_seen_at TIMESTAMP WITH TIME ZONE,

  -- Timestamps
  created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
  expires_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() + INTERVAL '30 days'
);

-- Add indexes if they don't exist
DO $$
BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_indexes WHERE indexname = 'idx_matrix_users_user_id') THEN
    CREATE INDEX idx_matrix_users_user_id ON matrix_users(user_id);
  END IF;

  IF NOT EXISTS (SELECT 1 FROM pg_indexes WHERE indexname = 'idx_matrix_users_matrix_id') THEN
    CREATE INDEX idx_matrix_users_matrix_id ON matrix_users(matrix_user_id);
  END IF;

  IF NOT EXISTS (SELECT 1 FROM pg_indexes WHERE indexname = 'idx_matrix_users_active') THEN
    CREATE INDEX idx_matrix_users_active ON matrix_users(is_active);
  END IF;

  IF NOT EXISTS (SELECT 1 FROM pg_indexes WHERE indexname = 'idx_matrix_users_expires') THEN
    CREATE INDEX idx_matrix_users_expires ON matrix_users(expires_at) WHERE is_active = true;
  END IF;
END $$;

-- Add columns if table already exists but missing columns
DO $$
BEGIN
  -- Add display_name if it doesn't exist
  IF NOT EXISTS (SELECT 1 FROM information_schema.columns
                 WHERE table_name = 'matrix_users' AND column_name = 'display_name') THEN
    ALTER TABLE matrix_users ADD COLUMN display_name VARCHAR(100);
  END IF;

  -- Add avatar_url if it doesn't exist
  IF NOT EXISTS (SELECT 1 FROM information_schema.columns
                 WHERE table_name = 'matrix_users' AND column_name = 'avatar_url') THEN
    ALTER TABLE matrix_users ADD COLUMN avatar_url TEXT;
  END IF;

  -- Add expires_at if it doesn't exist
  IF NOT EXISTS (SELECT 1 FROM information_schema.columns
                 WHERE table_name = 'matrix_users' AND column_name = 'expires_at') THEN
    ALTER TABLE matrix_users ADD COLUMN expires_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() + INTERVAL '30 days';
  END IF;
END $$;

-- Comments for documentation
COMMENT ON TABLE matrix_users IS 'Mapping between TrainBlink users and Matrix users';
COMMENT ON COLUMN matrix_users.access_token IS 'Encrypted Matrix access token for authentication';
COMMENT ON COLUMN matrix_users.expires_at IS 'Token expiration time (30 days from creation)';
COMMENT ON COLUMN matrix_users.is_active IS 'Whether this Matrix user is currently active';
