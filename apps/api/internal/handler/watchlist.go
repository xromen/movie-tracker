package handler

import (
	"log/slog"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/xromen/movietracker/internal/domain"
	"github.com/xromen/movietracker/internal/service"
)

type setStatusRequest struct {
	ID          int64  `json:"id"`
	WatchStatus string `json:"watch_status"`
	MediaType   string `json:"media_type"`
}

type mediaResponse struct {
	ID          int64   `json:"id"`
	Title       string  `json:"title"`
	Overview    string  `json:"overview"`
	PosterPath  string  `json:"poster_path"`
	ReleaseDate string  `json:"release_date"`
	VoteAverage float32 `json:"vote_average"`
	VoteCount   int64   `json:"vote_count"`
	Type        string  `json:"type"`
}

type userMediaResponse struct {
	Media  *mediaResponse     `json:"media"`
	Status domain.WatchStatus `json:"status"`
	Rating *int               `json:"rating"`
}

type userWatchListResponse struct {
	Medias  []userMediaResponse `json:"medias"`
	Total   int                 `json:"total"`
	Page    int                 `json:"page"`
	PerPage int                 `json:"per_page"`
}

type watchStatusResponse struct {
	WatchStatus *string `json:"watch_status"`
}

type WatchListHandler struct {
	watchListService service.WatchListService
	logger           *slog.Logger
}

func NewWatchListHandler(
	watchListService service.WatchListService,
	logger *slog.Logger) *WatchListHandler {
	return &WatchListHandler{
		watchListService: watchListService,
		logger:           logger,
	}
}

func (h *WatchListHandler) GetUserWatchList(c *gin.Context) {
	userID, _ := c.Get(ContextUserID)
	status := domain.WatchStatus(c.Query("status"))
	mediaType := domain.MediaType(c.Query("media_type"))
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "20"))

	input := service.UserWatchListInput{
		UserID:    userID.(int64),
		Status:    status,
		MediaType: mediaType,
		Page:      page,
		PerPage:   perPage,
	}

	result, err := h.watchListService.GetUserWatchList(c.Request.Context(), input)
	if err != nil {
		handleServiceError(c, err, h.logger)
		return
	}

	response := userWatchListResponse{
		Medias:  toUserMediasResponses(result.Medias),
		Total:   result.Total,
		Page:    result.Page,
		PerPage: result.PerPage,
	}

	c.JSON(http.StatusOK, response)
}

func (h *WatchListHandler) DeleteUserStatus(c *gin.Context) {
	userID, _ := c.Get(ContextUserID)
	mediaID, _ := strconv.ParseInt(c.Query("media_id"), 10, 64)
	mediaType := domain.MediaType(c.Query("media_type"))

	result, err := h.watchListService.DeleteMediaUserStatus(c.Request.Context(), mediaType, userID.(int64), mediaID)
	if err != nil {
		handleServiceError(c, err, h.logger)
		return
	}

	c.JSON(http.StatusOK, toUserMediaResponse(*result))
}

func (h *WatchListHandler) GetMediaUserStatus(c *gin.Context) {
	userID, _ := c.Get(ContextUserID)
	mediaID, _ := strconv.ParseInt(c.Query("media_id"), 10, 64)
	mediaType := domain.MediaType(c.Query("media_type"))

	status, err := h.watchListService.GetMediaUserStatus(c.Request.Context(), mediaType, userID.(int64), mediaID)
	if err != nil {
		handleServiceError(c, err, h.logger)
		return
	}

	if status == nil {
		c.JSON(http.StatusOK, watchStatusResponse{WatchStatus: nil})
		return
	}

	result := string(*status)

	c.JSON(http.StatusOK, watchStatusResponse{WatchStatus: &result})
}

func (h *WatchListHandler) SetStatus(c *gin.Context) {
	var req setStatusRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse(err.Error()))
		return
	}

	userID, _ := c.Get(ContextUserID)

	result, err := h.watchListService.SetMediaUserStatus(c.Request.Context(), service.SetMediaUserStatusInput{
		UserID:    userID.(int64),
		Status:    domain.WatchStatus(req.WatchStatus),
		MediaType: domain.MediaType(req.MediaType),
		MediaID:   req.ID,
	})

	if err != nil {
		handleServiceError(c, err, h.logger)
		return
	}

	c.JSON(http.StatusOK, toUserMediaResponse(*result))
}

func toUserMediasResponses(medias []domain.UserMedia) []userMediaResponse {
	result := make([]userMediaResponse, 0, len(medias))
	for _, media := range medias {
		result = append(result, toUserMediaResponse(media))
	}

	return result
}

func toUserMediaResponse(media domain.UserMedia) userMediaResponse {
	return userMediaResponse{
		Media:  toMediaResponse(media.Media),
		Status: media.Status,
		Rating: media.Rating,
	}
}

func toMediaResponse(media *domain.Media) *mediaResponse {
	return &mediaResponse{
		ID:          media.ID,
		Title:       media.Title,
		Overview:    media.Overview,
		PosterPath:  media.PosterPath,
		ReleaseDate: media.ReleaseDate,
		VoteAverage: media.VoteAverage,
		VoteCount:   media.VoteCount,
		Type:        string(media.Type),
	}
}
