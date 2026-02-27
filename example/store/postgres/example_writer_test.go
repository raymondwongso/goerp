package postgres

import (
	"context"
	"errors"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/raymondwoongso/goerp/domain"
	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel/trace/noop"
)

func TestWriter_Insert(t *testing.T) {
	tracer := noop.NewTracerProvider().Tracer("test")
	input := domain.Example{Name: "test-name"}

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		db, mock, err := sqlmock.New()
		assert.NoError(t, err)
		defer db.Close()

		rows := sqlmock.NewRows([]string{"id", "name"}).
			AddRow(int64(1), "test-name")

		mock.ExpectQuery(`INSERT INTO example \(name\) VALUES \(\$1\) RETURNING id, name`).
			WithArgs("test-name").
			WillReturnRows(rows)

		writer := NewWriter(tracer, sqlx.NewDb(db, "sqlmock"))
		result, err := writer.Insert(context.Background(), input)

		assert.NoError(t, err)
		assert.Equal(t, domain.Example{ID: 1, Name: "test-name"}, result)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("error — db unknown error", func(t *testing.T) {
		t.Parallel()

		db, mock, err := sqlmock.New()
		assert.NoError(t, err)
		defer db.Close()

		mock.ExpectQuery(`INSERT INTO example \(name\) VALUES \(\$1\) RETURNING id, name`).
			WithArgs("test-name").
			WillReturnError(errors.New("some db error"))

		writer := NewWriter(tracer, sqlx.NewDb(db, "sqlmock"))
		result, err := writer.Insert(context.Background(), input)

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
			AddRow("not-a-number", "test-name")

		mock.ExpectQuery(`INSERT INTO example \(name\) VALUES \(\$1\) RETURNING id, name`).
			WithArgs("test-name").
			WillReturnRows(rows)

		writer := NewWriter(tracer, sqlx.NewDb(db, "sqlmock"))
		result, err := writer.Insert(context.Background(), input)

		assert.Error(t, err)
		assert.Empty(t, result)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}
