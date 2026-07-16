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

type UserWatchListInput struct {
	UserID    int64
	Status    domain.WatchStatus
	MediaType domain.MediaType
	Page      int
	PerPage   int
}

type SetMediaUserStatusInput struct {
	UserID    int64
	Status    domain.WatchStatus
	MediaID   int64
	MediaType domain.MediaType
}

type UserWatchListOutput struct {
	Medias  []domain.UserMedia
	Total   int
	Page    int
	PerPage int
}

type mediaRepository interface {
	Upsert(ctx context.Context, media *domain.Media) error
	GetByTmdbID(ctx context.Context, tmdbID int64) (*domain.Media, error)
	GetUserList(ctx context.Context, userID int64, status domain.WatchStatus, mediaType domain.MediaType, page, perPage int) ([]domain.UserMedia, int, error)
	GetMediaUserStatus(ctx context.Context, mediaType domain.MediaType, userID, mediaID int64) (*domain.WatchStatus, error)
	DeleteUserStatus(ctx context.Context, mediaType domain.MediaType, userID, mediaID int64) (*domain.UserMedia, error)
	SetMediaUserStatus(ctx context.Context, userID int64, media *domain.UserMedia) error
}

type WatchListService interface {
	GetUserWatchList(ctx context.Context, input UserWatchListInput) (*UserWatchListOutput, error)
	DeleteMediaUserStatus(ctx context.Context, mediaType domain.MediaType, userID, mediaID int64) (*domain.UserMedia, error)
	GetMediaUserStatus(ctx context.Context, mediaType domain.MediaType, userID, mediaID int64) (*domain.WatchStatus, error)
	SetMediaUserStatus(ctx context.Context, input SetMediaUserStatusInput) (*domain.UserMedia, error)
}

type watchListService struct {
	mediaRepo mediaRepository
	tmdb      tmdb.Client
	cache     cache.Cache
	logger    *slog.Logger
}

func NewWatchListService(
	mediaRepo mediaRepository,
	tmdb tmdb.Client,
	cache cache.Cache,
	logger *slog.Logger,
) WatchListService {
	return &watchListService{
		mediaRepo: mediaRepo,
		tmdb:      tmdb,
		cache:     cache,
		logger:    logger,
	}
}

func (s *watchListService) GetUserWatchList(ctx context.Context, input UserWatchListInput) (*UserWatchListOutput, error) {
	if !input.Status.IsValid() && input.Status != "" {
		return nil, domain.NewValidationError("status", "invalid status value")
	}
	if !input.MediaType.IsValid() && input.MediaType != "" {
		return nil, domain.NewValidationError("media type", "invalid media type value")
	}
	if input.PerPage < 1 || input.PerPage > 100 {
		input.PerPage = 20
	}
	if input.Page < 1 {
		input.Page = 1
	}

	//cacheKey := cache.UserMediaListKey(
	//	input.UserID,
	//	string(input.Status),
	//	string(input.MediaType),
	//	input.Page,
	//	input.PerPage,
	//)
	//
	//var output UserWatchListOutput
	//if err := s.cache.Get(ctx, cacheKey, &output); err == nil {
	//	return &output, nil
	//}

	medias, total, err := s.mediaRepo.GetUserList(ctx, input.UserID, input.Status, input.MediaType, input.Page, input.PerPage)
	if err != nil {
		return nil, fmt.Errorf("get tv shows: %w", err)
	}

	output := UserWatchListOutput{
		Medias:  medias,
		Total:   total,
		Page:    input.Page,
		PerPage: input.PerPage,
	}

	//cache.InBackground(func() {
	//	cacheCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	//	defer cancel()
	//	if err := s.cache.Set(cacheCtx, cacheKey, output, userMoviesCacheTTL); err != nil {
	//		s.logger.Warn("failed to cache user watch list", "error", err)
	//	}
	//})

	return &output, nil
}

func (s *watchListService) DeleteMediaUserStatus(ctx context.Context, mediaType domain.MediaType, userID, mediaID int64) (*domain.UserMedia, error) {
	if !mediaType.IsValid() && mediaType != "" {
		return nil, domain.NewValidationError("media type", "invalid media type value")
	}
	
	media, err := s.mediaRepo.DeleteUserStatus(ctx, mediaType, userID, mediaID)

	if err != nil {
		return nil, fmt.Errorf("delete user status: %w", err)
	}

	cache.InBackground(func() {
		cacheCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		var key string
		if media.Media.Type == domain.MediaTypeMovie {
			key = cache.UserMoviesTemplate(userID, string(media.Status))
		} else {
			key = cache.UserTVShowsTemplate(userID, string(media.Status))
		}

		if err := s.cache.DeleteByPattern(cacheCtx, key); err != nil {
			s.logger.Warn("failed to invalidate cache", "error", err)
		}
	})

	return media, nil
}

func (s *watchListService) GetMediaUserStatus(ctx context.Context, mediaType domain.MediaType, userID, mediaID int64) (*domain.WatchStatus, error) {
	if !mediaType.IsValid() && mediaType != "" {
		return nil, domain.NewValidationError("media type", "invalid media type value")
	}
	
	status, err := s.mediaRepo.GetMediaUserStatus(ctx, mediaType, userID, mediaID)

	if err == nil {
		return status, nil
	}

	if errors.Is(err, domain.ErrNotFound) {
		return nil, nil
	} else {
		return nil, fmt.Errorf("get media user status: %w", err)
	}
}

func (s *watchListService) SetMediaUserStatus(ctx context.Context, input SetMediaUserStatusInput) (*domain.UserMedia, error) {
	if !input.Status.IsValid() {
		return nil, domain.NewValidationError("status", "invalid status value")
	}

	if !input.MediaType.IsValid() {
		return nil, domain.NewValidationError("media type", "invalid media type value")
	}

	var media *domain.Media
	var err error

	switch input.MediaType {
	case domain.MediaTypeMovie:
		media, err = s.tmdb.GetMovie(ctx, input.MediaID)

		if err != nil {
			s.logger.Warn("failed to get media movie", "error", err)
		}
	case domain.MediaTypeTV:
		media, err = s.tmdb.GetTVShow(ctx, input.MediaID)

		if err != nil {
			s.logger.Warn("failed to get media movie", "error", err)
		}
	}

	if media == nil {
		media, err = s.mediaRepo.GetByTmdbID(ctx, input.MediaID)

		if err != nil {
			return nil, fmt.Errorf("get media from tmdb: %w", err)
		}
	}

	err = s.mediaRepo.Upsert(ctx, media)

	if err != nil {
		return nil, fmt.Errorf("upsert media: %w", err)
	}

	userMedia := &domain.UserMedia{
		Status: input.Status,
		Media:  media,
		Rating: nil,
	}

	err = s.mediaRepo.SetMediaUserStatus(ctx, input.UserID, userMedia)

	if err != nil {
		return nil, fmt.Errorf("set media user status: %w", err)
	}

	return userMedia, nil
}
