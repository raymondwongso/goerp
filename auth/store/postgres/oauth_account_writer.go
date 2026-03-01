package postgres

import (
	"context"

	"github.com/jmoiron/sqlx"
	"github.com/raymondwongso/goerp/domain"
	"go.opentelemetry.io/otel/trace"
)

type oauthAccountWriter struct {
	tracer trace.Tracer
	db     sqlx.QueryerContext
}

// NewOAuthAccountWriter creates a new oauthAccountWriter
func NewOAuthAccountWriter(tracer trace.Tracer, db sqlx.QueryerContext) *oauthAccountWriter {
	return &oauthAccountWriter{
		tracer: tracer,
		db:     db,
	}
}

// Upsert inserts or updates an oauth account by (provider, provider_sub).
// On conflict it updates email and last_login.
func (w *oauthAccountWriter) Upsert(ctx context.Context, account domain.OAuthAccount) (domain.OAuthAccount, error) {
	ctx, span := w.tracer.Start(ctx, "OAuthAccountWriter/Upsert")
	defer span.End()

	query := `
		INSERT INTO oauth_accounts (user_id, provider, provider_sub, email)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (provider, provider_sub) DO UPDATE
			SET email      = EXCLUDED.email,
			    last_login = now()
		RETURNING id, user_id, provider, provider_sub, email, last_login, created_at`

	row := w.db.QueryRowxContext(ctx, query, account.UserID, account.Provider, account.ProviderSub, account.Email)
	if err := row.Err(); err != nil {
		return domain.OAuthAccount{}, err
	}

	var res domain.OAuthAccount
	if err := row.StructScan(&res); err != nil {
		return domain.OAuthAccount{}, err
	}
	return res, nil
}
