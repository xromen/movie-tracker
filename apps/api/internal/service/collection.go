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
	GetDetails(ctx context.Context, id int64) (*domain.Collection, error)
}

type collectionService struct {
	tmdbClient tmdb.Client
	cache      cache.Cache
	logger     *slog.Logger
}

func NewCollectionService(
	tmdbClient tmdb.Client,
	cache cache.Cache,
	logger *slog.Logger) CollectionService {
	return &collectionService{
		tmdbClient: tmdbClient,
		cache:      cache,
		logger:     logger,
	}
}

func (s *collectionService) GetDetails(ctx context.Context, id int64) (*domain.Collection, error) {
	cacheKey := cache.CollectionKey(id)

	var cacheResult domain.Collection
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

	return result, nil
}
