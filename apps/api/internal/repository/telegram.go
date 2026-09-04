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

type TelegramRepository interface {
	CreateBindingToken(ctx context.Context, token *domain.BindingToken) error
	GetBindingToken(ctx context.Context, userID int64) (*domain.BindingToken, error)
	GetUserByBindingToken(ctx context.Context, token string) (*domain.User, error)
	SetTelegramID(ctx context.Context, userID, telegramID int64) error
}

type telegramRepository struct {
	pool *pgxpool.Pool
}

func NewTelegramRepository(pool *pgxpool.Pool) TelegramRepository {
	return &telegramRepository{pool: pool}
}

func (r *telegramRepository) SetTelegramID(ctx context.Context, userID, telegramID int64) error {
	query := `
		UPDATE users
		SET telegram_id = $1,
			updated_at = NOW()
		WHERE id = $2;
	`

	if _, err := r.pool.Exec(ctx, query, telegramID, userID); err != nil {
		if isDuplicateError(err) {
			return fmt.Errorf("set telegram id: %w", domain.ErrAlreadyExists)
		}
		return fmt.Errorf("set telegram id: %w", err)
	}

	return nil
}

func (r *telegramRepository) GetUserByBindingToken(ctx context.Context, token string) (*domain.User, error) {
	query := `
		SELECT 
			u.id,
			u.email,
			u.username,
			u.created_at,
			u.updated_at,
			u.telegram_id,
			t.expires_at
		FROM telegram_binding_tokens t
			JOIN users u on t.user_id = u.id
		WHERE token = $1;
	`

	var user domain.User
	var expiresAt time.Time
	err := r.pool.QueryRow(ctx, query, token).Scan(
		&user.ID,
		&user.Email,
		&user.Username,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.TelegramId,
		&expiresAt,
	)
	if err != nil {
		return nil, fmt.Errorf("get user by binding token: %w", err)
	}

	if time.Now().After(expiresAt) {
		return nil, fmt.Errorf("get user by binding token: %w", domain.ErrBindingTokenExpired)
	}

	return &user, nil
}

func (r *telegramRepository) CreateBindingToken(ctx context.Context, token *domain.BindingToken) error {
	query := `
		INSERT INTO telegram_binding_tokens(
			user_id, 
			token, 
			expires_at
		)
		VALUES($1, $2, $3)
		RETURNING id, created_at;
	`

	err := r.pool.QueryRow(
		ctx, query,
		token.UserID,
		token.Token,
		token.ExpiresAt,
	).Scan(
		&token.ID,
		&token.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("create binding token: %w", err)
	}

	return nil
}

func (r *telegramRepository) GetBindingToken(ctx context.Context, userID int64) (*domain.BindingToken, error) {
	query := `
		SELECT
			t.id,
			t.user_id,
			t.token,
			t.created_at,
			t.expires_at
		FROM telegram_binding_tokens t
		WHERE user_id = $1
		  AND expires_at > NOW();
	`

	var bindingToken domain.BindingToken
	err := r.pool.QueryRow(ctx, query, userID).Scan(
		&bindingToken.ID,
		&bindingToken.UserID,
		&bindingToken.Token,
		&bindingToken.CreatedAt,
		&bindingToken.ExpiresAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("get binding token: %w", domain.ErrNotFound)
		} else {
			return nil, fmt.Errorf("get binding token: %w", err)
		}
	}

	return &bindingToken, nil
}
