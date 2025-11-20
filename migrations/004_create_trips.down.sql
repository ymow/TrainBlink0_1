-- Drop trips table and related objects
DROP TRIGGER IF EXISTS trips_updated_at_trigger ON trips;
DROP FUNCTION IF EXISTS update_trips_updated_at();
DROP TABLE IF EXISTS trips CASCADE;
