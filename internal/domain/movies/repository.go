package movies

import "context"

// Repository defines the interface for movie data access
type Repository interface {
	Save(ctx context.Context, movie *Movie) (*Movie, error)
	GetByID(ctx context.Context, id MovieID) (*Movie, error)
	GetAll(ctx context.Context, options ...SearchOption) ([]*Movie, error)
	SearchByTitle(ctx context.Context, title string, options ...SearchOption) ([]*Movie, error)
	Count(ctx context.Context) (int64, error)
	Exists(ctx context.Context, id MovieID) (bool, error)
	GetByGenre(ctx context.Context, genre string, options ...SearchOption) ([]*Movie, error)
	GetByDirector(ctx context.Context, director string, options ...SearchOption) ([]*Movie, error)
	GetByYearRange(ctx context.Context, startYear, endYear int, options ...SearchOption) ([]*Movie, error)
}
