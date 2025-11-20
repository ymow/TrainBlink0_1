ALTER TABLE users DROP CONSTRAINT IF EXISTS fk_users_banned_by;
DROP TRIGGER IF EXISTS update_roles_updated_at ON roles;
DROP TRIGGER IF EXISTS update_admins_updated_at ON admins;
DROP TABLE IF EXISTS roles CASCADE;
DROP TABLE IF EXISTS admins CASCADE;
