package handler

import (
	"log/slog"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/xromen/movietracker/internal/domain"
	"github.com/xromen/movietracker/internal/service"
)

type companyDetailsResponse struct {
	ID            int64  `json:"id"`
	Name          string `json:"name"`
	Description   string `json:"description"`
	LogoPath      string `json:"logo_path"`
	OriginCountry string `json:"origin_country"`
	Homepage      string `json:"homepage"`
	Headquarters  string `json:"headquarters"`
}

type CompanyHandler struct {
	companyService service.CompanyService
	logger         *slog.Logger
}

func NewCompanyHandler(companyService service.CompanyService, logger *slog.Logger) *CompanyHandler {
	return &CompanyHandler{
		companyService: companyService,
		logger:         logger,
	}
}

func (h *CompanyHandler) GetDetails(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("company_id"), 10, 64)

	result, err := h.companyService.GetDetails(c.Request.Context(), id)
	if err != nil {
		handleServiceError(c, err, h.logger)
		return
	}

	response := toCompanyDetailsResponse(result)

	c.JSON(http.StatusOK, response)
}

func toCompanyDetailsResponse(company *domain.CompanyDetails) *companyDetailsResponse {
	return &companyDetailsResponse{
		ID: company.ID,
		Name: company.Name,
		Description: company.Description,
		LogoPath: company.LogoPath,
		OriginCountry: company.OriginCountry,
		Homepage: company.Homepage,
		Headquarters: company.Headquarters,
	}
}
