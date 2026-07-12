package service_test

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/xromen/movietracker/internal/domain"
	"github.com/xromen/movietracker/internal/platform/jwt"
)

type inMemoryUserRepo struct {
	users map[string]*domain.User
}

func newInMemoryUserRepo() *inMemoryUserRepo {
	return &inMemoryUserRepo{users: make(map[string]*domain.User)}
}

func (s *inMemoryUserRepo) Create(_ context.Context, user *domain.User) error {
	if _, exists := s.users[user.Email]; exists {
		return fmt.Errorf("create: %w", domain.ErrAlreadyExists)
	}
	user.ID = int64(len(s.users) + 1)
	s.users[user.Email] = user
	return nil
}

func (s *inMemoryUserRepo) GetByEmail(_ context.Context, email string) (*domain.User, error) {
	if user, exists := s.users[email]; exists {
		return user, nil
	}
	return nil, fmt.Errorf("get: %w", domain.ErrNotFound)
}

func (s *inMemoryUserRepo) GetByUsername(_ context.Context, username string) (*domain.User, error) {
	for _, user := range s.users {
		if user.Username == username {
			return user, nil
		}
	}
	return nil, fmt.Errorf("get: %w", domain.ErrNotFound)
}

func (s *inMemoryUserRepo) GetByID(_ context.Context, id int64) (*domain.User, error) {
	for _, u := range s.users {
		if u.ID == id {
			return u, nil
		}
	}
	return nil, fmt.Errorf("get: %w", domain.ErrNotFound)
}

func (s *inMemoryUserRepo) AddUserToRole(_ context.Context, id, roleID int64) error {
	for i, u := range s.users {
		if u.ID == id {
			s.users[i].Roles = append(s.users[i].Roles, fmt.Sprintf("%d", roleID))
			return nil
		}
	}

	return fmt.Errorf("get: %w", domain.ErrNotFound)
}

func (s *inMemoryUserRepo) GetAuthVersion(ctx context.Context, userID int64) (*int64, error) {
	user, _ := s.GetByID(ctx, userID)

	return &user.AuthVersion, nil
}

type fakeHasher struct{}

func (fakeHasher) Hash(p string) (string, error) {
	return "hashed:" + p, nil
}

func (fakeHasher) Verify(p, h string) error {
	if "hashed:"+p != h {
		return errors.New("mismatch")
	}
	return nil
}

// fakeJWT — возвращает предсказуемые токены.
type fakeJWT struct{}

func (fakeJWT) Generate(id int64, username string, authVersion int64, roles []string) (string, error) {
	return fmt.Sprintf("token:%d:%s:%d:%s", id, username, authVersion, strings.Join(roles, ":")), nil
}
func (fakeJWT) Validate(token string) (*jwt.Claims, error) {
	return nil, nil
}

// func TestUserService_Register(t *testing.T) {
// 	tests := []struct {
// 		name    string
// 		input   service.RegisterInput
// 		wantErr error
// 	}{
// 		{
// 			name:    "success",
// 			input:   service.RegisterInput{Email: "a@b.com", Username: "alice", Password: "password123"},
// 			wantErr: nil,
// 		},
// 		{
// 			name:    "duplicate",
// 			input:   service.RegisterInput{Email: "a@b.com", Username: "alice2", Password: "password123"},
// 			wantErr: domain.ErrAlreadyExists,
// 		},
// 		{
// 			name:    "short password",
// 			input:   service.RegisterInput{Email: "b@b.com", Username: "bob", Password: "pw"},
// 			wantErr: domain.ErrInvalidInput,
// 		},
// 	}

// 	repo := newInMemoryUserRepo()
// 	svc := service.NewUserService(repo, fakeHasher{}, fakeJWT{}, nil)

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			out, err := svc.Register(context.Background(), tt.input)

// 			if tt.wantErr != nil {
// 				if !errors.Is(err, tt.wantErr) {
// 					t.Errorf("got %v, want %v", err, tt.wantErr)
// 				}
// 				return
// 			}
// 			if err != nil {
// 				t.Fatalf("unexpected error: %v", err)
// 			}
// 			if out.Token == "" {
// 				t.Error("expected non-empty token")
// 			}
// 			if out.User.ID == 0 {
// 				t.Error("expected user ID to be set")
// 			}
// 		})
// 	}
// }
