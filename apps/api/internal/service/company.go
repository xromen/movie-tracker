package service

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/xromen/movietracker/internal/domain"
	"github.com/xromen/movietracker/internal/platform/cache"
	"github.com/xromen/movietracker/internal/platform/tmdb"
)

type companyService struct {
	tmdbClient tmdb.Client
	cache      cache.Cache
	logger     *slog.Logger
}

type CompanyService interface {
	GetDetails(ctx context.Context, tmdbID int64) (*domain.CompanyDetails, error)
}

func NewCompanyService(
	tmdbClient tmdb.Client, 
	cache cache.Cache, 
	logger *slog.Logger,
) CompanyService {
	return &companyService{
		tmdbClient: tmdbClient,
		cache: cache,
		logger: logger,
	}
}

func (s companyService) GetDetails(ctx context.Context, id int64) (*domain.CompanyDetails, error) {
	cacheKey := cache.CompanyDetailsKey(id)

	var output domain.CompanyDetails
	if err := s.cache.Get(ctx, cacheKey, &output); err == nil {
		return &output, nil
	}

	company, err := s.tmdbClient.GetCompanyDetails(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get company: %w", err)
	}

	cache.InBackground(func() {
		cacheCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		if err := s.cache.Set(cacheCtx, cacheKey, company, detailCacheTTL); err != nil {
			s.logger.Warn("failed to cache company", "error", err)
		}
	})

	return company, nil
}
