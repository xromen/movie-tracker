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

type MediaPageOutput struct {
	Medias     []domain.Media
	TotalPages int
	TotalItems int
}

type SearchService interface {
	SearchMulti(ctx context.Context, query string, page int) (*MediaPageOutput, error)
}

type searchService struct {
	tmdbClient tmdb.Client
	cache      cache.Cache
	logger     *slog.Logger
}

func NewSearchService(
	tmdbClient tmdb.Client,
	cache cache.Cache,
	logger *slog.Logger,
) SearchService {
	return &searchService{
		tmdbClient: tmdbClient,
		cache:      cache,
		logger:     logger,
	}
}

func (s *searchService) SearchMulti(ctx context.Context, query string, page int) (*MediaPageOutput, error) {
	if query == "" {
		return nil, domain.NewValidationError("query", "is required")
	}

	if page < 1 {
		page = 1
	}

	cacheKey := cache.SearchMultiKey(query, page)

	var cacheResult MediaPageOutput
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

	result, err := s.tmdbClient.SearchMulti(ctx, query, page)

	if err != nil {
		return nil, fmt.Errorf("search multi: %w", err)
	}

	output := toMediaPageOutput(result)

	cache.InBackground(func() {
		cacheCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		if err := s.cache.Set(cacheCtx, cacheKey, output, searchCacheTTL); err != nil {
			s.logger.Warn("failed to cache search nulti result", "error", err)
		}
	})

	return output, nil
}

func toMediaPageOutput(result *tmdb.Paginated[domain.Media]) *MediaPageOutput {
	return &MediaPageOutput{
		Medias:     result.Items,
		TotalPages: result.TotalPages,
		TotalItems: result.TotalItems,
	}
}
