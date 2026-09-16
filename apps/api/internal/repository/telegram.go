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

	GetDueReportMessages(ctx context.Context, limit int) ([]domain.DueReportMessage, error)
	MarkMessageSuccess(ctx context.Context, messageID, telegramMessageID int64) error
	MarkMessageFailure(ctx context.Context, messageID int64, nextRetry time.Time, error string) error
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

func (r *telegramRepository) GetDueReportMessages(ctx context.Context, limit int) ([]domain.DueReportMessage, error) {
	query := `
		SELECT
			rm.id,
			u.telegram_id,
			rm.message,
			rm.position
		FROM report_messages rm
				JOIN reports r ON rm.report_id = r.id
				JOIN users u ON r.user_id = u.id
		WHERE rm.sent_at IS NULL
		AND rm.next_attempt_at <= NOW()
		ORDER BY u.id, rm.position
		LIMIT $1;
	`

	var result []domain.DueReportMessage

	rows, err := r.pool.Query(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("get due report messages: %w")
	}
	defer rows.Close()

	for rows.Next() {
		var message domain.DueReportMessage
		err := rows.Scan(
			&message.ID,
			&message.UserTelegramID,
			&message.Message,
			&message.Position,
		)
		if err != nil {
			return nil, fmt.Errorf("scan report message: %w", err)
		}

		result = append(result, message)
	}

	return result, rows.Err()
}

func (r *telegramRepository) MarkMessageSuccess(ctx context.Context, messageID, telegramMessageID int64) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE report_messages
		SET sent_at = NOW(),
			telegram_message_id = $2
		WHERE id = $1;
	`, messageID, telegramMessageID)

	if err != nil {
		return fmt.Errorf("mark message success: %w", err)
	}

	return nil
}

func (r *telegramRepository) MarkMessageFailure(ctx context.Context, messageID int64, nextRetry time.Time, error string) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE report_messages
		SET attempts        = attempts + 1,
			next_attempt_at = $2,
			last_error      = $3
		WHERE id = $1;
	`,
		messageID,
		nextRetry,
		error,
	)

	if err != nil {
		return fmt.Errorf("mark message failure: %w", err)
	}

	return nil
}
