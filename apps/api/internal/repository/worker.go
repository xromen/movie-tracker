package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/xromen/movietracker/internal/domain"
)

type WorkerRepository interface {
	CollectionListDue(ctx context.Context, limit int) ([]int64, error)
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

	UpdateMovieCollectionID(ctx context.Context, movieID int64, collectionID *int64) error

	UpsertCollection(ctx context.Context, collection domain.Collection) error

	MarkMediaSuccess(ctx context.Context, mediaID int64, nextSync time.Time) error
	MarkMediaFailure(ctx context.Context, mediaID int64, nextRetry time.Time, message string) error

	MarkCollectionSuccess(ctx context.Context, collectionID int64, nextSync time.Time) error
	MarkCollectionFailure(ctx context.Context, collectionID int64, nextRetry time.Time, message string) error
}

type workerRepository struct {
	pool *pgxpool.Pool
}

func NewWorkerRepository(pool *pgxpool.Pool) WorkerRepository {
	return &workerRepository{pool: pool}
}

func (r *workerRepository) CollectionListDue(ctx context.Context, limit int) ([]int64, error) {
	query := `
		SELECT DISTINCT
			m.collection_tmdb_id,
    		mc.next_schedule_sync_at
		FROM medias m
			JOIN user_medias um on m.id = um.media_id
			LEFT JOIN movie_collections mc on m.collection_tmdb_id = mc.tmdb_id
		WHERE (mc.next_schedule_sync_at IS NULL
		   OR mc.next_schedule_sync_at <= NOW())
		  AND m.collection_tmdb_id IS NOT NULL
		ORDER BY mc.next_schedule_sync_at NULLS FIRST
		LIMIT $1;
	`

	rows, err := r.pool.Query(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("get collection list due: %w", err)
	}
	defer rows.Close()

	var result []int64

	for rows.Next() {
		var collectionID int64
		var nextScheduleSyncAt *time.Time

		if err := rows.Scan(&collectionID, &nextScheduleSyncAt); err != nil {
			return nil, fmt.Errorf("scan collection id: %w", err)
		}

		result = append(result, collectionID)
	}

	return result, rows.Err()
}

func (r *workerRepository) ListDue(ctx context.Context, limit int) ([]domain.TrackedMedia, error) {
	query := `
		SELECT
			m.id,
			m.tmdb_id,
			m.media_type,
			m.collection_tmdb_id,
			m.next_schedule_sync_at
		FROM medias m
		WHERE (m.next_schedule_sync_at IS NULL OR m.next_schedule_sync_at <= NOW())
		AND (EXISTS (SELECT
						1
					FROM user_medias um
					WHERE um.media_id = m.id)
			OR EXISTS (SELECT
						1
					FROM user_medias um
								JOIN medias m2 ON m2.id = um.media_id
					WHERE m2.collection_tmdb_id = m.collection_tmdb_id)
			)
		ORDER BY m.next_schedule_sync_at NULLS FIRST
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
		var nextScheduleSyncAt *time.Time

		if err := rows.Scan(&media.ID, &media.TMDBID, &media.Type, &media.CollectionID, &nextScheduleSyncAt); err != nil {
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
		_, err := tx.Exec(ctx, `
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

func (r *workerRepository) UpsertCollection(ctx context.Context, collection domain.Collection) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("upsert collection begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, `
		INSERT INTO movie_collections(tmdb_id, name)
		VALUES ($1, $2)
		ON CONFLICT(tmdb_id) DO UPDATE SET
							 name = EXCLUDED.name;
	`, collection.ID, collection.Name)

	for _, part := range collection.Parts {
		_, err := tx.Exec(ctx, `
			INSERT INTO medias(tmdb_id, title, overview, poster_path, release_date, vote_average, media_type, collection_tmdb_id)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
			ON CONFLICT(tmdb_id, media_type) DO UPDATE SET
											 title = EXCLUDED.title,
											 overview = EXCLUDED.overview,
											 poster_path = EXCLUDED.poster_path,
											 collection_tmdb_id = EXCLUDED.collection_tmdb_id,
											 updated_at = NOW();
		`,
			part.ID,
			part.Title,
			part.Overview,
			part.PosterPath,
			part.ReleaseDate,
			part.VoteAverage,
			part.MediaType,
			collection.ID,
		)

		if err != nil {
			return fmt.Errorf("upsert collection insert collection %d: %w", collection.ID, err)
		}
	}

	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("upsert collection commit transaction: %w", err)
	}

	return nil
}

func (r *workerRepository) UpdateMovieCollectionID(ctx context.Context, mediaID int64, collectionID *int64) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE medias
		SET collection_tmdb_id = $2
		WHERE id = $1;
	`, mediaID, collectionID)

	return err
}

func (r *workerRepository) MarkMediaSuccess(ctx context.Context, mediaID int64, nextSync time.Time) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE medias
		SET schedule_synced_at = NOW(),
		    next_schedule_sync_at = $2,
		    schedule_sync_error = NULL
		WHERE id = $1;
	`, mediaID, nextSync)

	return err
}

func (r *workerRepository) MarkMediaFailure(ctx context.Context, mediaID int64, nextRetry time.Time, message string) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE medias
		SET next_schedule_sync_at = $2,
		    schedule_sync_error = $3
		WHERE id = $1;
	`, mediaID, nextRetry, message)

	return err
}

func (r *workerRepository) MarkCollectionSuccess(ctx context.Context, collectionID int64, nextSync time.Time) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE movie_collections
		SET schedule_synced_at = NOW(),
		    next_schedule_sync_at = $2,
		    schedule_sync_error = NULL
		WHERE tmdb_id = $1;
	`, collectionID, nextSync)

	return err
}

func (r *workerRepository) MarkCollectionFailure(ctx context.Context, collectionID int64, nextRetry time.Time, message string) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE movie_collections
		SET next_schedule_sync_at = $2,
		    schedule_sync_error = $3
		WHERE tmdb_id = $1;
	`, collectionID, nextRetry, message)

	return err
}
