package domain

import (
	"context"
	"time"

	"github.com/guregu/null"
)

//go:generate mockgen -package=mockdomain -source=$GOFILE -destination=mock/mock_$GOFILE

// OAuthStateWriter defines interface for OAuthState repository write operations
type OAuthStateWriter interface {
	Insert(ctx context.Context, state OAuthState) (OAuthState, error)
	DeleteByState(ctx context.Context, state string) (OAuthState, error)
}

// OAuthAccountWriter defines interface for OAuthAccount repository write operations
type OAuthAccountWriter interface {
	Upsert(ctx context.Context, account OAuthAccount) (OAuthAccount, error)
}

type OAuthProvider string

const (
	OAuthProviderGoogle OAuthProvider = "google"
	OAuthProviderGithub OAuthProvider = "github"
)

type OAuthAccount struct {
	ID          string        `json:"id"           db:"id"`
	UserID      string        `json:"user_id"      db:"user_id"`
	Provider    OAuthProvider `json:"provider"     db:"provider"`
	ProviderSub string        `json:"provider_sub" db:"provider_sub"`
	Email       string        `json:"email"        db:"email"`
	LastLogin   time.Time     `json:"last_login"   db:"last_login"`
	CreatedAt   time.Time     `json:"created_at"   db:"created_at"`
}

type OAuthState struct {
	State        string      `json:"state"         db:"state"`
	CodeVerifier string      `json:"code_verifier" db:"code_verifier"`
	RedirectTo   null.String `json:"redirect_to"   db:"redirect_to"`
	IPAddress    null.String `json:"ip_address"    db:"ip_address"`
	CreatedAt    time.Time   `json:"created_at"    db:"created_at"`
	ExpiresAt    time.Time   `json:"expires_at"    db:"expires_at"`
}
