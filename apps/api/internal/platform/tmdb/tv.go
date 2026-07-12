package tmdb

import (
	"context"
	"fmt"
	"math"

	"github.com/xromen/movietracker/internal/domain"
)

func (c *client) SearchTVShows(ctx context.Context, query string, page int) (*Paginated[domain.Media], error) {
	opts := getDefaultOpts()

	opts["page"] = fmt.Sprintf("%d", page)

	tmdbClient, err := c.requestClient(ctx)
	if err != nil {
		return nil, err
	}

	results, err := tmdbClient.GetSearchTVShow(query, opts)
	if err != nil {
		return nil, fmt.Errorf("search tv shows: %w", err)
	}

	tvShows := make([]domain.Media, 0, len(results.Results))
	for _, result := range results.Results {
		tvShows = append(tvShows, domain.Media{
			ID:          result.ID,
			Title:       result.Name,
			Overview:    result.Overview,
			ReleaseDate: result.FirstAirDate,
			PosterPath:  c.getPosterPath(result.PosterPath),
			VoteAverage: result.VoteAverage,
			VoteCount:   result.VoteCount,
			Type:        domain.MediaTypeTV,
		})
	}

	return &Paginated[domain.Media]{
		Items:      tvShows,
		TotalPages: min(int(results.TotalPages), maxTmdbPagesCount),
		TotalItems: int(results.TotalResults),
	}, nil
}

func (c *client) GetTVShow(ctx context.Context, tmdbId int64) (*domain.Media, error) {
	opts := getDefaultOpts()

	tmdbClient, err := c.requestClient(ctx)
	if err != nil {
		return nil, err
	}

	result, err := tmdbClient.GetTVDetails(int(tmdbId), opts)
	if err != nil {
		return nil, fmt.Errorf("get tv show: %w", err)
	}

	tv := domain.Media{
		ID:          result.ID,
		Title:       result.Name,
		Overview:    result.Overview,
		PosterPath:  c.getPosterPath(result.PosterPath),
		ReleaseDate: result.FirstAirDate,
		VoteAverage: result.VoteAverage,
		VoteCount:   result.VoteCount,
		Type:        domain.MediaTypeTV,
	}

	return &tv, nil
}

func (c *client) GetTVShowAiringToday(ctx context.Context, page int) (*Paginated[domain.Media], error) {
	opts := getDefaultOpts()
	opts["page"] = fmt.Sprintf("%d", page)

	tmdbClient, err := c.requestClient(ctx)
	if err != nil {
		return nil, err
	}

	result, err := tmdbClient.GetTVAiringToday(opts)
	if err != nil {
		return nil, fmt.Errorf("get tv shows airing today: %w", err)
	}

	tvs := make([]domain.Media, 0, len(result.Results))
	for _, result := range result.Results {
		tvs = append(tvs, domain.Media{
			ID:          result.ID,
			Title:       result.Name,
			Overview:    result.Overview,
			ReleaseDate: result.FirstAirDate,
			PosterPath:  c.getPosterPath(result.PosterPath),
			VoteAverage: result.VoteAverage,
			VoteCount:   result.VoteCount,
			Type:        domain.MediaTypeTV,
		})
	}

	return &Paginated[domain.Media]{
		Items:      tvs,
		TotalPages: min(int(result.TotalPages), maxTmdbPagesCount),
		TotalItems: int(result.TotalResults),
	}, nil
}

func (c *client) GetTVShowOnTheAir(ctx context.Context, page int) (*Paginated[domain.Media], error) {
	opts := getDefaultOpts()
	opts["page"] = fmt.Sprintf("%d", page)

	tmdbClient, err := c.requestClient(ctx)
	if err != nil {
		return nil, err
	}

	result, err := tmdbClient.GetTVOnTheAir(opts)
	if err != nil {
		return nil, fmt.Errorf("get tv shows on the air: %w", err)
	}

	tvs := make([]domain.Media, 0, len(result.Results))
	for _, result := range result.Results {
		tvs = append(tvs, domain.Media{
			ID:          result.ID,
			Title:       result.Name,
			Overview:    result.Overview,
			ReleaseDate: result.FirstAirDate,
			PosterPath:  c.getPosterPath(result.PosterPath),
			VoteAverage: result.VoteAverage,
			VoteCount:   result.VoteCount,
			Type:        domain.MediaTypeTV,
		})
	}

	return &Paginated[domain.Media]{
		Items:      tvs,
		TotalPages: min(int(result.TotalPages), maxTmdbPagesCount),
		TotalItems: int(result.TotalResults),
	}, nil
}

func (c *client) GetTVShowTopRated(ctx context.Context, page int) (*Paginated[domain.Media], error) {
	opts := getDefaultOpts()
	opts["page"] = fmt.Sprintf("%d", page)

	tmdbClient, err := c.requestClient(ctx)
	if err != nil {
		return nil, err
	}

	result, err := tmdbClient.GetTVTopRated(opts)
	if err != nil {
		return nil, fmt.Errorf("get tv shows top rated: %w", err)
	}

	tvs := make([]domain.Media, 0, len(result.Results))
	for _, result := range result.Results {
		tvs = append(tvs, domain.Media{
			ID:          result.ID,
			Title:       result.Name,
			Overview:    result.Overview,
			ReleaseDate: result.FirstAirDate,
			PosterPath:  c.getPosterPath(result.PosterPath),
			VoteAverage: result.VoteAverage,
			VoteCount:   result.VoteCount,
			Type:        domain.MediaTypeTV,
		})
	}

	return &Paginated[domain.Media]{
		Items:      tvs,
		TotalPages: min(int(result.TotalPages), maxTmdbPagesCount),
		TotalItems: int(result.TotalResults),
	}, nil
}

func (c *client) GetTVShowPopular(ctx context.Context, page int) (*Paginated[domain.Media], error) {
	opts := getDefaultOpts()
	opts["page"] = fmt.Sprintf("%d", page)

	tmdbClient, err := c.requestClient(ctx)
	if err != nil {
		return nil, err
	}

	result, err := tmdbClient.GetTVPopular(opts)
	if err != nil {
		return nil, fmt.Errorf("get tv shows popular: %w", err)
	}

	tvs := make([]domain.Media, 0, len(result.Results))
	for _, result := range result.Results {
		tvs = append(tvs, domain.Media{
			ID:          result.ID,
			Title:       result.Name,
			Overview:    result.Overview,
			ReleaseDate: result.FirstAirDate,
			PosterPath:  c.getPosterPath(result.PosterPath),
			VoteAverage: result.VoteAverage,
			VoteCount:   result.VoteCount,
			Type:        domain.MediaTypeTV,
		})
	}

	return &Paginated[domain.Media]{
		Items:      tvs,
		TotalPages: min(int(result.TotalPages), maxTmdbPagesCount),
		TotalItems: int(result.TotalResults),
	}, nil
}

func (c *client) GetTVShowDetails(ctx context.Context, tmdbId int64) (*domain.TVShowDetail, error) {
	opts := getDefaultOpts()
	opts["append_to_response"] = "videos"

	tmdbClient, err := c.requestClient(ctx)
	if err != nil {
		return nil, err
	}

	result, err := tmdbClient.GetTVDetails(int(tmdbId), opts)
	if err != nil {
		return nil, fmt.Errorf("get tv show details: %w", err)
	}

	var seasons []domain.Season
	for _, season := range result.Seasons {
		// overview := season.Overview
		// title := season.Name

		// var translation *translationData

		// if overview == "" || title == "" {
		// 	translation, err = c.getTvSeasonEnglishTranslation(ctx, tmdbId, season.SeasonNumber)
		// }

		// if err == nil && translation != nil {
		// 	if overview == "" {
		// 		overview = translation.Overview
		// 	}
		// 	if title == "" {
		// 		title = translation.Title
		// 	}
		// }

		seasons = append(seasons, domain.Season{
			ID:           season.ID,
			ReleaseDate:  season.AirDate,
			EpisodeCount: season.EpisodeCount,
			Title:        season.Name,
			Overview:     season.Overview,
			PosterPath:   c.getPosterPath(season.PosterPath),
			SeasonNumber: season.SeasonNumber,
			VoteAverage:  season.VoteAverage,
			VoteCount:    result.VoteCount,
		})
	}

	productionCountries := make([]string, 0, len(result.ProductionCountries))
	for _, prod := range result.ProductionCountries {
		productionCountries = append(productionCountries, prod.Name)
	}

	return &domain.TVShowDetail{
		ID:                     result.ID,
		Title:                  result.Name,
		Overview:               result.Overview,
		ReleaseDate:            result.FirstAirDate,
		LastEpisodeReleaseDate: result.LastAirDate,
		NextEpisodeReleaseDate: result.NextEpisodeToAir.AirDate,
		PosterPath:             c.getPosterPath(result.PosterPath),
		Genres:                 toDomainGenres(result.Genres),
		OriginalLanguage:       result.OriginalLanguage,
		OriginCountry:          result.OriginCountry,
		OriginalTitle:          result.OriginalName,
		Popularity:             result.Popularity,
		Status:                 result.Status,
		Videos:                 toDomainViedos(result.Videos),
		VoteAverage:            result.VoteAverage,
		VoteCount:              result.VoteCount,
		NumberOfSeasons:        result.NumberOfSeasons,
		NumberOfEpisodes:       result.NumberOfEpisodes,
		Seasons:                seasons,
		ProductionCountries:    productionCountries,
	}, nil
}

func (c *client) GetTVShowRecommendations(ctx context.Context, tmdbId int64, page int) (*Paginated[domain.Media], error) {
	opts := getDefaultOpts()
	opts["page"] = fmt.Sprintf("%d", page)

	tmdbClient, err := c.requestClient(ctx)
	if err != nil {
		return nil, err
	}

	result, err := tmdbClient.GetTVRecommendations(int(tmdbId), opts)
	if err != nil {
		return nil, fmt.Errorf("get tv show recommendations: %w", err)
	}

	var tvs []domain.Media
	for _, result := range result.Results {
		tvs = append(tvs, domain.Media{
			ID:          result.ID,
			Title:       result.Name,
			Overview:    result.Overview,
			ReleaseDate: result.FirstAirDate,
			PosterPath:  c.getPosterPath(result.PosterPath),
			VoteAverage: result.VoteAverage,
			VoteCount:   result.VoteCount,
			Type:        domain.MediaTypeTV,
		})
	}

	return &Paginated[domain.Media]{
		Items:      tvs,
		TotalPages: min(int(result.TotalPages), maxTmdbPagesCount),
		TotalItems: int(result.TotalResults),
	}, nil
}

func (c *client) GetTvSeasonEpisodes(ctx context.Context, tvId int64, seasonNumber, page int) (*Paginated[domain.Episode], error) {
	opts := getDefaultOpts()
	perPage := 20

	tmdbClient, err := c.requestClient(ctx)
	if err != nil {
		return nil, err
	}

	result, err := tmdbClient.GetTVSeasonDetails(int(tvId), seasonNumber, opts)
	if err != nil {
		return nil, fmt.Errorf("get tv show season episodes: %w", err)
	}

	from := (page - 1) * perPage
	to := from + perPage

	var episodes []domain.Episode
	for index, episode := range result.Episodes {
		if page != 0 {
			if index < from {
				continue
			}
			if index >= to {
				break
			}
		}

		overview := episode.Overview
		title := episode.Name

		var translation *translationData

		if overview == "" || title == "" {
			translation, err = c.getTvEpisodeEnglishTranslation(ctx, tvId, seasonNumber, episode.EpisodeNumber)
		}

		if err == nil && translation != nil {
			if overview == "" {
				overview = translation.Overview
			}
			if title == "" {
				title = translation.Title
			}
		}

		episodes = append(episodes, domain.Episode{
			ID:            episode.ID,
			ReleaseDate:   episode.AirDate,
			EpisodeNumber: episode.EpisodeNumber,
			Title:         title,
			Overview:      overview,
			Runtime:       episode.Runtime,
			SeasonNumber:  episode.SeasonNumber,
			TVShowId:      episode.ShowID,
			PosterPath:    c.getPosterPath(episode.StillPath),
			VoteAverage:   episode.VoteAverage,
		})
	}

	totalItems := len(result.Episodes)
	totalPages := math.Ceil(float64(totalItems) / float64(perPage))

	return &Paginated[domain.Episode]{
		Items:      episodes,
		TotalPages: min(int(totalPages), maxTmdbPagesCount),
		TotalItems: len(result.Episodes),
	}, nil
}

func (c *client) getTvSeasonEnglishTranslation(ctx context.Context, tvShowId int64, seasonNumber int) (*translationData, error) {
	tmdbClient, err := c.requestClient(ctx)
	if err != nil {
		return nil, err
	}

	result, err := tmdbClient.GetTVSeasonTranslations(int(tvShowId), seasonNumber)
	if err != nil {
		return nil, fmt.Errorf("get tv show season english name and overview: %w", err)
	}

	for _, translation := range result.Translations {
		if translation.Iso3166_1 == "US" && translation.Iso639_1 == "en" {
			return &translationData{
				Title:    translation.Data.Name,
				Overview: translation.Data.Overview,
			}, nil
		}
	}

	return nil, fmt.Errorf("can not find translations for season %d", tvShowId)
}

func (c *client) getTvEpisodeEnglishTranslation(ctx context.Context, tvShowId int64, seasonNumber int, episodeNumber int) (*translationData, error) {
	tmdbClient, err := c.requestClient(ctx)
	if err != nil {
		return nil, err
	}

	result, err := tmdbClient.GetTVEpisodeTranslations(int(tvShowId), seasonNumber, episodeNumber)
	if err != nil {
		return nil, fmt.Errorf("get tv show season english name and overview: %w", err)
	}

	for _, translation := range result.Translations {
		if translation.Iso3166_1 == "US" && translation.Iso639_1 == "en" {
			return &translationData{
				Title:    translation.Data.Name,
				Overview: translation.Data.Overview,
			}, nil
		}
	}

	return nil, fmt.Errorf("can not find translations for season %d", tvShowId)
}
