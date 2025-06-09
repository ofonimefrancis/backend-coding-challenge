DROP TRIGGER IF EXISTS update_users_updated_at ON users;
DROP FUNCTION IF EXISTS update_users_updated_at_column() CASCADE;
DROP TABLE IF EXISTS users; 