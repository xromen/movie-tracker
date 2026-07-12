package jwt

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	UserID      int64    `json:"user_id"`
	Username    string   `json:"username"`
	Roles       []string `json:"roles"`
	AuthVersion int64    `json:"av"`
	jwt.RegisteredClaims
}

type Manager interface {
	Generate(userID int64, username string, authVersion int64, roles []string) (string, time.Time, error)
	Validate(token string) (*Claims, error)
}

type Config struct {
	Secret         string
	AccessTokenTTL time.Duration
}

type manager struct {
	cfg Config
}

func NewManager(cfg Config) Manager {
	return &manager{cfg: cfg}
}

func (m *manager) Generate(userID int64, username string, authVersion int64, roles []string) (string, time.Time, error) {
	now := time.Now()

	expiresAt := now.Add(m.cfg.AccessTokenTTL)

	claims := Claims{
		UserID:      userID,
		Username:    username,
		Roles:       roles,
		AuthVersion: authVersion,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   fmt.Sprintf("%d", userID),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	signed, err := token.SignedString([]byte(m.cfg.Secret))

	if err != nil {
		return "", now, fmt.Errorf("jwt signing failed: %w", err)
	}

	return signed, expiresAt, nil
}

func (m *manager) Validate(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(
		tokenString,
		&Claims{},
		func(token *jwt.Token) (any, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(m.cfg.Secret), nil
		},
	)
	if err != nil {
		return nil, fmt.Errorf("jwt parse failed: %w", err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	return claims, nil
}
