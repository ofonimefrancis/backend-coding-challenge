package movies

import (
	"errors"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"strings"
	"thermondo/internal/domain/movies"
	appErrors "thermondo/internal/pkg/errors"
	"thermondo/internal/pkg/http/response"
	movieService "thermondo/internal/platform/service/movies"
	"time"

	"github.com/go-chi/chi/v5"
)

const (
	DefaultLimit  = 20
	InitialOffset = 0
)

type Handler struct {
	movieService   movieService.Service
	logger         *slog.Logger
	responseWriter *response.Writer
}

func NewHandler(movieService movieService.Service, logger *slog.Logger) *Handler {
	if logger == nil {
		logger = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelInfo,
		}))
	}

	return &Handler{
		movieService:   movieService,
		logger:         logger,
		responseWriter: response.NewWriter(logger),
	}
}

func (h *Handler) RegisterRoutes(router chi.Router) {
	router.Route("/movies", func(r chi.Router) {
		r.Post("/", h.CreateMovie)
		r.Get("/", h.GetAllMovies)
		r.Get("/search", h.SearchMovies)

		r.Route("/{id}", func(r chi.Router) {
			r.Get("/", h.GetMovie)
		})
	})
}

func (h *Handler) handleServiceError(w http.ResponseWriter, err error) {
	var appErr *appErrors.AppError
	if errors.As(err, &appErr) {

		h.responseWriter.WriteError(w, appErr.Message, appErr.StatusCode)
		return
	}

	h.responseWriter.WriteError(w, "Internal server error", http.StatusInternalServerError)
}

type listParams struct {
	Limit  int
	Offset int
	SortBy string
	Order  string
}

func (h *Handler) parseListParams(r *http.Request) (*listParams, error) {
	params := &listParams{
		Limit:  DefaultLimit,  // Default limit
		Offset: InitialOffset, // Default offset
		SortBy: "created_at",  // Default sort
		Order:  "desc",        // Default order
	}

	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		limit, err := strconv.Atoi(limitStr)
		if err != nil || limit < 1 || limit > 100 {
			h.logger.Error("[parse_list_params] Invalid limit", "error", err)
			return nil, errors.New("limit must be between 1 and 100")
		}
		params.Limit = limit
	}

	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		offset, err := strconv.Atoi(offsetStr)
		if err != nil || offset < 0 {
			h.logger.Error("[parse_list_params] Invalid offset", "error", err)
			return nil, errors.New("offset must be non-negative")
		}
		params.Offset = offset
	}

	if sortBy := r.URL.Query().Get("sort_by"); sortBy != "" {
		if !h.isValidSortField(sortBy) {
			h.logger.Error("[parse_list_params] Invalid sort_by field", "sort_by", sortBy)
			return nil, errors.New("invalid sort_by field")
		}
		params.SortBy = sortBy
	}

	if order := r.URL.Query().Get("order"); order != "" {
		order = strings.ToLower(order)
		if order != "asc" && order != "desc" {
			h.logger.Error("[parse_list_params] Invalid order", "order", order)
			return nil, errors.New("order must be 'asc' or 'desc'")
		}
		params.Order = order
	}

	return params, nil
}

func (h *Handler) parseSearchParams(r *http.Request) (*movies.SearchMoviesRequest, error) {
	listParams, err := h.parseListParams(r)
	if err != nil {
		return nil, err
	}

	searchParams := &movies.SearchMoviesRequest{
		Limit:  listParams.Limit,
		Offset: listParams.Offset,
		SortBy: listParams.SortBy,
		Order:  listParams.Order,
	}

	searchParams.Query = strings.TrimSpace(r.URL.Query().Get("q"))
	searchParams.Genre = strings.TrimSpace(r.URL.Query().Get("genre"))
	searchParams.Director = strings.TrimSpace(r.URL.Query().Get("director"))

	if minYearStr := r.URL.Query().Get("min_year"); minYearStr != "" {
		minYear, err := strconv.Atoi(minYearStr)
		if err != nil || minYear < movies.FirstMovieYear || minYear > time.Now().Year()+movies.MaxFutureYears {
			h.logger.Error("[parse_search_params] Invalid min_year", "error", err)
			return nil, errors.New("min_year must be a valid year between 1888 and current year + 5")
		}
		searchParams.MinYear = &minYear
	}

	if maxYearStr := r.URL.Query().Get("max_year"); maxYearStr != "" {
		maxYear, err := strconv.Atoi(maxYearStr)
		if err != nil || maxYear < movies.FirstMovieYear || maxYear > time.Now().Year()+movies.MaxFutureYears {
			h.logger.Error("[parse_search_params] Invalid max_year", "error", err)
			return nil, errors.New("max_year must be a valid year between 1888 and current year + 5")
		}
		searchParams.MaxYear = &maxYear
	}

	if searchParams.MinYear != nil && searchParams.MaxYear != nil {
		if *searchParams.MinYear > *searchParams.MaxYear {
			h.logger.Error("[parse_search_params] Min year cannot be greater than max year")
			return nil, errors.New("min_year cannot be greater than max_year")
		}
	}

	return searchParams, nil
}

func (h *Handler) isValidSortField(field string) bool {
	validFields := map[string]bool{
		"title":        true,
		"release_year": true,
		"created_at":   true,
		"updated_at":   true,
		"genre":        true,
		"director":     true,
	}
	return validFields[field]
}

// Response transformation methods
func (h *Handler) moviesToResponse(moviesList []*movies.Movie) []MovieResponse {
	responses := make([]MovieResponse, len(moviesList))
	for i, movie := range moviesList {
		responses[i] = h.movieToResponse(movie)
	}
	return responses
}

func (h *Handler) movieToResponse(movie *movies.Movie) MovieResponse {
	return MovieResponse{
		ID:           string(movie.ID),
		Title:        movie.Title,
		Description:  movie.Description,
		ReleaseYear:  movie.ReleaseYear,
		Genre:        movie.Genre,
		Director:     movie.Director,
		DurationMins: movie.DurationMins,
		Rating:       string(movie.Rating),
		Language:     movie.Language,
		Country:      movie.Country,
		Budget:       movie.Budget,
		Revenue:      movie.Revenue,
		IMDbID:       movie.IMDbID,
		PosterURL:    movie.PosterURL,
		CreatedAt:    movie.CreatedAt.Format(time.RFC3339),
		UpdatedAt:    movie.UpdatedAt.Format(time.RFC3339),
	}
}
