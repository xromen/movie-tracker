package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/xromen/movietracker/internal/domain"
	"github.com/xromen/movietracker/internal/platform/hasher"
	"github.com/xromen/movietracker/internal/platform/jwt"
	"github.com/xromen/movietracker/internal/platform/refreshtoken"
)

type userRepository interface {
	Create(ctx context.Context, user *domain.User) error
	GetByEmail(ctx context.Context, email string) (*domain.User, error)
	GetByUsername(ctx context.Context, username string) (*domain.User, error)
	GetByID(ctx context.Context, id int64) (*domain.User, error)
	AddUserToRole(ctx context.Context, userID, roleID int64) error
	GetAuthVersion(ctx context.Context, id int64) (*int64, error)
}

type refreshTokenRepository interface {
	Create(ctx context.Context, token *domain.RefreshToken) error
	Rotate(ctx context.Context, oldTokenHash []byte, newToken *domain.RefreshToken) error
	RevokeByUserID(ctx context.Context, userID int64) error
}

type RegisterInput struct {
	Email    string
	Username string
	Password string
}

type AuthOutput struct {
	AccessToken           string
	AccessTokenExpiresAt  time.Time
	RefreshToken          string
	RefreshTokenExpiresAt time.Time
}

type UserService interface {
	Register(ctx context.Context, input RegisterInput) (*AuthOutput, error)
	Login(ctx context.Context, email, password string) (*AuthOutput, error)
	Refresh(ctx context.Context, refreshToken string) (*AuthOutput, error)
	GetByID(ctx context.Context, id int64) (*domain.User, error)
	AddUserToRole(ctx context.Context, id, roleID int64) error
	ValidateAuthVersion(ctx context.Context, userID, authVersion int64) (bool, error)
}

type userService struct {
	userRepo         userRepository
	refreshTokenRepo refreshTokenRepository
	hasher           hasher.Hasher
	jwt              jwt.Manager
	refreshToken     refreshtoken.Manager
	logger           *slog.Logger
}

func NewUserService(
	repo userRepository,
	refreshTokenRepo refreshTokenRepository,
	hasher hasher.Hasher,
	jwt jwt.Manager,
	refreshToken refreshtoken.Manager,
	logger *slog.Logger,
) UserService {
	return &userService{
		userRepo:         repo,
		refreshTokenRepo: refreshTokenRepo,
		hasher:           hasher,
		jwt:              jwt,
		refreshToken:     refreshToken,
		logger:           logger,
	}
}

func (s *userService) Register(ctx context.Context, input RegisterInput) (*AuthOutput, error) {
	err := validateRegisterInput(input)
	if err != nil {
		return nil, err
	}

	hash, err := s.hasher.Hash(input.Password)
	if err != nil {
		return nil, fmt.Errorf("register: %w", err)
	}

	user := &domain.User{
		Username:     input.Username,
		Email:        input.Email,
		PasswordHash: hash,
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		if errors.Is(err, domain.ErrAlreadyExists) {
			return nil, fmt.Errorf("register: user already exists: %w", domain.ErrAlreadyExists)
		}
		return nil, fmt.Errorf("register: %w", err)
	}

	accessToken, accessTokenExpiresAt, err := s.jwt.Generate(user.ID, user.Username, user.AuthVersion, nil)
	if err != nil {
		return nil, fmt.Errorf("register: %w", err)
	}

	refreshTokenRaw, refreshToken, err := s.createRefreshToken(ctx, user.ID)
	if err != nil {
		return nil, fmt.Errorf("generate refresh token: %w", err)
	}

	return &AuthOutput{
		AccessToken:           accessToken,
		AccessTokenExpiresAt:  accessTokenExpiresAt,
		RefreshToken:          refreshTokenRaw,
		RefreshTokenExpiresAt: refreshToken.ExpiresAt,
	}, nil
}

func (u *userService) Login(ctx context.Context, email, password string) (*AuthOutput, error) {
	user, err := u.userRepo.GetByEmail(ctx, email)
	if err != nil {
		if !errors.Is(err, domain.ErrNotFound) {
			return nil, fmt.Errorf("login: %w", err)
		}

		user, err = u.userRepo.GetByUsername(ctx, email)
		if err != nil {
			if errors.Is(err, domain.ErrNotFound) {
				return nil, fmt.Errorf("login: %w", domain.ErrUnauthorized)
			}
			return nil, fmt.Errorf("login: %w", err)
		}
	}

	if err := u.hasher.Verify(password, user.PasswordHash); err != nil {
		return nil, fmt.Errorf("login: %w", domain.ErrUnauthorized)
	}

	accessToken, accessTokenExpiresAt, err := u.jwt.Generate(user.ID, user.Username, user.AuthVersion, user.Roles)
	if err != nil {
		return nil, fmt.Errorf("login: %w", err)
	}

	refreshTokenRaw, refreshToken, err := u.createRefreshToken(ctx, user.ID)
	if err != nil {
		return nil, fmt.Errorf("generate refresh token: %w", err)
	}

	return &AuthOutput{
		AccessToken:           accessToken,
		AccessTokenExpiresAt:  accessTokenExpiresAt,
		RefreshToken:          refreshTokenRaw,
		RefreshTokenExpiresAt: refreshToken.ExpiresAt,
	}, nil
}

func (u *userService) Refresh(ctx context.Context, refreshToken string) (*AuthOutput, error) {
	rawNewRefreshToken, newRefreshToken, err := u.refreshToken.Generate(nil)
	if err != nil {
		return nil, fmt.Errorf("generete refresh token: %w", err)
	}

	err = u.refreshTokenRepo.Rotate(ctx, refreshtoken.Hash(refreshToken), newRefreshToken)
	if err != nil {
		return nil, fmt.Errorf("rotate refresh token: %w", err)
	}

	user, err := u.userRepo.GetByID(ctx, newRefreshToken.UserID)
	if err != nil {
		return nil, fmt.Errorf("get user by id for refresh token: %w", err)
	}

	accessToken, accessTokenExpiresAt, err := u.jwt.Generate(user.ID, user.Username, user.AuthVersion, user.Roles)
	if err != nil {
		return nil, fmt.Errorf("genrate access token for refresh token: %w", err)
	}

	return &AuthOutput{
		AccessToken:           accessToken,
		AccessTokenExpiresAt:  accessTokenExpiresAt,
		RefreshToken:          rawNewRefreshToken,
		RefreshTokenExpiresAt: newRefreshToken.ExpiresAt,
	}, nil
}

func (u *userService) GetByID(ctx context.Context, id int64) (*domain.User, error) {
	user, err := u.userRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get user: %w", err)
	}
	return user, nil
}

func (u *userService) AddUserToRole(ctx context.Context, id, roleID int64) error {
	err := u.userRepo.AddUserToRole(ctx, id, roleID)
	if err != nil {
		return fmt.Errorf("add user to role: %w", err)
	}

	return nil
}

func (u *userService) ValidateAuthVersion(ctx context.Context, userID, authVersion int64) (bool, error) {
	actualAuthVersion, err := u.userRepo.GetAuthVersion(ctx, userID)
	if err != nil {
		u.logger.Error("get user auth version",
			"userID", userID,
			"error", err)
		return false, fmt.Errorf("get user auth version: %w", err)
	}

	return *actualAuthVersion == authVersion, nil
}

func (u *userService) createRefreshToken(ctx context.Context, userID int64) (string, *domain.RefreshToken, error) {
	raw, token, err := u.refreshToken.Generate(&userID)
	if err != nil {
		return "", nil, fmt.Errorf("generate refresh token: %w", err)
	}

	err = u.refreshTokenRepo.Create(ctx, token)
	if err != nil {
		return "", nil, fmt.Errorf("create refsresh token: %w", err)
	}

	return raw, token, nil
}

func validateRegisterInput(input RegisterInput) error {
	if input.Email == "" {
		return domain.NewValidationError("email", "is required")
	}
	if len(input.Email) > 255 {
		return domain.NewValidationError("email", "is too long")
	}
	if input.Username == "" {
		return domain.NewValidationError("username", "is required")
	}
	if len(input.Username) < 3 {
		return domain.NewValidationError("username", "must be at least 3 characters")
	}
	if len(input.Password) < 3 {
		return domain.NewValidationError("password", "must be at least 3 characters")
	}
	return nil
}
