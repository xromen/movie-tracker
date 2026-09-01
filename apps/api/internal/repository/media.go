package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/xromen/movietracker/internal/domain"
)

type MediaRepository interface {
	Upsert(ctx context.Context, media *domain.Media) error
	GetByTmdbID(ctx context.Context, tmdbID int64, mediaType domain.MediaType) (*domain.Media, error)
	GetUserList(ctx context.Context, userID int64, status domain.WatchStatus, mediaType domain.MediaType, page, perPage int) ([]domain.UserMedia, int, error)
	GetMediaUserStatus(ctx context.Context, mediaType domain.MediaType, userID, mediaID int64) (*domain.WatchStatus, error)
	GetMediaUserStatuses(ctx context.Context, mediaType domain.MediaType, userID int64, mediaIDs []int64) (map[int64]domain.WatchStatus, error)
	DeleteUserStatus(ctx context.Context, mediaType domain.MediaType, userID, mediaID int64) (*domain.UserMedia, error)
	SetMediaUserStatus(ctx context.Context, userID int64, media *domain.UserMedia) error
	GetWatchedEpisodeNumbers(ctx context.Context, userID, tvShowID int64, seasonNumber int) ([]int, error)
	SetEpisodeWatched(ctx context.Context, userID, tvShowID int64, seasonNumber, episodeNumber int, watched bool) error
	MarkSeasonWatched(ctx context.Context, userID, tvShowID int64, seasonNumber int, episodeNumbers []int32) error
	UnmarkSeasonWatched(ctx context.Context, userID, tvShowID int64, seasonNumber int) error
}

type mediaRepository struct {
	pool *pgxpool.Pool
}

func NewMediaRepository(pool *pgxpool.Pool) MediaRepository {
	return &mediaRepository{pool: pool}
}

func (r *mediaRepository) Upsert(ctx context.Context, media *domain.Media) error {
	query := `
		INSERT INTO medias (tmdb_id, title, overview, poster_path, release_date, media_type, vote_average)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT(tmdb_id, media_type) DO UPDATE SET
										 title = EXCLUDED.title,
										 overview = EXCLUDED.overview,
										 poster_path = EXCLUDED.poster_path,
										 updated_at = NOW()
	`

	_, err := r.pool.Exec(ctx, query,
		media.ID,
		media.Title,
		media.Overview,
		media.PosterPath,
		media.ReleaseDate,
		media.Type,
		media.VoteAverage,
	)

	if err != nil {
		return fmt.Errorf("upsert media: %w", err)
	}

	return nil
}

func (r *mediaRepository) GetByTmdbID(ctx context.Context, tmdbID int64, mediaType domain.MediaType) (*domain.Media, error) {
	query := `
		SELECT tmdb_id, title, overview, poster_path, release_date, media_type
		FROM medias 
		WHERE tmdb_id = $1
		  AND media_type = $2
	`

	media := &domain.Media{}
	err := r.pool.QueryRow(ctx, query, tmdbID, mediaType).Scan(
		&media.ID, &media.Title, &media.Overview, &media.PosterPath,
		&media.ReleaseDate, &media.Type)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("get media by tmdb_id: %w", domain.ErrNotFound)
		}
		return nil, fmt.Errorf("get media by tmdb_id: %w", err)
	}

	return media, nil
}

func (r *mediaRepository) SetMediaUserStatus(ctx context.Context, userID int64, um *domain.UserMedia) error {
	query := `
		INSERT INTO user_medias (user_id, media_id, status, rating)
		SELECT
			$1,
			m.id,
			$3,
			$4
		FROM medias m
		WHERE m.tmdb_id = $2
		ON CONFLICT (user_id, media_id) DO UPDATE SET status     = EXCLUDED.status,
													  rating     = EXCLUDED.rating,
													  watched_at = EXCLUDED.watched_at,
													  updated_at = NOW();
	`

	_, err := r.pool.Exec(ctx, query,
		userID, um.Media.ID, um.Status, um.Rating,
	)

	if err != nil {
		return fmt.Errorf("add to user list: %w", err)
	}

	return nil
}

func (r *mediaRepository) GetUserList(
	ctx context.Context,
	userID int64,
	status domain.WatchStatus,
	mediaType domain.MediaType,
	page, perPage int,
) ([]domain.UserMedia, int, error) {
	countQuery := `
		SELECT 
			COUNT(*) 
		FROM user_medias um
			JOIN medias m on um.media_id = m.id
		WHERE um.user_id = $1 AND ($2::watch_status IS NULL OR um.status = $2) AND ($3::media_type IS NULL OR m.media_type = $3)
	`

	var total int
	var statusParam *string
	if status != "" {
		s := string(status)
		statusParam = &s
	}
	var mediaTypeParam *string
	if mediaType != "" {
		s := string(mediaType)
		mediaTypeParam = &s
	}

	err := r.pool.QueryRow(ctx, countQuery, userID, statusParam, mediaTypeParam).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("count user medias: %w", err)
	}

	if total == 0 {
		return make([]domain.UserMedia, 0), 0, nil
	}

	offset := (page - 1) * perPage

	query := `
		SELECT
			um.status, um.rating,
			m.tmdb_id, m.title, m.overview, m.poster_path, m.release_date, m.media_type, m.vote_average
		FROM user_medias um
			JOIN medias m ON m.id = um.media_id
		WHERE um.user_id = $1 AND ($2::media_type IS NULL OR m.media_type = $2) AND ($3::watch_status IS NULL OR um.status = $3)
		ORDER BY um.created_at DESC
		LIMIT $4 OFFSET $5
	`

	rows, err := r.pool.Query(ctx, query, userID, mediaTypeParam, statusParam, perPage, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("get user medias: %w", err)
	}

	defer rows.Close()

	var userMedias []domain.UserMedia

	for rows.Next() {
		var um domain.UserMedia
		var m domain.Media

		err := rows.Scan(
			&um.Status, &um.Rating,
			&m.ID, &m.Title, &m.Overview, &m.PosterPath, &m.ReleaseDate, &m.Type, &m.VoteAverage,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("scan user media: %w", err)
		}

		um.Media = &m
		userMedias = append(userMedias, um)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("iterate user medias: %w", err)
	}

	return userMedias, total, nil
}

func (r *mediaRepository) GetMediaUserStatus(ctx context.Context, mediaType domain.MediaType, userID, mediaID int64) (*domain.WatchStatus, error) {
	query := `
		SELECT status
		FROM user_medias um
			join medias m on um.media_id = m.id
		WHERE um.user_id = $1 AND m.tmdb_id = $2 AND m.media_type = $3
	`

	var status string
	err := r.pool.QueryRow(ctx, query, userID, mediaID, mediaType).Scan(&status)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("get user status for media: %w", domain.ErrNotFound)
		}
		return nil, fmt.Errorf("get user status for media: %w", err)
	}

	watchStatus := domain.WatchStatus(status)
	if !watchStatus.IsValid() {
		return nil, fmt.Errorf("invalid user watch status: %s", status)
	}

	return &watchStatus, nil
}

func (r *mediaRepository) GetMediaUserStatuses(ctx context.Context, mediaType domain.MediaType, userID int64, mediaIDs []int64) (map[int64]domain.WatchStatus, error) {
	query := `
		SELECT tmdb_id, status
		FROM user_medias um
			join medias m on um.media_id = m.id
		WHERE um.user_id = $1
		  AND m.tmdb_id = ANY($2)
		  AND m.media_type = $3
	`

	rows, err := r.pool.Query(ctx, query, userID, mediaIDs, mediaType)
	if err != nil {
		return nil, fmt.Errorf("get user media statuses: %w", err)
	}

	defer rows.Close()

	var statuses = map[int64]domain.WatchStatus{}

	for rows.Next() {
		var mediaID int64
		var status string

		err := rows.Scan(&mediaID, &status)
		if err != nil {
			return nil, fmt.Errorf("scan user media statuses: %w", err)
		}

		statuses[mediaID] = domain.WatchStatus(status)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate user media statuses: %w", err)
	}

	return statuses, nil
}

func (r *mediaRepository) DeleteUserStatus(ctx context.Context, mediaType domain.MediaType, userID, mediaID int64) (*domain.UserMedia, error) {
	query := `
		WITH deleted_rows AS (
		DELETE FROM user_medias um
			USING medias m
			WHERE um.media_id = m.id AND um.user_id = $1 AND m.tmdb_id = $2 AND m.media_type = $3
			RETURNING
				um.status,
				um.rating,
				m.tmdb_id,
				m.title,
				m.overview,
				m.poster_path,
				m.release_date,
				m.media_type,
				m.vote_average)
	SELECT
		*
	FROM deleted_rows d;
	`

	var result domain.UserMedia
	var media domain.Media

	err := r.pool.QueryRow(ctx, query, userID, mediaID, mediaType).Scan(
		&result.Status, &result.Rating,
		&media.ID, &media.Title, &media.Overview, &media.PosterPath, &media.ReleaseDate, &media.Type, &media.VoteAverage,
	)

	result.Media = &media

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("delete user status: %w", domain.ErrNotFound)
		}
		return nil, fmt.Errorf("delete user status: %w", err)
	}

	return &result, nil
}

func (r *mediaRepository) GetWatchedEpisodeNumbers(ctx context.Context, userID, tvShowID int64, seasonNumber int) ([]int, error) {
	query := `
		SELECT 
			episode_number
		FROM user_tv_episodes
		WHERE user_id = $1 AND tv_show_tmdb_id = $2 AND season_number = $3
	`

	rows, err := r.pool.Query(ctx, query, userID, tvShowID, seasonNumber)
	if err != nil {
		return nil, fmt.Errorf("get watched episode numbers: %w", err)
	}
	defer rows.Close()

	episodeNumbers := make([]int, 0)
	for rows.Next() {
		var episodeNumber int
		if err := rows.Scan(&episodeNumber); err != nil {
			return nil, fmt.Errorf("scan watched episode number: %w", err)
		}
		episodeNumbers = append(episodeNumbers, episodeNumber)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate watched episode numbers: %w", err)
	}

	return episodeNumbers, nil
}

func (r *mediaRepository) SetEpisodeWatched(ctx context.Context, userID, tvShowID int64, seasonNumber, episodeNumber int, watched bool) error {
	if watched {
		query := `
			INSERT INTO user_tv_episodes (
				user_id,
				tv_show_tmdb_id,
				season_number,
				episode_number
			)
			VALUES ($1, $2, $3, $4)
			ON CONFLICT (user_id, tv_show_tmdb_id, season_number, episode_number)
			DO UPDATE SET
				watched_at = NOW()
		`

		_, err := r.pool.Exec(
			ctx,
			query,
			userID,
			tvShowID,
			seasonNumber,
			episodeNumber,
		)
		if err != nil {
			return fmt.Errorf("mark episode watched: %w", err)
		}

		return nil
	}

	query := `
		DELETE FROM user_tv_episodes
		WHERE user_id = $1
		  AND tv_show_tmdb_id = $2
		  AND season_number = $3
		  AND episode_number = $4
	`

	_, err := r.pool.Exec(
		ctx,
		query,
		userID,
		tvShowID,
		seasonNumber,
		episodeNumber,
	)
	if err != nil {
		return fmt.Errorf("unmark episode watched: %w", err)
	}

	return nil
}

func (r *mediaRepository) MarkSeasonWatched(ctx context.Context, userID, tvShowID int64, seasonNumber int, episodeNumbers []int32) error {
	if len(episodeNumbers) == 0 {
		return nil
	}

	query := `
		INSERT INTO user_tv_episodes (
			user_id,
			tv_show_tmdb_id,
			season_number,
			episode_number,
			watched_at
		)
		SELECT $1, $2, $3, episode_number, NOW()
		FROM unnest($4::integer[]) AS episode_number
		ON CONFLICT (user_id, tv_show_tmdb_id, season_number, episode_number)
		DO UPDATE SET
			watched_at = NOW()
	`

	_, err := r.pool.Exec(
		ctx,
		query,
		userID,
		tvShowID,
		seasonNumber,
		episodeNumbers,
	)
	if err != nil {
		return fmt.Errorf("mark season watched: %w", err)
	}

	return nil
}

func (r *mediaRepository) UnmarkSeasonWatched(ctx context.Context, userID, tvShowID int64, seasonNumber int) error {
	query := `
		DELETE FROM user_tv_episodes
		WHERE user_id = $1
		  AND tv_show_tmdb_id = $2
		  AND season_number = $3
	`

	_, err := r.pool.Exec(ctx, query, userID, tvShowID, seasonNumber)
	if err != nil {
		return fmt.Errorf("unmark season watched: %w", err)
	}

	return nil
}
