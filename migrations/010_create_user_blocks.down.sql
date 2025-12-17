-- Rollback user blocks table
DROP INDEX IF EXISTS idx_user_blocks_blocker;
DROP INDEX IF EXISTS idx_user_blocks_reverse;
DROP INDEX IF EXISTS idx_user_blocks_check;
DROP TABLE IF EXISTS user_blocks CASCADE;
