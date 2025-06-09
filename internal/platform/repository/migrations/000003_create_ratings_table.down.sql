DROP TRIGGER IF EXISTS update_ratings_updated_at ON ratings;
DROP FUNCTION IF EXISTS update_ratings_updated_at_column() CASCADE;
DROP TABLE IF EXISTS ratings; 