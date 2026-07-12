package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/xromen/movietracker/internal/domain"
)

type UserRepository interface {
	Create(ctx context.Context, user *domain.User) error
	GetByEmail(ctx context.Context, email string) (*domain.User, error)
	GetByUsername(ctx context.Context, username string) (*domain.User, error)
	GetByID(ctx context.Context, id int64) (*domain.User, error)
	AddUserToRole(ctx context.Context, userID, roleID int64) error
	GetAuthVersion(ctx context.Context, id int64) (*int64, error)
}

type userRepository struct {
	pool *pgxpool.Pool
}

func NewUserRepository(pool *pgxpool.Pool) UserRepository {
	return &userRepository{pool: pool}
}

func (r *userRepository) Create(ctx context.Context, user *domain.User) error {
	query := `
		INSERT INTO users (email, username, password_hash)
		VALUES ($1, $2, $3)
		RETURNING id, created_at, updated_at, auth_version
	`

	err := r.pool.QueryRow(ctx, query,
		user.Email,
		user.Username,
		user.PasswordHash,
	).Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt, &user.AuthVersion)

	if err != nil {
		if isDuplicateError(err) {
			return fmt.Errorf("create user: %w", domain.ErrAlreadyExists)
		}
		return fmt.Errorf("create user: %w", err)
	}

	return nil
}

func (r *userRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	query := `
        SELECT
			u.id,
			u.email,
			u.username,
			password_hash,
			u.created_at,
			u.updated_at,
			u.auth_version,
			COALESCE(
				ARRAY_AGG(r.name) FILTER (WHERE r.name IS NOT NULL),
				'{}'::text[]
			) AS roles
		FROM users u
				LEFT JOIN user_roles ur ON ur.user_id = u.id
				LEFT JOIN roles r ON r.id = ur.role_id
		WHERE u.email = $1
		GROUP BY u.id,
				u.email,
				u.username,
				u.created_at,
				u.updated_at,
				u.auth_version;
    `

	user := &domain.User{}
	err := r.pool.QueryRow(ctx, query, email).Scan(
		&user.ID,
		&user.Email,
		&user.Username,
		&user.PasswordHash,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.AuthVersion,
		&user.Roles,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("get user by email: %w", domain.ErrNotFound)
		}
		return nil, fmt.Errorf("get user by email: %w", err)
	}

	return user, nil
}

func (r *userRepository) GetByUsername(ctx context.Context, username string) (*domain.User, error) {
	query := `
		SELECT
			u.id,
			u.email,
			u.username,
			password_hash,
			u.created_at,
			u.updated_at,
			u.auth_version,
			COALESCE(
				ARRAY_AGG(r.name) FILTER (WHERE r.name IS NOT NULL),
				'{}'::text[]
			) AS roles
		FROM users u
				LEFT JOIN user_roles ur ON ur.user_id = u.id
				LEFT JOIN roles r ON r.id = ur.role_id
		WHERE u.username = $1
		GROUP BY u.id,
				u.email,
				u.username,
				u.created_at,
				u.updated_at,
				u.auth_version;
    `

	user := &domain.User{}
	err := r.pool.QueryRow(ctx, query, username).Scan(
		&user.ID,
		&user.Email,
		&user.Username,
		&user.PasswordHash,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.AuthVersion,
		&user.Roles,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("get user by username: %w", domain.ErrNotFound)
		}
		return nil, fmt.Errorf("get user by username: %w", err)
	}

	return user, nil
}

func (r *userRepository) GetByID(ctx context.Context, id int64) (*domain.User, error) {
	query := `
		SELECT
			u.id,
			u.email,
			u.username,
			password_hash,
			u.created_at,
			u.updated_at,
			u.auth_version,
			COALESCE(
				ARRAY_AGG(r.name) FILTER (WHERE r.name IS NOT NULL),
				'{}'::text[]
			) AS roles
		FROM users u
				LEFT JOIN user_roles ur ON ur.user_id = u.id
				LEFT JOIN roles r ON r.id = ur.role_id
		WHERE u.id = $1
		GROUP BY u.id,
				u.email,
				u.username,
				u.created_at,
				u.updated_at,
				u.auth_version;
    `

	user := &domain.User{}
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&user.ID,
		&user.Email,
		&user.Username,
		&user.PasswordHash,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.AuthVersion,
		&user.Roles,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("get user by id: %w", domain.ErrNotFound)
		}
		return nil, fmt.Errorf("get user by id: %w", err)
	}

	return user, nil
}

func (r *userRepository) AddUserToRole(ctx context.Context, userID, roleID int64) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin add user to role: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	
	query := `
		INSERT INTO user_roles VALUES ($1, $2)
		ON CONFLICT DO NOTHING;

		UPDATE users
		SET auth_version = auth_version + 1;
	`

	if _, err := tx.Exec(ctx, query, userID, roleID); err != nil {
		return fmt.Errorf("add user to role: %w", err)
	}

	err = tx.Commit(ctx)
	if err != nil {
		return fmt.Errorf("commit add user role: %w", err)
	}

	return nil
}

func (r *userRepository) GetAuthVersion(ctx context.Context, id int64) (*int64, error) {
	query := `
		SELECT auth_version
		FROM users u
		WHERE u.id = $1;
	`

	var authVersion int64
	err := r.pool.QueryRow(ctx, query, id).Scan(&authVersion)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("get user auth version: %w", domain.ErrNotFound)
		}
		return nil, fmt.Errorf("get user auth version: %w", err)
	}

	return &authVersion, nil
}

// 23505 — стандартный SQLSTATE код для unique_violation.
func isDuplicateError(err error) bool {
	// pgx оборачивает PostgreSQL ошибки в pgconn.PgError.
	// errors.As разворачивает цепочку ошибок и ищет нужный тип.
	var pgErr interface{ SQLState() string }
	if errors.As(err, &pgErr) {
		return pgErr.SQLState() == "23505"
	}
	return false
}
