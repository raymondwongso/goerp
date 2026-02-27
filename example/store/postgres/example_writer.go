package postgres

import (
	"context"

	"github.com/jmoiron/sqlx"
	"github.com/raymondwoongso/goerp/domain"
	"go.opentelemetry.io/otel/trace"
)

type writer struct {
	tracer trace.Tracer
	db     sqlx.QueryerContext
}

func NewWriter(tracer trace.Tracer, db sqlx.QueryerContext) *writer {
	return &writer{
		tracer: tracer,
		db:     db,
	}
}

func (w *writer) Insert(ctx context.Context, e domain.Example) (domain.Example, error) {
	ctx, span := w.tracer.Start(ctx, "ExampleWriter/Insert")
	defer span.End()

	query := `INSERT INTO example (name) VALUES ($1) RETURNING id, name`

	row := w.db.QueryRowxContext(ctx, query, e.Name)
	if err := row.Err(); err != nil {
		return domain.Example{}, err
	}

	var res domain.Example
	if err := row.StructScan(&res); err != nil {
		return domain.Example{}, err
	}
	return res, nil
}
