package domain

import (
	"time"

	"github.com/guregu/null"
)

type Role struct {
	ID          string      `json:"id"          db:"id"`
	Name        string      `json:"name"        db:"name"`
	Description null.String `json:"description" db:"description"`
	CreatedAt   time.Time   `json:"created_at"  db:"created_at"`
}
