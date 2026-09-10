package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/xromen/movietracker/internal/domain"
)

type WorkerRepository interface {
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

type workerRepository struct {
	pool *pgxpool.Pool
}

func NewWorkerRepository(pool *pgxpool.Pool) WorkerRepository {
	return &workerRepository{pool: pool}
}

func (r *workerRepository) ListDue(ctx context.Context, limit int) ([]domain.TrackedMedia, error) {
	query := `
		SELECT DISTINCT
			m.id,
			m.tmdb_id,
			m.media_type
		FROM medias m
			JOIN user_medias um on m.id = um.media_id
		WHERE m.next_schedule_sync_at IS NULL
		   OR m.next_schedule_sync_at <= NOW()
		ORDER BY m.next_schedule_sync_at  NULLS FIRST
		LIMIT $1;
	`

	rows, err := r.pool.Query(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("get media list due: %w", err)
	}
	defer rows.Close()

	var result []domain.TrackedMedia

	for rows.Next() {
		var media domain.TrackedMedia

		if err := rows.Scan(&media.ID, &media.TMDBID, &media.Type); err != nil {
			return nil, fmt.Errorf("scan due media: %w", err)
		}

		result = append(result, media)
	}

	return result, rows.Err()
}

func (r *workerRepository) ReplaceMovieReleases(ctx context.Context, mediaID int64, region string, releases []domain.MovieRelease) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("replace movie releases begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, `
		DELETE FROM movie_releases
		WHERE media_id = $1 AND region = $2
	`, mediaID, region)
	if err != nil {
		return fmt.Errorf("replace movie releases delete releases: %w", err)
	}

	for _, release := range releases {
		_, err := r.pool.Exec(ctx, `
			INSERT INTO movie_releases (
				media_id,
				region,
				release_type,
				release_at,
				certification
			)
			VALUES ($1, $2, $3, $4, $5)
		`, 
		mediaID,
		region,
		release.Type,
		release.ReleaseAt,
		release.Certification,
		)
		if err != nil {
			return fmt.Errorf("replace movie releases insert release: %w", err)
		}
	}

	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("replace movie releases commit transaction: %w", err)
	}

	return nil
}

func (r *workerRepository) ReplaceSeasonEpisodes(ctx context.Context, mediaID int64, seasonNumber int, episodes []domain.Episode) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("replace season episodes begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, `
		DELETE FROM tv_episodes
		WHERE media_id = $1 AND season_number = $2
	`, mediaID, seasonNumber)
	if err != nil {
		return fmt.Errorf("replace season episodes delete episodes: %w", err)
	}

	for _, episode := range episodes {
		_, err = tx.Exec(ctx, `
			INSERT INTO tv_episodes (
				tmdb_episode_id,
				media_id,
				season_number,
				episode_number,
				title,
				overview,
				air_date,
				still_path,
				runtime
			)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)
		`,
			episode.ID,
			mediaID,
			episode.SeasonNumber,
			episode.EpisodeNumber,
			episode.Title,
			episode.Overview,
			episode.AirDate,
			episode.StillPath,
			episode.Runtime,
		)
		if err != nil {
			return fmt.Errorf("replace season episodes insert episode: %w", err)
		}
	}

	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("replace season episodes commit transaction: %w", err)
	}

	return nil
}

func (r *workerRepository) MarkSuccess(ctx context.Context, mediaID int64, nextSync time.Time) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE medias
		SET schedule_synced_at = NOW(),
		    next_schedule_sync_at = $2,
		    schedule_sync_error = NULL
		WHERE id = $1
	`, mediaID, nextSync)

	return err
}

func (r *workerRepository) MarkFailure(ctx context.Context, mediaID int64, nextRetry time.Time, message string) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE medias
		SET next_schedule_sync_at = $2,
		    schedule_sync_error = $3
		WHERE id = $1
	`, mediaID, nextRetry, message)

	return err
}
