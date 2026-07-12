package tmdb

import (
	"context"
	"fmt"

	"github.com/xromen/movietracker/internal/domain"
)

func (c *client) SearchMovies(ctx context.Context, query string, page int) (*Paginated[domain.Media], error) {
	opts := getDefaultOpts()
	opts["page"] = fmt.Sprintf("%d", page)

	tmdbClient, err := c.requestClient(ctx)
	if err != nil {
		return nil, err
	}

	results, err := tmdbClient.GetSearchMovies(query, opts)
	if err != nil {
		return nil, fmt.Errorf("search movies: %w", err)
	}

	movies := make([]domain.Media, 0, len(results.Results))
	for _, result := range results.Results {
		movies = append(movies, domain.Media{
			ID:          result.ID,
			Title:       result.Title,
			Overview:    result.Overview,
			ReleaseDate: result.ReleaseDate,
			PosterPath:  c.getPosterPath(result.PosterPath),
			VoteAverage: result.VoteAverage,
			VoteCount:   result.VoteCount,
			Type:        domain.MediaTypeMovie,
		})
	}

	return &Paginated[domain.Media]{
		Items:      movies,
		TotalPages: min(int(results.TotalPages), maxTmdbPagesCount),
		TotalItems: int(results.TotalResults),
	}, nil
}

func (c *client) GetMovie(ctx context.Context, tmdbId int64) (*domain.Media, error) {
	opts := getDefaultOpts()

	tmdbClient, err := c.requestClient(ctx)
	if err != nil {
		return nil, err
	}

	result, err := tmdbClient.GetMovieDetails(int(tmdbId), opts)
	if err != nil {
		return nil, fmt.Errorf("get movie: %w", err)
	}

	movie := domain.Media{
		ID:          result.ID,
		Title:       result.Title,
		Overview:    result.Overview,
		PosterPath:  c.getPosterPath(result.PosterPath),
		ReleaseDate: result.ReleaseDate,
		VoteAverage: result.VoteAverage,
		VoteCount:   result.VoteCount,
		Type:        domain.MediaTypeMovie,
	}

	return &movie, nil
}

func (c *client) GetMovieNowPlaying(ctx context.Context, page int) (*Paginated[domain.Media], error) {
	opts := getDefaultOpts()
	opts["page"] = fmt.Sprintf("%d", page)

	tmdbClient, err := c.requestClient(ctx)
	if err != nil {
		return nil, err
	}

	result, err := tmdbClient.GetMovieNowPlaying(opts)
	if err != nil {
		return nil, fmt.Errorf("get movie now playing: %w", err)
	}

	movies := make([]domain.Media, 0, len(result.Results))
	for _, result := range result.Results {
		movies = append(movies, domain.Media{
			ID:          result.ID,
			Title:       result.Title,
			Overview:    result.Overview,
			ReleaseDate: result.ReleaseDate,
			PosterPath:  c.getPosterPath(result.PosterPath),
			VoteAverage: result.VoteAverage,
			VoteCount:   result.VoteCount,
			Type:        domain.MediaTypeMovie,
		})
	}

	return &Paginated[domain.Media]{
		Items:      movies,
		TotalPages: min(int(result.TotalPages), maxTmdbPagesCount),
		TotalItems: int(result.TotalResults),
	}, nil
}

func (c *client) GetMoviePopular(ctx context.Context, page int) (*Paginated[domain.Media], error) {
	opts := getDefaultOpts()
	opts["page"] = fmt.Sprintf("%d", page)

	tmdbClient, err := c.requestClient(ctx)
	if err != nil {
		return nil, err
	}

	result, err := tmdbClient.GetMoviePopular(opts)
	if err != nil {
		return nil, fmt.Errorf("get movie popular: %w", err)
	}

	movies := make([]domain.Media, 0, len(result.Results))
	for _, result := range result.Results {
		movies = append(movies, domain.Media{
			ID:          result.ID,
			Title:       result.Title,
			Overview:    result.Overview,
			ReleaseDate: result.ReleaseDate,
			PosterPath:  c.getPosterPath(result.PosterPath),
			VoteAverage: result.VoteAverage,
			VoteCount:   result.VoteCount,
			Type:        domain.MediaTypeMovie,
		})
	}

	return &Paginated[domain.Media]{
		Items:      movies,
		TotalPages: min(int(result.TotalPages), maxTmdbPagesCount),
		TotalItems: int(result.TotalResults),
	}, nil
}

func (c *client) GetMovieTopRated(ctx context.Context, page int) (*Paginated[domain.Media], error) {
	opts := getDefaultOpts()
	opts["page"] = fmt.Sprintf("%d", page)

	tmdbClient, err := c.requestClient(ctx)
	if err != nil {
		return nil, err
	}

	result, err := tmdbClient.GetMovieTopRated(opts)
	if err != nil {
		return nil, fmt.Errorf("get movie top rated: %w", err)
	}

	movies := make([]domain.Media, 0, len(result.Results))
	for _, result := range result.Results {
		movies = append(movies, domain.Media{
			ID:          result.ID,
			Title:       result.Title,
			Overview:    result.Overview,
			ReleaseDate: result.ReleaseDate,
			PosterPath:  c.getPosterPath(result.PosterPath),
			VoteAverage: result.VoteAverage,
			VoteCount:   result.VoteCount,
			Type:        domain.MediaTypeMovie,
		})
	}

	return &Paginated[domain.Media]{
		Items:      movies,
		TotalPages: min(int(result.TotalPages), maxTmdbPagesCount),
		TotalItems: int(result.TotalResults),
	}, nil
}

func (c *client) GetMovieUpcoming(ctx context.Context, page int) (*Paginated[domain.Media], error) {
	opts := getDefaultOpts()
	opts["page"] = fmt.Sprintf("%d", page)

	tmdbClient, err := c.requestClient(ctx)
	if err != nil {
		return nil, err
	}

	result, err := tmdbClient.GetMovieUpcoming(opts)
	if err != nil {
		return nil, fmt.Errorf("get movie upcoming: %w", err)
	}

	movies := make([]domain.Media, 0, len(result.Results))
	for _, result := range result.Results {
		movies = append(movies, domain.Media{
			ID:          result.ID,
			Title:       result.Title,
			Overview:    result.Overview,
			ReleaseDate: result.ReleaseDate,
			PosterPath:  c.getPosterPath(result.PosterPath),
			VoteAverage: result.VoteAverage,
			VoteCount:   result.VoteCount,
			Type:        domain.MediaTypeMovie,
		})
	}

	return &Paginated[domain.Media]{
		Items:      movies,
		TotalPages: min(int(result.TotalPages), maxTmdbPagesCount),
		TotalItems: int(result.TotalResults),
	}, nil
}

func (c *client) GetMovieDetails(ctx context.Context, tmdbId int64) (*domain.MovieDetail, error) {
	opts := getDefaultOpts()
	opts["append_to_response"] = "videos"

	tmdbClient, err := c.requestClient(ctx)
	if err != nil {
		return nil, err
	}

	result, err := tmdbClient.GetMovieDetails(int(tmdbId), opts)
	if err != nil {
		return nil, fmt.Errorf("get movie details: %w", err)
	}

	var collectionID *int64
	if result.BelongsToCollection.ID == 0 {
		collectionID = nil
	} else {
		collectionID = &result.BelongsToCollection.ID
	}

	productionCountries := make([]string, 0, len(result.ProductionCountries))
	for _, prod := range result.ProductionCountries {
		productionCountries = append(productionCountries, prod.Name)
	}

	return &domain.MovieDetail{
		ID:                  result.ID,
		Title:               result.Title,
		Overview:            result.Overview,
		ReleaseDate:         result.ReleaseDate,
		PosterPath:          c.getPosterPath(result.PosterPath),
		Genres:              toDomainGenres(result.Genres),
		OriginalLanguage:    result.OriginalLanguage,
		OriginCountry:       result.OriginCountry,
		OriginalTitle:       result.OriginalTitle,
		ProductionCountries: productionCountries,
		Popularity:          result.Popularity,
		Status:              result.Status,
		Videos:              toDomainViedos(result.Videos),
		VoteAverage:         result.VoteAverage,
		VoteCount:           result.VoteCount,
		Runtime:             result.Runtime,
		Budget:              result.Budget,
		Revenue:             result.Revenue,
		CollectionID:        collectionID,
	}, nil
}

func (c *client) GetMovieRecommendations(ctx context.Context, tmdbId int64, page int) (*Paginated[domain.Media], error) {
	opts := getDefaultOpts()
	opts["page"] = fmt.Sprintf("%d", page)

	tmdbClient, err := c.requestClient(ctx)
	if err != nil {
		return nil, err
	}

	result, err := tmdbClient.GetMovieRecommendations(int(tmdbId), opts)
	if err != nil {
		return nil, fmt.Errorf("get movie recommendations: %w", err)
	}

	var movies []domain.Media
	for _, result := range result.Results {
		movies = append(movies, domain.Media{
			ID:          result.ID,
			Title:       result.Title,
			Overview:    result.Overview,
			ReleaseDate: result.ReleaseDate,
			PosterPath:  c.getPosterPath(result.PosterPath),
			VoteAverage: result.VoteAverage,
			VoteCount:   result.VoteCount,
			Type:        domain.MediaTypeMovie,
		})
	}

	return &Paginated[domain.Media]{
		Items:      movies,
		TotalPages: min(int(result.TotalPages), maxTmdbPagesCount),
		TotalItems: int(result.TotalResults),
	}, nil
}
