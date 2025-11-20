-- Rollback matrix_users updates
-- Note: We don't drop the table entirely as it might have been created earlier
-- Just remove columns we added

DO $$
BEGIN
  -- Remove display_name if it exists
  IF EXISTS (SELECT 1 FROM information_schema.columns
             WHERE table_name = 'matrix_users' AND column_name = 'display_name') THEN
    ALTER TABLE matrix_users DROP COLUMN display_name;
  END IF;

  -- Remove avatar_url if it exists
  IF EXISTS (SELECT 1 FROM information_schema.columns
             WHERE table_name = 'matrix_users' AND column_name = 'avatar_url') THEN
    ALTER TABLE matrix_users DROP COLUMN avatar_url;
  END IF;

  -- Remove expires_at if it exists
  IF EXISTS (SELECT 1 FROM information_schema.columns
             WHERE table_name = 'matrix_users' AND column_name = 'expires_at') THEN
    ALTER TABLE matrix_users DROP COLUMN expires_at;
  END IF;
END $$;
