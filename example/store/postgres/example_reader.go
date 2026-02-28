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

type reader struct {
	tracer trace.Tracer
	db     sqlx.QueryerContext
}

func NewReader(tracer trace.Tracer, db sqlx.QueryerContext) *reader {
	return &reader{
		tracer: tracer,
		db:     db,
	}
}

func (r *reader) Get(ctx context.Context, id int64) (domain.Example, error) {
	ctx, span := r.tracer.Start(ctx, "ExampleReader/Get")
	defer span.End()

	query := `SELECT id, name FROM example WHERE id = $1`

	row := r.db.QueryRowxContext(ctx, query, id)
	if err := row.Err(); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Example{}, xerror.New(xerror.CodeNotFound, "example not found")
		}

		return domain.Example{}, err
	}

	var res domain.Example
	if err := row.StructScan(&res); err != nil {
		return domain.Example{}, err
	}
	return res, nil
}
