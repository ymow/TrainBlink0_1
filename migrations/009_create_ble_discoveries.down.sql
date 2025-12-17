-- Rollback BLE discoveries table
DROP INDEX IF EXISTS idx_ble_discoveries_trip;
DROP INDEX IF EXISTS idx_ble_discoveries_active;
DROP INDEX IF EXISTS idx_ble_discoveries_expired;
DROP INDEX IF EXISTS idx_ble_discoveries_bidirectional;
DROP INDEX IF EXISTS idx_ble_discoveries_permission_check;
DROP TABLE IF EXISTS ble_discoveries CASCADE;
