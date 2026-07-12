package tmdb

import (
	"context"
	"fmt"

	"github.com/xromen/movietracker/internal/domain"
)

func (c *client) SearchMulti(ctx context.Context, query string, page int) (*Paginated[domain.Media], error) {
	opts := getDefaultOpts()
	opts["page"] = fmt.Sprintf("%d", page)

	tmdbClient, err := c.requestClient(ctx)
	if err != nil {
		return nil, err
	}

	result, err := tmdbClient.GetSearchMulti(query, opts)
	if err != nil {
		return nil, fmt.Errorf("search medias multi: %w", err)
	}

	medias := make([]domain.Media, 0, len(result.Results))
	for _, result := range result.Results {
		mediaType := domain.MediaType(result.MediaType)
		if !mediaType.IsValid() {
			continue
		}

		title := ""
		if result.Title != "" {
			title = result.Title
		} else if result.Name != "" {
			title = result.Name
		}

		medias = append(medias, domain.Media{
			ID:          result.ID,
			Title:       title,
			Overview:    result.Overview,
			ReleaseDate: result.ReleaseDate,
			PosterPath:  c.getPosterPath(result.PosterPath),
			VoteAverage: result.VoteAverage,
			VoteCount:   result.VoteCount,
			Type:        mediaType,
		})
	}

	return &Paginated[domain.Media]{
		Items:      medias,
		TotalPages: min(int(result.TotalPages), maxTmdbPagesCount),
		TotalItems: int(result.TotalResults),
	}, nil
}
