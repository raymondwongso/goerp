package postgres

import (
	"context"
	"errors"
	"testing"
	"time"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/guregu/null"
	"github.com/jmoiron/sqlx"
	"github.com/raymondwongso/goerp/domain"
	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel/trace/noop"
)

func TestUserWriter_Upsert(t *testing.T) {
	tracer := noop.NewTracerProvider().Tracer("test")

	now := time.Now().UTC().Truncate(time.Second)

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		db, mock, err := sqlmock.New()
		assert.NoError(t, err)
		defer db.Close()

		rows := sqlmock.NewRows([]string{"id", "email", "display_name", "avatar_url", "is_active", "created_at", "updated_at"}).
			AddRow("user-id-1", "user@example.com", "Test User", "https://example.com/pic.jpg", true, now, now)

		mock.ExpectQuery(`INSERT INTO users`).
			WithArgs("user@example.com", null.StringFrom("Test User"), null.StringFrom("https://example.com/pic.jpg")).
			WillReturnRows(rows)

		w := NewUserWriter(tracer, sqlx.NewDb(db, "sqlmock"))
		result, err := w.Upsert(context.Background(), domain.User{
			Email:       "user@example.com",
			DisplayName: null.StringFrom("Test User"),
			AvatarURL:   null.StringFrom("https://example.com/pic.jpg"),
		})

		assert.NoError(t, err)
		assert.Equal(t, "user-id-1", result.ID)
		assert.Equal(t, "user@example.com", result.Email)
		assert.Equal(t, null.StringFrom("Test User"), result.DisplayName)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("error — db error", func(t *testing.T) {
		t.Parallel()

		db, mock, err := sqlmock.New()
		assert.NoError(t, err)
		defer db.Close()

		mock.ExpectQuery(`INSERT INTO users`).
			WithArgs("user@example.com", null.StringFrom("Test User"), null.StringFrom("https://example.com/pic.jpg")).
			WillReturnError(errors.New("db error"))

		w := NewUserWriter(tracer, sqlx.NewDb(db, "sqlmock"))
		result, err := w.Upsert(context.Background(), domain.User{
			Email:       "user@example.com",
			DisplayName: null.StringFrom("Test User"),
			AvatarURL:   null.StringFrom("https://example.com/pic.jpg"),
		})

		assert.Error(t, err)
		assert.Empty(t, result)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}
