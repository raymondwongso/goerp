package postgres

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/guregu/null"
	"github.com/jmoiron/sqlx"
	"github.com/raymondwongso/goerp/domain"
	"github.com/raymondwongso/goerp/domain/xerror"
	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel/trace/noop"
)

func TestOAuthStateWriter_Insert(t *testing.T) {
	tracer := noop.NewTracerProvider().Tracer("test")

	now := time.Now().UTC().Truncate(time.Second)

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		db, mock, err := sqlmock.New()
		assert.NoError(t, err)
		defer db.Close()

		rows := sqlmock.NewRows([]string{"state", "code_verifier", "redirect_to", "ip_address", "created_at", "expires_at"}).
			AddRow("state-abc", "verifier-xyz", "/dashboard", "192.168.1.1", now, now.Add(5*time.Minute))

		mock.ExpectQuery(`INSERT INTO oauth_states`).
			WithArgs("state-abc", "verifier-xyz", null.StringFrom("/dashboard"), null.StringFrom("192.168.1.1")).
			WillReturnRows(rows)

		w := NewOAuthStateWriter(tracer, sqlx.NewDb(db, "sqlmock"))
		result, err := w.Insert(context.Background(), domain.OAuthState{
			State:        "state-abc",
			CodeVerifier: "verifier-xyz",
			RedirectTo:   null.StringFrom("/dashboard"),
			IPAddress:    null.StringFrom("192.168.1.1"),
		})

		assert.NoError(t, err)
		assert.Equal(t, "state-abc", result.State)
		assert.Equal(t, "verifier-xyz", result.CodeVerifier)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("error — db error", func(t *testing.T) {
		t.Parallel()

		db, mock, err := sqlmock.New()
		assert.NoError(t, err)
		defer db.Close()

		mock.ExpectQuery(`INSERT INTO oauth_states`).
			WithArgs("state-abc", "verifier-xyz", null.StringFrom("/dashboard"), null.StringFrom("192.168.1.1")).
			WillReturnError(errors.New("db error"))

		w := NewOAuthStateWriter(tracer, sqlx.NewDb(db, "sqlmock"))
		result, err := w.Insert(context.Background(), domain.OAuthState{
			State:        "state-abc",
			CodeVerifier: "verifier-xyz",
			RedirectTo:   null.StringFrom("/dashboard"),
			IPAddress:    null.StringFrom("192.168.1.1"),
		})

		assert.Error(t, err)
		assert.Empty(t, result)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestOAuthStateWriter_DeleteByState(t *testing.T) {
	tracer := noop.NewTracerProvider().Tracer("test")

	now := time.Now().UTC().Truncate(time.Second)

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		db, mock, err := sqlmock.New()
		assert.NoError(t, err)
		defer db.Close()

		rows := sqlmock.NewRows([]string{"state", "code_verifier", "redirect_to", "ip_address", "created_at", "expires_at"}).
			AddRow("state-abc", "verifier-xyz", "/dashboard", "192.168.1.1", now, now.Add(5*time.Minute))

		mock.ExpectQuery(`DELETE FROM oauth_states`).
			WithArgs("state-abc").
			WillReturnRows(rows)

		w := NewOAuthStateWriter(tracer, sqlx.NewDb(db, "sqlmock"))
		result, err := w.DeleteByState(context.Background(), "state-abc")

		assert.NoError(t, err)
		assert.Equal(t, "state-abc", result.State)
		assert.Equal(t, "verifier-xyz", result.CodeVerifier)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("error — not found or expired", func(t *testing.T) {
		t.Parallel()

		db, mock, err := sqlmock.New()
		assert.NoError(t, err)
		defer db.Close()

		mock.ExpectQuery(`DELETE FROM oauth_states`).
			WithArgs("invalid-state").
			WillReturnError(sql.ErrNoRows)

		w := NewOAuthStateWriter(tracer, sqlx.NewDb(db, "sqlmock"))
		result, err := w.DeleteByState(context.Background(), "invalid-state")

		assert.Error(t, err)
		assert.Empty(t, result)
		assert.Equal(t, xerror.CodeNotFound, xerror.GetCode(err))
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("error — db error", func(t *testing.T) {
		t.Parallel()

		db, mock, err := sqlmock.New()
		assert.NoError(t, err)
		defer db.Close()

		mock.ExpectQuery(`DELETE FROM oauth_states`).
			WithArgs("state-abc").
			WillReturnError(errors.New("connection error"))

		w := NewOAuthStateWriter(tracer, sqlx.NewDb(db, "sqlmock"))
		result, err := w.DeleteByState(context.Background(), "state-abc")

		assert.Error(t, err)
		assert.Empty(t, result)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}
