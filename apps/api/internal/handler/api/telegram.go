package handler

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/xromen/movietracker/internal/service"
)

type TelegramHandler struct {
	telegramService service.TelegramService
	logger          *slog.Logger
}

type bindingUrlResponse struct {
	URL       string    `json:"url"`
	ExpiresAt time.Time `json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
}

func NewTelegramHandler(
	telegramService service.TelegramService,
	logger *slog.Logger) *TelegramHandler {
	return &TelegramHandler{
		telegramService: telegramService,
		logger:          logger,
	}
}

func (h *TelegramHandler) GenerateBindingUrl(c *gin.Context) {
	userID, _ := c.Get(ContextUserID)

	out, err := h.telegramService.GenerateBindingUrl(c.Request.Context(), userID.(int64))
	if err != nil {
		handleServiceError(c, err, h.logger)
		return
	}

	response := bindingUrlResponse{
		URL: out.URL,
		ExpiresAt: out.ExpiresAt,
		CreatedAt: out.CreatedAt,
	}

	c.JSON(http.StatusOK, response)
}
