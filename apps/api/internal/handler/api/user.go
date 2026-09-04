package handler

import (
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/xromen/movietracker/internal/domain"
	"github.com/xromen/movietracker/internal/service"
)

type UserHandler struct {
	userService service.UserService
	logger      *slog.Logger
}

func NewUserHandler(userService service.UserService, logger *slog.Logger) *UserHandler {
	return &UserHandler{userService: userService, logger: logger}
}

type registerRequest struct {
	Email    string `json:"email"    binding:"required,email"`
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type loginRequest struct {
	Email    string `json:"email"    binding:"required"`
	Password string `json:"password" binding:"required"`
}

const (
	accessTokenCookie  = "access_token"
	refreshTokenCookie = "refresh_token"
)

func (h *UserHandler) Register(c *gin.Context) {
	var req registerRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse(err.Error()))
		return
	}

	out, err := h.userService.Register(c.Request.Context(), service.RegisterInput{
		Email:    req.Email,
		Username: req.Username,
		Password: req.Password,
	})
	if err != nil {
		handleServiceError(c, err, h.logger)
		return
	}

	setAuthCookies(c, out)

	c.Status(http.StatusCreated)
}

func (h *UserHandler) Login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse(err.Error()))
		return
	}

	out, err := h.userService.Login(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		handleServiceError(c, err, h.logger)
		return
	}

	setAuthCookies(c, out)

	c.Status(http.StatusOK)
}

func (h *UserHandler) Logout(c *gin.Context) {
	setAuthCookies(c, nil)

	c.Status(http.StatusOK)
}

func (h *UserHandler) Refresh(c *gin.Context) {
	_, refreshToken := getAuthTokens(c)

	if refreshToken == nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, errorResponse("refresh token invalid"))
		return
	}

	out, err := h.userService.Refresh(c.Request.Context(), *refreshToken)
	if err != nil {
		h.logger.Error("error refresh token",
			"token", refreshToken,
			"error", err,
		)
		c.AbortWithStatusJSON(http.StatusUnauthorized, errorResponse("refresh token invalid"))
		return
	}

	setAuthCookies(c, out)

	c.Status(http.StatusOK)
}

func setAuthCookies(c *gin.Context, out *service.AuthOutput) {
	accessToken := ""
	accessTokenExpiresAt := time.Unix(0, 0)
	refreshToken := ""
	refreshTokenExpiresAt := time.Unix(0, 0)

	if out != nil {
		accessToken = out.AccessToken
		accessTokenExpiresAt = out.AccessTokenExpiresAt
		refreshToken = out.RefreshToken
		refreshTokenExpiresAt = out.RefreshTokenExpiresAt
	}

	http.SetCookie(c.Writer, &http.Cookie{
		Name:     accessTokenCookie,
		Value:    accessToken,
		Path:     "/",
		Expires:  accessTokenExpiresAt,
		MaxAge:   int(time.Until(accessTokenExpiresAt).Seconds()),
		HttpOnly: true,
		Secure:   false, // в локальной разработке обычно false без HTTPS
		SameSite: http.SameSiteLaxMode,
	})

	http.SetCookie(c.Writer, &http.Cookie{
		Name:  refreshTokenCookie,
		Value: refreshToken,
		Path:  "/",
		//Path:     "/api/v1/auth/refresh",
		Expires:  refreshTokenExpiresAt,
		MaxAge:   int(time.Until(refreshTokenExpiresAt).Seconds()),
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	})
}

func getAuthTokens(c *gin.Context) (*string, *string) {
	var accessToken *string
	var refreshToken *string

	if accessTokenFromCookie, err := c.Cookie(accessTokenCookie); err == nil {
		accessToken = &accessTokenFromCookie
	}

	if refreshTokenFromCookie, err := c.Cookie(refreshTokenCookie); err == nil {
		refreshToken = &refreshTokenFromCookie
	}

	return accessToken, refreshToken
}

func handleServiceError(c *gin.Context, err error, logger *slog.Logger) {
	switch {
	case errors.Is(err, domain.ErrNotFound):
		c.JSON(http.StatusNotFound, errorResponse("not found"))
	case errors.Is(err, domain.ErrAlreadyExists):
		c.JSON(http.StatusConflict, errorResponse("already exists"))
	case errors.Is(err, domain.ErrUnauthorized):
		c.JSON(http.StatusUnauthorized, errorResponse("unauthorized"))
	case errors.Is(err, domain.ErrInvalidInput):
		c.JSON(http.StatusBadRequest, errorResponse(err.Error()))
	default:
		{
			logger.Error("internal server error", "err", err)
			c.JSON(http.StatusInternalServerError, errorResponse("internal server error"))
		}
	}
}

type errorResp struct {
	Error string `json:"error"`
}

func errorResponse(msg string) errorResp {
	return errorResp{Error: msg}
}
