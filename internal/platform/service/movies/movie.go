package movies

import (
	"context"
	"fmt"
	"log/slog"
	"thermondo/internal/domain/movies"
	"thermondo/internal/domain/shared"
	"thermondo/internal/pkg/errors"
)

type Service interface {
	CreateMovie(ctx context.Context, req movies.CreateMovieRequest) (*movies.Movie, error)
	GetAllMovies(ctx context.Context, limit, offset int, sortBy, order string) ([]*movies.Movie, int64, error)
	GetMovieByID(ctx context.Context, id string) (*movies.Movie, error)
	SearchMovies(ctx context.Context, req movies.SearchMoviesRequest) ([]*movies.Movie, int64, error)
}

type movieService struct {
	movieRepo    movies.Repository
	idGenerator  shared.IDGenerator
	timeProvider shared.TimeProvider
	logger       *slog.Logger
}

func (m *movieService) CreateMovie(ctx context.Context, req movies.CreateMovieRequest) (*movies.Movie, error) {
	var options []movies.MovieOption

	if req.Rating != nil {
		options = append(options, movies.WithRating(*req.Rating))
	}
	if req.Budget != nil {
		options = append(options, movies.WithBudget(*req.Budget))
	}
	if req.Revenue != nil {
		options = append(options, movies.WithRevenue(*req.Revenue))
	}
	if req.IMDbID != nil {
		options = append(options, movies.WithIMDbID(*req.IMDbID))
	}
	if req.PosterURL != nil {
		options = append(options, movies.WithPosterURL(*req.PosterURL))
	}

	movie, err := movies.NewMovie(
		req.Title,
		req.Description,
		req.ReleaseYear,
		req.Genre,
		req.Director,
		req.DurationMins,
		req.Language,
		req.Country,
		m.idGenerator,
		m.timeProvider,
		options...,
	)
	if err != nil {
		m.logger.Error("Failed to create movie", "error", err)
		return nil, errors.NewBadRequestError(err.Error())
	}

	savedMovie, err := m.movieRepo.Save(ctx, movie)
	if err != nil {
		if isConflictError(err) {
			m.logger.Error("Movie with this ID already exists", "error", err)
			return nil, errors.NewConflictError("Movie with this ID already exists")
		}
		m.logger.Error("Failed to create movie", "error", err)
		return nil, errors.NewInternalError("Failed to create movie")
	}

	return savedMovie, nil
}

func (m *movieService) GetAllMovies(ctx context.Context, limit int, offset int, sortBy string, order string) ([]*movies.Movie, int64, error) {
	searchOptions := []movies.SearchOption{
		movies.WithLimit(limit),
		movies.WithOffset(offset),
		movies.WithSort(sortBy, order),
	}

	moviesList, err := m.movieRepo.GetAll(ctx, searchOptions...)
	if err != nil {
		m.logger.Error("Failed to get movies", "error", err)
		return nil, 0, errors.NewInternalError("Failed to get movies")
	}

	totalCount, err := m.movieRepo.Count(ctx)
	if err != nil {
		m.logger.Error("Failed to get movie count", "error", err)
		return nil, 0, errors.NewInternalError("Failed to get movie count")
	}

	return moviesList, totalCount, nil
}

func (m *movieService) GetMovieByID(ctx context.Context, id string) (*movies.Movie, error) {
	movie, err := m.movieRepo.GetByID(ctx, movies.MovieID(id))
	if err != nil {
		if isNotFoundError(err) {
			m.logger.Error("Movie not found", "error", err)
			return nil, errors.NewNotFoundError("Movie not found")
		}
		return nil, errors.NewInternalError("Failed to get movie")
	}

	return movie, nil
}

func (m *movieService) SearchMovies(ctx context.Context, req movies.SearchMoviesRequest) ([]*movies.Movie, int64, error) {
	searchOptions := []movies.SearchOption{
		movies.WithLimit(req.Limit),
		movies.WithOffset(req.Offset),
		movies.WithSort(req.SortBy, req.Order),
	}

	var moviesList []*movies.Movie
	var err error

	switch {
	case req.Query != "":
		moviesList, err = m.movieRepo.SearchByTitle(ctx, req.Query, searchOptions...)
	case req.Genre != "":
		moviesList, err = m.movieRepo.GetByGenre(ctx, req.Genre, searchOptions...)
	case req.Director != "":
		moviesList, err = m.movieRepo.GetByDirector(ctx, req.Director, searchOptions...)
	case req.MinYear != nil || req.MaxYear != nil:
		minYear := movies.FirstMovieYear
		maxYear := m.timeProvider.Now().Year() + movies.MaxFutureYears
		if req.MinYear != nil {
			minYear = *req.MinYear
		}
		if req.MaxYear != nil {
			maxYear = *req.MaxYear
		}
		moviesList, err = m.movieRepo.GetByYearRange(ctx, minYear, maxYear, searchOptions...)
	default:
		moviesList, err = m.movieRepo.GetAll(ctx, searchOptions...)
	}

	if err != nil {
		m.logger.Error("Failed to search movies", "error", err)
		return nil, 0, errors.NewInternalError("Failed to search movies")
	}

	totalCount, err := m.movieRepo.Count(ctx)
	if err != nil {
		m.logger.Error("Failed to get movie count", "error", err)
		return nil, 0, errors.NewInternalError("Failed to get movie count")
	}

	return moviesList, totalCount, nil
}

func NewMovieService(movieRepo movies.Repository,
	idGenerator shared.IDGenerator,
	timeProvider shared.TimeProvider,
	logger *slog.Logger,
) Service {
	return &movieService{
		movieRepo:    movieRepo,
		idGenerator:  idGenerator,
		timeProvider: timeProvider,
		logger:       logger,
	}
}

func isConflictError(err error) bool {
	return fmt.Sprintf("%v", err) == "movie with ID already exists"
}

func isNotFoundError(err error) bool {
	return fmt.Sprintf("%v", err) == "not found"
}
