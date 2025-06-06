CREATE TABLE movies (
    id CHAR(26) NOT NULL,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    release_year INTEGER NOT NULL,
    genre VARCHAR(100) NOT NULL,
    director VARCHAR(255) NOT NULL,
    duration_mins INTEGER NOT NULL,
    rating VARCHAR(10),
    language VARCHAR(100) NOT NULL,
    country VARCHAR(100) NOT NULL,
    budget BIGINT,
    revenue BIGINT,
    imdb_id VARCHAR(20),
    poster_url TEXT, 
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    
    PRIMARY KEY (id),
    

    CONSTRAINT chk_title_not_empty CHECK (TRIM(title) != ''),
    CONSTRAINT chk_genre_not_empty CHECK (TRIM(genre) != ''),
    CONSTRAINT chk_director_not_empty CHECK (TRIM(director) != ''),
    CONSTRAINT chk_language_not_empty CHECK (TRIM(language) != ''),
    CONSTRAINT chk_country_not_empty CHECK (TRIM(country) != ''),
    CONSTRAINT chk_duration_positive CHECK (duration_mins > 0),
    CONSTRAINT chk_release_year_valid CHECK (
        release_year >= 1888 AND 
        release_year <= EXTRACT(YEAR FROM NOW()) + 5
    ),
    CONSTRAINT chk_budget_non_negative CHECK (budget IS NULL OR budget >= 0),
    CONSTRAINT chk_revenue_non_negative CHECK (revenue IS NULL OR revenue >= 0),
    CONSTRAINT chk_imdb_id_format CHECK (
        imdb_id IS NULL OR 
        imdb_id ~ '^tt[0-9]{7,8}$'
    )
);
CREATE INDEX idx_movies_created_at ON movies (created_at DESC);
CREATE INDEX idx_movies_title_lower ON movies (LOWER(title));
CREATE INDEX idx_movies_genre ON movies (genre);
CREATE INDEX idx_movies_director_lower ON movies (LOWER(director));
CREATE INDEX idx_movies_release_year ON movies (release_year);

CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$ language 'plpgsql';

CREATE TRIGGER update_movies_updated_at 
    BEFORE UPDATE ON movies 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();