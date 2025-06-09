CREATE TABLE ratings (
    id CHAR(26) NOT NULL,
    user_id CHAR(26) NOT NULL,
    movie_id CHAR(26) NOT NULL,
    score INTEGER NOT NULL,
    review TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    
    PRIMARY KEY (id),
    
    CONSTRAINT fk_ratings_user_id FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    CONSTRAINT fk_ratings_movie_id FOREIGN KEY (movie_id) REFERENCES movies(id) ON DELETE CASCADE,
    CONSTRAINT chk_score_valid CHECK (score >= 1 AND score <= 5),
    
    UNIQUE (user_id, movie_id)
);

CREATE INDEX idx_ratings_user_created ON ratings (user_id, created_at DESC);

CREATE INDEX idx_ratings_movie_id ON ratings (movie_id);

CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS '
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
' LANGUAGE plpgsql;

CREATE TRIGGER update_ratings_updated_at 
    BEFORE UPDATE ON ratings 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column(); 