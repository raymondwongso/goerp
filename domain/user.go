package domain

import (
	"context"
	"time"

	"github.com/guregu/null"
)

//go:generate mockgen -package=mockdomain -source=$GOFILE -destination=mock/mock_$GOFILE

// UserWriter defines interface for User repository write operations
type UserWriter interface {
	Upsert(ctx context.Context, user User) (User, error)
}

type User struct {
	ID          string      `json:"id"           db:"id"`
	Email       string      `json:"email"        db:"email"`
	DisplayName null.String `json:"display_name" db:"display_name"`
	AvatarURL   null.String `json:"avatar_url"   db:"avatar_url"`
	IsActive    bool        `json:"is_active"    db:"is_active"`
	CreatedAt   time.Time   `json:"created_at"   db:"created_at"`
	UpdatedAt   time.Time   `json:"updated_at"   db:"updated_at"`
}
