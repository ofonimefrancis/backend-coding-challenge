DROP TRIGGER IF EXISTS update_movies_updated_at ON movies;
DROP FUNCTION IF EXISTS update_updated_at_column() CASCADE;
DROP TABLE IF EXISTS movies; 