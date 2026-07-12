package domain

import "time"

type RefreshToken struct {
	ID         string
	UserID     int64
	TokenHash  []byte
	FamilyID   string
	ExpiresAt  time.Time
	CreatedAt  time.Time
	RevokedAt  *time.Time
	ReplacedBy *string
}