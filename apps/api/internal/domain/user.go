package domain

import "time"

type User struct {
	ID           int64
	Email        string
	Username     string
	Roles        []string
	PasswordHash string
	AuthVersion  int64
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
