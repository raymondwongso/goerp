package auth

import (
	"context"
	"net/http"

	"github.com/jmoiron/sqlx"
	authhttp "github.com/raymondwongso/goerp/auth/http"
	googlestore "github.com/raymondwongso/goerp/auth/store/google"
	authpostgres "github.com/raymondwongso/goerp/auth/store/postgres"
	googleuc "github.com/raymondwongso/goerp/auth/usecase/google"
	"go.opentelemetry.io/otel/trace"
)

// Config holds the configuration for the auth module.
type Config struct {
	GoogleClientID     string
	GoogleClientSecret string
	GoogleRedirectURL  string
}

// RegisterHTTPHandlers initializes all auth dependencies and registers routes on the given mux.
func RegisterHTTPHandlers(ctx context.Context, mux *http.ServeMux, db *sqlx.DB, tracer trace.Tracer, cfg Config) error {
	googleProvider, err := googlestore.NewProvider(ctx, googlestore.ProviderParam{
		ClientID:     cfg.GoogleClientID,
		ClientSecret: cfg.GoogleClientSecret,
		RedirectURL:  cfg.GoogleRedirectURL,
	})
	if err != nil {
		return err
	}

	oauthStateWriter := authpostgres.NewOAuthStateWriter(tracer, db)
	userWriter := authpostgres.NewUserWriter(tracer, db)
	oauthAccountWriter := authpostgres.NewOAuthAccountWriter(tracer, db)
	sessionWriter := authpostgres.NewSessionWriter(tracer, db)

	loginUC := googleuc.NewLogin(googleProvider, oauthStateWriter)
	callbackUC := googleuc.NewCallback(googleProvider, oauthStateWriter, userWriter, oauthAccountWriter, sessionWriter)

	handler := authhttp.NewHandler(authhttp.HandlerParam{
		GoogleLogin:    loginUC,
		GoogleCallback: callbackUC,
	})

	mux.HandleFunc("PUT /auth/google/login", handler.GoogleLogin)
	mux.HandleFunc("POST /auth/google/callback", handler.GoogleCallback)

	return nil
}
