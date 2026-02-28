package domain

import (
	"time"

	"github.com/guregu/null"
)

type Session struct {
	ID             string      `json:"id"              db:"id"`
	UserID         string      `json:"user_id"         db:"user_id"`
	IPAddress      null.String `json:"ip_address"      db:"ip_address"`
	UserAgent      null.String `json:"user_agent"      db:"user_agent"`
	IsRevoked      bool        `json:"is_revoked"      db:"is_revoked"`
	AbsoluteExpiry time.Time   `json:"absolute_expiry" db:"absolute_expiry"`
	CreatedAt      time.Time   `json:"created_at"      db:"created_at"`
	LastSeenAt     time.Time   `json:"last_seen_at"    db:"last_seen_at"`
}
