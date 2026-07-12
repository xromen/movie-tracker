package tmdb

import (
	"context"
	"fmt"

	"github.com/xromen/movietracker/internal/domain"
)

func (c *client) GetCollectionDetails(ctx context.Context, id int64) (*domain.Collection, error) {
	opts := getDefaultOpts()

	tmdbClient, err := c.requestClient(ctx)
	if err != nil {
		return nil, err
	}

	result, err := tmdbClient.GetCollectionDetails(int(id), opts)
	if err != nil {
		return nil, fmt.Errorf("get collection details: %w", err)
	}

	parts := make([]domain.CollectionPart, 0, len(result.Parts))

	for _, part := range result.Parts {
		parts = append(parts, domain.CollectionPart{
			ID:          part.ID,
			Title:       part.Title,
			Overview:    part.Overview,
			ReleaseDate: part.ReleaseDate,
			PosterPath:  c.getPosterPath(part.PosterPath),
			MediaType:   part.MediaType,
			VoteAverage: part.VoteAverage,
		})
	}

	return &domain.Collection{
		ID:         result.ID,
		Name:       result.Name,
		PosterPath: c.getPosterPath(result.PosterPath),
		Overview:   result.Overview,
		Parts:      parts,
	}, nil
}
