package domain

import (
	"time"

	"github.com/guregu/null"
)

type UserRole struct {
	UserID    string      `json:"user_id"    db:"user_id"`
	RoleID    string      `json:"role_id"    db:"role_id"`
	GrantedBy null.String `json:"granted_by" db:"granted_by"`
	GrantedAt time.Time   `json:"granted_at" db:"granted_at"`
}
