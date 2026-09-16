package service

import (
	"context"
	"errors"
	"fmt"
	"html"
	"log/slog"
	"strings"
	"time"
	"unicode/utf16"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/xromen/movietracker/internal/domain"
	"github.com/xromen/movietracker/internal/platform/tmdb"
)

const (
	batchSize            = 100
	retryDelay           = time.Hour
	pollInterval         = time.Hour
	telegramMessageLimit = 3500
)

type workerRepository interface {
	CollectionListDue(ctx context.Context, limit int) ([]int64, error)
	ListDue(ctx context.Context, limit int) ([]domain.TrackedMedia, error)
	ListUsersDueForReport(ctx context.Context, limit int) ([]domain.DueReport, error)

	ReplaceMovieReleases(
		ctx context.Context,
		mediaID int64,
		region string,
		releases []domain.MovieRelease,
	) error

	UpdateMovieCollectionID(ctx context.Context, mediaID int64, collectionID *int64) error

	ReplaceSeasonEpisodes(
		ctx context.Context,
		mediaID int64,
		seasonNumber int,
		episodes []domain.Episode,
	) error

	UpsertCollection(ctx context.Context, collection domain.Collection) error

	MarkMediaSuccess(ctx context.Context, mediaID int64, nextSync time.Time) error
	MarkMediaFailure(ctx context.Context, mediaID int64, nextRetry time.Time, message string) error

	MarkCollectionSuccess(ctx context.Context, collectionID int64, nextSync time.Time) error
	MarkCollectionFailure(ctx context.Context, collectionID int64, nextRetry time.Time, message string) error

	GetMoviesForReport(ctx context.Context, userID int64, from, to time.Time) ([]domain.ReportMovie, error)
	GetTvEpisodesForReport(ctx context.Context, userID int64, from, to time.Time) ([]domain.ReportTvEpisode, error)

	CreateReport(ctx context.Context, report domain.Report, messages []string) error
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
		items, err := w.repo.CollectionListDue(ctx, batchSize)
		if err != nil {
			w.logger.Error("failed to list collection for schedule sync", "error", err)
			break
		}
		if len(items) == 0 {
			break
		}

		for _, collectionID := range items {
			if err := w.syncCollection(ctx, collectionID); err != nil {
				_ = w.repo.MarkCollectionFailure(
					ctx,
					collectionID,
					time.Now().Add(retryDelay),
					err.Error(),
				)
			}
		}
	}

	for {
		items, err := w.repo.ListDue(ctx, batchSize)
		if err != nil {
			w.logger.Error("failed to list media for schedule sync", "error", err)
			break
		}
		if len(items) == 0 {
			break
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

				_ = w.repo.MarkMediaFailure(
					ctx,
					item.ID,
					time.Now().Add(retryDelay),
					err.Error(),
				)
			}
		}
	}

	for {
		items, err := w.repo.ListUsersDueForReport(ctx, batchSize)
		if err != nil {
			w.logger.Error("failed to get users due for report", "error", err)
			break
		}
		if len(items) == 0 {
			break
		}

		for _, item := range items {
			if err := w.createReport(ctx, item); err != nil {
				w.logger.Warn(
					"failed to create report",
					"user_id", item.UserID,
					"error", err,
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

func (w *Worker) syncCollection(ctx context.Context, id int64) error {
	collection, err := w.tmdb.GetCollectionDetails(ctx, id)
	if err != nil {
		return fmt.Errorf("get collection details: %w", err)
	}

	if err := w.repo.UpsertCollection(ctx, *collection); err != nil {
		return fmt.Errorf("save collection: %w", err)
	}

	nextSync := time.Now().Add(7 * 24 * time.Hour)

	return w.repo.MarkCollectionSuccess(ctx, id, nextSync)
}

func (w *Worker) syncMovie(ctx context.Context, media domain.TrackedMedia) error {
	releases, err := w.tmdb.GetMovieReleases(ctx, media.TMDBID, "RU")
	if err != nil {
		return fmt.Errorf("sync movie get movie releases: %w", err)
	}

	if err := w.repo.ReplaceMovieReleases(
		ctx,
		media.ID,
		"RU",
		releases,
	); err != nil {
		return fmt.Errorf("sync movie save movie releases: %w", err)
	}

	details, err := w.tmdb.GetMovieDetails(ctx, media.TMDBID)
	if err != nil {
		return fmt.Errorf("sync movie get movie details: %w", err)
	}

	if details.CollectionID != media.CollectionID {
		if details.CollectionID == nil {
			err := w.repo.UpdateMovieCollectionID(ctx, media.ID, nil)

			if err != nil {
				return fmt.Errorf("sync movie update collection id: %w", err)
			}
		} else {
			err := w.syncCollection(ctx, *details.CollectionID)

			if err != nil {
				return fmt.Errorf("sync movie sync collection: %w", err)
			}
		}
	}

	nextSync := nextMovieSync(time.Now(), releases)
	return w.repo.MarkMediaSuccess(ctx, media.ID, nextSync)
}

func (w *Worker) syncTV(ctx context.Context, media domain.TrackedMedia) error {
	schedule, err := w.tmdb.GetTVSchedule(ctx, media.TMDBID)
	if err != nil {
		return fmt.Errorf("sync tv get tv schedule: %w", err)
	}

	for _, seasonNumber := range schedule.SeasonNumbers {
		episodes, err := w.tmdb.GetTvSeasonEpisodes(
			ctx,
			media.TMDBID,
			seasonNumber,
			0,
		)
		if err != nil {
			return fmt.Errorf("sync tv get season %d: %w", seasonNumber, err)
		}

		if err := w.repo.ReplaceSeasonEpisodes(
			ctx,
			media.ID,
			seasonNumber,
			episodes.Items,
		); err != nil {
			return fmt.Errorf("sync tv save season %d: %w", seasonNumber, err)
		}
	}

	nextSync := time.Now().Add(24 * time.Hour)
	if schedule.Status == "Ended" || schedule.Status == "Canceled" {
		nextSync = time.Now().Add(30 * 24 * time.Hour)
	}

	return w.repo.MarkMediaSuccess(ctx, media.ID, nextSync)
}

func (w *Worker) createReport(ctx context.Context, report domain.DueReport) error {
	interval := report.PeriodTo.Sub(report.PeriodFrom)
	periodFuture := report.PeriodTo.Add(interval)

	moviesForReport, err := w.repo.GetMoviesForReport(ctx, report.UserID, report.PeriodFrom, periodFuture)
	if err != nil {
		return fmt.Errorf("get movies for report: %w", err)
	}

	episodesForReport, err := w.repo.GetTvEpisodesForReport(ctx, report.UserID, report.PeriodFrom, periodFuture)
	if err != nil {
		return fmt.Errorf("get episodes for report: %w")
	}

	messages := buildReportMessages(report.PeriodFrom, report.PeriodTo, periodFuture, moviesForReport, episodesForReport)

	err = w.repo.CreateReport(
		ctx,
		domain.Report{
			UserID:               report.UserID,
			PeriodFrom:           report.PeriodFrom,
			PeriodTo:             report.PeriodTo,
			NextScheduleCreateAt: time.Now().Add(interval),
		},
		messages,
	)
	if err != nil {
		return fmt.Errorf("create report: %w", err)
	}

	return nil
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

func buildReportMessages(
	periodFrom time.Time,
	periodTo time.Time,
	futureTo time.Time,
	movies []domain.ReportMovie,
	episodes []domain.ReportTvEpisode,
) []string {
	if len(movies) == 0 && len(episodes) == 0 {
		return nil
	}

	blocks := []string{
		fmt.Sprintf(
			"🎬 <b>Киноотчёт</b>\n\n"+
				"📌 Вышло: %s–%s\n"+
				"🔭 Ожидается: %s–%s",
			periodFrom.Format("02.01"),
			periodTo.Format("02.01"),
			periodTo.Format("02.01"),
			futureTo.Format("02.01"),
		),
	}

	appendSection := func(heading string, items []string) {
		if len(items) == 0 {
			return
		}

		blocks = append(blocks, heading+"\n"+items[0])
		blocks = append(blocks, items[1:]...)
	}

	var releasedMovies []string
	var upcomingMovies []string

	for _, movie := range movies {
		line := fmt.Sprintf(
			"• <b>%s</b> — %s",
			html.EscapeString(movie.Title),
			movie.ReleaseAt.Format("02.01.2006"),
		)

		if movie.ReleaseAt.Before(periodTo) {
			releasedMovies = append(releasedMovies, line)
		} else {
			upcomingMovies = append(upcomingMovies, line)
		}
	}

	appendSection(
		"🍿 <b>Вышедшие фильмы</b>",
		releasedMovies,
	)
	appendSection(
		"📅 <b>Предстоящие фильмы</b>",
		upcomingMovies,
	)

	var releasedEpisodes []string
	var upcomingEpisodes []string

	for _, episode := range episodes {
		episodeString := fmt.Sprintf(
			"S%02dE%02d",
			episode.SeasonNumber,
			episode.EpisodeNumber,
		)

		if episode.EpisodeTitle != "" {
			episodeString += fmt.Sprintf(
				" «%s»",
				html.EscapeString(episode.EpisodeTitle),
			)
		}

		line := fmt.Sprintf(
			"• <b>%s</b> — %s — %s",
			html.EscapeString(episode.SeriesTitle),
			episodeString,
			episode.EpisodeAirDate.Format("02.01.2006"),
		)

		if episode.EpisodeAirDate.Before(periodTo) {
			releasedEpisodes = append(releasedEpisodes, line)
		} else {
			upcomingEpisodes = append(upcomingEpisodes, line)
		}
	}

	appendSection(
		"🆕 <b>Вышедшие серии</b>",
		releasedEpisodes,
	)
	appendSection(
		"⏳ <b>Предстоящие серии</b>",
		upcomingEpisodes,
	)

	return splitReportMessages(blocks)
}

func splitReportMessages(blocks []string) []string {
	var messages []string
	var message strings.Builder

	for _, block := range blocks {
		separator := ""
		if message.Len() > 0 {
			separator = "\n\n"
		}

		if telegramTextLength(message.String()+separator+block) > telegramMessageLimit {
			messages = append(messages, message.String())
			message.Reset()
			message.WriteString("🎬 <b>Киноотчёт — продолжение</b>\n\n")
		}

		message.WriteString(separator + block)
	}

	if message.Len() > 0 {
		messages = append(messages, message.String())
	}

	return messages
}

func telegramTextLength(value string) int {
	return len(utf16.Encode([]rune(value)))
}
