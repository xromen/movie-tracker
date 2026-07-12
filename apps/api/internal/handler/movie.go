package handler

import (
	"log/slog"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/xromen/movietracker/internal/domain"
	"github.com/xromen/movietracker/internal/service"
)

type movieResponse struct {
	ID          int64   `json:"id"`
	Title       string  `json:"title"`
	Overview    string  `json:"overview"`
	PosterPath  string  `json:"poster_path"`
	ReleaseDate string  `json:"release_date"`
	Popularity  float32 `json:"popularity"`
	VoteAverage float32 `json:"vote_average"`
	VoteCount   int64   `json:"vote_count"`
	WatchStatus string  `json:"watch_status,omitempty"`
}

type searchResponse struct {
	Movies     []movieResponse `json:"results"`
	TotalPages int             `json:"total_pages"`
	TotalItems int             `json:"total_items"`
}

type genreResponse struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

type videoResponse struct {
	ID          string `json:"id"`
	Iso639_1    string `json:"iso_639_1"`
	Iso3166_1   string `json:"iso_3166_1"`
	Key         string `json:"key"`
	Name        string `json:"name"`
	Official    bool   `json:"official"`
	PublishedAt string `json:"published_at"`
	Site        string `json:"site"`
	Size        int    `json:"size"`
	Type        string `json:"type"`
}

type movieDetailResponse struct {
	ID                  int64           `json:"id"`
	Title               string          `json:"title"`
	Overview            string          `json:"overview"`
	ReleaseDate         string          `json:"release_date"`
	PosterPath          string          `json:"poster_path"`
	Genres              []genreResponse `json:"genres"`
	OriginalLanguage    string          `json:"original_language"`
	OriginCountry       []string        `json:"origin_country"`
	OriginalTitle       string          `json:"original_title"`
	ProductionCountries []string        `json:"production_countries"`
	Popularity          float32         `json:"popularity"`
	Status              string          `json:"status"`
	Videos              []videoResponse `json:"videos"`
	VoteAverage         float32         `json:"vote_average"`
	VoteCount           int64           `json:"vote_count"`
	Revenue             int64           `json:"revenue"`
	Budget              int64           `json:"budget"`
	Runtime             int             `json:"runtime"`
	CollectionID        *int64          `json:"collection_id"`
}

type MovieHandler struct {
	movieService service.MovieService
	logger       *slog.Logger
}

func NewMovieHandler(movieService service.MovieService, logger *slog.Logger) *MovieHandler {
	return &MovieHandler{
		movieService: movieService,
		logger:       logger,
	}
}

func (h *MovieHandler) Search(c *gin.Context) {
	query := c.Query("q")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))

	result, err := h.movieService.Search(c.Request.Context(), query, page)
	if err != nil {
		handleServiceError(c, err, h.logger)
		return
	}

	c.JSON(http.StatusOK, toPaginatedMovies(result))
}

func (h *MovieHandler) NowPlaying(c *gin.Context) {
	var userID *int64
	if id, exists := c.Get(ContextUserID); exists {
		if v, ok := id.(int64); ok {
			userID = &v
		}
	}
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))

	result, err := h.movieService.GetNowPlaying(c.Request.Context(), userID, page)
	if err != nil {
		handleServiceError(c, err, h.logger)
		return
	}

	c.JSON(http.StatusOK, toPaginatedMovies(result))
}

func (h *MovieHandler) Popular(c *gin.Context) {
	var userID *int64
	if id, exists := c.Get(ContextUserID); exists {
		if v, ok := id.(int64); ok {
			userID = &v
		}
	}
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))

	result, err := h.movieService.GetPopular(c.Request.Context(), userID, page)
	if err != nil {
		handleServiceError(c, err, h.logger)
		return
	}

	c.JSON(http.StatusOK, toPaginatedMovies(result))
}

func (h *MovieHandler) TopRated(c *gin.Context) {
	var userID *int64
	if id, exists := c.Get(ContextUserID); exists {
		if v, ok := id.(int64); ok {
			userID = &v
		}
	}
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))

	result, err := h.movieService.GetTopRated(c.Request.Context(), userID, page)
	if err != nil {
		handleServiceError(c, err, h.logger)
		return
	}

	c.JSON(http.StatusOK, toPaginatedMovies(result))
}

func (h *MovieHandler) Upcoming(c *gin.Context) {
	var userID *int64
	if id, exists := c.Get(ContextUserID); exists {
		if v, ok := id.(int64); ok {
			userID = &v
		}
	}
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))

	result, err := h.movieService.GetUpcoming(c.Request.Context(), userID, page)
	if err != nil {
		handleServiceError(c, err, h.logger)
		return
	}

	c.JSON(http.StatusOK, toPaginatedMovies(result))
}

func (h *MovieHandler) GetMovieDetail(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	result, err := h.movieService.GetDetails(c.Request.Context(), id)
	if err != nil {
		handleServiceError(c, err, h.logger)
		return
	}

	c.JSON(http.StatusOK, toMovieDetailResponse(result))
}

func (h *MovieHandler) GetMovieRecommendations(c *gin.Context) {
	var userID *int64
	if id, exists := c.Get(ContextUserID); exists {
		if v, ok := id.(int64); ok {
			userID = &v
		}
	}
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))

	result, err := h.movieService.GetRecommendations(c.Request.Context(), userID, id, page)
	if err != nil {
		handleServiceError(c, err, h.logger)
		return
	}

	c.JSON(http.StatusOK, toPaginatedMovies(result))
}

func toPaginatedMovies(result *service.MoviePageOutput) searchResponse {
	movies := make([]movieResponse, len(result.Movies))
	for i, m := range result.Movies {
		movies[i] = toMovieResponse(m)
	}
	return searchResponse{
		Movies:     movies,
		TotalPages: result.TotalPages,
		TotalItems: result.TotalItems,
	}
}

func toMovieResponse(movie domain.Media) movieResponse {
	return movieResponse{
		ID:          movie.ID,
		Title:       movie.Title,
		Overview:    movie.Overview,
		PosterPath:  movie.PosterPath,
		ReleaseDate: movie.ReleaseDate,
		VoteAverage: movie.VoteAverage,
		VoteCount:   movie.VoteCount,
		WatchStatus: string(movie.WatchStatus),
	}
}

func toMovieDetailResponse(movie *domain.MovieDetail) movieDetailResponse {
	return movieDetailResponse{
		ID:                  movie.ID,
		Title:               movie.Title,
		Overview:            movie.Overview,
		ReleaseDate:         movie.ReleaseDate,
		PosterPath:          movie.PosterPath,
		Genres:              toGenreResponses(movie.Genres),
		OriginalLanguage:    movie.OriginalLanguage,
		OriginCountry:       movie.OriginCountry,
		OriginalTitle:       movie.OriginalTitle,
		ProductionCountries: movie.ProductionCountries,
		Popularity:          movie.Popularity,
		Status:              movie.Status,
		Videos:              toVideoResponses(movie.Videos),
		VoteAverage:         movie.VoteAverage,
		VoteCount:           movie.VoteCount,
		Revenue:             movie.Revenue,
		Budget:              movie.Budget,
		Runtime:             movie.Runtime,
		CollectionID:        movie.CollectionID,
	}
}

func toGenreResponses(genres []domain.Genre) []genreResponse {
	responses := make([]genreResponse, len(genres))
	for i, genre := range genres {
		responses[i] = genreResponse{
			ID:   genre.ID,
			Name: genre.Name,
		}
	}
	return responses
}

func toVideoResponses(videos []domain.Video) []videoResponse {
	responses := make([]videoResponse, len(videos))
	for i, video := range videos {
		responses[i] = videoResponse{
			ID:          video.ID,
			Iso639_1:    video.Iso639_1,
			Iso3166_1:   video.Iso3166_1,
			Key:         video.Key,
			Name:        video.Name,
			Official:    video.Official,
			PublishedAt: video.PublishedAt,
			Site:        video.Site,
			Size:        video.Size,
			Type:        video.Type,
		}
	}
	return responses
}
