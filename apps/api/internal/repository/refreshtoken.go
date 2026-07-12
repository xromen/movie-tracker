package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/xromen/movietracker/internal/domain"
)

var (
	ErrRefreshTokenNotFound = errors.New("refresh token not found")
	ErrRefreshTokenExpired  = errors.New("refresh token expired")
	ErrRefreshTokenReuse    = errors.New("refresh token reuse detected")
)

type RefreshTokenRepository interface {
	Create(ctx context.Context, token *domain.RefreshToken) error
	Rotate(ctx context.Context, oldTokenHash []byte, newToken *domain.RefreshToken) error
	RevokeByUserID(ctx context.Context, userID int64) error
}

type refreshTokenRepository struct {
	pool *pgxpool.Pool
}

func NewRefreshTokenRepository(pool *pgxpool.Pool) RefreshTokenRepository {
	return &refreshTokenRepository{pool: pool}
}

func (r *refreshTokenRepository) Create(ctx context.Context, token *domain.RefreshToken) error {
	const query = `
		INSERT INTO refresh_tokens (
			user_id,
			token_hash,
			family_id,
			expires_at
		)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at;
	`

	err := r.pool.QueryRow(
		ctx,
		query,
		token.UserID,
		token.TokenHash,
		token.FamilyID,
		token.ExpiresAt,
	).Scan(
		&token.ID,
		&token.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("create refresh token: %w", err)
	}

	return nil
}

func (r *refreshTokenRepository) Rotate(ctx context.Context, oldTokenHash []byte, newToken *domain.RefreshToken) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin refresh rotation: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	const findQuery = `
		SELECT
			id,
			user_id,
			token_hash,
			family_id,
			expires_at,
			created_at,
			revoked_at,
			replaced_by
		FROM refresh_tokens
		WHERE token_hash = $1
		FOR UPDATE;
	`

	old := &domain.RefreshToken{}
	err = tx.QueryRow(ctx, findQuery, oldTokenHash).Scan(
		&old.ID,
		&old.UserID,
		&old.TokenHash,
		&old.FamilyID,
		&old.ExpiresAt,
		&old.CreatedAt,
		&old.RevokedAt,
		&old.ReplacedBy,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrRefreshTokenNotFound
	}
	if err != nil {
		return fmt.Errorf("find refresh token for rotation: %w", err)
	}

	// Повторно предъявленный токен означает вероятную кражу.
	// Отзываем всю цепочку токенов данной сессии.
	if old.RevokedAt != nil {
		if err := revokeFamily(ctx, tx, old.FamilyID); err != nil {
			return err
		}
		if err := tx.Commit(ctx); err != nil {
			return fmt.Errorf("commit reused refresh token: %w", err)
		}
		return ErrRefreshTokenReuse
	}

	if !old.ExpiresAt.After(time.Now()) {
		return ErrRefreshTokenExpired
	}

	const createQuery = `
		INSERT INTO refresh_tokens (
			user_id,
			token_hash,
			family_id,
			expires_at
		)
		VALUES ($1, $2, $3, $4)
		RETURNING id, user_id, family_id, created_at;
	`

	err = tx.QueryRow(
		ctx,
		createQuery,
		old.UserID,
		newToken.TokenHash,
		old.FamilyID,
		newToken.ExpiresAt,
	).Scan(
		&newToken.ID,
		&newToken.UserID,
		&newToken.FamilyID,
		&newToken.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("create rotated refresh token: %w", err)
	}

	const revokeOldQuery = `
		UPDATE refresh_tokens
		SET revoked_at = now(),
		    replaced_by = $2
		WHERE id = $1;
	`

	if _, err := tx.Exec(ctx, revokeOldQuery, old.ID, newToken.ID); err != nil {
		return fmt.Errorf("revoke previous refresh token: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit refresh rotation: %w", err)
	}

	return nil
}

func revokeFamily(
	ctx context.Context,
	tx pgx.Tx,
	familyID string,
) error {
	const query = `
		UPDATE refresh_tokens
		SET revoked_at = now()
		WHERE family_id = $1
		  AND revoked_at IS NULL;
	`

	if _, err := tx.Exec(ctx, query, familyID); err != nil {
		return fmt.Errorf("revoke refresh token family: %w", err)
	}

	return nil
}

func (r *refreshTokenRepository) RevokeByUserID(
	ctx context.Context,
	userID int64,
) error {
	const query = `
		UPDATE refresh_tokens
		SET revoked_at = now()
		WHERE user_id = $1
		  AND revoked_at IS NULL;
	`

	if _, err := r.pool.Exec(ctx, query, userID); err != nil {
		return fmt.Errorf("revoke user refresh tokens: %w", err)
	}

	return nil
}