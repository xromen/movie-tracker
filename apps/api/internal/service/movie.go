package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/xromen/movietracker/internal/domain"
	"github.com/xromen/movietracker/internal/platform/cache"
	"github.com/xromen/movietracker/internal/platform/tmdb"
)

const (
	searchCacheTTL     = 24 * time.Hour
	detailCacheTTL     = 24 * time.Hour
	userMoviesCacheTTL = 24 * time.Hour
)

type movieRepository interface {
	Upsert(ctx context.Context, movie *domain.Media) error
	GetByTmdbID(ctx context.Context, tmdbID int64, mediaType domain.MediaType) (*domain.Media, error)
	GetMediaUserStatus(ctx context.Context, mediaType domain.MediaType, userID, mediaID int64) (*domain.WatchStatus, error)
	GetMediaUserStatuses(ctx context.Context, mediaType domain.MediaType, userID int64, mediaIDs []int64) (map[int64]domain.WatchStatus, error)
}

type AddToListInput struct {
	UserID int64
	TmdbID int64
	Status domain.WatchStatus
	Rating *int
}

type MoviePageOutput struct {
	Movies     []domain.Media
	TotalPages int
	TotalItems int
}

type MovieListOutput struct {
	Movies  []domain.UserMedia
	Total   int
	Page    int
	PerPage int
}

type MovieService interface {
	Search(ctx context.Context, query string, page int) (*MoviePageOutput, error)
	GetNowPlaying(ctx context.Context, userID *int64, page int) (*MoviePageOutput, error)
	GetPopular(ctx context.Context, userID *int64, page int) (*MoviePageOutput, error)
	GetTopRated(ctx context.Context, userID *int64, page int) (*MoviePageOutput, error)
	GetUpcoming(ctx context.Context, userID *int64, page int) (*MoviePageOutput, error)
	GetDetails(ctx context.Context, id int64) (*domain.MovieDetail, error)
	GetRecommendations(ctx context.Context, userID *int64, id int64, page int) (*MoviePageOutput, error)
}

type movieService struct {
	movieRepo  movieRepository
	tmdbClient tmdb.Client
	cache      cache.Cache
	logger     *slog.Logger
}

func NewMovieService(
	movieRepo movieRepository,
	tmdbClient tmdb.Client,
	cache cache.Cache,
	logger *slog.Logger,
) MovieService {
	return &movieService{
		movieRepo:  movieRepo,
		tmdbClient: tmdbClient,
		cache:      cache,
		logger:     logger,
	}
}

func (s *movieService) Search(ctx context.Context, query string, page int) (*MoviePageOutput, error) {
	if query == "" {
		return nil, domain.NewValidationError("query", "is required")
	}

	if page < 1 {
		page = 1
	}

	cacheKey := cache.MovieSearchKey(query, page)

	var cacheResult MoviePageOutput
	err := s.cache.Get(ctx, cacheKey, &cacheResult)
	if err == nil {
		return &cacheResult, nil
	}

	if !errors.Is(err, cache.ErrCacheMiss) {
		s.logger.Warn("cache get failed, falling back to tmdb",
			"key", cacheKey,
			"error", err,
		)
	}

	result, err := s.tmdbClient.SearchMovies(ctx, query, page)

	if err != nil {
		return nil, fmt.Errorf("search movies: %w", err)
	}

	output := toMoviePageOutput(result)

	cache.InBackground(func() {
		cacheCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		if err := s.cache.Set(cacheCtx, cacheKey, output, searchCacheTTL); err != nil {
			s.logger.Warn("failed to cache search result", "error", err)
		}
	})

	return output, nil
}

func (s *movieService) GetNowPlaying(ctx context.Context, userID *int64, page int) (*MoviePageOutput, error) {
	if page < 1 {
		page = 1
	}

	cacheKey := cache.MovieNowPlayingKey(page)

	var cacheResult MoviePageOutput
	err := s.cache.Get(ctx, cacheKey, &cacheResult)
	if err == nil {
		return s.withWatchStatuses(ctx, &cacheResult, userID), nil
	}

	if !errors.Is(err, cache.ErrCacheMiss) {
		s.logger.Warn("cache get failed, falling back to tmdb",
			"key", cacheKey,
			"error", err,
		)
	}

	result, err := s.tmdbClient.GetMovieNowPlaying(ctx, page)

	if err != nil {
		return nil, fmt.Errorf("now playing movies: %w", err)
	}

	output := toMoviePageOutput(result)

	cache.InBackground(func() {
		cacheCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		if err := s.cache.Set(cacheCtx, cacheKey, output, searchCacheTTL); err != nil {
			s.logger.Warn("failed to cache search result", "error", err)
		}
	})

	return s.withWatchStatuses(ctx, output, userID), nil
}

func (s *movieService) GetPopular(ctx context.Context, userID *int64, page int) (*MoviePageOutput, error) {
	if page < 1 {
		page = 1
	}

	cacheKey := cache.MoviePopularKey(page)

	var cacheResult MoviePageOutput
	err := s.cache.Get(ctx, cacheKey, &cacheResult)
	if err == nil {
		return s.withWatchStatuses(ctx, &cacheResult, userID), nil
	}

	if !errors.Is(err, cache.ErrCacheMiss) {
		s.logger.Warn("cache get failed, falling back to tmdb",
			"key", cacheKey,
			"error", err,
		)
	}

	result, err := s.tmdbClient.GetMoviePopular(ctx, page)

	if err != nil {
		return nil, fmt.Errorf("get popular movies: %w", err)
	}

	output := toMoviePageOutput(result)

	cache.InBackground(func() {
		cacheCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		if err := s.cache.Set(cacheCtx, cacheKey, output, searchCacheTTL); err != nil {
			s.logger.Warn("failed to cache search result", "error", err)
		}
	})

	return s.withWatchStatuses(ctx, output, userID), nil
}

func (s *movieService) GetTopRated(ctx context.Context, userID *int64, page int) (*MoviePageOutput, error) {
	if page < 1 {
		page = 1
	}

	cacheKey := cache.MovieTopRatedKey(page)

	var cacheResult MoviePageOutput
	err := s.cache.Get(ctx, cacheKey, &cacheResult)
	if err == nil {
		return s.withWatchStatuses(ctx, &cacheResult, userID), nil
	}

	if !errors.Is(err, cache.ErrCacheMiss) {
		s.logger.Warn("cache get failed, falling back to tmdb",
			"key", cacheKey,
			"error", err,
		)
	}

	result, err := s.tmdbClient.GetMovieTopRated(ctx, page)

	if err != nil {
		return nil, fmt.Errorf("get top rated movies: %w", err)
	}

	output := toMoviePageOutput(result)

	cache.InBackground(func() {
		cacheCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		if err := s.cache.Set(cacheCtx, cacheKey, output, searchCacheTTL); err != nil {
			s.logger.Warn("failed to cache search result", "error", err)
		}
	})

	return s.withWatchStatuses(ctx, output, userID), nil
}

func (s *movieService) GetUpcoming(ctx context.Context, userID *int64, page int) (*MoviePageOutput, error) {
	if page < 1 {
		page = 1
	}

	cacheKey := cache.MovieUpcomingKey(page)

	var cacheResult MoviePageOutput
	err := s.cache.Get(ctx, cacheKey, &cacheResult)
	if err == nil {
		return s.withWatchStatuses(ctx, &cacheResult, userID), nil
	}

	if !errors.Is(err, cache.ErrCacheMiss) {
		s.logger.Warn("cache get failed, falling back to tmdb",
			"key", cacheKey,
			"error", err,
		)
	}

	result, err := s.tmdbClient.GetMovieUpcoming(ctx, page)

	if err != nil {
		return nil, fmt.Errorf("get upcoming movies: %w", err)
	}

	output := toMoviePageOutput(result)

	cache.InBackground(func() {
		cacheCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		if err := s.cache.Set(cacheCtx, cacheKey, output, searchCacheTTL); err != nil {
			s.logger.Warn("failed to cache search result", "error", err)
		}
	})

	return s.withWatchStatuses(ctx, output, userID), nil
}

func (s *movieService) GetDetails(ctx context.Context, id int64) (*domain.MovieDetail, error) {
	var cacheKey string

	cacheKey = cache.MovieDetailKey(id)

	var output domain.MovieDetail
	if err := s.cache.Get(ctx, cacheKey, &output); err == nil {
		return &output, nil
	}

	movie, err := s.tmdbClient.GetMovieDetails(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get movie: %w", err)
	}

	cache.InBackground(func() {
		cacheCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		if err := s.cache.Set(cacheCtx, cacheKey, movie, detailCacheTTL); err != nil {
			s.logger.Warn("failed to cache user movies", "error", err)
		}
	})

	return movie, nil
}

func (s *movieService) GetRecommendations(ctx context.Context, userID *int64, id int64, page int) (*MoviePageOutput, error) {
	var cacheKey = cache.MovieRecommendationsKey(id, page)

	var cacheResult MoviePageOutput
	if err := s.cache.Get(ctx, cacheKey, &cacheResult); err == nil {
		return s.withWatchStatuses(ctx, &cacheResult, userID), nil
	}

	result, err := s.tmdbClient.GetMovieRecommendations(ctx, id, page)
	if err != nil {
		return nil, fmt.Errorf("get recommendations: %w", err)
	}

	output := toMoviePageOutput(result)

	cache.InBackground(func() {
		cacheCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		if err := s.cache.Set(cacheCtx, cacheKey, output, detailCacheTTL); err != nil {
			s.logger.Warn("failed to cache movie recommendations", "error", err)
		}
	})

	return s.withWatchStatuses(ctx, output, userID), nil
}

func (s *movieService) withWatchStatuses(ctx context.Context, out *MoviePageOutput, userID *int64) *MoviePageOutput {
	if userID == nil {
		return out
	}

	mediaIDs := make([]int64, 0, len(out.Movies))
	for _, movie := range out.Movies {
		mediaIDs = append(mediaIDs, movie.ID)
	}

	statuses, err := s.movieRepo.GetMediaUserStatuses(ctx, domain.MediaTypeMovie, *userID, mediaIDs)
	if err != nil {
		s.logger.Warn("failed to get movie statuses", "error", err)
		return out
	}

	for i := range out.Movies {
		if status, ok := statuses[out.Movies[i].ID]; ok {
			out.Movies[i].WatchStatus = status
		}
	}

	return out
}

func toMoviePageOutput(result *tmdb.Paginated[domain.Media]) *MoviePageOutput {
	return &MoviePageOutput{
		Movies:     result.Items,
		TotalPages: result.TotalPages,
		TotalItems: result.TotalItems,
	}
}
