package handler

import (
	"context"
	"net/http"
	"runtime"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/xromen/movietracker/internal/platform/cache"
)

type HealthHandler struct {
	db        *pgxpool.Pool
	cache     cache.Cache
	version   string
	startedAt time.Time
}

func NewHealthHandler(db *pgxpool.Pool, cache cache.Cache, version string) *HealthHandler {
	return &HealthHandler{
		db:        db,
		cache:     cache,
		version:   version,
		startedAt: time.Now(),
	}
}

// Live — liveness probe.
// Отвечает 200 пока процесс жив. Никаких проверок зависимостей.
// Если этот эндпоинт не отвечает — процесс завис и его нужно перезапустить.
func (h *HealthHandler) Live(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
	})
}

func (h *HealthHandler) Ready(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 3*time.Second)
	defer cancel()

	type depStatus struct {
		Status               string `json:"status"`
		Latency              string `json:"latency,omitempty"`
		Error                string `json:"error,omitempty"`
		TotalConns           int32  `json:"total_conns,omitempty"`
		AcquiredConns        int32  `json:"acquired_conns,omitempty"`
		IdleConns            int32  `json:"idle_conns,omitempty"`
		EmptyAcquireCount    int64  `json:"empty_acquire_count,omitempty"`
		EmptyAcquireWaitTime string `json:"empty_acquire_wait_time,omitempty"`
	}

	type response struct {
		Status       string               `json:"status"`
		Version      string               `json:"version"`
		Uptime       string               `json:"uptime"`
		Goroutines   int                  `json:"goroutines"`
		Dependencies map[string]depStatus `json:"dependencies"`
	}

	deps := make(map[string]depStatus)
	allOK := true

	// Проверяем PostgreSQL.
	start := time.Now()
	poolStat := h.db.Stat()
	if err := h.db.Ping(ctx); err != nil {
		deps["postgres"] = depStatus{
			Status:               "unavailable",
			Error:                err.Error(),
			TotalConns:           poolStat.TotalConns(),
			AcquiredConns:        poolStat.AcquiredConns(),
			IdleConns:            poolStat.IdleConns(),
			EmptyAcquireCount:    poolStat.EmptyAcquireCount(),
			EmptyAcquireWaitTime: poolStat.EmptyAcquireWaitTime().String(),
		}
		allOK = false
	} else {
		deps["postgres"] = depStatus{
			Status:               "ok",
			Latency:              time.Since(start).String(),
			TotalConns:           poolStat.TotalConns(),
			AcquiredConns:        poolStat.AcquiredConns(),
			IdleConns:            poolStat.IdleConns(),
			EmptyAcquireCount:    poolStat.EmptyAcquireCount(),
			EmptyAcquireWaitTime: poolStat.EmptyAcquireWaitTime().String(),
		}
	}

	// Проверяем Redis.
	start = time.Now()
	if err := h.cache.Ping(ctx); err != nil {
		deps["redis"] = depStatus{Status: "unavailable", Error: err.Error()}
		allOK = false
	} else {
		deps["redis"] = depStatus{
			Status:  "ok",
			Latency: time.Since(start).String(),
		}
	}

	status := "ok"
	httpStatus := http.StatusOK
	if !allOK {
		status = "degraded"
		httpStatus = http.StatusServiceUnavailable
	}

	c.JSON(httpStatus, response{
		Status:       status,
		Version:      h.version,
		Uptime:       time.Since(h.startedAt).String(),
		Goroutines:   runtime.NumGoroutine(),
		Dependencies: deps,
	})
}
