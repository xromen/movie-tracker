package handler

import (
	"log/slog"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/xromen/movietracker/internal/domain"
	"github.com/xromen/movietracker/internal/service"
)

type tvShowResponse struct {
	ID          int64   `json:"id"`
	Title       string  `json:"title"`
	Overview    string  `json:"overview"`
	PosterPath  string  `json:"poster_path"`
	ReleaseDate string  `json:"release_date"`
	VoteAverage float32 `json:"vote_average"`
	VoteCount   int64   `json:"vote_count"`
	WatchStatus string  `json:"watch_status,omitempty"`
}

type tvShowSearchResponse struct {
	TVShows    []tvShowResponse `json:"results"`
	TotalPages int              `json:"total_pages"`
	TotalItems int              `json:"total_items"`
}

type tvShowDetailResponse struct {
	ID                     int64                  `json:"id"`
	Title                  string                 `json:"title"`
	Overview               string                 `json:"overview"`
	ReleaseDate            string                 `json:"release_date"`
	LastEpisodeReleaseDate string                 `json:"last_episode_release_date"`
	NextEpisodeReleaseDate string                 `json:"next_episode_release_date"`
	PosterPath             string                 `json:"poster_path"`
	Genres                 []genreResponse        `json:"genres"`
	OriginalLanguage       string                 `json:"original_language"`
	OriginCountry          []string               `json:"origin_country"`
	OriginalTitle          string                 `json:"original_title"`
	ProductionCountries    []string               `json:"production_countries"`
	Status                 string                 `json:"status"`
	Videos                 []videoResponse        `json:"videos"`
	VoteAverage            float32                `json:"vote_average"`
	VoteCount              int64                  `json:"vote_count"`
	NumberOfSeasons        int                    `json:"number_of_seasons"`
	NumberOfEpisodes       int                    `json:"number_of_episodes"`
	Seasons                []tvShowSeasonResponse `json:"seasons"`
}

type tvShowSeasonResponse struct {
	ID           int64   `json:"id"`
	ReleaseDate  string  `json:"release_date"`
	EpisodeCount int     `json:"episode_count"`
	Title        string  `json:"title"`
	Overview     string  `json:"overview"`
	PosterPath   string  `json:"poster_path"`
	SeasonNumber int     `json:"season_number"`
	VoteAverage  float32 `json:"vote_average"`
	VoteCount    int64   `json:"vote_count"`
	IsWatched    *bool   `json:"is_watched,omitempty"`
}

type pagedTVShowEpisodesResponse struct {
	Episodes   []tvShowEpisodeResponse `json:"episodes"`
	TotalPages int                     `json:"total_pages"`
	TotalItems int                     `json:"total_items"`
}

type tvShowEpisodeResponse struct {
	ID            int64   `json:"id"`
	ReleaseDate   string  `json:"release_date"`
	EpisodeNumber int     `json:"episode_number"`
	Title         string  `json:"title"`
	Overview      string  `json:"overview"`
	Runtime       int     `json:"runtime"`
	SeasonNumber  int     `json:"season_number"`
	TVShowId      int64   `json:"tv_show_id"`
	PosterPath    string  `json:"poster_path"`
	VoteAverage   float32 `json:"vote_average"`
	VoteCount     int64   `json:"vote_count"`
	IsWatched     *bool   `json:"is_watched,omitempty"`
}

type TVShowHandler struct {
	tvShowService service.TVShowService
	logger        *slog.Logger
}

func NewTVShowHandler(tvShowService service.TVShowService, logger *slog.Logger) *TVShowHandler {
	return &TVShowHandler{
		tvShowService: tvShowService,
		logger:        logger,
	}
}

func (h *TVShowHandler) Search(c *gin.Context) {
	query := c.Query("q")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))

	result, err := h.tvShowService.Search(c.Request.Context(), query, page)
	if err != nil {
		handleServiceError(c, err, h.logger)
		return
	}

	c.JSON(http.StatusOK, toPaginatedTVShows(result))
}

func (h *TVShowHandler) GetAiringToday(c *gin.Context) {
	var userID *int64
	if id, exists := c.Get(ContextUserID); exists {
		if v, ok := id.(int64); ok {
			userID = &v
		}
	}
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))

	result, err := h.tvShowService.GetAiringToday(c.Request.Context(), userID, page)
	if err != nil {
		handleServiceError(c, err, h.logger)
		return
	}

	c.JSON(http.StatusOK, toPaginatedTVShows(result))
}

func (h *TVShowHandler) GetPopular(c *gin.Context) {
	var userID *int64
	if id, exists := c.Get(ContextUserID); exists {
		if v, ok := id.(int64); ok {
			userID = &v
		}
	}
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))

	result, err := h.tvShowService.GetPopular(c.Request.Context(), userID, page)
	if err != nil {
		handleServiceError(c, err, h.logger)
		return
	}

	c.JSON(http.StatusOK, toPaginatedTVShows(result))
}

func (h *TVShowHandler) GetTopRated(c *gin.Context) {
	var userID *int64
	if id, exists := c.Get(ContextUserID); exists {
		if v, ok := id.(int64); ok {
			userID = &v
		}
	}
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))

	result, err := h.tvShowService.GetTopRated(c.Request.Context(), userID, page)
	if err != nil {
		handleServiceError(c, err, h.logger)
		return
	}

	c.JSON(http.StatusOK, toPaginatedTVShows(result))
}

func (h *TVShowHandler) GetOnTheAir(c *gin.Context) {
	var userID *int64
	if id, exists := c.Get(ContextUserID); exists {
		if v, ok := id.(int64); ok {
			userID = &v
		}
	}
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))

	result, err := h.tvShowService.GetOnTheAir(c.Request.Context(), userID, page)
	if err != nil {
		handleServiceError(c, err, h.logger)
		return
	}

	c.JSON(http.StatusOK, toPaginatedTVShows(result))
}

func (h *TVShowHandler) GetTVShowDetail(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	result, err := h.tvShowService.GetDetails(c.Request.Context(), id)
	if err != nil {
		handleServiceError(c, err, h.logger)
		return
	}

	c.JSON(http.StatusOK, toTVShowDetailResponse(result))
}

func (h *TVShowHandler) GetRecommendations(c *gin.Context) {
	var userID *int64
	if id, exists := c.Get(ContextUserID); exists {
		if v, ok := id.(int64); ok {
			userID = &v
		}
	}
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))

	result, err := h.tvShowService.GetRecommendations(c.Request.Context(), userID, id, page)
	if err != nil {
		handleServiceError(c, err, h.logger)
		return
	}

	c.JSON(http.StatusOK, toPaginatedTVShows(result))
}

func (h *TVShowHandler) GetSeasonEpisodes(c *gin.Context) {
	var userID *int64
	if id, exists := c.Get(ContextUserID); exists {
		if v, ok := id.(int64); ok {
			userID = &v
		}
	}
	tvShowID, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	seasonNumber, _ := strconv.Atoi(c.Param("season_number"))
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))

	result, err := h.tvShowService.GetSeasonEpisodes(c.Request.Context(), userID, tvShowID, seasonNumber, page)
	if err != nil {
		handleServiceError(c, err, h.logger)
		return
	}

	c.JSON(http.StatusOK, toEpisodesResponse(result))
}

func (h *TVShowHandler) MarkEpisodeWatched(c *gin.Context) {
	h.setEpisodeWatched(c, true)
}

func (h *TVShowHandler) UnmarkEpisodeWatched(c *gin.Context) {
	h.setEpisodeWatched(c, false)
}

func (h *TVShowHandler) MarkSeasonWatched(c *gin.Context) {
	h.setSeasonWatched(c, true)
}

func (h *TVShowHandler) UnmarkSeasonWatched(c *gin.Context) {
	h.setSeasonWatched(c, false)
}

func (h *TVShowHandler) setSeasonWatched(c *gin.Context, watched bool) {
	tvShowID, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	seasonNumber, _ := strconv.Atoi(c.Param("season_number"))

	userID, _ := c.Get(ContextUserID)

	var err error
	if watched {
		err = h.tvShowService.MarkSeasonWatched(
			c.Request.Context(),
			userID.(int64),
			tvShowID,
			seasonNumber,
		)
	} else {
		err = h.tvShowService.UnmarkSeasonWatched(
			c.Request.Context(),
			userID.(int64),
			tvShowID,
			seasonNumber,
		)
	}

	if err != nil {
		handleServiceError(c, err, h.logger)
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *TVShowHandler) setEpisodeWatched(c *gin.Context, watched bool) {
	tvShowID, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	seasonNumber, _ := strconv.Atoi(c.Param("season_number"))
	episodeNumber, _ := strconv.Atoi(c.Param("episode_number"))

	userID, _ := c.Get(ContextUserID)

	err := h.tvShowService.SetEpisodeWatched(
		c.Request.Context(),
		userID.(int64),
		tvShowID,
		seasonNumber,
		episodeNumber,
		watched,
	)
	if err != nil {
		handleServiceError(c, err, h.logger)
		return
	}

	c.Status(http.StatusNoContent)
}

func toPaginatedTVShows(result *service.TVShowPageOutput) tvShowSearchResponse {
	tvShows := make([]tvShowResponse, len(result.TVShows))
	for i, m := range result.TVShows {
		tvShows[i] = toTVShowResponse(m)
	}
	return tvShowSearchResponse{
		TVShows:    tvShows,
		TotalPages: result.TotalPages,
		TotalItems: result.TotalItems,
	}
}

func toTVShowResponse(tvShow domain.Media) tvShowResponse {
	return tvShowResponse{
		ID:          tvShow.ID,
		Title:       tvShow.Title,
		Overview:    tvShow.Overview,
		PosterPath:  tvShow.PosterPath,
		ReleaseDate: tvShow.ReleaseDate,
		VoteAverage: tvShow.VoteAverage,
		VoteCount:   tvShow.VoteCount,
		WatchStatus: string(tvShow.WatchStatus),
	}
}

func toTVShowDetailResponse(tvShow *domain.TVShowDetail) tvShowDetailResponse {
	return tvShowDetailResponse{
		ID:                     tvShow.ID,
		Title:                  tvShow.Title,
		Overview:               tvShow.Overview,
		ReleaseDate:            tvShow.ReleaseDate,
		PosterPath:             tvShow.PosterPath,
		Genres:                 toGenreResponses(tvShow.Genres),
		OriginalLanguage:       tvShow.OriginalLanguage,
		OriginCountry:          tvShow.OriginCountry,
		OriginalTitle:          tvShow.OriginalTitle,
		Status:                 tvShow.Status,
		Videos:                 toVideoResponses(tvShow.Videos),
		VoteAverage:            tvShow.VoteAverage,
		VoteCount:              tvShow.VoteCount,
		NumberOfSeasons:        tvShow.NumberOfSeasons,
		NumberOfEpisodes:       tvShow.NumberOfEpisodes,
		Seasons:                toSeasonsResponse(tvShow.Seasons),
		LastEpisodeReleaseDate: tvShow.LastEpisodeReleaseDate,
		NextEpisodeReleaseDate: tvShow.NextEpisodeReleaseDate,
		ProductionCountries:    tvShow.ProductionCountries,
	}
}

func toSeasonsResponse(seasons []domain.Season) []tvShowSeasonResponse {
	result := make([]tvShowSeasonResponse, 0, len(seasons))
	for _, season := range seasons {
		result = append(result, tvShowSeasonResponse{
			ID:           season.ID,
			ReleaseDate:  season.ReleaseDate,
			EpisodeCount: season.EpisodeCount,
			Title:        season.Title,
			Overview:     season.Overview,
			PosterPath:   season.PosterPath,
			SeasonNumber: season.SeasonNumber,
			VoteAverage:  season.VoteAverage,
			VoteCount:    season.VoteCount,
			IsWatched:    season.IsWatched,
		})
	}
	return result
}

func toEpisodesResponse(result *service.EpisodePageOutput) *pagedTVShowEpisodesResponse {
	episodes := make([]tvShowEpisodeResponse, 0, len(result.Episodes))
	for _, episode := range result.Episodes {
		episodes = append(episodes, tvShowEpisodeResponse{
			ID:            episode.ID,
			ReleaseDate:   episode.ReleaseDate,
			EpisodeNumber: episode.EpisodeNumber,
			Title:         episode.Title,
			Overview:      episode.Overview,
			Runtime:       episode.Runtime,
			SeasonNumber:  episode.SeasonNumber,
			TVShowId:      episode.TVShowId,
			PosterPath:    episode.PosterPath,
			VoteAverage:   episode.VoteAverage,
			VoteCount:     episode.VoteCount,
			IsWatched:     episode.IsWatched,
		})
	}

	return &pagedTVShowEpisodesResponse{
		Episodes:   episodes,
		TotalPages: result.TotalPages,
		TotalItems: result.TotalItems,
	}
}
