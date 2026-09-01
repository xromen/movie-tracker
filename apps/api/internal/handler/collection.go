package handler

import (
	"log/slog"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/xromen/movietracker/internal/domain"
	"github.com/xromen/movietracker/internal/service"
)

type collectionResponse struct {
	ID         int64                    `json:"id"`
	Name       string                   `json:"name"`
	Overview   string                   `json:"overview"`
	PosterPath string                   `json:"poster_path"`
	Parts      []collectionPartResponse `json:"parts"`
}

type collectionPartResponse struct {
	ID          int64   `json:"id"`
	Title       string  `json:"title"`
	Overview    string  `json:"overview"`
	PosterPath  string  `json:"poster_path"`
	MediaType   string  `json:"type"`
	ReleaseDate string  `json:"release_date"`
	VoteAverage float32 `json:"vote_average"`
	WatchStatus string  `json:"watch_status"`
}

type CollectionHandler struct {
	collectionService service.CollectionService
	logger            *slog.Logger
}

func NewCollectionHandler(collectionService service.CollectionService, logger *slog.Logger) *CollectionHandler {
	return &CollectionHandler{
		collectionService: collectionService,
		logger:            logger,
	}
}

func (h *CollectionHandler) GetDetails(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var userID *int64
	if id, exists := c.Get(ContextUserID); exists {
		if v, ok := id.(int64); ok {
			userID = &v
		}
	}

	result, err := h.collectionService.GetDetails(c.Request.Context(), id, userID)
	if err != nil {
		handleServiceError(c, err, h.logger)
		return
	}

	response := toCollectionResponse(result)

	c.JSON(http.StatusOK, response)
}

func toCollectionResponse(collection *domain.Collection) *collectionResponse {
	var parts []collectionPartResponse
	for _, part := range collection.Parts {
		parts = append(parts, collectionPartResponse{
			ID:          part.ID,
			Title:       part.Title,
			Overview:    part.Overview,
			PosterPath:  part.PosterPath,
			MediaType:   part.MediaType,
			ReleaseDate: part.ReleaseDate,
			VoteAverage: part.VoteAverage,
			WatchStatus: string(part.WatchStatus),
		})
	}
	return &collectionResponse{
		ID:         collection.ID,
		Name:       collection.Name,
		Overview:   collection.Overview,
		PosterPath: collection.PosterPath,
		Parts:      parts,
	}
}
