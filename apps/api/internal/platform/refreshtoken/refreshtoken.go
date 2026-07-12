package refreshtoken

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/xromen/movietracker/internal/domain"
)

const tokenBytes = 128

type Manager interface {
	Generate(userID *int64) (string, *domain.RefreshToken, error)
}

type Config struct {
	RefreshTokenTTL time.Duration
}

func NewManager(cfg Config) Manager {
	return &manager{
		cfg: cfg,
	}
}

type manager struct {
	cfg Config
}

func (m *manager) Generate(userID *int64) (string, *domain.RefreshToken, error) {
	bytes := make([]byte, tokenBytes)

	if _, err := rand.Read(bytes); err != nil {
		return "", nil, fmt.Errorf("generate refresh token: %w", err)
	}

	raw := base64.RawURLEncoding.EncodeToString(bytes)

	now := time.Now()
	token := domain.RefreshToken{
		TokenHash: Hash(raw),
		FamilyID:  uuid.NewString(),
		ExpiresAt: now.Add(m.cfg.RefreshTokenTTL),
		CreatedAt: now,
	}

	if userID != nil {
		token.UserID = *userID
	}

	return raw, &token, nil
}

func Hash(raw string) []byte {
	sum := sha256.Sum256([]byte(raw))
	return sum[:]
}
