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

type CollectionService interface {
	GetDetails(ctx context.Context, id int64, userID *int64) (*domain.Collection, error)
}

type collectionRepository interface {
	GetMediaUserStatuses(ctx context.Context, mediaType domain.MediaType, userID int64, mediaIDs []int64) (map[int64]domain.WatchStatus, error)
}

type collectionService struct {
	tmdbClient     tmdb.Client
	collectionRepo collectionRepository
	cache          cache.Cache
	logger         *slog.Logger
}

func NewCollectionService(
	tmdbClient tmdb.Client,
	collectionRepo collectionRepository,
	cache cache.Cache,
	logger *slog.Logger) CollectionService {
	return &collectionService{
		tmdbClient:     tmdbClient,
		collectionRepo: collectionRepo,
		cache:          cache,
		logger:         logger,
	}
}

func (s *collectionService) GetDetails(ctx context.Context, id int64, userID *int64) (*domain.Collection, error) {
	cacheKey := cache.CollectionKey(id)

	var cacheResult domain.Collection
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

	var result *domain.Collection

	result, err = s.tmdbClient.GetCollectionDetails(ctx, id)

	if err != nil {
		return nil, fmt.Errorf("collection get: %w", err)
	}

	cache.InBackground(func() {
		cacheCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		if err := s.cache.Set(cacheCtx, cacheKey, result, detailCacheTTL); err != nil {
			s.logger.Warn("failed to cache search result", "error", err)
		}
	})

	return s.withWatchStatuses(ctx, result, userID), nil
}

func (s *collectionService) withWatchStatuses(ctx context.Context, out *domain.Collection, userID *int64) *domain.Collection {
	if userID == nil {
		return out
	}

	mediaIDs := make([]int64, 0, len(out.Parts))
	for _, media := range out.Parts {
		mediaIDs = append(mediaIDs, media.ID)
	}

	statuses, err := s.collectionRepo.GetMediaUserStatuses(ctx, domain.MediaTypeMovie, *userID, mediaIDs)
	if err != nil {
		s.logger.Warn("failed to get movie statuses", "error", err)
		return out
	}

	for i := range out.Parts {
		if status, ok := statuses[out.Parts[i].ID]; ok {
			out.Parts[i].WatchStatus = status
		}
	}

	return out
}
