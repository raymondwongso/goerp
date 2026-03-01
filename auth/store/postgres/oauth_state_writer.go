package postgres

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jmoiron/sqlx"
	"github.com/raymondwongso/goerp/domain"
	"github.com/raymondwongso/goerp/domain/xerror"
	"go.opentelemetry.io/otel/trace"
)

type oauthStateWriter struct {
	tracer trace.Tracer
	db     sqlx.QueryerContext
}

// NewOAuthStateWriter creates a new oauthStateWriter
func NewOAuthStateWriter(tracer trace.Tracer, db sqlx.QueryerContext) *oauthStateWriter {
	return &oauthStateWriter{
		tracer: tracer,
		db:     db,
	}
}

func (w *oauthStateWriter) Insert(ctx context.Context, state domain.OAuthState) (domain.OAuthState, error) {
	ctx, span := w.tracer.Start(ctx, "OAuthStateWriter/Insert")
	defer span.End()

	query := `
		INSERT INTO oauth_states (state, code_verifier, redirect_to)
		VALUES ($1, $2, $3)
		RETURNING state, code_verifier, redirect_to, created_at, expires_at`

	row := w.db.QueryRowxContext(ctx, query, state.State, state.CodeVerifier, state.RedirectTo)
	if err := row.Err(); err != nil {
		return domain.OAuthState{}, err
	}

	var res domain.OAuthState
	if err := row.StructScan(&res); err != nil {
		return domain.OAuthState{}, err
	}
	return res, nil
}

// DeleteByState atomically deletes and returns the oauth state if it exists and has not expired.
// Returns xerror.CodeNotFound if the state is not found or has expired.
func (w *oauthStateWriter) DeleteByState(ctx context.Context, state string) (domain.OAuthState, error) {
	ctx, span := w.tracer.Start(ctx, "OAuthStateWriter/DeleteByState")
	defer span.End()

	query := `
		DELETE FROM oauth_states
		WHERE state = $1 AND expires_at > now()
		RETURNING state, code_verifier, redirect_to, created_at, expires_at`

	row := w.db.QueryRowxContext(ctx, query, state)
	if err := row.Err(); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.OAuthState{}, xerror.New(xerror.CodeNotFound, "oauth state not found or expired")
		}
		return domain.OAuthState{}, err
	}

	var res domain.OAuthState
	if err := row.StructScan(&res); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.OAuthState{}, xerror.New(xerror.CodeNotFound, "oauth state not found or expired")
		}
		return domain.OAuthState{}, err
	}
	return res, nil
}
