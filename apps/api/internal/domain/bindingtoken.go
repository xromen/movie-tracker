package domain

import "time"

type BindingToken struct {
	ID        int64
	UserID    int64
	Token     string
	CreatedAt time.Time
	ExpiresAt time.Time
}
