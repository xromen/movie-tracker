package handler

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"log/slog"
	"net/http"
	"slices"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/xromen/movietracker/internal/platform/jwt"
	"golang.org/x/time/rate"
)

type contextKey string

const (
	traceIDKey    contextKey = "trace_id"
	headerTraceID            = "X-Request-ID"
)

func TraceID() gin.HandlerFunc {
	return func(c *gin.Context) {
		traceID := c.GetHeader(headerTraceID)
		if traceID == "" {
			traceID = generateTraceID()
		}

		c.Set(string(traceIDKey), traceID)

		ctx := context.WithValue(c.Request.Context(), traceIDKey, traceID)
		c.Request = c.Request.WithContext(ctx)

		c.Header(headerTraceID, traceID)

		c.Next()
	}
}

func TraceIDFromContext(ctx context.Context) string {
	if id, ok := ctx.Value(traceIDKey).(string); ok {
		return id
	}
	return ""
}

func generateTraceID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

func StructuredLogger(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.URL.Path == "/metrics" {
			c.Next()
			return
		}

		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		c.Next()

		latency := time.Since(start)
		status := c.Writer.Status()
		traceID := TraceIDFromContext(c.Request.Context())

		attrs := []any{
			"trace_id", traceID,
			"method", c.Request.Method,
			"path", path,
			"status", status,
			"latency_ms", latency.Milliseconds(),
			"ip", c.ClientIP(),
			"user_agent", c.Request.UserAgent(),
		}

		// Опциональные атрибуты.
		if query != "" {
			attrs = append(attrs, "query", query)
		}
		if len(c.Errors) > 0 {
			attrs = append(attrs, "errors", c.Errors.String())
		}
		if id, exists := c.Get(ContextUserID); exists {
			if v, ok := id.(int64); ok {
				attrs = append(attrs, "userID", v)
			}
		}
		if username, exists := c.Get(ContextUsername); exists {
			attrs = append(attrs, "username", username)
		}

		// Уровень лога зависит от статуса ответа.
		switch {
		case status >= 500:
			logger.Error("request", attrs...)
		case status >= 400:
			logger.Warn("request", attrs...)
		default:
			logger.Info("request", attrs...)
		}
	}
}

type ipLimiter struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

const (
	ContextUserID   = "user_id"
	ContextUsername = "username"
)

type userService interface {
	ValidateAuthVersion(ctx context.Context, userID, authVersion int64) (bool, error)
}

func AuthMiddleware(jwtManager jwt.Manager, userService userService, role *string) gin.HandlerFunc {
	return func(c *gin.Context) {
		accessToken, _ := getAuthTokens(c)
		if accessToken == nil {
			c.JSON(http.StatusUnauthorized, errorResponse("no access token"))
			c.Abort()
			return
		}

		claims, err := jwtManager.Validate(*accessToken)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, errorResponse("invalid or expired token"))
			return
		}

		authVersionValidated, err := userService.ValidateAuthVersion(c.Request.Context(), claims.UserID, claims.AuthVersion)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, errorResponse("invalid or expired token"))
			return
		}
		if !authVersionValidated {
			c.AbortWithStatusJSON(http.StatusUnauthorized, errorResponse("invalid or expired token"))
			return
		}

		if role != nil && !slices.Contains(claims.Roles, *role) {
			c.AbortWithStatusJSON(http.StatusForbidden, errorResponse("dont have permission"))
			return
		}

		c.Set(ContextUserID, claims.UserID)
		c.Set(ContextUsername, claims.Username)
		c.Next()
	}
}

func OptionalAuthMiddleware(jwtManager jwt.Manager, userService userService) gin.HandlerFunc {
	return func(c *gin.Context) {
		accessToken, _ := getAuthTokens(c)
		if accessToken == nil {
			c.Next()
			return
		}

		claims, err := jwtManager.Validate(*accessToken)
		if err != nil {
			setAuthCookies(c, nil)
			c.Next()
			return
		}

		authVersionValidated, err := userService.ValidateAuthVersion(c.Request.Context(), claims.UserID, claims.AuthVersion)
		if err != nil || !authVersionValidated {
			setAuthCookies(c, nil)
			c.Next()
			return
		}

		c.Set(ContextUserID, claims.UserID)
		c.Set(ContextUsername, claims.Username)
		c.Next()
	}
}

func RateLimiter(r rate.Limit, burst int) gin.HandlerFunc {
	var (
		mu       sync.Mutex
		limiters = make(map[string]*ipLimiter)
	)

	go func() {
		ticker := time.NewTicker(time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			mu.Lock()
			for ip, l := range limiters {
				if time.Since(l.lastSeen) > 3*time.Minute {
					delete(limiters, ip)
				}
			}
			mu.Unlock()
		}
	}()

	getLimiter := func(ip string) *rate.Limiter {
		mu.Lock()
		defer mu.Unlock()

		l, exists := limiters[ip]
		if !exists {
			l = &ipLimiter{
				limiter: rate.NewLimiter(r, burst),
			}
			limiters[ip] = l
		}
		l.lastSeen = time.Now()
		return l.limiter
	}

	return func(c *gin.Context) {
		limiter := getLimiter(c.ClientIP())
		if !limiter.Allow() {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, errorResponse("rate limit exceeded"))
			return
		}
		c.Next()
	}
}
