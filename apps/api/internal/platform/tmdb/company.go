package tmdb

import (
	"context"
	"fmt"

	"github.com/xromen/movietracker/internal/domain"
)

func (c *client) GetCompanyDetails(ctx context.Context, tmdbID int64) (*domain.CompanyDetails, error) {
	tmdbClient, err := c.requestClient(ctx)
	if err != nil {
		return nil, err
	}

	result, err := tmdbClient.GetCompanyDetails(int(tmdbID))
	if err != nil {
		return nil, fmt.Errorf("get company details: %w", err)
	}

	return &domain.CompanyDetails{
		ID:            result.ID,
		Name:          result.Name,
		Description:   result.Description,
		LogoPath:      c.getPosterPath(result.LogoPath),
		OriginCountry: result.OriginCountry,
		Homepage:      result.Homepage,
		Headquarters:  result.Headquarters,
	}, nil
}
