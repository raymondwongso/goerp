package postgres

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/raymondwoongso/goerp/domain"
	"github.com/raymondwoongso/goerp/domain/xerror"
	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel/trace/noop"
)

func TestReader_Get(t *testing.T) {
	tracer := noop.NewTracerProvider().Tracer("test")

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		db, mock, err := sqlmock.New()
		assert.NoError(t, err)
		defer db.Close()

		rows := sqlmock.NewRows([]string{"id", "name"}).
			AddRow(int64(1), "test-name")

		mock.ExpectQuery(`SELECT id, name FROM example WHERE id = \$1`).
			WithArgs(int64(1)).
			WillReturnRows(rows)

		reader := NewReader(tracer, sqlx.NewDb(db, "sqlmock"))
		result, err := reader.Get(context.Background(), 1)

		assert.NoError(t, err)
		assert.Equal(t, domain.Example{ID: 1, Name: "test-name"}, result)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("error — not found", func(t *testing.T) {
		t.Parallel()

		db, mock, err := sqlmock.New()
		assert.NoError(t, err)
		defer db.Close()

		mock.ExpectQuery(`SELECT id, name FROM example WHERE id = \$1`).
			WithArgs(int64(999)).
			WillReturnError(sql.ErrNoRows)

		reader := NewReader(tracer, sqlx.NewDb(db, "sqlmock"))
		result, err := reader.Get(context.Background(), 999)

		assert.Error(t, err)
		assert.Empty(t, result)
		assert.Equal(t, xerror.CodeNotFound, xerror.GetCode(err))
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("error — db unknown error", func(t *testing.T) {
		t.Parallel()

		db, mock, err := sqlmock.New()
		assert.NoError(t, err)
		defer db.Close()

		mock.ExpectQuery(`SELECT id, name FROM example WHERE id = \$1`).
			WithArgs(int64(1)).
			WillReturnError(errors.New("some db error"))

		reader := NewReader(tracer, sqlx.NewDb(db, "sqlmock"))
		result, err := reader.Get(context.Background(), 1)

		assert.Error(t, err)
		assert.Empty(t, result)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("error — struct scan error", func(t *testing.T) {
		t.Parallel()

		db, mock, err := sqlmock.New()
		assert.NoError(t, err)
		defer db.Close()

		rows := sqlmock.NewRows([]string{"id", "name"}).
			AddRow("string not number", "test-name")

		mock.ExpectQuery(`SELECT id, name FROM example WHERE id = \$1`).
			WithArgs(int64(1)).
			WillReturnRows(rows)

		reader := NewReader(tracer, sqlx.NewDb(db, "sqlmock"))
		result, err := reader.Get(context.Background(), 1)

		assert.Error(t, err)
		assert.Empty(t, result)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}
