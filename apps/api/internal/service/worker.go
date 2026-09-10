package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/xromen/movietracker/internal/domain"
	"github.com/xromen/movietracker/internal/platform/tmdb"
)

const (
	batchSize    = 100
	retryDelay   = time.Hour
	pollInterval = time.Hour
)

type workerRepository interface {
	ListDue(ctx context.Context, limit int) ([]domain.TrackedMedia, error)

	ReplaceMovieReleases(
		ctx context.Context,
		mediaID int64,
		region string,
		releases []domain.MovieRelease,
	) error

	ReplaceSeasonEpisodes(
		ctx context.Context,
		mediaID int64,
		seasonNumber int,
		episodes []domain.Episode,
	) error

	MarkSuccess(ctx context.Context, mediaID int64, nextSync time.Time) error
	MarkFailure(ctx context.Context, mediaID int64, nextRetry time.Time, message string) error
}

type Worker struct {
	repo   workerRepository
	pool   *pgxpool.Pool
	tmdb   tmdb.Client
	logger *slog.Logger
}

func NewWorker(
	repo workerRepository,
	pool *pgxpool.Pool,
	tmdb tmdb.Client,
	logger *slog.Logger,
) *Worker {
	return &Worker{
		repo:   repo,
		pool:   pool,
		tmdb:   tmdb,
		logger: logger,
	}
}

func (w *Worker) Run(ctx context.Context) error {
	w.logger.Info("worker started")

	w.runOnce(ctx)

	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil

		case <-ticker.C:
			w.runOnce(ctx)
		}
	}
}

func (w *Worker) runOnce(ctx context.Context) {
	unlock, acquired, err := w.tryLock(ctx)
	if err != nil {
		w.logger.Error("failed to acquire schedule lock", "error", err)
		return
	}
	if !acquired {
		return
	}
	defer unlock()

	for {
		items, err := w.repo.ListDue(ctx, batchSize)
		if err != nil {
			w.logger.Error("failed to list media for schedule sync", "error", err)
			return
		}
		if len(items) == 0 {
			return
		}

		for _, item := range items {
			if err := w.syncMedia(ctx, item); err != nil {
				w.logger.Warn(
					"failed to sync media schedule",
					"media_id", item.ID,
					"tmdb_id", item.TMDBID,
					"type", item.Type,
					"error", err,
				)

				_ = w.repo.MarkFailure(
					ctx,
					item.ID,
					time.Now().Add(retryDelay),
					err.Error(),
				)
			}
		}
	}
}

func (w *Worker) syncMedia(ctx context.Context, media domain.TrackedMedia) error {
	switch media.Type {
	case domain.MediaTypeMovie:
		return w.syncMovie(ctx, media)

	case domain.MediaTypeTV:
		return w.syncTV(ctx, media)

	default:
		return fmt.Errorf("unsupported media type %q", media.Type)
	}
}

func (w *Worker) syncMovie(ctx context.Context, media domain.TrackedMedia) error {
	releases, err := w.tmdb.GetMovieReleases(ctx, media.TMDBID, "RU")
	if err != nil {
		return fmt.Errorf("get movie releases: %w", err)
	}

	if err := w.repo.ReplaceMovieReleases(
		ctx,
		media.ID,
		"RU",
		releases,
	); err != nil {
		return fmt.Errorf("save movie releases: %w", err)
	}

	nextSync := nextMovieSync(time.Now(), releases)
	return w.repo.MarkSuccess(ctx, media.ID, nextSync)
}

func (w *Worker) syncTV(ctx context.Context, media domain.TrackedMedia) error {
	schedule, err := w.tmdb.GetTVSchedule(ctx, media.TMDBID)
	if err != nil {
		return fmt.Errorf("get tv schedule: %w", err)
	}

	for _, seasonNumber := range schedule.SeasonNumbers {
		episodes, err := w.tmdb.GetTvSeasonEpisodes(
			ctx,
			media.TMDBID,
			seasonNumber,
			0,
		)
		if err != nil {
			return fmt.Errorf("get season %d: %w", seasonNumber, err)
		}

		if err := w.repo.ReplaceSeasonEpisodes(
			ctx,
			media.ID,
			seasonNumber,
			episodes.Items,
		); err != nil {
			return fmt.Errorf("save season %d: %w", seasonNumber, err)
		}
	}

	nextSync := time.Now().Add(24 * time.Hour)
	if schedule.Status == "Ended" || schedule.Status == "Canceled" {
		nextSync = time.Now().Add(30 * 24 * time.Hour)
	}

	return w.repo.MarkSuccess(ctx, media.ID, nextSync)
}

func (w *Worker) tryLock(
	ctx context.Context,
) (unlock func(), acquired bool, err error) {
	conn, err := w.pool.Acquire(ctx)
	if err != nil {
		return nil, false, err
	}

	var locked bool
	err = conn.QueryRow(
		ctx,
		`SELECT pg_try_advisory_lock(hashtext('tmdb-schedule-sync'))`,
	).Scan(&locked)

	if err != nil || !locked {
		conn.Release()
		return func() {}, locked, err
	}

	unlock = func() {
		cleanupCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		_, unlockErr := conn.Exec(
			cleanupCtx,
			`SELECT pg_advisory_unlock(hashtext('tmdb-schedule-sync'))`,
		)
		if unlockErr != nil && !errors.Is(unlockErr, context.Canceled) {
			w.logger.Error("failed to release schedule lock", "error", unlockErr)
		}

		conn.Release()
	}

	return unlock, true, nil
}

func nextMovieSync(now time.Time, releases []domain.MovieRelease) time.Time {
	var latest time.Time

	for _, release := range releases {
		if release.ReleaseAt.After(latest) {
			latest = release.ReleaseAt
		}
	}

	if latest.IsZero() || latest.After(now.Add(-30*24*time.Hour)) {
		return now.Add(24 * time.Hour)
	}

	return now.Add(30 * 24 * time.Hour)
}
