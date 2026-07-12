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

type tvShowRepository interface {
	Upsert(ctx context.Context, movie *domain.Media) error
	GetByTmdbID(ctx context.Context, tmdbID int64) (*domain.Media, error)
	GetUserList(ctx context.Context, userID int64, status domain.WatchStatus, mediaType domain.MediaType, page, perPage int) ([]domain.UserMedia, int, error)
	GetMediaUserStatus(ctx context.Context, userID, mediaID int64) (*domain.WatchStatus, error)
	GetWatchedEpisodeNumbers(ctx context.Context, userID, tvShowID int64, seasonNumber int) ([]int, error)
	SetEpisodeWatched(ctx context.Context, userID, tvShowID int64, seasonNumber, episodeNumber int, watched bool) error
	MarkSeasonWatched(ctx context.Context, userID, tvShowID int64, seasonNumber int, episodeNumbers []int32) error
	UnmarkSeasonWatched(ctx context.Context, userID, tvShowID int64, seasonNumber int) error
}

type AddToTVShowListInput struct {
	UserID int64
	TmdbID int64
	Status domain.WatchStatus
	Rating *int
}

type TVShowPageOutput struct {
	TVShows    []domain.Media
	TotalPages int
	TotalItems int
}

type EpisodePageOutput struct {
	Episodes   []domain.Episode
	TotalPages int
	TotalItems int
}

type TVShowListOutput struct {
	TVShows []domain.UserMedia
	Total   int
	Page    int
	PerPage int
}

type TVShowService interface {
	Search(ctx context.Context, query string, page int) (*TVShowPageOutput, error)
	GetAiringToday(ctx context.Context, page int) (*TVShowPageOutput, error)
	GetOnTheAir(ctx context.Context, page int) (*TVShowPageOutput, error)
	GetPopular(ctx context.Context, page int) (*TVShowPageOutput, error)
	GetTopRated(ctx context.Context, page int) (*TVShowPageOutput, error)
	GetDetails(ctx context.Context, id int64) (*domain.TVShowDetail, error)
	GetRecommendations(ctx context.Context, id int64, page int) (*TVShowPageOutput, error)
	GetUserTVShows(ctx context.Context, userID int64, status domain.WatchStatus, page, perPage int) (*TVShowListOutput, error)
	GetSeasonEpisodes(ctx context.Context, userID *int64, tvID int64, seasonNumber int, page int) (*EpisodePageOutput, error)
	SetEpisodeWatched(ctx context.Context, userID, tvShowID int64, seasonNumber, episodeNumber int, watched bool) error
	MarkSeasonWatched(ctx context.Context, userID, tvShowID int64, seasonNumber int) error
	UnmarkSeasonWatched(ctx context.Context, userID, tvShowID int64, seasonNumber int) error
}

type tvShowService struct {
	tvShowRepo tvShowRepository
	tmdbClient tmdb.Client
	cache      cache.Cache
	logger     *slog.Logger
}

func NewTVShowService(
	tvShowRepo tvShowRepository,
	tmdbClient tmdb.Client,
	cache cache.Cache,
	logger *slog.Logger,
) TVShowService {
	return &tvShowService{
		tvShowRepo: tvShowRepo,
		tmdbClient: tmdbClient,
		cache:      cache,
		logger:     logger,
	}
}

func (s *tvShowService) Search(ctx context.Context, query string, page int) (*TVShowPageOutput, error) {
	if query == "" {
		return nil, domain.NewValidationError("query", "is required")
	}

	if page < 1 {
		page = 1
	}

	cacheKey := cache.TVShowSearchKey(query, page)

	var cacheResult TVShowPageOutput
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

	result, err := s.tmdbClient.SearchTVShows(ctx, query, page)

	if err != nil {
		return nil, fmt.Errorf("search tv shows: %w", err)
	}

	output := toTvPageOutput(result)

	cache.InBackground(func() {
		cacheCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		if err := s.cache.Set(cacheCtx, cacheKey, output, searchCacheTTL); err != nil {
			s.logger.Warn("failed to cache search tv shows result", "error", err)
		}
	})

	return output, nil
}

func (s *tvShowService) GetAiringToday(ctx context.Context, page int) (*TVShowPageOutput, error) {
	if page < 1 {
		page = 1
	}

	cacheKey := cache.TVShowAiringTodayKey(page)

	var cacheResult TVShowPageOutput
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

	result, err := s.tmdbClient.GetTVShowAiringToday(ctx, page)

	if err != nil {
		return nil, fmt.Errorf("airing today rv shows: %w", err)
	}

	output := toTvPageOutput(result)

	cache.InBackground(func() {
		cacheCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		if err := s.cache.Set(cacheCtx, cacheKey, output, searchCacheTTL); err != nil {
			s.logger.Warn("failed to cache airing today tv shows result", "error", err)
		}
	})

	return output, nil
}

func (s *tvShowService) GetPopular(ctx context.Context, page int) (*TVShowPageOutput, error) {
	if page < 1 {
		page = 1
	}

	cacheKey := cache.TVShowPopularKey(page)

	var cacheResult TVShowPageOutput
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

	result, err := s.tmdbClient.GetTVShowPopular(ctx, page)

	if err != nil {
		return nil, fmt.Errorf("get popular tv shows: %w", err)
	}

	output := toTvPageOutput(result)

	cache.InBackground(func() {
		cacheCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		if err := s.cache.Set(cacheCtx, cacheKey, output, searchCacheTTL); err != nil {
			s.logger.Warn("failed to cache popular tv shows result", "error", err)
		}
	})

	return output, nil
}

func (s *tvShowService) GetTopRated(ctx context.Context, page int) (*TVShowPageOutput, error) {
	if page < 1 {
		page = 1
	}

	cacheKey := cache.TVShowTopRatedKey(page)

	var cacheResult TVShowPageOutput
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

	result, err := s.tmdbClient.GetTVShowTopRated(ctx, page)

	if err != nil {
		return nil, fmt.Errorf("get top rated tv shows: %w", err)
	}

	output := toTvPageOutput(result)

	cache.InBackground(func() {
		cacheCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		if err := s.cache.Set(cacheCtx, cacheKey, output, searchCacheTTL); err != nil {
			s.logger.Warn("failed to cache top rated tv shows result", "error", err)
		}
	})

	return output, nil
}

func (s *tvShowService) GetOnTheAir(ctx context.Context, page int) (*TVShowPageOutput, error) {
	if page < 1 {
		page = 1
	}

	cacheKey := cache.TVShowOnTheAirKey(page)

	var cacheResult TVShowPageOutput
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

	result, err := s.tmdbClient.GetTVShowOnTheAir(ctx, page)

	if err != nil {
		return nil, fmt.Errorf("get on the air tv shows: %w", err)
	}

	output := toTvPageOutput(result)

	cache.InBackground(func() {
		cacheCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		if err := s.cache.Set(cacheCtx, cacheKey, output, searchCacheTTL); err != nil {
			s.logger.Warn("failed to cache on the air tv shows result", "error", err)
		}
	})

	return output, nil
}

// func (s *tvShowService) AddToList(ctx context.Context, input AddToTVShowListInput) (*domain.UserMedia, error) {
// 	if !input.Status.IsValid() {
// 		return nil, domain.NewValidationError("status", "invalid status value")
// 	}

// 	tvShow, err := s.tvShowRepo.GetByTmdbID(ctx, input.TmdbID)
// 	if err != nil {
// 		if !errors.Is(err, domain.ErrNotFound) {
// 			return nil, fmt.Errorf("add to list: %w", err)
// 		}

// 		tvShow, err = s.tmdbClient.GetTVShow(ctx, input.TmdbID)
// 		if err != nil {
// 			return nil, fmt.Errorf("add to list: %w", err)
// 		}

// 		if err := s.tvShowRepo.Upsert(ctx, tvShow); err != nil {
// 			return nil, fmt.Errorf("add to list, upsert tv show: %w", err)
// 		}
// 	}

// 	userTVShow := &domain.UserMedia{
// 		Media:  tvShow,
// 		Status: input.Status,
// 		Rating: input.Rating,
// 	}

// 	if err := s.tvShowRepo.AddToUserList(ctx, input.UserID, userTVShow); err != nil {
// 		return nil, fmt.Errorf("add to userlist: %w", err)
// 	}

// 	cache.InBackground(func() {
// 		cacheCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
// 		defer cancel()

// 		if err := s.cache.Delete(cacheCtx, cache.TVShowUserDetailKey(input.UserID, tvShow.ID)); err != nil {
// 			s.logger.Warn("failed to invalidate cache", "error", err)
// 		}
// 		if err := s.cache.DeleteByPattern(cacheCtx, cache.UserMediaListTemplate(input.UserID, string(input.Status), string(domain.MediaTypeTV))); err != nil {
// 			s.logger.Warn("failed to invalidate cache", "error", err)
// 		}
// 		if err := s.cache.DeleteByPattern(cacheCtx, cache.UserMediaListTemplate(input.UserID, "", string(domain.MediaTypeTV))); err != nil {
// 			s.logger.Warn("failed to invalidate cache", "error", err)
// 		}
// 	})

// 	return userTVShow, nil
// }

func (s *tvShowService) GetUserTVShows(
	ctx context.Context,
	userID int64,
	status domain.WatchStatus,
	page, perPage int,
) (*TVShowListOutput, error) {
	if !status.IsValid() && status != "" {
		return nil, domain.NewValidationError("status", "invalid status value")
	}
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}
	if page < 1 {
		page = 1
	}

	cacheKey := cache.UserTVShowsKey(userID, string(status), page, perPage)
	var output TVShowListOutput
	if err := s.cache.Get(ctx, cacheKey, &output); err == nil {
		return &output, nil
	}

	tvShows, total, err := s.tvShowRepo.GetUserList(ctx, userID, status, domain.MediaTypeTV, page, perPage)
	if err != nil {
		return nil, fmt.Errorf("get tv shows: %w", err)
	}

	result := &TVShowListOutput{
		TVShows: tvShows,
		Total:   total,
		Page:    page,
		PerPage: perPage,
	}

	cache.InBackground(func() {
		cacheCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		if err := s.cache.Set(cacheCtx, cacheKey, result, userMoviesCacheTTL); err != nil {
			s.logger.Warn("failed to cache user tv shows", "error", err)
		}
	})

	return result, nil
}

func (s *tvShowService) GetDetails(ctx context.Context, id int64) (*domain.TVShowDetail, error) {
	cacheKey := cache.TVShowDetailKey(id)

	var output domain.TVShowDetail
	if err := s.cache.Get(ctx, cacheKey, &output); err == nil {
		return &output, nil
	}

	tvShow, err := s.tmdbClient.GetTVShowDetails(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get tv show details: %w", err)
	}

	// var watchStatus *domain.WatchStatus
	// if userID != nil {
	// 	watchStatus, err = s.tvShowRepo.GetMediaUserStatus(ctx, *userID, tvShow.ID)
	// 	if err != nil {
	// 		if !errors.Is(err, domain.ErrNotFound) {
	// 			return nil, fmt.Errorf("get tv show user status: %w", err)
	// 		}
	// 	}

	// 	for i := range tvShow.Seasons {
	// 		watchedEpisodes, err := s.tvShowRepo.GetWatchedEpisodeNumbers(ctx, *userID, id, tvShow.Seasons[i].SeasonNumber)
	// 		if err != nil {
	// 			s.logger.Warn("failed to get watched episode numbers",
	// 				"error", err,
	// 				"tvshowid", id,
	// 				"seasonnumber", tvShow.Seasons[i].SeasonNumber)
	// 			continue
	// 		}

	// 		seasonWatched := len(watchedEpisodes) == tvShow.Seasons[i].EpisodeCount && tvShow.Seasons[i].EpisodeCount != 0
	// 		tvShow.Seasons[i].IsWatched = &seasonWatched
	// 	}
	// }

	cache.InBackground(func() {
		cacheCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		if err := s.cache.Set(cacheCtx, cacheKey, tvShow, detailCacheTTL); err != nil {
			s.logger.Warn("failed to cache user tv show", "error", err)
		}
	})

	return tvShow, nil
}

func (s *tvShowService) GetRecommendations(ctx context.Context, id int64, page int) (*TVShowPageOutput, error) {
	var cacheKey = cache.TVShowRecommendationsKey(id, page)

	var output TVShowPageOutput
	if err := s.cache.Get(ctx, cacheKey, &output); err == nil {
		return &output, nil
	}

	result, err := s.tmdbClient.GetTVShowRecommendations(ctx, id, page)
	if err != nil {
		return nil, fmt.Errorf("get tv show recommendations: %w", err)
	}

	output = *toTvPageOutput(result)

	cache.InBackground(func() {
		cacheCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		if err := s.cache.Set(cacheCtx, cacheKey, output, detailCacheTTL); err != nil {
			s.logger.Warn("failed to cache tv show recommendations", "error", err)
		}
	})

	return &output, nil
}

func (s *tvShowService) GetSeasonEpisodes(ctx context.Context, userID *int64, tvShowID int64, seasonNumber int, page int) (*EpisodePageOutput, error) {
	if tvShowID <= 0 {
		return nil, domain.NewValidationError("tv_show_id", "must be positive")
	}
	if seasonNumber < 0 {
		return nil, domain.NewValidationError("season_number", "must not be negative")
	}

	var cacheKey = cache.TVShowSeasonEpisodesKey(tvShowID, seasonNumber, page)

	var output *EpisodePageOutput
	if err := s.cache.Get(ctx, cacheKey, &output); err == nil {
		return s.withWatchedEpisodes(ctx, userID, tvShowID, seasonNumber, output)
	}

	result, err := s.tmdbClient.GetTvSeasonEpisodes(ctx, tvShowID, seasonNumber, page)
	if err != nil {
		return nil, fmt.Errorf("get tv show season episodes: %w", err)
	}

	cache.InBackground(func() {
		cacheCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		if err := s.cache.Set(cacheCtx, cacheKey, result, detailCacheTTL); err != nil {
			s.logger.Warn("failed to cache tv show season episodes", "error", err)
		}
	})

	return s.withWatchedEpisodes(ctx, userID, tvShowID, seasonNumber, &EpisodePageOutput{
		Episodes:   result.Items,
		TotalPages: result.TotalPages,
		TotalItems: result.TotalItems,
	})
}

func (s *tvShowService) SetEpisodeWatched(
	ctx context.Context,
	userID, tvShowID int64,
	seasonNumber, episodeNumber int,
	watched bool,
) error {
	if tvShowID <= 0 {
		return domain.NewValidationError("tv_show_id", "must be positive")
	}
	if seasonNumber < 0 {
		return domain.NewValidationError("season_number", "must not be negative")
	}
	if episodeNumber <= 0 {
		return domain.NewValidationError("episode_number", "must be positive")
	}

	if err := s.tvShowRepo.SetEpisodeWatched(
		ctx,
		userID,
		tvShowID,
		seasonNumber,
		episodeNumber,
		watched,
	); err != nil {
		return fmt.Errorf("set episode watched: %w", err)
	}

	cache.InBackground(func() {
		cacheCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := s.cache.Delete(cacheCtx, cache.TVShowUserDetailKey(userID, tvShowID)); err != nil {
			s.logger.Warn("failed to invalidate cache", "error", err)
		}
	})

	return nil
}

func (s *tvShowService) MarkSeasonWatched(ctx context.Context, userID, tvShowID int64, seasonNumber int) error {
	if tvShowID <= 0 {
		return domain.NewValidationError("tv_show_id", "must be positive")
	}
	if seasonNumber < 0 {
		return domain.NewValidationError("season_number", "must not be negative")
	}

	// nil означает: получить данные эпизодов без пользовательских is_watched.
	episodesOutput, err := s.GetSeasonEpisodes(ctx, nil, tvShowID, seasonNumber, 0)
	if err != nil {
		return fmt.Errorf("get season episodes for marking watched: %w", err)
	}

	episodeNumbers := make([]int32, 0, len(episodesOutput.Episodes))
	for _, episode := range episodesOutput.Episodes {
		episodeNumbers = append(episodeNumbers, int32(episode.EpisodeNumber))
	}

	if err := s.tvShowRepo.MarkSeasonWatched(
		ctx,
		userID,
		tvShowID,
		seasonNumber,
		episodeNumbers,
	); err != nil {
		return fmt.Errorf("mark season watched: %w", err)
	}

	cache.InBackground(func() {
		cacheCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := s.cache.Delete(cacheCtx, cache.TVShowUserDetailKey(userID, tvShowID)); err != nil {
			s.logger.Warn("failed to invalidate cache", "error", err)
		}
	})

	return nil
}

func (s *tvShowService) UnmarkSeasonWatched(ctx context.Context, userID, tvShowID int64, seasonNumber int) error {
	if tvShowID <= 0 {
		return domain.NewValidationError("tv_show_id", "must be positive")
	}
	if seasonNumber < 0 {
		return domain.NewValidationError("season_number", "must not be negative")
	}

	if err := s.tvShowRepo.UnmarkSeasonWatched(
		ctx,
		userID,
		tvShowID,
		seasonNumber,
	); err != nil {
		return fmt.Errorf("unmark season watched: %w", err)
	}

	cache.InBackground(func() {
		cacheCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := s.cache.Delete(cacheCtx, cache.TVShowUserDetailKey(userID, tvShowID)); err != nil {
			s.logger.Warn("failed to invalidate cache", "error", err)
		}
	})

	return nil
}

func (s *tvShowService) withWatchedEpisodes(
	ctx context.Context,
	userID *int64,
	tvShowID int64,
	seasonNumber int,
	output *EpisodePageOutput,
) (*EpisodePageOutput, error) {
	for i := range output.Episodes {
		output.Episodes[i].IsWatched = nil
	}

	if userID == nil {
		return output, nil
	}

	watchedNumbers, err := s.tvShowRepo.GetWatchedEpisodeNumbers(
		ctx,
		*userID,
		tvShowID,
		seasonNumber,
	)
	if err != nil {
		return nil, fmt.Errorf("get watched episodes: %w", err)
	}

	watched := make(map[int]struct{}, len(watchedNumbers))
	for _, number := range watchedNumbers {
		watched[number] = struct{}{}
	}

	for i := range output.Episodes {
		_, contains := watched[output.Episodes[i].EpisodeNumber]
		output.Episodes[i].IsWatched = &contains
	}

	return output, nil
}

func toTvPageOutput(result *tmdb.Paginated[domain.Media]) *TVShowPageOutput {
	return &TVShowPageOutput{
		TVShows:    result.Items,
		TotalPages: result.TotalPages,
		TotalItems: result.TotalItems,
	}
}
