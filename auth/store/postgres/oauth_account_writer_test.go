package postgres

import (
	"context"
	"errors"
	"testing"
	"time"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/raymondwongso/goerp/domain"
	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel/trace/noop"
)

func TestOAuthAccountWriter_Upsert(t *testing.T) {
	tracer := noop.NewTracerProvider().Tracer("test")

	now := time.Now().UTC().Truncate(time.Second)

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		db, mock, err := sqlmock.New()
		assert.NoError(t, err)
		defer db.Close()

		rows := sqlmock.NewRows([]string{"id", "user_id", "provider", "provider_sub", "email", "last_login", "created_at"}).
			AddRow("account-id-1", "user-id-1", "google", "google-sub-123", "user@example.com", now, now)

		mock.ExpectQuery(`INSERT INTO oauth_accounts`).
			WithArgs("user-id-1", domain.OAuthProviderGoogle, "google-sub-123", "user@example.com").
			WillReturnRows(rows)

		w := NewOAuthAccountWriter(tracer, sqlx.NewDb(db, "sqlmock"))
		result, err := w.Upsert(context.Background(), domain.OAuthAccount{
			UserID:      "user-id-1",
			Provider:    domain.OAuthProviderGoogle,
			ProviderSub: "google-sub-123",
			Email:       "user@example.com",
		})

		assert.NoError(t, err)
		assert.Equal(t, "account-id-1", result.ID)
		assert.Equal(t, "user-id-1", result.UserID)
		assert.Equal(t, domain.OAuthProviderGoogle, result.Provider)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("error — db error", func(t *testing.T) {
		t.Parallel()

		db, mock, err := sqlmock.New()
		assert.NoError(t, err)
		defer db.Close()

		mock.ExpectQuery(`INSERT INTO oauth_accounts`).
			WithArgs("user-id-1", domain.OAuthProviderGoogle, "google-sub-123", "user@example.com").
			WillReturnError(errors.New("db error"))

		w := NewOAuthAccountWriter(tracer, sqlx.NewDb(db, "sqlmock"))
		result, err := w.Upsert(context.Background(), domain.OAuthAccount{
			UserID:      "user-id-1",
			Provider:    domain.OAuthProviderGoogle,
			ProviderSub: "google-sub-123",
			Email:       "user@example.com",
		})

		assert.Error(t, err)
		assert.Empty(t, result)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}
