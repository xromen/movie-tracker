package worker

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"
	"github.com/xromen/movietracker/internal/config"
	"github.com/xromen/movietracker/internal/platform/database"
	"github.com/xromen/movietracker/internal/platform/tmdb"
	"github.com/xromen/movietracker/internal/repository"
	"github.com/xromen/movietracker/internal/service"
)

func main() {
	_ = godotenv.Load()
	
	ctx, stop := signal.NotifyContext(
		context.Background(),
		syscall.SIGINT,
		syscall.SIGTERM,
	)
	defer stop()

	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	pool, err := database.NewPool(ctx, cfg.Database.PoolConfig())
	if err != nil {
		slog.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	tmdbClient, err := tmdb.NewClient(tmdb.Config{
		BaseURL:       cfg.TMDB.BaseURL,
		BearerToken:   cfg.TMDB.BearerToken,
		ImagesBaseURL: cfg.TMDB.ImagesBaseURL,
		Timeout:       cfg.TMDB.Timeout,
		RPM:           cfg.TMDB.RPM,
		Burst:         cfg.TMDB.Burst,
	})
	if err != nil {
		slog.Error("failed to create TMDB client", "error", err)
		os.Exit(1)
	}

	repo := repository.NewWorkerRepository(pool)
	worker := service.NewWorker(repo, pool, tmdbClient, slog.Default())

	if err := worker.Run(ctx); err != nil {
		slog.Error("worker stopped", "error", err)
		os.Exit(1)
	}
}
