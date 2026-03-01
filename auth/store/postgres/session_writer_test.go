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

func TestSessionWriter_Insert(t *testing.T) {
	tracer := noop.NewTracerProvider().Tracer("test")

	now := time.Now().UTC().Truncate(time.Second)

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		db, mock, err := sqlmock.New()
		assert.NoError(t, err)
		defer db.Close()

		rows := sqlmock.NewRows([]string{"id", "user_id", "ip_address", "user_agent", "is_revoked", "absolute_expiry", "created_at", "last_seen_at"}).
			AddRow("session-id-1", "user-id-1", "192.168.1.1", "Mozilla/5.0", false, now.Add(30*24*time.Hour), now, now)

		mock.ExpectQuery(`INSERT INTO sessions`).
			WithArgs("user-id-1", null.StringFrom("192.168.1.1"), null.StringFrom("Mozilla/5.0")).
			WillReturnRows(rows)

		w := NewSessionWriter(tracer, sqlx.NewDb(db, "sqlmock"))
		result, err := w.Insert(context.Background(), domain.Session{
			UserID:    "user-id-1",
			IPAddress: null.StringFrom("192.168.1.1"),
			UserAgent: null.StringFrom("Mozilla/5.0"),
		})

		assert.NoError(t, err)
		assert.Equal(t, "session-id-1", result.ID)
		assert.Equal(t, "user-id-1", result.UserID)
		assert.False(t, result.IsRevoked)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("error — db error", func(t *testing.T) {
		t.Parallel()

		db, mock, err := sqlmock.New()
		assert.NoError(t, err)
		defer db.Close()

		mock.ExpectQuery(`INSERT INTO sessions`).
			WithArgs("user-id-1", null.StringFrom("192.168.1.1"), null.StringFrom("Mozilla/5.0")).
			WillReturnError(errors.New("db error"))

		w := NewSessionWriter(tracer, sqlx.NewDb(db, "sqlmock"))
		result, err := w.Insert(context.Background(), domain.Session{
			UserID:    "user-id-1",
			IPAddress: null.StringFrom("192.168.1.1"),
			UserAgent: null.StringFrom("Mozilla/5.0"),
		})

		assert.Error(t, err)
		assert.Empty(t, result)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}
