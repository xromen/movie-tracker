package repository

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/xromen/movietracker/internal/domain"
	"github.com/xromen/movietracker/internal/platform/database"
)

var testPool *pgxpool.Pool

func TestMain(m *testing.M) {
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		dsn = "postgres://postgres:maxim11@localhost:5432/movietracker?sslmode=disable"
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var err error
	testPool, err = database.NewPool(ctx, database.Config{
		DSN:             dsn,
		MaxConns:        5,
		MinConns:        1,
		MaxConnLifetime: time.Minute,
		MaxConnIdleTime: time.Minute,
	})

	if err != nil {
		panic("failed to connect to test database: " + err.Error())
	}
	defer testPool.Close()

	os.Exit(m.Run())
}

func cleanupUsers(t *testing.T, emails ...string) {
	t.Helper()

	for _, email := range emails {
		_, err := testPool.Exec(
			context.Background(),
			"DELETE FROM users WHERE email = $1",
			email,
		)

		if err != nil {
			t.Logf("clenup warning: failed to delete user '%s': %v", email, err)
		}
	}
}

func TestUserRepository_Create(t *testing.T) {
	repo := NewUserRepository(testPool)

	tests := []struct {
		name    string
		user    *domain.User
		wantErr error // nil = ожидаем успех, иначе ожидаем эту ошибку
	}{
		{
			name: "success",
			user: &domain.User{
				Email:        "test_create@example.com",
				Username:     "testuser_create",
				PasswordHash: "hashed_password_here",
			},
			wantErr: nil,
		},
		{
			name: "duplicate email",
			user: &domain.User{
				Email:        "test_create@example.com",
				Username:     "different_username",
				PasswordHash: "hashed_password_here",
			},
			wantErr: domain.ErrAlreadyExists,
		},
		{
			name: "duplicate username",
			user: &domain.User{
				Email:        "another@example.com",
				Username:     "testuser_create",
				PasswordHash: "hashed_password_here",
			},
			wantErr: domain.ErrAlreadyExists,
		},
	}

	t.Cleanup(func() {
		cleanupUsers(t,
			"test_create@example.com",
			"another@example.com")
	})

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			err := repo.Create(ctx, tt.user)

			if tt.wantErr != nil {
				if err == nil {
					t.Fatal("expected error, got nil")
				}

				if !errors.Is(err, tt.wantErr) {
					t.Errorf("expected error %v, got %v", tt.wantErr, err)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if tt.user.ID == 0 {
				t.Error("expected ID to be set after create, got 0")
			}
			if tt.user.CreatedAt.IsZero() {
				t.Error("expected CreatedAt to be set after create")
			}
		})
	}
}

func TestUserRepository_GetByEmail(t *testing.T) {
	repo := NewUserRepository(testPool)
	ctx := context.Background()

	seed := &domain.User{
		Email:        "test_get@example.com",
		Username:     "testuser_get",
		PasswordHash: "some_hash",
	}
	if err := repo.Create(ctx, seed); err != nil {
		t.Fatalf("seed: failed to create user: %v", err)
	}
	t.Cleanup(func() { cleanupUsers(t, seed.Email) })

	tests := []struct {
		name    string
		email   string
		wantErr error
		check   func(t *testing.T, user *domain.User)
	}{
		{
			name:    "found",
			email:   seed.Email,
			wantErr: nil,
			check: func(t *testing.T, user *domain.User) {
				t.Helper()
				if user.Email != seed.Email {
					t.Errorf("email: got %q, want %q", user.Email, seed.Email)
				}
				if user.Username != seed.Username {
					t.Errorf("username: got %q, want %q", user.Username, seed.Username)
				}
				if user.PasswordHash == "" {
					t.Error("expected PasswordHash to be non-empty")
				}
			},
		},
		{
			name:    "not found",
			email:   "nonexistent@example.com",
			wantErr: domain.ErrNotFound,
			check:   nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user, err := repo.GetByEmail(ctx, tt.email)

			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("got error %v, want %v", err, tt.wantErr)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tt.check != nil {
				tt.check(t, user)
			}
		})
	}
}
