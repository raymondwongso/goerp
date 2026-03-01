package postgres

import (
	"context"

	"github.com/jmoiron/sqlx"
	"github.com/raymondwongso/goerp/domain"
	"go.opentelemetry.io/otel/trace"
)

type userWriter struct {
	tracer trace.Tracer
	db     sqlx.QueryerContext
}

// NewUserWriter creates a new userWriter
func NewUserWriter(tracer trace.Tracer, db sqlx.QueryerContext) *userWriter {
	return &userWriter{
		tracer: tracer,
		db:     db,
	}
}

// Upsert inserts or updates a user by email.
// On conflict it updates display_name, avatar_url and updated_at.
func (w *userWriter) Upsert(ctx context.Context, user domain.User) (domain.User, error) {
	ctx, span := w.tracer.Start(ctx, "UserWriter/Upsert")
	defer span.End()

	query := `
		INSERT INTO users (email, display_name, avatar_url)
		VALUES ($1, $2, $3)
		ON CONFLICT (email) DO UPDATE
			SET display_name = EXCLUDED.display_name,
			    avatar_url   = EXCLUDED.avatar_url,
			    updated_at   = now()
		RETURNING id, email, display_name, avatar_url, is_active, created_at, updated_at`

	row := w.db.QueryRowxContext(ctx, query, user.Email, user.DisplayName, user.AvatarURL)
	if err := row.Err(); err != nil {
		return domain.User{}, err
	}

	var res domain.User
	if err := row.StructScan(&res); err != nil {
		return domain.User{}, err
	}
	return res, nil
}
