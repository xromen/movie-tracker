package handler

import (
	"log/slog"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/xromen/movietracker/internal/service"
)

type paginatedMediaResponse struct {
	Medias     []mediaResponse `json:"results"`
	TotalPages int             `json:"total_pages"`
	TotalItems int             `json:"total_items"`
}

type SearchHandler struct {
	searchService service.SearchService
	logger        *slog.Logger
}

func NewSearchHandler(searchService service.SearchService, logger *slog.Logger) *SearchHandler {
	return &SearchHandler{
		searchService: searchService,
		logger:        logger,
	}
}

func (h *SearchHandler) SearchMulti(c *gin.Context) {
	query := c.Query("query")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))

	result, err := h.searchService.SearchMulti(c.Request.Context(), query, page)
	if err != nil {
		handleServiceError(c, err, h.logger)
		return
	}

	c.JSON(http.StatusOK, toPaginatedMedias(result))
}

func toPaginatedMedias(result *service.MediaPageOutput) paginatedMediaResponse {
	medias := make([]mediaResponse, len(result.Medias))
	for i, m := range result.Medias {
		medias[i] = *toMediaResponse(&m)
	}
	return paginatedMediaResponse{
		Medias:     medias,
		TotalPages: result.TotalPages,
		TotalItems: result.TotalItems,
	}
}
